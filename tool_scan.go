package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ScanBarcodeInput defines the input parameters for the scan_barcode tool.
type ScanBarcodeInput struct {
	ImageData string `json:"image_data" jsonschema:"required,description=Base64-encoded image data (PNG\\, JPEG\\, GIF\\, TIFF\\, or BMP)"`
}

// makeScanHandler creates the handler for the scan_barcode tool.
func makeScanHandler(client *AsposeClient) mcp.ToolHandlerFor[ScanBarcodeInput, any] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[ScanBarcodeInput]) (*mcp.CallToolResult, error) {
		input := params.Arguments

		body := barcode.ScanBase64Request{
			FileBase64: input.ImageData,
		}

		result, _, err := client.API.ScanAPI.ScanBase64(client.AuthCtx, body)
		if err != nil {
			return nil, fmt.Errorf("Aspose API error: %w", err)
		}

		text := formatBarcodeResults(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, nil
	}
}

// formatBarcodeResults formats a BarcodeResponseList into human-readable text.
func formatBarcodeResults(result barcode.BarcodeResponseList) string {
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
