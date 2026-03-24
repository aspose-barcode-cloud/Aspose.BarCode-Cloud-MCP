package main

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMakeListHandler_ReturnsResult(t *testing.T) {
	handler := makeListHandler()

	params := &mcp.CallToolParamsFor[ListBarcodeTypesInput]{
		Arguments: ListBarcodeTypesInput{},
	}

	result, err := handler(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(result.Content))
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	text := textContent.Text

	// Verify generation types section
	if !strings.Contains(text, "Supported barcode types for GENERATION:") {
		t.Error("missing generation types header")
	}

	// Verify recognition types section
	if !strings.Contains(text, "Supported barcode types for RECOGNITION:") {
		t.Error("missing recognition types header")
	}

	// Verify some well-known types are present
	expectedTypes := []string{"QR", "Code128", "DataMatrix", "EAN13", "Pdf417"}
	for _, et := range expectedTypes {
		if !strings.Contains(text, "- "+et) {
			t.Errorf("expected type %q not found in output", et)
		}
	}
}

func TestMakeListHandler_ContainsAllEncodeTypes(t *testing.T) {
	handler := makeListHandler()

	params := &mcp.CallToolParamsFor[ListBarcodeTypesInput]{
		Arguments: ListBarcodeTypesInput{},
	}

	result, err := handler(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text

	for _, et := range SupportedEncodeTypes {
		if !strings.Contains(text, string(et)) {
			t.Errorf("encode type %q not found in list output", et)
		}
	}
}

func TestMakeListHandler_ContainsAllDecodeTypes(t *testing.T) {
	handler := makeListHandler()

	params := &mcp.CallToolParamsFor[ListBarcodeTypesInput]{
		Arguments: ListBarcodeTypesInput{},
	}

	result, err := handler(context.Background(), nil, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text

	for _, dt := range SupportedDecodeTypes {
		if !strings.Contains(text, string(dt)) {
			t.Errorf("decode type %q not found in list output", dt)
		}
	}
}
