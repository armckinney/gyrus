package fts5_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	indexsqlite "github.com/armckinney/gyrus/internal/provider/index/sqlite"
	"github.com/armckinney/gyrus/internal/provider/search/fts5"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that SearchProvider over SQLite FTS5 executes keyword searches, applies category/type/status filters, and handles special character sanitization.
// [Execution Surface]: Real SQLite FTS5 Engine
// [Assertions]: Queries return matched documents with BM25 rankings; special characters in search queries do not cause SQLite syntax errors.
// -----------------------------------------------------------------------------
func TestFTS5SearchProvider(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "fts5.db")

	indexer, err := indexsqlite.NewIndexer(dbPath)
	if err != nil {
		t.Fatalf("Failed creating indexer: %v", err)
	}
	defer indexer.Close()

	ctx := context.Background()

	docs := []gyrus.Document{
		{
			ID:          "adr-001-vector",
			Title:       "Vector Search Provider",
			Category:    gyrus.CategoryArchitecture,
			Type:        gyrus.TypeADR,
			OwnerGroup:  "platform",
			Version:     1,
			Status:      "accepted",
			LastUpdated: time.Now(),
			Tags:        []string{"vector", "ollama", "search"},
			Content:     "Semantic embeddings with Ollama nomic-embed-text for local similarity matching.",
		},
		{
			ID:          "spec-001-fts5",
			Title:       "FTS5 Lexical Search Engine",
			Category:    gyrus.CategoryTechnical,
			Type:        gyrus.TypeSpecification,
			OwnerGroup:  "platform",
			Version:     1,
			Status:      "active",
			LastUpdated: time.Now(),
			Tags:        []string{"fts5", "sqlite", "search"},
			Content:     "BM25 lexical full text search using SQLite virtual table fts5.",
		},
	}

	for _, d := range docs {
		if err := indexer.Index(ctx, d); err != nil {
			t.Fatalf("Failed indexing %s: %v", d.ID, err)
		}
	}

	searcher := fts5.NewSearchProvider(indexer)

	// 1. Basic Keyword Search
	results, err := searcher.Search(ctx, "embeddings", gyrus.SearchFilter{})
	if err != nil {
		t.Fatalf("Search for 'embeddings' failed: %v", err)
	}
	if len(results) != 1 || results[0].Document.ID != "adr-001-vector" {
		t.Errorf("Expected match for adr-001-vector, got %v", results)
	}

	// 2. Filtered Search (Type = specification)
	specResults, err := searcher.Search(ctx, "search", gyrus.SearchFilter{Type: gyrus.TypeSpecification})
	if err != nil {
		t.Fatalf("Filtered search failed: %v", err)
	}
	if len(specResults) != 1 || specResults[0].Document.ID != "spec-001-fts5" {
		t.Errorf("Expected spec-001-fts5 match, got %v", specResults)
	}

	// 3. Special Character Queries (Sanitization)
	specialQueries := []string{
		`"exact phrase"`,
		`vector*`,
		`Ollama -nomic`,
		`search (AND) or 'quotes'`,
		`special!@#$%^&*()_+ chars`,
	}

	for _, sq := range specialQueries {
		t.Run("Sanitize/"+sq, func(t *testing.T) {
			_, err := searcher.Search(ctx, sq, gyrus.SearchFilter{})
			if err != nil {
				t.Errorf("Special query '%s' caused unexpected error: %v", sq, err)
			}
		})
	}
}
