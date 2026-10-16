package okf_test

import (
	"context"
	"errors"
	"testing"

	graph "github.com/armckinney/gyrus/internal/provider/graph/okf"
	"github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

type failingStore struct{ *localfs.Store }

func (s failingStore) Update(ctx context.Context, id string, p gyrus.ConceptPatch, v int) (gyrus.ConceptRef, error) {
	if id == "root/reference/b" {
		return gyrus.ConceptRef{}, errors.New("injected persistence failure")
	}
	return s.Store.Update(ctx, id, p, v)
}
func TestBatchPreflightPartialFailureAndImmutableMetadata(t *testing.T) {
	ctx := context.Background()
	store, _ := localfs.NewStore(t.TempDir())
	for _, id := range []string{"a", "b"} {
		if _, err := store.Create(ctx, gyrus.Concept{ID: id, Title: id, Type: gyrus.TypeADR, Category: gyrus.CategoryArchitecture, Tenant: "root", Version: 1, Status: "accepted", Immutable: true, Content: "historical body", Name: "root/reference/" + id}); err != nil {
			t.Fatal(err)
		}
	}
	g, _ := graph.NewGraphStore(failingStore{store})
	edges := []gyrus.ConceptEdge{{FromConceptID: "root/reference/a", ToConceptID: "root/reference/future", RelationshipType: gyrus.RelSupersedes}, {FromConceptID: "root/reference/b", ToConceptID: "root/reference/future", RelationshipType: gyrus.RelMitigates}}
	invalid := append(append([]gyrus.ConceptEdge{}, edges...), gyrus.ConceptEdge{FromConceptID: "root/reference/missing", ToConceptID: "root/reference/future", RelationshipType: gyrus.RelImplements})
	if err := g.UpsertEdges(ctx, invalid); err == nil {
		t.Fatal("missing source accepted")
	}
	first, _ := store.Get(ctx, "root/reference/a")
	if first.Version != 1 {
		t.Fatal("preflight wrote a source")
	}
	err := g.UpsertEdges(ctx, edges)
	var partial *gyrus.PartialGraphMutationError
	if !errors.As(err, &partial) || partial.FailedSourceID != "root/reference/b" || len(partial.AppliedSourceIDs) != 1 || partial.AppliedSourceIDs[0] != "root/reference/a" {
		t.Fatalf("partial error: %v", err)
	}
	first, _ = store.Get(ctx, "root/reference/a")
	second, _ := store.Get(ctx, "root/reference/b")
	if first.Version != 2 || first.Content != "historical body" || second.Version != 1 {
		t.Fatalf("partial preservation: %+v %+v", first, second)
	}
	if err := g.UpsertEdges(ctx, edges[:1]); err != nil {
		t.Fatal(err)
	}
	duplicate, _ := store.Get(ctx, "root/reference/a")
	if duplicate.Version != 2 || duplicate.Links[0] != first.Links[0] {
		t.Fatal("duplicate changed revision/audit")
	}
}
