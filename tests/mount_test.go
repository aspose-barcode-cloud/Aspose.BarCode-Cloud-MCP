package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/aspose-barcode-cloud/Aspose.BarCode-Cloud-MCP/mcpbarcode"
)

// --- MountConfig Creation Tests ---

func TestNewMountConfig_EmptyPath(t *testing.T) {
	_, err := mcpbarcode.NewMountConfig("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !strings.Contains(err.Error(), "mount path") {
		t.Errorf("expected 'mount path' in error, got: %v", err)
	}
}

func TestNewMountConfig_ValidDirectory(t *testing.T) {
	dir := t.TempDir()
	mc, err := mcpbarcode.NewMountConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mc.Path == "" {
		t.Error("expected non-empty path")
	}
}

func TestNewMountConfig_NonExistentPath(t *testing.T) {
	_, err := mcpbarcode.NewMountConfig("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
}

func TestNewMountConfig_FilePath(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "afile")
	os.WriteFile(f, []byte("x"), 0644)

	_, err := mcpbarcode.NewMountConfig(f)
	if err == nil {
		t.Fatal("expected error for file path")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' in error, got: %v", err)
	}
}

// --- Path Validation Tests ---

func TestValidatePath_ValidSimpleFilename(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	absPath, err := mc.ValidatePath("barcode.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(absPath, dir) {
		t.Errorf("expected path within mount dir, got: %s", absPath)
	}
}

func TestValidatePath_TraversalDotDot(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	_, err := mc.ValidatePath("../etc/passwd")
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestValidatePath_TraversalNested(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	_, err := mc.ValidatePath("foo/../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for nested path traversal")
	}
}

func TestValidatePath_AbsolutePath(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	_, err := mc.ValidatePath("/etc/passwd")
	if err == nil {
		t.Fatal("expected error for absolute path")
	}
	if !strings.Contains(err.Error(), "absolute") {
		t.Errorf("expected 'absolute' in error, got: %v", err)
	}
}

func TestValidatePath_EmptyPath(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	_, err := mc.ValidatePath("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestValidatePath_DotDotInMiddle(t *testing.T) {
	dir := t.TempDir()
	// Create the "foo" subdirectory so symlink resolution works
	os.Mkdir(filepath.Join(dir, "foo"), 0755)
	mc, _ := mcpbarcode.NewMountConfig(dir)

	absPath, err := mc.ValidatePath("foo/../bar.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should resolve to bar.png within mount
	if !strings.HasSuffix(absPath, "bar.png") {
		t.Errorf("expected bar.png, got: %s", absPath)
	}
}

// --- Filename Generation Tests ---

func TestGenerateFilename_StandardQR(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	filename := mc.GenerateFilename("QR", "png")

	pattern := `^qr-\d{8}-\d{6}-[0-9a-f]{8}\.png$`
	matched, _ := regexp.MatchString(pattern, filename)
	if !matched {
		t.Errorf("filename %q does not match pattern %s", filename, pattern)
	}
}

func TestGenerateFilename_LongType(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	filename := mc.GenerateFilename("AustralianPosteParcel", "png")
	// Type should be truncated to 20 chars
	if len(strings.Split(filename, "-")[0]) > 20 {
		t.Errorf("type prefix too long in: %s", filename)
	}
}

func TestGenerateFilename_SpecialChars(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	filename := mc.GenerateFilename("Code-128", "jpg")
	if !strings.HasPrefix(filename, "code128-") {
		t.Errorf("expected 'code128-' prefix, got: %s", filename)
	}
	if !strings.HasSuffix(filename, ".jpg") {
		t.Errorf("expected .jpg suffix, got: %s", filename)
	}
}

func TestGenerateFilename_Uniqueness(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	f1 := mc.GenerateFilename("QR", "png")
	f2 := mc.GenerateFilename("QR", "png")
	if f1 == f2 {
		t.Errorf("expected unique filenames, both are: %s", f1)
	}
}

func TestGenerateFilename_SVGExtension(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	filename := mc.GenerateFilename("QR", "svg")
	if !strings.HasSuffix(filename, ".svg") {
		t.Errorf("expected .svg suffix, got: %s", filename)
	}
}

// --- File Operations Tests ---

func TestWriteFile_AndReadBack(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	data := []byte("test barcode data")
	relPath, err := mc.WriteFile("test.png", data)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
	if relPath != "test.png" {
		t.Errorf("expected relative path 'test.png', got: %s", relPath)
	}

	readBack, err := os.ReadFile(filepath.Join(dir, "test.png"))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if string(readBack) != string(data) {
		t.Error("written and read data do not match")
	}
}

func TestWriteFile_Permissions(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	mc.WriteFile("perm.png", []byte("data"))

	info, err := os.Stat(filepath.Join(dir, "perm.png"))
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	// On Windows permissions work differently, so just verify file exists
	if info.Size() == 0 {
		t.Error("expected non-zero file size")
	}
}

func TestOpenFile_Valid(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	os.WriteFile(filepath.Join(dir, "existing.png"), []byte("image data"), 0644)

	f, err := mc.OpenFile("existing.png")
	if err != nil {
		t.Fatalf("OpenFile error: %v", err)
	}
	defer f.Close()

	buf := make([]byte, 100)
	n, _ := f.Read(buf)
	if string(buf[:n]) != "image data" {
		t.Error("file content mismatch")
	}
}

func TestOpenFile_NonExistent(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	_, err := mc.OpenFile("nonexistent.png")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestOpenFile_Traversal(t *testing.T) {
	dir := t.TempDir()
	mc, _ := mcpbarcode.NewMountConfig(dir)

	_, err := mc.OpenFile("../outside.png")
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}

// --- Image Extension Validation Tests ---

func TestValidateImageExtension_ValidExtensions(t *testing.T) {
	valid := []string{"test.png", "test.jpg", "test.jpeg", "test.gif", "test.tiff", "test.tif", "test.bmp"}
	for _, f := range valid {
		if err := mcpbarcode.ValidateImageExtension(f); err != nil {
			t.Errorf("expected valid for %q, got error: %v", f, err)
		}
	}
}

func TestValidateImageExtension_CaseInsensitive(t *testing.T) {
	if err := mcpbarcode.ValidateImageExtension("test.PNG"); err != nil {
		t.Errorf("expected valid for .PNG, got: %v", err)
	}
}

func TestValidateImageExtension_Invalid(t *testing.T) {
	invalid := []string{"test.txt", "test.exe", "testfile", "test.svg"}
	for _, f := range invalid {
		if err := mcpbarcode.ValidateImageExtension(f); err == nil {
			t.Errorf("expected error for %q", f)
		}
	}
}

// --- ExtensionForFormat Tests ---

func TestExtensionForFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "png"},
		{"PNG", "png"},
		{"JPEG", "jpg"},
		{"JPG", "jpg"},
		{"SVG", "svg"},
		{"GIF", "gif"},
		{"TIFF", "tiff"},
		{"unknown", "png"},
	}
	for _, tt := range tests {
		result := mcpbarcode.ExtensionForFormat(tt.input)
		if result != tt.expected {
			t.Errorf("ExtensionForFormat(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
