package setup

import (
	"fmt"
	"os"
	"path/filepath"
)

// ConfigSetupOptions defines options for workspace configuration setup (gyrus init config).
type ConfigSetupOptions struct {
	WorkspaceDir string
	Profile      Profile
	OwnerGroup   string
}

// ConfigSetupResult contains status details from workspace configuration setup.
type ConfigSetupResult struct {
	ConfigFile string
	StorageDir string
	Profile    Profile
	OwnerGroup string
}

// RunConfigSetup generates .gyrus.yaml configuration for the target workspace.
func RunConfigSetup(opts ConfigSetupOptions) (*ConfigSetupResult, error) {
	if opts.WorkspaceDir == "" {
		opts.WorkspaceDir = "."
	}
	if opts.Profile == "" {
		opts.Profile = ProfileLocal
	}
	if opts.OwnerGroup == "" {
		opts.OwnerGroup = "armckinney"
	}

	cfgPath, err := WriteConfigFile(opts.WorkspaceDir, opts.Profile, opts.OwnerGroup)
	if err != nil {
		return nil, err
	}

	return &ConfigSetupResult{
		ConfigFile: cfgPath,
		StorageDir: filepath.Join(opts.WorkspaceDir, ".gyrus"),
		Profile:    opts.Profile,
		OwnerGroup: opts.OwnerGroup,
	}, nil
}

// ClientSetupOptions defines options for installing Agent Plugins and registering MCP (gyrus init client).
type ClientSetupOptions struct {
	WorkspaceDir   string
	Target         ClientTarget // antigravity, claude, codex, copilot
	Mode           MCPMode      // container, local
	Global         bool
	ContainerImage string
	BinaryCmd      string
	PluginDir      string
}

// ClientSetupResult contains execution details from client plugin & MCP equipping.
type ClientSetupResult struct {
	Target       ClientTarget
	PluginDir    string
	PluginFiles  []string
	InstalledMCP []string
}

// RunClientSetup installs the Gyrus Agent Plugin and registers MCP servers for an explicit client target.
func RunClientSetup(opts ClientSetupOptions) (*ClientSetupResult, error) {
	if opts.WorkspaceDir == "" {
		opts.WorkspaceDir = "."
	}
	target, err := ValidateClientTarget(string(opts.Target))
	if err != nil {
		return nil, err
	}
	if opts.Mode == "" {
		opts.Mode = MCPModeContainer
	}
	if opts.BinaryCmd == "" {
		opts.BinaryCmd = "gyrus"
	}
	if opts.ContainerImage == "" {
		opts.ContainerImage = "ghcr.io/armckinney/gyrus:latest"
	}

	result := &ClientSetupResult{
		Target: target,
	}

	// 1. Install Agent Plugin for targets that discover plugins (antigravity, copilot, codex)
	if target != ClientTargetClaude {
		pluginDir := opts.PluginDir
		if pluginDir == "" {
			if opts.Global {
				userHome, _ := os.UserHomeDir()
				if target == ClientTargetAntigravity {
					pluginDir = filepath.Join(userHome, ".gemini", "config", "plugins", "gyrus")
				} else {
					pluginDir = filepath.Join(userHome, ".agents", "plugins", "gyrus")
				}
			} else {
				pluginDir = filepath.Join(opts.WorkspaceDir, ".agents", "plugins", "gyrus")
			}
		}
		result.PluginDir = pluginDir

		files, err := InstallAgentPlugin(PluginInstallOptions{
			TargetDir:      pluginDir,
			WorkspaceDir:   opts.WorkspaceDir,
			Mode:           opts.Mode,
			ContainerImage: opts.ContainerImage,
			BinaryCmd:      opts.BinaryCmd,
		})
		if err != nil {
			return nil, fmt.Errorf("failed installing Agent Plugin: %w", err)
		}
		result.PluginFiles = files
	}

	// 2. Register MCP server configurations
	mcpFiles, err := RegisterMCPServer(opts.WorkspaceDir, MCPTarget(target), opts.Mode, opts.Global, opts.BinaryCmd, opts.ContainerImage)
	if err != nil {
		return nil, fmt.Errorf("failed registering MCP server: %w", err)
	}
	result.InstalledMCP = mcpFiles

	return result, nil
}

// SetupOptions defines legacy options for workspace initialization (deprecated in Phase 2.6).
type SetupOptions struct {
	WorkspaceDir   string
	Profile        Profile
	OwnerGroup     string
	MCPTarget      MCPTarget
	MCPMode        MCPMode
	GlobalMCP      bool
	ContainerImage string
	BinaryCmd      string
	SkipMCP        bool
	SkipPlugin     bool
	SkipConfig     bool
	PluginDir      string
}

// SetupResult contains execution status details from workspace setup (deprecated in Phase 2.6).
type SetupResult struct {
	ConfigFile   string
	StorageDir   string
	InstalledMCP []string
	PluginFiles  []string
	PluginDir    string
}

// RunSetup provides compatibility by running both config and client setup when an explicit target is specified.
func RunSetup(opts SetupOptions) (*SetupResult, error) {
	if opts.WorkspaceDir == "" {
		opts.WorkspaceDir = "."
	}
	result := &SetupResult{
		StorageDir: filepath.Join(opts.WorkspaceDir, ".gyrus"),
	}

	if !opts.SkipConfig {
		cfgRes, err := RunConfigSetup(ConfigSetupOptions{
			WorkspaceDir: opts.WorkspaceDir,
			Profile:      opts.Profile,
			OwnerGroup:   opts.OwnerGroup,
		})
		if err != nil {
			return nil, err
		}
		result.ConfigFile = cfgRes.ConfigFile
	}

	if opts.MCPTarget != "" {
		clientRes, err := RunClientSetup(ClientSetupOptions{
			WorkspaceDir:   opts.WorkspaceDir,
			Target:         ClientTarget(opts.MCPTarget),
			Mode:           opts.MCPMode,
			Global:         opts.GlobalMCP,
			ContainerImage: opts.ContainerImage,
			BinaryCmd:      opts.BinaryCmd,
			PluginDir:      opts.PluginDir,
		})
		if err != nil {
			return nil, err
		}
		result.InstalledMCP = clientRes.InstalledMCP
		result.PluginFiles = clientRes.PluginFiles
		result.PluginDir = clientRes.PluginDir
	}

	return result, nil
}
