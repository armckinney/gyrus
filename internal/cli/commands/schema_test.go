package commands_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/cli/commands"
	"github.com/spf13/cobra"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies CLI schema subcommands: get, set, list, delete, and backward-compatible direct calls.
// [Execution Surface]: In-Memory App with LocalFS Store
// [Assertions]: Subcommands set/get/list/delete schemas to/from .gyrus/schemas/, supporting both text and JSON outputs.
// -----------------------------------------------------------------------------
func TestCLISchemaSubcommands(t *testing.T) {
	tempDir := t.TempDir()
	storageRoot := filepath.Join(tempDir, ".gyrus")
	_ = os.MkdirAll(storageRoot, 0755)

	application, err := app.New(storageRoot)
	if err != nil {
		t.Fatalf("app.New failed: %v", err)
	}

	execCmd := func(args ...string) (string, error) {
		rootCmd := &cobra.Command{Use: "gyrus"}
		commands.Register(rootCmd, application)

		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs(args)

		err := rootCmd.ExecuteContext(context.Background())
		return buf.String(), err
	}

	// 1. gyrus schema get adr (default embedded)
	out, err := execCmd("schema", "get", "adr")
	if err != nil {
		t.Fatalf("schema get adr failed: %v", err)
	}
	if !strings.Contains(out, "Architecture Decision Record") {
		t.Errorf("expected embedded ADR template in output, got: %s", out)
	}

	// 2. Backward compatibility: gyrus schema adr
	out, err = execCmd("schema", "adr")
	if err != nil {
		t.Fatalf("schema adr failed: %v", err)
	}
	if !strings.Contains(out, "Architecture Decision Record") {
		t.Errorf("expected embedded ADR template in backward-compat output, got: %s", out)
	}

	// 3. gyrus schema set runbook --content "..."
	schemaContent := `---
id: <unique-id>
title: <Title>
category: operations
type: runbook
owner_group: ops
version: 1
status: active
---

# Runbook Template Body
`
	_, err = execCmd("schema", "set", "runbook", "--content", schemaContent)
	if err != nil {
		t.Fatalf("schema set failed: %v", err)
	}

	// Verify file was written to .gyrus/schemas/runbook.md
	schemaFilePath := filepath.Join(storageRoot, "schemas", "runbook.md")
	if _, err := os.Stat(schemaFilePath); err != nil {
		t.Errorf("expected schema file at %s, got: %v", schemaFilePath, err)
	}

	// 4. gyrus schema list
	out, err = execCmd("schema", "list")
	if err != nil {
		t.Fatalf("schema list failed: %v", err)
	}
	if !strings.Contains(out, "runbook") || !strings.Contains(out, "persisted") {
		t.Errorf("expected 'runbook' with 'persisted' in schema list, got: %s", out)
	}

	// 5. gyrus schema get runbook
	out, err = execCmd("schema", "get", "runbook")
	if err != nil {
		t.Fatalf("schema get runbook failed: %v", err)
	}
	if !strings.Contains(out, "Runbook Template Body") {
		t.Errorf("expected custom runbook content, got: %s", out)
	}

	// 6. Test --content-file with schema set
	contentFilePath := filepath.Join(tempDir, "policy.md")
	policyContent := `---
id: <unique-id>
title: <Policy>
category: technical
type: policy
owner_group: compliance
version: 1
status: active
---

# Policy Template
`
	if err := os.WriteFile(contentFilePath, []byte(policyContent), 0644); err != nil {
		t.Fatalf("failed writing policy content file: %v", err)
	}

	_, err = execCmd("schema", "set", "policy", "--content-file", contentFilePath)
	if err != nil {
		t.Fatalf("schema set with --content-file failed: %v", err)
	}

	out, err = execCmd("schema", "get", "policy")
	if err != nil {
		t.Fatalf("schema get policy failed: %v", err)
	}
	if !strings.Contains(out, "Policy Template") {
		t.Errorf("expected policy template content, got: %s", out)
	}

	// 7. gyrus schema delete runbook
	_, err = execCmd("schema", "delete", "runbook")
	if err != nil {
		t.Fatalf("schema delete runbook failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(schemaFilePath); !os.IsNotExist(err) {
		t.Errorf("expected schema file to be deleted from %s", schemaFilePath)
	}

	// 8. gyrus schema get runbook (deleted/non-existent, should fail)
	_, err = execCmd("schema", "get", "runbook")
	if err == nil {
		t.Fatalf("expected error when getting deleted/non-existent schema, got nil")
	}
	if !strings.Contains(err.Error(), "schema template not found") {
		t.Errorf("expected 'schema template not found' in error, got: %v", err)
	}
}
