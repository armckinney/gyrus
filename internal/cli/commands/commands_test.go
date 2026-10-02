package commands_test

import (
	"os"
	"testing"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/cli/commands"
	"github.com/spf13/cobra"
)

func TestMain(m *testing.M) {
	os.Unsetenv("GYRUS_WORKSPACE")
	os.Unsetenv("GYRUS_CONFIG")
	os.Exit(m.Run())
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that commands.Register successfully attaches all subcommands to root and executes them in-memory.
// [Execution Surface]: In-Memory Cobra Command Tree
// [Assertions]: Root command contains all subcommands and runs init, create, and get without errors.
// -----------------------------------------------------------------------------
func TestCommandsRegistrationAndExecution(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	origWd, err := os.Getwd()
	if err == nil {
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(origWd) }()
	}

	application, err := app.New()
	if err != nil {
		t.Fatalf("Failed creating app container: %v", err)
	}

	rootCmd := &cobra.Command{Use: "gyrus"}
	commands.Register(rootCmd, application)

	// 1. gyrus config init
	rootCmd.SetArgs([]string{"config", "init", "--profile", "local"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Config init command failed: %v", err)
	}
	if err := application.Reset(); err != nil {
		t.Fatalf("Application reset failed: %v", err)
	}

	// 1b. gyrus client install
	rootCmd.SetArgs([]string{"client", "install", "--target", "antigravity", "--mode", "local"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Client install command failed: %v", err)
	}

	// 1c. gyrus init bare (must error)
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatalf("Expected bare init command to fail, but succeeded")
	}

	// 2. gyrus create with explicit owner-group
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

	// 2b. gyrus create without owner-group (should default to root)
	rootCmd.SetArgs([]string{
		"create",
		"--id", "doc-cmd-default-owner",
		"--title", "Default Owner Architecture",
		"--category", "architecture",
		"--type", "adr",
		"--content", "Testing default root owner.",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Create command without owner-group failed: %v", err)
	}

	// 3. gyrus get
	rootCmd.SetArgs([]string{"get", "doc-cmd-001"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Get command failed: %v", err)
	}

	// 4. Verify gyrus ui and gyrus ui serve registration
	uiCmd, _, err := rootCmd.Find([]string{"ui"})
	if err != nil || uiCmd == nil {
		t.Fatalf("Expected 'ui' command to be registered, found error: %v", err)
	}
	uiServeCmd, _, err := rootCmd.Find([]string{"ui", "serve"})
	if err != nil || uiServeCmd == nil {
		t.Fatalf("Expected 'ui serve' subcommand to be registered, found error: %v", err)
	}
	if uiServeCmd.Flags().Lookup("port") == nil {
		t.Errorf("Expected --port flag on 'ui serve'")
	}
	if uiServeCmd.Flags().Lookup("host") == nil {
		t.Errorf("Expected --host flag on 'ui serve'")
	}
	if uiServeCmd.Flags().Lookup("workspace") == nil {
		t.Errorf("Expected --workspace flag on 'ui serve'")
	}
	if uiServeCmd.Flags().Lookup("config") == nil {
		t.Errorf("Expected --config flag on 'ui serve'")
	}

	_ = os.RemoveAll
}
