package postgres_test

import (
	"context"
	"testing"
	"time"

	graphpg "github.com/armckinney/gyrus/internal/provider/graph/postgres"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/pashagolub/pgxmock/v4"
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

	now := time.Now()
	edges := []gyrus.DocumentEdge{
		{
			FromDocumentID:   "prd-001",
			ToDocumentID:     "adr-001",
			RelationshipType: gyrus.RelDependsOn,
			CreatedBy:        "dev-1",
			CreatedAt:        now,
		},
	}

	// 1. Upsert Edges
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO document_edges").
		WithArgs("prd-001", "adr-001", "depends_on", "dev-1", now).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	if err := graphStore.UpsertEdges(ctx, edges); err != nil {
		t.Fatalf("UpsertEdges failed: %v", err)
	}

	// 2. Neighbors Outgoing
	rows := pgxmock.NewRows([]string{"from_document_id", "to_document_id", "relationship_type", "created_by", "created_at"}).
		AddRow("prd-001", "adr-001", "depends_on", "dev-1", now)

	mock.ExpectQuery("SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE from_document_id = \\$1").
		WithArgs("prd-001").
		WillReturnRows(rows)

	outgoing, err := graphStore.Neighbors(ctx, "prd-001", gyrus.EdgeFilter{Direction: "outgoing"})
	if err != nil {
		t.Fatalf("Neighbors failed: %v", err)
	}
	if len(outgoing) != 1 || outgoing[0].ToDocumentID != "adr-001" {
		t.Errorf("Expected outgoing edge to adr-001, got %v", outgoing)
	}

	// 3. Delete Edges
	mock.ExpectExec("DELETE FROM document_edges WHERE from_document_id = \\$1 AND to_document_id = \\$2 AND relationship_type = \\$3").
		WithArgs("prd-001", "adr-001", "depends_on").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	if err := graphStore.DeleteEdges(ctx, "prd-001", "adr-001", gyrus.RelDependsOn); err != nil {
		t.Fatalf("DeleteEdges failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}
