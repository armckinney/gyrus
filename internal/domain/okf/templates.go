package okf

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed templates/*.md
var embeddedTemplates embed.FS

// ListEmbeddedTemplates returns all built-in concept template types compiled in the binary.
func ListEmbeddedTemplates() []string {
	entries, err := embeddedTemplates.ReadDir("templates")
	if err != nil {
		return nil
	}
	var res []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			name := strings.TrimSuffix(entry.Name(), ".md")
			res = append(res, name)
		}
	}
	sort.Strings(res)
	return res
}

// ValidateSchemaContent verifies that a schema template contains valid OKF frontmatter
// and matches the expected concept type.
func ValidateSchemaContent(conceptType string, content string) error {
	if conceptType == "" || !idRegex.MatchString(conceptType) {
		return fmt.Errorf("invalid schema concept type '%s': must match regex ^[a-z0-9-_]+$", conceptType)
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("schema content cannot be empty")
	}

	doc, err := ParseMarkdown([]byte(content))
	if err != nil {
		return fmt.Errorf("invalid schema template frontmatter: %w", err)
	}

	if doc.Type != "" && string(doc.Type) != conceptType {
		return fmt.Errorf("schema template type '%s' does not match target concept type '%s'", doc.Type, conceptType)
	}

	return nil
}

// ValidateSchemaPath verifies that a remote or local schema path conforms to enforced locations (.gyrus/schemas/<conceptType>.md)
// and does not contain directory traversals.
func ValidateSchemaPath(path string) error {
	clean := filepath.ToSlash(filepath.Clean(path))
	if strings.Contains(clean, "..") {
		return fmt.Errorf("invalid schema path '%s': directory traversal is not permitted", path)
	}
	if !strings.HasSuffix(clean, ".md") {
		return fmt.Errorf("invalid schema path '%s': must end with .md", path)
	}
	base := strings.TrimSuffix(filepath.Base(clean), ".md")
	if !idRegex.MatchString(base) {
		return fmt.Errorf("invalid schema path filename '%s': concept type must match ^[a-z0-9-_]+$", base)
	}
	if !strings.Contains(clean, ".gyrus/schemas/") && !strings.Contains(clean, "schemas/") {
		return fmt.Errorf("invalid schema path '%s': must be located within .gyrus/schemas/", path)
	}
	return nil
}

// GetTemplate retrieves the schema template for a concept type.
// It prioritizes custom templates in customSchemasDir, falling back to embedded binary templates.
func GetTemplate(conceptType string, customSchemasDir string) (string, error) {
	// 1. Check custom user schemas directory in project repository if specified
	if customSchemasDir != "" {
		customPath := filepath.Join(customSchemasDir, fmt.Sprintf("%s.md", conceptType))
		if data, err := os.ReadFile(customPath); err == nil {
			return string(data), nil
		}
	}

	// 2. Check embedded templates compiled inside the Gyrus binary
	embedPath := fmt.Sprintf("templates/%s.md", conceptType)
	if data, err := embeddedTemplates.ReadFile(embedPath); err == nil {
		return string(data), nil
	}

	return "", fmt.Errorf("schema template not found for concept type '%s'", conceptType)
}
