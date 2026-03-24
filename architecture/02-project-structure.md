# Project Structure

## Directory Layout

```
aspose-barcode-mcp/                 # root of the MCP server source (separate repo or subdirectory)
├── main.go                         # Entry point: env validation, client init, server setup, tool registration, ServeStdio
├── client.go                       # AsposeClient wrapper: auth context creation, SDK client lifecycle
├── tool_generate.go                # generate_barcode tool: definition + handler
├── tool_recognize.go               # recognize_barcode tool: definition + handler
├── tool_scan.go                    # scan_barcode tool: definition + handler
├── tool_list.go                    # list_barcode_types tool: definition + handler (no API call)
├── barcode_types.go                # Constants: supported encode/decode barcode type lists
├── go.mod                          # Module: module github.com/aspose-barcode-cloud/Aspose.BarCode-Cloud-MCP
├── go.sum                          # Dependency checksums
├── Dockerfile                      # Multi-stage build for Docker MCP Registry
├── README.md                       # Usage, configuration, examples
├── LICENSE                         # MIT license
└── registry/                       # Docker MCP Registry submission files
    ├── server.yaml                 # Server metadata for registry
    └── tools.json                  # Tool definitions for registry catalog
```

## File Responsibilities

### main.go
- Read and validate `ASPOSE_CLIENT_ID` and `ASPOSE_CLIENT_SECRET` from environment
- Create `AsposeClient` instance
- Create MCP server via `server.NewMCPServer()`
- Register all 4 tools with their handlers
- Call `server.ServeStdio(s)` to start

### client.go
- Define `AsposeClient` struct holding the SDK `*barcode.APIClient` and auth `context.Context`
- Constructor `NewAsposeClient(clientID, clientSecret string) *AsposeClient`
- The auth context with JWT token source is created once and reused (SDK handles token refresh)

### tool_generate.go
- MCP tool definition for `generate_barcode` with all parameters
- Handler function that:
  1. Extracts and validates parameters from `mcp.CallToolRequest`
  2. Maps parameters to SDK `GenerateAPIGenerateOpts`
  3. Calls `client.GenerateAPI.Generate()`
  4. Returns image as base64 (`type: "image"`) or SVG as text (`type: "text"`)

### tool_recognize.go
- MCP tool definition for `recognize_barcode`
- Handler that:
  1. Decodes base64 `image_data` parameter
  2. Calls `client.RecognizeAPI.RecognizeMultipart()` with a temp file or uses `RecognizeBase64()`
  3. Formats results as structured text

### tool_scan.go
- MCP tool definition for `scan_barcode`
- Handler that:
  1. Decodes base64 `image_data` parameter
  2. Calls `client.ScanAPI.ScanMultipart()` or `ScanBase64()`
  3. Formats results as structured text

### tool_list.go
- MCP tool definition for `list_barcode_types` (no parameters)
- Handler returns a formatted text listing from `barcode_types.go` constants

### barcode_types.go
- Two exported string slices: `SupportedEncodeTypes` and `SupportedDecodeTypes`
- Populated from SDK's `EncodeBarcodeType` and `DecodeBarcodeType` enum values
- Used by `list_barcode_types` tool and optionally for input validation
