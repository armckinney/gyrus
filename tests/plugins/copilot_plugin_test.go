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
// [Purpose]: Verifies GitHub Copilot, Codex, Claude Code, and VSCode multi-agent tool configuration generation.
// [Execution Surface]: Filesystem Workspace Integration (.vscode/, .claude/, .codex/)
// [Assertions]: Target MCP configuration JSON files exist, are valid JSON, and configure Gyrus stdio server.
// -----------------------------------------------------------------------------
func TestMultiAgentCopilotAndClaudeIntegration(t *testing.T) {
	tempWorkspace := t.TempDir()

	_, err := setup.RunSetup(setup.SetupOptions{
		WorkspaceDir: tempWorkspace,
		Profile:      setup.ProfileLocal,
		OwnerGroup:   "armckinney",
		MCPTarget:    setup.MCPTargetAll,
		SkillTarget:  setup.SkillTargetAll,
		BinaryCmd:    "gyrus",
	})
	if err != nil {
		t.Fatalf("RunSetup for all targets failed: %v", err)
	}

	targets := []struct {
		name     string
		filePath string
	}{
		{"VSCode/Copilot", filepath.Join(tempWorkspace, ".vscode", "mcp.json")},
		{"Claude", filepath.Join(tempWorkspace, ".claude", "mcp.json")},
		{"Codex", filepath.Join(tempWorkspace, ".codex", "mcp.json")},
	}

	for _, tc := range targets {
		t.Run(tc.name, func(t *testing.T) {
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
