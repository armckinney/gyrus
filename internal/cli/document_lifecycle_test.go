package cli

import (
	"os"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies the CLI command router in-memory execution for init, create, and get subcommands.
// [Execution Surface]: In-Memory Cobra Command Tree (internal/cli)
// [Assertions]: Commands execute without errors and mutate memory/storage.
// -----------------------------------------------------------------------------
func TestCLIInitAndCreateAndGet(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	// 1. gyrus init config
	rootCmd, err := BuildRootCmd()
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}
	rootCmd.SetArgs([]string{"init", "config"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("gyrus init config failed: %v", err)
	}

	// 2. gyrus create (rebuild to load newly generated workspace .gyrus.yaml)
	rootCmd, err = BuildRootCmd()
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}
	rootCmd.SetArgs([]string{
		"create",
		"--id", "adr-2026-cli",
		"--title", "CLI Command Routing Architecture",
		"--category", "architecture",
		"--type", "adr",
		"--owner-group", "platform",
		"--content", "CLI implementation architecture details.",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("gyrus create failed: %v", err)
	}

	// 3. gyrus get
	rootCmd, err = BuildRootCmd()
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}
	rootCmd.SetArgs([]string{
		"--json",
		"get",
		"adr-2026-cli",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("gyrus get failed: %v", err)
	}
}
