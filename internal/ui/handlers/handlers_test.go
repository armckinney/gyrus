package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/ui/agent"
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

func TestHandlers_LifecycleAndHelpers(t *testing.T) {
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

	if h.SessionStore() == nil {
		t.Errorf("expected session store to be initialized")
	}

	// Test SetRunner
	mockRunner := agent.NewGenericRunner("Mock", "mock", "echo", tmpDir)
	h.SetRunner(mockRunner)
	if h.uiClientName() != "mock" {
		t.Errorf("expected client name 'mock', got '%s'", h.uiClientName())
	}

	// Test storageBackend
	if h.storageBackend() != "localfs" {
		t.Errorf("expected storage backend 'localfs', got '%s'", h.storageBackend())
	}

	// Test extractSnippet helper
	longContent := "# Title\n\nFirst line of content that has some descriptive text for the document snippet preview.\nAnother paragraph follows."
	snippet := extractSnippet(longContent)
	if snippet == "" || snippet[0] == '#' {
		t.Errorf("unexpected snippet output: %s", snippet)
	}

	// Test buildTaxonomyCounts helper
	docs := []gyrus.SearchResult{
		{Document: gyrus.Document{Category: "architecture", Type: gyrus.TypeADR}},
		{Document: gyrus.Document{Category: "architecture", Type: gyrus.TypeSpecification}},
		{Document: gyrus.Document{Category: "product", Type: gyrus.TypePRD}},
	}
	catCounts, typeCounts := h.buildTaxonomyCounts(docs)
	if len(catCounts) != 2 {
		t.Errorf("expected 2 categories, got %d", len(catCounts))
	}
	if len(typeCounts) != 3 {
		t.Errorf("expected 3 types, got %d", len(typeCounts))
	}
}

func TestHandlers_ScopesAndTenants(t *testing.T) {
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

	// Test defaultOwnerGroup
	if h.defaultOwnerGroup() != "test-core" {
		t.Errorf("expected default owner group 'test-core', got '%s'", h.defaultOwnerGroup())
	}

	// Test extractAvailableTenants
	docs := []gyrus.SearchResult{
		{Document: gyrus.Document{OwnerGroup: "alpha-team"}},
		{Document: gyrus.Document{OwnerGroup: "beta-team"}},
	}
	tenants := h.extractAvailableTenants(docs)
	if len(tenants) < 3 { // test-core, alpha-team, beta-team
		t.Errorf("expected at least 3 tenants, got %v", tenants)
	}

	// Test determineDocScope
	docRef := gyrus.Document{ID: "spec-001", Category: gyrus.CategoryTechnical, Type: gyrus.TypeSpecification}
	if scope := h.determineDocScope(&docRef, nil); scope != "reference" {
		t.Errorf("expected scope 'reference', got '%s'", scope)
	}

	docWS := gyrus.Document{ID: "prd-001", Category: gyrus.CategoryProduct, Type: gyrus.TypePRD}
	if scope := h.determineDocScope(&docWS, nil); scope != "workspaces/main" {
		t.Errorf("expected scope 'workspaces/main', got '%s'", scope)
	}

	// Test buildScopeData
	scopeDocs := []gyrus.SearchResult{
		{Document: docRef},
		{Document: docWS},
	}
	data := h.buildScopeData(scopeDocs, nil)
	if data.ReferenceCount != 1 {
		t.Errorf("expected 1 reference document, got %d", data.ReferenceCount)
	}
	if len(data.Workspaces) != 1 || data.Workspaces[0].Name != "main" || data.Workspaces[0].Count != 1 {
		t.Errorf("expected 1 workspace 'main' with count 1, got %+v", data.Workspaces)
	}
}

func TestHandlers_DocsPage_ScopesAndTenants_HTTP(t *testing.T) {
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

	// Create a reference document
	_, err = engine.Create(ctx, gyrus.Document{
		ID:         "spec-101",
		Title:      "Core Storage Spec",
		Category:   gyrus.CategoryArchitecture,
		Type:       gyrus.TypeSpecification,
		OwnerGroup: "test-core",
		Version:    1,
		Status:     "accepted",
		Content:    "# Storage Spec\n\nDetailed storage specification.",
	})
	if err != nil {
		t.Fatalf("failed creating reference doc: %v", err)
	}

	// Create a workspace document
	_, err = engine.Create(ctx, gyrus.Document{
		ID:         "prd-101",
		Title:      "Workspace Feature PRD",
		Category:   gyrus.CategoryProduct,
		Type:       gyrus.TypePRD,
		OwnerGroup: "test-core",
		Version:    1,
		Status:     "draft",
		Content:    "# Feature PRD\n\nDetailed feature requirement.",
	})
	if err != nil {
		t.Fatalf("failed creating workspace doc: %v", err)
	}

	// 1. Test GET /docs renders Scopes and tenant selector, and does NOT render Categories nav group
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rr := httptest.NewRecorder()
	h.Docs(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	body := rr.Body.String()

	if !strings.Contains(body, "Scopes") {
		t.Errorf("expected sidebar to contain 'Scopes'")
	}
	if !strings.Contains(body, "Reference") {
		t.Errorf("expected sidebar to contain 'Reference'")
	}
	if !strings.Contains(body, "Workspaces") {
		t.Errorf("expected sidebar to contain 'Workspaces'")
	}
	if !strings.Contains(body, "tenant-select") {
		t.Errorf("expected header to contain 'tenant-select'")
	}
	if strings.Contains(body, `<div class="nav-group-title">Categories</div>`) {
		t.Errorf("expected Categories section to be removed from sidebar")
	}

	// 2. Test GET /docs?scope=reference
	reqRef := httptest.NewRequest(http.MethodGet, "/docs?scope=reference", nil)
	rrRef := httptest.NewRecorder()
	h.Docs(rrRef, reqRef)

	if rrRef.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for scope=reference, got %d", rrRef.Code)
	}
	refBody := rrRef.Body.String()
	if !strings.Contains(refBody, "Scope: Reference") {
		t.Errorf("expected page title 'Scope: Reference'")
	}
	if !strings.Contains(refBody, "spec-101") {
		t.Errorf("expected reference document 'spec-101' to be listed")
	}
	if strings.Contains(refBody, "prd-101") {
		t.Errorf("did not expect workspace document 'prd-101' in scope=reference filter")
	}

	// 3. Test GET /docs?scope=workspaces/main
	reqWS := httptest.NewRequest(http.MethodGet, "/docs?scope=workspaces/main", nil)
	rrWS := httptest.NewRecorder()
	h.Docs(rrWS, reqWS)

	if rrWS.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for scope=workspaces/main, got %d", rrWS.Code)
	}
	wsBody := rrWS.Body.String()
	if !strings.Contains(wsBody, "Workspace: main") {
		t.Errorf("expected page title 'Workspace: main'")
	}
	if !strings.Contains(wsBody, "prd-101") {
		t.Errorf("expected workspace document 'prd-101' to be listed")
	}
	if strings.Contains(wsBody, "spec-101") {
		t.Errorf("did not expect reference document 'spec-101' in workspace filter")
	}

	// 4. Test GET /api/graph?tenant=test-core
	reqGraph := httptest.NewRequest(http.MethodGet, "/api/graph?tenant=test-core", nil)
	rrGraph := httptest.NewRecorder()
	h.APIGraph(rrGraph, reqGraph)

	if rrGraph.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/graph, got %d", rrGraph.Code)
	}
	if !strings.Contains(rrGraph.Body.String(), "spec-101") {
		t.Errorf("expected /api/graph to contain 'spec-101'")
	}
}
