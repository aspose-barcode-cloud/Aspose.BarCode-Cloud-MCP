package main

import (
	"testing"

	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
)

func TestMapImageFormat_ValidFormats(t *testing.T) {
	tests := []struct {
		input    string
		expected barcode.BarcodeImageFormat
	}{
		{"PNG", barcode.BarcodeImageFormatPng},
		{"png", barcode.BarcodeImageFormatPng},
		{"Png", barcode.BarcodeImageFormatPng},
		{"JPEG", barcode.BarcodeImageFormatJpeg},
		{"jpeg", barcode.BarcodeImageFormatJpeg},
		{"JPG", barcode.BarcodeImageFormatJpeg},
		{"jpg", barcode.BarcodeImageFormatJpeg},
		{"SVG", barcode.BarcodeImageFormatSvg},
		{"svg", barcode.BarcodeImageFormatSvg},
		{"TIFF", barcode.BarcodeImageFormatTiff},
		{"tiff", barcode.BarcodeImageFormatTiff},
		{"GIF", barcode.BarcodeImageFormatGif},
		{"gif", barcode.BarcodeImageFormatGif},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := mapImageFormat(tt.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("mapImageFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapImageFormat_InvalidFormat(t *testing.T) {
	invalidFormats := []string{"BMP", "WEBP", "", "invalid"}
	for _, f := range invalidFormats {
		t.Run(f, func(t *testing.T) {
			_, err := mapImageFormat(f)
			if err == nil {
				t.Fatalf("expected error for format %q, got nil", f)
			}
		})
	}
}

func TestMapCodeLocation_ValidLocations(t *testing.T) {
	tests := []struct {
		input    string
		expected barcode.CodeLocation
	}{
		{"Below", barcode.CodeLocationBelow},
		{"below", barcode.CodeLocationBelow},
		{"BELOW", barcode.CodeLocationBelow},
		{"Above", barcode.CodeLocationAbove},
		{"above", barcode.CodeLocationAbove},
		{"None", barcode.CodeLocationNone},
		{"none", barcode.CodeLocationNone},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := mapCodeLocation(tt.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("mapCodeLocation(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapCodeLocation_InvalidLocation(t *testing.T) {
	invalidLocations := []string{"Left", "Right", "", "invalid"}
	for _, loc := range invalidLocations {
		t.Run(loc, func(t *testing.T) {
			_, err := mapCodeLocation(loc)
			if err == nil {
				t.Fatalf("expected error for location %q, got nil", loc)
			}
		})
	}
}

func TestMimeTypeForFormat(t *testing.T) {
	tests := []struct {
		input    barcode.BarcodeImageFormat
		expected string
	}{
		{barcode.BarcodeImageFormatPng, "image/png"},
		{barcode.BarcodeImageFormatJpeg, "image/jpeg"},
		{barcode.BarcodeImageFormatGif, "image/gif"},
		{barcode.BarcodeImageFormatTiff, "image/tiff"},
		{barcode.BarcodeImageFormatSvg, "image/png"}, // SVG falls through to default
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := mimeTypeForFormat(tt.input)
			if result != tt.expected {
				t.Errorf("mimeTypeForFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
