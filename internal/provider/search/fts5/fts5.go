package fts5

import (
	"context"

	indexsqlite "github.com/armckinney/gyrus/internal/provider/index/sqlite"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// SearchProvider implements gyrus.SearchProvider over SQLite FTS5 indexer.
type SearchProvider struct {
	indexer *indexsqlite.Indexer
}

// NewSearchProvider returns a new SearchProvider wrapping a SQLite indexer.
func NewSearchProvider(indexer *indexsqlite.Indexer) *SearchProvider {
	return &SearchProvider{indexer: indexer}
}

// Search executes FTS5 full-text search query.
func (p *SearchProvider) Search(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	return p.indexer.Search(ctx, gyrus.SearchQuery{
		Query:      query,
		Filter:     filter,
		MaxResults: 50,
	})
}
