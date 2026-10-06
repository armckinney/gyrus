package plugins_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/setup"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies that .claude-plugin/marketplace.json is present as the single canonical
//
//	marketplace manifest, syntactically valid, and points to packaging/plugins/gyrus.
//
// [Execution Surface]: Filesystem Repository Root (.claude-plugin/, packaging/plugins/gyrus/)
// [Assertions]: marketplace.json adheres to schema and points to packaging/plugins/gyrus.
// -----------------------------------------------------------------------------
func TestPluginMarketplaceManifest(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repository root: %v", err)
	}

	claudeMarketplace := filepath.Join(repoRoot, ".claude-plugin", "marketplace.json")

	// 1. Verify .claude-plugin/marketplace.json exists as a regular file (not a symlink)
	fi, err := os.Lstat(claudeMarketplace)
	if err != nil {
		t.Fatalf("Failed to stat .claude-plugin/marketplace.json: %v", err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Errorf(".claude-plugin/marketplace.json should be a standalone regular file, not a symlink")
	}

	data, err := os.ReadFile(claudeMarketplace)
	if err != nil {
		t.Fatalf("Failed to read .claude-plugin/marketplace.json: %v", err)
	}

	var marketplace struct {
		Schema      string `json:"$schema"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Owner       struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"owner"`
		Plugins []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Version     string `json:"version"`
			Source      string `json:"source"`
			Category    string `json:"category"`
			Homepage    string `json:"homepage"`
		} `json:"plugins"`
	}

	if err := json.Unmarshal(data, &marketplace); err != nil {
		t.Fatalf("Failed to parse .claude-plugin/marketplace.json as valid JSON: %v", err)
	}

	if marketplace.Name != "gyrus" {
		t.Errorf("Expected marketplace name 'gyrus', got '%s'", marketplace.Name)
	}

	if len(marketplace.Plugins) == 0 {
		t.Fatalf("Marketplace must declare at least one plugin")
	}

	foundGyrus := false
	for _, plugin := range marketplace.Plugins {
		if plugin.Name == "gyrus" {
			foundGyrus = true
			if plugin.Source != "./packaging/plugins/gyrus" {
				t.Errorf("Expected plugin source './packaging/plugins/gyrus', got '%s'", plugin.Source)
			}

			// Verify source directory exists on disk
			resolvedSource := filepath.Join(repoRoot, plugin.Source)
			sfi, err := os.Stat(resolvedSource)
			if err != nil {
				t.Fatalf("Plugin source path %s does not exist on disk: %v", resolvedSource, err)
			}
			if !sfi.IsDir() {
				t.Fatalf("Plugin source path %s is not a directory", resolvedSource)
			}
		}
	}

	if !foundGyrus {
		t.Errorf("Plugin 'gyrus' not found in marketplace.json plugins catalog")
	}
}

func TestClaudeCodePluginManifestInPackaging(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repository root: %v", err)
	}

	pluginDir := filepath.Join(repoRoot, "packaging", "plugins", "gyrus")

	// 1. Check .claude-plugin/plugin.json in packaging/plugins/gyrus
	claudePluginManifest := filepath.Join(pluginDir, ".claude-plugin", "plugin.json")
	data, err := os.ReadFile(claudePluginManifest)
	if err != nil {
		t.Fatalf("Missing .claude-plugin/plugin.json at %s: %v", claudePluginManifest, err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("Invalid JSON in %s: %v", claudePluginManifest, err)
	}

	if manifest["name"] != "gyrus" {
		t.Errorf("Expected plugin name 'gyrus', got '%v'", manifest["name"])
	}

	// 2. Check .mcp.json in packaging/plugins/gyrus
	claudeMCPPath := filepath.Join(pluginDir, ".mcp.json")
	mcpData, err := os.ReadFile(claudeMCPPath)
	if err != nil {
		t.Fatalf("Missing .mcp.json at %s: %v", claudeMCPPath, err)
	}

	var mcpCfg setup.MCPConfigFile
	if err := json.Unmarshal(mcpData, &mcpCfg); err != nil {
		t.Fatalf("Invalid JSON in %s: %v", claudeMCPPath, err)
	}

	server, ok := mcpCfg.MCPServers["gyrus"]
	if !ok {
		t.Fatalf("Claude MCP config missing 'gyrus' server entry")
	}
	if server.Command != "gyrus" {
		t.Errorf("Expected command 'gyrus', got '%s'", server.Command)
	}
}

func TestInstallAgentPluginGeneratesClaudeManifest(t *testing.T) {
	tempWorkspace := t.TempDir()
	pluginDir := filepath.Join(tempWorkspace, ".agents", "plugins", "gyrus")

	files, err := setup.InstallAgentPlugin(setup.PluginInstallOptions{
		TargetDir:    pluginDir,
		WorkspaceDir: tempWorkspace,
		BinaryCmd:    "gyrus",
	})
	if err != nil {
		t.Fatalf("InstallAgentPlugin failed: %v", err)
	}

	hasClaudePluginJSON := false
	hasClaudeMCP := false
	for _, f := range files {
		if f == filepath.Join(pluginDir, ".claude-plugin", "plugin.json") {
			hasClaudePluginJSON = true
		}
		if f == filepath.Join(pluginDir, ".mcp.json") {
			hasClaudeMCP = true
		}
	}

	if !hasClaudePluginJSON {
		t.Errorf("InstallAgentPlugin did not return .claude-plugin/plugin.json in created files list")
	}
	if !hasClaudeMCP {
		t.Errorf("InstallAgentPlugin did not return .mcp.json in created files list")
	}

	// Verify file contents
	data, err := os.ReadFile(filepath.Join(pluginDir, ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatalf("Failed to read generated .claude-plugin/plugin.json: %v", err)
	}
	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("Invalid JSON in generated .claude-plugin/plugin.json: %v", err)
	}
	if manifest["name"] != "gyrus" {
		t.Errorf("Expected plugin name 'gyrus', got %v", manifest["name"])
	}
}
