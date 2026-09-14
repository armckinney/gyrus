package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/cli/commands"
	"github.com/spf13/cobra"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that commands.Register successfully attaches all subcommands to root and executes them in-memory.
// [Execution Surface]: In-Memory Cobra Command Tree
// [Assertions]: Root command contains all subcommands and runs init, create, and get without errors.
// -----------------------------------------------------------------------------
func TestCommandsRegistrationAndExecution(t *testing.T) {
	tempDir := t.TempDir()
	origWd, err := os.Getwd()
	if err == nil {
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(origWd) }()
	}
	storagePath := filepath.Join(tempDir, "storage")

	application, err := app.New(storagePath)
	if err != nil {
		t.Fatalf("Failed creating app container: %v", err)
	}

	rootCmd := &cobra.Command{Use: "gyrus"}
	commands.Register(rootCmd, application)

	// 1. gyrus init config
	rootCmd.SetArgs([]string{"init", "config", "--profile", "local"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Init config command failed: %v", err)
	}

	// 1b. gyrus init client
	rootCmd.SetArgs([]string{"init", "client", "--target", "antigravity", "--mode", "local"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Init client command failed: %v", err)
	}

	// 1c. gyrus init bare (must error)
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatalf("Expected bare init command to fail, but succeeded")
	}

	// 2. gyrus create
	rootCmd.SetArgs([]string{
		"create",
		"--id", "doc-cmd-001",
		"--title", "CLI Commands Architecture",
		"--category", "architecture",
		"--type", "adr",
		"--owner-group", "platform",
		"--content", "Modular subpackages for CLI commands.",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Create command failed: %v", err)
	}

	// 3. gyrus get
	rootCmd.SetArgs([]string{"get", "doc-cmd-001"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Get command failed: %v", err)
	}
	_ = os.RemoveAll
}
