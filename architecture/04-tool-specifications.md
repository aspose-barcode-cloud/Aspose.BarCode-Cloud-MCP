# MCP Tool Specifications

Detailed specifications for each MCP tool. These map directly to mcp-go `mcp.NewTool()` definitions.

---

## Tool 1: `generate_barcode`

**Description**: Generate a barcode image. Returns the barcode as a base64-encoded image (PNG by default) or SVG text.

### MCP Definition (mcp-go)

```go
mcp.NewTool("generate_barcode",
    mcp.WithDescription("Generate a barcode image of the specified type encoding the given data. "+
        "Returns the image as base64-encoded content. "+
        "Use list_barcode_types to see all supported barcode types."),
    mcp.WithString("barcode_type",
        mcp.Required(),
        mcp.Description("Barcode symbology to generate (e.g., QR, Code128, DataMatrix, EAN13, PDF417)"),
    ),
    mcp.WithString("data",
        mcp.Required(),
        mcp.Description("Data to encode in the barcode"),
    ),
    mcp.WithString("image_format",
        mcp.Description("Output image format"),
        mcp.Enum("PNG", "JPEG", "SVG", "GIF", "TIFF", "BMP"),
    ),
    mcp.WithString("text_location",
        mcp.Description("Where to display human-readable text on the barcode"),
        mcp.Enum("Below", "Above", "None"),
    ),
    mcp.WithString("foreground_color",
        mcp.Description("Foreground color as color name (e.g., Black) or #AARRGGBB hex"),
    ),
    mcp.WithString("background_color",
        mcp.Description("Background color as color name (e.g., White) or #AARRGGBB hex"),
    ),
    mcp.WithNumber("resolution",
        mcp.Description("Image resolution in DPI (1-100000)"),
    ),
    mcp.WithNumber("rotation_angle",
        mcp.Description("Rotation angle: 0, 90, 180, or 270 degrees"),
    ),
    mcp.WithNumber("image_width",
        mcp.Description("Image width in pixels"),
    ),
    mcp.WithNumber("image_height",
        mcp.Description("Image height in pixels"),
    ),
)
```

### Handler Logic

```
1. requiredParam "barcode_type" → map to barcode.EncodeBarcodeType constant
2. requiredParam "data" → string
3. optionalParam "image_format" → default "PNG", map to barcode.BarcodeImageFormat
4. optionalParam "text_location" → map to barcode.CodeLocation
5. optionalParam foreground_color, background_color → pass as strings
6. optionalParam resolution, rotation_angle, image_width, image_height → numeric opts
7. Build barcode.GenerateAPIGenerateOpts struct
8. Call client.API.GenerateAPI.Generate(client.AuthCtx, barcodeType, data, &opts)
9. If error → return mcp.NewToolResultError(err.Error())
10. If SVG format → return mcp.NewToolResultText(string(responseBytes))
11. Else → return CallToolResult with image content:
    {Type: "image", Data: base64.StdEncoding.EncodeToString(responseBytes), MimeType: "image/png"}
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

### MCP Definition (mcp-go)

```go
mcp.NewTool("recognize_barcode",
    mcp.WithDescription("Recognize barcodes of a specific type from a base64-encoded image. "+
        "Allows specifying the barcode type and recognition quality. "+
        "For automatic detection of most commonly barcode types, use scan_barcode instead or set `MostCommonlyUsed` barcode type."),
    mcp.WithString("image_data",
        mcp.Required(),
        mcp.Description("Base64-encoded image data (PNG, JPEG, GIF, TIFF, or BMP)"),
    ),
    mcp.WithString("barcode_type",
        mcp.Description("Barcode type to look for (e.g., QR, Code128). You can pass many types. Default: most commonly used types"),
    ),
    mcp.WithString("recognition_mode",
        mcp.Description("Recognition quality vs speed trade-off"),
        mcp.Enum("Fast", "Normal", "Excellent"),
    ),
    mcp.WithString("recognition_image_kind",
        mcp.Description("Hint about the image source for better recognition"),
        mcp.Enum("Photo", "ScannedDocument", "ClearImage"),
    ),
)
```

### Handler Logic

```
1. requiredParam "image_data" → base64 decode to []byte
2. optionalParam "barcode_type" → map to []barcode.DecodeBarcodeType (default: MostCommonlyUsed). 
3. optionalParam "recognition_mode" → map to SDK enum
4. optionalParam "recognition_image_kind" → map to SDK enum
5. Use RecognizeBase64 endpoint (preferred — avoids temp files):
   - Build request body with base64 image data
   - Call client.API.RecognizeAPI.RecognizeBase64(client.AuthCtx, body)
6. If error → return mcp.NewToolResultError(err.Error())
7. Format BarcodeResponseList as structured text:
   "Found N barcode(s):\n\n1. Type: QR\n   Value: hello\n   Checksum: ...\n\n2. ..."
8. Return mcp.NewToolResultText(formattedText)
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

### MCP Definition (mcp-go)

```go
mcp.NewTool("scan_barcode",
    mcp.WithDescription("Automatically detect and read commonly used barcodes in a base64-encoded image. "+
        "Scans for most commonly used supported barcode types without requiring you to specify which type. "+
        "For targeted recognition of a specific barcode type, use recognize_barcode instead."),
    mcp.WithString("image_data",
        mcp.Required(),
        mcp.Description("Base64-encoded image data (PNG, JPEG, GIF, TIFF, or BMP)"),
    ),
)
```

### Handler Logic

```
1. requiredParam "image_data" → base64 decode to []byte
2. Use ScanBase64 endpoint (preferred):
   - Build request body with base64 image data
   - Call client.API.ScanAPI.ScanBase64(client.AuthCtx, body)
3. If error → return mcp.NewToolResultError(err.Error())
4. Format BarcodeResponseList as structured text:
   "Found N barcode(s):\n\n1. Type: QR\n   Value: hello\n\n2. ..."
5. If no barcodes found → return "No barcodes detected in the image."
6. Return mcp.NewToolResultText(formattedText)
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

### MCP Definition (mcp-go)

```go
mcp.NewTool("list_barcode_types",
    mcp.WithDescription("List all supported barcode types for generation and recognition. "+
        "Use this to discover valid barcode_type values for generate_barcode and recognize_barcode."),
)
```

### Handler Logic

```
1. No parameters needed
2. Build formatted text from barcode_types.go constants:
   "Supported barcode types for GENERATION:\n- QR\n- Code128\n- ...\n\n
    Supported barcode types for RECOGNITION:\n- QR\n- Code128\n- ..."
3. Return mcp.NewToolResultText(formattedText)
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
