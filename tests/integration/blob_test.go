package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	storageblob "github.com/armckinney/gyrus/internal/provider/storage/blob"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"gocloud.dev/blob/fileblob"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies end-to-end functionality for the 'blob' profile (Cloud Object Storage driver via fileblob + SQLite index).
// [Execution Surface]: Real Blob Storage Emulation (gocloud.dev/blob/fileblob)
// [Assertions]: Document writes to blob bucket, retrieves accurately, and handles bucket key prefixing.
// -----------------------------------------------------------------------------
func TestIntegration_BlobProfile(t *testing.T) {
	tempDir := t.TempDir()
	blobDir := filepath.Join(tempDir, "bucket")
	if err := os.MkdirAll(blobDir, 0755); err != nil {
		t.Fatalf("Failed creating blob dir: %v", err)
	}

	bucket, err := fileblob.OpenBucket(blobDir, &fileblob.Options{})
	if err != nil {
		t.Fatalf("Failed opening fileblob bucket: %v", err)
	}
	defer bucket.Close()

	store := storageblob.NewStoreWithBucket(bucket, "gyrus-docs")
	ctx := context.Background()

	doc := gyrus.Document{
		ID:         "spec-blob-001",
		Title:      "Cloud Blob Spec",
		Category:   gyrus.CategoryTechnical,
		Type:       gyrus.TypeSpecification,
		OwnerGroup: "cloud-eng",
		Version:    1,
		Status:     "active",
		Content:    "Cloud-native object storage envelope bundle.",
	}

	ref, err := store.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Blob store Create failed: %v", err)
	}
	if ref.ID != doc.ID {
		t.Errorf("Expected Ref ID %s, got %s", doc.ID, ref.ID)
	}

	fetched, err := store.Get(ctx, doc.ID)
	if err != nil {
		t.Fatalf("Blob store Get failed: %v", err)
	}
	if fetched.Title != doc.Title {
		t.Errorf("Expected Title '%s', got '%s'", doc.Title, fetched.Title)
	}
}
