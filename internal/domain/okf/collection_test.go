package okf_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestTypedLinksRoundTripAndCollectionFailures(t *testing.T) {
	doc := gyrus.Concept{ID: "a", Title: "A", Category: gyrus.CategoryTechnical, Type: gyrus.TypeGuide, Tenant: "root", Version: 1, Status: "active", Links: []gyrus.ConceptLink{{Concept: "root/reference/b", Rel: gyrus.RelDependsOn}, {Concept: "root/reference/b", Rel: gyrus.RelDependsOn}}, Name: "root/reference/a"}
	data, err := okf.SerializeMarkdown(&doc)
	if err != nil {
		t.Fatal(err)
	}
	round, err := okf.ParseMarkdown(data)
	if err != nil {
		t.Fatal(err)
	}
	jsonData, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	jsonRound, err := okf.ParseJSON(jsonData)
	if err != nil || len(jsonRound.Links) != 2 {
		t.Fatalf("JSON links: %+v %v", jsonRound, err)
	}
	edges := gyrus.ConceptEdges(*round)
	if len(edges) != 1 {
		t.Fatalf("normalized audit: %+v", edges)
	}
	round.Links[0].Rel = ""
	if err := okf.Validate(round); err == nil {
		t.Fatal("empty typed relation accepted")
	}
	root := t.TempDir()
	store, _ := localfs.NewStore(root)
	if _, err := store.Create(context.Background(), doc); err != nil {
		t.Fatal(err)
	}
	original, _ := store.Get(context.Background(), doc.Name)
	duplicate := filepath.Join(root, "concepts/root/projects/duplicate/a.md")
	if err := os.MkdirAll(filepath.Dir(duplicate), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(duplicate, data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("ignored"), 0644); err != nil {
		t.Fatal(err)
	}
	report, err := store.Scan(context.Background(), func(gyrus.Concept) error { return nil })
	if err != nil || report.ScannedFiles != 2 || len(report.Errors) != 0 {
		t.Fatalf("duplicate scan: %+v %v", report, err)
	}
	if err := os.Remove(duplicate); err != nil {
		t.Fatal(err)
	}
	title := "updated"
	if _, err := store.Update(context.Background(), doc.Name, gyrus.ConceptPatch{Title: &title}, original.Version); err != nil {
		t.Fatal(err)
	}
	preserved, _ := store.Get(context.Background(), doc.Name)
	if len(preserved.Links) != 2 {
		t.Fatal("unrelated patch lost links")
	}
	clear := []gyrus.ConceptLink{}
	if _, err := store.Update(context.Background(), doc.Name, gyrus.ConceptPatch{Links: &clear}, preserved.Version); err != nil {
		t.Fatal(err)
	}
	cleared, _ := store.Get(context.Background(), doc.Name)
	if len(cleared.Links) != 0 {
		t.Fatal("empty patch did not clear typed links")
	}
	for _, bad := range []string{"../escape", "a/b", "a\\b", "..", "", "con", "lpt1"} {
		if _, err := store.Get(context.Background(), bad); err == nil {
			t.Fatalf("unsafe ID accepted: %q", bad)
		}
	}
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "concepts/root/reference/outside.md")); err == nil {
		if _, err := store.Scan(context.Background(), func(gyrus.Concept) error { return nil }); err == nil {
			t.Fatal("symlink accepted")
		}
	}
}
