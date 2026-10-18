package localfs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that LocalFS Store creates, reads, updates, and deletes Markdown files on disk with optimistic concurrency locking.
// [Execution Surface]: Real Local Filesystem
// [Assertions]: File creation follows OKF path convention, updates increment version, concurrency conflict fails outdated versions, delete removes file.
// -----------------------------------------------------------------------------
func TestLocalfsStoreCRUD(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-localfs-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := localfs.NewStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize Store: %v", err)
	}

	ctx := context.Background()

	concept := gyrus.Concept{
		ID:             "guide-2026-auth",
		Title:          "OAuth2 Authentication Guide",
		Category:       gyrus.CategoryArchitecture,
		Type:           gyrus.TypeGuide,
		Tenant:         "security",
		Version:        1,
		Status:         "active",
		LastModifiedBy: "sec-engineer",
		Content:        "# OAuth2 Guide\n\nInstructions here.", Name: "security/reference/guide-2026-auth",
	}

	// Create
	ref, err := store.Create(ctx, concept)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if ref.ID != concept.ID {
		t.Errorf("Expected Ref ID %s, got %s", concept.ID, ref.ID)
	}

	// Get
	fetched, err := store.Get(ctx, concept.Name)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Title != concept.Title {
		t.Errorf("Expected Title '%s', got '%s'", concept.Title, fetched.Title)
	}
	if fetched.Tenant != concept.Tenant {
		t.Errorf("Expected Tenant '%s', got '%s'", concept.Tenant, fetched.Tenant)
	}

	// Update
	newTitle := "OAuth2 & OIDC Authentication Guide"
	patch := gyrus.ConceptPatch{
		Title: &newTitle,
	}
	updatedRef, err := store.Update(ctx, concept.Name, patch, 1)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updatedRef.Version != 2 {
		t.Errorf("Expected version 2, got %d", updatedRef.Version)
	}

	// Concurrency conflict test
	_, err = store.Update(ctx, concept.Name, patch, 1)
	if err == nil {
		t.Fatal("Expected concurrency error for outdated version 1, got nil")
	}

	// Delete
	if err := store.Delete(ctx, concept.Name); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = store.Get(ctx, concept.Name)
	if err == nil {
		t.Fatal("Expected error fetching deleted concept, got nil")
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that LocalFS Store implements gyrus.SchemaStore for CRUD operations under .gyrus/schemas/.
// [Execution Surface]: Real Local Filesystem
// [Assertions]: Saves schema to .gyrus/schemas/<conceptType>.md, lists persisted schemas, retrieves content, and deletes schema.
// -----------------------------------------------------------------------------
func TestLocalfsSchemaStoreCRUD(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-localfs-schema-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := localfs.NewStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize Store: %v", err)
	}

	var schemaStore gyrus.SchemaStore = store
	ctx := context.Background()

	// 1. Initially empty
	schemas, err := schemaStore.ListSchemas(ctx)
	if err != nil {
		t.Fatalf("ListSchemas failed: %v", err)
	}
	if len(schemas) != 0 {
		t.Errorf("Expected 0 schemas initially, got %d", len(schemas))
	}

	// 2. Save schema
	content := `---
id: <unique-id>
title: <Title>
category: operations
type: runbook
tenant: ops
version: 1
status: active
---

# Runbook Template
`
	if err := schemaStore.SaveSchema(ctx, "runbook", content); err != nil {
		t.Fatalf("SaveSchema failed: %v", err)
	}

	// 3. List schemas
	schemas, err = schemaStore.ListSchemas(ctx)
	if err != nil {
		t.Fatalf("ListSchemas failed: %v", err)
	}
	if len(schemas) != 1 || schemas[0] != "runbook" {
		t.Errorf("Expected ['runbook'], got: %v", schemas)
	}

	// 4. Get schema
	fetched, err := schemaStore.GetSchema(ctx, "runbook")
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}
	if fetched != content {
		t.Errorf("Fetched schema content mismatch")
	}

	// 5. Delete schema
	if err := schemaStore.DeleteSchema(ctx, "runbook"); err != nil {
		t.Fatalf("DeleteSchema failed: %v", err)
	}

	// 6. Verify deleted
	_, err = schemaStore.GetSchema(ctx, "runbook")
	if err == nil {
		t.Fatal("Expected error getting deleted schema, got nil")
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that LocalFS Store routes concepts to projects/<project>/ and reference/ paths and parses scope on Get.
// [Assertions]: Project concepts persist to projects/<project>/<id>.md, reference concepts to reference/<id>.md.
// -----------------------------------------------------------------------------
func TestLocalfsStoreProjectScoping(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-localfs-project-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := localfs.NewStoreWithProject(tempDir, "gyrus")
	if err != nil {
		t.Fatalf("Failed to initialize Store: %v", err)
	}

	ctx := context.Background()

	// 1. Create project concept
	wsConcept := gyrus.Concept{
		ID:       "ticket-101",
		Title:    "Implement Scoping",
		Category: gyrus.CategoryProduct,
		Type:     gyrus.TypePRD,
		Tenant:   "root",
		Version:  1,
		Status:   "active",
		Content:  "# Implementation Ticket", Name: "root/projects/gyrus/ticket-101",
	}

	if _, err := store.Create(ctx, wsConcept); err != nil {
		t.Fatalf("Failed creating wsConcept: %v", err)
	}

	// 2. Create reference concept
	refConcept := gyrus.Concept{
		ID:       "adr-007",
		Title:    "Scoping Architecture",
		Category: gyrus.CategoryArchitecture,
		Type:     gyrus.TypeADR,
		Tenant:   "root",
		Scope:    "reference",
		Version:  1,
		Status:   "proposed",
		Content:  "# ADR 007", Name: "root/reference/adr-007",
	}

	if _, err := store.Create(ctx, refConcept); err != nil {
		t.Fatalf("Failed creating refConcept: %v", err)
	}

	// Verify file path on disk for wsConcept
	expectedWsPath := filepath.Join(tempDir, "concepts/root/projects/gyrus/ticket-101.md")
	if _, err := os.Stat(expectedWsPath); err != nil {
		t.Fatalf("Expected project file at %s, got: %v", expectedWsPath, err)
	}

	// Verify file path on disk for refConcept
	expectedRefPath := filepath.Join(tempDir, "concepts/root/reference/adr-007.md")
	if _, err := os.Stat(expectedRefPath); err != nil {
		t.Fatalf("Expected reference file at %s, got: %v", expectedRefPath, err)
	}

	// 3. Test Get() parses scope and project
	fetchedWs, err := store.Get(ctx, "root/projects/gyrus/ticket-101")
	if err != nil {
		t.Fatalf("Get wsConcept failed: %v", err)
	}
	if fetchedWs.Scope != "project" {
		t.Errorf("Expected Scope 'project', got '%s'", fetchedWs.Scope)
	}
	if fetchedWs.Project != "gyrus" {
		t.Errorf("Expected Project 'gyrus', got '%s'", fetchedWs.Project)
	}

	fetchedRef, err := store.Get(ctx, "root/reference/adr-007")
	if err != nil {
		t.Fatalf("Get refConcept failed: %v", err)
	}
	if fetchedRef.Scope != "reference" {
		t.Errorf("Expected Scope 'reference', got '%s'", fetchedRef.Scope)
	}
	if fetchedRef.Project != "" {
		t.Errorf("Expected empty Project for reference concept, got '%s'", fetchedRef.Project)
	}
}
