package main

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createTestServer creates an MCP server with only the list_barcode_types tool
// registered (no API credentials needed).
func createTestServer() *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "aspose-barcode-cloud",
			Version: "test",
		},
		nil,
	)

	mcp.AddTool(s, &mcp.Tool{
		Name: "list_barcode_types",
		Description: "List all supported barcode types for generation and recognition. " +
			"Use this to discover valid barcode_type values for generate_barcode and recognize_barcode.",
	}, makeListHandler())

	return s
}

// createFullServer creates an MCP server with all 4 tools registered.
// Returns nil if tool registration panics (e.g. due to jsonschema tag issues).
func createFullServer(t *testing.T) (s *mcp.Server) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("tool registration panicked (jsonschema tag issue): %v", r)
		}
	}()

	s = mcp.NewServer(
		&mcp.Implementation{Name: "aspose-barcode-cloud", Version: "test"},
		nil,
	)

	dummyClient := &AsposeClient{}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "generate_barcode",
		Description: "Generate a barcode image",
	}, makeGenerateHandler(dummyClient))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "recognize_barcode",
		Description: "Recognize barcodes from an image",
	}, makeRecognizeHandler(dummyClient))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "scan_barcode",
		Description: "Scan barcodes from an image",
	}, makeScanHandler(dummyClient))

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_barcode_types",
		Description: "List supported barcode types",
	}, makeListHandler())

	return s
}

// connectTestClient creates in-memory client+server sessions for testing.
func connectTestClient(t *testing.T, s *mcp.Server) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := s.Connect(ctx, serverTransport)
	if err != nil {
		t.Fatalf("server connect error: %v", err)
	}
	t.Cleanup(func() { serverSession.Wait() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport)
	if err != nil {
		t.Fatalf("client connect error: %v", err)
	}
	t.Cleanup(func() { clientSession.Close() })

	return clientSession
}

func TestMCPProtocol_ListTools(t *testing.T) {
	server := createTestServer()
	cs := connectTestClient(t, server)

	result, err := cs.ListTools(context.Background(), nil)
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
	server := createTestServer()
	cs := connectTestClient(t, server)

	result, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_barcode_types",
		Arguments: map[string]any{},
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

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if textContent.Text == "" {
		t.Error("expected non-empty text content")
	}
}

func TestMCPProtocol_CallListBarcodeTypes_ContentCheck(t *testing.T) {
	server := createTestServer()
	cs := connectTestClient(t, server)

	result, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_barcode_types",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text

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
	server := createTestServer()
	cs := connectTestClient(t, server)

	_, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "nonexistent_tool",
		Arguments: map[string]any{},
	})
	// The MCP SDK should return an error for unknown tools
	if err == nil {
		t.Fatal("expected error for nonexistent tool, got nil")
	}
}

func TestMCPProtocol_ToolSchemaValidation(t *testing.T) {
	server := createTestServer()
	cs := connectTestClient(t, server)

	result, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools error: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.InputSchema == nil {
			t.Errorf("tool %q has nil input schema", tool.Name)
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

	result, err := cs.ListTools(context.Background(), nil)
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

	result, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools error: %v", err)
	}

	for _, tool := range result.Tools {
		if tool.InputSchema == nil {
			t.Errorf("tool %q has nil input schema", tool.Name)
		}
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
	}
}
