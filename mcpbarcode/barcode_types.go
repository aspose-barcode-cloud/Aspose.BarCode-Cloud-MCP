package mcpbarcode

import (
	"fmt"
	"strings"

	"github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4/barcode"
)

// SupportedEncodeTypes lists all barcode types available for generation.
var SupportedEncodeTypes = []barcode.EncodeBarcodeType{
	barcode.EncodeBarcodeTypeQR,
	barcode.EncodeBarcodeTypeAustraliaPost,
	barcode.EncodeBarcodeTypeAustralianPosteParcel,
	barcode.EncodeBarcodeTypeAztec,
	barcode.EncodeBarcodeTypeCodabar,
	barcode.EncodeBarcodeTypeCodablockF,
	barcode.EncodeBarcodeTypeCode11,
	barcode.EncodeBarcodeTypeCode128,
	barcode.EncodeBarcodeTypeCode16K,
	barcode.EncodeBarcodeTypeCode32,
	barcode.EncodeBarcodeTypeCode39,
	barcode.EncodeBarcodeTypeCode39FullASCII,
	barcode.EncodeBarcodeTypeCode93,
	barcode.EncodeBarcodeTypeDataLogic2of5,
	barcode.EncodeBarcodeTypeDataMatrix,
	barcode.EncodeBarcodeTypeDatabarExpanded,
	barcode.EncodeBarcodeTypeDatabarExpandedStacked,
	barcode.EncodeBarcodeTypeDatabarLimited,
	barcode.EncodeBarcodeTypeDatabarOmniDirectional,
	barcode.EncodeBarcodeTypeDatabarStacked,
	barcode.EncodeBarcodeTypeDatabarStackedOmniDirectional,
	barcode.EncodeBarcodeTypeDatabarTruncated,
	barcode.EncodeBarcodeTypeDeutschePostIdentcode,
	barcode.EncodeBarcodeTypeDeutschePostLeitcode,
	barcode.EncodeBarcodeTypeDotCode,
	barcode.EncodeBarcodeTypeDutchKIX,
	barcode.EncodeBarcodeTypeEAN13,
	barcode.EncodeBarcodeTypeEAN14,
	barcode.EncodeBarcodeTypeEAN8,
	barcode.EncodeBarcodeTypeGS1Aztec,
	barcode.EncodeBarcodeTypeGS1CodablockF,
	barcode.EncodeBarcodeTypeGS1Code128,
	barcode.EncodeBarcodeTypeGS1DataMatrix,
	barcode.EncodeBarcodeTypeGS1DotCode,
	barcode.EncodeBarcodeTypeGS1HanXin,
	barcode.EncodeBarcodeTypeGS1MicroPdf417,
	barcode.EncodeBarcodeTypeGS1QR,
	barcode.EncodeBarcodeTypeHanXin,
	barcode.EncodeBarcodeTypeIATA2of5,
	barcode.EncodeBarcodeTypeISBN,
	barcode.EncodeBarcodeTypeISMN,
	barcode.EncodeBarcodeTypeISSN,
	barcode.EncodeBarcodeTypeITF14,
	barcode.EncodeBarcodeTypeITF6,
	barcode.EncodeBarcodeTypeInterleaved2of5,
	barcode.EncodeBarcodeTypeItalianPost25,
	barcode.EncodeBarcodeTypeMSI,
	barcode.EncodeBarcodeTypeMacroPdf417,
	barcode.EncodeBarcodeTypeMailmark,
	barcode.EncodeBarcodeTypeMatrix2of5,
	barcode.EncodeBarcodeTypeMaxiCode,
	barcode.EncodeBarcodeTypeMicroPdf417,
	barcode.EncodeBarcodeTypeMicroQR,
	barcode.EncodeBarcodeTypeOPC,
	barcode.EncodeBarcodeTypeOneCode,
	barcode.EncodeBarcodeTypePZN,
	barcode.EncodeBarcodeTypePatchCode,
	barcode.EncodeBarcodeTypePdf417,
	barcode.EncodeBarcodeTypePharmacode,
	barcode.EncodeBarcodeTypePlanet,
	barcode.EncodeBarcodeTypePostnet,
	barcode.EncodeBarcodeTypeRM4SCC,
	barcode.EncodeBarcodeTypeRectMicroQR,
	barcode.EncodeBarcodeTypeSCC14,
	barcode.EncodeBarcodeTypeSSCC18,
	barcode.EncodeBarcodeTypeSingaporePost,
	barcode.EncodeBarcodeTypeStandard2of5,
	barcode.EncodeBarcodeTypeSwissPostParcel,
	barcode.EncodeBarcodeTypeUPCA,
	barcode.EncodeBarcodeTypeUPCE,
	barcode.EncodeBarcodeTypeUpcaGs1Code128Coupon,
	barcode.EncodeBarcodeTypeUpcaGs1DatabarCoupon,
	barcode.EncodeBarcodeTypeVIN,
}

