package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestGraphScopeBoundaryAndStableOrdering(t *testing.T) {
	h, cleanup := setupDocsTestEnvironment(t)
	defer cleanup()
	ctx := context.Background()
	for _, doc := range []gyrus.Concept{
		{ID: "b", Title: "Visible B", Scope: "reference"},
		{ID: "a", Title: "Visible A", Scope: "reference"},
		{ID: "hidden", Title: "Private project title", Scope: "project", Project: "other"},
	} {
		doc.Tenant, doc.Type, doc.Category, doc.Status, doc.Version = "boundary", gyrus.TypeGuide, gyrus.CategoryTechnical, "draft", 1
		if _, err := h.engine.Create(ctx, doc); err != nil {
			t.Fatal(err)
		}
	}
	const a, b = "boundary/reference/a", "boundary/reference/b"
	const hidden, shared = "boundary/projects/other/hidden", "boundary/reference/missing"
	for _, edge := range []gyrus.ConceptEdge{
		{FromConceptID: a, ToConceptID: hidden, RelationshipType: gyrus.RelDependsOn},
		{FromConceptID: hidden, ToConceptID: "boundary/projects/other/onward", RelationshipType: gyrus.RelDependsOn},
		{FromConceptID: a, ToConceptID: a, RelationshipType: gyrus.RelMitigates},
		{FromConceptID: a, ToConceptID: b, RelationshipType: gyrus.RelImplements},
		{FromConceptID: b, ToConceptID: a, RelationshipType: gyrus.RelSupersedes},
		{FromConceptID: a, ToConceptID: shared, RelationshipType: gyrus.RelDependsOn},
		{FromConceptID: b, ToConceptID: shared, RelationshipType: gyrus.RelDependsOn},
	} {
		if err := h.engine.Link(ctx, edge.FromConceptID, edge.ToConceptID, edge.RelationshipType); err != nil {
			t.Fatal(err)
		}
	}
	var baseline string
	const repeatCount = 32
	for i := 0; i < repeatCount; i++ {
		rec := httptest.NewRecorder()
		h.APIGraph(rec, httptest.NewRequest(http.MethodGet, "/api/graph?tenant=boundary", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("response: %d %s", rec.Code, rec.Body.String())
		}
		var graph GraphDataJSON
		if err := json.Unmarshal(rec.Body.Bytes(), &graph); err != nil {
			t.Fatal(err)
		}
		if len(graph.Nodes) != 4 || len(graph.Edges) != 6 {
			t.Fatalf("scope expanded through ghost: %+v", graph)
		}
		if !sort.SliceIsSorted(graph.Nodes, func(i, j int) bool { return graph.Nodes[i].ID < graph.Nodes[j].ID }) {
			t.Fatal("nodes not sorted")
		}
		for _, node := range graph.Nodes {
			if node.ID == hidden && (node.Title != hidden || node.Category != "external") {
				t.Fatalf("filtered metadata exposed: %+v", node)
			}
		}
		for _, edge := range graph.Edges {
			if edge.Source != a && edge.Source != b {
				t.Fatalf("ghost expanded: %+v", edge)
			}
		}
		if i == 0 {
			baseline = rec.Body.String()
		} else if rec.Body.String() != baseline {
			t.Fatal("identical requests returned different graph output")
		}
	}
}

type graphTestSearch struct{ docs []gyrus.SearchResult }

func (s graphTestSearch) Search(context.Context, gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	return s.docs, nil
}

type graphTestStore struct {
	edges []gyrus.ConceptEdge
	err   error
}

func (g graphTestStore) UpsertEdges(context.Context, []gyrus.ConceptEdge) error { return nil }
func (g graphTestStore) DeleteEdges(context.Context, string, string, gyrus.RelationshipType) error {
	return nil
}
func (g graphTestStore) Traverse(context.Context, gyrus.GraphQuery) ([]gyrus.GraphPath, error) {
	return nil, nil
}
func (g graphTestStore) Neighbors(context.Context, string, gyrus.EdgeFilter) ([]gyrus.ConceptEdge, error) {
	return g.edges, g.err
}

func TestGraphDeduplicatesAndReportsEdgeErrors(t *testing.T) {
	doc := gyrus.Concept{Name: "root/reference/source", Title: "Source"}
	search := graphTestSearch{docs: []gyrus.SearchResult{{Concept: doc}, {Concept: doc}}}
	edge := gyrus.ConceptEdge{FromConceptID: doc.Name, ToConceptID: "root/reference/missing", RelationshipType: gyrus.RelDependsOn}
	for _, fail := range []bool{false, true} {
		graph := graphTestStore{edges: []gyrus.ConceptEdge{edge, edge}}
		if fail {
			graph.err = errors.New("edge query failed")
		}
		h := &Handlers{engine: lifecycle.NewEngine(nil, search, nil, graph, "")}
		rec := httptest.NewRecorder()
		h.APIGraph(rec, httptest.NewRequest(http.MethodGet, "/api/graph", nil))
		if fail {
			if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), `"nodes"`) {
				t.Fatalf("partial graph returned: %d %s", rec.Code, rec.Body.String())
			}
			continue
		}
		var result GraphDataJSON
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if len(result.Nodes) != 2 || len(result.Edges) != 1 {
			t.Fatalf("duplicate graph entries: %+v", result)
		}
	}
}

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
	if edge.Source != "test-core/reference/prd-001-roadmap" || edge.Target != "test-core/reference/adr-001-architecture" || edge.RelationshipType != string(gyrus.RelDependsOn) {
		t.Errorf("unexpected edge structure: %+v", edge)
	}
}
