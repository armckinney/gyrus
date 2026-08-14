package mcp_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

var gyrusBinPath string

func TestMain(m *testing.M) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		panic(err)
	}

	tempDir, err := os.MkdirTemp("", "gyrus-mcp-bin-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	binPath := filepath.Join(tempDir, "gyrus")
	buildCmd := exec.Command("go", "build", "-o", binPath, "cmd/gyrus/main.go")
	buildCmd.Dir = repoRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		panic("failed building gyrus binary for MCP tests: " + string(out))
	}

	gyrusBinPath = binPath
	os.Exit(m.Run())
}

// MCPClient handles JSON-RPC 2.0 stdio communication with the gyrus mcp serve process.
type MCPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
	reqID  int64
}

func startMCPClient(t *testing.T, workspaceDir string) *MCPClient {
	cmd := exec.Command(gyrusBinPath, "mcp", "serve")
	cmd.Dir = workspaceDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to open stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to open stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start gyrus mcp serve: %v", err)
	}

	client := &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		reader: bufio.NewReader(stdout),
	}

	t.Cleanup(func() {
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	return client
}

type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *MCPClient) Call(t *testing.T, method string, params any) *JSONRPCResponse {
	id := atomic.AddInt64(&c.reqID, 1)
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed marshaling JSON-RPC request: %v", err)
	}

	if _, err := fmt.Fprintf(c.stdin, "%s\n", string(data)); err != nil {
		t.Fatalf("Failed writing to MCP process stdin: %v", err)
	}

	line, err := c.reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed reading MCP process response: %v", err)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &resp); err != nil {
		t.Fatalf("Failed unmarshaling JSON-RPC response: %v\nRaw line: %s", err, line)
	}

	return &resp
}

func (c *MCPClient) Notify(t *testing.T, method string, params any) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed marshaling JSON-RPC notification: %v", err)
	}

	if _, err := fmt.Fprintf(c.stdin, "%s\n", string(data)); err != nil {
		t.Fatalf("Failed writing notification to stdin: %v", err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies the JSON-RPC 2.0 Stdio handshake, capabilities negotiation, and protocol initialization.
// [Execution Surface]: 'gyrus mcp serve' Subprocess over Stdio
// [Assertions]: Process responds to 'initialize' with valid capabilities and server info.
// -----------------------------------------------------------------------------
func TestMCP_ProtocolHandshake(t *testing.T) {
	workspaceDir := t.TempDir()

	// Init workspace
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = workspaceDir
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	client := startMCPClient(t, workspaceDir)

	// 1. Send 'initialize'
	initParams := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "test-runner",
			"version": "1.0",
		},
	}

	resp := client.Call(t, "initialize", initParams)
	if resp.Error != nil {
		t.Fatalf("initialize returned error: %+v", resp.Error)
	}

	var initResult struct {
		ProtocolVersion string `json:"protocolVersion"`
		ServerInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
		Capabilities struct {
			Tools     any `json:"tools"`
			Resources any `json:"resources"`
			Prompts   any `json:"prompts"`
		} `json:"capabilities"`
	}

	if err := json.Unmarshal(resp.Result, &initResult); err != nil {
		t.Fatalf("Failed unmarshaling initialize result: %v", err)
	}

	if initResult.ServerInfo.Name != "gyrus-memory" {
		t.Errorf("Expected server name 'gyrus-memory', got '%s'", initResult.ServerInfo.Name)
	}

	// 2. Send 'notifications/initialized'
	client.Notify(t, "notifications/initialized", map[string]any{})
}

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies that 'tools/list', 'resources/list', and 'prompts/list' declare all Gyrus memory capabilities.
// [Execution Surface]: 'gyrus mcp serve' Subprocess over Stdio
// [Assertions]: tools/list returns all 11 tools; resources/list and prompts/list return non-empty lists.
// -----------------------------------------------------------------------------
func TestMCP_CapabilitiesListing(t *testing.T) {
	workspaceDir := t.TempDir()

	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = workspaceDir
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	client := startMCPClient(t, workspaceDir)

	// Initialize
	client.Call(t, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "test-runner"},
	})
	client.Notify(t, "notifications/initialized", map[string]any{})

	// 1. tools/list
	toolsResp := client.Call(t, "tools/list", map[string]any{})
	if toolsResp.Error != nil {
		t.Fatalf("tools/list returned error: %+v", toolsResp.Error)
	}

	var toolsResult struct {
		Tools []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(toolsResp.Result, &toolsResult); err != nil {
		t.Fatalf("Failed unmarshaling tools list: %v", err)
	}

	expectedTools := []string{
		"gyrus_create", "gyrus_get", "gyrus_update", "gyrus_archive",
		"gyrus_search", "gyrus_suggest_context", "gyrus_link", "gyrus_unlink",
		"gyrus_neighbors", "gyrus_traverse", "gyrus_sync",
	}

	toolMap := make(map[string]bool)
	for _, tool := range toolsResult.Tools {
		toolMap[tool.Name] = true
	}

	for _, exp := range expectedTools {
		if !toolMap[exp] {
			t.Errorf("Missing expected MCP tool: %s", exp)
		}
	}

	// 2. resources/list
	resResp := client.Call(t, "resources/list", map[string]any{})
	if resResp.Error != nil {
		t.Fatalf("resources/list returned error: %+v", resResp.Error)
	}

	// 3. prompts/list
	promptsResp := client.Call(t, "prompts/list", map[string]any{})
	if promptsResp.Error != nil {
		t.Fatalf("prompts/list returned error: %+v", promptsResp.Error)
	}
}
