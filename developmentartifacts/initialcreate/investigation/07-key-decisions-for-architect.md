# Key Architecture Decisions

## Decisions Required

### 1. Tool Granularity

**Option A (Recommended)**: 3-4 focused tools
- `generate_barcode` - generation
- `recognize_barcode` - targeted recognition with type hint
- `scan_barcode` - auto-detection scan
- `list_barcode_types` - discovery helper

**Option B**: Fine-grained tools per endpoint variant
- `generate_barcode`, `generate_barcode_from_json`, etc.
- More complex, less useful for LLMs

**Recommendation**: Option A. LLMs work better with fewer, well-described tools. The server internally selects the optimal API endpoint.

### 2. Image Data Transfer

**Option A (Recommended)**: Base64 strings in tool parameters
- Works universally across all MCP clients
- No filesystem dependency
- For generate: return base64 image in response
- For recognize/scan: accept base64 image in request

**Option B**: File paths
- Requires shared filesystem between MCP host and server container
- Breaks containerization isolation
- Would need volume mounts

**Recommendation**: Option A. Base64 is the standard approach for MCP image exchange.

### 3. Barcode Type Validation

**Option A**: Validate barcode types on server side before API call
- Better error messages
- Requires maintaining type list in server code

**Option B (Recommended)**: Pass through to Aspose API, return API errors
- Simpler, always up-to-date with API
- Provide `list_barcode_types` tool for discovery

**Option C**: Use enum constraints in MCP tool schema
- MCP `inputSchema` supports `enum` for string types
- Constrains choices at protocol level
- Long enum lists may be unwieldy for LLM display

**Recommendation**: Option B + provide `list_barcode_types`. Keep server simple.

### 4. Project Structure

```
aspose-barcode-mcp/
├── main.go              # entry point, server setup, tool registration
├── tools_generate.go    # generate tool handler
├── tools_recognize.go   # recognize tool handler
├── tools_scan.go        # scan tool handler
├── tools_list.go        # list types tool handler
├── aspose_client.go     # Aspose SDK client wrapper
├── go.mod
├── go.sum
├── Dockerfile
├── README.md
└── registry/            # Files for Docker MCP Registry submission
    ├── server.yaml
    └── tools.json
```

**Alternative**: Single-file approach (everything in main.go)
- Simpler for a small server
- Acceptable if total code is under ~500 lines

### 5. Dockerfile Strategy

Multi-stage build:
```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /mcp-server .

# Runtime stage
FROM alpine:latest
COPY --from=builder /mcp-server /mcp-server
ENTRYPOINT ["/mcp-server"]
```

Minimal image size, no Go toolchain in production.

### 6. Configuration

**Environment variables** (set via Docker MCP Registry secrets mechanism):
- `ASPOSE_CLOUD_CLIENT_ID` (required, secret)
- `ASPOSE_CLOUD_CLIENT_SECRET` (required, secret)

No other configuration needed. API base URL uses SDK defaults.

### 7. Error Handling Strategy

- Invalid/missing parameters: return `mcp.NewToolResultError()` with clear message
- Aspose API errors: extract error message from API response, return as tool error
- Auth failures: return clear error about invalid/missing credentials
- Network failures: return timeout/connectivity error

### 8. SVG vs Raster Output

For `generate_barcode`:
- PNG/JPEG/GIF/TIFF: return as base64 image content (`type: "image"`)
- SVG: return as text content (`type: "text"`) since SVG is XML text

### 9. Testing Strategy

- Unit tests: mock Aspose API client, test tool handlers
- Integration tests: require real Aspose credentials (CI secret)
- Docker build test: `docker build` + `docker run` with `--tools` flag
- Registry validation: `task validate` from Docker MCP Registry tooling

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| Aspose API rate limits | Document in README; not controllable server-side |
| Large image sizes in base64 | Set reasonable default resolution; document limits |
| API credential exposure | Docker secrets mechanism; never log credentials |
| SDK breaking changes | Pin SDK version in go.mod |
| MCP spec changes | Pin mcp-go version; monitor updates |
