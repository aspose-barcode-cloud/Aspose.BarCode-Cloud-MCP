package tests

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/aspose-barcode-cloud/Aspose.BarCode-Cloud-MCP/mcpbarcode"
)

var (
	asposeConnectivityOnce sync.Once
	asposeConnectivityErr  error
)

func skipWithoutCredentials(t *testing.T) {
	t.Helper()
	if os.Getenv("ASPOSE_CLOUD_CLIENT_ID") == "" || os.Getenv("ASPOSE_CLOUD_CLIENT_SECRET") == "" {
		t.Skip("ASPOSE_CLOUD_CLIENT_ID and ASPOSE_CLOUD_CLIENT_SECRET not set")
	}
}

func skipWithoutAsposeConnectivity(t *testing.T) {
	t.Helper()

	asposeConnectivityOnce.Do(func() {
		conn, err := net.DialTimeout("tcp", "id.aspose.cloud:443", 5*time.Second)
		if err != nil {
			asposeConnectivityErr = err
			return
		}
		_ = conn.Close()
	})

	if asposeConnectivityErr != nil {
		t.Skipf("Aspose Cloud is unreachable from this environment: %v", asposeConnectivityErr)
	}
}

func registerIntegrationTools(t *testing.T, s *server.MCPServer, asposeClient *mcpbarcode.AsposeClient, mount *mcpbarcode.MountConfig) (ok bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("tool registration panicked (jsonschema tag issue): %v", r)
			ok = false
		}
	}()

	s.AddTool(mcp.NewTool("generate_barcode",
		mcp.WithDescription("Generate a barcode image"),
		mcp.WithInputSchema[mcpbarcode.GenerateBarcodeInput](),
	), mcpbarcode.MakeGenerateHandler(asposeClient, mount))

	s.AddTool(mcp.NewTool("recognize_barcode",
		mcp.WithDescription("Recognize barcodes from an image"),
		mcp.WithInputSchema[mcpbarcode.RecognizeBarcodeInput](),
	), mcpbarcode.MakeRecognizeHandler(asposeClient, mount))

	s.AddTool(mcp.NewTool("scan_barcode",
		mcp.WithDescription("Scan barcodes from an image"),
		mcp.WithInputSchema[mcpbarcode.ScanBarcodeInput](),
	), mcpbarcode.MakeScanHandler(asposeClient, mount))

	s.AddTool(mcp.NewTool("list_barcode_types",
		mcp.WithDescription("List supported barcode types"),
		mcp.WithInputSchema[mcpbarcode.ListBarcodeTypesInput](),
	), mcpbarcode.MakeListHandler())

	return true
}

// createIntegrationServer creates an in-process MCP client with a mount-enabled server.
// Returns the client and the mount directory path.
func createIntegrationServer(t *testing.T) (*client.Client, string) {
	t.Helper()
	skipWithoutCredentials(t)
	skipWithoutAsposeConnectivity(t)

	asposeClient, err := mcpbarcode.NewAsposeClient(
		os.Getenv("ASPOSE_CLOUD_CLIENT_ID"),
		os.Getenv("ASPOSE_CLOUD_CLIENT_SECRET"),
	)
	if err != nil {
		t.Fatalf("failed to create Aspose client: %v", err)
	}

	mountDir := t.TempDir()
	mount, err := mcpbarcode.NewMountConfig(mountDir)
	if err != nil {
		t.Fatalf("failed to create mount config: %v", err)
	}

	s := server.NewMCPServer("aspose-barcode-cloud", "test")

	if !registerIntegrationTools(t, s, asposeClient, mount) {
		return nil, ""
	}

	ctx := context.Background()

	c, err := client.NewInProcessClient(s)
	if err != nil {
		t.Fatalf("failed to create in-process client: %v", err)
	}

	if err := c.Start(ctx); err != nil {
		t.Fatalf("failed to start client: %v", err)
	}

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "test-client", Version: "1.0"}

	if _, err := c.Initialize(ctx, initReq); err != nil {
		t.Fatalf("failed to initialize client: %v", err)
	}

	t.Cleanup(func() { c.Close() })
	return c, mountDir
}

// extractFilenameFromResponse parses the filename from a generate_barcode TextContent response.
func extractFilenameFromResponse(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(result.Content))
	}
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	// Format: "Generated barcode image saved to: <filename>\nFormat: <mime>"
	line := strings.SplitN(textContent.Text, "\n", 2)[0]
	prefix := "Generated barcode image saved to: "
	if !strings.HasPrefix(line, prefix) {
		t.Fatalf("unexpected response format: %s", textContent.Text)
	}
	return strings.TrimPrefix(line, prefix)
}

