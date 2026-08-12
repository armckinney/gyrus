package okf

import (
	"github.com/armckinney/gyrus/internal/format"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

var okfFmt = format.NewOKFFormat()
var jsonFmt = format.NewRelationalFormat()

// ParseMarkdown reads a Markdown document with YAML frontmatter into a gyrus.Document.
func ParseMarkdown(data []byte) (*gyrus.Document, error) {
	return okfFmt.Deserialize(data)
}

// ParseJSON parses a raw JSON OKF envelope into a gyrus.Document.
func ParseJSON(data []byte) (*gyrus.Document, error) {
	return jsonFmt.Deserialize(data)
}

// SerializeMarkdown formats a gyrus.Document into a YAML frontmatter Markdown payload.
func SerializeMarkdown(doc *gyrus.Document) ([]byte, error) {
	return okfFmt.Serialize(doc)
}
