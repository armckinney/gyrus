package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCLILinkAndSyncAndValidate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-cli-maint-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storagePath := filepath.Join(tempDir, "storage")

	rootCmd, err := BuildRootCmd(storagePath)
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}

	// 1. gyrus create 2 docs
	rootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"create",
		"--id", "doc-1",
		"--title", "Doc One",
		"--category", "architecture",
		"--type", "adr",
		"--owner-group", "team",
		"--content", "Body 1",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("create doc-1 failed: %v", err)
	}

	rootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"create",
		"--id", "doc-2",
		"--title", "Doc Two",
		"--category", "architecture",
		"--type", "specification",
		"--owner-group", "team",
		"--content", "Body 2",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("create doc-2 failed: %v", err)
	}

	// 2. gyrus link
	rootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"link",
		"doc-1", "doc-2",
		"--rel-type", "depends_on",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	// 3. gyrus sync
	rootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"sync",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	// 4. gyrus validate file
	docFile := filepath.Join(storagePath, "docs", "team", "reference", "doc-1.md")
	rootCmd.SetArgs([]string{
		"validate",
		docFile,
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
}
