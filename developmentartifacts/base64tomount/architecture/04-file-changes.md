# File-Level Change Specifications

This document provides exact code-level guidance for each file that needs modification. Developers should implement changes in the order defined in [03-implementation-plan.md](03-implementation-plan.md).

---

## Section 1: New File -- `mcpbarcode/mount.go`

### Purpose
Centralizes all mount-related logic: configuration, path validation, filename generation, file I/O.

### Full Specification

```go
package mcpbarcode

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MountConfig holds the configuration for mount-based file exchange.
type MountConfig struct {
	Path string // Absolute path to mount directory inside container
}

// NewMountConfig creates a MountConfig from the given path.
// Returns an error if path is empty or invalid.
func NewMountConfig(path string) (*MountConfig, error) {
	if path == "" {
		return nil, fmt.Errorf("ASPOSE_CLOUD_MOUNT_PATH is required but not set")
	}

	// Verify directory exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("mount path %q: %w", path, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("mount path %q is not a directory", path)
	}

	// Verify writable by creating and removing a temp file
	testFile := filepath.Join(path, ".mount-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return nil, fmt.Errorf("mount path %q is not writable: %w", path, err)
	}
	os.Remove(testFile)

	return &MountConfig{
		Path: filepath.Clean(path),
	}, nil
}

// ValidatePath validates a relative path is safe and resolves it to an absolute
// path within the mount directory. Returns the absolute path or an error.
//
// Security: prevents path traversal, rejects absolute paths, rejects symlinks
// that escape the mount directory.
func (m *MountConfig) ValidatePath(relativePath string) (string, error) {
	// Reject empty paths
	if relativePath == "" {
		return "", fmt.Errorf("file path is empty")
	}

	// Reject absolute paths
	if filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("absolute paths are not allowed: %q", relativePath)
	}

	// Clean and join with mount root
	cleaned := filepath.Clean(relativePath)

	// Reject paths that try to escape (after cleaning)
	if strings.HasPrefix(cleaned, "..") {
		return "", fmt.Errorf("path traversal not allowed: %q", relativePath)
	}

	absPath := filepath.Join(m.Path, cleaned)

	// Resolve symlinks and verify still within mount
	resolved, err := filepath.EvalSymlinks(filepath.Dir(absPath))
	if err != nil {
		// If the parent directory doesn't exist yet, check without symlink resolution
		// This handles the case where we're about to create a new file
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("cannot resolve path %q: %w", relativePath, err)
		}
		// For non-existent parent, just verify the cleaned path
		resolved = filepath.Dir(absPath)
	} else {
		resolved = filepath.Join(resolved, filepath.Base(absPath))
		absPath = resolved
	}

	// Verify the resolved path is still within mount directory
	mountResolved, err := filepath.EvalSymlinks(m.Path)
	if err != nil {
		mountResolved = m.Path
	}
	if !strings.HasPrefix(filepath.Clean(resolved), mountResolved) {
		return "", fmt.Errorf("path %q resolves outside mount directory", relativePath)
	}

	return absPath, nil
}

// GenerateFilename creates a unique filename for a generated barcode.
// Format: {type}-{YYYYMMDD-HHMMSS}-{8char_hex}.{ext}
func (m *MountConfig) GenerateFilename(barcodeType string, ext string) string {
	// Normalize barcode type: lowercase, truncate
	typeName := strings.ToLower(barcodeType)
	if len(typeName) > 20 {
		typeName = typeName[:20]
	}
	// Remove non-alphanumeric characters
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, typeName)

	timestamp := time.Now().UTC().Format("20060102-150405")

	// Generate 4 random bytes -> 8 hex chars
	b := make([]byte, 4)
	rand.Read(b)
	hex := fmt.Sprintf("%x", b)

	return fmt.Sprintf("%s-%s-%s.%s", safe, timestamp, hex, ext)
}

// WriteFile writes data to a file in the mount directory.
// Returns the relative filename (not absolute path).
func (m *MountConfig) WriteFile(filename string, data []byte) (string, error) {
	absPath, err := m.ValidatePath(filename)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file %q: %w", filename, err)
	}

	return filename, nil
}

// OpenFile opens a file from the mount directory for reading.
// Caller is responsible for closing the returned file.
func (m *MountConfig) OpenFile(relativePath string) (*os.File, error) {
	absPath, err := m.ValidatePath(relativePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", relativePath, err)
	}

	return file, nil
}

// ValidateImageExtension checks that a file has a recognized image extension.
var allowedImageExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".tiff": true,
	".tif":  true,
	".bmp":  true,
}

func ValidateImageExtension(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedImageExtensions[ext] {
		return fmt.Errorf("unsupported image file extension: %q", ext)
	}
	return nil
}

// ExtensionForFormat returns the file extension for a barcode image format string.
func ExtensionForFormat(format string) string {
	switch strings.ToUpper(format) {
	case "JPEG", "JPG":
		return "jpg"
	case "GIF":
		return "gif"
	case "TIFF":
		return "tiff"
	case "SVG":
		return "svg"
	default:
		return "png"
	}
}
```

