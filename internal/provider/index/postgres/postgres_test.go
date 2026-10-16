package postgres

import (
	"context"
	"encoding/json"
	"github.com/pashagolub/pgxmock/v4"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that the PostgreSQL Indexer properly indexes, searches, and removes concept envelopes.
// [Execution Surface]: In-Memory PostgreSQL Mock Pool (pgxmock)
// [Assertions]: SQL queries for INSERT/UPDATE, DELETE, and SELECT execute matching expectations.
// -----------------------------------------------------------------------------
func TestPostgresIndexer_IndexSearchRemove(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed opening pgxmock pool: %v", err)
	}
	defer mock.Close()

	indexer := &Indexer{pool: mock}
	ctx := context.Background()

	doc := gyrus.Concept{
		ID:       "adr-001-pg",
		Title:    "Postgres Storage Driver",
		Category: gyrus.CategoryArchitecture,
		Type:     gyrus.TypeADR,
		Tenant:   "platform",
		Version:  1,
		Status:   "accepted",
		Content:  "PostgreSQL backend persistence.", Name: "platform/reference/adr-001-pg",
	}

	// 1. Index
	mock.ExpectExec("INSERT INTO concepts").
		WithArgs("platform/reference/adr-001-pg", pgxmock.AnyArg(), "PostgreSQL backend persistence.").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := indexer.Index(ctx, doc); err != nil {
		t.Fatalf("Index failed: %v", err)
	}

	// 2. Search
	fmBytes, _ := json.Marshal(doc)
	rows := pgxmock.NewRows([]string{"id", "frontmatter", "content", "score"}).
		AddRow(doc.Name, fmBytes, "PostgreSQL backend persistence.", 1000.0)

	mock.ExpectQuery("SELECT id, frontmatter, content").
		WithArgs("platform/reference/adr-001-pg", gyrus.DefaultSearchLimit).
		WillReturnRows(rows)

	results, err := indexer.Search(ctx, gyrus.SearchQuery{Query: "platform/reference/adr-001-pg"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].Concept.ID != doc.ID {
		t.Errorf("Expected result for %s, got %v", doc.ID, results)
	}

	// 3. Remove
	mock.ExpectExec("DELETE FROM concepts WHERE id = \\$1").
		WithArgs("platform/reference/adr-001-pg").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	if err := indexer.Remove(ctx, "platform/reference/adr-001-pg"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}
