package plugins_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies Agent Plugin packaging compliance with Agent Plugins Standard 1.0.0 and Google Antigravity.
// [Execution Surface]: Canonical packaging repository directory (packaging/plugins/gyrus/)
// [Assertions]: plugin.json and mcp.json adhere to schemas; rules and skills are correctly populated.
// -----------------------------------------------------------------------------
func TestAntigravityPluginIntegration(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repository root: %v", err)
	}

	pluginDir := filepath.Join(repoRoot, "packaging", "plugins", "gyrus")

	// 1. Validate plugin mcp_config.json (Antigravity discovers MCP directly from plugin bundle)
	mcpPath := filepath.Join(pluginDir, "mcp_config.json")
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("Failed to read plugin mcp_config.json: %v", err)
	}

	var mcpCfg struct {
		MCPServers map[string]struct {
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &mcpCfg); err != nil {
		t.Fatalf("Failed to parse plugin mcp_config.json: %v", err)
	}

	server, exists := mcpCfg.MCPServers["gyrus"]
	if !exists {
		t.Fatalf("Antigravity MCP config does not contain 'gyrus' server")
	}
	if server.Command != "gyrus" {
		t.Errorf("Unexpected server command: %s", server.Command)
	}
	if server.Env["GYRUS_WORKSPACE"] != "${workspaceFolder}" {
		t.Errorf("Expected GYRUS_WORKSPACE env to be '${workspaceFolder}', got '%s'", server.Env["GYRUS_WORKSPACE"])
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
	var stdMCP struct {
		Schema     string `json:"$schema"`
		MCPServers map[string]struct {
			Type    string            `json:"type"`
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(stdMCPBytes, &stdMCP); err != nil {
		t.Fatalf("Failed to unmarshal mcp.json: %v", err)
	}
	if stdMCP.Schema != "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json" {
		t.Errorf("mcp.json missing valid $schema: %v", stdMCP.Schema)
	}
	stdServer, exists := stdMCP.MCPServers["gyrus"]
	if !exists {
		t.Fatalf("mcp.json missing 'gyrus' server")
	}
	if stdServer.Command != "gyrus" {
		t.Errorf("Unexpected command in mcp.json: %s", stdServer.Command)
	}
	if stdServer.Env["GYRUS_WORKSPACE"] != "${workspaceFolder}" {
		t.Errorf("Expected GYRUS_WORKSPACE env to be '${workspaceFolder}', got '%s'", stdServer.Env["GYRUS_WORKSPACE"])
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
	unifiedSkillPath := filepath.Join(pluginDir, "skills", "gyrus", "SKILL.md")
	skillContent, err := os.ReadFile(unifiedSkillPath)
	if err != nil {
		t.Fatalf("Failed reading equipped Gyrus skill: %v", err)
	}
	contentStr := string(skillContent)
	if !strings.Contains(contentStr, "name: gyrus") {
		t.Errorf("Skill missing valid YAML frontmatter name")
	}
}
