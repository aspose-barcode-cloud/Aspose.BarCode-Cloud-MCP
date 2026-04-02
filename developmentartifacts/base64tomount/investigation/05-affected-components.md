# Affected Components Inventory

## Files That Must Change

### 1. mcpbarcode/tool_generate.go
- **Lines 108-118**: Replace `base64.StdEncoding.EncodeToString` + `mcp.ImageContent` with file write + path return
- **Line 5**: `encoding/base64` import can be removed
- **Input struct** (lines 16-27): May need new parameter for output filename/path
- **Handler function**: Must write `imageBytes` to mounted directory, return file path

### 2. mcpbarcode/tool_recognize.go
- **Lines 14-15**: `RecognizeBarcodeInput.ImageData` description says "Base64-encoded image data" - needs to change to file path
- **Lines 48-51**: Replace `RecognizeBase64Request{FileBase64: input.ImageData}` with file-based call
- **Handler function**: Must open file from mounted directory, use `RecognizeMultipart` or read+encode

### 3. mcpbarcode/tool_scan.go
- **Lines 14-15**: `ScanBarcodeInput.ImageData` description says "Base64-encoded image data" - needs to change to file path
- **Lines 26-28**: Replace `ScanBase64Request{FileBase64: input.ImageData}` with file-based call
- **Handler function**: Must open file from mounted directory, use `ScanMultipart` or read+encode

### 4. main.go
- **Lines 37-39**: `generate_barcode` tool description mentions "base64-encoded content"
- **Lines 44-45**: `recognize_barcode` tool description mentions "base64-encoded image"
- **Lines 51**: `scan_barcode` tool description mentions "base64-encoded image"
- **Lines 20-24**: May need new env var for mount path (e.g., `ASPOSE_CLOUD_MOUNT_PATH`)
- **Server setup**: May need to pass mount path to handlers

### 5. Dockerfile
- May need `VOLUME` declaration for mount point
- May need `WORKDIR` or directory creation for the mount target
- May need to set permissions for the mount directory

### 6. .vscode/mcp.json
- Must add `-v` or `--mount` flags to Docker args
- Must define host-side and container-side paths

### 7. registry/server.yaml
- Must add `run` section with `volumes` configuration
- May need `config.parameters` for user-configurable mount path

### 8. registry/tools.json
- All tool descriptions mentioning "base64" must be updated
- `image_data` parameter description must change from "Base64-encoded image bytes" to file path
- `generate_barcode` description must change from "Returns the image as base64-encoded content" to file path

## Files That May Need Changes

### 9. mcpbarcode/client.go
- No direct changes expected, but may need mount path passed through

### 10. tests/integration_test.go
- **Lines 99-158**: Generate+scan round-trip test uses `imgContent.Data` (base64) - must change to file path
- **Lines 278-280**: Creates base64 from `createMinimalWhitePNG()` - must change to write file
- All base64 string handling in tests must be replaced with file operations

### 11. tests/docker_integration_test.go
- Docker run commands must include `-v` mount flags
- Test assertions must handle file paths instead of base64 strings

### 12. tests/tool_generate_test.go, tests/tool_recognize_test.go, tests/tool_scan_test.go
- Unit tests may need updates if input/output types change

### 13. architecture/03-component-design.md
- Data flow diagrams show base64 encoding - need update
- Error handling table mentions "Invalid base64 input"

### 14. architecture/04-tool-specifications.md
- Tool specs reference base64 encoding throughout

## Files That Should NOT Change

- `mcpbarcode/barcode_types.go` - Type mapping unrelated to file transfer
- `mcpbarcode/tool_list.go` - No file transfer involved
- `go.mod` / `go.sum` - No new dependencies needed (unless adding file utilities)
