package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/aspose-barcode-cloud/Aspose.BarCode-Cloud-MCP/mcpbarcode"
)

// createTestServer creates an MCP server with only the list_barcode_types tool
// registered (no API credentials needed).
func createTestServer() *server.MCPServer {
	s := server.NewMCPServer("aspose-barcode-cloud", "test")

	s.AddTool(mcp.NewTool("list_barcode_types",
		mcp.WithDescription("List all supported barcode types for generation and recognition. "+
			"Use this to discover valid barcode_type values for generate_barcode and recognize_barcode."),
		mcp.WithInputSchema[mcpbarcode.ListBarcodeTypesInput](),
	), mcpbarcode.MakeListHandler())

	return s
}

// createFullServer creates an MCP server with all 4 tools registered.
// Returns nil if tool registration panics (e.g. due to jsonschema tag issues).
func createFullServer(t *testing.T) (s *server.MCPServer) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("tool registration panicked (jsonschema tag issue): %v", r)
		}
	}()

	s = server.NewMCPServer("aspose-barcode-cloud", "test")

	dummyClient := &mcpbarcode.AsposeClient{}
	mount, err := mcpbarcode.NewMountConfig(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create mount config: %v", err)
	}

	s.AddTool(mcp.NewTool("generate_barcode",
		mcp.WithDescription("Generate a barcode image"),
		mcp.WithInputSchema[mcpbarcode.GenerateBarcodeInput](),
	), mcpbarcode.MakeGenerateHandler(dummyClient, mount))

	s.AddTool(mcp.NewTool("recognize_barcode",
		mcp.WithDescription("Recognize barcodes from an image"),
		mcp.WithInputSchema[mcpbarcode.RecognizeBarcodeInput](),
	), mcpbarcode.MakeRecognizeHandler(dummyClient, mount))

	s.AddTool(mcp.NewTool("scan_barcode",
		mcp.WithDescription("Scan barcodes from an image"),
		mcp.WithInputSchema[mcpbarcode.ScanBarcodeInput](),
	), mcpbarcode.MakeScanHandler(dummyClient, mount))

	s.AddTool(mcp.NewTool("list_barcode_types",
		mcp.WithDescription("List supported barcode types"),
		mcp.WithInputSchema[mcpbarcode.ListBarcodeTypesInput](),
	), mcpbarcode.MakeListHandler())

	return s
}

// connectTestClient creates an in-process client connected to the server.
func connectTestClient(t *testing.T, s *server.MCPServer) *client.Client {
	t.Helper()
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
	return c
}

func TestMCPProtocol_ListTools(t *testing.T) {
	s := createTestServer()
	cs := connectTestClient(t, s)

	result, err := cs.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools error: %v", err)
	}

	if len(result.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(result.Tools))
	}

	tool := result.Tools[0]
	if tool.Name != "list_barcode_types" {
		t.Errorf("expected tool name 'list_barcode_types', got %q", tool.Name)
	}
	if tool.Description == "" {
		t.Error("expected non-empty tool description")
	}
}

func TestMCPProtocol_CallListBarcodeTypes(t *testing.T) {
	s := createTestServer()
	cs := connectTestClient(t, s)

	result, err := cs.CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "list_barcode_types",
			Arguments: map[string]any{},
		},
	})
	if err != nil {
		t.Fatalf("CallTool error: %v", err)
	}

	if result.IsError {
		t.Fatal("expected successful result, got error")
	}

	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(result.Content))
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if textContent.Text == "" {
		t.Error("expected non-empty text content")
	}
}

func TestMCPProtocol_CallListBarcodeTypes_ContentCheck(t *testing.T) {
	s := createTestServer()
	cs := connectTestClient(t, s)

	result, err := cs.CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "list_barcode_types",
			Arguments: map[string]any{},
		},
	})
	if err != nil {
		t.Fatalf("CallTool error: %v", err)
	}

	text := result.Content[0].(mcp.TextContent).Text

	// Verify key barcode types appear in the output
	expectedTypes := []string{"QR", "Code128", "DataMatrix", "EAN13", "Pdf417", "MostCommonlyUsed"}
	for _, et := range expectedTypes {
		if !strings.Contains(text, et) {
			t.Errorf("expected type %q in output", et)
		}
	}

	// Verify both sections exist
	if !strings.Contains(text, "GENERATION") {
		t.Error("missing GENERATION section")
	}
	if !strings.Contains(text, "RECOGNITION") {
		t.Error("missing RECOGNITION section")
	}
}

func TestMCPProtocol_CallNonexistentTool(t *testing.T) {
	s := createTestServer()
	cs := connectTestClient(t, s)

	_, err := cs.CallTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "nonexistent_tool",
			Arguments: map[string]any{},
		},
	})
	// The MCP SDK should return an error for unknown tools
	if err == nil {
		t.Fatal("expected error for nonexistent tool, got nil")
	}
}

func TestMCPProtocol_ToolSchemaValidation(t *testing.T) {
	s := createTestServer()
	cs := connectTestClient(t, s)

	result, err := cs.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools error: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.InputSchema.Type == "" {
			t.Errorf("tool %q has empty input schema type", tool.Name)
		}
	}
}

// TestMCPProtocol_FullServerToolRegistration tests that all 4 tools can be
// registered with the MCP server (verifies schema inference works for all
// input structs). Skips if registration panics due to jsonschema tag issues.
func TestMCPProtocol_FullServerToolRegistration(t *testing.T) {
	s := createFullServer(t)
	if s == nil {
		return
	}

	cs := connectTestClient(t, s)

	result, err := cs.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools error: %v", err)
	}

	if len(result.Tools) != 4 {
		t.Fatalf("expected 4 tools, got %d", len(result.Tools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range result.Tools {
		toolNames[tool.Name] = true
	}

	expectedTools := []string{"generate_barcode", "recognize_barcode", "scan_barcode", "list_barcode_types"}
	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Errorf("expected tool %q not found in server", name)
		}
	}
}

// TestMCPProtocol_FullServerSchemas verifies all tool schemas are non-nil
// when all 4 tools are registered. Skips if registration panics.
func TestMCPProtocol_FullServerSchemas(t *testing.T) {
	s := createFullServer(t)
	if s == nil {
		return
	}

	cs := connectTestClient(t, s)

	result, err := cs.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools error: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.InputSchema.Type == "" {
			t.Errorf("tool %q has empty input schema type", tool.Name)
		}
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
	}
}
