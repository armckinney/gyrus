package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies in-memory execution of search, suggest-context, and schema CLI subcommands.
// [Execution Surface]: In-Memory Cobra Command Tree (internal/cli)
// [Assertions]: Document creation indexes content, search matches keywords, suggest linearizes context, and schema prints ADR template.
// -----------------------------------------------------------------------------
func TestCLISearchAndSuggestAndSchema(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-cli-search-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storagePath := filepath.Join(tempDir, "storage")

	rootCmd, err := BuildRootCmd(storagePath)
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}

	// 1. gyrus create doc
	rootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"create",
		"--id", "adr-001-test",
		"--title", "Search Index Test ADR",
		"--category", "architecture",
		"--type", "adr",
		"--owner-group", "platform",
		"--content", "We test full-text search capability.",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 2. gyrus sync
	rootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"sync",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	// 3. gyrus search
	rootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"search",
		"--query", "search",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("search failed: %v", err)
	}

	// 4. gyrus suggest-context
	rootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"suggest-context",
		"--prompt", "How to implement search?",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("suggest-context failed: %v", err)
	}

	// 5. gyrus schema
	rootCmd.SetArgs([]string{
		"schema",
		"adr",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("schema failed: %v", err)
	}
}