---

## Section 2: Changes to `main.go`

### Current lines to change

**Add mount config initialization (after client creation, ~line 29):**

```go
// Read mount path configuration (required)
mountPath := os.Getenv("ASPOSE_CLOUD_MOUNT_PATH")
mount, err := mcpbarcode.NewMountConfig(mountPath)
if err != nil {
    log.Fatalf("Mount configuration error: %v", err)
}
log.Printf("Mount mode enabled: %s", mount.Path)
```

**Update handler registrations (pass mount):**

```go
// Before:
), mcpbarcode.MakeGenerateHandler(client))
), mcpbarcode.MakeRecognizeHandler(client))
), mcpbarcode.MakeScanHandler(client))

// After:
), mcpbarcode.MakeGenerateHandler(client, mount))
), mcpbarcode.MakeRecognizeHandler(client, mount))
), mcpbarcode.MakeScanHandler(client, mount))
```

**Update tool descriptions:**

```go
// generate_barcode
"Generate a barcode image of the specified type encoding the given data. " +
"Saves the image file to the mounted data directory and returns the file path. " +
"Use list_barcode_types to see all supported barcode types."

// recognize_barcode
"Recognize barcodes of a specific type from an image file in the mounted data directory. " +
"Allows specifying the barcode type and recognition quality. " +
"For automatic detection of most commonly used barcode types, use scan_barcode instead."

// scan_barcode
"Automatically detect and read commonly used barcodes from an image file " +
"in the mounted data directory. " +
"For targeted recognition of a specific barcode type, use recognize_barcode instead."
```

---

## Section 3: Changes to `mcpbarcode/tool_generate.go`

### Update function signature

```go
// Before:
func MakeGenerateHandler(client *AsposeClient) server.ToolHandlerFunc {

// After:
func MakeGenerateHandler(client *AsposeClient, mount *MountConfig) server.ToolHandlerFunc {
```

### Remove import

Remove `"encoding/base64"` from imports -- no longer needed.

### Replace entire output block (lines 108-118)

Replace everything from `if imageFormat == barcode.BarcodeImageFormatSvg` through end of function:

```go
// Determine file extension and MIME type
ext := ExtensionForFormat(input.ImageFormat)
mimeType := MimeTypeForFormat(imageFormat)
if imageFormat == barcode.BarcodeImageFormatSvg {
    mimeType = "image/svg+xml"
}

// Write file to mount directory
filename := mount.GenerateFilename(input.BarcodeType, ext)
relPath, err := mount.WriteFile(filename, imageBytes)
if err != nil {
    return nil, fmt.Errorf("failed to save barcode image: %w", err)
}

return &mcp.CallToolResult{
    Content: []mcp.Content{
        mcp.TextContent{
            Type: "text",
            Text: fmt.Sprintf("Generated barcode image saved to: %s\nFormat: %s", relPath, mimeType),
        },
    },
}, nil
```

**Key changes from previous version:**
- No SVG special case -- all formats write to file
- No base64 fallback branch
- `encoding/base64` import removed
- `mcp.ImageContent` no longer used

---

## Section 4: Changes to `mcpbarcode/tool_recognize.go`

### Update input struct

Replace `ImageData` with `ImagePath`:

```go
type RecognizeBarcodeInput struct {
    ImagePath            string `json:"image_path"                             jsonschema:"description=Path to image file in the mounted data directory (PNG, JPEG, GIF, TIFF, or BMP)"`
    BarcodeType          string `json:"barcode_type,omitempty"                 jsonschema:"description=Barcode type to look for (e.g. QR, Code128). Default: most commonly used types"`
    RecognitionMode      string `json:"recognition_mode,omitempty"             jsonschema:"description=Recognition quality vs speed trade-off: Fast, Normal, or Excellent"`
    RecognitionImageKind string `json:"recognition_image_kind,omitempty"       jsonschema:"description=Hint about the image source for better recognition: Photo, ScannedDocument, or ClearImage"`
}
```

