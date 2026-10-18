package okf_test

import (
	"context"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestQualifiedNameContract(t *testing.T) {
	for _, name := range []string{"root/reference/adr-001", "root/projects/gyrus/ticket-003", "Root/projects/Other/ticket-003", "concepts/reference/adr-001", "root/projects/concepts/ticket-003"} {
		tenant, project, id, err := okf.ParseName(name)
		if err != nil {
			t.Fatal(err)
		}
		doc := gyrus.Concept{ID: id, Tenant: tenant, Project: project, Scope: "reference"}
		if err := okf.SetName(&doc, ""); err != nil || doc.Name != name {
			t.Fatalf("name: %+v %v", doc, err)
		}
		if err := okf.SetLocation(&doc, "/parent/concepts/repo/concepts/"+name+".md"); err != nil || doc.Name != name {
			t.Fatalf("anchored location: %+v %v", doc, err)
		}
	}
	for _, name := range []string{"adr-001", "/root/reference/adr-001", "concepts/root/reference/adr-001", "root/reference/adr-001.md", "root/projects//adr-001", "root/reference/../adr-001", "root\\reference\\adr-001", "root/projects/gyrus/ADR-001", "root/other/adr-001", "root/reference/adr-001/"} {
		if err := okf.ValidateName(name); err == nil {
			t.Fatalf("accepted malformed name %q", name)
		}
	}
}

func TestQualifiedLocationsAndRepeatedLocalIDs(t *testing.T) {
	ctx := context.Background()
	store, err := localfs.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	doc := gyrus.Concept{ID: "same", Title: "Same", Tenant: "root", Scope: "reference", Category: gyrus.CategoryTechnical, Type: gyrus.TypeGuide, Version: 1, Status: "active"}
	for _, tc := range []struct{ tenant, project, name string }{{"root", "", "root/reference/same"}, {"root", "one", "root/projects/one/same"}, {"root", "two", "root/projects/two/same"}, {"other", "", "other/reference/same"}} {
		doc.Tenant, doc.Project = tc.tenant, tc.project
		ref, err := store.Create(ctx, doc)
		if err != nil || ref.Name != tc.name {
			t.Fatalf("create %s: %+v %v", tc.name, ref, err)
		}
		got, err := store.Get(ctx, tc.name)
		if err != nil || got.ID != "same" || got.Name != tc.name {
			t.Fatalf("get: %+v %v", got, err)
		}
		if _, err := store.Create(ctx, doc); err == nil {
			t.Fatal("duplicate qualified name accepted")
		}
	}
	if _, err := store.Get(ctx, "same"); err == nil {
		t.Fatal("short lookup accepted")
	}
	bad := doc
	bad.ID = "wrong"
	if err := okf.SetLocation(&bad, "concepts/other/reference/same.md"); err == nil {
		t.Fatal("filename mismatch accepted")
	}
	bad = doc
	bad.Tenant = "wrong"
	if err := okf.SetLocation(&bad, "concepts/other/reference/same.md"); err == nil {
		t.Fatal("tenant mismatch accepted")
	}
}

func TestObsoleteRelationshipFieldsRejected(t *testing.T) {
	for _, fields := range []string{"dependencies: []", "links:\n  - target_id: old\n    rel_type: depends_on", "links:\n  - concept: root/reference/target\n    rel: depends_on\n    created_at: 2026-10-16T00:00:00Z", "links:\n  - concept: root/reference/target\n    rel: depends_on\n    created_by: human", "links:\n  - concept: root/reference/target"} {
		if _, err := okf.ParseMarkdown([]byte("---\nid: source\ntype: guide\n" + fields + "\n---\nbody\n")); err == nil {
			t.Fatalf("accepted %s", fields)
		}
	}
	for _, data := range []string{`{"dependencies":[]}`, `{"links":[{"target_id":"old","rel_type":"depends_on"}]}`, `{"links":[{"concept":"root/reference/target","rel":"depends_on","created_by":"human"}]}`} {
		if _, err := okf.ParseJSON([]byte(data)); err == nil {
			t.Fatalf("accepted JSON %s", data)
		}
	}
	doc := gyrus.Concept{ID: "source", Name: "root/reference/source", Links: []gyrus.ConceptLink{{Concept: "root/reference/target", Rel: gyrus.RelDependsOn}}}
	data, err := okf.SerializeMarkdown(&doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"name:", "target_id:", "dependencies:", "created_by:", "created_at:"} {
		if strings.Contains(string(data), field) {
			t.Fatalf("leaked %s", field)
		}
	}
}
