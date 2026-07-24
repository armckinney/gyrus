package blob

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"gocloud.dev/blob/fileblob"
)

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
	doc := gyrus.Document{
		ID:             "adr-2026-001",
		Title:          "Test ADR",
		Category:       gyrus.CategoryArchitecture,
		Type:           gyrus.TypeADR,
		Format:         "markdown",
		OwnerGroup:     "platform",
		Version:        1,
		Status:         "draft",
		LastModifiedBy: "tester",
		LastUpdated:    now,
		Tags:           []string{"test"},
		Content:        "# Hello World\nThis is a test.",
	}

	// Test Create
	ref, err := store.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if ref.ID != doc.ID {
		t.Errorf("expected ref ID %q, got %q", doc.ID, ref.ID)
	}

	// Test Get
	fetched, err := store.Get(ctx, doc.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if fetched.Title != doc.Title {
		t.Errorf("expected title %q, got %q", doc.Title, fetched.Title)
	}
	if fetched.Content != doc.Content {
		t.Errorf("expected content %q, got %q", doc.Content, fetched.Content)
	}

	// Test Update
	newTitle := "Updated ADR"
	patch := gyrus.DocumentPatch{
		Title: &newTitle,
	}
	
	updRef, err := store.Update(ctx, doc.ID, patch, 1)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updRef.Version != 2 {
		t.Errorf("expected version 2, got %d", updRef.Version)
	}

	fetched2, err := store.Get(ctx, doc.ID)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if fetched2.Title != newTitle {
		t.Errorf("expected updated title %q, got %q", newTitle, fetched2.Title)
	}

	// Test Archive
	err = store.Archive(ctx, doc.ID)
	if err != nil {
		t.Fatalf("Archive failed: %v", err)
	}
	
	fetched3, err := store.Get(ctx, doc.ID)
	if err != nil {
		t.Fatalf("Get after archive failed: %v", err)
	}
	if fetched3.Status != "archived" {
		t.Errorf("expected status 'archived', got %q", fetched3.Status)
	}

	// Test Delete
	err = store.Delete(ctx, doc.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	
	_, err = store.Get(ctx, doc.ID)
	if err == nil {
		t.Fatalf("expected error getting deleted document")
	}
}
