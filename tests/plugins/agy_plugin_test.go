package plugins_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/setup"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies Agent Plugin packaging compliance with Agent Plugins Standard 1.0.0 and Google Antigravity.
// [Execution Surface]: Filesystem Workspace Integration (.agents/plugins/gyrus/ and .antigravity/mcp.json)
// [Assertions]: plugin.json and mcp.json adhere to schemas; rules and skills are correctly populated.
// -----------------------------------------------------------------------------
func TestAntigravityPluginIntegration(t *testing.T) {
	tempWorkspace := t.TempDir()

	res, err := setup.RunClientSetup(setup.ClientSetupOptions{
		WorkspaceDir: tempWorkspace,
		Target:       setup.ClientTargetAntigravity,
		Mode:         setup.MCPModeLocal,
		BinaryCmd:    "gyrus",
	})
	if err != nil {
		t.Fatalf("RunClientSetup for Antigravity failed: %v", err)
	}

	pluginDir := res.PluginDir
	if pluginDir != filepath.Join(tempWorkspace, ".agents", "plugins", "gyrus") {
		t.Errorf("Unexpected plugin dir: %s", pluginDir)
	}

	// 1. Validate .antigravity/mcp.json
	mcpPath := filepath.Join(tempWorkspace, ".antigravity", "mcp.json")
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("Failed to read .antigravity/mcp.json: %v", err)
	}

	var mcpCfg setup.MCPConfigFile
	if err := json.Unmarshal(data, &mcpCfg); err != nil {
		t.Fatalf("Failed to parse .antigravity/mcp.json: %v", err)
	}

	server, exists := mcpCfg.MCPServers["gyrus"]
	if !exists {
		t.Fatalf("Antigravity MCP config does not contain 'gyrus' server")
	}
	if server.Command != "gyrus" {
		t.Errorf("Unexpected server command: %s", server.Command)
	}

	// 2. Validate plugin.json compliance against Agent Plugins Standard 1.0.0
	pluginManifestPath := filepath.Join(pluginDir, "plugin.json")
	manifestBytes, err := os.ReadFile(pluginManifestPath)
	if err != nil {
		t.Fatalf("Failed to read plugin.json: %v", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("Failed to unmarshal plugin.json: %v", err)
	}

	if manifest["$schema"] != "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json" {
		t.Errorf("plugin.json missing valid $schema: %v", manifest["$schema"])
	}

	pluginName, ok := manifest["name"].(string)
	if !ok || pluginName == "" {
		t.Fatalf("plugin.json missing name field")
	}
	if strings.Contains(pluginName, "--") || strings.Contains(pluginName, "..") {
		t.Errorf("Plugin name '%s' contains consecutive hyphens or periods", pluginName)
	}
	namePattern := regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)
	if !namePattern.MatchString(pluginName) {
		t.Errorf("Plugin name '%s' violates schema naming pattern", pluginName)
	}

	// 3. Validate mcp.json compliance against Agent Plugins Standard 1.0.0
	stdMCPPath := filepath.Join(pluginDir, "mcp.json")
	stdMCPBytes, err := os.ReadFile(stdMCPPath)
	if err != nil {
		t.Fatalf("Failed to read mcp.json: %v", err)
	}
	var stdMCP map[string]interface{}
	if err := json.Unmarshal(stdMCPBytes, &stdMCP); err != nil {
		t.Fatalf("Failed to unmarshal mcp.json: %v", err)
	}
	if stdMCP["$schema"] != "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json" {
		t.Errorf("mcp.json missing valid $schema: %v", stdMCP["$schema"])
	}

	// 4. Validate rules/AGENTS.md
	rulesPath := filepath.Join(pluginDir, "rules", "AGENTS.md")
	rulesBytes, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("Failed reading plugin rules: %v", err)
	}
	if !strings.Contains(string(rulesBytes), "gyrus suggest-context") {
		t.Errorf("Plugin rules missing suggest-context instruction")
	}

	// 5. Validate skills inside plugin
	cliSkillPath := filepath.Join(pluginDir, "skills", "gyrus-cli", "SKILL.md")
	skillContent, err := os.ReadFile(cliSkillPath)
	if err != nil {
		t.Fatalf("Failed reading equipped CLI skill: %v", err)
	}
	contentStr := string(skillContent)
	if !strings.Contains(contentStr, "name: gyrus-cli") {
		t.Errorf("Skill missing valid YAML frontmatter name")
	}
}
