package main

import (
	"testing"

	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
)

func TestMapRecognitionMode_ValidModes(t *testing.T) {
	tests := []struct {
		input    string
		expected barcode.RecognitionMode
	}{
		{"Fast", barcode.RecognitionModeFast},
		{"fast", barcode.RecognitionModeFast},
		{"FAST", barcode.RecognitionModeFast},
		{"Normal", barcode.RecognitionModeNormal},
		{"normal", barcode.RecognitionModeNormal},
		{"Excellent", barcode.RecognitionModeExcellent},
		{"excellent", barcode.RecognitionModeExcellent},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := mapRecognitionMode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("mapRecognitionMode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapRecognitionMode_InvalidMode(t *testing.T) {
	invalidModes := []string{"SuperFast", "", "invalid", "Quick"}
	for _, m := range invalidModes {
		t.Run(m, func(t *testing.T) {
			_, err := mapRecognitionMode(m)
			if err == nil {
				t.Fatalf("expected error for mode %q, got nil", m)
			}
		})
	}
}

func TestMapRecognitionImageKind_ValidKinds(t *testing.T) {
	tests := []struct {
		input    string
		expected barcode.RecognitionImageKind
	}{
		{"Photo", barcode.RecognitionImageKindPhoto},
		{"photo", barcode.RecognitionImageKindPhoto},
		{"ScannedDocument", barcode.RecognitionImageKindScannedDocument},
		{"scanneddocument", barcode.RecognitionImageKindScannedDocument},
		{"ClearImage", barcode.RecognitionImageKindClearImage},
		{"clearimage", barcode.RecognitionImageKindClearImage},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := mapRecognitionImageKind(tt.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("mapRecognitionImageKind(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapRecognitionImageKind_InvalidKind(t *testing.T) {
	invalidKinds := []string{"Blurry", "", "invalid", "HighRes"}
	for _, k := range invalidKinds {
		t.Run(k, func(t *testing.T) {
			_, err := mapRecognitionImageKind(k)
			if err == nil {
				t.Fatalf("expected error for kind %q, got nil", k)
			}
		})
	}
}
