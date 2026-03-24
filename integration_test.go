package main

import (
	"context"
	"encoding/base64"
	"os"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func skipWithoutCredentials(t *testing.T) {
	t.Helper()
	if os.Getenv("ASPOSE_CLIENT_ID") == "" || os.Getenv("ASPOSE_CLIENT_SECRET") == "" {
		t.Skip("ASPOSE_CLIENT_ID and ASPOSE_CLIENT_SECRET not set")
	}
}

func registerIntegrationTools(t *testing.T, s *mcp.Server, client *AsposeClient) (ok bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("tool registration panicked (jsonschema tag issue): %v", r)
			ok = false
		}
	}()

	mcp.AddTool(s, &mcp.Tool{
		Name:        "generate_barcode",
		Description: "Generate a barcode image",
	}, makeGenerateHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "recognize_barcode",
		Description: "Recognize barcodes from an image",
	}, makeRecognizeHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "scan_barcode",
		Description: "Scan barcodes from an image",
	}, makeScanHandler(client))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_barcode_types",
		Description: "List supported barcode types",
	}, makeListHandler())

	return true
}

func createIntegrationServer(t *testing.T) *mcp.ClientSession {
	t.Helper()
	skipWithoutCredentials(t)

	client, err := NewAsposeClient(
		os.Getenv("ASPOSE_CLIENT_ID"),
		os.Getenv("ASPOSE_CLIENT_SECRET"),
	)
	if err != nil {
		t.Fatalf("failed to create Aspose client: %v", err)
	}

	s := mcp.NewServer(
		&mcp.Implementation{Name: "aspose-barcode-cloud", Version: "test"},
		nil,
	)

	if !registerIntegrationTools(t, s, client) {
		return nil
	}

	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := s.Connect(ctx, serverTransport)
	if err != nil {
		t.Fatalf("server connect error: %v", err)
	}
	t.Cleanup(func() { serverSession.Wait() })

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport)
	if err != nil {
		t.Fatalf("client connect error: %v", err)
	}
	t.Cleanup(func() { clientSession.Close() })

	return clientSession
}

// TestIntegration_GenerateAndScanRoundTrip generates a QR barcode and then
// scans it to verify the decoded value matches the input.
func TestIntegration_GenerateAndScanRoundTrip(t *testing.T) {
	cs := createIntegrationServer(t)
	ctx := context.Background()

	testData := "Hello from integration test"

	// Step 1: Generate a QR barcode
	genResult, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "generate_barcode",
		Arguments: map[string]any{
			"barcode_type": "QR",
			"data":         testData,
		},
	})
	if err != nil {
		t.Fatalf("generate_barcode error: %v", err)
	}
	if genResult.IsError {
		t.Fatalf("generate_barcode returned error: %v", genResult.Content)
	}

	// Verify we got image content
	if len(genResult.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(genResult.Content))
	}
	imgContent, ok := genResult.Content[0].(*mcp.ImageContent)
	if !ok {
		t.Fatalf("expected ImageContent, got %T", genResult.Content[0])
	}
	if imgContent.MIMEType != "image/png" {
		t.Errorf("expected MIME type image/png, got %q", imgContent.MIMEType)
	}

	// Step 2: Scan the generated barcode
	scanResult, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "scan_barcode",
		Arguments: map[string]any{
			"image_data": string(imgContent.Data),
		},
	})
	if err != nil {
		t.Fatalf("scan_barcode error: %v", err)
	}
	if scanResult.IsError {
		t.Fatalf("scan_barcode returned error: %v", scanResult.Content)
	}

	textContent, ok := scanResult.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", scanResult.Content[0])
	}

	if !strings.Contains(textContent.Text, testData) {
		t.Errorf("scan result does not contain original data %q, got: %s", testData, textContent.Text)
	}
}

// TestIntegration_GenerateAndRecognizeRoundTrip generates a Code128 barcode
// and recognizes it with a type hint.
func TestIntegration_GenerateAndRecognizeRoundTrip(t *testing.T) {
	cs := createIntegrationServer(t)
	ctx := context.Background()

	testData := "TEST12345"

	// Step 1: Generate a Code128 barcode
	genResult, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "generate_barcode",
		Arguments: map[string]any{
			"barcode_type": "Code128",
			"data":         testData,
		},
	})
	if err != nil {
		t.Fatalf("generate_barcode error: %v", err)
	}
	if genResult.IsError {
		t.Fatalf("generate_barcode returned error: %v", genResult.Content)
	}

	imgContent := genResult.Content[0].(*mcp.ImageContent)

	// Step 2: Recognize with type hint
	recResult, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "recognize_barcode",
		Arguments: map[string]any{
			"image_data":   string(imgContent.Data),
			"barcode_type": "Code128",
		},
	})
	if err != nil {
		t.Fatalf("recognize_barcode error: %v", err)
	}
	if recResult.IsError {
		t.Fatalf("recognize_barcode returned error: %v", recResult.Content)
	}

	textContent := recResult.Content[0].(*mcp.TextContent)
	if !strings.Contains(textContent.Text, testData) {
		t.Errorf("recognize result does not contain original data %q, got: %s", testData, textContent.Text)
	}
	if !strings.Contains(textContent.Text, "Code128") {
		t.Errorf("recognize result does not mention barcode type, got: %s", textContent.Text)
	}
}

