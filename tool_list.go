package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListBarcodeTypesInput is empty — the tool takes no parameters.
type ListBarcodeTypesInput struct{}

// makeListHandler creates the handler for the list_barcode_types tool.
func makeListHandler() mcp.ToolHandlerFor[ListBarcodeTypesInput, any] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ListBarcodeTypesInput]) (*mcp.CallToolResult, error) {
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
				&mcp.TextContent{Text: sb.String()},
			},
		}, nil
	}
}
