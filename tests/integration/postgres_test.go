package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/armckinney/gyrus/internal/provider/db"
	graphpostgres "github.com/armckinney/gyrus/internal/provider/graph/postgres"
	indexpostgres "github.com/armckinney/gyrus/internal/provider/index/postgres"
	searchpostgres "github.com/armckinney/gyrus/internal/provider/search/postgres_fts"
	storagepostgres "github.com/armckinney/gyrus/internal/provider/storage/postgres"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies end-to-end functionality for the 'postgres' profile against a real PostgreSQL database server.
// [Execution Surface]: Real PostgreSQL Database Server (localhost:5432 / postgres:5432)
// [Assertions]: Postgres storage persists documents, tsvector generates search index, ts_rank_cd scores searches, and graph tables resolve edges.
// -----------------------------------------------------------------------------
func TestIntegration_PostgresLive(t *testing.T) {
	connStrings := []string{
		os.Getenv("POSTGRES_URL"),
		"postgres://postgres:postgres@postgres:5432/gyrus?sslmode=disable",
		"postgres://postgres:postgres@localhost:5432/gyrus?sslmode=disable",
	}

	var connString string
	var poolErr error
	ctx := context.Background()

	for _, cs := range connStrings {
		if cs == "" {
			continue
		}
		pool, err := db.OpenPostgres(ctx, cs)
		if err == nil && pool.Ping(ctx) == nil {
			connString = cs
			break
		}
		poolErr = err
	}

	if connString == "" {
		t.Skipf("PostgreSQL database is not reachable, skipping live postgres integration test: %v", poolErr)
	}

	// 1. Initialize Postgres Store (runs DDL migrations)
	store, err := storagepostgres.NewStore(ctx, connString)
	if err != nil {
		t.Fatalf("Failed initializing Postgres store: %v", err)
	}

	// 2. Initialize Indexer, Searcher, and GraphStore
	indexer, err := indexpostgres.NewIndexer(ctx, connString)
	if err != nil {
		t.Fatalf("Failed initializing Postgres indexer: %v", err)
	}
	defer indexer.Close()

	searcher, err := searchpostgres.NewSearchProvider(ctx, connString)
	if err != nil {
		t.Fatalf("Failed initializing Postgres searcher: %v", err)
	}

	graphStore := graphpostgres.NewGraphStore(indexer.Pool())

	runID := time.Now().UnixNano()
	docAID := fmt.Sprintf("adr-live-pg-%d", runID)
	docBID := fmt.Sprintf("prd-live-pg-%d", runID)
	uniqueOwner := fmt.Sprintf("owner-%d", runID)

	docA := gyrus.Document{
		ID:         docAID,
		Title:      "Postgres Relational Architecture",
		Category:   gyrus.CategoryArchitecture,
		Type:       gyrus.TypeADR,
		OwnerGroup: uniqueOwner,
		Version:    1,
		Status:     "accepted",
		Tags:       []string{"postgres", "relational", "sql"},
		Content:    "High performance PostgreSQL backend with tsvector indexing and JSONB storage.",
	}
	docB := gyrus.Document{
		ID:         docBID,
		Title:      "Product Spec for Cloud Memory",
		Category:   gyrus.CategoryProduct,
		Type:       gyrus.TypePRD,
		OwnerGroup: uniqueOwner,
		Version:    1,
		Status:     "active",
		Tags:       []string{"cloud", "postgres"},
		Content:    "Product specifications utilizing PostgreSQL storage backend.",
	}

	defer func() {
		_ = store.Delete(ctx, docAID)
		_ = store.Delete(ctx, docBID)
		_ = indexer.Remove(ctx, docAID)
		_ = indexer.Remove(ctx, docBID)
	}()

	// 3. Store Create & Get
	if _, err := store.Create(ctx, docA); err != nil {
		t.Fatalf("Postgres store create docA failed: %v", err)
	}
	if _, err := store.Create(ctx, docB); err != nil {
		t.Fatalf("Postgres store create docB failed: %v", err)
	}

	fetched, err := store.Get(ctx, docA.ID)
	if err != nil {
		t.Fatalf("Postgres store get failed: %v", err)
	}
	if fetched.Title != docA.Title {
		t.Errorf("Expected Title '%s', got '%s'", docA.Title, fetched.Title)
	}

	// 4. Indexer Index
	if err := indexer.Index(ctx, docA); err != nil {
		t.Fatalf("Postgres indexing failed: %v", err)
	}
	if err := indexer.Index(ctx, docB); err != nil {
		t.Fatalf("Postgres indexing failed: %v", err)
	}

	// 5. Tsvector FTS Search
	results, err := searcher.Search(ctx, "relational tsvector", gyrus.SearchFilter{OwnerGroup: uniqueOwner})
	if err != nil {
		t.Fatalf("Postgres FTS search failed: %v", err)
	}
	if len(results) == 0 || results[0].Document.ID != docA.ID {
		t.Errorf("Expected top FTS match to be docA, got %v", results)
	}

	// 6. Graph Edges & Neighbors
	edge := gyrus.DocumentEdge{
		FromDocumentID:   docB.ID,
		ToDocumentID:     docA.ID,
		RelationshipType: gyrus.RelDependsOn,
		CreatedBy:        "test-runner",
		CreatedAt:        time.Now(),
	}
	if err := graphStore.UpsertEdges(ctx, []gyrus.DocumentEdge{edge}); err != nil {
		t.Fatalf("Postgres graph upsert failed: %v", err)
	}

	outgoing, err := graphStore.Neighbors(ctx, docB.ID, gyrus.EdgeFilter{Direction: "outgoing"})
	if err != nil {
		t.Fatalf("Postgres graph neighbors failed: %v", err)
	}
	if len(outgoing) != 1 || outgoing[0].ToDocumentID != docA.ID {
		t.Errorf("Expected 1 outgoing edge to docA, got %v", outgoing)
	}
}
