package postgres_test

import (
	"context"
	"github.com/pashagolub/pgxmock/v4"
	"testing"

	graphpg "github.com/armckinney/gyrus/internal/provider/graph/postgres"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that PostgreSQL GraphStore handles edge insertion, neighbor queries, and edge deletions.
// [Execution Surface]: In-Memory PostgreSQL Mock Pool (pgxmock)
// [Assertions]: SQL queries for INSERT/UPDATE edges, SELECT neighbors, and DELETE execute against mock pool.
// -----------------------------------------------------------------------------
func TestPostgresGraphStore_EdgesAndNeighbors(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed opening pgxmock pool: %v", err)
	}
	defer mock.Close()

	graphStore := graphpg.NewGraphStore(mock)
	ctx := context.Background()

	edges := []gyrus.ConceptEdge{
		{
			FromConceptID:    "root/reference/prd-001",
			ToConceptID:      "root/reference/adr-001",
			RelationshipType: gyrus.RelDependsOn,
		},
	}

	// 1. Upsert Edges
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO concept_edges").
		WithArgs("root/reference/prd-001", "root/reference/adr-001", "depends_on").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	if err := graphStore.UpsertEdges(ctx, edges); err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}

	// 2. Neighbors Outgoing
	rows := pgxmock.NewRows([]string{"from_concept_id", "to_concept_id", "relationship_type"}).
		AddRow("root/reference/prd-001", "root/reference/adr-001", "depends_on")

	mock.ExpectQuery("SELECT from_concept_id, to_concept_id, relationship_type FROM concept_edges WHERE from_concept_id = \\$1").
		WithArgs("root/reference/prd-001").
		WillReturnRows(rows)

	outgoing, err := graphStore.Neighbors(ctx, "root/reference/prd-001", gyrus.EdgeFilter{Direction: "outgoing"})
	if err != nil {
		t.Fatalf("Neighbors failed: %v", err)
	}
	if len(outgoing) != 1 || outgoing[0].ToConceptID != "root/reference/adr-001" {
		t.Errorf("Expected outgoing edge to adr-001, got %v", outgoing)
	}

	// 3. Delete Edges
	mock.ExpectExec("DELETE FROM concept_edges WHERE from_concept_id = \\$1 AND to_concept_id = \\$2 AND relationship_type = \\$3").
		WithArgs("root/reference/prd-001", "root/reference/adr-001", "depends_on").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	if err := graphStore.DeleteEdges(ctx, "root/reference/prd-001", "root/reference/adr-001", gyrus.RelDependsOn); err != nil {
		t.Fatalf("DeleteEdges failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}
