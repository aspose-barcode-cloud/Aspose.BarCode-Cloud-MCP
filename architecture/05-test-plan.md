# Test Plan: Mount-Based File Distribution

## Overview

This test plan covers all testing required to verify the mount-based file distribution feature. Tests are organized by level (unit, integration, Docker integration) and by phase matching [03-implementation-plan.md](03-implementation-plan.md).

---

## Section 1: Unit Tests -- `tests/mount_test.go`

### 1.1 MountConfig Creation

| ID | Test Case | Input | Expected Result |
|----|-----------|-------|-----------------|
| MC-01 | Create with empty path | `""` | Error: ASPOSE_CLOUD_MOUNT_PATH is required |
| MC-02 | Create with valid directory | temp dir path | `MountConfig` with `Path` set |
| MC-03 | Create with non-existent path | `/nonexistent/path` | Error: path does not exist |
| MC-04 | Create with file path (not dir) | path to a file | Error: not a directory |
| MC-05 | Create with read-only directory | read-only temp dir | Error: not writable |

### 1.2 Path Validation (ValidatePath)

| ID | Test Case | Input | Expected Result |
|----|-----------|-------|-----------------|
| PV-01 | Valid simple filename | `"barcode.png"` | Absolute path within mount |
| PV-02 | Valid with subdirectory | `"output/barcode.png"` | Absolute path within mount |
| PV-03 | Path traversal `..` | `"../etc/passwd"` | Error: path traversal |
| PV-04 | Path traversal nested | `"foo/../../etc/passwd"` | Error: path traversal |
| PV-05 | Absolute path | `"/etc/passwd"` | Error: absolute paths not allowed |
| PV-06 | Empty path | `""` | Error: empty path |
| PV-07 | Path with `..` in middle | `"foo/../bar.png"` | Valid (resolves to `bar.png` within mount) |
| PV-08 | Windows-style path separators | `"foo\\bar.png"` | Handled correctly (cleaned by filepath) |

### 1.3 Filename Generation (GenerateFilename)

| ID | Test Case | Input | Expected Result |
|----|-----------|-------|-----------------|
| FG-01 | Standard QR PNG | `"QR", "png"` | Matches pattern `qr-YYYYMMDD-HHMMSS-XXXXXXXX.png` |
| FG-02 | Long barcode type | `"AustralianPosteParcel", "png"` | Type truncated to 20 chars |
| FG-03 | Type with special chars | `"Code-128", "jpg"` | Special chars removed: `code128-...` |
| FG-04 | Uniqueness | Call twice | Two different filenames |
| FG-05 | Extension preserved | `"QR", "jpeg"` | Ends with `.jpeg` |
| FG-06 | SVG extension | `"QR", "svg"` | Ends with `.svg` |

### 1.4 File Operations

| ID | Test Case | Input | Expected Result |
|----|-----------|-------|-----------------|
| FO-01 | WriteFile + read back | Write bytes, read file | Content matches |
| FO-02 | WriteFile permissions | Write file, check stat | File mode `0644` |
| FO-03 | OpenFile valid | Write file, then open | File descriptor returned, content readable |
| FO-04 | OpenFile nonexistent | `"nonexistent.png"` | Error: file not found |
| FO-05 | OpenFile traversal | `"../outside.png"` | Error: path traversal |

### 1.5 Image Extension Validation

| ID | Test Case | Input | Expected Result |
|----|-----------|-------|-----------------|
| IE-01 | Valid .png | `"test.png"` | No error |
| IE-02 | Valid .jpg | `"test.jpg"` | No error |
| IE-03 | Valid .jpeg | `"test.jpeg"` | No error |
| IE-04 | Valid .gif | `"test.gif"` | No error |
| IE-05 | Valid .tiff | `"test.tiff"` | No error |
| IE-06 | Valid .bmp | `"test.bmp"` | No error |
| IE-07 | Invalid .txt | `"test.txt"` | Error: unsupported extension |
| IE-08 | Invalid .exe | `"test.exe"` | Error: unsupported extension |
| IE-09 | No extension | `"testfile"` | Error: unsupported extension |
| IE-10 | Case insensitive | `"test.PNG"` | No error |

### 1.6 ExtensionForFormat

| ID | Test Case | Input | Expected Result |
|----|-----------|-------|-----------------|
| EF-01 | PNG (default) | `""` | `"png"` |
| EF-02 | JPEG | `"JPEG"` | `"jpg"` |
| EF-03 | SVG | `"SVG"` | `"svg"` |
| EF-04 | GIF | `"GIF"` | `"gif"` |
| EF-05 | TIFF | `"TIFF"` | `"tiff"` |

---

## Section 2: Integration Tests -- `tests/integration_test.go`

