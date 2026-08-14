package integration_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	graphsqlite "github.com/armckinney/gyrus/internal/provider/graph/sqlite"
	indexsqlite "github.com/armckinney/gyrus/internal/provider/index/sqlite"
	searchfts5 "github.com/armckinney/gyrus/internal/provider/search/fts5"
	storagelocal "github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies end-to-end functionality for the 'local' profile (LocalFS storage + SQLite Indexer + SQLite Graph + SQLite FTS5 search).
// [Execution Surface]: Real Local Filesystem + Real SQLite Database
// [Assertions]: Document creation persists to disk, SQLite index syncs files, FTS5 search finds keywords, and Graph store manages directed edges.
// -----------------------------------------------------------------------------
func TestIntegration_LocalProfile(t *testing.T) {
	tempDir := t.TempDir()
	storageRoot := filepath.Join(tempDir, "docs")
	dbPath := filepath.Join(tempDir, "index.db")

	store, err := storagelocal.NewStore(storageRoot)
	if err != nil {
		t.Fatalf("Failed localfs initialization: %v", err)
	}

	indexer, err := indexsqlite.NewIndexer(dbPath)
	if err != nil {
		t.Fatalf("Failed sqlite indexer initialization: %v", err)
	}
	defer indexer.Close()

	graphStore := graphsqlite.NewGraphStore(indexer.DB())
	searcher := searchfts5.NewSearchProvider(indexer)

	ctx := context.Background()

	doc := gyrus.Document{
		ID:         "standard-001",
		Title:      "Coding Standard",
		Category:   gyrus.CategoryTechnical,
		Type:       gyrus.TypeStandards,
		OwnerGroup: "engineering",
		Version:    1,
		Status:     "active",
		Content:    "All Go code must follow standard idiomatic patterns.",
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
	report, err := indexer.Sync(ctx, storageRoot)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if report.IndexedFiles != 1 {
		t.Errorf("Expected 1 indexed file, got %d", report.IndexedFiles)
	}

	// 3. Search via SearchProvider
	results, err := searcher.Search(ctx, "idiomatic", gyrus.SearchFilter{})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].Document.ID != doc.ID {
		t.Errorf("Expected search match for standard-001, got %v", results)
	}

	// 4. Graph Edge Link
	err = graphStore.UpsertEdges(ctx, []gyrus.DocumentEdge{{
		FromDocumentID:   "standard-001",
		ToDocumentID:     "adr-001",
		RelationshipType: gyrus.RelDependsOn,
		CreatedAt:        time.Now(),
	}})
	if err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}

	neighbors, err := graphStore.Neighbors(ctx, "standard-001", gyrus.EdgeFilter{Direction: "outgoing"})
	if err != nil {
		t.Fatalf("Neighbors failed: %v", err)
	}
	if len(neighbors) != 1 || neighbors[0].ToDocumentID != "adr-001" {
		t.Errorf("Expected outgoing edge to adr-001, got %v", neighbors)
	}
}
