package main

import (
	"strings"
	"testing"

	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
)

func TestFormatBarcodeResults_NoBarcodes(t *testing.T) {
	result := formatBarcodeResults(barcode.BarcodeResponseList{
		Barcodes: []barcode.BarcodeResponse{},
	})

	expected := "No barcodes detected in the image."
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestFormatBarcodeResults_NilBarcodes(t *testing.T) {
	result := formatBarcodeResults(barcode.BarcodeResponseList{
		Barcodes: nil,
	})

	expected := "No barcodes detected in the image."
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestFormatBarcodeResults_SingleBarcode(t *testing.T) {
	result := formatBarcodeResults(barcode.BarcodeResponseList{
		Barcodes: []barcode.BarcodeResponse{
			{
				Type:         "QR",
				BarcodeValue: "Hello World",
			},
		},
	})

	if !strings.Contains(result, "Found 1 barcode(s):") {
		t.Errorf("expected 'Found 1 barcode(s):' in result, got: %q", result)
	}
	if !strings.Contains(result, "Type: QR") {
		t.Errorf("expected 'Type: QR' in result, got: %q", result)
	}
	if !strings.Contains(result, "Value: Hello World") {
		t.Errorf("expected 'Value: Hello World' in result, got: %q", result)
	}
}

func TestFormatBarcodeResults_MultipleBarcodes(t *testing.T) {
	result := formatBarcodeResults(barcode.BarcodeResponseList{
		Barcodes: []barcode.BarcodeResponse{
			{
				Type:         "QR",
				BarcodeValue: "First",
			},
			{
				Type:         "Code128",
				BarcodeValue: "Second",
				Checksum:     "ABC123",
			},
		},
	})

	if !strings.Contains(result, "Found 2 barcode(s):") {
		t.Errorf("expected 'Found 2 barcode(s):' in result, got: %q", result)
	}
	if !strings.Contains(result, "1. Type: QR") {
		t.Errorf("expected '1. Type: QR' in result, got: %q", result)
	}
	if !strings.Contains(result, "2. Type: Code128") {
		t.Errorf("expected '2. Type: Code128' in result, got: %q", result)
	}
	if !strings.Contains(result, "Checksum: ABC123") {
		t.Errorf("expected 'Checksum: ABC123' in result, got: %q", result)
	}
}

func TestFormatBarcodeResults_WithChecksum(t *testing.T) {
	result := formatBarcodeResults(barcode.BarcodeResponseList{
		Barcodes: []barcode.BarcodeResponse{
			{
				Type:         "EAN13",
				BarcodeValue: "5901234123457",
				Checksum:     "7",
			},
		},
	})

	if !strings.Contains(result, "Checksum: 7") {
		t.Errorf("expected checksum in result, got: %q", result)
	}
}

func TestFormatBarcodeResults_WithoutChecksum(t *testing.T) {
	result := formatBarcodeResults(barcode.BarcodeResponseList{
		Barcodes: []barcode.BarcodeResponse{
			{
				Type:         "QR",
				BarcodeValue: "test",
				Checksum:     "",
			},
		},
	})

	if strings.Contains(result, "Checksum:") {
		t.Errorf("did not expect checksum in result, got: %q", result)
	}
}
