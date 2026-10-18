package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

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
	connString := os.Getenv("GYRUS_TEST_POSTGRES_URL")
	if connString == "" {
		t.Skip("set GYRUS_TEST_POSTGRES_URL to an isolated PostgreSQL test database")
	}
	ctx := context.Background()

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
	uniqueTenant := fmt.Sprintf("tenant-%d", runID)

	docA := gyrus.Concept{
		ID:       docAID,
		Title:    "Postgres Relational Architecture",
		Category: gyrus.CategoryArchitecture,
		Type:     gyrus.TypeADR,
		Tenant:   uniqueTenant,
		Version:  1,
		Status:   "accepted",
		Tags:     []string{"postgres", "relational", "sql"},
		Content:  "High performance PostgreSQL backend with tsvector indexing and JSONB storage.",
		Scope:    "reference", Name: uniqueTenant + "/reference/" + docAID,
	}
	docB := gyrus.Concept{
		ID:       docBID,
		Title:    "Product Spec for Cloud Memory",
		Category: gyrus.CategoryProduct,
		Type:     gyrus.TypePRD,
		Tenant:   uniqueTenant,
		Version:  1,
		Status:   "active",
		Tags:     []string{"cloud", "postgres"},
		Content:  "Product specifications utilizing PostgreSQL storage backend.",
		Scope:    "project", Project: "gyrus", Name: uniqueTenant + "/projects/gyrus/" + docBID,
	}

	defer func() {
		_ = store.Delete(ctx, docA.Name)
		_ = store.Delete(ctx, docB.Name)
	}()

	// 3. Store Create & Get
	if _, err := store.Create(ctx, docA); err != nil {
		t.Fatalf("Postgres store create docA failed: %v", err)
	}
	if _, err := store.Create(ctx, docB); err != nil {
		t.Fatalf("Postgres store create docB failed: %v", err)
	}

	fetched, err := store.Get(ctx, docA.Name)
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
	results, err := searcher.Search(ctx, gyrus.SearchQuery{Query: "relational tsvector", Filter: gyrus.SearchFilter{Tenant: uniqueTenant}})
	if err != nil {
		t.Fatalf("Postgres FTS search failed: %v", err)
	}
	if len(results) == 0 || results[0].Concept.ID != docA.ID {
		t.Errorf("Expected top FTS match to be docA, got %v", results)
	}
	for _, tc := range []struct {
		scope, project, name string
		want                 int
	}{
		{"reference", "", docA.Name, 1}, {"project", "gyrus", docB.Name, 1}, {"project", "other", "", 0}, {"all", "", "", 2},
	} {
		matches, err := searcher.Search(ctx, gyrus.SearchQuery{Filter: gyrus.SearchFilter{Tenant: uniqueTenant, Scope: tc.scope, Project: tc.project}})
		if err != nil || len(matches) != tc.want {
			t.Fatalf("scope=%s project=%s: %+v %v", tc.scope, tc.project, matches, err)
		}
		if tc.name != "" && matches[0].Concept.Name != tc.name {
			t.Errorf("qualified identity: got %s, want %s", matches[0].Concept.Name, tc.name)
		}
	}

	// 6. Graph Edges & Neighbors
	edge := gyrus.ConceptEdge{
		FromConceptID:    docB.Name,
		ToConceptID:      docA.Name,
		RelationshipType: gyrus.RelDependsOn,
	}
	if err := graphStore.UpsertEdges(ctx, []gyrus.ConceptEdge{edge}); err != nil {
		t.Fatalf("Postgres graph upsert failed: %v", err)
	}

	outgoing, err := graphStore.Neighbors(ctx, docB.Name, gyrus.EdgeFilter{Direction: "outgoing"})
	if err != nil {
		t.Fatalf("Postgres graph neighbors failed: %v", err)
	}
	if len(outgoing) != 1 || outgoing[0].ToConceptID != docA.Name {
		t.Errorf("Expected 1 outgoing edge to docA, got %v", outgoing)
	}
}
