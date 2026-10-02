package markdown

import (
	"strings"
	"testing"
)

func TestRenderer_Render(t *testing.T) {
	renderer := NewRenderer()

	t.Run("GFM table rendering", func(t *testing.T) {
		input := `
| Header 1 | Header 2 |
|----------|----------|
| Val A    | Val B    |
`
		out, err := renderer.Render(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rendered := string(out)
		if !strings.Contains(rendered, "<table>") || !strings.Contains(rendered, "<th>Header 1</th>") {
			t.Errorf("expected HTML table, got: %s", rendered)
		}
	})

	t.Run("Auto heading IDs", func(t *testing.T) {
		input := "# System Architecture Overview"
		out, err := renderer.Render(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rendered := string(out)
		if !strings.Contains(rendered, `id="system-architecture-overview"`) {
			t.Errorf("expected auto heading ID, got: %s", rendered)
		}
	})

	t.Run("Task lists", func(t *testing.T) {
		input := `
- [x] Phase 1 Complete
- [ ] Phase 2 Pending
`
		out, err := renderer.Render(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rendered := string(out)
		if !strings.Contains(rendered, `type="checkbox"`) {
			t.Errorf("expected checkbox input in tasklist, got: %s", rendered)
		}
	})

	t.Run("Code block", func(t *testing.T) {
		input := "```go\nfmt.Println(\"hello\")\n```"
		out, err := renderer.Render(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rendered := string(out)
		if !strings.Contains(rendered, "<pre><code") || !strings.Contains(rendered, "fmt.Println") {
			t.Errorf("expected pre/code block, got: %s", rendered)
		}
	})
}
