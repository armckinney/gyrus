package postgres

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/pashagolub/pgxmock/v4"
)

func TestStore_SearchFTS(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	store := &Store{pool: mock}
	
	doc := gyrus.Document{
		ID:      "doc-1",
		Title:   "Postgres FTS Search",
		Category: gyrus.CategoryTechnical,
	}
	fmBytes, _ := json.Marshal(doc)

	// Since we construct raw SQL string, we need a flexible matcher or exact query.
	// For this test we will just match the prefix with Regex.
	mock.ExpectQuery("SELECT frontmatter, content").
		WithArgs("technical", "postgres | search").
		WillReturnRows(pgxmock.NewRows([]string{"frontmatter", "content", "score"}).AddRow(fmBytes, "Some content here", 0.9))

	filter := gyrus.SearchFilter{
		Category: gyrus.CategoryTechnical,
	}
	results, err := store.SearchFTS(context.Background(), "postgres search", filter)
	if err != nil {
		t.Errorf("error was not expected while searching: %s", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Document.ID != "doc-1" {
		t.Errorf("expected ID 'doc-1', got '%s'", results[0].Document.ID)
	}
	if results[0].Score != 0.9 {
		t.Errorf("expected Score 0.9, got %f", results[0].Score)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