These tests require Aspose Cloud API credentials (`ASPOSE_CLOUD_CLIENT_ID`, `ASPOSE_CLOUD_CLIENT_SECRET`).

### 2.1 Generate with Mount

| ID | Test Case | Setup | Expected Result |
|----|-----------|-------|-----------------|
| GI-01 | Generate QR PNG to mount | Mount with temp dir | File exists on disk, TextContent response with path ending `.png` |
| GI-02 | Generate Code128 JPEG to mount | Mount with temp dir | `.jpg` file exists, response contains `image/jpeg` |
| GI-03 | Generate SVG to mount | Mount with temp dir | `.svg` file exists on disk, TextContent response with path |
| GI-04 | Generated PNG is valid image | Mount with temp dir | File bytes start with PNG magic bytes (`\x89PNG`) |
| GI-05 | Generated SVG is valid XML | Mount with temp dir | File content starts with `<` or `<?xml` |
| GI-06 | Response contains relative path | Mount with temp dir | Path in response has no `/mnt/data/` prefix |

### 2.2 Recognize/Scan with image_path

| ID | Test Case | Setup | Expected Result |
|----|-----------|-------|-----------------|
| RI-01 | Recognize with image_path | Write test barcode PNG to mount dir | Barcode recognized correctly |
| RI-02 | Scan with image_path | Write test barcode PNG to mount dir | Barcode detected |
| RI-03 | Error: empty image_path | `image_path: ""` | Error: image_path is required |
| RI-04 | Error: path traversal | `image_path: "../etc/passwd"` | Error: path traversal |
| RI-05 | Error: invalid extension | `image_path: "test.txt"` | Error: unsupported extension |
| RI-06 | Error: nonexistent file | `image_path: "nothere.png"` | Error: file not found |
| RI-07 | Recognize with mode and kind | image_path + recognition_mode + image_kind | Works correctly |

### 2.3 Round-Trip (Generate + Scan via Mount)

| ID | Test Case | Steps | Expected Result |
|----|-----------|-------|-----------------|
| RT-01 | Generate then scan | 1. Generate QR with data "test123" to mount 2. Extract filename from response 3. Scan with that filename as image_path | Scan finds QR with value "test123" |
| RT-02 | Generate then recognize | 1. Generate Code128 to mount 2. Recognize with filename + barcode_type=Code128 | Recognition finds correct value |
| RT-03 | Generate SVG then verify | 1. Generate QR as SVG to mount 2. Read file from disk | File contains valid SVG content |

---

## Section 3: Docker Integration Tests -- `tests/docker_integration_test.go`

These tests build and run the Docker container with mount volumes.

### 3.1 Docker Mount Setup

For all Docker mount tests:
- Create a temp directory on the host
- Add `-v {tempdir}:/mnt/data` to Docker run args
- Add `-e ASPOSE_CLOUD_MOUNT_PATH=/mnt/data` to Docker run args
- After test, verify file presence on host temp dir

### 3.2 Test Cases

| ID | Test Case | Docker Args | JSON-RPC Request | Expected Result |
|----|-----------|-------------|-----------------|-----------------|
| DM-01 | Generate writes to host | `-v`, `-e MOUNT_PATH` | generate_barcode QR | File appears in host temp dir |
| DM-02 | Scan reads from host | `-v`, `-e MOUNT_PATH` + pre-place image | scan_barcode with image_path | Barcode detected |
| DM-03 | Round-trip via mount | `-v`, `-e MOUNT_PATH` | generate then scan | Scan finds generated barcode |
| DM-04 | Missing mount path env | No `-e MOUNT_PATH` | (startup) | Server exits with error |
| DM-05 | Path traversal blocked | `-v`, `-e MOUNT_PATH` | scan with `../etc/passwd` | Error in response |
| DM-06 | SVG written to file | `-v`, `-e MOUNT_PATH` | generate_barcode QR SVG | .svg file appears in host temp dir |

### 3.3 Docker Run Command Template

```bash
# Normal operation
docker run -i --rm \
  -e ASPOSE_CLOUD_CLIENT_ID \
  -e ASPOSE_CLOUD_CLIENT_SECRET \
  -e ASPOSE_CLOUD_MOUNT_PATH=/mnt/data \
  -v "${TEMP_DIR}:/mnt/data" \
  aspose-barcode-cloud-mcp

# Missing mount path (should fail)
docker run -i --rm \
  -e ASPOSE_CLOUD_CLIENT_ID \
  -e ASPOSE_CLOUD_CLIENT_SECRET \
  aspose-barcode-cloud-mcp
# Expected: exits with non-zero code, stderr contains "ASPOSE_CLOUD_MOUNT_PATH"
```

---

## Section 4: Security Tests

