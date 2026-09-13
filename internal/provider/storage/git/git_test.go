package git

import (
	"context"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/go-git/go-git/v5"
)

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
	doc := gyrus.Document{
		ID:      "test-doc",
		Title:   "Test Document",
		Content: "This is a test document.",
		Status:  "draft",
	}
	ref, err := store.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if ref.Version != 1 {
		t.Errorf("expected version 1, got %d", ref.Version)
	}

	// 2. Get
	fetched, err := store.Get(ctx, "test-doc")
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
	patch := gyrus.DocumentPatch{
		Title:  &newTitle,
		Reason: "testing update",
	}
	ref, err = store.Update(ctx, "test-doc", patch, fetched.Version)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if ref.Version != 2 {
		t.Errorf("expected version 2, got %d", ref.Version)
	}

	fetched2, _ := store.Get(ctx, "test-doc")
	if fetched2.Title != "Updated Title" {
		t.Errorf("expected updated title")
	}

	// 4. Archive
	err = store.Archive(ctx, "test-doc")
	if err != nil {
		t.Fatalf("Archive failed: %v", err)
	}
	fetched3, _ := store.Get(ctx, "test-doc")
	if fetched3.Status != "archived" {
		t.Errorf("expected status archived, got %s", fetched3.Status)
	}

	// 5. Delete
	err = store.Delete(ctx, "test-doc")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = store.Get(ctx, "test-doc")
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
owner_group: dev
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
