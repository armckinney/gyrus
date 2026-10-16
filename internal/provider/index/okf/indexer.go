// Package okf provides a metadata index derived directly from stored OKF concepts.
package okf

import (
	"context"

	domain "github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/internal/provider/okfutil"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

type Indexer struct{ scanner gyrus.ConceptScanner }

func NewIndexer(store gyrus.ConceptStore) (*Indexer, error) {
	scanner, err := okfutil.Scanner(store)
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
	return okfutil.Search(ctx, idx.scanner, q)
}
func (idx *Indexer) Sync(ctx context.Context, _ string) (gyrus.SyncReport, error) {
	return idx.scanner.Scan(ctx, func(gyrus.Concept) error { return nil })
}

var _ gyrus.IndexStore = (*Indexer)(nil)
