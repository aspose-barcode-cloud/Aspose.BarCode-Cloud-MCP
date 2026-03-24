package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/antihax/optional"
	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GenerateBarcodeInput defines the input parameters for the generate_barcode tool.
type GenerateBarcodeInput struct {
	BarcodeType     string  `json:"barcode_type"                jsonschema:"Barcode symbology to generate (e.g. QR, Code128, DataMatrix, EAN13, PDF417)"`
	Data            string  `json:"data"                        jsonschema:"Data to encode in the barcode"`
	ImageFormat     string  `json:"image_format,omitempty"      jsonschema:"Output image format: PNG, JPEG, SVG, GIF, or TIFF"`
	TextLocation    string  `json:"text_location,omitempty"     jsonschema:"Where to display human-readable text on the barcode: Below, Above, or None"`
	ForegroundColor string  `json:"foreground_color,omitempty"  jsonschema:"Foreground color as color name (e.g. Black) or #AARRGGBB hex"`
	BackgroundColor string  `json:"background_color,omitempty"  jsonschema:"Background color as color name (e.g. White) or #AARRGGBB hex"`
	Resolution      float64 `json:"resolution,omitempty"        jsonschema:"Image resolution in DPI (1-100000)"`
	RotationAngle   float64 `json:"rotation_angle,omitempty"    jsonschema:"Rotation angle: 0, 90, 180, or 270 degrees"`
	ImageWidth      float64 `json:"image_width,omitempty"       jsonschema:"Image width in pixels"`
	ImageHeight     float64 `json:"image_height,omitempty"      jsonschema:"Image height in pixels"`
}

// makeGenerateHandler creates the handler for the generate_barcode tool.
func makeGenerateHandler(client *AsposeClient) mcp.ToolHandlerFor[GenerateBarcodeInput, any] {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[GenerateBarcodeInput]) (*mcp.CallToolResult, error) {
		input := params.Arguments

		// Map barcode type
		barcodeType, err := mapEncodeType(input.BarcodeType)
		if err != nil {
			return nil, err
		}

		// Build optional parameters
		opts := &barcode.GenerateAPIGenerateOpts{}

		// Image format
		imageFormat := barcode.BarcodeImageFormatPng
		if input.ImageFormat != "" {
			mapped, err := mapImageFormat(input.ImageFormat)
			if err != nil {
				return nil, err
			}
			imageFormat = mapped
		}
		opts.ImageFormat = optional.NewInterface(imageFormat)

		// Text location
		if input.TextLocation != "" {
			loc, err := mapCodeLocation(input.TextLocation)
			if err != nil {
				return nil, err
			}
			opts.TextLocation = optional.NewInterface(loc)
		}

		// Colors
		if input.ForegroundColor != "" {
			opts.ForegroundColor = optional.NewString(input.ForegroundColor)
		}
		if input.BackgroundColor != "" {
			opts.BackgroundColor = optional.NewString(input.BackgroundColor)
		}

		// Numeric options
		if input.Resolution != 0 {
			opts.Resolution = optional.NewFloat32(float32(input.Resolution))
		}
		if input.RotationAngle != 0 {
			opts.RotationAngle = optional.NewInt32(int32(input.RotationAngle))
		}
		if input.ImageWidth != 0 {
			opts.ImageWidth = optional.NewFloat32(float32(input.ImageWidth))
		}
		if input.ImageHeight != 0 {
			opts.ImageHeight = optional.NewFloat32(float32(input.ImageHeight))
		}

		// Call the Aspose API
		imageBytes, _, err := client.API.GenerateAPI.Generate(
			client.AuthCtx,
			barcodeType,
			input.Data,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("Aspose API error: %w", err)
		}

		// Return SVG as text, raster formats as base64 image
		if imageFormat == barcode.BarcodeImageFormatSvg {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(imageBytes)},
				},
			}, nil
		}

		mimeType := mimeTypeForFormat(imageFormat)
		encoded := base64.StdEncoding.EncodeToString(imageBytes)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.ImageContent{
					Data:     []byte(encoded),
					MIMEType: mimeType,
				},
			},
		}, nil
	}
}

// mapImageFormat maps a user-provided format string to the SDK enum.
func mapImageFormat(s string) (barcode.BarcodeImageFormat, error) {
	switch strings.ToUpper(s) {
	case "PNG":
		return barcode.BarcodeImageFormatPng, nil
	case "JPEG", "JPG":
		return barcode.BarcodeImageFormatJpeg, nil
	case "SVG":
		return barcode.BarcodeImageFormatSvg, nil
	case "TIFF":
		return barcode.BarcodeImageFormatTiff, nil
	case "GIF":
		return barcode.BarcodeImageFormatGif, nil
	default:
		return "", fmt.Errorf("unsupported image format: %q (supported: PNG, JPEG, SVG, GIF, TIFF)", s)
	}
}

// mapCodeLocation maps a user-provided text location to the SDK enum.
func mapCodeLocation(s string) (barcode.CodeLocation, error) {
	switch strings.ToLower(s) {
	case "below":
		return barcode.CodeLocationBelow, nil
	case "above":
		return barcode.CodeLocationAbove, nil
	case "none":
		return barcode.CodeLocationNone, nil
	default:
		return "", fmt.Errorf("unsupported text location: %q (supported: Below, Above, None)", s)
	}
}

// mimeTypeForFormat returns the MIME type for a given barcode image format.
func mimeTypeForFormat(f barcode.BarcodeImageFormat) string {
	switch f {
	case barcode.BarcodeImageFormatPng:
		return "image/png"
	case barcode.BarcodeImageFormatJpeg:
		return "image/jpeg"
	case barcode.BarcodeImageFormatGif:
		return "image/gif"
	case barcode.BarcodeImageFormatTiff:
		return "image/tiff"
	default:
		return "image/png"
	}
}
