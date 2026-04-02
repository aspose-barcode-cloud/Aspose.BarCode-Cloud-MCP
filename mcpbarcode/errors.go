package mcpbarcode

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// toolError returns a CallToolResult with IsError set to true.
// Use this for user-facing errors (bad input, API failures, file issues).
// Protocol errors (nil, error) should only be used for truly unexpected failures.
func toolError(format string, args ...any) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(format, args...),
			},
		},
	}, nil
}
