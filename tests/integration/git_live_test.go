package integration_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/armckinney/gyrus/internal/config"
	"github.com/armckinney/gyrus/internal/provider"
	storagegit "github.com/armckinney/gyrus/internal/provider/storage/git"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// TestGitLiveRemote writes disposable fixtures only when a remote is explicitly
// supplied. Credentials come from the same Git credential helper as production.
func TestGitLiveRemote(t *testing.T) {
	remote := os.Getenv("GYRUS_TEST_GIT_URL")
	if remote == "" {
		t.Skip("set GYRUS_TEST_GIT_URL to a disposable remote to enable live Git validation")
	}
	run := os.Getenv("GYRUS_TEST_GIT_RUN_ID")
	if run == "" {
		run = fmt.Sprintf("run-%d", time.Now().UnixNano())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	trunk, err := storagegit.NewStoreContext(ctx, storagegit.Options{RepoURL: remote})
	if err != nil {
		t.Fatal(err)
	}
	name := "root/reference/" + run
	_, err = trunk.Create(ctx, gyrus.Concept{ID: run, Tenant: "root", Title: "Trunk " + run, Category: gyrus.CategoryTechnical, Type: gyrus.TypeGuide, Status: "active", Content: "Disposable live validation fixture."})
	if err != nil {
		t.Fatal(err)
	}
	for _, initiative := range []string{"a", "b"} {
		branch := "validation/" + run + "/" + initiative
		session, err := storagegit.NewStoreContext(ctx, storagegit.Options{RepoURL: remote, CommitStrategy: "branching", Branch: branch})
		if err != nil {
			t.Fatal(err)
		}
		title := "Initiative " + initiative
		if _, err = session.Update(ctx, name, gyrus.ConceptPatch{Title: &title}, 1); err != nil {
			t.Fatal(err)
		}
		cfg := &config.Config{}
		cfg.Storage.Provider = "git"
		idx, err := provider.NewIndexStore(cfg, session)
		if err != nil {
			t.Fatal(err)
		}
		graph, err := provider.NewGraphStore(cfg, session, idx)
		if err != nil {
			t.Fatal(err)
		}
		if err = graph.UpsertEdges(ctx, []gyrus.ConceptEdge{{FromConceptID: name, ToConceptID: "root/reference/future", RelationshipType: gyrus.RelDependsOn}}); err != nil {
			t.Fatal(err)
		}
		search, err := provider.NewSearchProvider(cfg, session, idx)
		if err != nil {
			t.Fatal(err)
		}
		results, err := search.Search(ctx, gyrus.SearchQuery{Query: run})
		if err != nil || len(results) != 1 || results[0].Concept.Title != title {
			t.Fatalf("initiative search: %+v %v", results, err)
		}
		schema := "---\nid: <id>\ntitle: " + title + "\ntype: guide\n---\n# " + title + "\n"
		if err = session.SaveSchema(ctx, "guide", schema); err != nil {
			t.Fatal(err)
		}
		if err = session.SaveSchema(ctx, "guide", schema); err != nil {
			t.Fatal(err)
		}
		fresh, err := storagegit.NewStoreContext(ctx, storagegit.Options{RepoURL: remote, CommitStrategy: "branching", Branch: branch})
		if err != nil {
			t.Fatal(err)
		}
		doc, err := fresh.Get(ctx, name)
		if err != nil || doc.Title != title || len(doc.Links) != 1 {
			t.Fatalf("initiative reload: %+v %v", doc, err)
		}
		content, err := fresh.GetSchema(ctx, "guide")
		if err != nil || !strings.Contains(content, title) {
			t.Fatalf("schema reload: %v", err)
		}
		if _, err = fresh.Switch(ctx, "direct", ""); err != nil {
			t.Fatal(err)
		}
		doc, err = fresh.Get(ctx, name)
		if err != nil || doc.Title != "Trunk "+run || len(doc.Links) != 0 {
			t.Fatalf("trunk was changed by initiative: %+v %v", doc, err)
		}
		t.Logf("Validated direct -> %s -> fresh session -> direct", branch)
	}
	t.Logf("Fixture for GitHub PR validation: %s", name)
}
