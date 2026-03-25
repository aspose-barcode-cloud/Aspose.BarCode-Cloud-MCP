package mcpbarcode

import (
	"context"
	"fmt"
	"strings"

	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ScanBarcodeInput defines the input parameters for the scan_barcode tool.
type ScanBarcodeInput struct {
	ImageData string `json:"image_data" jsonschema:"description=Base64-encoded image data (PNG, JPEG, GIF, TIFF, or BMP)"`
}

// MakeScanHandler creates the handler for the scan_barcode tool.
func MakeScanHandler(client *AsposeClient) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input ScanBarcodeInput
		if err := request.BindArguments(&input); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		body := barcode.ScanBase64Request{
			FileBase64: input.ImageData,
		}

		result, _, err := client.API.ScanAPI.ScanBase64(client.AuthCtx, body)
		if err != nil {
			return nil, fmt.Errorf("Aspose API error: %w", err)
		}

		text := FormatBarcodeResults(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: text},
			},
		}, nil
	}
}

// FormatBarcodeResults formats a BarcodeResponseList into human-readable text.
func FormatBarcodeResults(result barcode.BarcodeResponseList) string {
	if len(result.Barcodes) == 0 {
		return "No barcodes detected in the image."
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d barcode(s):\n", len(result.Barcodes))

	for i, b := range result.Barcodes {
		fmt.Fprintf(&sb, "\n%d. Type: %s\n", i+1, b.Type)
		fmt.Fprintf(&sb, "   Value: %s\n", b.BarcodeValue)
		if b.Checksum != "" {
			fmt.Fprintf(&sb, "   Checksum: %s\n", b.Checksum)
		}
	}

	return sb.String()
}
