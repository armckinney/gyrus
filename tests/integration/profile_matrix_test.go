package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/internal/provider/sqlite"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestProfileMatrixIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-integration-*")
	if err != nil {
		t.Fatalf("Failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storageRoot := filepath.Join(tempDir, "docs")
	dbPath := filepath.Join(tempDir, "index.db")

	store, err := localfs.NewStore(storageRoot)
	if err != nil {
		t.Fatalf("Failed localfs initialization: %v", err)
	}

	indexer, err := sqlite.NewIndexer(dbPath)
	if err != nil {
		t.Fatalf("Failed sqlite indexer initialization: %v", err)
	}
	defer indexer.Close()

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

	// 1. Storage
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

	// 3. Search
	results, err := indexer.Search(ctx, gyrus.SearchQuery{Query: "idiomatic"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].Document.ID != doc.ID {
		t.Errorf("Expected search match for standard-001, got %v", results)
	}
}
