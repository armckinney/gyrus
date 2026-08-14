package plugins_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/setup"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies Google Antigravity IDE & CLI integration manifest, skills equipping, and MCP registration.
// [Execution Surface]: Filesystem Workspace Integration (.antigravity/ and .agents/skills/)
// [Assertions]: Antigravity MCP config is valid JSON; contains 'gyrus' server; skill files match Antigravity skill schema.
// -----------------------------------------------------------------------------
func TestAntigravityPluginIntegration(t *testing.T) {
	tempWorkspace := t.TempDir()

	res, err := setup.RunSetup(setup.SetupOptions{
		WorkspaceDir: tempWorkspace,
		Profile:      setup.ProfileLocal,
		OwnerGroup:   "armckinney",
		MCPTarget:    setup.MCPTargetAntigravity,
		SkillTarget:  setup.SkillTargetAntigravity,
		BinaryCmd:    "gyrus",
	})
	if err != nil {
		t.Fatalf("RunSetup for Antigravity failed: %v", err)
	}

	// 1. Validate .antigravity/mcp.json
	mcpPath := filepath.Join(tempWorkspace, ".antigravity", "mcp.json")
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("Failed to read .antigravity/mcp.json: %v", err)
	}

	var mcpCfg setup.MCPConfigFile
	if err := json.Unmarshal(data, &mcpCfg); err != nil {
		t.Fatalf("Failed to parse .antigravity/mcp.json: %v", err)
	}

	server, exists := mcpCfg.MCPServers["gyrus"]
	if !exists {
		t.Fatalf("Antigravity MCP config does not contain 'gyrus' server")
	}
	if server.Command != "docker" && server.Command != "gyrus" {
		t.Errorf("Unexpected server command: %s", server.Command)
	}

	// 2. Validate Antigravity Skill manifest
	cliSkillPath := filepath.Join(tempWorkspace, ".agents", "skills", "gyrus-cli", "SKILL.md")
	skillContent, err := os.ReadFile(cliSkillPath)
	if err != nil {
		t.Fatalf("Failed reading equipped CLI skill: %v", err)
	}

	contentStr := string(skillContent)
	if !strings.Contains(contentStr, "name: gyrus-cli") {
		t.Errorf("Skill missing valid YAML frontmatter name")
	}
	if !strings.Contains(contentStr, "applyTo:") {
		t.Errorf("Skill missing applyTo frontmatter")
	}
	_ = res
}
