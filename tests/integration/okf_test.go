package integration_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/domain/okf"
	graphokf "github.com/armckinney/gyrus/internal/provider/graph/okf"
	indexokf "github.com/armckinney/gyrus/internal/provider/index/okf"
	searchokf "github.com/armckinney/gyrus/internal/provider/search/okf"
	"github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func okfConcept(id string) gyrus.Concept {
	return gyrus.Concept{ID: id, Title: "Keyword guide " + id, Tenant: "root", Category: gyrus.CategoryTechnical, Type: gyrus.TypeGuide, Version: 1, Status: "active", Scope: "reference", Content: "Unicode café and keyword content", Tags: []string{"go"}, Name: "root/reference/" + id}
}
func TestOKFSearchAndAuthoritativeEdits(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store, err := localfs.NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 61; i++ {
		if _, err := store.Create(ctx, okfConcept(fmt.Sprintf("guide-%02d", i))); err != nil {
			t.Fatal(err)
		}
	}
	search, err := searchokf.NewSearchProvider(store)
	if err != nil {
		t.Fatal(err)
	}
	index, err := indexokf.NewIndexer(store)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		query       string
		limit, want int
	}{{"keyword", 0, 10}, {"", 0, 61}, {"", 7, 7}, {"keyword", 15, 15}, {"???", 0, 0}, {"CAFÉ unknown", 0, 10}} {
		results, err := index.Search(ctx, gyrus.SearchQuery{Query: tc.query, MaxResults: tc.limit})
		if err != nil || len(results) != tc.want {
			t.Fatalf("query=%q limit=%d: count=%d err=%v", tc.query, tc.limit, len(results), err)
		}
	}
	results, err := search.Search(ctx, gyrus.SearchQuery{Query: "root/reference/guide-60"})
	if err != nil || len(results) == 0 || results[0].Concept.ID != "guide-60" {
		t.Fatalf("exact-ID ranking: %v %v", results, err)
	}
	doc, err := store.Get(ctx, "root/reference/guide-60")
	if err != nil {
		t.Fatal(err)
	}
	doc.Content = "manually-edited-marker"
	data, _ := okf.SerializeMarkdown(&doc)
	if err := os.WriteFile(doc.FilePath, data, 0644); err != nil {
		t.Fatal(err)
	}
	results, err = search.Search(ctx, gyrus.SearchQuery{Query: "manually-edited-marker"})
	if err != nil || len(results) != 1 {
		t.Fatalf("external edit: %v %v", results, err)
	}
	report, err := store.Scan(ctx, func(gyrus.Concept) error { return nil })
	if err != nil || report.IndexedFiles != 61 || report.UnchangedFiles != 0 {
		t.Fatalf("sync: %+v %v", report, err)
	}
	if _, err := os.Stat(filepath.Join(root, "index.db")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected database: %v", err)
	}
	if err := index.Remove(ctx, "root/reference/guide-60"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, "root/reference/guide-60"); err != nil {
		t.Fatal("index remove changed storage", err)
	}
	if err := os.WriteFile(filepath.Join(root, "concepts/root/reference/broken.md"), []byte("not an OKF concept"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := search.Search(ctx, gyrus.SearchQuery{Query: "keyword", Filter: gyrus.SearchFilter{Tenant: "other"}}); err == nil {
		t.Fatal("invalid collection was hidden by filter")
	}
	report, err = store.Scan(ctx, func(gyrus.Concept) error { return nil })
	if err != nil || report.ScannedFiles != 62 || len(report.Errors) != 1 {
		t.Fatalf("invalid record report: %+v %v", report, err)
	}
}

func TestOKFGraphPersistenceAndDependencyExpansion(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store, _ := localfs.NewStore(root)
	source := okfConcept("project-task")
	source.Project = "demo"
	source.Scope = "project"
	source.Name = "root/projects/demo/project-task"
	source.Links = []gyrus.ConceptLink{{Concept: "root/reference/shared-guide", Rel: gyrus.RelDependsOn}}
	for _, doc := range []gyrus.Concept{source, okfConcept("shared-guide"), okfConcept("typed-guide")} {
		if _, err := store.Create(ctx, doc); err != nil {
			t.Fatal(err)
		}
	}
	graph, _ := graphokf.NewGraphStore(store)
	if err := graph.UpsertEdges(ctx, []gyrus.ConceptEdge{{FromConceptID: source.Name, ToConceptID: "root/reference/shared-guide", RelationshipType: gyrus.RelDependsOn}}); err != nil {
		t.Fatal(err)
	}
	doc, _ := store.Get(ctx, source.Name)
	if doc.Version != 1 {
		t.Fatal("duplicate legacy edge changed version")
	}
	if err := graph.UpsertEdges(ctx, []gyrus.ConceptEdge{{FromConceptID: source.Name, ToConceptID: "root/reference/typed-guide", RelationshipType: gyrus.RelDependsOn}, {FromConceptID: source.Name, ToConceptID: "root/reference/future-concept", RelationshipType: gyrus.RelImplements}}); err != nil {
		t.Fatal(err)
	}
	doc, _ = store.Get(ctx, source.Name)
	if doc.Version != 2 || len(doc.Links) != 3 {
		t.Fatalf("batch patch: %+v", doc)
	}
	fresh, _ := localfs.NewStore(root)
	restarted, _ := graphokf.NewGraphStore(fresh)
	edges, err := restarted.Neighbors(ctx, source.Name, gyrus.EdgeFilter{})
	if err != nil || len(edges) != 3 {
		t.Fatalf("restart: %v %v", edges, err)
	}
	incoming, _ := restarted.Neighbors(ctx, "root/reference/future-concept", gyrus.EdgeFilter{Direction: "incoming"})
	if len(incoming) != 1 {
		t.Fatal(incoming)
	}
	search, _ := searchokf.NewSearchProvider(fresh)
	index, _ := indexokf.NewIndexer(fresh)
	engine := lifecycle.NewEngineWithProject(fresh, search, index, restarted, root, "demo")
	contextText, err := engine.SuggestContext(ctx, "project-task", "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(contextText, "typed-guide") {
		t.Fatal("typed dependency missing from context")
	}
	if err := restarted.DeleteEdges(ctx, source.Name, "root/reference/shared-guide", gyrus.RelDependsOn); err != nil {
		t.Fatal(err)
	}
	doc, _ = fresh.Get(ctx, source.Name)
	if len(doc.Links) != 2 {
		t.Fatal("legacy dependency survived unlink")
	}
	version := doc.Version
	if err := restarted.DeleteEdges(ctx, source.Name, "root/reference/shared-guide", gyrus.RelDependsOn); err != nil {
		t.Fatal(err)
	}
	doc, _ = fresh.Get(ctx, source.Name)
	if doc.Version != version {
		t.Fatal("absent unlink changed version")
	}
	if err := restarted.UpsertEdges(ctx, []gyrus.ConceptEdge{{FromConceptID: "root/reference/typed-guide", ToConceptID: source.Name, RelationshipType: gyrus.RelDependsOn}, {FromConceptID: "root/reference/typed-guide", ToConceptID: "root/reference/typed-guide", RelationshipType: gyrus.RelMitigates}}); err != nil {
		t.Fatal(err)
	}
	paths, err := restarted.Traverse(ctx, gyrus.GraphQuery{StartID: source.Name, MaxDepth: 8})
	if err != nil || len(paths) != 1 || len(paths[0].Nodes) != 3 || len(paths[0].Edges) != 4 {
		t.Fatalf("cycle traversal: %+v %v", paths, err)
	}
	if _, err := restarted.Neighbors(ctx, source.Name, gyrus.EdgeFilter{Direction: "sideways"}); err == nil {
		t.Fatal("invalid direction accepted")
	}
	if err := fresh.Delete(ctx, "root/reference/typed-guide"); err != nil {
		t.Fatal(err)
	}
	edges, err = restarted.Neighbors(ctx, "root/reference/typed-guide", gyrus.EdgeFilter{Direction: "incoming"})
	if err != nil || len(edges) != 1 {
		t.Fatalf("dangling incoming edge: %v %v", edges, err)
	}
}
func contains(s, part string) bool { return strings.Contains(s, part) }

func TestOKFFiltersScopeAndRanking(t *testing.T) {
	ctx := context.Background()
	store, _ := localfs.NewStore(t.TempDir())
	docs := []gyrus.Concept{okfConcept("reference-a"), okfConcept("reference-b"), okfConcept("project-a"), okfConcept("other-project")}
	docs[0].Tags = []string{"go"}
	docs[0].Title = "ranking needle"
	docs[0].Content = "plain"
	docs[1].Tags = []string{"golang"}
	docs[1].Tenant = "other"
	docs[1].Status = "draft"
	docs[1].Type = gyrus.TypeSpecification
	docs[1].Category = gyrus.CategoryArchitecture
	docs[1].Title = "plain"
	docs[1].Content = "ranking needle"
	docs[2].Project, docs[2].Scope = "demo", "project"
	docs[3].Project, docs[3].Scope = "other", "project"
	for _, doc := range docs {
		if _, err := store.Create(ctx, doc); err != nil {
			t.Fatal(err)
		}
	}
	idx, _ := indexokf.NewIndexer(store)
	for _, tc := range []struct {
		filter gyrus.SearchFilter
		want   int
	}{
		{gyrus.SearchFilter{}, 2}, {gyrus.SearchFilter{Project: "demo"}, 3}, {gyrus.SearchFilter{Scope: "all"}, 4}, {gyrus.SearchFilter{Scope: "project"}, 2}, {gyrus.SearchFilter{Scope: "project", Project: "demo"}, 1}, {gyrus.SearchFilter{Scope: "reference", Project: "demo"}, 2}, {gyrus.SearchFilter{Scope: "all", Tenant: "other"}, 1}, {gyrus.SearchFilter{Scope: "all", Type: gyrus.TypeSpecification}, 1}, {gyrus.SearchFilter{Scope: "all", Category: gyrus.CategoryArchitecture}, 1}, {gyrus.SearchFilter{Scope: "all", Status: "draft"}, 1}, {gyrus.SearchFilter{Tag: "go"}, 1}, {gyrus.SearchFilter{Tag: "gol"}, 0},
	} {
		results, err := idx.Search(ctx, gyrus.SearchQuery{Filter: tc.filter})
		if err != nil || len(results) != tc.want {
			t.Fatalf("filter %+v: count=%d err=%v", tc.filter, len(results), err)
		}
	}
	results, err := idx.Search(ctx, gyrus.SearchQuery{Query: "ranking needle", MaxResults: 1})
	if err != nil || len(results) != 1 || results[0].Concept.ID != "reference-a" {
		t.Fatalf("rank before limit: %v %v", results, err)
	}
	if _, err := idx.Search(ctx, gyrus.SearchQuery{Filter: gyrus.SearchFilter{Scope: "unknown"}}); err == nil {
		t.Fatal("invalid scope accepted")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := idx.Search(cancelled, gyrus.SearchQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
}
