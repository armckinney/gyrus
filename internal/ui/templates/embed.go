package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
)

//go:embed *.html
var templateFS embed.FS

// Manager coordinates parsing and rendering of embedded HTML templates.
type Manager struct {
	fullTemplates    map[string]*template.Template
	partialTemplates map[string]*template.Template
}

// NewManager parses and caches layout and view templates.
func NewManager() (*Manager, error) {
	views := []string{"index.html", "doc.html", "search_results.html", "graph.html"}
	full := make(map[string]*template.Template)
	partial := make(map[string]*template.Template)

	for _, v := range views {
		// Full page template combining layout.html with the view template
		tFull, err := template.New("layout.html").ParseFS(templateFS, "layout.html", v)
		if err != nil {
			return nil, fmt.Errorf("failed parsing full template for '%s': %w", v, err)
		}
		full[v] = tFull

		// Partial template for HTMX fragments
		tPart, err := template.New(v).ParseFS(templateFS, v)
		if err != nil {
			return nil, fmt.Errorf("failed parsing partial template for '%s': %w", v, err)
		}
		partial[v] = tPart
	}

	return &Manager{
		fullTemplates:    full,
		partialTemplates: partial,
	}, nil
}

// Render renders either a full page or a partial HTML fragment depending on isHTMX.
func (m *Manager) Render(w io.Writer, view string, data any, isHTMX bool) error {
	if isHTMX {
		t, ok := m.partialTemplates[view]
		if !ok {
			return fmt.Errorf("partial template '%s' not found", view)
		}
		return t.ExecuteTemplate(w, "content", data)
	}

	t, ok := m.fullTemplates[view]
	if !ok {
		return fmt.Errorf("full template '%s' not found", view)
	}
	return t.Execute(w, data)
}
