package mcpbarcode

import (
	"context"
	"fmt"
	"strings"

	"github.com/antihax/optional"
	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RecognizeBarcodeInput defines the input parameters for the recognize_barcode tool.
type RecognizeBarcodeInput struct {
	ImagePath            string `json:"image_path"                             jsonschema:"description=Path to image file in the mounted data directory (PNG, JPEG, GIF, TIFF, or BMP)"`
	BarcodeType          string `json:"barcode_type,omitempty"                 jsonschema:"description=Barcode type to look for (e.g. QR, Code128). Default: most commonly used types"`
	RecognitionMode      string `json:"recognition_mode,omitempty"             jsonschema:"description=Recognition quality vs speed trade-off: Fast, Normal, or Excellent"`
	RecognitionImageKind string `json:"recognition_image_kind,omitempty"       jsonschema:"description=Hint about the image source for better recognition: Photo, ScannedDocument, or ClearImage"`
}

// MakeRecognizeHandler creates the handler for the recognize_barcode tool.
func MakeRecognizeHandler(client *AsposeClient, mount *MountConfig) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var input RecognizeBarcodeInput
		if err := request.BindArguments(&input); err != nil {
			return nil, fmt.Errorf("invalid arguments: %w", err)
		}

		if input.ImagePath == "" {
			return nil, fmt.Errorf("'image_path' is required")
		}
		if err := ValidateImageExtension(input.ImagePath); err != nil {
			return nil, err
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
					return nil, err
				}
				barcodeTypes = append(barcodeTypes, dt)
			}
		}

		// Open file from mount
		file, err := mount.OpenFile(input.ImagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open image: %w", err)
		}
		defer file.Close()

		// RecognizeMultipart requires a single DecodeBarcodeType
		recognizeType := barcode.DecodeBarcodeTypeMostCommonlyUsed
		if len(barcodeTypes) == 1 {
			recognizeType = barcodeTypes[0]
		}

		opts := &barcode.RecognizeAPIRecognizeMultipartOpts{}
		if input.RecognitionMode != "" {
			mode, err := MapRecognitionMode(input.RecognitionMode)
			if err != nil {
				return nil, err
			}
			opts.RecognitionMode = optional.NewInterface(mode)
		}
		if input.RecognitionImageKind != "" {
			kind, err := MapRecognitionImageKind(input.RecognitionImageKind)
			if err != nil {
				return nil, err
			}
			opts.RecognitionImageKind = optional.NewInterface(kind)
		}

		result, _, err := client.API.RecognizeAPI.RecognizeMultipart(
			client.AuthCtx,
			recognizeType,
			file,
			opts,
		)
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
