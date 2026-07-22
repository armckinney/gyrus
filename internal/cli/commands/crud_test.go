package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/cli"
	_ "github.com/armckinney/gyrus/internal/cli/commands"
)

func TestCLIInitAndCreateAndGet(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-cli-crud-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storagePath := filepath.Join(tempDir, "storage")

	// 1. gyrus init
	cli.RootCmd.SetArgs([]string{"--storage-path", storagePath, "init"})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("gyrus init failed: %v", err)
	}

	// 2. gyrus create
	cli.RootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"create",
		"--id", "adr-2026-cli",
		"--title", "CLI Command Routing Architecture",
		"--category", "architecture",
		"--type", "adr",
		"--owner-group", "platform",
		"--content", "CLI implementation architecture details.",
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("gyrus create failed: %v", err)
	}

	// 3. gyrus get
	cli.RootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"--json",
		"get",
		"adr-2026-cli",
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("gyrus get failed: %v", err)
	}
}
