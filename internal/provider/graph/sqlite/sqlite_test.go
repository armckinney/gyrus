package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"

	graphsqlite "github.com/armckinney/gyrus/internal/provider/graph/sqlite"
	indexsqlite "github.com/armckinney/gyrus/internal/provider/index/sqlite"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that the SQLite GraphStore handles edge insertion, neighbor filtering (incoming/outgoing/both), and edge deletion.
// [Execution Surface]: Real SQLite Database
// [Assertions]: Directed edges are upserted, neighbor queries filter by direction and relationship type, and deletion removes specified edges.
// -----------------------------------------------------------------------------
func TestSQLiteGraphStore_EdgesAndNeighbors(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "graph.db")

	indexer, err := indexsqlite.NewIndexer(dbPath)
	if err != nil {
		t.Fatalf("Failed creating indexer: %v", err)
	}
	defer indexer.Close()

	graphStore := graphsqlite.NewGraphStore(indexer.DB())
	ctx := context.Background()

	edges := []gyrus.ConceptEdge{
		{
			FromConceptID:    "root/reference/prd-001",
			ToConceptID:      "root/reference/adr-001",
			RelationshipType: gyrus.RelDependsOn,
		},
		{
			FromConceptID:    "root/reference/prd-001",
			ToConceptID:      "root/reference/adr-002",
			RelationshipType: gyrus.RelImplements,
		},
		{
			FromConceptID:    "root/reference/adr-002",
			ToConceptID:      "root/reference/adr-003",
			RelationshipType: gyrus.RelSupersedes,
		},
	}

	// 1. Upsert Edges
	if err := graphStore.UpsertEdges(ctx, edges); err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}

	// 2. Outgoing Neighbors for prd-001
	outgoing, err := graphStore.Neighbors(ctx, "root/reference/prd-001", gyrus.EdgeFilter{Direction: "outgoing"})
	if err != nil {
		t.Fatalf("Neighbors outgoing failed: %v", err)
	}
	if len(outgoing) != 2 {
		t.Errorf("Expected 2 outgoing edges from prd-001, got %d", len(outgoing))
	}

	// 3. Incoming Neighbors for adr-001
	incoming, err := graphStore.Neighbors(ctx, "root/reference/adr-001", gyrus.EdgeFilter{Direction: "incoming"})
	if err != nil {
		t.Fatalf("Neighbors incoming failed: %v", err)
	}
	if len(incoming) != 1 || incoming[0].FromConceptID != "root/reference/prd-001" {
		t.Errorf("Expected incoming edge from prd-001 to adr-001, got %v", incoming)
	}

	// 4. Filter by RelationshipType
	dependsOnEdges, err := graphStore.Neighbors(ctx, "root/reference/prd-001", gyrus.EdgeFilter{
		Direction:        "outgoing",
		RelationshipType: gyrus.RelDependsOn,
	})
	if err != nil {
		t.Fatalf("Neighbors with rel filter failed: %v", err)
	}
	if len(dependsOnEdges) != 1 || dependsOnEdges[0].ToConceptID != "root/reference/adr-001" {
		t.Errorf("Expected 1 depends_on edge to adr-001, got %v", dependsOnEdges)
	}

	// 5. Delete Edge
	if err := graphStore.DeleteEdges(ctx, "root/reference/prd-001", "root/reference/adr-001", gyrus.RelDependsOn); err != nil {
		t.Fatalf("DeleteEdges failed: %v", err)
	}

	afterDelete, err := graphStore.Neighbors(ctx, "root/reference/prd-001", gyrus.EdgeFilter{Direction: "outgoing"})
	if err != nil {
		t.Fatalf("Neighbors after delete failed: %v", err)
	}
	if len(afterDelete) != 1 || afterDelete[0].ToConceptID != "root/reference/adr-002" {
		t.Errorf("Expected only 1 edge remaining to adr-002, got %v", afterDelete)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies recursive multi-hop graph traversal over directed concept edges in SQLite.
// [Execution Surface]: Real SQLite Database
// [Assertions]: Traverse resolves all reachable nodes and edges up to MaxDepth.
// -----------------------------------------------------------------------------
func TestSQLiteGraphStore_Traverse(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "graph.db")

	indexer, err := indexsqlite.NewIndexer(dbPath)
	if err != nil {
		t.Fatalf("Failed creating indexer: %v", err)
	}
	defer indexer.Close()

	graphStore := graphsqlite.NewGraphStore(indexer.DB())
	ctx := context.Background()

	// Chain: doc-a -> doc-b -> doc-c
	edges := []gyrus.ConceptEdge{
		{
			FromConceptID:    "root/reference/doc-a",
			ToConceptID:      "root/reference/doc-b",
			RelationshipType: gyrus.RelDependsOn,
		},
		{
			FromConceptID:    "root/reference/doc-b",
			ToConceptID:      "root/reference/doc-c",
			RelationshipType: gyrus.RelDependsOn,
		},
	}

	if err := graphStore.UpsertEdges(ctx, edges); err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}

	paths, err := graphStore.Traverse(ctx, gyrus.GraphQuery{
		StartID:  "root/reference/doc-a",
		MaxDepth: 3,
	})
	if err != nil {
		t.Fatalf("Traverse failed: %v", err)
	}

	if len(paths) == 0 {
		t.Fatalf("Expected traversal paths, got empty")
	}

	foundNodes := make(map[string]bool)
	for _, node := range paths[0].Nodes {
		foundNodes[node] = true
	}

	if !foundNodes["root/reference/doc-a"] || !foundNodes["root/reference/doc-b"] || !foundNodes["root/reference/doc-c"] {
		t.Errorf("Expected nodes doc-a, doc-b, doc-c in traversal path, got %v", paths[0].Nodes)
	}
}
