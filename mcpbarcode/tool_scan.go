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
	ImagePath string `json:"image_path" jsonschema:"Relative path to image file in the mounted data directory (PNG, JPEG, GIF, TIFF, or BMP). Must be relative to the mount root, e.g. 'photo.png' or 'subdir/photo.png'"`
}

// MakeScanHandler creates the handler for the scan_barcode tool.
func MakeScanHandler(client *AsposeClient, mount *MountConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input ScanBarcodeInput
		if err := request.BindArguments(&input); err != nil {
			return toolError("invalid arguments: %v", err)
		}

		if input.ImagePath == "" {
			return toolError("'image_path' is required")
		}
		if err := ValidateImageExtension(input.ImagePath); err != nil {
			return toolError("%v", err)
		}

		file, err := mount.OpenFile(input.ImagePath)
		if err != nil {
			return toolError("failed to open image: %v", err)
		}
		defer file.Close()

		result, _, err := client.API.ScanAPI.ScanMultipart(client.AuthCtx, file)
		if err != nil {
			return toolError("Aspose API error: %v", err)
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
