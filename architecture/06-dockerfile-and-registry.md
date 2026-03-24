# Dockerfile & Docker MCP Registry Files

## Dockerfile

```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /mcp-server .

# Runtime stage
FROM alpine:3.21

# Add ca-certificates for HTTPS calls to Aspose Cloud API
RUN apk --no-cache add ca-certificates

COPY --from=builder /mcp-server /mcp-server

ENTRYPOINT ["/mcp-server"]
```

**Key points**:
- `CGO_ENABLED=0` for static binary
- `-ldflags="-s -w"` strips debug info for smaller image
- `ca-certificates` required for TLS to Aspose Cloud API and OAuth endpoint
- No `CMD` — the server has no arguments, it reads env vars and serves via stdio
- No `EXPOSE` — stdio transport, no network ports

---

## registry/server.yaml

```yaml
name: aspose-barcode-cloud
image: mcp/aspose-barcode-cloud
type: server
meta:
  category: productivity
  tags:
    - barcode
    - qr-code
    - image-generation
    - read-barcode
    - aspose
about:
  title: Aspose BarCode Cloud
  description: >-
    Generate, recognize, and scan 60+ barcode types (QR, Code128, DataMatrix,
    EAN, PDF417, and more) using the Aspose BarCode Cloud API. Supports barcode
    image generation in multiple formats and barcode reading from images.
  icon: https://avatars.githubusercontent.com/u/3location0918?s=200&v=4
source:
  project: https://github.com/aspose-barcode-cloud/Aspose.BarCode-Cloud-MCP
  commit: <FILL_WITH_40_CHAR_COMMIT_HASH>
config:
  description: >-
    Configure Aspose Cloud API credentials. Get free credentials at
    https://dashboard.aspose.cloud/applications
  secrets:
    - name: aspose-barcode.client_id
      env: ASPOSE_CLIENT_ID
      example: your-client-id-from-aspose-dashboard
    - name: aspose-barcode.client_secret
      env: ASPOSE_CLIENT_SECRET
      example: your-client-secret-from-aspose-dashboard
```

**Notes**:
- `image: mcp/aspose-barcode-cloud` — Docker will build and publish under the `mcp/` namespace
- `commit` must be filled with the actual 40-char commit hash before submission
- `icon` URL should point to the Aspose organization avatar or a custom icon
- Secrets are injected as environment variables by Docker Desktop

---

## registry/tools.json

```json
[
  {
    "name": "generate_barcode",
    "description": "Generate a barcode image of the specified type encoding the given data. Returns the image as base64-encoded content.",
    "arguments": [
      {
        "name": "barcode_type",
        "type": "string",
        "desc": "Barcode symbology (e.g., QR, Code128, DataMatrix, EAN13, PDF417)",
        "required": true
      },
      {
        "name": "data",
        "type": "string",
        "desc": "Data to encode in the barcode",
        "required": true
      },
      {
        "name": "image_format",
        "type": "string",
        "desc": "Output format: PNG (default), JPEG, SVG, GIF, TIFF, BMP"
      },
      {
        "name": "text_location",
        "type": "string",
        "desc": "Text position: Below, Above, None. Default (Above or None) depends on barcode type"
      }
    ]
  },
  {
    "name": "recognize_barcode",
    "description": "Recognize barcodes of a specific type from a base64-encoded image. Allows specifying barcode types and recognition quality.",
    "arguments": [
      {
        "name": "image_data",
        "type": "string",
        "desc": "Base64-encoded image data",
        "required": true
      },
      {
        "name": "barcode_types",
        "type": "string",
        "desc": "Types to look for (e.g., QR, Code128). Default: most commonly used types"
      },
      {
        "name": "recognition_mode",
        "type": "string",
        "desc": "Quality vs speed: Fast, Normal (default), Excellent"
      }
    ]
  },
  {
    "name": "scan_barcode",
    "description": "Automatically detect and read most common barcodes in a base64-encoded image without specifying barcode type.",
    "arguments": [
      {
        "name": "image_data",
        "type": "string",
        "desc": "Base64-encoded image data",
        "required": true
      }
    ]
  },
  {
    "name": "list_barcode_types",
    "description": "List all supported barcode types for generation and recognition.",
    "arguments": []
  }
]
```
