package okf

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed templates/*.md
var embeddedTemplates embed.FS

// GetTemplate retrieves the schema template for a document type.
// It prioritizes custom templates in customSchemasDir, falling back to embedded binary templates.
func GetTemplate(docType string, customSchemasDir string) (string, error) {
	// 1. Check custom user schemas directory in project repository if specified
	if customSchemasDir != "" {
		customPath := filepath.Join(customSchemasDir, fmt.Sprintf("%s.md", docType))
		if data, err := os.ReadFile(customPath); err == nil {
			return string(data), nil
		}
	}

	// 2. Check embedded templates compiled inside the Gyrus binary
	embedPath := fmt.Sprintf("templates/%s.md", docType)
	if data, err := embeddedTemplates.ReadFile(embedPath); err == nil {
		return string(data), nil
	}

	// 3. Fallback default template generator
	fallback := fmt.Sprintf(`---
id: %s-001
title: Example Title
category: architecture
type: %s
owner_group: platform
version: 1
status: draft
tags:
  - example
dependencies: []
---

# Title

Document description body here.
`, docType, docType)

	return fallback, nil
}