| ID | Test Case | Category | Input | Expected Result |
|----|-----------|----------|-------|-----------------|
| SEC-01 | Path traversal via `..` | Input validation | `image_path: "../../etc/shadow"` | Rejected |
| SEC-02 | Path traversal via encoded | Input validation | `image_path: "%2e%2e/etc/shadow"` | Rejected (filepath.Clean handles) |
| SEC-03 | Absolute path injection | Input validation | `image_path: "/etc/passwd"` | Rejected |
| SEC-04 | Null byte injection | Input validation | `image_path: "test\x00.png"` | Rejected or handled safely |
| SEC-05 | Very long filename | Input validation | 10000 char path | Error (OS limit or validation) |
| SEC-06 | Special characters | Input validation | `image_path: "test;rm -rf /.png"` | Filename treated literally, no injection |
| SEC-07 | Symlink escape | Filesystem | Create symlink in mount pointing outside | Rejected by EvalSymlinks check |
| SEC-08 | File type check | Input validation | `image_path: "malware.exe"` | Rejected: unsupported extension |

---

## Section 5: Regression Tests

These verify that unchanged functionality still works.

| ID | Test Case | Condition | Expected Result |
|----|-----------|-----------|-----------------|
| REG-01 | List barcode types | Mount configured | Unchanged output |
| REG-02 | Invalid barcode type | Mount configured | Same error as before |
| REG-03 | Missing credentials | Mount configured | Same startup error |
| REG-04 | All format options | Mount configured | PNG, JPEG, GIF, TIFF, SVG all generate valid files |
| REG-05 | Text location options | Mount configured | Below, Above, None all work |
| REG-06 | Color options | Mount configured | foreground_color, background_color work |
| REG-07 | Resolution and rotation | Mount configured | resolution, rotation_angle work |
| REG-08 | Recognition modes | Mount configured | Fast, Normal, Excellent all work with image_path |
| REG-09 | Recognition image kinds | Mount configured | Photo, ScannedDocument, ClearImage all work |

---

## Section 6: Manual QA Checklist

### Pre-requisites
- [ ] Aspose Cloud API credentials available
- [ ] Docker Desktop installed
- [ ] A directory on host machine designated for barcode data exchange

### Claude Desktop Testing
- [ ] Configure `claude_desktop_config.json` with mount args and `ASPOSE_CLOUD_MOUNT_PATH`
- [ ] Ask Claude to generate a QR code -> verify file appears in host directory
- [ ] Place an image in the host directory -> ask Claude to scan it via image_path
- [ ] Verify Claude can access and reference the generated file path

### VS Code Copilot Testing
- [ ] Configure `.vscode/mcp.json` with mount args
- [ ] Generate a barcode -> verify file appears in workspace `.barcode-data/` directory
- [ ] Place an image in `.barcode-data/` -> scan it via image_path
- [ ] Verify tool descriptions show image_path parameter in Copilot tool panel

### Docker MCP Toolkit Testing
- [ ] Deploy via Docker MCP Toolkit with registry metadata
- [ ] Configure `data_dir` parameter in Toolkit UI
- [ ] Verify mount is created automatically
- [ ] Run generate + scan round-trip

### Startup Error Testing
- [ ] Run Docker container WITHOUT `ASPOSE_CLOUD_MOUNT_PATH` env var
- [ ] Verify server exits immediately with clear error message
- [ ] Run Docker container with `ASPOSE_CLOUD_MOUNT_PATH` pointing to non-existent dir
- [ ] Verify server exits with clear error message

---

## Test Environment Matrix

| Environment | Priority |
|-------------|----------|
| Go test (native, temp dir as mount) | P0 |
| Docker container (bind mount) | P0 |
| Claude Desktop (manual config) | P1 |
| VS Code Copilot (workspace mount) | P1 |
| Docker MCP Toolkit (registry) | P1 |

---

## Acceptance Criteria

1. Server exits with error if `ASPOSE_CLOUD_MOUNT_PATH` is not set.
2. `generate_barcode` writes files to mount for all formats (PNG, JPEG, GIF, TIFF, SVG).
3. `generate_barcode` returns `TextContent` with relative file path.
4. `recognize_barcode` and `scan_barcode` accept `image_path` (required parameter).
5. `image_data` parameter is fully removed from recognize and scan tools.
6. Path traversal attacks are blocked.
7. Invalid file extensions are rejected.
8. Docker container creates `/mnt/data` directory.
9. Registry `server.yaml` correctly configures volume mount via `data_dir` parameter.
10. Round-trip works: generate barcode to mount -> scan from mount -> correct result.
11. No `encoding/base64` import remains in tool_generate.go.
12. No `mcp.ImageContent` usage remains in codebase.
