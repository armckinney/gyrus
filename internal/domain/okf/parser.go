package okf

import (
	"github.com/armckinney/gyrus/internal/format"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

var okfFmt = format.NewOKFFormat()
var jsonFmt = format.NewRelationalFormat()

// ParseMarkdown reads a Markdown concept with YAML frontmatter into a gyrus.Concept.
func ParseMarkdown(data []byte) (*gyrus.Concept, error) {
	return okfFmt.Deserialize(data)
}

// ParseJSON parses a raw JSON OKF envelope into a gyrus.Concept.
func ParseJSON(data []byte) (*gyrus.Concept, error) {
	return jsonFmt.Deserialize(data)
}

// SerializeMarkdown formats a gyrus.Concept into a YAML frontmatter Markdown payload.
func SerializeMarkdown(doc *gyrus.Concept) ([]byte, error) {
	return okfFmt.Serialize(doc)
}
