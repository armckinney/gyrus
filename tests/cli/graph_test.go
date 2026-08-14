package cli_test

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus link' and 'gyrus unlink' manage directed relationship edges between documents.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: 'link' creates directed edge and confirms in stdout; 'unlink' removes edge and confirms in stdout.
// -----------------------------------------------------------------------------
func TestCLI_LinkAndUnlink(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Init
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	// Create 2 test documents
	doc1Cmd := exec.Command(gyrusBinPath, "create",
		"--id", "prd-graph-001",
		"--title", "Product Roadmap PRD",
		"--category", "product",
		"--type", "prd",
		"--owner-group", "product-team",
		"--status", "active",
		"--content", "Roadmap planning document.",
	)
	doc1Cmd.Dir = tempWorkspace
	if out, err := doc1Cmd.CombinedOutput(); err != nil {
		t.Fatalf("Create doc1 failed: %v\nOutput: %s", err, string(out))
	}

	doc2Cmd := exec.Command(gyrusBinPath, "create",
		"--id", "adr-graph-001",
		"--title", "Storage Graph Driver",
		"--category", "architecture",
		"--type", "adr",
		"--owner-group", "platform-team",
		"--status", "accepted",
		"--content", "SQLite and Postgres graph relationship edges.",
	)
	doc2Cmd.Dir = tempWorkspace
	if out, err := doc2Cmd.CombinedOutput(); err != nil {
		t.Fatalf("Create doc2 failed: %v\nOutput: %s", err, string(out))
	}

	// 1. Link (prd-graph-001 -> adr-graph-001 with depends_on)
	linkCmd := exec.Command(gyrusBinPath, "link", "prd-graph-001", "adr-graph-001", "--rel-type", "depends_on")
	linkCmd.Dir = tempWorkspace
	var linkOut bytes.Buffer
	linkCmd.Stdout = &linkOut
	if err := linkCmd.Run(); err != nil {
		t.Fatalf("Link command failed: %v", err)
	}

	if !strings.Contains(linkOut.String(), "Linked") || !strings.Contains(linkOut.String(), "prd-graph-001") {
		t.Errorf("Expected link confirmation in stdout, got:\n%s", linkOut.String())
	}

	// 2. Link with implements relationship
	linkImplCmd := exec.Command(gyrusBinPath, "link", "prd-graph-001", "adr-graph-001", "--rel-type", "implements")
	linkImplCmd.Dir = tempWorkspace
	if out, err := linkImplCmd.CombinedOutput(); err != nil {
		t.Fatalf("Link implements failed: %v\nOutput: %s", err, string(out))
	}

	// 3. Unlink depends_on
	unlinkCmd := exec.Command(gyrusBinPath, "unlink", "prd-graph-001", "adr-graph-001", "--rel-type", "depends_on")
	unlinkCmd.Dir = tempWorkspace
	var unlinkOut bytes.Buffer
	unlinkCmd.Stdout = &unlinkOut
	if err := unlinkCmd.Run(); err != nil {
		t.Fatalf("Unlink command failed: %v", err)
	}

	if !strings.Contains(unlinkOut.String(), "Unlinked") || !strings.Contains(unlinkOut.String(), "depends_on") {
		t.Errorf("Expected unlink confirmation in stdout, got:\n%s", unlinkOut.String())
	}
}
