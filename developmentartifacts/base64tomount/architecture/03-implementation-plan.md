# Implementation Plan

## Prerequisites

- Read [01-decisions.md](01-decisions.md) for all architecture decisions
- Read [02-design.md](02-design.md) for data flow diagrams and component design
- Read [04-file-changes.md](04-file-changes.md) for exact code change specifications

## Phase 1: Mount Infrastructure

**Goal:** Add the mount configuration layer and path validation. All existing tests must continue to pass.

### Step 1.1: Create `mcpbarcode/mount.go`

Create a new file with `MountConfig` struct and associated methods:

- `MountConfig` struct: `Path string`
- `NewMountConfig(envPath string) (*MountConfig, error)` -- errors if path is empty, validates directory exists and is writable
- `ValidatePath(relativePath string) (string, error)` -- path traversal prevention
- `GenerateFilename(barcodeType string, ext string) string` -- naming per ADR-3
- `WriteFile(filename string, data []byte) (string, error)` -- writes file, returns relative path
- `OpenFile(relativePath string) (*os.File, error)` -- opens file for reading with validation

See [04-file-changes.md](04-file-changes.md) Section 1 for detailed specification.

### Step 1.2: Create `tests/mount_test.go`

Unit tests for all mount.go functions:
- Path validation (traversal attacks, symlinks, valid paths)
- Filename generation (format, uniqueness, extension mapping)
- File write/read round-trip
- Error cases (empty path, directory not writable, file not found)

See [05-test-plan.md](05-test-plan.md) Section 1 for test cases.

**Checkpoint:** `go test ./...` passes. No existing behavior changed.

---

## Phase 2: Update Handler Signatures and Startup

**Goal:** Thread `MountConfig` through to all handlers. Update `main.go` to require `ASPOSE_CLOUD_MOUNT_PATH`.

### Step 2.1: Update handler factory signatures

Change handler factory functions to accept `*MountConfig`:

```go
// Before:
func MakeGenerateHandler(client *AsposeClient) server.ToolHandlerFunc
func MakeRecognizeHandler(client *AsposeClient) server.ToolHandlerFunc
func MakeScanHandler(client *AsposeClient) server.ToolHandlerFunc

// After:
func MakeGenerateHandler(client *AsposeClient, mount *MountConfig) server.ToolHandlerFunc
func MakeRecognizeHandler(client *AsposeClient, mount *MountConfig) server.ToolHandlerFunc
func MakeScanHandler(client *AsposeClient, mount *MountConfig) server.ToolHandlerFunc
```

### Step 2.2: Update `main.go`

- Read `ASPOSE_CLOUD_MOUNT_PATH` environment variable
- Exit with error if not set (required, not optional)
- Create `MountConfig` instance
- Pass to all handler factories

**Checkpoint:** `go test ./...` passes. Handlers accept mount config. Server requires mount path at startup.

---

## Phase 3: Implement Generate with Mount

**Goal:** `generate_barcode` writes all files (including SVG) to mount directory.

### Step 3.1: Update `mcpbarcode/tool_generate.go`

Replace the entire output section:
- Remove `encoding/base64` import
- Remove `mcp.ImageContent` usage
- For all formats (including SVG): write file to mount, return `TextContent` with relative path

See [04-file-changes.md](04-file-changes.md) Section 3 for exact code.

### Step 3.2: Update tests

- Add integration test: generate with mount, verify file exists on disk
- Add test: SVG also written to file (not inline)
- Update existing tests that expect `ImageContent` or inline SVG

**Checkpoint:** Generate writes files for all formats. All tests pass.

---

## Phase 4: Replace image_data with image_path in Recognize/Scan

**Goal:** `recognize_barcode` and `scan_barcode` accept only `image_path` parameter.

### Step 4.1: Update input structs

Replace `ImageData` with `ImagePath` in `RecognizeBarcodeInput` and `ScanBarcodeInput`:
```go
// Before:
ImageData string `json:"image_data" jsonschema:"description=Base64-encoded image data..."`

// After:
ImagePath string `json:"image_path" jsonschema:"description=Path to image file in the mounted data directory"`
```

### Step 4.2: Update handler logic

- Remove `RecognizeBase64` / `ScanBase64` code paths
- Use `RecognizeMultipart` / `ScanMultipart` exclusively
- Validate `image_path` is not empty
- Validate image file extension
- Open file via `mount.OpenFile()`

See [04-file-changes.md](04-file-changes.md) Section 4 for exact code.

### Step 4.3: Update tests

- Add test: recognize with `image_path` (file on disk)
- Add test: scan with `image_path`
- Add test: error when `image_path` is empty
- Add test: error for path traversal attempt
- Add test: error for invalid file extension
- Remove tests that used `image_data` base64 input

**Checkpoint:** Recognize/scan work with `image_path` only. All tests pass.

---

## Phase 5: Update Tool Descriptions and Metadata

**Goal:** Tool descriptions, registry metadata, and client configs reflect the file-based approach.

### Step 5.1: Update `main.go` tool descriptions

Remove all mentions of "base64". Descriptions should reference file paths and mount directory.

### Step 5.2: Update `registry/tools.json`

- Remove `image_data` parameter from recognize and scan tools
- Add `image_path` parameter to recognize and scan tools
- Update `generate_barcode` description to reference file output
- Update `required` arrays

### Step 5.3: Update `registry/server.yaml`

Add `run` section with volumes, env, and `config.parameters` for `data_dir`.

### Step 5.4: Update `.vscode/mcp.json`

Add `-v` mount flag and `ASPOSE_CLOUD_MOUNT_PATH` environment variable.

### Step 5.5: Update `Dockerfile`

Add directory creation for mount point and set permissions.

**Checkpoint:** All metadata is consistent. Docker build succeeds.

---

## Phase 6: Update Docker Integration Tests

**Goal:** Docker tests exercise the mount-based flow.

### Step 6.1: Update `tests/docker_integration_test.go`

- Add `-v` flag to Docker run commands to mount a temp directory
- Add `-e ASPOSE_CLOUD_MOUNT_PATH=/mnt/data` to Docker run commands
- Update all existing tests to use mount-based I/O
- Add tests for generate -> verify file on host
- Add tests for round-trip: generate -> scan via file path

See [05-test-plan.md](05-test-plan.md) Section 3 for test cases.

**Checkpoint:** All tests pass including Docker integration.

---

## Phase Summary

| Phase | Files Changed | Risk | Rollback |
|-------|--------------|------|----------|
| 1 | New: `mount.go`, `mount_test.go` | Low | Delete new files |
| 2 | `main.go`, `tool_generate.go`, `tool_recognize.go`, `tool_scan.go` | Low | Revert signature changes |
| 3 | `tool_generate.go`, tests | Medium | Revert output logic |
| 4 | `tool_recognize.go`, `tool_scan.go`, tests | Medium | Revert input changes |
| 5 | `main.go`, registry/*, `.vscode/mcp.json`, `Dockerfile` | Low | Revert metadata |
| 6 | `docker_integration_test.go` | Low | Revert test changes |

## Recommended Git Strategy

One commit per phase. Each commit should leave tests passing. This enables bisect if issues arise.

```
commit 1: "Add mount configuration infrastructure"
commit 2: "Require ASPOSE_CLOUD_MOUNT_PATH, thread MountConfig through handlers"
commit 3: "Replace base64 output with file-based output for generate_barcode"
commit 4: "Replace image_data with image_path for recognize and scan"
commit 5: "Update tool descriptions, registry, and Docker config"
commit 6: "Update Docker integration tests for mount-based flow"
```
