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
// [Purpose]: Verifies that PostgreSQL Store executes full-text search against tsvector index.
// [Execution Surface]: In-Memory PostgreSQL Mock Pool (pgxmock)
// [Assertions]: SQL query executes with category filter and tsquery args, returning expected score and concept.
// -----------------------------------------------------------------------------
func TestStore_SearchFTS(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	store := &Store{pool: mock}

	doc := gyrus.Concept{
		ID:       "doc-1",
		Title:    "Postgres FTS Search",
		Category: gyrus.CategoryTechnical, Tenant: "root", Name: "root/reference/doc-1",
	}
	fmBytes, _ := json.Marshal(doc)

	// Since we construct raw SQL string, we need a flexible matcher or exact query.
	// For this test we will just match the prefix with Regex.
	mock.ExpectQuery("SELECT id, frontmatter, content").
		WithArgs("technical", "postgres OR search", gyrus.DefaultSearchLimit).
		WillReturnRows(pgxmock.NewRows([]string{"id", "frontmatter", "content", "score"}).AddRow(doc.Name, fmBytes, "Some content here", 0.9))

	filter := gyrus.SearchFilter{
		Category: gyrus.CategoryTechnical,
	}
	results, err := store.Search(context.Background(), gyrus.SearchQuery{Query: "postgres search", Filter: filter})
	if err != nil {
		t.Errorf("error was not expected while searching: %s", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Concept.ID != "doc-1" {
		t.Errorf("expected ID 'doc-1', got '%s'", results[0].Concept.ID)
	}
	if results[0].Score != 0.9 {
		t.Errorf("expected Score 0.9, got %f", results[0].Score)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
