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
		return nil, fmt.Errorf("mount path is required, use --mount-path flag")
	}

	// Verify directory exists
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("mount path %q: %w", path, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("mount path %q is not a directory", path)
	}

	// Verify writable by creating and removing a unique temp file
	f, err := os.CreateTemp(path, ".mount-test-*")
	if err != nil {
		return nil, fmt.Errorf("mount path %q is not writable: %w", path, err)
	}
	f.Close()
	os.Remove(f.Name())

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
	if relativePath == "" {
		return "", fmt.Errorf("file path is empty")
	}

	if isPathRooted(relativePath) {
		return "", fmt.Errorf("absolute paths are not allowed: %q", relativePath)
	}

	cleaned := filepath.Clean(relativePath)
	if cleaned == "." {
		return "", fmt.Errorf("file path is empty")
	}

	if isPathRooted(cleaned) {
		return "", fmt.Errorf("absolute paths are not allowed: %q", relativePath)
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal not allowed: %q", relativePath)
	}

	absPath := filepath.Join(m.Path, cleaned)
	relToMount, err := filepath.Rel(m.Path, absPath)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path %q: %w", relativePath, err)
	}
	if relToMount == ".." || strings.HasPrefix(relToMount, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal not allowed: %q", relativePath)
	}

	if err := ensurePathWithinMount(m.Path, cleaned); err != nil {
		return "", fmt.Errorf("cannot resolve path %q: %w", relativePath, err)
	}

	return absPath, nil
}

func isPathRooted(path string) bool {
	return filepath.IsAbs(path) ||
		filepath.VolumeName(path) != "" ||
		strings.HasPrefix(path, "/") ||
		strings.HasPrefix(path, `\`)
}

func ensurePathWithinMount(mountPath string, relativePath string) error {
	mountResolved, err := filepath.EvalSymlinks(mountPath)
	if err != nil {
		mountResolved = filepath.Clean(mountPath)
	}

	current := mountResolved
	for _, part := range strings.Split(relativePath, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}

		nextPath := filepath.Join(current, part)
		info, err := os.Lstat(nextPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if info.Mode()&os.ModeSymlink == 0 {
			current = nextPath
			continue
		}

		resolved, err := filepath.EvalSymlinks(nextPath)
		if err != nil {
			return err
		}

		relToMount, err := filepath.Rel(mountResolved, resolved)
		if err != nil {
			return err
		}
		if relToMount == ".." || strings.HasPrefix(relToMount, ".."+string(filepath.Separator)) {
			return fmt.Errorf("path %q resolves outside mount directory", relativePath)
		}

		current = resolved
	}

	return nil
}

// GenerateFilename creates a unique filename for a generated barcode.
// Format: {type}-{YYYYMMDD-HHMMSS}-{8char_hex}.{ext}
func (m *MountConfig) GenerateFilename(barcodeType string, ext string) string {
	typeName := strings.ToLower(barcodeType)
	if len(typeName) > 20 {
		typeName = typeName[:20]
	}
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, typeName)

	timestamp := time.Now().UTC().Format("20060102-150405")

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

// allowedImageExtensions defines recognized image file extensions for input validation.
var allowedImageExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".tiff": true,
	".tif":  true,
	".bmp":  true,
}

// ValidateImageExtension checks that a file has a recognized image extension.
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
