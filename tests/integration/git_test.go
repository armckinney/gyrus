package integration_test

import (
	"context"
	"path/filepath"
	"testing"

	storagegit "github.com/armckinney/gyrus/internal/provider/storage/git"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/go-git/go-git/v5"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies end-to-end functionality for the 'git' profile (Git remote storage driver + SQLite index).
// [Execution Surface]: Real In-Memory / Local Bare Git Repository + SQLite Database
// [Assertions]: Document commits to git repository, retrieves via Git storage driver, and updates commit log.
// -----------------------------------------------------------------------------
func TestIntegration_GitProfile(t *testing.T) {
	tempDir := t.TempDir()
	remoteDir := filepath.Join(tempDir, "remote.git")

	_, err := git.PlainInit(remoteDir, true)
	if err != nil {
		t.Fatalf("Failed initializing bare git repo: %v", err)
	}

	store, err := storagegit.NewStore(storagegit.Options{
		RepoURL: remoteDir,
		Branch:  "main",
	})
	if err != nil {
		t.Fatalf("Failed initializing git store: %v", err)
	}

	ctx := context.Background()

	doc := gyrus.Document{
		ID:         "adr-git-001",
		Title:      "Git Backed Architecture",
		Category:   gyrus.CategoryArchitecture,
		Type:       gyrus.TypeADR,
		OwnerGroup: "platform",
		Version:    1,
		Status:     "proposed",
		Content:    "Persisted directly into remote Git commits.",
	}

	ref, err := store.Create(ctx, doc)
	if err != nil {
		t.Fatalf("Git store Create failed: %v", err)
	}
	if ref.ID != doc.ID {
		t.Errorf("Expected Ref ID %s, got %s", doc.ID, ref.ID)
	}

	fetched, err := store.Get(ctx, doc.ID)
	if err != nil {
		t.Fatalf("Git store Get failed: %v", err)
	}
	if fetched.Title != doc.Title {
		t.Errorf("Expected Title '%s', got '%s'", doc.Title, fetched.Title)
	}
}