### Update function signature

```go
func MakeRecognizeHandler(client *AsposeClient, mount *MountConfig) server.ToolHandlerFunc {
```

### Add import

Add `"github.com/antihax/optional"` for `RecognizeAPIRecognizeMultipartOpts`.

### Replace handler body

Replace the entire handler body (after `BindArguments`) with:

```go
// Validate input
if input.ImagePath == "" {
    return nil, fmt.Errorf("'image_path' is required")
}
if err := ValidateImageExtension(input.ImagePath); err != nil {
    return nil, err
}

// Build barcode types list
barcodeTypes := []barcode.DecodeBarcodeType{barcode.DecodeBarcodeTypeMostCommonlyUsed}
if input.BarcodeType != "" {
    parts := strings.Split(input.BarcodeType, ",")
    barcodeTypes = make([]barcode.DecodeBarcodeType, 0, len(parts))
    for _, p := range parts {
        p = strings.TrimSpace(p)
        if p == "" {
            continue
        }
        dt, err := MapDecodeType(p)
        if err != nil {
            return nil, err
        }
        barcodeTypes = append(barcodeTypes, dt)
    }
}

// Open file from mount
file, err := mount.OpenFile(input.ImagePath)
if err != nil {
    return nil, fmt.Errorf("failed to open image: %w", err)
}
defer file.Close()

// RecognizeMultipart requires a single DecodeBarcodeType
recognizeType := barcode.DecodeBarcodeTypeMostCommonlyUsed
if len(barcodeTypes) == 1 {
    recognizeType = barcodeTypes[0]
}

opts := &barcode.RecognizeAPIRecognizeMultipartOpts{}
if input.RecognitionMode != "" {
    mode, err := MapRecognitionMode(input.RecognitionMode)
    if err != nil {
        return nil, err
    }
    opts.RecognitionMode = optional.NewInterface(mode)
}
if input.RecognitionImageKind != "" {
    kind, err := MapRecognitionImageKind(input.RecognitionImageKind)
    if err != nil {
        return nil, err
    }
    opts.RecognitionImageKind = optional.NewInterface(kind)
}

result, _, err := client.API.RecognizeAPI.RecognizeMultipart(
    client.AuthCtx,
    recognizeType,
    file,
    opts,
)
if err != nil {
    return nil, fmt.Errorf("Aspose API error: %w", err)
}

text := FormatBarcodeResults(result)
return &mcp.CallToolResult{
    Content: []mcp.Content{
        mcp.TextContent{Type: "text", Text: text},
    },
}, nil
```

**Key changes from previous version:**
- `ImageData` field removed entirely -- no base64 input
- No dual-input validation (only `image_path`)
- Only `RecognizeMultipart` code path -- `RecognizeBase64` removed
- Check SDK for `RecognizeAPIRecognizeMultipartOpts` exact field names

---

## Section 5: Changes to `mcpbarcode/tool_scan.go`

### Update input struct

Replace `ImageData` with `ImagePath`:

```go
type ScanBarcodeInput struct {
    ImagePath string `json:"image_path" jsonschema:"description=Path to image file in the mounted data directory (PNG, JPEG, GIF, TIFF, or BMP)"`
}
```

### Update function signature

```go
func MakeScanHandler(client *AsposeClient, mount *MountConfig) server.ToolHandlerFunc {
```

### Replace handler body

Replace the entire handler body (after `BindArguments`) with:

```go
// Validate input
if input.ImagePath == "" {
    return nil, fmt.Errorf("'image_path' is required")
}
if err := ValidateImageExtension(input.ImagePath); err != nil {
    return nil, err
}

// Open file from mount
file, err := mount.OpenFile(input.ImagePath)
if err != nil {
    return nil, fmt.Errorf("failed to open image: %w", err)
}
defer file.Close()

result, _, err := client.API.ScanAPI.ScanMultipart(client.AuthCtx, file)
if err != nil {
    return nil, fmt.Errorf("Aspose API error: %w", err)
}

text := FormatBarcodeResults(result)
return &mcp.CallToolResult{
    Content: []mcp.Content{
        mcp.TextContent{Type: "text", Text: text},
    },
}, nil
```

