package localfs_test

import (
	"context"
	"os"
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

	doc := gyrus.Document{
		ID:             "guide-2026-auth",
		Title:          "OAuth2 Authentication Guide",
		Category:       gyrus.CategoryArchitecture,
		Type:           gyrus.TypeGuide,
		OwnerGroup:     "security",
		Version:        1,
		Status:         "active",
		LastModifiedBy: "sec-engineer",
		Content:        "# OAuth2 Guide\n\nInstructions here.",
	}

	// Create
	ref, err := store.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if ref.ID != doc.ID {
		t.Errorf("Expected Ref ID %s, got %s", doc.ID, ref.ID)
	}

	// Get
	fetched, err := store.Get(ctx, doc.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Title != doc.Title {
		t.Errorf("Expected Title '%s', got '%s'", doc.Title, fetched.Title)
	}

	// Update
	newTitle := "OAuth2 & OIDC Authentication Guide"
	patch := gyrus.DocumentPatch{
		Title: &newTitle,
	}
	updatedRef, err := store.Update(ctx, doc.ID, patch, 1)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updatedRef.Version != 2 {
		t.Errorf("Expected version 2, got %d", updatedRef.Version)
	}

	// Concurrency conflict test
	_, err = store.Update(ctx, doc.ID, patch, 1)
	if err == nil {
		t.Fatal("Expected concurrency error for outdated version 1, got nil")
	}

	// Delete
	if err := store.Delete(ctx, doc.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	_, err = store.Get(ctx, doc.ID)
	if err == nil {
		t.Fatal("Expected error fetching deleted document, got nil")
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that LocalFS Store implements gyrus.SchemaStore for CRUD operations under .gyrus/schemas/.
// [Execution Surface]: Real Local Filesystem
// [Assertions]: Saves schema to .gyrus/schemas/<docType>.md, lists persisted schemas, retrieves content, and deletes schema.
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
owner_group: ops
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
// [Purpose]: Verifies that LocalFS Store routes documents to workspaces/<ws>/ and reference/ paths and parses scope on Get.
// [Assertions]: Workspace documents persist to workspaces/<ws>/<id>.md, reference docs to reference/<id>.md.
// -----------------------------------------------------------------------------
func TestLocalfsStoreWorkspaceScoping(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-localfs-ws-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := localfs.NewStoreWithWorkspace(tempDir, "gyrus")
	if err != nil {
		t.Fatalf("Failed to initialize Store: %v", err)
	}

	ctx := context.Background()

	// 1. Create workspace document
	wsDoc := gyrus.Document{
		ID:         "ticket-101",
		Title:      "Implement Scoping",
		Category:   gyrus.CategoryProduct,
		Type:       gyrus.TypePRD,
		OwnerGroup: "root",
		Version:    1,
		Status:     "active",
		Content:    "# Implementation Ticket",
	}

	if _, err := store.Create(ctx, wsDoc); err != nil {
		t.Fatalf("Failed creating wsDoc: %v", err)
	}

	// 2. Create reference document
	refDoc := gyrus.Document{
		ID:         "adr-007",
		Title:      "Scoping Architecture",
		Category:   gyrus.CategoryArchitecture,
		Type:       gyrus.TypeADR,
		OwnerGroup: "root",
		Scope:      "reference",
		Version:    1,
		Status:     "proposed",
		Content:    "# ADR 007",
	}

	if _, err := store.Create(ctx, refDoc); err != nil {
		t.Fatalf("Failed creating refDoc: %v", err)
	}

	// Verify file path on disk for wsDoc
	expectedWsPath := tempDir + "/docs/root/workspaces/gyrus/ticket-101.md"
	if _, err := os.Stat(expectedWsPath); err != nil {
		t.Fatalf("Expected workspace file at %s, got: %v", expectedWsPath, err)
	}

	// Verify file path on disk for refDoc
	expectedRefPath := tempDir + "/docs/root/reference/adr-007.md"
	if _, err := os.Stat(expectedRefPath); err != nil {
		t.Fatalf("Expected reference file at %s, got: %v", expectedRefPath, err)
	}

	// 3. Test Get() parses scope and workspace
	fetchedWs, err := store.Get(ctx, "ticket-101")
	if err != nil {
		t.Fatalf("Get wsDoc failed: %v", err)
	}
	if fetchedWs.Scope != "workspace" {
		t.Errorf("Expected Scope 'workspace', got '%s'", fetchedWs.Scope)
	}
	if fetchedWs.Workspace != "gyrus" {
		t.Errorf("Expected Workspace 'gyrus', got '%s'", fetchedWs.Workspace)
	}

	fetchedRef, err := store.Get(ctx, "adr-007")
	if err != nil {
		t.Fatalf("Get refDoc failed: %v", err)
	}
	if fetchedRef.Scope != "reference" {
		t.Errorf("Expected Scope 'reference', got '%s'", fetchedRef.Scope)
	}
	if fetchedRef.Workspace != "" {
		t.Errorf("Expected empty Workspace for reference doc, got '%s'", fetchedRef.Workspace)
	}
}
