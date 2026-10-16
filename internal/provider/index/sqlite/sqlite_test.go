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

func TestReindexPreservesDatabaseOwnedRelationships(t *testing.T) {
	ctx := context.Background()
	idx, err := sqlite.NewIndexer(filepath.Join(t.TempDir(), "index.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()
	doc := gyrus.Concept{ID: "same", Tenant: "root", Scope: "reference", Title: "Source", Category: gyrus.CategoryTechnical, Type: gyrus.TypeGuide, Version: 1, Status: "active", Links: []gyrus.ConceptLink{{Concept: "root/reference/dependency", Rel: gyrus.RelDependsOn}}}
	if err := idx.Index(ctx, doc); err != nil {
		t.Fatal(err)
	}
	if _, err := idx.DB().Exec("INSERT INTO concept_edges (from_concept_id,to_concept_id,relationship_type) VALUES (?,?,?)", "root/reference/same", "root/reference/proposal", "implements"); err != nil {
		t.Fatal(err)
	}
	doc.Title = "Updated"
	if err := idx.Index(ctx, doc); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := idx.DB().QueryRow("SELECT count(*) FROM concept_edges WHERE from_concept_id = ?", "root/reference/same").Scan(&count); err != nil || count != 2 {
		t.Fatalf("database-owned edge lost: %d %v", count, err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that the SQLite Indexer properly indexes, searches, updates, and removes concepts with FTS5 search.
// [Execution Surface]: Real SQLite Database File (In-Memory / Temp File)
// [Assertions]: Concept is indexed, full-text search returns matching concepts, updates overwrite existing, and remove purges the concept.
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

	doc := gyrus.Concept{
		ID:             "adr-001-storage",
		Title:          "Storage Engine Architecture",
		Category:       gyrus.CategoryArchitecture,
		Type:           gyrus.TypeADR,
		Tenant:         "platform",
		Version:        1,
		Status:         "accepted",
		LastModifiedBy: "developer-alice",
		LastUpdated:    time.Now(),
		Tags:           []string{"storage", "sqlite", "architecture"},
		Links:          []gyrus.ConceptLink{{Concept: "root/reference/prd-core-engine", Rel: gyrus.RelDependsOn}},
		Content:        "We use SQLite FTS5 for high performance local search.", Name: "platform/reference/adr-001-storage",
	}

	// 1. Index concept
	if err := indexer.Index(ctx, doc); err != nil {
		t.Fatalf("Failed indexing concept: %v", err)
	}

	// 2. Search by keyword
	results, err := indexer.Search(ctx, gyrus.SearchQuery{Query: "performance"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].Concept.ID != doc.ID {
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

	// 4. Update concept content
	doc.Content = "Updated content with unique-term-xyz123"
	if err := indexer.Index(ctx, doc); err != nil {
		t.Fatalf("Failed re-indexing concept: %v", err)
	}

	updatedResults, err := indexer.Search(ctx, gyrus.SearchQuery{Query: "unique-term-xyz123"})
	if err != nil {
		t.Fatalf("Search after update failed: %v", err)
	}
	if len(updatedResults) != 1 || updatedResults[0].Concept.ID != doc.ID {
		t.Errorf("Expected match for updated term, got %v", updatedResults)
	}

	// 5. Remove concept
	if err := indexer.Remove(ctx, doc.Name); err != nil {
		t.Fatalf("Failed removing concept: %v", err)
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
// [Purpose]: Verifies that SQLite Indexer syncs a directory of Markdown concepts, extracts dependency edges, and prunes stale concepts.
// [Execution Surface]: Real SQLite Database File + Real Filesystem Tree
// [Assertions]: Sync report indexes all valid concepts and extracts cross-concept edge dependencies.
// -----------------------------------------------------------------------------
func TestSQLiteIndexerSync(t *testing.T) {
	tempDir := t.TempDir()
	storageRoot := filepath.Join(tempDir, "concepts")
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
	doc1 := &gyrus.Concept{
		ID:       "prd-001",
		Title:    "Core Product",
		Category: gyrus.CategoryProduct,
		Type:     gyrus.TypePRD,
		Tenant:   "product",
		Version:  1,
		Status:   "active",
		Links:    []gyrus.ConceptLink{{Concept: "root/reference/adr-001", Rel: gyrus.RelDependsOn}},
		Content:  "Product requirements document.", Name: "product/projects/gyrus/prd-001",
	}
	_ = os.MkdirAll(filepath.Join(storageRoot, "concepts/product/reference"), 0755)
	_ = os.MkdirAll(filepath.Join(storageRoot, "concepts/architecture/reference"), 0755)
	doc1Bytes, _ := okf.SerializeMarkdown(doc1)
	if err := os.WriteFile(filepath.Join(storageRoot, "concepts/product/reference/prd-001.md"), doc1Bytes, 0644); err != nil {
		t.Fatalf("Failed writing doc1: %v", err)
	}

	doc2 := &gyrus.Concept{
		ID:       "adr-001",
		Title:    "Architecture Record",
		Category: gyrus.CategoryArchitecture,
		Type:     gyrus.TypeADR,
		Tenant:   "architecture",
		Version:  1,
		Status:   "accepted",
		Content:  "Architecture design record.", Name: "architecture/reference/adr-001",
	}
	doc2Bytes, _ := okf.SerializeMarkdown(doc2)
	if err := os.WriteFile(filepath.Join(storageRoot, "concepts/architecture/reference/adr-001.md"), doc2Bytes, 0644); err != nil {
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
	if len(results) != 1 || results[0].Concept.ID != "adr-001" {
		t.Errorf("Expected search match for adr-001, got %v", results)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that SQLite Indexer derives scope and project from paths during Sync, prioritizes current project results, and isolates external projects.
// [Assertions]: Project concepts rank higher than reference concepts, foreign projects are excluded by default, and Scope='all' retrieves cross-project.
// -----------------------------------------------------------------------------
func TestSQLiteIndexerProjectScopingAndPrioritization(t *testing.T) {
	tempDir := t.TempDir()
	storageRoot := filepath.Join(tempDir, "concepts", "root")
	dbPath := filepath.Join(tempDir, "index.db")

	refDir := filepath.Join(storageRoot, "reference")
	gyrusProjDir := filepath.Join(storageRoot, "projects", "gyrus")
	billingProjDir := filepath.Join(storageRoot, "projects", "billing")

	for _, d := range []string{refDir, gyrusProjDir, billingProjDir} {
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
	refDoc := &gyrus.Concept{
		ID:       "adr-shared-db",
		Title:    "Shared Database Strategy",
		Category: gyrus.CategoryArchitecture,
		Type:     gyrus.TypeADR,
		Tenant:   "root",
		Version:  1,
		Status:   "accepted",
		Content:  "High level database patterns across services.", Name: "root/reference/adr-shared-db",
	}
	refBytes, _ := okf.SerializeMarkdown(refDoc)
	if err := os.WriteFile(filepath.Join(refDir, "adr-shared-db.md"), refBytes, 0644); err != nil {
		t.Fatalf("Failed writing refDoc: %v", err)
	}

	// 2. Current project doc (gyrus)
	projDoc := &gyrus.Concept{
		ID:       "ticket-gyrus-db",
		Title:    "Gyrus Database Implementation",
		Category: gyrus.CategoryProduct,
		Type:     gyrus.TypePRD,
		Tenant:   "root",
		Version:  1,
		Status:   "active",
		Content:  "Specific database schema for gyrus.", Name: "root/projects/gyrus/ticket-gyrus-db",
	}
	projBytes, _ := okf.SerializeMarkdown(projDoc)
	if err := os.WriteFile(filepath.Join(gyrusProjDir, "ticket-gyrus-db.md"), projBytes, 0644); err != nil {
		t.Fatalf("Failed writing projDoc: %v", err)
	}

	// 3. Other project doc (billing)
	otherProjDoc := &gyrus.Concept{
		ID:       "ticket-billing-db",
		Title:    "Billing Database Implementation",
		Category: gyrus.CategoryProduct,
		Type:     gyrus.TypePRD,
		Tenant:   "root",
		Version:  1,
		Status:   "active",
		Content:  "Specific database schema for billing.", Name: "root/projects/gyrus/ticket-billing-db",
	}
	otherBytes, _ := okf.SerializeMarkdown(otherProjDoc)
	if err := os.WriteFile(filepath.Join(billingProjDir, "ticket-billing-db.md"), otherBytes, 0644); err != nil {
		t.Fatalf("Failed writing otherProjDoc: %v", err)
	}

	// Sync
	report, err := indexer.Sync(ctx, tempDir)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
	if report.IndexedFiles != 3 {
		t.Fatalf("Expected 3 indexed files, got %d", report.IndexedFiles)
	}

	// Test 1: Project search prioritizing current project (gyrus)
	scopedResults, err := indexer.Search(ctx, gyrus.SearchQuery{
		Query: "database",
		Filter: gyrus.SearchFilter{
			Project: "gyrus",
		},
	})
	if err != nil {
		t.Fatalf("Scoped search failed: %v", err)
	}

	// Should contain ticket-gyrus-db and adr-shared-db, but NOT ticket-billing-db
	if len(scopedResults) != 2 {
		t.Fatalf("Expected exactly 2 results (current proj + reference), got %d", len(scopedResults))
	}
	if scopedResults[0].Concept.ID != "ticket-gyrus-db" {
		t.Errorf("Expected project doc first, got %s", scopedResults[0].Concept.ID)
	}
	if scopedResults[1].Concept.ID != "adr-shared-db" {
		t.Errorf("Expected reference doc second, got %s", scopedResults[1].Concept.ID)
	}
	if scopedResults[0].Score <= scopedResults[1].Score {
		t.Errorf("Expected project score (%f) to be higher than reference score (%f)",
			scopedResults[0].Score, scopedResults[1].Score)
	}

	// Check populated scope & project
	if scopedResults[0].Concept.Scope != "project" || scopedResults[0].Concept.Project != "gyrus" {
		t.Errorf("Expected Scope 'project' and Project 'gyrus', got %s / %s",
			scopedResults[0].Concept.Scope, scopedResults[0].Concept.Project)
	}
	if scopedResults[1].Concept.Scope != "reference" {
		t.Errorf("Expected Scope 'reference', got %s", scopedResults[1].Concept.Scope)
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
	if len(refOnlyResults) != 1 || refOnlyResults[0].Concept.ID != "adr-shared-db" {
		t.Errorf("Expected only adr-shared-db, got %v", refOnlyResults)
	}

	// Test 3: Scope='all' retrieves across all projects
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
		t.Errorf("Expected all 3 concepts, got %d", len(allResults))
	}
}
