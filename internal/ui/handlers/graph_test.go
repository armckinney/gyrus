package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestGraph_Page(t *testing.T) {
	h, cleanup := setupDocsTestEnvironment(t)
	defer cleanup()

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
}

func TestGraph_APIGraph_JSONContract(t *testing.T) {
	h, cleanup := setupDocsTestEnvironment(t)
	defer cleanup()

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
}
