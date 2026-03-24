# MCP Tool Specifications

Detailed specifications for each MCP tool. These use the official `github.com/modelcontextprotocol/go-sdk/mcp` SDK with typed input structs and `mcp.AddTool()` registration.

---

## Tool 1: `generate_barcode`

**Description**: Generate a barcode image. Returns the barcode as a base64-encoded image (PNG by default) or SVG text.

### MCP Definition (go-sdk)

```go
type GenerateBarcodeInput struct {
    BarcodeType     string  `json:"barcode_type"     jsonschema:"required,description=Barcode symbology to generate (e.g. QR\\, Code128\\, DataMatrix\\, EAN13\\, PDF417)"`
    Data            string  `json:"data"             jsonschema:"required,description=Data to encode in the barcode"`
    ImageFormat     string  `json:"image_format"     jsonschema:"description=Output image format,enum=PNG,enum=JPEG,enum=SVG,enum=GIF,enum=TIFF,enum=BMP"`
    TextLocation    string  `json:"text_location"    jsonschema:"description=Where to display human-readable text on the barcode,enum=Below,enum=Above,enum=None"`
    ForegroundColor string  `json:"foreground_color" jsonschema:"description=Foreground color as color name (e.g. Black) or #AARRGGBB hex"`
    BackgroundColor string  `json:"background_color" jsonschema:"description=Background color as color name (e.g. White) or #AARRGGBB hex"`
    Resolution      float64 `json:"resolution"       jsonschema:"description=Image resolution in DPI (1-100000)"`
    RotationAngle   float64 `json:"rotation_angle"   jsonschema:"description=Rotation angle: 0\\, 90\\, 180\\, or 270 degrees"`
    ImageWidth      float64 `json:"image_width"      jsonschema:"description=Image width in pixels"`
    ImageHeight     float64 `json:"image_height"     jsonschema:"description=Image height in pixels"`
}

mcp.AddTool(s, &mcp.Tool{
    Name: "generate_barcode",
    Description: "Generate a barcode image of the specified type encoding the given data. " +
        "Returns the image as base64-encoded content. " +
        "Use list_barcode_types to see all supported barcode types.",
}, makeGenerateHandler(client))
```

### Handler Logic

```
1. input.BarcodeType (required, validated by SDK) → map to barcode.EncodeBarcodeType constant
2. input.Data (required, validated by SDK) → string
3. input.ImageFormat → default "PNG" if empty, map to barcode.BarcodeImageFormat
4. input.TextLocation → map to barcode.CodeLocation if non-empty
5. input.ForegroundColor, input.BackgroundColor → pass as strings if non-empty
6. input.Resolution, input.RotationAngle, input.ImageWidth, input.ImageHeight → numeric opts if non-zero
7. Build barcode.GenerateAPIGenerateOpts struct
8. Call client.API.GenerateAPI.Generate(client.AuthCtx, barcodeType, data, &opts)
9. If error → return nil, nil, fmt.Errorf("...") (SDK wraps into IsError response)
10. If SVG format → return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: svgString}}}
11. Else → return &mcp.CallToolResult{Content: []mcp.Content{&mcp.ImageContent{Data: base64String, MIMEType: "image/png"}}}
```

### SDK Call Reference

```go
opts := &barcode.GenerateAPIGenerateOpts{
    ImageFormat:  optional.NewInterface(barcode.BarcodeImageFormatPng),
    TextLocation: optional.NewInterface(barcode.CodeLocationNone),
    // ... other optional fields
}
imageBytes, httpResp, err := client.API.GenerateAPI.Generate(
    client.AuthCtx,
    barcode.EncodeBarcodeTypeQR,  // mapped from barcode_type param
    "data to encode",             // from data param
    opts,
)
```

---

## Tool 2: `recognize_barcode`

**Description**: Recognize barcodes of a specific type from an image. Use this when you know what barcode types to look for. For auto-detection, use `scan_barcode` instead or set `MostCommonlyUsed` barcode type.

### MCP Definition (go-sdk)

```go
type RecognizeBarcodeInput struct {
    ImageData          string `json:"image_data"            jsonschema:"required,description=Base64-encoded image data (PNG\\, JPEG\\, GIF\\, TIFF\\, or BMP)"`
    BarcodeType        string `json:"barcode_type"          jsonschema:"description=Barcode type to look for (e.g. QR\\, Code128). You can pass many types. Default: most commonly used types"`
    RecognitionMode    string `json:"recognition_mode"      jsonschema:"description=Recognition quality vs speed trade-off,enum=Fast,enum=Normal,enum=Excellent"`
    RecognitionImageKind string `json:"recognition_image_kind" jsonschema:"description=Hint about the image source for better recognition,enum=Photo,enum=ScannedDocument,enum=ClearImage"`
}

mcp.AddTool(s, &mcp.Tool{
    Name: "recognize_barcode",
    Description: "Recognize barcodes of a specific type from a base64-encoded image. " +
        "Allows specifying the barcode type and recognition quality. " +
        "For automatic detection of most commonly barcode types, use scan_barcode instead or set MostCommonlyUsed barcode type.",
}, makeRecognizeHandler(client))
```

