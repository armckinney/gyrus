package okf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"gopkg.in/yaml.v3"
)

// ParseMarkdown reads a Markdown document with YAML frontmatter into a gyrus.Document.
func ParseMarkdown(data []byte) (*gyrus.Document, error) {
	str := string(data)
	trimmed := strings.TrimSpace(str)

	if !strings.HasPrefix(trimmed, "---") {
		return nil, fmt.Errorf("invalid OKF document: missing leading frontmatter '---' delimiter")
	}

	// Find closing frontmatter delimiter
	rest := trimmed[3:]
	closeIdx := strings.Index(rest, "\n---")
	if closeIdx == -1 {
		return nil, fmt.Errorf("invalid OKF document: missing closing frontmatter '---' delimiter")
	}

	frontmatterYAML := rest[:closeIdx]
	bodyContent := strings.TrimPrefix(rest[closeIdx+4:], "\n")

	var doc gyrus.Document
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &doc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	doc.Content = strings.TrimSpace(bodyContent)
	if doc.Format == "" {
		doc.Format = "markdown"
	}

	return &doc, nil
}

// ParseJSON parses a raw JSON OKF envelope into a gyrus.Document.
func ParseJSON(data []byte) (*gyrus.Document, error) {
	var doc gyrus.Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON envelope: %w", err)
	}
	if doc.Format == "" {
		doc.Format = "json"
	}
	return &doc, nil
}

// SerializeMarkdown formats a gyrus.Document into a YAML frontmatter Markdown payload.
func SerializeMarkdown(doc *gyrus.Document) ([]byte, error) {
	yamlData, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal frontmatter to YAML: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlData)
	buf.WriteString("---\n\n")
	buf.WriteString(strings.TrimSpace(doc.Content))
	buf.WriteString("\n")

	return buf.Bytes(), nil
}
