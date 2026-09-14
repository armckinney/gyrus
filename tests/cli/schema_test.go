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
// [Purpose]: Verifies that 'gyrus schema' subcommands (get, set, list, delete) manage schemas in the persistence layer.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Schemas can be saved to .gyrus/schemas/, listed, retrieved, and deleted with CLI status messages and JSON output.
// -----------------------------------------------------------------------------
func TestCLI_SchemaLifecycle(t *testing.T) {
	tempWorkspace := t.TempDir()

	// 1. Init workspace
	initCmd := exec.Command(gyrusBinPath, "init", "config", "--profile", "local")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	// 2. gyrus schema get adr
	getCmd := exec.Command(gyrusBinPath, "schema", "get", "adr")
	getCmd.Dir = tempWorkspace
	var getOut bytes.Buffer
	getCmd.Stdout = &getOut
	if err := getCmd.Run(); err != nil {
		t.Fatalf("schema get adr failed: %v", err)
	}
	if !strings.Contains(getOut.String(), "Architecture Decision Record") {
		t.Errorf("Expected embedded ADR template, got:\n%s", getOut.String())
	}

	// 3. Backward compatibility: gyrus schema adr
	legacyCmd := exec.Command(gyrusBinPath, "schema", "adr")
	legacyCmd.Dir = tempWorkspace
	var legacyOut bytes.Buffer
	legacyCmd.Stdout = &legacyOut
	if err := legacyCmd.Run(); err != nil {
		t.Fatalf("legacy schema adr failed: %v", err)
	}
	if !strings.Contains(legacyOut.String(), "Architecture Decision Record") {
		t.Errorf("Expected embedded ADR template in legacy invocation, got:\n%s", legacyOut.String())
	}

	// 4. Create custom schema file in /tmp/
	schemaFilePath := filepath.Join(tempWorkspace, "test-runbook.md")
	customContent := `---
id: <unique-dashed-id>
title: <Clean Document Title>
category: operations
type: runbook
format: markdown
owner_group: <team-owning-document>
version: 1
status: active
tags: []
dependencies: []
---

# Runbook: [Service/Procedure Name]

## Purpose
[Describe operational purpose]
`
	if err := os.WriteFile(schemaFilePath, []byte(customContent), 0644); err != nil {
		t.Fatalf("Failed writing test schema file: %v", err)
	}

	// 5. gyrus schema set runbook --content-file <file>
	setCmd := exec.Command(gyrusBinPath, "schema", "set", "runbook", "--content-file", schemaFilePath)
	setCmd.Dir = tempWorkspace
	if out, err := setCmd.CombinedOutput(); err != nil {
		t.Fatalf("schema set failed: %v\nOutput: %s", err, string(out))
	}

	// Verify file was written to .gyrus/schemas/runbook.md
	persistedPath := filepath.Join(tempWorkspace, ".gyrus", "schemas", "runbook.md")
	if _, err := os.Stat(persistedPath); err != nil {
		t.Fatalf("Expected persisted schema file at %s: %v", persistedPath, err)
	}

	// 6. gyrus schema list
	listCmd := exec.Command(gyrusBinPath, "schema", "list")
	listCmd.Dir = tempWorkspace
	var listOut bytes.Buffer
	listCmd.Stdout = &listOut
	if err := listCmd.Run(); err != nil {
		t.Fatalf("schema list failed: %v", err)
	}
	if !strings.Contains(listOut.String(), "runbook") || !strings.Contains(listOut.String(), "persisted") {
		t.Errorf("Expected 'runbook' with 'persisted' in list, got:\n%s", listOut.String())
	}

	// 7. gyrus schema list --json
	listJSONCmd := exec.Command(gyrusBinPath, "schema", "list", "--json")
	listJSONCmd.Dir = tempWorkspace
	var listJSONOut bytes.Buffer
	listJSONCmd.Stdout = &listJSONOut
	if err := listJSONCmd.Run(); err != nil {
		t.Fatalf("schema list --json failed: %v", err)
	}
	if !strings.Contains(listJSONOut.String(), `"type": "runbook"`) || !strings.Contains(listJSONOut.String(), `"source": "persisted"`) {
		t.Errorf("Expected JSON array with persisted runbook, got:\n%s", listJSONOut.String())
	}

	// 8. gyrus schema get runbook
	getRunbookCmd := exec.Command(gyrusBinPath, "schema", "get", "runbook")
	getRunbookCmd.Dir = tempWorkspace
	var getRunbookOut bytes.Buffer
	getRunbookCmd.Stdout = &getRunbookOut
	if err := getRunbookCmd.Run(); err != nil {
		t.Fatalf("schema get runbook failed: %v", err)
	}
	if !strings.Contains(getRunbookOut.String(), "Runbook: [Service/Procedure Name]") {
		t.Errorf("Expected custom runbook content, got:\n%s", getRunbookOut.String())
	}

	// 9. gyrus schema delete runbook
	delCmd := exec.Command(gyrusBinPath, "schema", "delete", "runbook")
	delCmd.Dir = tempWorkspace
	if out, err := delCmd.CombinedOutput(); err != nil {
		t.Fatalf("schema delete failed: %v\nOutput: %s", err, string(out))
	}

	// Verify file is gone from .gyrus/schemas/runbook.md
	if _, err := os.Stat(persistedPath); !os.IsNotExist(err) {
		t.Errorf("Expected schema file to be removed from %s", persistedPath)
	}

	// 10. List schemas after deletion
	listAfterCmd := exec.Command(gyrusBinPath, "schema", "list")
	listAfterCmd.Dir = tempWorkspace
	var listAfterOut bytes.Buffer
	listAfterCmd.Stdout = &listAfterOut
	if err := listAfterCmd.Run(); err != nil {
		t.Fatalf("schema list after delete failed: %v", err)
	}
	if strings.Contains(listAfterOut.String(), "persisted") {
		t.Errorf("Expected no persisted schemas after delete, got:\n%s", listAfterOut.String())
	}

	// 11. gyrus schema get runbook (non-existent, must fail with Exit Code 5)
	getDeletedCmd := exec.Command(gyrusBinPath, "schema", "get", "runbook")
	getDeletedCmd.Dir = tempWorkspace
	var getDeletedErr bytes.Buffer
	getDeletedCmd.Stderr = &getDeletedErr
	err := getDeletedCmd.Run()
	if err == nil {
		t.Fatalf("Expected CLI failure when getting non-existent schema 'runbook', but succeeded")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 5 {
		t.Errorf("Expected Exit Code 5 for schema not found, got %d. Stderr:\n%s", exitErr.ExitCode(), getDeletedErr.String())
	}
}
