package markdown

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// Renderer formats Markdown content into safe, styled HTML using goldmark.
type Renderer struct {
	md goldmark.Markdown
}

// NewRenderer initializes a Goldmark Markdown converter equipped with GitHub Flavored Markdown
// extensions (tables, autolinks, strikethrough, task lists) and automatic heading IDs.
func NewRenderer() *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
			extension.Linkify,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	return &Renderer{md: md}
}

// Render converts markdown source text to sanitized, template-safe HTML.
func (r *Renderer) Render(content string) (template.HTML, error) {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(content), &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
