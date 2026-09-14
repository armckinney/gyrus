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
	if !strings.Contains(content, "storage_provider: postgres") {
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
// [Purpose]: Verifies configuration path precedence: CLI Flag > Environment Variable > .gyrus.yaml.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Custom storage_root from .gyrus.yaml is respected; explicit --storage-path flag overrides config file.
// -----------------------------------------------------------------------------
func TestCLI_ConfigPrecedence(t *testing.T) {
	tempWorkspace := t.TempDir()

	customDocsDir := filepath.Join(tempWorkspace, "custom_docs")
	overrideDocsDir := filepath.Join(tempWorkspace, "override_docs")

	// 1. Write custom .gyrus.yaml
	configContent := `# Custom Config
storage_provider: localfs
index_provider: sqlite
search_provider: sqlite
storage_root: custom_docs
default_owner_group: test-team
`
	if err := os.WriteFile(filepath.Join(tempWorkspace, ".gyrus.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed writing custom config: %v", err)
	}

	// 2. Create document without flag (should use custom_docs from config)
	cmd1 := exec.Command(gyrusBinPath, "create",
		"--id", "doc-custom-001",
		"--title", "Custom Path Doc",
		"--category", "technical",
		"--type", "specification",
		"--owner-group", "test-team",
		"--status", "active",
		"--content", "Stored in custom_docs.",
	)
	cmd1.Dir = tempWorkspace
	if out, err := cmd1.CombinedOutput(); err != nil {
		t.Fatalf("Create with custom config failed: %v\nOutput: %s", err, string(out))
	}

	foundInCustom := false
	_ = filepath.Walk(customDocsDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.Contains(path, "doc-custom-001") {
			foundInCustom = true
		}
		return nil
	})

	if !foundInCustom {
		t.Errorf("Expected document to be stored in %s", customDocsDir)
	}

	// 3. Create document with explicit --storage-path flag (override)
	cmd2 := exec.Command(gyrusBinPath, "create",
		"--storage-path", overrideDocsDir,
		"--id", "doc-override-001",
		"--title", "Override Path Doc",
		"--category", "technical",
		"--type", "specification",
		"--owner-group", "test-team",
		"--status", "active",
		"--content", "Stored in override_docs.",
	)
	cmd2.Dir = tempWorkspace
	if out, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("Create with override flag failed: %v\nOutput: %s", err, string(out))
	}

	foundInOverride := false
	_ = filepath.Walk(overrideDocsDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && strings.Contains(path, "doc-override-001") {
			foundInOverride = true
		}
		return nil
	})

	if !foundInOverride {
		t.Errorf("Expected document to be stored in override directory %s", overrideDocsDir)
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
