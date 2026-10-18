package postgres_fts_test

import (
	"context"
	"encoding/json"
	"github.com/pashagolub/pgxmock/v4"
	"regexp"
	"testing"

	"github.com/armckinney/gyrus/internal/provider/search/postgres_fts"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestSearchScopeAndLimits(t *testing.T) {
	for _, tc := range []struct {
		name      string
		request   gyrus.SearchQuery
		condition string
		args      []any
	}{
		{"reference", gyrus.SearchQuery{Filter: gyrus.SearchFilter{Scope: "reference"}}, "COALESCE(frontmatter->>'scope', 'reference') = 'reference'", nil},
		{"project", gyrus.SearchQuery{Filter: gyrus.SearchFilter{Scope: "project", Project: "demo"}}, "COALESCE(frontmatter->>'project', '') = $1", []any{"demo"}},
		{"default project plus reference", gyrus.SearchQuery{Filter: gyrus.SearchFilter{Project: "demo"}}, "(COALESCE(frontmatter->>'scope', 'reference') = 'reference' OR COALESCE(frontmatter->>'project', '') = $1)", []any{"demo"}},
		{"all uncapped", gyrus.SearchQuery{Filter: gyrus.SearchFilter{Scope: "all"}}, "FROM concepts ORDER BY", nil},
		{"metadata limit", gyrus.SearchQuery{Filter: gyrus.SearchFilter{Scope: "all"}, MaxResults: 3}, "LIMIT $1", []any{3}},
		{"keyword default", gyrus.SearchQuery{Query: "needle", Filter: gyrus.SearchFilter{Scope: "all"}}, "LIMIT $2", []any{"needle", gyrus.DefaultSearchLimit}},
		{"keyword larger limit", gyrus.SearchQuery{Query: "needle", Filter: gyrus.SearchFilter{Scope: "all"}, MaxResults: 20}, "LIMIT $2", []any{"needle", 20}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatal(err)
			}
			defer mock.Close()
			mock.ExpectQuery("SELECT id, frontmatter, content.*" + regexp.QuoteMeta(tc.condition)).WithArgs(tc.args...).WillReturnRows(pgxmock.NewRows([]string{"id", "frontmatter", "content", "score"}))
			if _, err := postgres_fts.NewSearchProviderWithPool(mock).Search(context.Background(), tc.request); err != nil {
				t.Fatal(err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSearchRejectsInvalidScope(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()
	_, err = postgres_fts.NewSearchProviderWithPool(mock).Search(context.Background(), gyrus.SearchQuery{Filter: gyrus.SearchFilter{Scope: "invalid"}})
	if err == nil {
		t.Fatal("invalid scope accepted")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

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

	doc := gyrus.Concept{
		ID:       "adr-001-pg-fts",
		Title:    "Postgres FTS Search Engine",
		Category: gyrus.CategoryTechnical,
		Type:     gyrus.TypeSpecification,
		Tenant:   "platform",
		Version:  1,
		Status:   "active",
		Content:  "High performance tsvector search engine.", Name: "platform/reference/adr-001-pg-fts",
	}

	fmBytes, _ := json.Marshal(doc)
	rows := pgxmock.NewRows([]string{"id", "frontmatter", "content", "score"}).
		AddRow(doc.Name, fmBytes, doc.Content, 0.85)

	mock.ExpectQuery("SELECT id, frontmatter, content, ts_rank_cd").
		WithArgs("technical", "tsvector OR engine", gyrus.DefaultSearchLimit).
		WillReturnRows(rows)

	results, err := searcher.Search(ctx, gyrus.SearchQuery{Query: "tsvector engine", Filter: gyrus.SearchFilter{
		Category: gyrus.CategoryTechnical,
	}})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 || results[0].Concept.ID != doc.ID {
		t.Errorf("Expected result for %s, got %v", doc.ID, results)
	}
	if results[0].Score != 0.85 {
		t.Errorf("Expected score 0.85, got %f", results[0].Score)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled mock expectations: %v", err)
	}
}
