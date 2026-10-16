package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus init config' generates profile-specific .gyrus.yaml configurations and sets default owner group.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Configuration file contains specified profile settings and owner group.
// -----------------------------------------------------------------------------
func TestCLI_Config_ProfileFlags(t *testing.T) {
	tempWorkspace := t.TempDir()

	cmd := exec.Command(gyrusBinPath, "config", "init",
		"--profile", "postgres",
		"--owner-group", "platform-infra",
	)
	cmd.Dir = tempWorkspace
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("Config init failed: %v", err)
	}

	cfgPath := filepath.Join(tempWorkspace, ".gyrus.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("Failed reading config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "provider: postgres") {
		t.Errorf("Expected postgres storage provider in config, got:\n%s", content)
	}
	if !strings.Contains(content, "default_owner_group: platform-infra") {
		t.Errorf("Expected owner group 'platform-infra', got:\n%s", content)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that deprecated 'gyrus client' command is permanently removed.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Command exits with unknown command error.
// -----------------------------------------------------------------------------
func TestCLI_Client_CommandRejection(t *testing.T) {
	tempWorkspace := t.TempDir()

	for _, args := range [][]string{
		{"client"},
		{"client", "install"},
		{"client", "uninstall"},
	} {
		cmd := exec.Command(gyrusBinPath, args...)
		cmd.Dir = tempWorkspace
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("Expected 'gyrus %s' to fail, but succeeded with output: %s", strings.Join(args, " "), string(out))
		}
		if !strings.Contains(string(out), "unknown command \"client\"") {
			t.Errorf("Expected unknown command error for 'gyrus %s', got:\n%s", strings.Join(args, " "), string(out))
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that deprecated 'gyrus init' command is permanently removed.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Command exits with unknown command error.
// -----------------------------------------------------------------------------
func TestCLI_Init_BareCommandRejection(t *testing.T) {
	tempWorkspace := t.TempDir()

	cmd := exec.Command(gyrusBinPath, "init")
	cmd.Dir = tempWorkspace
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected 'gyrus init' to fail, but succeeded with output: %s", string(out))
	}

	outStr := string(out)
	if !strings.Contains(outStr, "unknown command \"init\"") {
		t.Errorf("Expected unknown command error in output, got:\n%s", outStr)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies configuration precedence: Workspace .gyrus.yaml > Global ~/.gyrus.yaml > Defaults.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Workspace .gyrus.yaml completely overrides global ~/.gyrus.yaml.
// -----------------------------------------------------------------------------
func TestCLI_ConfigPrecedence(t *testing.T) {
	tempWorkspace := t.TempDir()
	tempHome := t.TempDir()

	globalDocsDir := filepath.Join(tempHome, "global_docs")
	customDocsDir := filepath.Join(tempWorkspace, "custom_docs")

	// 1. Write global ~/.gyrus.yaml
	globalContent := `storage:
  provider: localfs
  root: global_docs
default_owner_group: global-team
`
	if err := os.WriteFile(filepath.Join(tempHome, ".gyrus.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("Failed writing global config: %v", err)
	}

	// 2. Write workspace .gyrus.yaml
	wsContent := `storage:
  provider: localfs
  root: custom_docs
default_owner_group: workspace-team
`
	if err := os.WriteFile(filepath.Join(tempWorkspace, ".gyrus.yaml"), []byte(wsContent), 0644); err != nil {
		t.Fatalf("Failed writing workspace config: %v", err)
	}

	// 3. Create document in workspace (should use workspace custom_docs, overriding global)
	cmd := exec.Command(gyrusBinPath, "create",
		"--id", "doc-precedence-001",
		"--title", "Precedence Doc",
		"--category", "technical",
		"--type", "specification",
		"--owner-group", "workspace-team",
		"--status", "active",
		"--content", "Stored in custom_docs.",
	)
	cmd.Dir = tempWorkspace
	cmd.Env = append(os.Environ(), "HOME="+tempHome)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Create with precedence failed: %v\nOutput: %s", err, string(out))
	}

	foundInCustom := false
	_ = filepath.Walk(customDocsDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.Contains(path, "doc-precedence-001") {
			foundInCustom = true
		}
		return nil
	})
	if !foundInCustom {
		t.Errorf("Expected document to be stored in workspace %s", customDocsDir)
	}

	foundInGlobal := false
	_ = filepath.Walk(globalDocsDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.Contains(path, "doc-precedence-001") {
			foundInGlobal = true
		}
		return nil
	})
	if foundInGlobal {
		t.Errorf("Document should not be stored in global directory %s", globalDocsDir)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus config show' displays resolved config and source.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Displays source header and resolved storage root.
// -----------------------------------------------------------------------------
func TestCLI_ConfigShow(t *testing.T) {
	tempWorkspace := t.TempDir()

	wsContent := `storage:
  provider: localfs
  root: my_workspace_docs
default_owner_group: test-group
`
	if err := os.WriteFile(filepath.Join(tempWorkspace, ".gyrus.yaml"), []byte(wsContent), 0644); err != nil {
		t.Fatalf("Failed writing workspace config: %v", err)
	}

	cmd := exec.Command(gyrusBinPath, "config", "show")
	cmd.Dir = tempWorkspace
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("config show failed: %v\nOutput: %s", err, string(out))
	}

	outStr := string(out)
	if !strings.Contains(outStr, "# Source: workspace") {
		t.Errorf("Expected output to contain '# Source: workspace', got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "# Resolved Storage Root:") {
		t.Errorf("Expected output to contain '# Resolved Storage Root:', got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "my_workspace_docs") {
		t.Errorf("Expected output to contain 'my_workspace_docs', got:\n%s", outStr)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus config show --json' outputs valid JSON with source metadata.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Valid JSON containing source, storage_root, and config fields.
// -----------------------------------------------------------------------------
func TestCLI_ConfigShowJSON(t *testing.T) {
	tempWorkspace := t.TempDir()

	wsContent := `storage:
  provider: localfs
  root: json_docs
default_owner_group: json-group
`
	if err := os.WriteFile(filepath.Join(tempWorkspace, ".gyrus.yaml"), []byte(wsContent), 0644); err != nil {
		t.Fatalf("Failed writing workspace config: %v", err)
	}

	cmd := exec.Command(gyrusBinPath, "config", "show", "--json")
	cmd.Dir = tempWorkspace
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("config show --json failed: %v\nOutput: %s", err, string(out))
	}

	var payload struct {
		Source      string                 `json:"source"`
		SourcePath  string                 `json:"source_path"`
		StorageRoot string                 `json:"storage_root"`
		Config      map[string]interface{} `json:"config"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatalf("Failed unmarshaling config show JSON: %v\nOutput: %s", err, string(out))
	}

	if payload.Source != "workspace" {
		t.Errorf("Expected source 'workspace', got '%s'", payload.Source)
	}
	if !strings.Contains(payload.StorageRoot, "json_docs") {
		t.Errorf("Expected storage_root to contain 'json_docs', got '%s'", payload.StorageRoot)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus init config --global' writes ~/.gyrus.yaml in user home.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: ~/.gyrus.yaml created with specified profile and owner group.
// -----------------------------------------------------------------------------
func TestCLI_InitConfigGlobal(t *testing.T) {
	tempHome := t.TempDir()

	cmd := exec.Command(gyrusBinPath, "config", "init",
		"--global",
		"--profile", "local",
		"--owner-group", "global-platform",
	)
	cmd.Env = append(os.Environ(), "HOME="+tempHome)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("config init --global failed: %v\nOutput: %s", err, string(out))
	}

	globalPath := filepath.Join(tempHome, ".gyrus.yaml")
	data, err := os.ReadFile(globalPath)
	if err != nil {
		t.Fatalf("Expected ~/.gyrus.yaml to exist: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "default_owner_group: global-platform") {
		t.Errorf("Expected owner group 'global-platform', got:\n%s", content)
	}
	if !strings.Contains(content, "provider: localfs") {
		t.Errorf("Expected 'provider: localfs', got:\n%s", content)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus config init --global' refuses to overwrite without --force.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Exits with error mentioning --force when file already exists.
// -----------------------------------------------------------------------------
func TestCLI_InitConfigGlobalRefuse(t *testing.T) {
	tempHome := t.TempDir()

	// Pre-create ~/.gyrus.yaml
	if err := os.WriteFile(filepath.Join(tempHome, ".gyrus.yaml"), []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed writing initial config: %v", err)
	}

	cmd := exec.Command(gyrusBinPath, "config", "init", "--global")
	cmd.Env = append(os.Environ(), "HOME="+tempHome)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected config init --global to fail when file exists, but succeeded: %s", string(out))
	}

	outStr := string(out)
	if !strings.Contains(outStr, "already exists") || !strings.Contains(outStr, "--force") {
		t.Errorf("Expected refusal message mentioning --force, got:\n%s", outStr)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus config init --global --force' overwrites existing configuration.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Overwrites ~/.gyrus.yaml successfully.
// -----------------------------------------------------------------------------
func TestCLI_InitConfigGlobalForce(t *testing.T) {
	tempHome := t.TempDir()

	// Pre-create ~/.gyrus.yaml
	if err := os.WriteFile(filepath.Join(tempHome, ".gyrus.yaml"), []byte("old_config"), 0644); err != nil {
		t.Fatalf("Failed writing initial config: %v", err)
	}

	cmd := exec.Command(gyrusBinPath, "config", "init", "--global", "--force", "--profile", "postgres")
	cmd.Env = append(os.Environ(), "HOME="+tempHome)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("config init --global --force failed: %v\nOutput: %s", err, string(out))
	}

	data, err := os.ReadFile(filepath.Join(tempHome, ".gyrus.yaml"))
	if err != nil {
		t.Fatalf("Failed reading ~/.gyrus.yaml: %v", err)
	}
	if !strings.Contains(string(data), "provider: postgres") {
		t.Errorf("Expected config to contain 'provider: postgres', got:\n%s", string(data))
	}
}


// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus config init' generates .gyrus.yaml configuration.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Configuration file contains specified profile settings and owner group.
// -----------------------------------------------------------------------------
func TestCLI_ConfigInit(t *testing.T) {
	tempWorkspace := t.TempDir()

	cmd := exec.Command(gyrusBinPath, "config", "init",
		"--profile", "local",
		"--owner-group", "dev-team",
	)
	cmd.Dir = tempWorkspace
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	cfgPath := filepath.Join(tempWorkspace, ".gyrus.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("Failed reading config: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "provider: localfs") {
		t.Errorf("Expected localfs storage provider, got:\n%s", content)
	}
	if !strings.Contains(content, "default_owner_group: dev-team") {
		t.Errorf("Expected owner group 'dev-team', got:\n%s", content)
	}
}
