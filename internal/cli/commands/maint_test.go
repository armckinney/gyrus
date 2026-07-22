package commands_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/cli"
	_ "github.com/armckinney/gyrus/internal/cli/commands"
)

func TestCLILinkAndSyncAndValidate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gyrus-cli-maint-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storagePath := filepath.Join(tempDir, "storage")

	// 1. gyrus create 2 docs
	cli.RootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"create",
		"--id", "doc-1",
		"--title", "Doc One",
		"--category", "architecture",
		"--type", "adr",
		"--owner-group", "team",
		"--content", "Body 1",
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("create doc-1 failed: %v", err)
	}

	cli.RootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"create",
		"--id", "doc-2",
		"--title", "Doc Two",
		"--category", "architecture",
		"--type", "specification",
		"--owner-group", "team",
		"--content", "Body 2",
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("create doc-2 failed: %v", err)
	}

	// 2. gyrus link
	cli.RootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"link",
		"doc-1", "doc-2",
		"--rel-type", "depends_on",
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("link failed: %v", err)
	}

	// 3. gyrus sync
	cli.RootCmd.SetArgs([]string{
		"--storage-path", storagePath,
		"sync",
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	// 4. gyrus validate file
	docFile := filepath.Join(storagePath, "okf", "team", "reference", "doc-1.md")
	cli.RootCmd.SetArgs([]string{
		"validate",
		docFile,
	})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
}
