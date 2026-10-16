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
project: gyrus
storage:
  provider: localfs
  localfs:
    path: .gyrus/concepts
tenant: test-core
search:
  provider: sqlite_fts5
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".gyrus.yaml"), []byte(cfgContent), 0644); err != nil {
		t.Fatalf("failed writing config: %v", err)
	}

	application, err := app.NewWithProject(tmpDir)
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
		{Concept: gyrus.Concept{Category: "architecture", Type: gyrus.TypeADR}},
		{Concept: gyrus.Concept{Category: "architecture", Type: gyrus.TypeSpecification}},
		{Concept: gyrus.Concept{Category: "product", Type: gyrus.TypePRD}},
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

	// Test defaultTenant
	if h.defaultTenant() != "test-core" {
		t.Errorf("expected default tenant 'test-core', got '%s'", h.defaultTenant())
	}

	// Test extractAvailableTenants
	docs := []gyrus.SearchResult{
		{Concept: gyrus.Concept{Tenant: "alpha-team"}},
		{Concept: gyrus.Concept{Tenant: "beta-team"}},
	}
	tenants := h.extractAvailableTenants(docs)
	if len(tenants) < 3 { // test-core, alpha-team, beta-team
		t.Errorf("expected at least 3 tenants, got %v", tenants)
	}

	// Test determineDocScope
	docRef := gyrus.Concept{ID: "spec-001", Category: gyrus.CategoryTechnical, Type: gyrus.TypeSpecification}
	if scope := h.determineDocScope(&docRef, nil); scope != "reference" {
		t.Errorf("expected scope 'reference', got '%s'", scope)
	}

	docProj := gyrus.Concept{ID: "prd-001", Category: gyrus.CategoryProduct, Type: gyrus.TypePRD}
	if scope := h.determineDocScope(&docProj, nil); scope != "projects/gyrus" {
		t.Errorf("expected scope 'projects/gyrus', got '%s'", scope)
	}

	// Test buildScopeData
	scopeDocs := []gyrus.SearchResult{
		{Concept: docRef},
		{Concept: docProj},
	}
	data := h.buildScopeData(scopeDocs, nil)
	if data.ReferenceCount != 1 {
		t.Errorf("expected 1 reference concept, got %d", data.ReferenceCount)
	}
	if len(data.Projects) != 1 || data.Projects[0].Name != "gyrus" || data.Projects[0].Count != 1 {
		t.Errorf("expected 1 project 'gyrus' with count 1, got %+v", data.Projects)
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
	_, err = engine.Create(ctx, gyrus.Concept{
		ID:       "spec-101",
		Title:    "Core Storage Spec",
		Category: gyrus.CategoryArchitecture,
		Type:     gyrus.TypeSpecification,
		Tenant:   "test-core",
		Scope:    "reference",
		Version:  1,
		Status:   "accepted",
		Content:  "# Storage Spec\n\nDetailed storage specification.",
	})
	if err != nil {
		t.Fatalf("failed creating reference doc: %v", err)
	}

	// Create a project document
	_, err = engine.Create(ctx, gyrus.Concept{
		ID:       "prd-101",
		Title:    "Project Feature PRD",
		Category: gyrus.CategoryProduct,
		Type:     gyrus.TypePRD,
		Tenant:   "test-core",
		Version:  1,
		Status:   "draft",
		Content:  "# Feature PRD\n\nDetailed feature requirement.",
	})
	if err != nil {
		t.Fatalf("failed creating project doc: %v", err)
	}

	// 1. Test GET /concepts renders Scopes and tenant selector, and does NOT render Categories nav group
	req := httptest.NewRequest(http.MethodGet, "/concepts", nil)
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
	if !strings.Contains(body, "Projects") {
		t.Errorf("expected sidebar to contain 'Projects'")
	}
	if !strings.Contains(body, "tenant-select") {
		t.Errorf("expected header to contain 'tenant-select'")
	}
	if strings.Contains(body, `<div class="nav-group-title">Categories</div>`) {
		t.Errorf("expected Categories section to be removed from sidebar")
	}
	if strings.Contains(body, `<div class="nav-group-title">Document Types</div>`) {
		t.Errorf("expected Document Types section to be removed from sidebar")
	}
	if !strings.Contains(body, `type-filter-bar`) {
		t.Errorf("expected type filter bar to be present in the main content area")
	}
	if !strings.Contains(body, `type-filter-pill`) {
		t.Errorf("expected type filter pills to be present in the main content area")
	}

	// 2. Test GET /concepts?scope=reference
	reqRef := httptest.NewRequest(http.MethodGet, "/concepts?scope=reference", nil)
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
		t.Errorf("did not expect project document 'prd-101' in scope=reference filter")
	}

	// 3. Test GET /concepts?scope=projects/gyrus
	reqWS := httptest.NewRequest(http.MethodGet, "/concepts?scope=projects/gyrus", nil)
	rrWS := httptest.NewRecorder()
	h.Docs(rrWS, reqWS)

	if rrWS.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for scope=projects/gyrus, got %d", rrWS.Code)
	}
	wsBody := rrWS.Body.String()
	if !strings.Contains(wsBody, "Project: gyrus") {
		t.Errorf("expected page title 'Project: gyrus'")
	}
	if !strings.Contains(wsBody, "prd-101") {
		t.Errorf("expected project document 'prd-101' to be listed")
	}
	if strings.Contains(wsBody, "spec-101") {
		t.Errorf("did not expect reference document 'spec-101' in project filter")
	}

	// 4. Test GET /concepts?scope=reference&type=specification (combined scope + type filter)
	reqScopeType := httptest.NewRequest(http.MethodGet, "/concepts?scope=reference&type=specification", nil)
	rrScopeType := httptest.NewRecorder()
	h.Docs(rrScopeType, reqScopeType)

	if rrScopeType.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for scope+type filter, got %d", rrScopeType.Code)
	}
	scopeTypeBody := rrScopeType.Body.String()
	if !strings.Contains(scopeTypeBody, "Scope: Reference · specification") {
		t.Errorf("expected page title 'Scope: Reference · specification'")
	}
	if !strings.Contains(scopeTypeBody, "spec-101") {
		t.Errorf("expected 'spec-101' in scope=reference&type=specification results")
	}
	if strings.Contains(scopeTypeBody, "prd-101") {
		t.Errorf("did not expect 'prd-101' in scope=reference&type=specification results")
	}

	// 5. Test GET /api/graph?tenant=test-core
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
