package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/setup"
)

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus init config' generates profile-specific .gyrus.yaml configurations and sets default owner group.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Configuration file contains specified profile settings and owner group.
// -----------------------------------------------------------------------------
func TestCLI_Init_ProfileFlags(t *testing.T) {
	tempWorkspace := t.TempDir()

	cmd := exec.Command(gyrusBinPath, "init", "config",
		"--profile", "postgres",
		"--owner-group", "platform-infra",
	)
	cmd.Dir = tempWorkspace
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("Init config failed: %v", err)
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
// [Purpose]: Verifies that 'gyrus init client' equips the Agent Plugin and registers MCP server configurations.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Agent Plugin files are written to .agents/plugins/gyrus/ and MCP server config is written to .antigravity/mcp.json.
// -----------------------------------------------------------------------------
func TestCLI_Init_MCPAndPluginEquipping(t *testing.T) {
	tempWorkspace := t.TempDir()

	cmd := exec.Command(gyrusBinPath, "init", "client",
		"--target", "antigravity",
		"--mode", "local",
	)
	cmd.Dir = tempWorkspace
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Init client failed: %v\nOutput: %s", err, string(out))
	}

	// 1. Verify Plugin installation
	pluginDir := filepath.Join(tempWorkspace, ".agents", "plugins", "gyrus")
	pluginJSON := filepath.Join(pluginDir, "plugin.json")
	if _, err := os.Stat(pluginJSON); os.IsNotExist(err) {
		t.Errorf("Expected plugin.json at %s", pluginJSON)
	}

	mcpJSON := filepath.Join(pluginDir, "mcp.json")
	if _, err := os.Stat(mcpJSON); os.IsNotExist(err) {
		t.Errorf("Expected mcp.json at %s", mcpJSON)
	}

	rulesFile := filepath.Join(pluginDir, "rules", "AGENTS.md")
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		t.Errorf("Expected rules/AGENTS.md at %s", rulesFile)
	}

	gyrusSkill := filepath.Join(pluginDir, "skills", "gyrus", "SKILL.md")
	if _, err := os.Stat(gyrusSkill); os.IsNotExist(err) {
		t.Errorf("Expected Gyrus skill in plugin at %s", gyrusSkill)
	}

	// 2. Verify MCP registration
	mcpJSONPath := filepath.Join(tempWorkspace, ".antigravity", "mcp.json")
	mcpData, err := os.ReadFile(mcpJSONPath)
	if err != nil {
		t.Fatalf("Failed reading MCP config: %v", err)
	}

	var mcpCfg setup.MCPConfigFile
	if err := json.Unmarshal(mcpData, &mcpCfg); err != nil {
		t.Fatalf("Failed unmarshaling MCP JSON: %v", err)
	}

	if _, ok := mcpCfg.MCPServers["gyrus"]; !ok {
		t.Errorf("Expected 'gyrus' server in mcpServers")
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that bare 'gyrus init' without subcommands is rejected with guidance.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Command exits with error indicating explicit subcommand is required.
// -----------------------------------------------------------------------------
func TestCLI_Init_BareCommandRejection(t *testing.T) {
	tempWorkspace := t.TempDir()

	cmd := exec.Command(gyrusBinPath, "init")
	cmd.Dir = tempWorkspace
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected bare 'gyrus init' to fail, but succeeded with output: %s", string(out))
	}

	outStr := string(out)
	if !strings.Contains(outStr, "initialization requires an explicit subcommand") {
		t.Errorf("Expected guidance message in output, got:\n%s", outStr)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus init client' without --target is rejected.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Command exits with validation error.
// -----------------------------------------------------------------------------
func TestCLI_Init_ClientMissingTarget(t *testing.T) {
	tempWorkspace := t.TempDir()

	cmd := exec.Command(gyrusBinPath, "init", "client")
	cmd.Dir = tempWorkspace
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected 'gyrus init client' without target to fail, but succeeded with output: %s", string(out))
	}

	outStr := string(out)
	if !strings.Contains(outStr, "--target is required") {
		t.Errorf("Expected missing target message, got:\n%s", outStr)
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

	cmd := exec.Command(gyrusBinPath, "init", "config",
		"--global",
		"--profile", "local",
		"--owner-group", "global-platform",
	)
	cmd.Env = append(os.Environ(), "HOME="+tempHome)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init config --global failed: %v\nOutput: %s", err, string(out))
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
// [Purpose]: Verifies that 'gyrus init config --global' refuses to overwrite without --force.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Exits with error mentioning --force when file already exists.
// -----------------------------------------------------------------------------
func TestCLI_InitConfigGlobalRefuse(t *testing.T) {
	tempHome := t.TempDir()

	// Pre-create ~/.gyrus.yaml
	if err := os.WriteFile(filepath.Join(tempHome, ".gyrus.yaml"), []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed writing initial config: %v", err)
	}

	cmd := exec.Command(gyrusBinPath, "init", "config", "--global")
	cmd.Env = append(os.Environ(), "HOME="+tempHome)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected init config --global to fail when file exists, but succeeded: %s", string(out))
	}

	outStr := string(out)
	if !strings.Contains(outStr, "already exists") || !strings.Contains(outStr, "--force") {
		t.Errorf("Expected refusal message mentioning --force, got:\n%s", outStr)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus init config --global --force' overwrites existing configuration.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Overwrites ~/.gyrus.yaml successfully.
// -----------------------------------------------------------------------------
func TestCLI_InitConfigGlobalForce(t *testing.T) {
	tempHome := t.TempDir()

	// Pre-create ~/.gyrus.yaml
	if err := os.WriteFile(filepath.Join(tempHome, ".gyrus.yaml"), []byte("old_config"), 0644); err != nil {
		t.Fatalf("Failed writing initial config: %v", err)
	}

	cmd := exec.Command(gyrusBinPath, "init", "config", "--global", "--force", "--profile", "postgres")
	cmd.Env = append(os.Environ(), "HOME="+tempHome)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init config --global --force failed: %v\nOutput: %s", err, string(out))
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
// [Purpose]: Verifies that re-running 'gyrus init client' repairs missing or deleted plugin files idempotently.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Deleted plugin file is restored on subsequent init client invocation.
// -----------------------------------------------------------------------------
func TestCLI_SetupIdempotencyAndRepair(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Initial setup
	initCmd := exec.Command(gyrusBinPath, "init", "client", "--target", "antigravity", "--mode", "local")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Initial init client failed: %v\nOutput: %s", err, string(out))
	}

	pluginFile := filepath.Join(tempWorkspace, ".agents", "plugins", "gyrus", "plugin.json")
	if _, err := os.Stat(pluginFile); os.IsNotExist(err) {
		t.Fatalf("Expected plugin.json to exist before deletion")
	}

	// Delete plugin file to simulate corruption/loss
	if err := os.Remove(pluginFile); err != nil {
		t.Fatalf("Failed deleting plugin file: %v", err)
	}

	// Re-run init client (repair)
	repairCmd := exec.Command(gyrusBinPath, "init", "client", "--target", "antigravity", "--mode", "local")
	repairCmd.Dir = tempWorkspace
	if out, err := repairCmd.CombinedOutput(); err != nil {
		t.Fatalf("Repair init client failed: %v\nOutput: %s", err, string(out))
	}

	// Verify plugin file is restored
	if _, err := os.Stat(pluginFile); os.IsNotExist(err) {
		t.Errorf("Expected deleted plugin file to be restored after re-running init client")
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

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus client install' and 'gyrus client uninstall' work symmetrically.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: 'client install' creates plugin and mcp.json; 'client uninstall' removes them.
// -----------------------------------------------------------------------------
func TestCLI_ClientInstallAndUninstall(t *testing.T) {
	tempWorkspace := t.TempDir()

	// 1. gyrus client install --target antigravity
	installCmd := exec.Command(gyrusBinPath, "client", "install", "--target", "antigravity")
	installCmd.Dir = tempWorkspace
	if out, err := installCmd.CombinedOutput(); err != nil {
		t.Fatalf("client install failed: %v\nOutput: %s", err, string(out))
	}

	pluginDir := filepath.Join(tempWorkspace, ".agents", "plugins", "gyrus")
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		t.Fatalf("Expected plugin directory to exist after install")
	}

	mcpFile := filepath.Join(tempWorkspace, ".antigravity", "mcp.json")
	mcpData, err := os.ReadFile(mcpFile)
	if err != nil {
		t.Fatalf("Expected .antigravity/mcp.json to exist: %v", err)
	}
	if !strings.Contains(string(mcpData), "GYRUS_WORKSPACE") {
		t.Errorf("Expected GYRUS_WORKSPACE env in mcp.json, got:\n%s", string(mcpData))
	}

	// 2. gyrus client uninstall --target antigravity
	uninstallCmd := exec.Command(gyrusBinPath, "client", "uninstall", "--target", "antigravity")
	uninstallCmd.Dir = tempWorkspace
	if out, err := uninstallCmd.CombinedOutput(); err != nil {
		t.Fatalf("client uninstall failed: %v\nOutput: %s", err, string(out))
	}

	// Plugin directory should be deleted
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Errorf("Expected plugin directory %s to be deleted after uninstall", pluginDir)
	}

	// MCP config should no longer have gyrus
	cleanedData, err := os.ReadFile(mcpFile)
	if err != nil {
		t.Fatalf("Expected mcp.json to still exist: %v", err)
	}
	if strings.Contains(string(cleanedData), `"gyrus":`) {
		t.Errorf("Expected gyrus to be removed from mcp.json, got:\n%s", string(cleanedData))
	}
}
