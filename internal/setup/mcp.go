package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// MCPTarget identifies an AI agent platform for MCP auto-registration.
type MCPTarget string

const (
	MCPTargetClaude      MCPTarget = "claude"
	MCPTargetAntigravity MCPTarget = "antigravity" // Cursor / Antigravity
	MCPTargetCodex       MCPTarget = "codex"
	MCPTargetCopilot     MCPTarget = "copilot"
	MCPTargetAll         MCPTarget = "all"
)

// ServerConfig defines an individual stdio MCP server definition.
type ServerConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// MCPConfigFile represents the standard JSON envelope for mcpServers.
type MCPConfigFile struct {
	MCPServers map[string]ServerConfig `json:"mcpServers,omitempty"`
}

// RegisterMCPServer registers the Gyrus stdio MCP server into target config JSON files.
func RegisterMCPServer(workspaceDir string, target MCPTarget, binaryCmd string) ([]string, error) {
	if binaryCmd == "" {
		binaryCmd = "gyrus"
	}

	targets := []MCPTarget{target}
	if target == MCPTargetAll {
		targets = []MCPTarget{MCPTargetAntigravity, MCPTargetClaude, MCPTargetCodex, MCPTargetCopilot}
	}

	var updatedFiles []string

	for _, t := range targets {
		paths := getTargetConfigPaths(workspaceDir, t)
		for _, path := range paths {
			if err := injectMCPServerJSON(path, binaryCmd); err != nil {
				return updatedFiles, fmt.Errorf("failed to register MCP for target %s at %s: %w", t, path, err)
			}
			updatedFiles = append(updatedFiles, path)
		}
	}

	return updatedFiles, nil
}

func getTargetConfigPaths(workspaceDir string, target MCPTarget) []string {
	userHome, _ := os.UserHomeDir()

	switch target {
	case MCPTargetClaude:
		var path string
		switch runtime.GOOS {
		case "darwin":
			path = filepath.Join(userHome, "Library", "Application Support", "Claude", "claude_desktop_config.json")
		case "windows":
			appData := os.Getenv("APPDATA")
			if appData == "" {
				appData = filepath.Join(userHome, "AppData", "Roaming")
			}
			path = filepath.Join(appData, "Claude", "claude_desktop_config.json")
		default: // linux
			path = filepath.Join(userHome, ".config", "Claude", "claude_desktop_config.json")
		}
		return []string{
			filepath.Join(workspaceDir, ".claude", "mcp.json"),
			path,
		}

	case MCPTargetAntigravity:
		return []string{
			filepath.Join(workspaceDir, ".antigravity", "mcp.json"),
		}

	case MCPTargetCodex:
		return []string{
			filepath.Join(workspaceDir, ".codex", "mcp.json"),
		}

	case MCPTargetCopilot:
		return []string{
			filepath.Join(workspaceDir, ".vscode", "mcp.json"),
		}

	default:
		return nil
	}
}

func injectMCPServerJSON(filePath string, binaryCmd string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var root map[string]interface{}
	if data, err := os.ReadFile(filePath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &root); err != nil {
			root = make(map[string]interface{})
		}
	} else {
		root = make(map[string]interface{})
	}

	var mcpServers map[string]interface{}
	if rawServers, ok := root["mcpServers"].(map[string]interface{}); ok {
		mcpServers = rawServers
	} else {
		mcpServers = make(map[string]interface{})
	}

	mcpServers["gyrus"] = map[string]interface{}{
		"command": binaryCmd,
		"args":    []string{"mcp", "serve"},
	}

	root["mcpServers"] = mcpServers

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, out, 0644)
}
