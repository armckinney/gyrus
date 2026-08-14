package skills_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/mcp"
	"github.com/armckinney/gyrus/internal/setup"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Validates Agent Skills structure, YAML frontmatter, references, and tool mappings on disk.
// [Execution Surface]: Packaging Skills Directory (packaging/skills/)
// [Assertions]: Skills have valid name, applyTo block, existing reference docs, and documented tool actions.
// -----------------------------------------------------------------------------
func TestSkillStaticAnalysis(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repo root: %v", err)
	}

	skills := []struct {
		name            string
		references      []string
		expectedActions []string
	}{
		{
			name:       "gyrus-cli",
			references: []string{"okf-schemas.md", "mcp-setup.md", "storage-providers.md"},
			expectedActions: []string{
				"suggest-context",
				"search",
				"get",
				"create",
				"update",
				"link",
				"sync",
			},
		},
		{
			name:       "gyrus-mcp",
			references: []string{"okf-schemas.md", "mcp-setup.md", "storage-providers.md"},
			expectedActions: []string{
				"gyrus_suggest_context",
				"gyrus_search",
				"gyrus_get",
				"gyrus_create",
				"gyrus_update",
				"gyrus_link",
				"gyrus_sync",
			},
		},
	}

	for _, s := range skills {
		t.Run(s.name, func(t *testing.T) {
			skillDir := filepath.Join(repoRoot, "packaging", "skills", s.name)
			skillMD := filepath.Join(skillDir, "SKILL.md")

			// a. Verify SKILL.md existence & readable frontmatter
			content, err := os.ReadFile(skillMD)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", skillMD, err)
			}

			strContent := string(content)

			// b. Verify YAML frontmatter fields
			if !strings.Contains(strContent, "name: "+s.name) {
				t.Errorf("SKILL.md missing valid name frontmatter: name: %s", s.name)
			}
			if !strings.Contains(strContent, "applyTo:") {
				t.Errorf("SKILL.md missing applyTo frontmatter block")
			}

			// c. Verify reference guides exist on disk
			for _, ref := range s.references {
				refPath := filepath.Join(skillDir, "references", ref)
				if _, err := os.Stat(refPath); os.IsNotExist(err) {
					t.Errorf("Missing required skill reference file: %s", refPath)
				}
			}

			// d. Verify all documented actions are present in SKILL.md text
			for _, action := range s.expectedActions {
				if !strings.Contains(strContent, action) {
					t.Errorf("SKILL.md for %s does not document expected action/tool: %s", s.name, action)
				}
			}
		})
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies that equipping skills for Antigravity creates config, skills, and starts MCP server.
// [Execution Surface]: Filesystem Workspace Integration
// [Assertions]: Antigravity MCP config generated, skill files present, and MCP server initializes targeting storage dir.
// -----------------------------------------------------------------------------
func TestAntigravityBehavioralIntegration(t *testing.T) {
	tempDir := t.TempDir()

	res, err := setup.RunSetup(setup.SetupOptions{
		WorkspaceDir: tempDir,
		Profile:      setup.ProfileLocal,
		OwnerGroup:   "armckinney",
		MCPTarget:    setup.MCPTargetAntigravity,
		SkillTarget:  setup.SkillTargetAntigravity,
		BinaryCmd:    "gyrus",
	})
	if err != nil {
		t.Fatalf("setup.RunSetup for Antigravity failed: %v", err)
	}

	antigravityMCP := filepath.Join(tempDir, ".antigravity", "mcp.json")
	if _, err := os.Stat(antigravityMCP); os.IsNotExist(err) {
		t.Fatalf("Dedicated Antigravity MCP config file not created at %s", antigravityMCP)
	}

	mcpContent, err := os.ReadFile(antigravityMCP)
	if err != nil {
		t.Fatalf("Failed reading Antigravity MCP config: %v", err)
	}

	if !strings.Contains(string(mcpContent), "gyrus") {
		t.Errorf("Antigravity MCP config missing gyrus server definition: %s", string(mcpContent))
	}

	cliSkillFile := filepath.Join(tempDir, ".agents", "skills", "gyrus-cli", "SKILL.md")
	if _, err := os.Stat(cliSkillFile); os.IsNotExist(err) {
		t.Errorf("Equipped CLI skill file missing at %s", cliSkillFile)
	}

	mcpSkillFile := filepath.Join(tempDir, ".agents", "skills", "gyrus-mcp", "SKILL.md")
	if _, err := os.Stat(mcpSkillFile); os.IsNotExist(err) {
		t.Errorf("Equipped MCP skill file missing at %s", mcpSkillFile)
	}

	srv, err := mcp.NewServer(res.StorageDir)
	if err != nil {
		t.Fatalf("Failed initializing MCP server for Antigravity: %v", err)
	}
	_ = srv
}

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies that the compiled gyrus binary is built and responsive to CLI commands in the workspace.
// [Execution Surface]: Subprocess execution of compiled ./gyrus binary
// [Assertions]: Process exits successfully and prints CLI help text.
// -----------------------------------------------------------------------------
func TestCLICommandSuiteBehavior(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repo root: %v", err)
	}

	binPath := filepath.Join(repoRoot, "gyrus")
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		buildCmd := exec.Command("go", "build", "-o", "gyrus", "cmd/gyrus/main.go")
		buildCmd.Dir = repoRoot
		if out, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to auto-build gyrus binary for test: %v\nOutput:\n%s", err, string(out))
		}
	}

	cmd := exec.Command("./gyrus", "help")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI help command execution failed: %v\nOutput:\n%s", err, string(out))
	}

	if !strings.Contains(string(out), "gyrus") {
		t.Errorf("Unexpected CLI help output: %s", string(out))
	}
}
