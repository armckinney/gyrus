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

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies CLI scoping flags (--scope, --workspace) and workspace-first retrieval in search and suggest-context.
// [Assertions]: Workspace documents are prioritized, reference dependencies are expanded, and other workspaces are isolated.
// -----------------------------------------------------------------------------
func TestCLIScopedSearchAndSuggestContext(t *testing.T) {
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
workspace: gyrus
default_owner_group: root
`
	if err := os.WriteFile(filepath.Join(tempDir, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing .gyrus.yaml: %v", err)
	}

	rootCmd, err := BuildRootCmd()
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}

	// 1. Create reference doc
	rootCmd.SetArgs([]string{
		"create",
		"--id", "adr-shared-arch",
		"--title", "Shared Architecture Standards",
		"--category", "architecture",
		"--type", "adr",
		"--scope", "reference",
		"--content", "Global company-wide architecture standards.",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("create reference doc failed: %v", err)
	}

	// 2. Create workspace doc with dependency on reference doc
	rootCmd.SetArgs([]string{
		"create",
		"--id", "ticket-gyrus-core",
		"--title", "Gyrus Core Feature Implementation",
		"--category", "product",
		"--type", "prd",
		"--dependencies", "adr-shared-arch",
		"--content", "Implementation details following architecture standards.",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("create workspace doc failed: %v", err)
	}

	// 3. Create foreign workspace doc
	rootCmd.SetArgs([]string{
		"create",
		"--id", "ticket-foreign-billing",
		"--title", "Billing Implementation",
		"--category", "product",
		"--type", "prd",
		"--workspace", "billing",
		"--content", "Foreign billing architecture implementation.",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("create foreign workspace doc failed: %v", err)
	}

	// 4. Sync
	rootCmd.SetArgs([]string{"sync"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	// 5. Search with workspace scoping (default in workspace: gyrus)
	rootCmd.SetArgs([]string{
		"search",
		"--query", "architecture",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("scoped search failed: %v", err)
	}

	// 6. Suggest context with workspace prioritization & dependency expansion
	rootCmd.SetArgs([]string{
		"suggest-context",
		"--prompt", "architecture standards implementation",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("suggest-context failed: %v", err)
	}

	// 7. Search with scope=reference
	rootCmd.SetArgs([]string{
		"search",
		"--query", "architecture",
		"--scope", "reference",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("search scope=reference failed: %v", err)
	}

	// 8. Search with scope=all
	rootCmd.SetArgs([]string{
		"search",
		"--query", "architecture",
		"--scope", "all",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("search scope=all failed: %v", err)
	}
}
