package tests

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

const dockerImage = "aspose-barcode-cloud-mcp:test"

var (
	buildOnce  sync.Once
	buildError error
)

func isDockerPermissionError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "access is denied") ||
		strings.Contains(message, "permission denied") ||
		strings.Contains(message, "error loading config file") ||
		strings.Contains(message, ".docker\\buildx\\instances")
}

// jsonRPCRequest creates a JSON-RPC 2.0 request.
func jsonRPCRequest(id int, method string, params any) []byte {
	msg := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		msg["params"] = params
	}
	b, _ := json.Marshal(msg)
	return append(b, '\n')
}

// jsonRPCNotification creates a JSON-RPC 2.0 notification (no id).
func jsonRPCNotification(method string, params any) []byte {
	msg := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
	}
	if params != nil {
		msg["params"] = params
	}
	b, _ := json.Marshal(msg)
	return append(b, '\n')
}

// jsonRPCResponse represents a parsed JSON-RPC 2.0 response.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.Number     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// dockerMCPSession manages a Docker container running the MCP server.
type dockerMCPSession struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   *bufio.Reader
	stderr   bytes.Buffer
	mountDir string // host-side mount directory
}

func ensureDockerBuild(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not found in PATH")
	}

	buildOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		cmd := exec.CommandContext(ctx, "docker", "build", "-t", dockerImage, ".")
		cmd.Dir = ".."
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildError = fmt.Errorf("docker build failed: %v\n%s", err, out)
		}
	})

	if buildError != nil {
		if isDockerPermissionError(buildError) {
			t.Skipf("docker is unavailable in this environment: %v", buildError)
		}
		t.Fatalf("docker build: %v", buildError)
	}
}

func skipDockerWithoutCredentials(t *testing.T) {
	t.Helper()
	if os.Getenv("ASPOSE_CLOUD_CLIENT_ID") == "" || os.Getenv("ASPOSE_CLOUD_CLIENT_SECRET") == "" {
		t.Skip("ASPOSE_CLOUD_CLIENT_ID and ASPOSE_CLOUD_CLIENT_SECRET not set")
	}
}

func dockerCredential(envVar string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return "test"
}

// startDockerMCP starts the MCP server in a Docker container with credentials and mount.
func startDockerMCP(t *testing.T) *dockerMCPSession {
	t.Helper()

	mountDir := t.TempDir()

	args := []string{
		"run", "--rm", "-i",
		"-e", "ASPOSE_CLOUD_CLIENT_ID=" + dockerCredential("ASPOSE_CLOUD_CLIENT_ID"),
		"-e", "ASPOSE_CLOUD_CLIENT_SECRET=" + dockerCredential("ASPOSE_CLOUD_CLIENT_SECRET"),
		"-v", mountDir + ":/mnt/data",
		dockerImage,
	}

	cmd := exec.Command("docker", args...)
	s := &dockerMCPSession{cmd: cmd, mountDir: mountDir}

	var err error
	s.stdin, err = cmd.StdinPipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	s.stdout = bufio.NewReader(stdoutPipe)
	cmd.Stderr = &s.stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("docker run: %v", err)
	}

	t.Cleanup(func() {
		s.stdin.Close()
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()

		select {
		case <-done:
		case <-time.After(10 * time.Second):
			cmd.Process.Kill()
		}
	})

	return s
}

// sendRequest writes a JSON-RPC request to the container's stdin.
func (s *dockerMCPSession) sendRequest(t *testing.T, data []byte) {
	t.Helper()
	if _, err := s.stdin.Write(data); err != nil {
		t.Fatalf("write to docker stdin: %v", err)
	}
}

