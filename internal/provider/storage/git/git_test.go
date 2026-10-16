package git

import (
	"context"
	"github.com/go-git/go-git/v5"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestQualifiedNamesSurviveGitReload(t *testing.T) {
	ctx := context.Background()
	remote := setupTestRemote(t)
	store, err := NewStore(Options{RepoURL: remote})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ tenant, project, name string }{{"root", "", "root/reference/same"}, {"root", "demo", "root/projects/demo/same"}, {"other", "", "other/reference/same"}} {
		doc := gyrus.Concept{ID: "same", Tenant: tc.tenant, Scope: "reference", Project: tc.project, Title: tc.name, Type: gyrus.TypeGuide, Category: gyrus.CategoryTechnical, Version: 1, Status: "active"}
		ref, err := store.Create(ctx, doc)
		if err != nil || ref.Name != tc.name {
			t.Fatalf("create: %+v %v", ref, err)
		}
	}
	fresh, err := NewStore(Options{RepoURL: remote})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"root/reference/same", "root/projects/demo/same", "other/reference/same"} {
		doc, err := fresh.Get(ctx, name)
		if err != nil || doc.Name != name || doc.Title != name || doc.ID != "same" {
			t.Fatalf("reload: %+v %v", doc, err)
		}
	}
	if _, err := fresh.Get(ctx, "same"); err == nil {
		t.Fatal("short lookup accepted")
	}
	report, err := fresh.Scan(ctx, func(gyrus.Concept) error { return nil })
	if err != nil || len(report.Errors) > 0 || report.IndexedFiles != 3 {
		t.Fatalf("scoped scan: %+v %v", report, err)
	}
}

func setupTestRemote(t *testing.T) string {
	dir := t.TempDir()

	// Init bare repository to act as remote
	_, err := git.PlainInit(dir, true)
	if err != nil {
		t.Fatalf("failed to init bare test repo: %v", err)
	}

	return dir
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that GitStore executes versioned commits for document Create, Update, Archive, and Delete.
// [Execution Surface]: In-Memory / Local Bare Git Repository (go-git)
// [Assertions]: Git commits are generated for document changes, branches track updates, and history is maintained.
// -----------------------------------------------------------------------------
func TestGitStore(t *testing.T) {
	remoteURL := setupTestRemote(t)

	opts := Options{
		RepoURL: remoteURL,
		Branch:  "main",
	}

	store, err := NewStore(opts)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()

	// 1. Create
	doc := gyrus.Concept{
		ID:     "test-doc",
		Tenant: "root", Category: gyrus.CategoryTechnical, Type: gyrus.TypeGuide,
		Title:   "Test Document",
		Content: "This is a test document.",
		Status:  "draft", Name: "root/reference/test-doc",
	}
	ref, err := store.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if ref.Version != 1 {
		t.Errorf("expected version 1, got %d", ref.Version)
	}

	// 2. Get
	fetched, err := store.Get(ctx, "root/reference/test-doc")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Title != "Test Document" {
		t.Errorf("expected title 'Test Document', got %s", fetched.Title)
	}
	if fetched.Content != "This is a test document." {
		t.Errorf("expected content 'This is a test document.', got %s", fetched.Content)
	}

	// 3. Update
	newTitle := "Updated Title"
	patch := gyrus.ConceptPatch{
		Title:  &newTitle,
		Reason: "testing update",
	}
	ref, err = store.Update(ctx, "root/reference/test-doc", patch, fetched.Version)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if ref.Version != 2 {
		t.Errorf("expected version 2, got %d", ref.Version)
	}

	fetched2, _ := store.Get(ctx, "root/reference/test-doc")
	if fetched2.Title != "Updated Title" {
		t.Errorf("expected updated title")
	}

	// 4. Archive
	err = store.Archive(ctx, "root/reference/test-doc")
	if err != nil {
		t.Fatalf("Archive failed: %v", err)
	}
	fetched3, _ := store.Get(ctx, "root/reference/test-doc")
	if fetched3.Status != "archived" {
		t.Errorf("expected status archived, got %s", fetched3.Status)
	}

	// 5. Delete
	err = store.Delete(ctx, "root/reference/test-doc")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = store.Get(ctx, "root/reference/test-doc")
	if err == nil {
		t.Errorf("expected error getting deleted doc")
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that Git Store implements gyrus.SchemaStore for schema CRUD operations under .gyrus/schemas/.
// [Execution Surface]: In-Memory / Local Bare Git Repository (go-git)
// [Assertions]: Git commits are generated for schema Save, Get reads file, List enumerates schemas, and Delete commits removal.
// -----------------------------------------------------------------------------
func TestGitSchemaStore(t *testing.T) {
	remoteURL := setupTestRemote(t)

	opts := Options{
		RepoURL: remoteURL,
		Branch:  "main",
	}

	store, err := NewStore(opts)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	var schemaStore gyrus.SchemaStore = store
	ctx := context.Background()

	// 1. Initial list empty
	list, err := schemaStore.ListSchemas(ctx)
	if err != nil {
		t.Fatalf("ListSchemas failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 schemas initially, got %d", len(list))
	}

	// 2. Save schema
	content := `---
id: <unique-id>
title: <Title>
category: technical
type: guide
tenant: dev
version: 1
status: active
---

# Developer Guide Template
`
	if err := schemaStore.SaveSchema(ctx, "guide", content); err != nil {
		t.Fatalf("SaveSchema failed: %v", err)
	}

	// 3. List schemas
	list, err = schemaStore.ListSchemas(ctx)
	if err != nil {
		t.Fatalf("ListSchemas failed: %v", err)
	}
	if len(list) != 1 || list[0] != "guide" {
		t.Errorf("expected ['guide'], got: %v", list)
	}

	// 4. Get schema
	fetched, err := schemaStore.GetSchema(ctx, "guide")
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}
	if fetched != content {
		t.Errorf("fetched schema content mismatch")
	}

	// 5. Delete schema
	if err := schemaStore.DeleteSchema(ctx, "guide"); err != nil {
		t.Fatalf("DeleteSchema failed: %v", err)
	}

	// 6. Verify deleted
	_, err = schemaStore.GetSchema(ctx, "guide")
	if err == nil {
		t.Fatal("expected error getting deleted schema, got nil")
	}
}