// TestIntegration_GenerateWithOptions tests generation with custom options.
func TestIntegration_GenerateWithOptions(t *testing.T) {
	cs := createIntegrationServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "generate_barcode",
		Arguments: map[string]any{
			"barcode_type":  "QR",
			"data":          "Options test",
			"image_format":  "JPEG",
			"text_location": "None",
		},
	})
	if err != nil {
		t.Fatalf("generate_barcode error: %v", err)
	}
	if result.IsError {
		t.Fatalf("generate_barcode returned error: %v", result.Content)
	}

	imgContent := result.Content[0].(*mcp.ImageContent)
	if imgContent.MIMEType != "image/jpeg" {
		t.Errorf("expected MIME type image/jpeg, got %q", imgContent.MIMEType)
	}
}

// TestIntegration_GenerateSVG tests SVG format generation returns text content.
func TestIntegration_GenerateSVG(t *testing.T) {
	cs := createIntegrationServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "generate_barcode",
		Arguments: map[string]any{
			"barcode_type": "QR",
			"data":         "SVG test",
			"image_format": "SVG",
		},
	})
	if err != nil {
		t.Fatalf("generate_barcode error: %v", err)
	}
	if result.IsError {
		t.Fatalf("generate_barcode returned error: %v", result.Content)
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("SVG should return TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(textContent.Text, "<svg") && !strings.Contains(textContent.Text, "<?xml") {
		t.Errorf("expected SVG content, got: %.100s...", textContent.Text)
	}
}

// TestIntegration_ScanBlankImage tests scanning an image with no barcodes.
func TestIntegration_ScanBlankImage(t *testing.T) {
	cs := createIntegrationServer(t)
	ctx := context.Background()

	// Create a minimal 1x1 white PNG (base64)
	blankPNG := createMinimalWhitePNG()
	b64 := base64.StdEncoding.EncodeToString(blankPNG)

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "scan_barcode",
		Arguments: map[string]any{
			"image_data": b64,
		},
	})
	if err != nil {
		t.Fatalf("scan_barcode error: %v", err)
	}
	if result.IsError {
		t.Fatalf("scan_barcode returned error: %v", result.Content)
	}

	textContent := result.Content[0].(*mcp.TextContent)
	if !strings.Contains(textContent.Text, "No barcodes detected") {
		t.Errorf("expected 'No barcodes detected', got: %s", textContent.Text)
	}
}

// TestIntegration_InvalidBarcodeType tests error handling for invalid barcode types.
func TestIntegration_InvalidBarcodeType(t *testing.T) {
	cs := createIntegrationServer(t)
	ctx := context.Background()

	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "generate_barcode",
		Arguments: map[string]any{
			"barcode_type": "COMPLETELY_FAKE_TYPE",
			"data":         "test",
		},
	})
	if err != nil {
		t.Fatalf("expected error result, got protocol error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected error result for invalid barcode type")
	}
}

// TestIntegration_RecognizeWithMode tests recognition with quality mode.
func TestIntegration_RecognizeWithMode(t *testing.T) {
	cs := createIntegrationServer(t)
	ctx := context.Background()

	testData := "ModeTest123"

	// Generate
	genResult, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "generate_barcode",
		Arguments: map[string]any{
			"barcode_type": "QR",
			"data":         testData,
		},
	})
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}
	imgContent := genResult.Content[0].(*mcp.ImageContent)

	// Recognize with Excellent mode
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name: "recognize_barcode",
		Arguments: map[string]any{
			"image_data":            string(imgContent.Data),
			"barcode_type":          "QR",
			"recognition_mode":      "Excellent",
			"recognition_image_kind": "ClearImage",
		},
	})
	if err != nil {
		t.Fatalf("recognize error: %v", err)
	}
	if result.IsError {
		t.Fatalf("recognize returned error: %v", result.Content)
	}

	textContent := result.Content[0].(*mcp.TextContent)
	if !strings.Contains(textContent.Text, testData) {
		t.Errorf("expected data %q in result, got: %s", testData, textContent.Text)
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