### Handler Logic

```
1. input.ImageData (required, validated by SDK) → base64 string for API
2. input.BarcodeType → map to []barcode.DecodeBarcodeType if non-empty (default: MostCommonlyUsed)
3. input.RecognitionMode → map to SDK enum if non-empty
4. input.RecognitionImageKind → map to SDK enum if non-empty
5. Use RecognizeBase64 endpoint (preferred — avoids temp files):
   - Build request body with base64 image data
   - Call client.API.RecognizeAPI.RecognizeBase64(client.AuthCtx, body)
6. If error → return nil, nil, fmt.Errorf("...") (SDK wraps into IsError response)
7. Format BarcodeResponseList as structured text:
   "Found N barcode(s):\n\n1. Type: QR\n   Value: hello\n   Checksum: ...\n\n2. ..."
8. Return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: formattedText}}}
```

### SDK Call Reference (RecognizeBase64)

```go
body := barcode.RecognizeBase64Request{
    BarcodeTypes: []barcode.DecodeBarcodeType{barcode.DecodeBarcodeTypeQR},
    BarcodeImage: base64ImageString,
}
result, httpResp, err := client.API.RecognizeAPI.RecognizeBase64(client.AuthCtx, body)
// result.Barcodes contains []BarcodeResponse
```

---

## Tool 3: `scan_barcode`

**Description**: Auto-detect and read commonly used types of barcodes in an image. Simpler than `recognize_barcode` — no need to specify barcode type or additional params.

### MCP Definition (go-sdk)

```go
type ScanBarcodeInput struct {
    ImageData string `json:"image_data" jsonschema:"required,description=Base64-encoded image data (PNG\\, JPEG\\, GIF\\, TIFF\\, or BMP)"`
}

mcp.AddTool(s, &mcp.Tool{
    Name: "scan_barcode",
    Description: "Automatically detect and read commonly used barcodes in a base64-encoded image. " +
        "Scans for most commonly used supported barcode types without requiring you to specify which type. " +
        "For targeted recognition of a specific barcode type, use recognize_barcode instead.",
}, makeScanHandler(client))
```

### Handler Logic

```
1. input.ImageData (required, validated by SDK) → base64 string for API
2. Use ScanBase64 endpoint (preferred):
   - Build request body with base64 image data
   - Call client.API.ScanAPI.ScanBase64(client.AuthCtx, body)
3. If error → return nil, nil, fmt.Errorf("...") (SDK wraps into IsError response)
4. Format BarcodeResponseList as structured text:
   "Found N barcode(s):\n\n1. Type: QR\n   Value: hello\n\n2. ..."
5. If no barcodes found → return "No barcodes detected in the image."
6. Return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: formattedText}}}
```

### SDK Call Reference

```go
body := barcode.ScanBase64Request{
    BarcodeImage: base64ImageString,
}
result, httpResp, err := client.API.ScanAPI.ScanBase64(client.AuthCtx, body)
```

---

## Tool 4: `list_barcode_types`

**Description**: List all supported barcode symbologies. No API call needed — uses local constants.

### MCP Definition (go-sdk)

```go
type ListBarcodeTypesInput struct{}

mcp.AddTool(s, &mcp.Tool{
    Name: "list_barcode_types",
    Description: "List all supported barcode types for generation and recognition. " +
        "Use this to discover valid barcode_type values for generate_barcode and recognize_barcode.",
}, makeListHandler())
```

### Handler Logic

```
1. No parameters needed (empty input struct)
2. Build formatted text from barcode_types.go constants:
   "Supported barcode types for GENERATION:\n- QR\n- Code128\n- ...\n\n
    Supported barcode types for RECOGNITION:\n- QR\n- Code128\n- ..."
3. Return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: formattedText}}}
```

---

## Response Format Summary

| Tool | Success Response Type | Content |
|------|--------------------|---------|
| `generate_barcode` (raster) | `type: "image"` | `{data: "<base64>", mimeType: "image/png"}` |
| `generate_barcode` (SVG) | `type: "text"` | SVG XML string |
| `recognize_barcode` | `type: "text"` | Formatted list of recognized barcodes |
| `scan_barcode` | `type: "text"` | Formatted list of detected barcodes |
| `list_barcode_types` | `type: "text"` | Formatted list of supported types |
| Any tool (error) | `isError: true` | Error message string |
