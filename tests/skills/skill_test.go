package skills_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/mcp"
)

func TestSkillPackagesValidation(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repo root: %v", err)
	}

	skills := []struct {
		name       string
		references []string
	}{
		{
			name:       "gyrus-cli",
			references: []string{"okf-schemas.md", "mcp-setup.md", "storage-providers.md"},
		},
		{
			name:       "gyrus-mcp",
			references: []string{"okf-schemas.md", "mcp-setup.md", "storage-providers.md"},
		},
	}

	for _, s := range skills {
		t.Run(s.name, func(t *testing.T) {
			skillDir := filepath.Join(repoRoot, "skills", s.name)
			skillMD := filepath.Join(skillDir, "SKILL.md")

			// 1. Verify SKILL.md existence
			content, err := os.ReadFile(skillMD)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", skillMD, err)
			}

			// 2. Verify Frontmatter name
			expectedName := "name: " + s.name
			if !strings.Contains(string(content), expectedName) {
				t.Errorf("Expected frontmatter %s in %s", expectedName, skillMD)
			}

			// 3. Verify reference files existence
			for _, ref := range s.references {
				refPath := filepath.Join(skillDir, "references", ref)
				if _, err := os.Stat(refPath); os.IsNotExist(err) {
					t.Errorf("Missing required skill reference: %s", refPath)
				}
			}
		})
	}

	// 4. Verify CLI executable responsiveness (read-only help call)
	cmd := exec.Command("./gyrus", "help")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("CLI executable test failed: %v\nOutput:\n%s", err, string(out))
	}

	// 5. Verify MCP server initialization in an isolated temp directory (auto-deleted by Go testing framework)
	tempDir := t.TempDir()
	tempStorageDir := filepath.Join(tempDir, "docs")
	if err := os.MkdirAll(tempStorageDir, 0755); err != nil {
		t.Fatalf("Failed creating temp storage dir: %v", err)
	}

	srv, err := mcp.NewServer(tempStorageDir)
	if err != nil {
		t.Fatalf("Failed creating MCP server for skill validation: %v", err)
	}
	_ = srv
}