// SupportedDecodeTypes lists all barcode types available for recognition.
var SupportedDecodeTypes = []barcode.DecodeBarcodeType{
	barcode.DecodeBarcodeTypeQR,
	barcode.DecodeBarcodeTypeMostCommonlyUsed,
	barcode.DecodeBarcodeTypeAustraliaPost,
	barcode.DecodeBarcodeTypeAustralianPosteParcel,
	barcode.DecodeBarcodeTypeAztec,
	barcode.DecodeBarcodeTypeCodabar,
	barcode.DecodeBarcodeTypeCodablockF,
	barcode.DecodeBarcodeTypeCode11,
	barcode.DecodeBarcodeTypeCode128,
	barcode.DecodeBarcodeTypeCode16K,
	barcode.DecodeBarcodeTypeCode32,
	barcode.DecodeBarcodeTypeCode39,
	barcode.DecodeBarcodeTypeCode39FullASCII,
	barcode.DecodeBarcodeTypeCode93,
	barcode.DecodeBarcodeTypeCompactPdf417,
	barcode.DecodeBarcodeTypeDataLogic2of5,
	barcode.DecodeBarcodeTypeDataMatrix,
	barcode.DecodeBarcodeTypeDatabarExpanded,
	barcode.DecodeBarcodeTypeDatabarExpandedStacked,
	barcode.DecodeBarcodeTypeDatabarLimited,
	barcode.DecodeBarcodeTypeDatabarOmniDirectional,
	barcode.DecodeBarcodeTypeDatabarStacked,
	barcode.DecodeBarcodeTypeDatabarStackedOmniDirectional,
	barcode.DecodeBarcodeTypeDatabarTruncated,
	barcode.DecodeBarcodeTypeDeutschePostIdentcode,
	barcode.DecodeBarcodeTypeDeutschePostLeitcode,
	barcode.DecodeBarcodeTypeDotCode,
	barcode.DecodeBarcodeTypeDutchKIX,
	barcode.DecodeBarcodeTypeEAN13,
	barcode.DecodeBarcodeTypeEAN14,
	barcode.DecodeBarcodeTypeEAN8,
	barcode.DecodeBarcodeTypeGS1Aztec,
	barcode.DecodeBarcodeTypeGS1Code128,
	barcode.DecodeBarcodeTypeGS1CompositeBar,
	barcode.DecodeBarcodeTypeGS1DataMatrix,
	barcode.DecodeBarcodeTypeGS1DotCode,
	barcode.DecodeBarcodeTypeGS1HanXin,
	barcode.DecodeBarcodeTypeGS1MicroPdf417,
	barcode.DecodeBarcodeTypeGS1QR,
	barcode.DecodeBarcodeTypeHanXin,
	barcode.DecodeBarcodeTypeHIBCAztecLIC,
	barcode.DecodeBarcodeTypeHIBCAztecPAS,
	barcode.DecodeBarcodeTypeHIBCCode128LIC,
	barcode.DecodeBarcodeTypeHIBCCode128PAS,
	barcode.DecodeBarcodeTypeHIBCCode39LIC,
	barcode.DecodeBarcodeTypeHIBCCode39PAS,
	barcode.DecodeBarcodeTypeHIBCDataMatrixLIC,
	barcode.DecodeBarcodeTypeHIBCDataMatrixPAS,
	barcode.DecodeBarcodeTypeHIBCQRLIC,
	barcode.DecodeBarcodeTypeHIBCQRPAS,
	barcode.DecodeBarcodeTypeIATA2of5,
	barcode.DecodeBarcodeTypeISBN,
	barcode.DecodeBarcodeTypeISMN,
	barcode.DecodeBarcodeTypeISSN,
	barcode.DecodeBarcodeTypeITF14,
	barcode.DecodeBarcodeTypeITF6,
	barcode.DecodeBarcodeTypeInterleaved2of5,
	barcode.DecodeBarcodeTypeItalianPost25,
	barcode.DecodeBarcodeTypeMacroPdf417,
	barcode.DecodeBarcodeTypeMailmark,
	barcode.DecodeBarcodeTypeMatrix2of5,
	barcode.DecodeBarcodeTypeMaxiCode,
	barcode.DecodeBarcodeTypeMicrE13B,
	barcode.DecodeBarcodeTypeMicroPdf417,
	barcode.DecodeBarcodeTypeMicroQR,
	barcode.DecodeBarcodeTypeMSI,
	barcode.DecodeBarcodeTypeOneCode,
	barcode.DecodeBarcodeTypeOPC,
	barcode.DecodeBarcodeTypePatchCode,
	barcode.DecodeBarcodeTypePdf417,
	barcode.DecodeBarcodeTypePharmacode,
	barcode.DecodeBarcodeTypePlanet,
	barcode.DecodeBarcodeTypePostnet,
	barcode.DecodeBarcodeTypePZN,
	barcode.DecodeBarcodeTypeRectMicroQR,
	barcode.DecodeBarcodeTypeRM4SCC,
	barcode.DecodeBarcodeTypeSCC14,
	barcode.DecodeBarcodeTypeSSCC18,
	barcode.DecodeBarcodeTypeStandard2of5,
	barcode.DecodeBarcodeTypeSupplement,
	barcode.DecodeBarcodeTypeSwissPostParcel,
	barcode.DecodeBarcodeTypeUPCA,
	barcode.DecodeBarcodeTypeUPCE,
	barcode.DecodeBarcodeTypeVIN,
}

