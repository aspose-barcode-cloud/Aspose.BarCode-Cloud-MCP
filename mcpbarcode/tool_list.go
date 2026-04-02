package mcpbarcode

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListBarcodeTypesInput is empty — the tool takes no parameters.
type ListBarcodeTypesInput struct{}

// MakeListHandler creates the handler for the list_barcode_types tool.
func MakeListHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var sb strings.Builder

		sb.WriteString("Supported barcode types for GENERATION:\n")
		for _, t := range SupportedEncodeTypes {
			fmt.Fprintf(&sb, "- %s\n", string(t))
		}

		sb.WriteString("\nSupported barcode types for RECOGNITION:\n")
		for _, t := range SupportedDecodeTypes {
			fmt.Fprintf(&sb, "- %s\n", string(t))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: sb.String()},
			},
		}, nil
	}
}
