package blob

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"gocloud.dev/blob/fileblob"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that BlobStore handles concept lifecycle (Create, Get, Update, Archive, Delete) via Go CDK fileblob.
// [Execution Surface]: In-Memory / Local fileblob bucket
// [Assertions]: Concept is created, retrieved with content matching, updated with version increment, archived, and deleted.
// -----------------------------------------------------------------------------
func TestBlobStore(t *testing.T) {
	ctx := context.Background()

	dir, err := os.MkdirTemp("", "blobstore-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	bucket, err := fileblob.OpenBucket(dir, &fileblob.Options{})
	if err != nil {
		t.Fatalf("failed to open fileblob bucket: %v", err)
	}
	defer bucket.Close()

	store := NewStoreWithBucket(bucket, "testprefix")

	now := time.Now().Truncate(time.Second)
	concept := gyrus.Concept{
		ID:             "adr-2026-001",
		Title:          "Test ADR",
		Category:       gyrus.CategoryArchitecture,
		Type:           gyrus.TypeADR,
		Format:         "markdown",
		Tenant:         "platform",
		Version:        1,
		Status:         "draft",
		LastModifiedBy: "tester",
		LastUpdated:    now,
		Tags:           []string{"test"},
		Content:        "# Hello World\nThis is a test.",
	}

	// Test Create
	ref, err := store.Create(ctx, concept)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if ref.ID != concept.ID {
		t.Errorf("expected ref ID %q, got %q", concept.ID, ref.ID)
	}

	// Test Get
	fetched, err := store.Get(ctx, concept.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Title != concept.Title {
		t.Errorf("expected title %q, got %q", concept.Title, fetched.Title)
	}
	if fetched.Tenant != concept.Tenant {
		t.Errorf("expected tenant %q, got %q", concept.Tenant, fetched.Tenant)
	}
	if fetched.Content != concept.Content {
		t.Errorf("expected content %q, got %q", concept.Content, fetched.Content)
	}

	// Test Update
	newTitle := "Updated ADR"
	patch := gyrus.ConceptPatch{
		Title: &newTitle,
	}

	updRef, err := store.Update(ctx, concept.ID, patch, 1)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updRef.Version != 2 {
		t.Errorf("expected version 2, got %d", updRef.Version)
	}

	fetched2, err := store.Get(ctx, concept.ID)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if fetched2.Title != newTitle {
		t.Errorf("expected updated title %q, got %q", newTitle, fetched2.Title)
	}

	// Test Archive
	err = store.Archive(ctx, concept.ID)
	if err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	fetched3, err := store.Get(ctx, concept.ID)
	if err != nil {
		t.Fatalf("Get after archive failed: %v", err)
	}
	if fetched3.Status != "archived" {
		t.Errorf("expected status 'archived', got %q", fetched3.Status)
	}

	// Test Delete
	err = store.Delete(ctx, concept.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = store.Get(ctx, concept.ID)
	if err == nil {
		t.Fatalf("expected error getting deleted concept")
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that Blob Store implements gyrus.SchemaStore for remote schema CRUD operations.
// [Execution Surface]: In-Memory / Local fileblob bucket
// [Assertions]: Saves schema to enforced .gyrus/schemas/<conceptType>.md key, lists schemas, retrieves content, and deletes schema.
// -----------------------------------------------------------------------------
func TestBlobSchemaStore(t *testing.T) {
	ctx := context.Background()

	dir, err := os.MkdirTemp("", "blobstore-schema-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	bucket, err := fileblob.OpenBucket(dir, &fileblob.Options{})
	if err != nil {
		t.Fatalf("failed to open fileblob bucket: %v", err)
	}
	defer bucket.Close()

	store := NewStoreWithBucket(bucket, "testprefix")
	var schemaStore gyrus.SchemaStore = store

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
category: product
type: spec
tenant: product
version: 1
status: active
---

# Product Spec Template
`
	if err := schemaStore.SaveSchema(ctx, "spec", content); err != nil {
		t.Fatalf("SaveSchema failed: %v", err)
	}

	// 3. List schemas
	list, err = schemaStore.ListSchemas(ctx)
	if err != nil {
		t.Fatalf("ListSchemas failed: %v", err)
	}
	if len(list) != 1 || list[0] != "spec" {
		t.Errorf("expected ['spec'], got %v", list)
	}

	// 4. Get schema
	fetched, err := schemaStore.GetSchema(ctx, "spec")
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}
	if fetched != content {
		t.Errorf("fetched schema content mismatch")
	}

	// 5. Delete schema
	if err := schemaStore.DeleteSchema(ctx, "spec"); err != nil {
		t.Fatalf("DeleteSchema failed: %v", err)
	}

	// 6. Verify deleted
	_, err = schemaStore.GetSchema(ctx, "spec")
	if err == nil {
		t.Fatal("expected error getting deleted schema, got nil")
	}
}
