package plugins_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies that the Claude Code plugin starts the gyrus MCP server in the project Claude Code is open in.
// [Execution Surface]: Filesystem Canonical Plugin Bundle (packaging/plugins/gyrus/)
// [Assertions]: .claude-plugin/plugin.json redeclares the gyrus server with GYRUS_PROJECT=${CLAUDE_PROJECT_DIR};
// Claude Code leaves ${workspaceFolder}, used by the shared .mcp.json, as literal text.
// -----------------------------------------------------------------------------
func TestClaudePluginMCPServerUsesProjectDir(t *testing.T) {
	pluginDir := filepath.Join("..", "..", "packaging", "plugins", "gyrus")

	var manifest struct {
		MCPServers map[string]struct {
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env"`
		} `json:"mcpServers"`
	}
	data, err := os.ReadFile(filepath.Join(pluginDir, ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("invalid Claude plugin.json: %v", err)
	}
	server, ok := manifest.MCPServers["gyrus"]
	if !ok || server.Command != "gyrus" || len(server.Args) != 2 || server.Args[0] != "mcp" || server.Args[1] != "serve" {
		t.Fatalf("Claude plugin.json must redeclare the gyrus server as 'gyrus mcp serve', got %+v", manifest.MCPServers)
	}
	if server.Env["GYRUS_PROJECT"] != "${CLAUDE_PROJECT_DIR}" {
		t.Errorf("GYRUS_PROJECT must be ${CLAUDE_PROJECT_DIR}, got %q", server.Env["GYRUS_PROJECT"])
	}
}