**Key changes from previous version:**
- `ImageData` field removed entirely
- No dual-input validation
- Only `ScanMultipart` code path -- `ScanBase64` removed

---

## Section 6: Changes to `Dockerfile`

### Full updated Dockerfile

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

# Create mount point directory for file exchange
RUN mkdir -p /mnt/data && chmod 777 /mnt/data

COPY --from=builder /mcp-server /mcp-server

ENTRYPOINT ["/mcp-server"]
```

**Note:** `chmod 777` ensures the directory is writable regardless of the user the container runs as. The actual security boundary is the Docker volume mount, not filesystem permissions.

---

## Section 7: Changes to `registry/server.yaml`

### Full updated server.yaml

```yaml
name: aspose-barcode-cloud
image: mcp/aspose-barcode-cloud
type: server
meta:
  category: productivity
  tags:
    - barcode
    - qr-code
    - barcode-generation
    - read-barcode
    - aspose
about:
  title: Aspose BarCode Cloud
  description: >-
    Generate, recognize, and scan 60+ barcode types (QR, Code128, DataMatrix,
    EAN, PDF417, and more) using the Aspose BarCode Cloud API. Barcode images
    are exchanged through a mounted data directory. Generated barcodes are saved
    as files; images for recognition should be placed in the same directory.
  icon: https://avatars.githubusercontent.com/u/3location0918?s=200&v=4
source:
  project: https://github.com/aspose-barcode-cloud/Aspose.BarCode-Cloud-MCP
  commit: <FILL_WITH_40_CHAR_COMMIT_HASH>
run:
  volumes:
    - "{{aspose-barcode-cloud.data_dir}}:/mnt/data"
  env:
    ASPOSE_CLOUD_MOUNT_PATH: "/mnt/data"
config:
  description: >-
    Configure Aspose Cloud API credentials and data directory.
    Get free credentials at https://dashboard.aspose.cloud/applications
  secrets:
    - name: aspose-barcode.client_id
      env: ASPOSE_CLOUD_CLIENT_ID
      example: your-client-id-from-aspose-dashboard
    - name: aspose-barcode.client_secret
      env: ASPOSE_CLOUD_CLIENT_SECRET
      example: your-client-secret-from-aspose-dashboard
  parameters:
    type: object
    properties:
      data_dir:
        type: string
        description: >-
          Host directory for barcode file exchange.
          Generated barcodes will be saved here, and images for
          recognition/scanning should be placed here.
    required:
      - data_dir
```

---

## Section 8: Changes to `.vscode/mcp.json`

```json
{
  "servers": {
    "aspose-barcode-cloud": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "ASPOSE_CLOUD_CLIENT_ID",
        "-e",
        "ASPOSE_CLOUD_CLIENT_SECRET",
        "-e",
        "ASPOSE_CLOUD_MOUNT_PATH=/mnt/data",
        "-v",
        "${workspaceFolder}/.barcode-data:/mnt/data",
        "aspose-barcode-cloud-mcp"
      ]
    }
  }
}
```

---

## Section 9: Changes to `registry/tools.json`

### generate_barcode

1. **Description:** Change to `"Generate a barcode image of the specified type encoding the given data. Saves the image file to the mounted data directory and returns the file path."`

### recognize_barcode

1. **Remove** `image_data` parameter entirely
2. **Add** `image_path` parameter:
```json
"image_path": {
    "type": "string",
    "description": "Path to image file in the mounted data directory (PNG, JPEG, GIF, TIFF, or BMP)"
}
```
3. **Update** `required` array: replace `"image_data"` with `"image_path"`
4. **Update** tool description: remove "base64" mentions

### scan_barcode

1. **Remove** `image_data` parameter entirely
2. **Add** `image_path` parameter:
```json
"image_path": {
    "type": "string",
    "description": "Path to image file in the mounted data directory (PNG, JPEG, GIF, TIFF, or BMP)"
}
```
3. **Update** `required` array: replace `"image_data"` with `"image_path"`
4. **Update** tool description: remove "base64" mentions

---

## Files NOT Changed

| File | Reason |
|------|--------|
| `mcpbarcode/barcode_types.go` | Type mapping unrelated to file transfer |
| `mcpbarcode/tool_list.go` | No file transfer involved |
| `mcpbarcode/client.go` | No changes needed -- mount config is separate from API client |
| `go.mod` / `go.sum` | No new dependencies (uses stdlib `os`, `path/filepath`, `crypto/rand`) |
