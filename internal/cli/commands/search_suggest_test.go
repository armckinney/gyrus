package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/cli"
	_ "github.com/armckinney/gyrus/internal/cli/commands"
)

func TestCLISearchAndSuggestAndSchema(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-cli-search-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storagePath := filepath.Join(tempDir, "storage")

	// 1. gyrus create doc
	cli.RootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"create",
		"--id", "adr-001-test",
		"--title", "Search Index Test ADR",
		"--category", "architecture",
		"--type", "adr",
		"--owner-group", "platform",
		"--content", "We test full-text search capability.",
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 2. gyrus sync
	cli.RootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"sync",
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	// 3. gyrus search
	cli.RootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"search",
		"--query", "search",
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("search failed: %v", err)
	}

	// 4. gyrus suggest-context
	cli.RootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"suggest-context",
		"--prompt", "How to implement search?",
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("suggest-context failed: %v", err)
	}

	// 5. gyrus schema
	cli.RootCmd.SetArgs([]string{
		"schema",
		"adr",
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("schema failed: %v", err)
	}
}
