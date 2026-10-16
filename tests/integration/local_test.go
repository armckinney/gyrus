package integration_test

import (
	"context"
	"path/filepath"
	"testing"

	graphokf "github.com/armckinney/gyrus/internal/provider/graph/okf"
	indexokf "github.com/armckinney/gyrus/internal/provider/index/okf"
	searchokf "github.com/armckinney/gyrus/internal/provider/search/okf"
	storagelocal "github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies localfs storage with OKF index, graph, and search providers.
// [Execution Surface]: Real local concept files.
// [Assertions]: Creation persists, scans validate concepts, search finds keywords, and typed links persist.
// -----------------------------------------------------------------------------
func TestIntegration_LocalProfile(t *testing.T) {
	tempDir := t.TempDir()
	storageRoot := filepath.Join(tempDir, "docs")

	store, err := storagelocal.NewStore(storageRoot)
	if err != nil {
		t.Fatalf("Failed localfs initialization: %v", err)
	}

	indexer, err := indexokf.NewIndexer(store)
	if err != nil {
		t.Fatalf("Failed OKF indexer initialization: %v", err)
	}
	graphStore, err := graphokf.NewGraphStore(store)
	if err != nil {
		t.Fatal(err)
	}
	searcher, err := searchokf.NewSearchProvider(store)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	doc := gyrus.Concept{
		ID:       "standard-001",
		Title:    "Coding Standard",
		Category: gyrus.CategoryTechnical,
		Type:     gyrus.TypeStandards,
		Tenant:   "engineering",
		Version:  1,
		Status:   "active",
		Content:  "All Go code must follow standard idiomatic patterns.", Name: "engineering/reference/standard-001",
	}

	// 1. Storage Create
	ref, err := store.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Storage Create failed: %v", err)
	}
	if ref.ID != doc.ID {
		t.Errorf("Expected Ref ID %s, got %s", doc.ID, ref.ID)
	}

	// 2. Sync / Indexing
	report, err := store.Scan(ctx, func(doc gyrus.Concept) error { return indexer.Index(ctx, doc) })
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if report.IndexedFiles != 1 {
		t.Errorf("Expected 1 indexed file, got %d", report.IndexedFiles)
	}

	// 3. Search via SearchProvider
	results, err := searcher.Search(ctx, gyrus.SearchQuery{Query: "idiomatic"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].Concept.ID != doc.ID {
		t.Errorf("Expected search match for standard-001, got %v", results)
	}

	// 4. Graph Edge Link
	err = graphStore.UpsertEdges(ctx, []gyrus.ConceptEdge{{
		FromConceptID:    "engineering/reference/standard-001",
		ToConceptID:      "root/reference/adr-001",
		RelationshipType: gyrus.RelDependsOn,
	}})
	if err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}

	neighbors, err := graphStore.Neighbors(ctx, "engineering/reference/standard-001", gyrus.EdgeFilter{Direction: "outgoing"})
	if err != nil {
		t.Fatalf("Neighbors failed: %v", err)
	}
	if len(neighbors) != 1 || neighbors[0].ToConceptID != "root/reference/adr-001" {
		t.Errorf("Expected outgoing edge to adr-001, got %v", neighbors)
	}
}
