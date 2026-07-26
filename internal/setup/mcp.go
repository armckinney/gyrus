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
	MCPTargetAntigravity MCPTarget = "antigravity"
	MCPTargetCodex       MCPTarget = "codex"
	MCPTargetCopilot     MCPTarget = "copilot"
	MCPTargetAll         MCPTarget = "all"
)

// MCPMode specifies local binary execution vs containerized stdio execution.
type MCPMode string

const (
	MCPModeLocal     MCPMode = "local"
	MCPModeContainer MCPMode = "container"
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

// RegisterMCPServer registers or merges the Gyrus stdio MCP server into target config JSON files.
func RegisterMCPServer(workspaceDir string, target MCPTarget, mode MCPMode, isGlobal bool, binaryCmd string, containerImage string) ([]string, error) {
	if binaryCmd == "" {
		binaryCmd = "gyrus"
	}
	if containerImage == "" {
		containerImage = "ghcr.io/armckinney/gyrus:latest"
	}

	targets := []MCPTarget{target}
	if target == MCPTargetAll {
		targets = []MCPTarget{MCPTargetAntigravity, MCPTargetClaude, MCPTargetCodex, MCPTargetCopilot}
	}

	var command string
	var args []string

	if mode == MCPModeContainer {
		command = "docker"
		args = []string{"run", "-i", "--rm", "-v", workspaceDir + ":/workspace", "-w", "/workspace", containerImage, "mcp", "serve"}
	} else {
		command = binaryCmd
		args = []string{"mcp", "serve"}
	}

	var updatedFiles []string

	for _, t := range targets {
		paths := getTargetConfigPaths(workspaceDir, t, isGlobal)
		for _, path := range paths {
			if err := injectMCPServerJSON(path, command, args); err != nil {
				return updatedFiles, fmt.Errorf("failed to register MCP for target %s at %s: %w", t, path, err)
			}
			updatedFiles = append(updatedFiles, path)
		}
	}

	return updatedFiles, nil
}

func getTargetConfigPaths(workspaceDir string, target MCPTarget, isGlobal bool) []string {
	userHome, _ := os.UserHomeDir()
	baseDir := workspaceDir
	if isGlobal {
		baseDir = userHome
	}

	switch target {
	case MCPTargetClaude:
		var desktopPath string
		switch runtime.GOOS {
		case "darwin":
			desktopPath = filepath.Join(userHome, "Library", "Application Support", "Claude", "claude_desktop_config.json")
		case "windows":
			appData := os.Getenv("APPDATA")
			if appData == "" {
				appData = filepath.Join(userHome, "AppData", "Roaming")
			}
			desktopPath = filepath.Join(appData, "Claude", "claude_desktop_config.json")
		default: // linux
			desktopPath = filepath.Join(userHome, ".config", "Claude", "claude_desktop_config.json")
		}

		if isGlobal {
			return []string{desktopPath, filepath.Join(userHome, ".claude", "mcp.json")}
		}
		return []string{
			filepath.Join(baseDir, ".claude", "mcp.json"),
			desktopPath,
		}

	case MCPTargetAntigravity:
		return []string{
			filepath.Join(baseDir, ".antigravity", "mcp.json"),
		}

	case MCPTargetCodex:
		return []string{
			filepath.Join(baseDir, ".codex", "mcp.json"),
		}

	case MCPTargetCopilot:
		return []string{
			filepath.Join(baseDir, ".vscode", "mcp.json"),
		}

	default:
		return nil
	}
}

// injectMCPServerJSON safely merges the gyrus server entry into existing mcpServers without removing user configurations.
func injectMCPServerJSON(filePath string, command string, args []string) error {
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
		"command": command,
		"args":    args,
	}

	root["mcpServers"] = mcpServers

	out, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, out, 0644)
}
