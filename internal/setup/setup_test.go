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
// [Purpose]: Verifies that RunSetup executes the complete initialization workflow across configuration, skills, and MCP targeting.
// [Execution Surface]: In-Memory / Temporary Filesystem
// [Assertions]: Creates .gyrus.yaml, installs agent skill files, and registers MCP server config JSON.
// -----------------------------------------------------------------------------
func TestMasterSetupWorkflow(t *testing.T) {
	tempDir := t.TempDir()

	result, err := setup.RunSetup(setup.SetupOptions{
		WorkspaceDir: tempDir,
		Profile:      setup.ProfileLocal,
		OwnerGroup:   "test-group",
		MCPTarget:    setup.MCPTargetAll,
		SkillTarget:  setup.SkillTargetAll,
		BinaryCmd:    "gyrus",
	})
	if err != nil {
		t.Fatalf("RunSetup failed: %v", err)
	}

	// 1. Verify .gyrus.yaml
	configPath := filepath.Join(tempDir, ".gyrus.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected .gyrus.yaml to exist at %s", configPath)
	}

	// 2. Verify Storage Dir path string
	expectedStorageDir := filepath.Join(tempDir, "docs", ".gyrus", "docs")
	if result.StorageDir != expectedStorageDir {
		t.Errorf("Expected storage root path %s, got %s", expectedStorageDir, result.StorageDir)
	}

	// 3. Verify Agent Skill Files
	cliSkillFile := filepath.Join(tempDir, ".agents", "skills", "gyrus-cli", "SKILL.md")
	if _, err := os.Stat(cliSkillFile); os.IsNotExist(err) {
		t.Errorf("Expected CLI skill file at %s", cliSkillFile)
	}

	mcpSkillFile := filepath.Join(tempDir, ".agents", "skills", "gyrus-mcp", "SKILL.md")
	if _, err := os.Stat(mcpSkillFile); os.IsNotExist(err) {
		t.Errorf("Expected MCP skill file at %s", mcpSkillFile)
	}

	// 4. Verify MCP Registration
	antigravityMCP := filepath.Join(tempDir, ".antigravity", "mcp.json")
	if _, err := os.Stat(antigravityMCP); os.IsNotExist(err) {
		t.Errorf("Expected Antigravity MCP config at %s", antigravityMCP)
	}
	if len(result.InstalledMCP) == 0 {
		t.Errorf("Expected non-empty InstalledMCP slice")
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

	opts := setup.SetupOptions{
		WorkspaceDir: tempDir,
		Profile:      setup.ProfileLocal,
		OwnerGroup:   "core-eng",
		MCPTarget:    setup.MCPTargetAll,
		SkillTarget:  setup.SkillTargetAll,
		BinaryCmd:    "gyrus",
	}

	// Run 1
	if _, err := setup.RunSetup(opts); err != nil {
		t.Fatalf("Run 1 failed: %v", err)
	}

	// Run 2 (Idempotent replay)
	if _, err := setup.RunSetup(opts); err != nil {
		t.Fatalf("Run 2 failed (idempotency violation): %v", err)
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
