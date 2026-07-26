package setup_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/setup"
)

func TestMasterSetupWorkflow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-setup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	result, err := setup.RunSetup(setup.SetupOptions{
		WorkspaceDir: tempDir,
		Profile:      setup.ProfileLocal,
		OwnerGroup:   "test-group",
		MCPTarget:    setup.MCPTargetAll,
		SkillTarget:  setup.SkillTargetAll,
		BinaryCmd:    "gyrus",
	})
	if err != nil {
		t.Fatalf("RunSetup failed: %v", err)
	}

	// 1. Verify .gyrus.yaml
	configPath := filepath.Join(tempDir, ".gyrus.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected .gyrus.yaml to exist at %s", configPath)
	}

	// 2. Verify Storage Dir
	storageDir := filepath.Join(tempDir, "docs", ".gyrus", "docs")
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		t.Errorf("Expected storage root at %s", storageDir)
	}

	// 3. Verify Agent Skill Files
	cliSkillFile := filepath.Join(tempDir, ".agents", "skills", "gyrus-cli", "SKILL.md")
	if _, err := os.Stat(cliSkillFile); os.IsNotExist(err) {
		t.Errorf("Expected CLI skill file at %s", cliSkillFile)
	}

	mcpSkillFile := filepath.Join(tempDir, ".agents", "skills", "gyrus-mcp", "SKILL.md")
	if _, err := os.Stat(mcpSkillFile); os.IsNotExist(err) {
		t.Errorf("Expected MCP skill file at %s", mcpSkillFile)
	}

	// 4. Verify MCP Registration
	cursorMCP := filepath.Join(tempDir, ".cursor", "mcp.json")
	if _, err := os.Stat(cursorMCP); os.IsNotExist(err) {
		t.Errorf("Expected Cursor MCP config at %s", cursorMCP)
	}
	if len(result.InstalledMCP) == 0 {
		t.Errorf("Expected non-empty InstalledMCP slice")
	}
}