// encodeTypeMap maps lowercase barcode type names to EncodeBarcodeType constants.
var encodeTypeMap map[string]barcode.EncodeBarcodeType

// decodeTypeMap maps lowercase barcode type names to DecodeBarcodeType constants.
var decodeTypeMap map[string]barcode.DecodeBarcodeType

func init() {
	encodeTypeMap = make(map[string]barcode.EncodeBarcodeType, len(SupportedEncodeTypes))
	for _, t := range SupportedEncodeTypes {
		encodeTypeMap[strings.ToLower(string(t))] = t
	}

	decodeTypeMap = make(map[string]barcode.DecodeBarcodeType, len(SupportedDecodeTypes))
	for _, t := range SupportedDecodeTypes {
		decodeTypeMap[strings.ToLower(string(t))] = t
	}
}

// MapEncodeType performs case-insensitive lookup of an EncodeBarcodeType.
func MapEncodeType(s string) (barcode.EncodeBarcodeType, error) {
	if t, ok := encodeTypeMap[strings.ToLower(s)]; ok {
		return t, nil
	}
	return "", fmt.Errorf("unsupported barcode type for generation: %q", s)
}

// MapDecodeType performs case-insensitive lookup of a DecodeBarcodeType.
func MapDecodeType(s string) (barcode.DecodeBarcodeType, error) {
	if t, ok := decodeTypeMap[strings.ToLower(s)]; ok {
		return t, nil
	}
	return "", fmt.Errorf("unsupported barcode type for recognition: %q", s)
}
