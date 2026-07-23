package sqlite_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/armckinney/gyrus/internal/provider/sqlite"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestSQLiteIndexerAndFTS(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-sqlite-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "index.db")
	indexer, err := sqlite.NewIndexer(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize Indexer: %v", err)
	}
	defer indexer.Close()

	ctx := context.Background()

	doc1 := gyrus.Document{
		ID:             "adr-001-sqlite",
		Title:          "SQLite Embedded Search Provider Architecture",
		Category:       gyrus.CategoryArchitecture,
		Type:           gyrus.TypeADR,
		OwnerGroup:     "platform",
		Version:        1,
		Status:         "accepted",
		LastModifiedBy: "dev-joe",
		LastUpdated:    time.Now().Truncate(time.Second),
		Tags:           []string{"sqlite", "search", "fts5"},
		Content:        "We choose SQLite FTS5 for zero-dependency embedded lexical search queries.",
	}

	doc2 := gyrus.Document{
		ID:             "guide-002-git",
		Title:          "Git Repository Local File Store Guide",
		Category:       gyrus.CategoryTechnical,
		Type:           gyrus.TypeGuide,
		OwnerGroup:     "platform",
		Version:        1,
		Status:         "active",
		LastModifiedBy: "dev-alice",
		LastUpdated:    time.Now().Truncate(time.Second),
		Tags:           []string{"git", "storage"},
		Content:        "Instructions for mounting local git repository folders.",
	}

	if err := indexer.Index(ctx, doc1); err != nil {
		t.Fatalf("Indexing doc1 failed: %v", err)
	}
	if err := indexer.Index(ctx, doc2); err != nil {
		t.Fatalf("Indexing doc2 failed: %v", err)
	}

	// Search FTS keyword
	query := gyrus.SearchQuery{
		Query: "SQLite",
	}
	results, err := indexer.Search(ctx, query)
	if err != nil {
		t.Fatalf("Search query failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("Expected search results for 'SQLite', got 0")
	}

	if results[0].Document.ID != doc1.ID {
		t.Errorf("Expected first match ID %s, got %s", doc1.ID, results[0].Document.ID)
	}

	// Category filter query
	filterQuery := gyrus.SearchQuery{
		Filter: gyrus.SearchFilter{
			Category: gyrus.CategoryTechnical,
		},
	}
	filteredResults, err := indexer.Search(ctx, filterQuery)
	if err != nil {
		t.Fatalf("Filtered search failed: %v", err)
	}
	if len(filteredResults) != 1 || filteredResults[0].Document.ID != doc2.ID {
		t.Errorf("Expected 1 match (%s), got %v", doc2.ID, filteredResults)
	}
}
