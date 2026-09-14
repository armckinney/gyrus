package setup_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/setup"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that RunConfigSetup creates .gyrus.yaml with designated profile and owner group.
// [Execution Surface]: In-Memory / Temporary Filesystem
// [Assertions]: Creates .gyrus.yaml in workspace and returns correct status fields.
// -----------------------------------------------------------------------------
func TestRunConfigSetup(t *testing.T) {
	tempDir := t.TempDir()

	res, err := setup.RunConfigSetup(setup.ConfigSetupOptions{
		WorkspaceDir: tempDir,
		Profile:      setup.ProfileLocal,
		OwnerGroup:   "test-group",
	})
	if err != nil {
		t.Fatalf("RunConfigSetup failed: %v", err)
	}

	configPath := filepath.Join(tempDir, ".gyrus.yaml")
	if res.ConfigFile != configPath {
		t.Errorf("Expected config file %s, got %s", configPath, res.ConfigFile)
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected .gyrus.yaml to exist at %s", configPath)
	}

	expectedStorageDir := filepath.Join(tempDir, ".gyrus")
	if res.StorageDir != expectedStorageDir {
		t.Errorf("Expected storage root path %s, got %s", expectedStorageDir, res.StorageDir)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that RunClientSetup installs the complete Agent Plugin package and registers MCP.
// [Execution Surface]: In-Memory / Temporary Filesystem
// [Assertions]: Creates plugin.json, mcp.json, mcp_config.json, rules/AGENTS.md, skills/, and .antigravity/mcp.json.
// -----------------------------------------------------------------------------
func TestRunClientSetupAntigravity(t *testing.T) {
	tempDir := t.TempDir()

	result, err := setup.RunClientSetup(setup.ClientSetupOptions{
		WorkspaceDir: tempDir,
		Target:       setup.ClientTargetAntigravity,
		Mode:         setup.MCPModeLocal,
		BinaryCmd:    "gyrus",
	})
	if err != nil {
		t.Fatalf("RunClientSetup failed: %v", err)
	}

	pluginDir := filepath.Join(tempDir, ".agents", "plugins", "gyrus")
	if result.PluginDir != pluginDir {
		t.Errorf("Expected plugin dir %s, got %s", pluginDir, result.PluginDir)
	}

	// 1. Verify plugin.json
	pluginJSONPath := filepath.Join(pluginDir, "plugin.json")
	pluginData, err := os.ReadFile(pluginJSONPath)
	if err != nil {
		t.Fatalf("Failed reading plugin.json: %v", err)
	}
	var manifest map[string]interface{}
	if err := json.Unmarshal(pluginData, &manifest); err != nil {
		t.Fatalf("plugin.json is invalid JSON: %v", err)
	}
	if manifest["name"] != "gyrus" {
		t.Errorf("Expected manifest name 'gyrus', got %v", manifest["name"])
	}
	if manifest["$schema"] != "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json" {
		t.Errorf("Expected standard schema, got %v", manifest["$schema"])
	}

	// 2. Verify mcp.json (Standard)
	mcpJSONPath := filepath.Join(pluginDir, "mcp.json")
	if _, err := os.Stat(mcpJSONPath); os.IsNotExist(err) {
		t.Errorf("Expected mcp.json at %s", mcpJSONPath)
	}

	// 3. Verify mcp_config.json (Antigravity)
	mcpConfigPath := filepath.Join(pluginDir, "mcp_config.json")
	if _, err := os.Stat(mcpConfigPath); os.IsNotExist(err) {
		t.Errorf("Expected mcp_config.json at %s", mcpConfigPath)
	}

	// 4. Verify rules/AGENTS.md
	rulesPath := filepath.Join(pluginDir, "rules", "AGENTS.md")
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		t.Errorf("Expected rules/AGENTS.md at %s", rulesPath)
	}

	// 5. Verify skills
	gyrusSkill := filepath.Join(pluginDir, "skills", "gyrus", "SKILL.md")
	if _, err := os.Stat(gyrusSkill); os.IsNotExist(err) {
		t.Errorf("Expected Gyrus skill at %s", gyrusSkill)
	}

	// 6. Verify client MCP registration
	antigravityMCP := filepath.Join(tempDir, ".antigravity", "mcp.json")
	if _, err := os.Stat(antigravityMCP); os.IsNotExist(err) {
		t.Errorf("Expected Antigravity MCP config at %s", antigravityMCP)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that explicit client targets are validated and invalid targets are rejected.
// [Execution Surface]: Pure Go Logic
// [Assertions]: 'all' or random strings return an error; valid targets succeed.
// -----------------------------------------------------------------------------
func TestClientTargetValidation(t *testing.T) {
	invalidTargets := []string{"all", "invalid", "", "other"}
	for _, it := range invalidTargets {
		_, err := setup.ValidateClientTarget(it)
		if err == nil {
			t.Errorf("Expected error for invalid target '%s', got nil", it)
		}
	}

	validTargets := []setup.ClientTarget{
		setup.ClientTargetAntigravity,
		setup.ClientTargetClaude,
		setup.ClientTargetCodex,
		setup.ClientTargetCopilot,
	}
	for _, vt := range validTargets {
		res, err := setup.ValidateClientTarget(string(vt))
		if err != nil {
			t.Errorf("Expected valid target for '%s', got error: %v", vt, err)
		}
		if res != vt {
			t.Errorf("Expected %s, got %s", vt, res)
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that WriteConfigFile generates valid YAML for all storage & search profiles.
// [Execution Surface]: In-Memory / Temporary Filesystem
// [Assertions]: Generated config contains expected provider keywords for local, git, blob, s3, azure, gcs, postgres, and vector profiles.
// -----------------------------------------------------------------------------
func TestWriteConfigFileProfiles(t *testing.T) {
	profiles := []struct {
		profile setup.Profile
		keyword string
	}{
		{setup.ProfileLocal, "storage_provider: localfs"},
		{setup.ProfileGit, "storage_provider: git"},
		{setup.ProfileBlob, "storage_provider: blob"},
		{setup.ProfileS3, "storage_provider: s3"},
		{setup.ProfileAzure, "storage_provider: azure_blob"},
		{setup.ProfileGCS, "storage_provider: gcs"},
		{setup.ProfilePostgres, "storage_provider: postgres"},
		{setup.ProfileVector, "search_provider: vector"},
	}

	for _, p := range profiles {
		t.Run(string(p.profile), func(t *testing.T) {
			tempDir := t.TempDir()
			cfgPath, err := setup.WriteConfigFile(tempDir, p.profile, "my-team")
			if err != nil {
				t.Fatalf("WriteConfigFile failed for %s: %v", p.profile, err)
			}

			data, err := os.ReadFile(cfgPath)
			if err != nil {
				t.Fatalf("Failed reading generated config: %v", err)
			}

			content := string(data)
			if !strings.Contains(content, p.keyword) {
				t.Errorf("Expected config to contain '%s', got:\n%s", p.keyword, content)
			}
			if !strings.Contains(content, "default_owner_group: my-team") {
				t.Errorf("Expected config to contain owner group 'my-team', got:\n%s", content)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies MCP server registration across container vs local binary execution modes.
// [Execution Surface]: In-Memory / Temporary Filesystem
// [Assertions]: Container mode registers 'docker' command with volume mounts; local mode registers 'gyrus mcp serve'.
// -----------------------------------------------------------------------------
func TestRegisterMCPServerModes(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Container Mode
	mcpFiles, err := setup.RegisterMCPServer(tempDir, setup.MCPTargetAntigravity, setup.MCPModeContainer, false, "gyrus", "ghcr.io/armckinney/gyrus:v1.0")
	if err != nil {
		t.Fatalf("RegisterMCPServer container mode failed: %v", err)
	}
	if len(mcpFiles) != 1 {
		t.Fatalf("Expected 1 MCP config file, got %d", len(mcpFiles))
	}

	data, err := os.ReadFile(mcpFiles[0])
	if err != nil {
		t.Fatalf("Failed reading MCP config: %v", err)
	}

	var cfg setup.MCPConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Failed unmarshaling MCP JSON: %v", err)
	}

	srv, ok := cfg.MCPServers["gyrus"]
	if !ok {
		t.Fatalf("Expected 'gyrus' server key in mcpServers")
	}
	if srv.Command != "docker" {
		t.Errorf("Expected command 'docker', got '%s'", srv.Command)
	}

	// 2. Local Mode
	mcpLocalFiles, err := setup.RegisterMCPServer(tempDir, setup.MCPTargetAntigravity, setup.MCPModeLocal, false, "/usr/local/bin/gyrus", "")
	if err != nil {
		t.Fatalf("RegisterMCPServer local mode failed: %v", err)
	}
	if len(mcpLocalFiles) != 1 {
		t.Fatalf("Expected 1 MCP config file, got %d", len(mcpLocalFiles))
	}

	localData, _ := os.ReadFile(mcpLocalFiles[0])
	var localCfg setup.MCPConfigFile
	_ = json.Unmarshal(localData, &localCfg)

	localSrv, ok := localCfg.MCPServers["gyrus"]
	if !ok {
		t.Fatalf("Expected 'gyrus' server key in config")
	}
	if localSrv.Command != "/usr/local/bin/gyrus" {
		t.Errorf("Expected command '/usr/local/bin/gyrus', got '%s'", localSrv.Command)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that running setup multiple times is idempotent and preserves user workspace state.
// [Execution Surface]: In-Memory / Temporary Filesystem
// [Assertions]: Repeated setup invocations succeed without errors and preserve existing configurations.
// -----------------------------------------------------------------------------
func TestSetupIdempotencyAndPreservation(t *testing.T) {
	tempDir := t.TempDir()

	// Config setup 1 & 2
	cfgOpts := setup.ConfigSetupOptions{
		WorkspaceDir: tempDir,
		Profile:      setup.ProfileLocal,
		OwnerGroup:   "core-eng",
	}
	if _, err := setup.RunConfigSetup(cfgOpts); err != nil {
		t.Fatalf("Config run 1 failed: %v", err)
	}
	if _, err := setup.RunConfigSetup(cfgOpts); err != nil {
		t.Fatalf("Config run 2 failed: %v", err)
	}

	// Client setup 1 & 2
	clientOpts := setup.ClientSetupOptions{
		WorkspaceDir: tempDir,
		Target:       setup.ClientTargetAntigravity,
		Mode:         setup.MCPModeLocal,
		BinaryCmd:    "gyrus",
	}
	if _, err := setup.RunClientSetup(clientOpts); err != nil {
		t.Fatalf("Client run 1 failed: %v", err)
	}
	if _, err := setup.RunClientSetup(clientOpts); err != nil {
		t.Fatalf("Client run 2 failed: %v", err)
	}

	// Verify .gyrus.yaml remains intact
	data, err := os.ReadFile(filepath.Join(tempDir, ".gyrus.yaml"))
	if err != nil {
		t.Fatalf("Failed reading config after replay: %v", err)
	}
	if !strings.Contains(string(data), "core-eng") {
		t.Errorf("Expected config to preserve 'core-eng', got:\n%s", string(data))
	}
}
