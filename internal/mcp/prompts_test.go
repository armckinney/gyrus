package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/okf"
	mcpapi "github.com/mark3labs/mcp-go/mcp"
)

func TestPrepareADRUsesActiveSchema(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".gyrus.yaml"), []byte("storage:\n  provider: localfs\n  root: .gyrus\n"), 0644); err != nil {
		t.Fatal(err)
	}
	application, err := app.NewWithProject(root)
	if err != nil {
		t.Fatal(err)
	}
	server, err := newServerFromApp(application)
	if err != nil {
		t.Fatal(err)
	}
	req := mcpapi.GetPromptRequest{}
	req.Params.Arguments = map[string]string{"title": "Decision: \"quoted\"\nsecond line"}
	for _, custom := range []bool{false, true} {
		if custom {
			template := "---\ntitle: Placeholder\nid: custom-adr\ntype: adr\ncategory: architecture\nversion: 1\nstatus: proposed\ntenant: root\nlinks: []\n---\n\n# Custom ADR marker\n"
			if err := server.engine.SaveSchema(context.Background(), "adr", template); err != nil {
				t.Fatal(err)
			}
		}
		result, err := server.handlePrepareADRPrompt(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		content := result.Messages[0].Content.(mcpapi.TextContent).Text
		if strings.Contains(content, "dependencies:") {
			t.Fatal("obsolete dependencies field in ADR prompt")
		}
		if !strings.HasPrefix(content, "---\ntitle:") {
			t.Fatal("prompt does not use canonical frontmatter order")
		}
		if custom {
			doc, err := okf.ParseMarkdown([]byte(content))
			if err != nil {
				t.Fatal(err)
			}
			if err := okf.Validate(doc); err != nil {
				t.Fatal(err)
			}
			if doc.Title != req.Params.Arguments["title"] || doc.ID != "custom-adr" || !strings.Contains(doc.Content, "Custom ADR marker") {
				t.Fatalf("active template not preserved: %+v", doc)
			}
		} else if !strings.Contains(content, "## Alternatives Considered") {
			t.Fatal("prompt did not use the shared embedded ADR template")
		}
	}
}
