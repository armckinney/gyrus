package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies maintenance CLI commands: link, sync, and validate.
// [Execution Surface]: In-Memory Cobra Command Tree (internal/cli)
// [Assertions]: Commands execute without errors and mutate memory/storage.
// -----------------------------------------------------------------------------
func TestCLILinkAndSyncAndValidate(t *testing.T) {
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

	// 1. gyrus create 2 docs
	rootCmd.SetArgs([]string{
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
		"link",
		"doc-1", "doc-2",
		"--rel-type", "depends_on",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	// 3. gyrus sync
	rootCmd.SetArgs([]string{
		"sync",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	// 4. gyrus validate file
	docFile := filepath.Join(tempDir, ".gyrus", "docs", "team", "reference", "doc-1.md")
	rootCmd.SetArgs([]string{
		"validate",
		docFile,
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
}