// readResponse reads one JSON-RPC response line from stdout.
func (s *dockerMCPSession) readResponse(t *testing.T, timeout time.Duration) jsonRPCResponse {
	t.Helper()

	type result struct {
		line string
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		line, err := s.stdout.ReadString('\n')
		ch <- result{line, err}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			t.Fatalf("read from docker stdout: %v (stderr: %s)", r.err, s.stderr.String())
		}
		var resp jsonRPCResponse
		if err := json.Unmarshal([]byte(r.line), &resp); err != nil {
			t.Fatalf("unmarshal response: %v\nraw: %s", err, r.line)
		}
		return resp
	case <-time.After(timeout):
		t.Fatalf("timeout reading response after %v (stderr: %s)", timeout, s.stderr.String())
		return jsonRPCResponse{} // unreachable
	}
}

// initialize performs the MCP handshake (initialize + notifications/initialized).
func (s *dockerMCPSession) initialize(t *testing.T) jsonRPCResponse {
	t.Helper()

	s.sendRequest(t, jsonRPCRequest(1, "initialize", map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "docker-test", "version": "1.0"},
	}))

	resp := s.readResponse(t, 30*time.Second)

	// Send initialized notification
	s.sendRequest(t, jsonRPCNotification("notifications/initialized", nil))

	return resp
}

func TestDocker_Build(t *testing.T) {
	ensureDockerBuild(t)
	// If we get here, the build succeeded
}

func TestDocker_FailsWithoutCredentials(t *testing.T) {
	ensureDockerBuild(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		dockerImage)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit without credentials")
	}

	out := string(output)
	if !strings.Contains(out, "ASPOSE_CLOUD_CLIENT_ID") && !strings.Contains(out, "ASPOSE_CLOUD_CLIENT_SECRET") {
		t.Errorf("expected credential error message, got: %s", out)
	}
}

func TestDocker_FailsWithoutMountPath(t *testing.T) {
	ensureDockerBuild(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"--entrypoint", "/mcp-server",
		"-e", "ASPOSE_CLOUD_CLIENT_ID=test",
		"-e", "ASPOSE_CLOUD_CLIENT_SECRET=test",
		dockerImage)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit without mount path")
	}

	out := string(output)
	if !strings.Contains(out, "mount path") {
		t.Errorf("expected mount path error message, got: %s", out)
	}
}

func TestDocker_Initialize(t *testing.T) {
	ensureDockerBuild(t)

	s := startDockerMCP(t)
	resp := s.initialize(t)

	if resp.Error != nil {
		t.Fatalf("initialize error: code=%d msg=%s", resp.Error.Code, resp.Error.Message)
	}

	var result map[string]any
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal init result: %v", err)
	}

	if _, ok := result["serverInfo"]; !ok {
		t.Error("missing serverInfo in initialize response")
	}
	if _, ok := result["protocolVersion"]; !ok {
		t.Error("missing protocolVersion in initialize response")
	}
}

