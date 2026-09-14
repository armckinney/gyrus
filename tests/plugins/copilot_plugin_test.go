package plugins_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/setup"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies GitHub Copilot, Codex, and Claude Code client equipping and configuration generation.
// [Execution Surface]: Filesystem Workspace Integration (.vscode/, .claude/, .codex/)
// [Assertions]: Target MCP configuration JSON files exist, are valid JSON, and configure Gyrus stdio server.
// -----------------------------------------------------------------------------
func TestMultiAgentCopilotAndClaudeIntegration(t *testing.T) {
	tempWorkspace := t.TempDir()

	targets := []struct {
		target   setup.ClientTarget
		name     string
		filePath string
	}{
		{setup.ClientTargetCopilot, "VSCode/Copilot", filepath.Join(tempWorkspace, ".vscode", "mcp.json")},
		{setup.ClientTargetClaude, "Claude", filepath.Join(tempWorkspace, ".claude", "mcp.json")},
		{setup.ClientTargetCodex, "Codex", filepath.Join(tempWorkspace, ".codex", "mcp.json")},
	}

	for _, tc := range targets {
		t.Run(tc.name, func(t *testing.T) {
			_, err := setup.RunClientSetup(setup.ClientSetupOptions{
				WorkspaceDir: tempWorkspace,
				Target:       tc.target,
				Mode:         setup.MCPModeLocal,
				BinaryCmd:    "gyrus",
			})
			if err != nil {
				t.Fatalf("RunClientSetup for %s failed: %v", tc.name, err)
			}

			data, err := os.ReadFile(tc.filePath)
			if err != nil {
				t.Fatalf("Failed to read %s config at %s: %v", tc.name, tc.filePath, err)
			}

			var cfg setup.MCPConfigFile
			if err := json.Unmarshal(data, &cfg); err != nil {
				t.Fatalf("Invalid JSON in %s config: %v", tc.name, err)
			}

			if _, ok := cfg.MCPServers["gyrus"]; !ok {
				t.Errorf("%s config does not contain 'gyrus' server entry", tc.name)
			}
		})
	}
}
