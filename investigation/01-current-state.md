# Current State: Base64 File Distribution

## Overview

All binary file data (barcode images) is transferred between MCP host and server as base64-encoded strings embedded in JSON-RPC messages over stdio. No filesystem access is used.

## Generation Flow (Server -> Host)

**File:** `mcpbarcode/tool_generate.go:88-118`

1. Server calls Aspose Cloud API, receives `[]byte` image data
2. For raster formats (PNG, JPEG, GIF, TIFF): encodes bytes with `base64.StdEncoding.EncodeToString(imageBytes)` (line 109)
3. Returns `mcp.ImageContent{Type: "image", Data: encoded, MIMEType: mimeType}`
4. For SVG: returns raw SVG string as `mcp.TextContent`

The full base64 string is embedded in the JSON-RPC response sent over stdout.

## Recognition/Scan Flow (Host -> Server)

**Files:** `mcpbarcode/tool_recognize.go:14-15`, `mcpbarcode/tool_scan.go:14-15`

1. Host sends base64-encoded image as `image_data` string parameter in JSON-RPC request
2. Server passes the base64 string directly to Aspose SDK:
   - `RecognizeBase64(ctx, RecognizeBase64Request{FileBase64: input.ImageData})`
   - `ScanBase64(ctx, ScanBase64Request{FileBase64: input.ImageData})`
3. No decoding happens on the server side - the SDK handles it

## Tool Input/Output Schemas

| Tool | Input | Output |
|------|-------|--------|
| `generate_barcode` | `barcode_type` (string), `data` (string), format options | `mcp.ImageContent` with base64 `Data` field |
| `recognize_barcode` | `image_data` (base64 string), type/mode options | `mcp.TextContent` with formatted results |
| `scan_barcode` | `image_data` (base64 string) | `mcp.TextContent` with formatted results |
| `list_barcode_types` | none | `mcp.TextContent` with type list |

## Tool Descriptions Mentioning Base64

These descriptions are part of the MCP protocol and visible to AI hosts:

- **main.go:37-39**: `generate_barcode` - "Returns the image as base64-encoded content"
- **main.go:44-45**: `recognize_barcode` - "Recognize barcodes of a specific type from a base64-encoded image"
- **main.go:51**: `scan_barcode` - "read commonly used barcodes in a base64-encoded image"
- **registry/tools.json:4**: "Returns the image as base64-encoded content"
- **registry/tools.json:37**: "Base64-encoded image bytes"
- **registry/tools.json:59**: "Base64-encoded image bytes"

## Docker Configuration (No Mounts)

**Dockerfile**: Multi-stage build, no `VOLUME` declarations, no `WORKDIR` in runtime stage

**.vscode/mcp.json**: Docker args are `run -i --rm -e ASPOSE_CLOUD_CLIENT_ID -e ASPOSE_CLOUD_CLIENT_SECRET aspose-barcode-cloud-mcp` - no `-v` or `--mount` flags

**tests/docker_integration_test.go**: Also uses only `-e` flags, no volumes

## Data Size Implications

Base64 encoding increases data size by ~33%. For a typical barcode image:
- A QR code PNG: ~2-10 KB raw -> ~3-14 KB base64
- A detailed barcode with text: ~5-50 KB raw -> ~7-67 KB base64

These sizes are passed through the stdio JSON-RPC channel as part of the message payload.
