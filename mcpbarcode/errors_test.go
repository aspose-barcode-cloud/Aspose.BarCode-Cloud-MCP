package mcpbarcode

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestToolError_ReturnsIsErrorTrue(t *testing.T) {
	result, err := toolError("something went wrong")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !result.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestToolError_ContainsMessage(t *testing.T) {
	result, _ := toolError("bad input: %q", "test-value")

	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if textContent.Type != "text" {
		t.Errorf("expected Type %q, got %q", "text", textContent.Type)
	}

	expected := `bad input: "test-value"`
	if textContent.Text != expected {
		t.Errorf("expected Text %q, got %q", expected, textContent.Text)
	}
}
