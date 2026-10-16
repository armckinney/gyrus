package provider_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/config"
	"github.com/armckinney/gyrus/internal/provider"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestOKFFactoryCompositionAndNoHiddenDatabase(t *testing.T) {
	root := t.TempDir()
	cfg := &config.Config{}
	store, err := provider.NewConceptStore(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	idx, err := provider.NewIndexStore(cfg, root, store)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := provider.NewGraphStore(cfg, root, store, idx)
	if err != nil {
		t.Fatal(err)
	}
	search, err := provider.NewSearchProvider(cfg, root, store, idx)
	if err != nil {
		t.Fatal(err)
	}
	doc := gyrus.Concept{ID: "shared-store", Title: "Shared", Category: gyrus.CategoryTechnical, Type: gyrus.TypeGuide, Tenant: "root", Version: 1, Status: "active", Name: "root/reference/shared-store"}
	if _, err := store.Create(context.Background(), doc); err != nil {
		t.Fatal(err)
	}
	if err := graph.UpsertEdges(context.Background(), []gyrus.ConceptEdge{{FromConceptID: doc.Name, ToConceptID: "root/reference/future", RelationshipType: gyrus.RelImplements}}); err != nil {
		t.Fatal(err)
	}
	results, err := search.Search(context.Background(), "", gyrus.SearchFilter{})
	if err != nil || len(results) != 1 || len(results[0].Concept.Links) != 1 {
		t.Fatalf("shared store: %v %v", results, err)
	}
	if _, err := os.Stat(filepath.Join(root, "index.db")); !os.IsNotExist(err) {
		t.Fatal("hidden database created")
	}
	cfg.Index.Provider = "okf"
	cfg.Search.Provider = "sqlite_fts5"
	untouched := filepath.Join(root, "must-not-exist")
	if _, err := provider.NewConceptStore(cfg, untouched); err == nil {
		t.Fatal("incompatible config accepted")
	}
	if _, err := os.Stat(untouched); !os.IsNotExist(err) {
		t.Fatal("resource created before compatibility failure")
	}
}
