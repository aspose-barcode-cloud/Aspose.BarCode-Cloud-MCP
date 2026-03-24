# Implementation Plan

Step-by-step implementation order. Each step is a self-contained unit that can be built and tested independently.

---

## Step 1: Project Initialization

**Goal**: Bootable Go project with dependencies.

1. Initialize Go module:
   ```bash
   go mod init github.com/aspose-barcode-cloud/Aspose.BarCode-Cloud-MCP
   ```
2. Add dependencies:
   ```bash
   go get github.com/mark3labs/mcp-go
   go get github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4
   ```
3. Create `main.go` with minimal MCP server (no tools yet):
   ```go
   package main

   import (
       "log"
       "os"

       "github.com/mark3labs/mcp-go/server"
   )

   func main() {
       log.SetOutput(os.Stderr)

       s := server.NewMCPServer(
           "aspose-barcode-cloud",
           "0.2604.0",
           server.WithToolCapabilities(false),
           server.WithRecovery(),
       )

       if err := server.ServeStdio(s); err != nil {
           log.Fatalf("Server error: %v", err)
       }
   }
   ```
4. Verify it compiles: `go build .`

**Deliverables**: `main.go`, `go.mod`, `go.sum`

---

## Step 2: Aspose Client Wrapper

**Goal**: Authenticated SDK client ready to make API calls.

1. Create `client.go` with `AsposeClient` struct
2. Implement `NewAsposeClient(clientID, clientSecret string) (*AsposeClient, error)`
3. Wire into `main.go`: read env vars, create client, fail fast on missing credentials
4. Verify auth works: add a temporary test that calls a simple API endpoint (remove after)

**Deliverables**: `client.go`, updated `main.go`

**Developer notes**:
- Study the SDK source at `github.com/aspose-barcode-cloud/aspose-barcode-cloud-go/v4` to understand the exact auth pattern for v4
- The investigation shows a `jwt.NewConfig()` pattern — verify this exists in v4 SDK
- If the SDK config struct has fields for OAuthClientId/OAuthClientSecret, use those directly instead of manual JWT setup

---

## Step 3: Barcode Type Constants

**Goal**: Enumerated barcode types for the `list_barcode_types` tool.

1. Create `barcode_types.go`
2. Populate `SupportedEncodeTypes` from SDK's `EncodeBarcodeType` constants
3. Populate `SupportedDecodeTypes` from SDK's `DecodeBarcodeType` constants
4. Create helper function `mapEncodeType(s string) (barcode.EncodeBarcodeType, error)` for case-insensitive lookup
5. Create helper function `mapDecodeType(s string) (barcode.DecodeBarcodeType, error)` for case-insensitive lookup

**Deliverables**: `barcode_types.go`

---

## Step 4: `list_barcode_types` Tool

**Goal**: First working MCP tool — simplest one, no API call.

1. Create `tool_list.go`
2. Define tool with `mcp.NewTool("list_barcode_types", ...)`
3. Implement handler: format and return type lists as text
4. Register in `main.go`: `s.AddTool(listTool, listHandler)`
5. Test manually: run the server, send `tools/list` and `tools/call` via stdin JSON-RPC

**Deliverables**: `tool_list.go`, updated `main.go`

**Manual test** (pipe JSON-RPC to stdin):
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | go run .
```

---

## Step 5: `generate_barcode` Tool

**Goal**: Core generation tool.

1. Create `tool_generate.go`
2. Define tool with full parameter schema (see 04-tool-specifications.md)
3. Implement handler:
   - Map `barcode_type` string to `EncodeBarcodeType` using helper from Step 3
   - Map optional `image_format` to `BarcodeImageFormat`
   - Map optional `text_location` to `CodeLocation`
   - Build `GenerateAPIGenerateOpts`
   - Call SDK `Generate()` method
   - Return image as base64 content or SVG as text
4. Register in `main.go`
5. Test with real Aspose credentials

**Deliverables**: `tool_generate.go`, updated `main.go`

---

## Step 6: `scan_barcode` Tool

**Goal**: Simplest recognition tool (no parameters beyond image).

1. Create `tool_scan.go`
2. Define tool with `image_data` parameter
3. Implement handler:
   - Decode base64 input
   - Prefer `ScanBase64` endpoint
4. Register in `main.go`
5. Test: generate a barcode image (Step 5), then scan it

**Deliverables**: `tool_scan.go`, updated `main.go`

---

## Step 7: `recognize_barcode` Tool

**Goal**: Advanced recognition with type targeting and quality options.

1. Create `tool_recognize.go`
2. Define tool with all parameters
3. Implement handler:
   - Decode base64 input
   - Map optional `barcode_types` to `[]DecodeBarcodeType`
   - Map optional `recognition_mode` and `recognition_image_kind`
   - Prefer `RecognizeBase64` endpoint, fallback to `RecognizeMultipart`
   - Format response as text (include type, value, checksum, region)
4. Register in `main.go`
5. Test with known barcode images

**Deliverables**: `tool_recognize.go`, updated `main.go`

---

## Step 8: Dockerfile

**Goal**: Containerized server for Docker MCP Registry.

1. Create `Dockerfile` (multi-stage build — see 06-dockerfile-and-registry.md)
2. Test build: `docker build -t aspose-barcode-mcp .`
3. Test run: `docker run -e ASPOSE_CLOUD_CLIENT_ID=... -e ASPOSE_CLOUD_CLIENT_SECRET=... -i aspose-barcode-cloud-mcp`
4. Verify stdio transport works through Docker

**Deliverables**: `Dockerfile`

---

## Step 9: Docker MCP Registry Files

**Goal**: Submission-ready registry metadata.

1. Create `registry/server.yaml` (see 06-dockerfile-and-registry.md)
2. Create `registry/tools.json` (see 06-dockerfile-and-registry.md)
3. Validate against registry requirements

**Deliverables**: `registry/server.yaml`, `registry/tools.json`

---

## Step 10: README and Final Polish

**Goal**: Documentation and cleanup.

1. Write `README.md`: description, prerequisites, configuration, usage examples, Docker usage
2. Verify MIT LICENSE file is present
3. Run `go vet ./...` and `go mod tidy`
4. Final Docker build + integration test

**Deliverables**: `README.md`, clean build
