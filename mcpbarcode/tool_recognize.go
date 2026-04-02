package mcpbarcode

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RecognizeBarcodeInput defines the input parameters for the recognize_barcode tool.
type RecognizeBarcodeInput struct {
	ImagePath            string `json:"image_path"                             jsonschema:"Relative path to image file in the mounted data directory (PNG, JPEG, GIF, TIFF, or BMP). Must be relative to the mount root, e.g. 'photo.png' or 'subdir/photo.png'"`
	BarcodeType          string `json:"barcode_type,omitempty"                 jsonschema:"Barcode type to look for (e.g. QR, Code128). Default: most commonly used types"`
	RecognitionMode      string `json:"recognition_mode,omitempty"             jsonschema:"Recognition quality vs speed trade-off: Fast, Normal, or Excellent"`
	RecognitionImageKind string `json:"recognition_image_kind,omitempty"       jsonschema:"Hint about the image source for better recognition: Photo, ScannedDocument, or ClearImage"`
}

// MakeRecognizeHandler creates the handler for the recognize_barcode tool.
func MakeRecognizeHandler(client *AsposeClient, mount *MountConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input RecognizeBarcodeInput
		if err := request.BindArguments(&input); err != nil {
			return toolError("invalid arguments: %v", err)
		}

		if input.ImagePath == "" {
			return toolError("'image_path' is required")
		}
		if err := ValidateImageExtension(input.ImagePath); err != nil {
			return toolError("%v", err)
		}

		// Build barcode types list
		barcodeTypes := []barcode.DecodeBarcodeType{barcode.DecodeBarcodeTypeMostCommonlyUsed}
		if input.BarcodeType != "" {
			parts := strings.Split(input.BarcodeType, ",")
			barcodeTypes = make([]barcode.DecodeBarcodeType, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				dt, err := MapDecodeType(p)
				if err != nil {
					return toolError("%v", err)
				}
				barcodeTypes = append(barcodeTypes, dt)
			}
		}

		// Read file from mount and encode to base64
		file, err := mount.OpenFile(input.ImagePath)
		if err != nil {
			return toolError("failed to open image: %v", err)
		}
		defer file.Close()

		imageBytes, err := io.ReadAll(file)
		if err != nil {
			return toolError("failed to read image: %v", err)
		}
		fileBase64 := base64.StdEncoding.EncodeToString(imageBytes)

		body := barcode.RecognizeBase64Request{
			BarcodeTypes: barcodeTypes,
			FileBase64:   fileBase64,
		}
		if input.RecognitionMode != "" {
			mode, err := MapRecognitionMode(input.RecognitionMode)
			if err != nil {
				return toolError("%v", err)
			}
			body.RecognitionMode = mode
		}
		if input.RecognitionImageKind != "" {
			kind, err := MapRecognitionImageKind(input.RecognitionImageKind)
			if err != nil {
				return toolError("%v", err)
			}
			body.RecognitionImageKind = kind
		}

		result, _, err := client.API.RecognizeAPI.RecognizeBase64(
			client.AuthCtx,
			body,
		)
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

// MapRecognitionMode maps a user-provided mode string to the SDK enum.
func MapRecognitionMode(s string) (barcode.RecognitionMode, error) {
	switch strings.ToLower(s) {
	case "fast":
		return barcode.RecognitionModeFast, nil
	case "normal":
		return barcode.RecognitionModeNormal, nil
	case "excellent":
		return barcode.RecognitionModeExcellent, nil
	default:
		return "", fmt.Errorf("unsupported recognition mode: %q (supported: Fast, Normal, Excellent)", s)
	}
}

// MapRecognitionImageKind maps a user-provided kind string to the SDK enum.
func MapRecognitionImageKind(s string) (barcode.RecognitionImageKind, error) {
	switch strings.ToLower(s) {
	case "photo":
		return barcode.RecognitionImageKindPhoto, nil
	case "scanneddocument":
		return barcode.RecognitionImageKindScannedDocument, nil
	case "clearimage":
		return barcode.RecognitionImageKindClearImage, nil
	default:
		return "", fmt.Errorf("unsupported recognition image kind: %q (supported: Photo, ScannedDocument, ClearImage)", s)
	}
}
