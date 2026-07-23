package localfs_test

import (
	"context"
	"os"
	"testing"

	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

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

func TestResolveStoragePathPrecedence(t *testing.T) {
	// Flag precedence
	flagPath := "/tmp/gyrus-flag-path"
	resolved, err := localfs.ResolveStoragePath(flagPath)
	if err != nil {
		t.Fatalf("Unexpected error resolving flag path: %v", err)
	}
	if resolved != flagPath {
		t.Errorf("Expected '%s', got '%s'", flagPath, resolved)
	}
}

func TestDotGyrusYamlConfigResolution(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-dot-config-*")
	if err != nil {
		t.Fatalf("Failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	// Write .gyrus.yaml in tempDir
	configContent := `version: 1
storage:
  root: ./my-docs
`
	if err := os.WriteFile(tempDir+"/.gyrus.yaml", []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed writing .gyrus.yaml: %v", err)
	}

	_ = os.Chdir(tempDir)

	resolved, err := localfs.ResolveStoragePath("")
	if err != nil {
		t.Fatalf("ResolveStoragePath failed: %v", err)
	}

	expected := tempDir + "/my-docs"
	if resolved != expected {
		t.Errorf("Expected resolved path '%s', got '%s'", expected, resolved)
	}
}
