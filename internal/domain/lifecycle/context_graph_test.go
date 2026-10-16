package lifecycle_test

import (
	"context"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

type projectSearch struct{ doc gyrus.Concept }

func (s projectSearch) Search(context.Context, gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	return []gyrus.SearchResult{{Concept: s.doc}}, nil
}

type dependencyGraph struct{ edges []gyrus.ConceptEdge }

func (g dependencyGraph) UpsertEdges(context.Context, []gyrus.ConceptEdge) error { return nil }
func (g dependencyGraph) DeleteEdges(context.Context, string, string, gyrus.RelationshipType) error {
	return nil
}
func (g dependencyGraph) Traverse(context.Context, gyrus.GraphQuery) ([]gyrus.GraphPath, error) {
	return nil, nil
}
func (g dependencyGraph) Neighbors(context.Context, string, gyrus.EdgeFilter) ([]gyrus.ConceptEdge, error) {
	return g.edges, nil
}

func TestSuggestContextExpandsConfiguredGraph(t *testing.T) {
	source := gyrus.Concept{ID: "source", Name: "root/projects/demo/source", Tenant: "root", Scope: "project", Project: "demo", Title: "Source", Content: "projectmarker"}
	target := gyrus.Concept{ID: "target", Name: "root/reference/target", Tenant: "root", Scope: "reference", Title: "Target", Content: "databasedependencymarker"}
	store := &mockStoreWithSearch{docs: map[string]gyrus.Concept{source.Name: source, target.Name: target}}
	graph := dependencyGraph{edges: []gyrus.ConceptEdge{{FromConceptID: source.Name, ToConceptID: target.Name, RelationshipType: gyrus.RelDependsOn}}}
	engine := lifecycle.NewEngineWithProject(store, projectSearch{source}, nil, graph, "", "demo")
	content, err := engine.SuggestContext(context.Background(), "projectmarker", "", 2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, target.Content) || !strings.Contains(content, target.Name) {
		t.Fatalf("database-only dependency missing: %s", content)
	}
	if len(source.Links) != 0 {
		t.Fatal("fixture must not contain document-owned links")
	}
}
