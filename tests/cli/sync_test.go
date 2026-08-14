package cli_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus sync' scans disk and indexes all markdown files into search/index providers.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Sync scans documents, updates search index, and reports indexed file counts.
// -----------------------------------------------------------------------------
func TestCLI_Sync(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Init
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	// Create a document directly via create
	createCmd := exec.Command(gyrusBinPath, "create",
		"--id", "spec-sync-001",
		"--title", "Sync Specification",
		"--category", "technical",
		"--type", "specification",
		"--owner-group", "platform",
		"--status", "active",
		"--content", "Specification to verify sync scanning.",
	)
	createCmd.Dir = tempWorkspace
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("Create failed: %v\nOutput: %s", err, string(out))
	}

	// Run Sync
	syncCmd := exec.Command(gyrusBinPath, "sync")
	syncCmd.Dir = tempWorkspace
	var syncOut bytes.Buffer
	syncCmd.Stdout = &syncOut
	if err := syncCmd.Run(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if !strings.Contains(syncOut.String(), "Sync Completed") && !strings.Contains(syncOut.String(), "indexed") {
		t.Errorf("Expected sync summary output, got:\n%s", syncOut.String())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus validate' validates OKF frontmatter schemas and catches formatting errors.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Valid markdown passes validation; invalid markdown fails with Exit Code 1.
// -----------------------------------------------------------------------------
func TestCLI_Validate(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Init
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	// 1. Valid OKF File
	validDocContent := `---
id: adr-valid-001
title: Valid ADR Title
category: architecture
type: adr
owner_group: platform
version: 1
status: proposed
---

# Content
Valid body content.
`
	validPath := filepath.Join(tempWorkspace, "valid.md")
	if err := os.WriteFile(validPath, []byte(validDocContent), 0644); err != nil {
		t.Fatalf("Failed writing valid doc: %v", err)
	}

	valCmd := exec.Command(gyrusBinPath, "validate", validPath)
	valCmd.Dir = tempWorkspace
	var valOut bytes.Buffer
	valCmd.Stdout = &valOut
	if err := valCmd.Run(); err != nil {
		t.Fatalf("Validation of valid doc failed: %v", err)
	}

	if !strings.Contains(valOut.String(), "Validation successful") {
		t.Errorf("Expected success confirmation, got:\n%s", valOut.String())
	}

	// 2. Invalid OKF File (Missing Title & Invalid Category)
	invalidDocContent := `---
id: invalid-doc-001
category: not-a-real-category
type: adr
owner_group: platform
version: 1
status: proposed
---

# Content
`
	invalidPath := filepath.Join(tempWorkspace, "invalid.md")
	if err := os.WriteFile(invalidPath, []byte(invalidDocContent), 0644); err != nil {
		t.Fatalf("Failed writing invalid doc: %v", err)
	}

	invalidCmd := exec.Command(gyrusBinPath, "validate", invalidPath)
	invalidCmd.Dir = tempWorkspace
	err := invalidCmd.Run()
	if err == nil {
		t.Fatalf("Expected validation error for invalid markdown, but command succeeded")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("Expected Exit Code 1 for schema validation failure, got %d", exitErr.ExitCode())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus schema <doc-type>' prints the OKF document template and frontmatter contract.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Schema outputs valid YAML frontmatter template containing required fields.
// -----------------------------------------------------------------------------
func TestCLI_Schema(t *testing.T) {
	tempWorkspace := t.TempDir()

	cmd := exec.Command(gyrusBinPath, "schema", "adr")
	cmd.Dir = tempWorkspace
	var schemaOut bytes.Buffer
	cmd.Stdout = &schemaOut

	if err := cmd.Run(); err != nil {
		t.Fatalf("Schema command failed: %v", err)
	}

	output := schemaOut.String()
	if !strings.Contains(output, "type: adr") || !strings.Contains(output, "status:") {
		t.Errorf("Expected schema output to contain ADR template frontmatter, got:\n%s", output)
	}
}
