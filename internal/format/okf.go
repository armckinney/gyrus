package format

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"gopkg.in/yaml.v3"
)

// OKFFormat implements StorageFormat for Markdown documents with YAML frontmatter.
type OKFFormat struct{}

// NewOKFFormat returns a new OKFFormat serializer instance.
func NewOKFFormat() *OKFFormat {
	return &OKFFormat{}
}

// Serialize converts a gyrus.Document into YAML frontmatter + Markdown bytes.
func (f *OKFFormat) Serialize(doc *gyrus.Document) ([]byte, error) {
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

// Deserialize parses a YAML frontmatter + Markdown payload into a gyrus.Document.
func (f *OKFFormat) Deserialize(data []byte) (*gyrus.Document, error) {
	str := string(data)
	trimmed := strings.TrimSpace(str)

	if !strings.HasPrefix(trimmed, "---") {
		return nil, fmt.Errorf("invalid OKF document: missing leading frontmatter '---' delimiter")
	}

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