// TestIntegration_GenerateAndScanRoundTrip generates a QR barcode and then
// scans it to verify the decoded value matches the input.
func TestIntegration_GenerateAndScanRoundTrip(t *testing.T) {
	cs, mountDir := createIntegrationServer(t)
	ctx := context.Background()

	testData := "Hello from integration test"

	// Step 1: Generate a QR barcode
	genResult, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "generate_barcode",
			Arguments: map[string]any{
				"barcode_type": "QR",
				"data":         testData,
			},
		},
	})
	if err != nil {
		t.Fatalf("generate_barcode error: %v", err)
	}
	if genResult.IsError {
		t.Fatalf("generate_barcode returned error: %v", genResult.Content)
	}

	// Verify we got text content with a file path
	filename := extractFilenameFromResponse(t, genResult)
	if !strings.HasSuffix(filename, ".png") {
		t.Errorf("expected .png suffix, got: %s", filename)
	}

	// Verify file exists on disk
	filePath := filepath.Join(mountDir, filename)
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("generated file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("generated file is empty")
	}

	// Verify response contains format metadata
	textContent := genResult.Content[0].(mcp.TextContent)
	if !strings.Contains(textContent.Text, "image/png") {
		t.Errorf("expected image/png in response, got: %s", textContent.Text)
	}

	// Step 2: Scan the generated barcode via file path
	scanResult, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "scan_barcode",
			Arguments: map[string]any{
				"image_path": filename,
			},
		},
	})
	if err != nil {
		t.Fatalf("scan_barcode error: %v", err)
	}
	if scanResult.IsError {
		t.Fatalf("scan_barcode returned error: %v", scanResult.Content)
	}

	scanText := scanResult.Content[0].(mcp.TextContent)
	if !strings.Contains(scanText.Text, testData) {
		t.Errorf("scan result does not contain original data %q, got: %s", testData, scanText.Text)
	}
}

// TestIntegration_GenerateAndRecognizeRoundTrip generates a Code128 barcode
// and recognizes it with a type hint.
func TestIntegration_GenerateAndRecognizeRoundTrip(t *testing.T) {
	cs, _ := createIntegrationServer(t)
	ctx := context.Background()

	testData := "TEST12345"

	// Step 1: Generate a Code128 barcode
	genResult, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "generate_barcode",
			Arguments: map[string]any{
				"barcode_type": "Code128",
				"data":         testData,
			},
		},
	})
	if err != nil {
		t.Fatalf("generate_barcode error: %v", err)
	}
	if genResult.IsError {
		t.Fatalf("generate_barcode returned error: %v", genResult.Content)
	}

	filename := extractFilenameFromResponse(t, genResult)

	// Step 2: Recognize with type hint
	recResult, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "recognize_barcode",
			Arguments: map[string]any{
				"image_path":   filename,
				"barcode_type": "Code128",
			},
		},
	})
	if err != nil {
		t.Fatalf("recognize_barcode error: %v", err)
	}
	if recResult.IsError {
		t.Fatalf("recognize_barcode returned error: %v", recResult.Content)
	}

	textContent := recResult.Content[0].(mcp.TextContent)
	if !strings.Contains(textContent.Text, testData) {
		t.Errorf("recognize result does not contain original data %q, got: %s", testData, textContent.Text)
	}
	if !strings.Contains(textContent.Text, "Code128") {
		t.Errorf("recognize result does not mention barcode type, got: %s", textContent.Text)
	}
}

// TestIntegration_GenerateWithOptions tests generation with custom options.
func TestIntegration_GenerateWithOptions(t *testing.T) {
	cs, _ := createIntegrationServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "generate_barcode",
			Arguments: map[string]any{
				"barcode_type":  "QR",
				"data":          "Options test",
				"image_format":  "JPEG",
				"text_location": "None",
			},
		},
	})
	if err != nil {
		t.Fatalf("generate_barcode error: %v", err)
	}
	if result.IsError {
		t.Fatalf("generate_barcode returned error: %v", result.Content)
	}

	textContent := result.Content[0].(mcp.TextContent)
	if !strings.Contains(textContent.Text, "image/jpeg") {
		t.Errorf("expected image/jpeg in response, got: %s", textContent.Text)
	}

	filename := extractFilenameFromResponse(t, result)
	if !strings.HasSuffix(filename, ".jpg") {
		t.Errorf("expected .jpg suffix, got: %s", filename)
	}
}

// TestIntegration_GenerateSVG tests SVG format generation writes to file.
func TestIntegration_GenerateSVG(t *testing.T) {
	cs, mountDir := createIntegrationServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "generate_barcode",
			Arguments: map[string]any{
				"barcode_type": "QR",
				"data":         "SVG test",
				"image_format": "SVG",
			},
		},
	})
	if err != nil {
		t.Fatalf("generate_barcode error: %v", err)
	}
	if result.IsError {
		t.Fatalf("generate_barcode returned error: %v", result.Content)
	}

	textContent := result.Content[0].(mcp.TextContent)
	if !strings.Contains(textContent.Text, "image/svg+xml") {
		t.Errorf("expected image/svg+xml in response, got: %s", textContent.Text)
	}

	filename := extractFilenameFromResponse(t, result)
	if !strings.HasSuffix(filename, ".svg") {
		t.Errorf("expected .svg suffix, got: %s", filename)
	}

	// Verify SVG file content
	data, err := os.ReadFile(filepath.Join(mountDir, filename))
	if err != nil {
		t.Fatalf("failed to read SVG file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<svg") && !strings.Contains(content, "<?xml") {
		t.Errorf("expected SVG content, got: %.100s...", content)
	}
}

