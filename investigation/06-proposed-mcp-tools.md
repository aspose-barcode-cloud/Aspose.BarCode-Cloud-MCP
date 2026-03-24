# Proposed MCP Tools Design

## Mapping Aspose Barcode Cloud API to MCP Tools

The API has 3 groups of operations. Each maps naturally to MCP tools.

---

## Tool 1: `generate_barcode`

**Purpose**: Generate a barcode image from text/data.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `barcode_type` | string (enum) | yes | Symbology: QR, Code128, Code39, DataMatrix, PDF417, Aztec, EAN13, EAN8, UPCA, UPCE, etc. |
| `data` | string | yes | Content to encode |
| `image_format` | string (enum) | no | Output format: PNG (default), JPEG, SVG, GIF, TIFF |
| `text_location` | string (enum) | no | Text position: Below (default), Above, None |
| `foreground_color` | string | no | Foreground color (name or #AARRGGBB) |
| `background_color` | string | no | Background color (name or #AARRGGBB) |
| `resolution` | number | no | DPI (default varies) |
| `rotation_angle` | number | no | 0, 90, 180, 270 |
| `image_width` | number | no | Width in units |
| `image_height` | number | no | Height in units |

**Returns**: Image content (base64) with mime type, or SVG text.

**SDK mapping**: `client.GenerateAPI.Generate(ctx, barcodeType, data, opts)`

---

## Tool 2: `recognize_barcode`

**Purpose**: Recognize specific barcode types from an image.

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `image_data` | string | yes | Base64-encoded image data |
| `barcode_type` | string (enum) | no | Type to look for (default: all known types) |
| `recognition_mode` | string (enum) | no | Fast, Normal (default), Excellent |
| `recognition_image_kind` | string (enum) | no | Photo, ScannedDocument, ClearImage |

**Returns**: Text with list of recognized barcodes (type + value + region).

**SDK mapping**: `client.RecognizeAPI.RecognizeMultipart(ctx, file, opts)` or `RecognizeBase64`

---

## Tool 3: `scan_barcode`

**Purpose**: Auto-detect and read all barcodes from an image (simpler than recognize).

**Parameters:**
| Name | Type | Required | Description |
|------|------|----------|-------------|
| `image_data` | string | yes | Base64-encoded image data |

**Returns**: Text with list of detected barcodes (type + value).

**SDK mapping**: `client.ScanAPI.ScanMultipart(ctx, file)` or `ScanBase64`

---

## Tool 4 (optional): `list_barcode_types`

**Purpose**: List all supported barcode types for generation and recognition.

**Parameters**: None

**Returns**: Text listing all supported encode and decode barcode symbologies.

**Rationale**: Helps LLMs discover valid barcode types without hardcoding knowledge.

---

## Input/Output Considerations

### Image Input (recognize/scan)
- MCP clients may provide images as base64 strings
- The server should accept base64-encoded image data
- Use `RecognizeBase64` / `ScanBase64` endpoints which accept base64 in JSON body

### Image Output (generate)
- Return barcode images as base64-encoded content with proper MIME type
- MCP supports `"type": "image"` content with `data` (base64) and `mimeType` fields
- For SVG, return as text content

### Error Handling
- Return `mcp.NewToolResultError()` for invalid parameters or API failures
- Include meaningful error messages from the Aspose API

## Authentication Flow

1. Server reads `ASPOSE_CLIENT_ID` and `ASPOSE_CLIENT_SECRET` from environment
2. On startup, creates JWT config using the SDK
3. Creates auth context used for all API calls
4. Token refresh is handled automatically by the SDK
