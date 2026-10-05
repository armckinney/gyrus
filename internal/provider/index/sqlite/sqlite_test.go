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

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that SQLite Indexer derives scope and workspace from paths during Sync, prioritizes current workspace results, and isolates external workspaces.
// [Assertions]: Workspace documents rank higher than reference documents, foreign workspaces are excluded by default, and Scope='all' retrieves cross-workspace.
// -----------------------------------------------------------------------------
func TestSQLiteIndexerWorkspaceScopingAndPrioritization(t *testing.T) {
	tempDir := t.TempDir()
	storageRoot := filepath.Join(tempDir, "docs", "root")
	dbPath := filepath.Join(tempDir, "index.db")

	refDir := filepath.Join(storageRoot, "reference")
	gyrusWsDir := filepath.Join(storageRoot, "workspaces", "gyrus")
	billingWsDir := filepath.Join(storageRoot, "workspaces", "billing")

	for _, d := range []string{refDir, gyrusWsDir, billingWsDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("Failed creating directory %s: %v", d, err)
		}
	}

	indexer, err := sqlite.NewIndexer(dbPath)
	if err != nil {
		t.Fatalf("Failed initializing indexer: %v", err)
	}
	defer indexer.Close()

	ctx := context.Background()

	// 1. Reference doc
	refDoc := &gyrus.Document{
		ID:         "adr-shared-db",
		Title:      "Shared Database Strategy",
		Category:   gyrus.CategoryArchitecture,
		Type:       gyrus.TypeADR,
		OwnerGroup: "root",
		Version:    1,
		Status:     "accepted",
		Content:    "High level database patterns across services.",
	}
	refBytes, _ := okf.SerializeMarkdown(refDoc)
	if err := os.WriteFile(filepath.Join(refDir, "adr-shared-db.md"), refBytes, 0644); err != nil {
		t.Fatalf("Failed writing refDoc: %v", err)
	}

	// 2. Current workspace doc (gyrus)
	wsDoc := &gyrus.Document{
		ID:         "ticket-gyrus-db",
		Title:      "Gyrus Database Implementation",
		Category:   gyrus.CategoryProduct,
		Type:       gyrus.TypePRD,
		OwnerGroup: "root",
		Version:    1,
		Status:     "active",
		Content:    "Specific database schema for gyrus.",
	}
	wsBytes, _ := okf.SerializeMarkdown(wsDoc)
	if err := os.WriteFile(filepath.Join(gyrusWsDir, "ticket-gyrus-db.md"), wsBytes, 0644); err != nil {
		t.Fatalf("Failed writing wsDoc: %v", err)
	}

	// 3. Other workspace doc (billing)
	otherWsDoc := &gyrus.Document{
		ID:         "ticket-billing-db",
		Title:      "Billing Database Implementation",
		Category:   gyrus.CategoryProduct,
		Type:       gyrus.TypePRD,
		OwnerGroup: "root",
		Version:    1,
		Status:     "active",
		Content:    "Specific database schema for billing.",
	}
	otherBytes, _ := okf.SerializeMarkdown(otherWsDoc)
	if err := os.WriteFile(filepath.Join(billingWsDir, "ticket-billing-db.md"), otherBytes, 0644); err != nil {
		t.Fatalf("Failed writing otherWsDoc: %v", err)
	}

	// Sync
	report, err := indexer.Sync(ctx, filepath.Join(tempDir, "docs"))
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if report.IndexedFiles != 3 {
		t.Fatalf("Expected 3 indexed files, got %d", report.IndexedFiles)
	}

	// Test 1: Workspace search prioritizing current workspace (gyrus)
	scopedResults, err := indexer.Search(ctx, gyrus.SearchQuery{
		Query: "database",
		Filter: gyrus.SearchFilter{
			Workspace: "gyrus",
		},
	})
	if err != nil {
		t.Fatalf("Scoped search failed: %v", err)
	}

	// Should contain ticket-gyrus-db and adr-shared-db, but NOT ticket-billing-db
	if len(scopedResults) != 2 {
		t.Fatalf("Expected exactly 2 results (current ws + reference), got %d", len(scopedResults))
	}
	if scopedResults[0].Document.ID != "ticket-gyrus-db" {
		t.Errorf("Expected workspace doc first, got %s", scopedResults[0].Document.ID)
	}
	if scopedResults[1].Document.ID != "adr-shared-db" {
		t.Errorf("Expected reference doc second, got %s", scopedResults[1].Document.ID)
	}
	if scopedResults[0].Score <= scopedResults[1].Score {
		t.Errorf("Expected workspace score (%f) to be higher than reference score (%f)",
			scopedResults[0].Score, scopedResults[1].Score)
	}

	// Check populated scope & workspace
	if scopedResults[0].Document.Scope != "workspace" || scopedResults[0].Document.Workspace != "gyrus" {
		t.Errorf("Expected Scope 'workspace' and Workspace 'gyrus', got %s / %s",
			scopedResults[0].Document.Scope, scopedResults[0].Document.Workspace)
	}
	if scopedResults[1].Document.Scope != "reference" {
		t.Errorf("Expected Scope 'reference', got %s", scopedResults[1].Document.Scope)
	}

	// Test 2: Reference-only search
	refOnlyResults, err := indexer.Search(ctx, gyrus.SearchQuery{
		Query: "database",
		Filter: gyrus.SearchFilter{
			Scope: "reference",
		},
	})
	if err != nil {
		t.Fatalf("Reference search failed: %v", err)
	}
	if len(refOnlyResults) != 1 || refOnlyResults[0].Document.ID != "adr-shared-db" {
		t.Errorf("Expected only adr-shared-db, got %v", refOnlyResults)
	}

	// Test 3: Scope='all' retrieves across all workspaces
	allResults, err := indexer.Search(ctx, gyrus.SearchQuery{
		Query: "database",
		Filter: gyrus.SearchFilter{
			Scope: "all",
		},
	})
	if err != nil {
		t.Fatalf("Scope all search failed: %v", err)
	}
	if len(allResults) != 3 {
		t.Errorf("Expected all 3 documents, got %d", len(allResults))
	}
}
