package skills_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSkillStaticAnalysis(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
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
				"suggest-context",
				"search",
				"get",
				"create",
				"update",
				"link",
			},
		},
		{
			name:            "gyrus-migrate-sqlite-to-okf",
			references:      []string{"sqlite-to-okf.md"},
			expectedActions: []string{"gyrus config show", "gyrus validate"},
		},
		{
			name:            "gyrus-test-lifecycle",
			references:      []string{"consumer-lifecycle.md"},
			expectedActions: []string{"gyrus config init", "gyrus mcp serve", "gyrus ui serve", "gyrus validate"},
		},
		{
			name:            "gyrus-index",
			expectedActions: []string{"gyrus search"},
		},
	}

	for _, skill := range skills {
		t.Run(skill.name, func(t *testing.T) {
			skillDir := filepath.Join(repoRoot, "packaging", "plugins", "gyrus", "skills", skill.name)
			content, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
			if err != nil {
				t.Fatal(err)
			}
			text := strings.ReplaceAll(string(content), "\r\n", "\n")
			frontmatter, ok := strings.CutPrefix(text, "---\n")
			if !ok {
				t.Fatal("SKILL.md must start with YAML frontmatter")
			}
			frontmatter, _, ok = strings.Cut(frontmatter, "\n---\n")
			if !ok {
				t.Fatal("SKILL.md missing closing frontmatter delimiter")
			}
			var meta struct {
				Name        string `yaml:"name"`
				Description string `yaml:"description"`
			}
			if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
				t.Fatalf("invalid YAML frontmatter: %v", err)
			}
			if meta.Name != skill.name || strings.TrimSpace(meta.Description) == "" {
				t.Errorf("invalid skill metadata: %+v", meta)
			}

			documents := []string{text}
			for _, reference := range skill.references {
				link := "references/" + reference
				if !strings.Contains(text, "]("+link+")") {
					t.Errorf("SKILL.md does not link required reference: %s", link)
				}
				data, err := os.ReadFile(filepath.Join(skillDir, "references", reference))
				if err != nil {
					t.Fatalf("read required reference %s: %v", reference, err)
				}
				documents = append(documents, string(data))
			}

			instructions := strings.Join(documents, "\n")
			for _, action := range skill.expectedActions {
				if !strings.Contains(instructions, action) {
					t.Errorf("skill does not document expected action/tool: %s", action)
				}
			}
		})
	}
}
