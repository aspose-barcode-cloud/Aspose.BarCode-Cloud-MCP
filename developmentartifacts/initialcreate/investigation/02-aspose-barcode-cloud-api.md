# Aspose Barcode Cloud API v4.0

## Overview

Cloud REST API for barcode generation, recognition, and scanning. Supports 60+ barcode symbologies (1D, 2D, Postal).

- **Base URL**: `https://api.aspose.cloud/v4.0`
- **Auth**: OAuth 2.0 JWT token obtained from `https://id.aspose.cloud/connect/token`
- **SDK version**: v4.2603.0
- **Go module**: `github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4`

## Authentication

Requires Client ID and Client Secret from https://dashboard.aspose.cloud/applications.
Free quota available. JWT token is obtained automatically by the SDK.

```go
jwtConf := jwt.NewConfig("ClientId", "ClientSecret")
authCtx := context.WithValue(context.Background(),
    barcode.ContextJWT,
    jwtConf.TokenSource(context.Background()))
client := barcode.NewAPIClient(barcode.NewConfiguration())
```

## API Endpoints (3 groups, 9 endpoints)

### Generate API

| Method | HTTP | Description |
|--------|------|-------------|
| `Generate` | `GET /barcode/generate/{barcodeType}` | Generate via query params |
| `GenerateBody` | `POST /barcode/generate-body` | Generate via JSON/XML body |
| `GenerateMultipart` | `POST /barcode/generate-multipart` | Generate via multipart form |

**Key Parameters:**
- `barcodeType` (EncodeBarcodeType) - required, e.g. QR, Code128, DataMatrix
- `data` (string) - required, content to encode
- `dataType` (EncodeDataType) - StringData, Base64Bytes, HexBytes
- `imageFormat` (BarcodeImageFormat) - PNG, JPEG, SVG, TIFF, GIF
- `textLocation` (CodeLocation) - Below, Above, None
- `foregroundColor`, `backgroundColor` - color name or #AARRGGBB hex
- `units` (GraphicsUnit) - Pixel, Point, Inch, Millimeter
- `resolution` (float) - DPI, 1-100000
- `imageHeight`, `imageWidth` (float)
- `rotationAngle` (int) - 0, 90, 180, 270

**Returns:** image bytes (PNG/JPEG/SVG/etc.)

### Recognize API

| Method | HTTP | Description |
|--------|------|-------------|
| `Recognize` | `GET /barcode/recognize` | Recognize from URL |
| `RecognizeBase64` | `POST /barcode/recognize-body` | Recognize from base64 body |
| `RecognizeMultipart` | `POST /barcode/recognize-multipart` | Recognize from uploaded file |

**Key Parameters:**
- `barcodeType` (DecodeBarcodeType) - specific type(s) to look for
- `recognitionMode` - Fast, Normal, Excellent
- `recognitionImageKind` - Photo, ScannedDocument, ClearImage

### Scan API

| Method | HTTP | Description |
|--------|------|-------------|
| `Scan` | `GET /barcode/scan` | Auto-detect barcodes from URL |
| `ScanBase64` | `POST /barcode/scan-body` | Auto-detect from base64 body |
| `ScanMultipart` | `POST /barcode/scan-multipart` | Auto-detect from uploaded file |

**Scan vs Recognize:** Scan auto-detects all barcode types; Recognize targets specific types.

## Response Models

### BarcodeResponse
```
BarcodeValue  string        // decoded data
Type          string        // barcode classification
Region        []RegionPoint // bounding coordinates
Checksum      string        // validation checksum
```

### BarcodeResponseList
```
Barcodes []BarcodeResponse
```

## Supported Barcode Symbologies (60+)

Known types from SDK/API docs:
- **1D**: Code128, Code39, Code93, EAN13, EAN8, UPCA, UPCE, Codabar, ITF14, ISBN, ISSN, MSI, Pharmacode, Standard2of5, Interleaved2of5, etc.
- **2D**: QR, DataMatrix, PDF417, Aztec, MaxiCode, MicroQR, etc.
- **Postal**: Postnet, Planet, AustralianPost, RM4SCC, DutchKIX, etc.
- **GS1**: GS1Code128, GS1DataMatrix, GS1QR, GS1DotCode, etc.
- **Other**: DotCode, HanXin, PatchCode, etc.

The exact enum values are defined in `EncodeBarcodeType` and `DecodeBarcodeType` in the Go SDK source.

## Go SDK Dependencies

- `github.com/antihax/optional`
- `github.com/google/uuid`
- `golang.org/x/oauth2`

## SDK Code Examples

### Generate barcode
```go
opts := &barcode.GenerateAPIGenerateOpts{
    TextLocation: optional.NewInterface(barcode.CodeLocationNone),
}
data, _, err := client.GenerateAPI.Generate(authCtx,
    barcode.EncodeBarcodeTypeQR,
    "Go SDK example",
    opts)
// data is []byte containing the image
```

### Scan barcode from file
```go
imageFile, _ := os.Open("barcode.png")
recognized, _, err := client.ScanAPI.ScanMultipart(authCtx, imageFile)
for _, b := range recognized.Barcodes {
    fmt.Printf("Type: %s, Value: %s\n", b.Type, b.BarcodeValue)
}
```
