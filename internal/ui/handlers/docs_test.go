package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/ui/markdown"
	"github.com/armckinney/gyrus/internal/ui/templates"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func setupDocsTestEnvironment(t *testing.T) (*Handlers, func()) {
	t.Helper()
	application, tmpDir := setupTestApp(t)

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

	if err := engine.Link(ctx, "prd-001-roadmap", "adr-001-architecture", gyrus.RelDependsOn); err != nil {
		t.Fatalf("failed linking docs: %v", err)
	}

	teardown := func() {
		os.RemoveAll(tmpDir)
	}
	return h, teardown
}

func TestDocs_HomeAndDocsList(t *testing.T) {
	h, cleanup := setupDocsTestEnvironment(t)
	defer cleanup()

	// 1. Full Page GET /
	t.Run("Home full page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		h.Home(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("expected full html page with doctype")
		}
		if !strings.Contains(body, "Storage Engine Selection") {
			t.Errorf("expected doc1 title in body")
		}
	})

	// 2. HTMX Partial GET /
	t.Run("Home HTMX partial", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()

		h.Home(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if strings.Contains(body, "<!DOCTYPE html>") {
			t.Errorf("expected partial fragment without doctype for HTMX request")
		}
	})

	// 3. Category Filter GET /docs?category=architecture
	t.Run("Docs category filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs?category=architecture", nil)
		rec := httptest.NewRecorder()

		h.Docs(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Storage Engine Selection") {
			t.Errorf("expected architecture doc in response")
		}
	})

	// 4. Type Filter GET /docs?type=prd
	t.Run("Docs type filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs?type=prd", nil)
		rec := httptest.NewRecorder()

		h.Docs(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Product Roadmap Specification") {
			t.Errorf("expected prd doc in response")
		}
	})
}

func TestDocs_DocView(t *testing.T) {
	h, cleanup := setupDocsTestEnvironment(t)
	defer cleanup()

	t.Run("Doc view success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs/adr-001-architecture", nil)
		rec := httptest.NewRecorder()

		h.Doc(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Storage Engine Selection") {
			t.Errorf("expected doc title")
		}
		if !strings.Contains(body, "<table>") {
			t.Errorf("expected markdown table rendering")
		}
		if !strings.Contains(body, "prd-001-roadmap") {
			t.Errorf("expected incoming relationship to be rendered")
		}
	})

	t.Run("Doc view 404 for missing document", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs/non-existent-doc", nil)
		rec := httptest.NewRecorder()

		h.Doc(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})
}

func TestDocs_Search(t *testing.T) {
	h, cleanup := setupDocsTestEnvironment(t)
	defer cleanup()

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
}
