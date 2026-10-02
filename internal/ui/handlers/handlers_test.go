package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/ui/markdown"
	"github.com/armckinney/gyrus/internal/ui/templates"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func setupTestApp(t *testing.T) (*app.App, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "gyrus-ui-test-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}

	cfgContent := `
version: 1
storage:
  provider: localfs
  localfs:
    path: .gyrus/docs
default_owner_group: test-core
search:
  provider: sqlite_fts5
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".gyrus.yaml"), []byte(cfgContent), 0644); err != nil {
		t.Fatalf("failed writing config: %v", err)
	}

	application, err := app.NewWithWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("failed initializing app: %v", err)
	}

	return application, tmpDir
}

func TestHandlers_EndToEnd(t *testing.T) {
	application, tmpDir := setupTestApp(t)
	defer os.RemoveAll(tmpDir)

	engine, err := application.Engine()
	if err != nil {
		t.Fatalf("failed initializing engine: %v", err)
	}

	tmplMgr, err := templates.NewManager()
	if err != nil {
		t.Fatalf("failed initializing template manager: %v", err)
	}

	renderer := markdown.NewRenderer()
	h := New(application, engine, tmplMgr, renderer)

	ctx := context.Background()

	// 1. Create test documents
	doc1 := gyrus.Document{
		ID:         "adr-001-architecture",
		Title:      "Storage Engine Selection",
		Category:   "architecture",
		Type:       gyrus.TypeADR,
		Status:     "accepted",
		Version:    1,
		OwnerGroup: "test-core",
		Tags:       []string{"storage", "sqlite"},
		Content:    "# Decision\nWe choose SQLite and PostgreSQL for unified persistence.\n\n| Engine | Score |\n|---|---|\n| SQLite | 9.5 |",
	}
	if _, err := engine.Create(ctx, doc1); err != nil {
		t.Fatalf("failed creating doc1: %v", err)
	}

	doc2 := gyrus.Document{
		ID:         "prd-001-roadmap",
		Title:      "Product Roadmap Specification",
		Category:   "product",
		Type:       gyrus.TypePRD,
		Status:     "active",
		Version:    1,
		OwnerGroup: "test-core",
		Tags:       []string{"roadmap"},
		Content:    "# Roadmap\nPhase 2.9 includes Web UI development.",
	}
	if _, err := engine.Create(ctx, doc2); err != nil {
		t.Fatalf("failed creating doc2: %v", err)
	}

	// 2. Link documents
	if err := engine.Link(ctx, "prd-001-roadmap", "adr-001-architecture", gyrus.RelDependsOn); err != nil {
		t.Fatalf("failed linking docs: %v", err)
	}

	// 3. Test Home (GET /docs)
	t.Run("Home full page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs", nil)
		rec := httptest.NewRecorder()

		h.Home(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Storage Engine Selection") {
			t.Errorf("expected body to contain doc title, got %s", body)
		}
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("expected full HTML page layout")
		}
	})

	t.Run("Home HTMX partial", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		h.Home(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("expected HTMX partial without full HTML shell")
		}
		if !strings.Contains(body, "Storage Engine Selection") {
			t.Errorf("expected partial to contain doc title")
		}
	})

	// 4. Test Doc view (GET /docs/{id})
	t.Run("Doc view success with rendered markdown and relationships", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs/prd-001-roadmap", nil)
		rec := httptest.NewRecorder()

		h.Doc(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Product Roadmap Specification") {
			t.Errorf("missing title in doc view")
		}
		if !strings.Contains(body, "Phase 2.9 includes Web UI development.") {
			t.Errorf("missing rendered content")
		}
		if !strings.Contains(body, "adr-001-architecture") {
			t.Errorf("missing outgoing relationship link to adr-001-architecture")
		}
	})

	t.Run("Doc view 404 for missing document", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs/nonexistent-id", nil)
		rec := httptest.NewRecorder()

		h.Doc(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404 for missing doc, got %d", rec.Code)
		}
	})

	// 5. Test Search (GET /search?q=roadmap)
	t.Run("Search endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/search?q=roadmap", nil)
		rec := httptest.NewRecorder()

		h.Search(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Product Roadmap Specification") {
			t.Errorf("expected search result to contain match")
		}
	})

	// 6. Test Graph page (GET /graph)
	t.Run("Graph page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/graph?focus=prd-001-roadmap", nil)
		rec := httptest.NewRecorder()

		h.Graph(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Knowledge Graph Explorer") {
			t.Errorf("expected graph page title")
		}
	})

	// 7. Test API Graph JSON endpoint (GET /api/graph)
	t.Run("APIGraph JSON contract", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/graph", nil)
		rec := httptest.NewRecorder()

		h.APIGraph(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if rec.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content-type, got %s", rec.Header().Get("Content-Type"))
		}

		var payload GraphDataJSON
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed decoding json: %v", err)
		}

		if len(payload.Nodes) != 2 {
			t.Errorf("expected 2 nodes, got %d", len(payload.Nodes))
		}
		if len(payload.Edges) != 1 {
			t.Fatalf("expected 1 edge, got %d", len(payload.Edges))
		}

		edge := payload.Edges[0]
		if edge.Source != "prd-001-roadmap" || edge.Target != "adr-001-architecture" || edge.RelationshipType != string(gyrus.RelDependsOn) {
			t.Errorf("unexpected edge structure: %+v", edge)
		}
	})
}
