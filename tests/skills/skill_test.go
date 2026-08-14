package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
