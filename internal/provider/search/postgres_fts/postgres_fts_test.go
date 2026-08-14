package postgres_fts_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/armckinney/gyrus/internal/provider/search/postgres_fts"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/pashagolub/pgxmock/v4"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies that SearchProvider over PostgreSQL executes tsvector full-text search with metadata filters and ranking.
// [Execution Surface]: In-Memory PostgreSQL Mock Pool (pgxmock)
// [Assertions]: SQL queries with tsvector @@ to_tsquery and frontmatter JSONB filters execute as expected.
// -----------------------------------------------------------------------------
func TestPostgresFTSSearchProvider(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed opening pgxmock pool: %v", err)
	}
	defer mock.Close()

	searcher := postgres_fts.NewSearchProviderWithPool(mock)
	ctx := context.Background()

	doc := gyrus.Document{
		ID:         "adr-001-pg-fts",
		Title:      "Postgres FTS Search Engine",
		Category:   gyrus.CategoryTechnical,
		Type:       gyrus.TypeSpecification,
		OwnerGroup: "platform",
		Version:    1,
		Status:     "active",
		Content:    "High performance tsvector search engine.",
	}

	fmBytes, _ := json.Marshal(doc)
	rows := pgxmock.NewRows([]string{"frontmatter", "content", "score"}).
		AddRow(fmBytes, doc.Content, 0.85)

	mock.ExpectQuery("SELECT frontmatter, content, ts_rank_cd").
		WithArgs("technical", "tsvector | engine").
		WillReturnRows(rows)

	results, err := searcher.Search(ctx, "tsvector engine", gyrus.SearchFilter{
		Category: gyrus.CategoryTechnical,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].Document.ID != doc.ID {
		t.Errorf("Expected result for %s, got %v", doc.ID, results)
	}
	if results[0].Score != 0.85 {
		t.Errorf("Expected score 0.85, got %f", results[0].Score)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}
