package plugins_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies GitHub Copilot, Codex, and Claude Code native plugin configs in packaging/plugins/gyrus.
// [Execution Surface]: Filesystem Canonical Plugin Bundle (packaging/plugins/gyrus/)
// [Assertions]: Target plugin and MCP configuration JSON files exist, are valid JSON, and configure Gyrus stdio server.
// -----------------------------------------------------------------------------
func TestMultiAgentCopilotAndClaudeIntegration(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repository root: %v", err)
	}

	pluginDir := filepath.Join(repoRoot, "packaging", "plugins", "gyrus")

	// 1. Claude Code plugin manifests (.claude-plugin/plugin.json and .mcp.json)
	t.Run("ClaudePlugin", func(t *testing.T) {
		manifestPath := filepath.Join(pluginDir, ".claude-plugin", "plugin.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatalf("Missing Claude plugin.json at %s: %v", manifestPath, err)
		}
		var manifest map[string]interface{}
		if err := json.Unmarshal(data, &manifest); err != nil {
			t.Fatalf("Invalid JSON in Claude plugin.json: %v", err)
		}
		if manifest["name"] != "gyrus" {
			t.Errorf("Expected plugin name 'gyrus', got '%v'", manifest["name"])
		}

		mcpPath := filepath.Join(pluginDir, ".mcp.json")
		mcpData, err := os.ReadFile(mcpPath)
		if err != nil {
			t.Fatalf("Missing Claude .mcp.json at %s: %v", mcpPath, err)
		}
		var mcp struct {
			MCPServers map[string]struct {
				Command string            `json:"command"`
				Args    []string          `json:"args"`
				Env     map[string]string `json:"env"`
			} `json:"mcpServers"`
		}
		if err := json.Unmarshal(mcpData, &mcp); err != nil {
			t.Fatalf("Invalid JSON in Claude .mcp.json: %v", err)
		}
		server, ok := mcp.MCPServers["gyrus"]
		if !ok {
			t.Fatalf("Claude .mcp.json missing 'gyrus' server")
		}
		if server.Command != "gyrus" {
			t.Errorf("Expected command 'gyrus', got '%s'", server.Command)
		}
		if server.Env["GYRUS_PROJECT"] != "${workspaceFolder}" {
			t.Errorf("Expected GYRUS_PROJECT env to be '${workspaceFolder}', got '%s'", server.Env["GYRUS_PROJECT"])
		}
	})

	// 2. OpenAI Codex plugin manifest (.codex-plugin/plugin.json)
	t.Run("CodexPlugin", func(t *testing.T) {
		manifestPath := filepath.Join(pluginDir, ".codex-plugin", "plugin.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatalf("Missing Codex plugin.json at %s: %v", manifestPath, err)
		}
		var manifest map[string]interface{}
		if err := json.Unmarshal(data, &manifest); err != nil {
			t.Fatalf("Invalid JSON in Codex plugin.json: %v", err)
		}
		if manifest["name"] != "gyrus" {
			t.Errorf("Expected plugin name 'gyrus', got '%v'", manifest["name"])
		}
		if manifest["skills"] != "./skills/" {
			t.Errorf("Expected skills './skills/', got '%v'", manifest["skills"])
		}
		if manifest["mcpServers"] != "./.mcp.json" {
			t.Errorf("Expected mcpServers './.mcp.json', got '%v'", manifest["mcpServers"])
		}
	})

	// 3. GitHub Copilot & Standard Agent Plugin (plugin.json and mcp.json)
	t.Run("CopilotPlugin", func(t *testing.T) {
		manifestPath := filepath.Join(pluginDir, "plugin.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatalf("Missing Agent Plugins plugin.json at %s: %v", manifestPath, err)
		}
		var manifest map[string]interface{}
		if err := json.Unmarshal(data, &manifest); err != nil {
			t.Fatalf("Invalid JSON in plugin.json: %v", err)
		}
		if manifest["name"] != "gyrus" {
			t.Errorf("Expected plugin name 'gyrus', got '%v'", manifest["name"])
		}

		mcpPath := filepath.Join(pluginDir, "mcp.json")
		mcpData, err := os.ReadFile(mcpPath)
		if err != nil {
			t.Fatalf("Missing Agent Plugins mcp.json at %s: %v", mcpPath, err)
		}
		var mcp struct {
			MCPServers map[string]struct {
				Command string            `json:"command"`
				Args    []string          `json:"args"`
				Env     map[string]string `json:"env"`
			} `json:"mcpServers"`
		}
		if err := json.Unmarshal(mcpData, &mcp); err != nil {
			t.Fatalf("Invalid JSON in mcp.json: %v", err)
		}
		server, ok := mcp.MCPServers["gyrus"]
		if !ok {
			t.Fatalf("mcp.json missing 'gyrus' server")
		}
		if server.Command != "gyrus" {
			t.Errorf("Expected command 'gyrus', got '%s'", server.Command)
		}
		if server.Env["GYRUS_PROJECT"] != "${workspaceFolder}" {
			t.Errorf("Expected GYRUS_PROJECT env to be '${workspaceFolder}', got '%s'", server.Env["GYRUS_PROJECT"])
		}
	})
}
