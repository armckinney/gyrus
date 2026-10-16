package skills_test

import (
	"gopkg.in/yaml.v3"
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
			name:       "gyrus",
			references: []string{"okf-schemas.md"},
			expectedActions: []string{
				"gyrus_suggest_context",
				"gyrus_search",
				"gyrus_get",
				"gyrus_create",
				"gyrus_update",
				"gyrus_link",
				"gyrus_sync",
				"suggest-context",
				"search",
				"get",
				"create",
				"update",
				"link",
				"sync",
			},
		},
		{name: "gyrus-migrate-okf", references: []string{"sqlite-to-okf.md"}, expectedActions: []string{"gyrus config show", "gyrus sync"}},
		{
			name:       "gyrus-index",
			references: []string{},
			expectedActions: []string{
				"gyrus sync",
			},
		},
	}

	for _, s := range skills {
		t.Run(s.name, func(t *testing.T) {
			skillDir := filepath.Join(repoRoot, "packaging", "plugins", "gyrus", "skills", s.name)
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

func TestMigrationSkillFrontmatterAndReference(t *testing.T) {
	root := filepath.Join("..", "..", "packaging", "plugins", "gyrus", "skills", "gyrus-migrate-okf")
	data, err := os.ReadFile(filepath.Join(root, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) != 3 {
		t.Fatal("missing frontmatter")
	}
	var meta struct {
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		ApplyTo     []string `yaml:"applyTo"`
	}
	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		t.Fatal(err)
	}
	if meta.Name != "gyrus-migrate-okf" || meta.Description == "" || len(meta.ApplyTo) == 0 {
		t.Fatalf("invalid skill metadata: %+v", meta)
	}
	if _, err := os.Stat(filepath.Join(root, "references", "sqlite-to-okf.md")); err != nil {
		t.Fatal(err)
	}
}
