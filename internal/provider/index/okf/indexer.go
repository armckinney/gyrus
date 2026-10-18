// Package okf provides a metadata index derived directly from stored OKF concepts.
package okf

import (
	"context"

	domain "github.com/armckinney/gyrus/internal/domain/okf"
	sharedokf "github.com/armckinney/gyrus/internal/provider/shared/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

type Indexer struct{ scanner gyrus.ConceptScanner }

func NewIndexer(store gyrus.ConceptStore) (*Indexer, error) {
	scanner, err := sharedokf.Scanner(store)
	if err != nil {
		return nil, err
	}
	return &Indexer{scanner}, nil
}
func (idx *Indexer) Index(ctx context.Context, doc gyrus.Concept) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return domain.Validate(&doc)
}
func (idx *Indexer) Remove(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return domain.ValidateName(id)
}
func (idx *Indexer) Search(ctx context.Context, q gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	return sharedokf.Search(ctx, idx.scanner, q)
}

var _ gyrus.IndexStore = (*Indexer)(nil)
