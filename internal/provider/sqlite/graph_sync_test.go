package sqlite_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/internal/provider/sqlite"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestSQLiteGraphAndSync(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-graph-sync-test-*")
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

	// 1. Test Graph Edge Operations
	edge := gyrus.DocumentEdge{
		FromDocumentID:   "adr-001",
		ToDocumentID:     "prd-001",
		RelationshipType: gyrus.RelImplements,
		CreatedBy:        "dev-team",
		CreatedAt:        time.Now().Truncate(time.Second),
	}

	if err := indexer.UpsertEdges(ctx, []gyrus.DocumentEdge{edge}); err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}

	neighbors, err := indexer.Neighbors(ctx, "adr-001", gyrus.EdgeFilter{Direction: "outgoing"})
	if err != nil {
		t.Fatalf("Neighbors failed: %v", err)
	}
	if len(neighbors) != 1 || neighbors[0].ToDocumentID != "prd-001" {
		t.Errorf("Expected outgoing neighbor prd-001, got %v", neighbors)
	}

	// 2. Test File Sync Engine
	storageRoot := filepath.Join(tempDir, "docs")
	store, err := localfs.NewStore(storageRoot)
	if err != nil {
		t.Fatalf("Failed creating localfs store: %v", err)
	}

	doc1 := gyrus.Document{
		ID:           "doc-alpha",
		Title:        "Document Alpha",
		Category:     gyrus.CategoryArchitecture,
		Type:         gyrus.TypeADR,
		OwnerGroup:   "core",
		Version:      1,
		Status:       "proposed",
		Dependencies: []string{"doc-beta"},
		Content:      "Alpha content depends on beta.",
	}
	doc2 := gyrus.Document{
		ID:         "doc-beta",
		Title:      "Document Beta",
		Category:   gyrus.CategoryArchitecture,
		Type:       gyrus.TypeSpecification,
		OwnerGroup: "core",
		Version:    1,
		Status:     "active",
		Content:    "Beta content specification.",
	}

	_, _ = store.Create(ctx, doc1)
	_, _ = store.Create(ctx, doc2)

	report, err := indexer.Sync(ctx, storageRoot)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if report.IndexedFiles != 2 {
		t.Errorf("Expected 2 indexed files, got %d", report.IndexedFiles)
	}

	// Verify auto-extracted edge from sync
	betaNeighbors, err := indexer.Neighbors(ctx, "doc-alpha", gyrus.EdgeFilter{Direction: "outgoing"})
	if err != nil {
		t.Fatalf("Neighbors check after sync failed: %v", err)
	}
	if len(betaNeighbors) != 1 || betaNeighbors[0].ToDocumentID != "doc-beta" {
		t.Errorf("Expected auto-extracted dependency edge to doc-beta, got %v", betaNeighbors)
	}
}
