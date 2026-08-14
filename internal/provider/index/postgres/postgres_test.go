package postgres

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/pashagolub/pgxmock/v4"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that the PostgreSQL Indexer properly indexes, searches, and removes document envelopes.
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

	doc := gyrus.Document{
		ID:         "adr-001-pg",
		Title:      "Postgres Storage Driver",
		Category:   gyrus.CategoryArchitecture,
		Type:       gyrus.TypeADR,
		OwnerGroup: "platform",
		Version:    1,
		Status:     "accepted",
		Content:    "PostgreSQL backend persistence.",
	}

	// 1. Index
	mock.ExpectExec("INSERT INTO documents").
		WithArgs("adr-001-pg", pgxmock.AnyArg(), "PostgreSQL backend persistence.").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := indexer.Index(ctx, doc); err != nil {
		t.Fatalf("Index failed: %v", err)
	}

	// 2. Search
	fmBytes, _ := json.Marshal(doc)
	rows := pgxmock.NewRows([]string{"frontmatter", "content"}).
		AddRow(fmBytes, "PostgreSQL backend persistence.")

	mock.ExpectQuery("SELECT frontmatter, content FROM documents WHERE id = \\$1").
		WithArgs("adr-001-pg").
		WillReturnRows(rows)

	results, err := indexer.Search(ctx, gyrus.SearchQuery{Query: "adr-001-pg"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].Document.ID != doc.ID {
		t.Errorf("Expected result for %s, got %v", doc.ID, results)
	}

	// 3. Remove
	mock.ExpectExec("DELETE FROM documents WHERE id = \\$1").
		WithArgs("adr-001-pg").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	if err := indexer.Remove(ctx, "adr-001-pg"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}
