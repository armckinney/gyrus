package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies search, suggest-context, and schema CLI subcommands.
// [Execution Surface]: In-Memory Cobra Command Tree (internal/cli)
// [Assertions]: Commands execute without errors and return results.
// -----------------------------------------------------------------------------
func TestCLISearchAndSuggestAndSchema(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	wsYaml := `storage:
  provider: localfs
  root: .gyrus
`
	if err := os.WriteFile(filepath.Join(tempDir, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing .gyrus.yaml: %v", err)
	}

	rootCmd, err := BuildRootCmd()
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}

	// 1. gyrus create doc
	rootCmd.SetArgs([]string{
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
		"sync",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	// 3. gyrus search
	rootCmd.SetArgs([]string{
		"search",
		"--query", "search",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("search failed: %v", err)
	}

	// 4. gyrus suggest-context
	rootCmd.SetArgs([]string{
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