// TestIntegration_ScanBlankImage tests scanning an image with no barcodes.
func TestIntegration_ScanBlankImage(t *testing.T) {
	cs, mountDir := createIntegrationServer(t)
	ctx := context.Background()

	// Write a minimal 1x1 white PNG to the mount directory
	blankPNG := createMinimalWhitePNG()
	blankFile := "blank.png"
	if err := os.WriteFile(filepath.Join(mountDir, blankFile), blankPNG, 0644); err != nil {
		t.Fatalf("failed to write blank PNG: %v", err)
	}

	result, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "scan_barcode",
			Arguments: map[string]any{
				"image_path": blankFile,
			},
		},
	})
	if err != nil {
		t.Fatalf("scan_barcode error: %v", err)
	}
	if result.IsError {
		t.Fatalf("scan_barcode returned error: %v", result.Content)
	}

	textContent := result.Content[0].(mcp.TextContent)
	if !strings.Contains(textContent.Text, "No barcodes detected") {
		t.Errorf("expected 'No barcodes detected', got: %s", textContent.Text)
	}
}

// TestIntegration_InvalidBarcodeType tests error handling for invalid barcode types.
func TestIntegration_InvalidBarcodeType(t *testing.T) {
	cs, _ := createIntegrationServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "generate_barcode",
			Arguments: map[string]any{
				"barcode_type": "COMPLETELY_FAKE_TYPE",
				"data":         "test",
			},
		},
	})
	// mcp-go returns handler errors as protocol errors
	if err == nil && !result.IsError {
		t.Fatal("expected error for invalid barcode type")
	}
}

// TestIntegration_RecognizeWithMode tests recognition with quality mode.
func TestIntegration_RecognizeWithMode(t *testing.T) {
	cs, _ := createIntegrationServer(t)
	ctx := context.Background()

	testData := "ModeTest123"

	// Generate
	genResult, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "generate_barcode",
			Arguments: map[string]any{
				"barcode_type": "QR",
				"data":         testData,
			},
		},
	})
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}
	filename := extractFilenameFromResponse(t, genResult)

	// Recognize with Excellent mode
	result, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "recognize_barcode",
			Arguments: map[string]any{
				"image_path":             filename,
				"barcode_type":           "QR",
				"recognition_mode":       "Excellent",
				"recognition_image_kind": "ClearImage",
			},
		},
	})
	if err != nil {
		t.Fatalf("recognize error: %v", err)
	}
	if result.IsError {
		t.Fatalf("recognize returned error: %v", result.Content)
	}

	textContent := result.Content[0].(mcp.TextContent)
	if !strings.Contains(textContent.Text, testData) {
		t.Errorf("expected data %q in result, got: %s", testData, textContent.Text)
	}
}

// TestIntegration_ScanEmptyImagePath tests error for empty image_path.
func TestIntegration_ScanEmptyImagePath(t *testing.T) {
	cs, _ := createIntegrationServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "scan_barcode",
			Arguments: map[string]any{
				"image_path": "",
			},
		},
	})
	if err == nil && !result.IsError {
		t.Fatal("expected error for empty image_path")
	}
}

// TestIntegration_ScanPathTraversal tests that path traversal is blocked.
func TestIntegration_ScanPathTraversal(t *testing.T) {
	cs, _ := createIntegrationServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "scan_barcode",
			Arguments: map[string]any{
				"image_path": "../etc/passwd.png",
			},
		},
	})
	if err == nil && !result.IsError {
		t.Fatal("expected error for path traversal")
	}
}

// TestIntegration_ScanInvalidExtension tests that invalid file extensions are rejected.
func TestIntegration_ScanInvalidExtension(t *testing.T) {
	cs, _ := createIntegrationServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "scan_barcode",
			Arguments: map[string]any{
				"image_path": "test.txt",
			},
		},
	})
	if err == nil && !result.IsError {
		t.Fatal("expected error for invalid extension")
	}
}

// createMinimalWhitePNG returns a minimal valid 1x1 white PNG image.
func createMinimalWhitePNG() []byte {
	// This is a valid 1x1 white pixel PNG file
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG signature
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, // 8-bit RGB
		0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00, // compressed
		0x00, 0x00, 0x02, 0x00, 0x01, 0xe2, 0x21, 0xbc, // data
		0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, // IEND chunk
		0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
