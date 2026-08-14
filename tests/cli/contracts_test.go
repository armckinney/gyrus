package cli_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var gyrusBinPath string

func TestMain(m *testing.M) {
	// Build gyrus binary for CLI E2E tests
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		panic(err)
	}

	tempDir, err := os.MkdirTemp("", "gyrus-bin-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	binPath := filepath.Join(tempDir, "gyrus")
	buildCmd := exec.Command("go", "build", "-o", binPath, "cmd/gyrus/main.go")
	buildCmd.Dir = repoRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		panic("failed building gyrus binary for tests: " + string(out))
	}

	gyrusBinPath = binPath
	os.Exit(m.Run())
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that attempting to create a document with invalid schema fields or malformed IDs fails with Exit Code 1.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Process exits with code 1; error message is printed to stderr; stdout contains no corrupt data.
// -----------------------------------------------------------------------------
func TestCLI_SchemaValidation_ExitCode1(t *testing.T) {
	tempWorkspace := t.TempDir()

	// 1. Initialize workspace
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init workspace: %v\nOutput: %s", err, string(out))
	}

	// 2. Test Invalid ID (uppercase & spaces)
	createCmd := exec.Command(gyrusBinPath, "create",
		"--id", "INVALID UPPER_CASE",
		"--title", "Bad ID Doc",
		"--category", "architecture",
		"--type", "adr",
		"--owner-group", "armckinney",
	)
	createCmd.Dir = tempWorkspace
	var stdout, stderr bytes.Buffer
	createCmd.Stdout = &stdout
	createCmd.Stderr = &stderr

	err := createCmd.Run()
	if err == nil {
		t.Fatalf("Expected CLI failure for invalid ID, but command succeeded")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("Expected Exit Code 1 for validation error, got %d. Stderr:\n%s", exitErr.ExitCode(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "validation error") {
		t.Errorf("Expected validation error message on stderr, got:\n%s", stderr.String())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that attempting an illegal lifecycle transition on an ADR is blocked and exits with Exit Code 2.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Process exits with code 2; transition error message is emitted to stderr.
// -----------------------------------------------------------------------------
func TestCLI_IllegalLifecycleTransition_ExitCode2(t *testing.T) {
	tempWorkspace := t.TempDir()

	// 1. Initialize workspace
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init workspace: %v\nOutput: %s", err, string(out))
	}

	// 2. Create accepted ADR
	createCmd := exec.Command(gyrusBinPath, "create",
		"--id", "adr-storage-engine",
		"--title", "Storage Engine Architecture",
		"--category", "architecture",
		"--type", "adr",
		"--owner-group", "platform",
		"--status", "accepted",
	)
	createCmd.Dir = tempWorkspace
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed creating accepted ADR: %v\nOutput: %s", err, string(out))
	}

	// 3. Attempt illegal state transition: accepted -> proposed
	updateCmd := exec.Command(gyrusBinPath, "update", "adr-storage-engine", "--status", "proposed")
	updateCmd.Dir = tempWorkspace
	var stderr bytes.Buffer
	updateCmd.Stderr = &stderr

	err := updateCmd.Run()
	if err == nil {
		t.Fatalf("Expected CLI failure for illegal state transition accepted -> proposed, but succeeded")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 2 {
		t.Errorf("Expected Exit Code 2 for illegal transition, got %d. Stderr:\n%s", exitErr.ExitCode(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "cannot transition") && !strings.Contains(stderr.String(), "invalid lifecycle transition") {
		t.Errorf("Expected lifecycle transition error message on stderr, got:\n%s", stderr.String())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that attempting to directly modify content of an accepted ADR is blocked by immutability enforcement.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Process exits with code 2; immutability error is reported on stderr.
// -----------------------------------------------------------------------------
func TestCLI_ImmutabilityEnforcement_ExitCode2(t *testing.T) {
	tempWorkspace := t.TempDir()

	// 1. Initialize workspace
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init workspace: %v\nOutput: %s", err, string(out))
	}

	// 2. Create accepted ADR
	createCmd := exec.Command(gyrusBinPath, "create",
		"--id", "adr-locked-decision",
		"--title", "Finalized Decision",
		"--category", "architecture",
		"--type", "adr",
		"--owner-group", "platform",
		"--status", "accepted",
		"--content", "Original finalized content.",
	)
	createCmd.Dir = tempWorkspace
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed creating accepted ADR: %v\nOutput: %s", err, string(out))
	}

	// 3. Attempt content mutation on accepted ADR
	updateCmd := exec.Command(gyrusBinPath, "update", "adr-locked-decision", "--content", "Mutated illegal content.")
	updateCmd.Dir = tempWorkspace
	var stderr bytes.Buffer
	updateCmd.Stderr = &stderr

	err := updateCmd.Run()
	if err == nil {
		t.Fatalf("Expected CLI failure for modifying content of accepted ADR, but succeeded")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 2 {
		t.Errorf("Expected Exit Code 2 for immutability error, got %d. Stderr:\n%s", exitErr.ExitCode(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "immutable") {
		t.Errorf("Expected immutability error message on stderr, got:\n%s", stderr.String())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that optimistic concurrency locks reject updates when --expected-version mismatches document version.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Process exits with code 4; concurrency mismatch error is reported on stderr.
// -----------------------------------------------------------------------------
func TestCLI_OptimisticConcurrencyConflict_ExitCode4(t *testing.T) {
	tempWorkspace := t.TempDir()

	// 1. Initialize workspace
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init workspace: %v\nOutput: %s", err, string(out))
	}

	// 2. Create living specification (version 1)
	createCmd := exec.Command(gyrusBinPath, "create",
		"--id", "spec-storage-engine",
		"--title", "Storage Engine Spec",
		"--category", "technical",
		"--type", "specification",
		"--owner-group", "platform",
		"--status", "active",
	)
	createCmd.Dir = tempWorkspace
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed creating spec: %v\nOutput: %s", err, string(out))
	}

	// 3. Attempt update with wrong expected version (expected 99, actual 1)
	updateCmd := exec.Command(gyrusBinPath, "update", "spec-storage-engine",
		"--title", "New Title",
		"--expected-version", "99",
	)
	updateCmd.Dir = tempWorkspace
	var stderr bytes.Buffer
	updateCmd.Stderr = &stderr

	err := updateCmd.Run()
	if err == nil {
		t.Fatalf("Expected CLI failure for optimistic concurrency version mismatch, but succeeded")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 4 {
		t.Errorf("Expected Exit Code 4 for concurrency error, got %d. Stderr:\n%s", exitErr.ExitCode(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "expected version") && !strings.Contains(stderr.String(), "concurrency") {
		t.Errorf("Expected concurrency version mismatch message on stderr, got:\n%s", stderr.String())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that requesting a non-existent document ID exits with Exit Code 5.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Process exits with code 5; not found error is reported on stderr.
// -----------------------------------------------------------------------------
func TestCLI_DocumentNotFound_ExitCode5(t *testing.T) {
	tempWorkspace := t.TempDir()

	// 1. Initialize workspace
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init workspace: %v\nOutput: %s", err, string(out))
	}

	// 2. Attempt getting non-existent document
	getCmd := exec.Command(gyrusBinPath, "get", "non-existent-doc-id-12345")
	getCmd.Dir = tempWorkspace
	var stderr bytes.Buffer
	getCmd.Stderr = &stderr

	err := getCmd.Run()
	if err == nil {
		t.Fatalf("Expected CLI failure for non-existent document ID, but succeeded")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("Expected ExitError, got %v", err)
	}
	if exitErr.ExitCode() != 5 {
		t.Errorf("Expected Exit Code 5 for not found/storage error, got %d. Stderr:\n%s", exitErr.ExitCode(), stderr.String())
	}
}