func TestDocker_ListTools(t *testing.T) {
	ensureDockerBuild(t)

	s := startDockerMCP(t)
	s.initialize(t)

	s.sendRequest(t, jsonRPCRequest(2, "tools/list", map[string]any{}))
	resp := s.readResponse(t, 15*time.Second)

	if resp.Error != nil {
		t.Fatalf("tools/list error: code=%d msg=%s", resp.Error.Code, resp.Error.Message)
	}

	var result struct {
		Tools []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal tools/list: %v", err)
	}

	if len(result.Tools) != 4 {
		t.Fatalf("expected 4 tools, got %d", len(result.Tools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range result.Tools {
		toolNames[tool.Name] = true
		if tool.Description == "" {
			t.Errorf("tool %q has empty description", tool.Name)
		}
	}

	for _, expected := range []string{"generate_barcode", "recognize_barcode", "scan_barcode", "list_barcode_types"} {
		if !toolNames[expected] {
			t.Errorf("missing expected tool %q", expected)
		}
	}
}

func TestDocker_ListBarcodeTypes(t *testing.T) {
	ensureDockerBuild(t)

	s := startDockerMCP(t)
	s.initialize(t)

	s.sendRequest(t, jsonRPCRequest(2, "tools/call", map[string]any{
		"name":      "list_barcode_types",
		"arguments": map[string]any{},
	}))
	resp := s.readResponse(t, 15*time.Second)

	if resp.Error != nil {
		t.Fatalf("tools/call error: code=%d msg=%s", resp.Error.Code, resp.Error.Message)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content in response")
	}

	text := result.Content[0].Text
	for _, expected := range []string{"QR", "Code128", "DataMatrix", "GENERATION", "RECOGNITION"} {
		if !strings.Contains(text, expected) {
			t.Errorf("expected %q in output", expected)
		}
	}
}

func TestDocker_GenerateAndScanRoundTrip(t *testing.T) {
	ensureDockerBuild(t)
	skipDockerWithoutCredentials(t)

	s := startDockerMCP(t)
	s.initialize(t)

	testData := "Docker integration test"

	// Generate a QR barcode
	s.sendRequest(t, jsonRPCRequest(2, "tools/call", map[string]any{
		"name": "generate_barcode",
		"arguments": map[string]any{
			"barcode_type": "QR",
			"data":         testData,
		},
	}))
	genResp := s.readResponse(t, 60*time.Second)

	if genResp.Error != nil {
		t.Fatalf("generate error: code=%d msg=%s", genResp.Error.Code, genResp.Error.Message)
	}

	var genResult struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(genResp.Result, &genResult); err != nil {
		t.Fatalf("unmarshal generate: %v", err)
	}

	if len(genResult.Content) == 0 {
		t.Fatal("expected content in generate response")
	}
	if genResult.Content[0].Type != "text" {
		t.Fatalf("expected text content, got %q", genResult.Content[0].Type)
	}
	if !strings.Contains(genResult.Content[0].Text, "image/png") {
		t.Errorf("expected image/png in response, got: %s", genResult.Content[0].Text)
	}

	// Extract filename from response
	line := strings.SplitN(genResult.Content[0].Text, "\n", 2)[0]
	filename := strings.TrimPrefix(line, "Generated barcode image saved to: ")

	// Verify file exists on host mount
	hostFile := filepath.Join(s.mountDir, filename)
	if _, err := os.Stat(hostFile); err != nil {
		t.Fatalf("generated file not found on host: %v", err)
	}

	// Scan the generated barcode via file path
	s.sendRequest(t, jsonRPCRequest(3, "tools/call", map[string]any{
		"name": "scan_barcode",
		"arguments": map[string]any{
			"image_path": filename,
		},
	}))
	scanResp := s.readResponse(t, 60*time.Second)

	if scanResp.Error != nil {
		t.Fatalf("scan error: code=%d msg=%s", scanResp.Error.Code, scanResp.Error.Message)
	}

	var scanResult struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(scanResp.Result, &scanResult); err != nil {
		t.Fatalf("unmarshal scan: %v", err)
	}

	if len(scanResult.Content) == 0 {
		t.Fatal("expected content in scan response")
	}
	if !strings.Contains(scanResult.Content[0].Text, testData) {
		t.Errorf("scan result does not contain %q, got: %s", testData, scanResult.Content[0].Text)
	}
}

func TestDocker_GenerateAndRecognizeRoundTrip(t *testing.T) {
	ensureDockerBuild(t)
	skipDockerWithoutCredentials(t)

	s := startDockerMCP(t)
	s.initialize(t)

	testData := "DOCKER12345"

	// Generate a Code128 barcode
	s.sendRequest(t, jsonRPCRequest(2, "tools/call", map[string]any{
		"name": "generate_barcode",
		"arguments": map[string]any{
			"barcode_type": "Code128",
			"data":         testData,
		},
	}))
	genResp := s.readResponse(t, 60*time.Second)

	if genResp.Error != nil {
		t.Fatalf("generate error: code=%d msg=%s", genResp.Error.Code, genResp.Error.Message)
	}

	var genResult struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(genResp.Result, &genResult); err != nil {
		t.Fatalf("unmarshal generate: %v", err)
	}

	// Extract filename
	line := strings.SplitN(genResult.Content[0].Text, "\n", 2)[0]
	filename := strings.TrimPrefix(line, "Generated barcode image saved to: ")

	// Recognize with type hint
	s.sendRequest(t, jsonRPCRequest(3, "tools/call", map[string]any{
		"name": "recognize_barcode",
		"arguments": map[string]any{
			"image_path":   filename,
			"barcode_type": "Code128",
		},
	}))
	recResp := s.readResponse(t, 60*time.Second)

	if recResp.Error != nil {
		t.Fatalf("recognize error: code=%d msg=%s", recResp.Error.Code, recResp.Error.Message)
	}

	var recResult struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(recResp.Result, &recResult); err != nil {
		t.Fatalf("unmarshal recognize: %v", err)
	}

	if len(recResult.Content) == 0 {
		t.Fatal("expected content in recognize response")
	}
	if !strings.Contains(recResult.Content[0].Text, testData) {
		t.Errorf("recognize result does not contain %q, got: %s", testData, recResult.Content[0].Text)
	}
	if !strings.Contains(recResult.Content[0].Text, "Code128") {
		t.Errorf("recognize result does not mention Code128, got: %s", recResult.Content[0].Text)
	}
}

func TestDocker_GenerateSVG(t *testing.T) {
	ensureDockerBuild(t)
	skipDockerWithoutCredentials(t)

	s := startDockerMCP(t)
	s.initialize(t)

	s.sendRequest(t, jsonRPCRequest(2, "tools/call", map[string]any{
		"name": "generate_barcode",
		"arguments": map[string]any{
			"barcode_type": "QR",
			"data":         "SVG Docker test",
			"image_format": "SVG",
		},
	}))
	resp := s.readResponse(t, 60*time.Second)

	if resp.Error != nil {
		t.Fatalf("generate SVG error: code=%d msg=%s", resp.Error.Code, resp.Error.Message)
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content")
	}
	if result.Content[0].Type != "text" {
		t.Fatalf("SVG should return text content, got %q", result.Content[0].Type)
	}
	if !strings.Contains(result.Content[0].Text, "image/svg+xml") {
		t.Errorf("expected image/svg+xml in response, got: %s", result.Content[0].Text)
	}

	// Extract filename and verify SVG file on host
	line := strings.SplitN(result.Content[0].Text, "\n", 2)[0]
	filename := strings.TrimPrefix(line, "Generated barcode image saved to: ")
	hostFile := filepath.Join(s.mountDir, filename)
	data, err := os.ReadFile(hostFile)
	if err != nil {
		t.Fatalf("SVG file not found on host: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "<svg") && !strings.Contains(content, "<?xml") {
		t.Errorf("expected SVG content, got: %.100s...", content)
	}
}

func TestDocker_PathTraversalBlocked(t *testing.T) {
	ensureDockerBuild(t)

	s := startDockerMCP(t)
	s.initialize(t)

	s.sendRequest(t, jsonRPCRequest(2, "tools/call", map[string]any{
		"name": "scan_barcode",
		"arguments": map[string]any{
			"image_path": "../etc/passwd.png",
		},
	}))
	resp := s.readResponse(t, 30*time.Second)

	// Should get an error either as JSON-RPC error or as isError in result
	if resp.Error != nil {
		return // JSON-RPC level error — fine
	}

	var result struct {
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for path traversal")
	}
}

func TestDocker_InvalidBarcodeType(t *testing.T) {
	ensureDockerBuild(t)

	s := startDockerMCP(t)
	s.initialize(t)

	s.sendRequest(t, jsonRPCRequest(2, "tools/call", map[string]any{
		"name": "generate_barcode",
		"arguments": map[string]any{
			"barcode_type": "COMPLETELY_FAKE_TYPE",
			"data":         "test",
		},
	}))
	resp := s.readResponse(t, 30*time.Second)

	// Should get an error either as JSON-RPC error or as isError in result
	if resp.Error != nil {
		// JSON-RPC level error — this is fine
		return
	}

	var result struct {
		IsError bool `json:"isError"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for invalid barcode type")
	}
}
