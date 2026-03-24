package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RecognizeBarcodeInput defines the input parameters for the recognize_barcode tool.
type RecognizeBarcodeInput struct {
	ImageData            string `json:"image_data"             jsonschema:"required,description=Base64-encoded image data (PNG\\, JPEG\\, GIF\\, TIFF\\, or BMP)"`
	BarcodeType          string `json:"barcode_type"           jsonschema:"description=Barcode type to look for (e.g. QR\\, Code128). Default: most commonly used types"`
	RecognitionMode      string `json:"recognition_mode"       jsonschema:"description=Recognition quality vs speed trade-off,enum=Fast,enum=Normal,enum=Excellent"`
	RecognitionImageKind string `json:"recognition_image_kind" jsonschema:"description=Hint about the image source for better recognition,enum=Photo,enum=ScannedDocument,enum=ClearImage"`
}

// makeRecognizeHandler creates the handler for the recognize_barcode tool.
func makeRecognizeHandler(client *AsposeClient) mcp.ToolHandlerFor[RecognizeBarcodeInput, any] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[RecognizeBarcodeInput]) (*mcp.CallToolResult, error) {
		input := params.Arguments

		// Build barcode types list
		barcodeTypes := []barcode.DecodeBarcodeType{barcode.DecodeBarcodeTypeMostCommonlyUsed}
		if input.BarcodeType != "" {
			// Support comma-separated types
			parts := strings.Split(input.BarcodeType, ",")
			barcodeTypes = make([]barcode.DecodeBarcodeType, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				dt, err := mapDecodeType(p)
				if err != nil {
					return nil, err
				}
				barcodeTypes = append(barcodeTypes, dt)
			}
		}

		body := barcode.RecognizeBase64Request{
			BarcodeTypes: barcodeTypes,
			FileBase64:   input.ImageData,
		}

		// Map optional recognition mode
		if input.RecognitionMode != "" {
			mode, err := mapRecognitionMode(input.RecognitionMode)
			if err != nil {
				return nil, err
			}
			body.RecognitionMode = mode
		}

		// Map optional recognition image kind
		if input.RecognitionImageKind != "" {
			kind, err := mapRecognitionImageKind(input.RecognitionImageKind)
			if err != nil {
				return nil, err
			}
			body.RecognitionImageKind = kind
		}

		result, _, err := client.API.RecognizeAPI.RecognizeBase64(client.AuthCtx, body)
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

// mapRecognitionMode maps a user-provided mode string to the SDK enum.
func mapRecognitionMode(s string) (barcode.RecognitionMode, error) {
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

// mapRecognitionImageKind maps a user-provided kind string to the SDK enum.
func mapRecognitionImageKind(s string) (barcode.RecognitionImageKind, error) {
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
