# Testing Strategy

## Test Levels

### 1. Unit Tests (no API credentials required)

Test tool handlers with a mocked Aspose client.

**Files**: `tool_generate_test.go`, `tool_recognize_test.go`, `tool_scan_test.go`, `tool_list_test.go`

**Approach**:
- Extract an interface from `AsposeClient` for mockability:
  ```go
  type BarcodeGenerator interface {
      Generate(ctx context.Context, barcodeType barcode.EncodeBarcodeType, data string, opts *barcode.GenerateAPIGenerateOpts) ([]byte, *http.Response, error)
  }
  ```
- Or: test at the MCP handler level by creating a real `mcp.CallToolRequest` and asserting the result shape
- For `list_barcode_types`: no mock needed, just verify output contains expected types

**What to test**:
- Required parameter missing → error result
- Valid parameters → correct SDK call mapping
- API error → error result with message
- SVG format → text content type returned
- PNG format → image content type returned
- Empty scan results → "No barcodes detected" message
- Base64 decode failure → clear error message

### 2. Integration Tests (require ASPOSE_CLIENT_ID + ASPOSE_CLIENT_SECRET)

**File**: `integration_test.go`

**Guard**: Skip if credentials are not set:
```go
func skipWithoutCredentials(t *testing.T) {
    if os.Getenv("ASPOSE_CLIENT_ID") == "" || os.Getenv("ASPOSE_CLIENT_SECRET") == "" {
        t.Skip("ASPOSE_CLIENT_ID and ASPOSE_CLIENT_SECRET not set")
    }
}
```

**Test cases**:
1. **Generate + Scan round-trip**: Generate a QR barcode → scan the resulting image → verify decoded value matches input
2. **Generate + Recognize round-trip**: Generate a Code128 barcode → recognize with type hint → verify match
3. **Generate with options**: Generate with custom format (JPEG), text location, colors → verify non-empty response
4. **Invalid barcode type**: Pass nonsense type → verify error response
5. **Empty image scan**: Pass a blank/white image → verify "no barcodes" response

### 3. Docker Build Test

```bash
# Build
docker build -t aspose-barcode-cloud-mcp .

# Verify it starts (will fail on missing credentials but should print error to stderr, not crash)
docker run --rm aspose-barcode-cloud-mcp 2>&1 | head -5

# Full test with credentials
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  | docker run --rm -i \
    -e ASPOSE_CLIENT_ID=$ASPOSE_CLIENT_ID \
    -e ASPOSE_CLIENT_SECRET=$ASPOSE_CLIENT_SECRET \
    aspose-barcode-cloud-mcp
```

### 4. MCP Protocol Test

Verify the full JSON-RPC lifecycle via stdio:

```bash
# Send initialize → initialized notification → tools/list → tools/call sequence
# Can be scripted or tested with an MCP client library
```

**Verify**:
- `initialize` response includes tool capabilities
- `tools/list` returns all 4 tools with correct schemas
- `tools/call` for each tool returns expected content types

## Running Tests

```bash
# Unit tests (no credentials)
go test ./... -short

# Integration tests (with credentials)
ASPOSE_CLIENT_ID=xxx ASPOSE_CLIENT_SECRET=yyy go test ./... -v

# Lint
go vet ./...
```
