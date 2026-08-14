package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus init' sets up a new local workspace structure with configuration.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Process exits with code 0; .gyrus configuration directory and settings file are created.
// -----------------------------------------------------------------------------
func TestCLI_Init(t *testing.T) {
	tempWorkspace := t.TempDir()

	cmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	cmd.Dir = tempWorkspace
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Init failed: %v\nStderr: %s", err, stderr.String())
	}

	cfgFile := filepath.Join(tempWorkspace, ".gyrus.yaml")
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		t.Errorf("Expected .gyrus.yaml configuration file to exist at %s", cfgFile)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies document creation and retrieval across human-readable text and pure JSON outputs.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: 'create' returns new document ID, 'get' outputs valid document data, and '--json' outputs parseable JSON envelope.
// -----------------------------------------------------------------------------
func TestCLI_CreateAndGet(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Init
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	// 1. Create with text output
	createCmd := exec.Command(gyrusBinPath, "create",
		"--id", "spec-crud-001",
		"--title", "CRUD Specification",
		"--category", "technical",
		"--type", "specification",
		"--owner-group", "platform",
		"--status", "active",
		"--tags", "crud,test,e2e",
		"--content", "Specification body content for CRUD tests.",
	)
	createCmd.Dir = tempWorkspace
	var createOut bytes.Buffer
	createCmd.Stdout = &createOut
	if err := createCmd.Run(); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !strings.Contains(createOut.String(), "spec-crud-001") {
		t.Errorf("Expected output to mention 'spec-crud-001', got:\n%s", createOut.String())
	}

	// 2. Get with pure JSON output
	getCmd := exec.Command(gyrusBinPath, "get", "spec-crud-001", "--json")
	getCmd.Dir = tempWorkspace
	var getJSONOut bytes.Buffer
	getCmd.Stdout = &getJSONOut
	if err := getCmd.Run(); err != nil {
		t.Fatalf("Get with --json failed: %v", err)
	}

	var doc gyrus.Document
	if err := json.Unmarshal(getJSONOut.Bytes(), &doc); err != nil {
		t.Fatalf("Failed unmarshaling --json output: %v\nRaw Output:\n%s", err, getJSONOut.String())
	}

	if doc.ID != "spec-crud-001" {
		t.Errorf("Expected ID 'spec-crud-001', got '%s'", doc.ID)
	}
	if doc.Title != "CRUD Specification" {
		t.Errorf("Expected Title 'CRUD Specification', got '%s'", doc.Title)
	}
	if len(doc.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %v", doc.Tags)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus update' modifies document fields and increments version number.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Updated title and content are persisted, version increments from 1 to 2.
// -----------------------------------------------------------------------------
func TestCLI_Update(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Init
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	// Create
	createCmd := exec.Command(gyrusBinPath, "create",
		"--id", "guide-crud-001",
		"--title", "Initial Guide",
		"--category", "technical",
		"--type", "guide",
		"--owner-group", "docs-team",
		"--status", "draft",
		"--content", "Initial draft content.",
	)
	createCmd.Dir = tempWorkspace
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("Create failed: %v\nOutput: %s", err, string(out))
	}

	// Update
	updateCmd := exec.Command(gyrusBinPath, "update", "guide-crud-001",
		"--title", "Updated Guide Title",
		"--status", "active",
		"--content", "Updated guide body content.",
		"--expected-version", "1",
	)
	updateCmd.Dir = tempWorkspace
	if out, err := updateCmd.CombinedOutput(); err != nil {
		t.Fatalf("Update failed: %v\nOutput: %s", err, string(out))
	}

	// Verify via Get --json
	getCmd := exec.Command(gyrusBinPath, "get", "guide-crud-001", "--json")
	getCmd.Dir = tempWorkspace
	var getOut bytes.Buffer
	getCmd.Stdout = &getOut
	if err := getCmd.Run(); err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	var doc gyrus.Document
	if err := json.Unmarshal(getOut.Bytes(), &doc); err != nil {
		t.Fatalf("Failed unmarshaling get output: %v", err)
	}

	if doc.Title != "Updated Guide Title" {
		t.Errorf("Expected updated title, got '%s'", doc.Title)
	}
	if doc.Status != "active" {
		t.Errorf("Expected updated status 'active', got '%s'", doc.Status)
	}
	if doc.Version != 2 {
		t.Errorf("Expected version incremented to 2, got %d", doc.Version)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus archive' transitions document status to 'archived'.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Process exits with code 0; document status transitions to 'archived'.
// -----------------------------------------------------------------------------
func TestCLI_Archive(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Init
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	// Create
	createCmd := exec.Command(gyrusBinPath, "create",
		"--id", "spec-archive-001",
		"--title", "Archivable Specification",
		"--category", "technical",
		"--type", "specification",
		"--owner-group", "platform",
		"--status", "active",
		"--content", "Specification to be archived.",
	)
	createCmd.Dir = tempWorkspace
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("Create failed: %v\nOutput: %s", err, string(out))
	}

	// Archive
	archiveCmd := exec.Command(gyrusBinPath, "archive", "spec-archive-001", "--json")
	archiveCmd.Dir = tempWorkspace
	var archiveOut bytes.Buffer
	archiveCmd.Stdout = &archiveOut
	if err := archiveCmd.Run(); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	if !strings.Contains(archiveOut.String(), `"archived": true`) {
		t.Errorf("Expected archive confirmation in JSON output, got:\n%s", archiveOut.String())
	}

	// Verify that getting the archived document returns Exit Code 5 (document not found)
	getCmd := exec.Command(gyrusBinPath, "get", "spec-archive-001")
	getCmd.Dir = tempWorkspace
	err := getCmd.Run()
	if err == nil {
		t.Fatalf("Expected get to fail after archive, but succeeded")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 5 {
		t.Errorf("Expected Exit Code 5 after archiving document, got %d", exitErr.ExitCode())
	}
}
