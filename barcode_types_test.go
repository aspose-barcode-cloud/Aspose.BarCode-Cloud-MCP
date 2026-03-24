package main

import (
	"testing"

	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
)

func TestMapEncodeType_ValidTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected barcode.EncodeBarcodeType
	}{
		{"QR", barcode.EncodeBarcodeTypeQR},
		{"qr", barcode.EncodeBarcodeTypeQR},
		{"Qr", barcode.EncodeBarcodeTypeQR},
		{"Code128", barcode.EncodeBarcodeTypeCode128},
		{"code128", barcode.EncodeBarcodeTypeCode128},
		{"DataMatrix", barcode.EncodeBarcodeTypeDataMatrix},
		{"EAN13", barcode.EncodeBarcodeTypeEAN13},
		{"Pdf417", barcode.EncodeBarcodeTypePdf417},
		{"Aztec", barcode.EncodeBarcodeTypeAztec},
		{"UPCA", barcode.EncodeBarcodeTypeUPCA},
		{"UPCE", barcode.EncodeBarcodeTypeUPCE},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := mapEncodeType(tt.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("mapEncodeType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapEncodeType_InvalidType(t *testing.T) {
	_, err := mapEncodeType("NonExistentType")
	if err == nil {
		t.Fatal("expected error for invalid barcode type, got nil")
	}
}

func TestMapEncodeType_EmptyString(t *testing.T) {
	_, err := mapEncodeType("")
	if err == nil {
		t.Fatal("expected error for empty string, got nil")
	}
}

func TestMapDecodeType_ValidTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected barcode.DecodeBarcodeType
	}{
		{"QR", barcode.DecodeBarcodeTypeQR},
		{"qr", barcode.DecodeBarcodeTypeQR},
		{"Code128", barcode.DecodeBarcodeTypeCode128},
		{"MostCommonlyUsed", barcode.DecodeBarcodeTypeMostCommonlyUsed},
		{"mostcommonlyused", barcode.DecodeBarcodeTypeMostCommonlyUsed},
		{"DataMatrix", barcode.DecodeBarcodeTypeDataMatrix},
		{"Pdf417", barcode.DecodeBarcodeTypePdf417},
		{"HIBCAztecLIC", barcode.DecodeBarcodeTypeHIBCAztecLIC},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := mapDecodeType(tt.input)
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("mapDecodeType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMapDecodeType_InvalidType(t *testing.T) {
	_, err := mapDecodeType("NonExistentType")
	if err == nil {
		t.Fatal("expected error for invalid barcode type, got nil")
	}
}

func TestSupportedEncodeTypes_NotEmpty(t *testing.T) {
	if len(SupportedEncodeTypes) == 0 {
		t.Fatal("SupportedEncodeTypes should not be empty")
	}
}

func TestSupportedDecodeTypes_NotEmpty(t *testing.T) {
	if len(SupportedDecodeTypes) == 0 {
		t.Fatal("SupportedDecodeTypes should not be empty")
	}
}

func TestAllEncodeTypesAreMappable(t *testing.T) {
	for _, et := range SupportedEncodeTypes {
		_, err := mapEncodeType(string(et))
		if err != nil {
			t.Errorf("SupportedEncodeType %q is not mappable: %v", et, err)
		}
	}
}

func TestAllDecodeTypesAreMappable(t *testing.T) {
	for _, dt := range SupportedDecodeTypes {
		_, err := mapDecodeType(string(dt))
		if err != nil {
			t.Errorf("SupportedDecodeType %q is not mappable: %v", dt, err)
		}
	}
}
