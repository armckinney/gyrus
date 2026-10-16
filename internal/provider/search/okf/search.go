// Package okf provides direct keyword and metadata retrieval over stored concepts.
package okf

import (
	"context"

	sharedokf "github.com/armckinney/gyrus/internal/provider/shared/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

type SearchProvider struct{ scanner gyrus.ConceptScanner }

func NewSearchProvider(store gyrus.ConceptStore) (*SearchProvider, error) {
	scanner, err := sharedokf.Scanner(store)
	if err != nil {
		return nil, err
	}
	return &SearchProvider{scanner}, nil
}
func (p *SearchProvider) Search(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	return sharedokf.Search(ctx, p.scanner, gyrus.SearchQuery{Query: query, Filter: filter})
}

var _ gyrus.SearchProvider = (*SearchProvider)(nil)
