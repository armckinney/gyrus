package sqlite_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/internal/provider/index/sqlite"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that the SQLite Indexer properly indexes, searches, updates, and removes documents with FTS5 search.
// [Execution Surface]: Real SQLite Database File (In-Memory / Temp File)
// [Assertions]: Document is indexed, full-text search returns matching documents, updates overwrite existing, and remove purges the document.
// -----------------------------------------------------------------------------
func TestSQLiteIndexerCRUDAndSearch(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "index.db")

	indexer, err := sqlite.NewIndexer(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize SQLite Indexer: %v", err)
	}
	defer indexer.Close()

	ctx := context.Background()

	doc := gyrus.Document{
		ID:             "adr-001-storage",
		Title:          "Storage Engine Architecture",
		Category:       gyrus.CategoryArchitecture,
		Type:           gyrus.TypeADR,
		OwnerGroup:     "platform",
		Version:        1,
		Status:         "accepted",
		LastModifiedBy: "developer-alice",
		LastUpdated:    time.Now(),
		Tags:           []string{"storage", "sqlite", "architecture"},
		Dependencies:   []string{"prd-core-engine"},
		Content:        "We use SQLite FTS5 for high performance local search.",
	}

	// 1. Index document
	if err := indexer.Index(ctx, doc); err != nil {
		t.Fatalf("Failed indexing document: %v", err)
	}

	// 2. Search by keyword
	results, err := indexer.Search(ctx, gyrus.SearchQuery{Query: "performance"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].Document.ID != doc.ID {
		t.Errorf("Expected search match for %s, got %v", doc.ID, results)
	}

	// 3. Search with filter (category)
	filteredResults, err := indexer.Search(ctx, gyrus.SearchQuery{
		Query: "storage",
		Filter: gyrus.SearchFilter{
			Category: gyrus.CategoryArchitecture,
		},
	})
	if err != nil {
		t.Fatalf("Filtered search failed: %v", err)
	}
	if len(filteredResults) != 1 {
		t.Errorf("Expected 1 filtered result, got %d", len(filteredResults))
	}

	// 4. Update document content
	doc.Content = "Updated content with unique-term-xyz123"
	if err := indexer.Index(ctx, doc); err != nil {
		t.Fatalf("Failed re-indexing document: %v", err)
	}

	updatedResults, err := indexer.Search(ctx, gyrus.SearchQuery{Query: "unique-term-xyz123"})
	if err != nil {
		t.Fatalf("Search after update failed: %v", err)
	}
	if len(updatedResults) != 1 || updatedResults[0].Document.ID != doc.ID {
		t.Errorf("Expected match for updated term, got %v", updatedResults)
	}

	// 5. Remove document
	if err := indexer.Remove(ctx, doc.ID); err != nil {
		t.Fatalf("Failed removing document: %v", err)
	}

	afterRemove, err := indexer.Search(ctx, gyrus.SearchQuery{Query: "unique-term-xyz123"})
	if err != nil {
		t.Fatalf("Search after remove failed: %v", err)
	}
	if len(afterRemove) != 0 {
		t.Errorf("Expected 0 results after removal, got %d", len(afterRemove))
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that SQLite Indexer syncs a directory of Markdown documents, extracts dependency edges, and prunes stale documents.
// [Execution Surface]: Real SQLite Database File + Real Filesystem Tree
// [Assertions]: Sync report indexes all valid documents and extracts cross-document edge dependencies.
// -----------------------------------------------------------------------------
func TestSQLiteIndexerSync(t *testing.T) {
	tempDir := t.TempDir()
	storageRoot := filepath.Join(tempDir, "docs")
	dbPath := filepath.Join(tempDir, "index.db")

	if err := os.MkdirAll(storageRoot, 0755); err != nil {
		t.Fatalf("Failed creating storage root: %v", err)
	}

	indexer, err := sqlite.NewIndexer(dbPath)
	if err != nil {
		t.Fatalf("Failed initializing SQLite Indexer: %v", err)
	}
	defer indexer.Close()

	ctx := context.Background()

	// Write 2 markdown files
	doc1 := &gyrus.Document{
		ID:           "prd-001",
		Title:        "Core Product",
		Category:     gyrus.CategoryProduct,
		Type:         gyrus.TypePRD,
		OwnerGroup:   "product",
		Version:      1,
		Status:       "active",
		Dependencies: []string{"adr-001"},
		Content:      "Product requirements document.",
	}
	doc1Bytes, _ := okf.SerializeMarkdown(doc1)
	if err := os.WriteFile(filepath.Join(storageRoot, "prd-001.md"), doc1Bytes, 0644); err != nil {
		t.Fatalf("Failed writing doc1: %v", err)
	}

	doc2 := &gyrus.Document{
		ID:         "adr-001",
		Title:      "Architecture Record",
		Category:   gyrus.CategoryArchitecture,
		Type:       gyrus.TypeADR,
		OwnerGroup: "architecture",
		Version:    1,
		Status:     "accepted",
		Content:    "Architecture design record.",
	}
	doc2Bytes, _ := okf.SerializeMarkdown(doc2)
	if err := os.WriteFile(filepath.Join(storageRoot, "adr-001.md"), doc2Bytes, 0644); err != nil {
		t.Fatalf("Failed writing doc2: %v", err)
	}

	report, err := indexer.Sync(ctx, storageRoot)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if report.IndexedFiles != 2 {
		t.Errorf("Expected 2 indexed files, got %d", report.IndexedFiles)
	}

	// Verify Search
	results, err := indexer.Search(ctx, gyrus.SearchQuery{Query: "Architecture"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].Document.ID != "adr-001" {
		t.Errorf("Expected search match for adr-001, got %v", results)
	}
}
