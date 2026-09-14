package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/mcp"
	"github.com/armckinney/gyrus/internal/setup"
	"github.com/spf13/cobra"
)

// NewInitCmd constructs the parent 'init' Cobra command with required subcommands.
func NewInitCmd(application *app.App) *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Gyrus workspace configuration or equip AI agent clients",
		Long: `Initialize Gyrus workspace configuration or equip target AI agent clients with Agent Plugins and MCP.

Explicit subcommands are required:
  gyrus init config - Generate .gyrus.yaml workspace configuration
  gyrus init client - Install Agent Plugin and configure MCP across target AI tools`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("initialization requires an explicit subcommand: 'gyrus init config' or 'gyrus init client'. Run 'gyrus init --help' for options")
		},
	}

	initCmd.AddCommand(NewInitConfigCmd(application))
	initCmd.AddCommand(NewInitClientCmd(application))

	return initCmd
}

// NewInitConfigCmd constructs the 'init config' Cobra subcommand.
func NewInitConfigCmd(application *app.App) *cobra.Command {
	var (
		profile    string
		ownerGroup string
	)

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Generate .gyrus.yaml workspace configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			targetDir := cwd
			if storageFlag, _ := cmd.Flags().GetString("storage-path"); storageFlag != "" {
				targetDir = filepath.Dir(storageFlag)
			} else if GlobalStoragePath != "" {
				targetDir = filepath.Dir(GlobalStoragePath)
			}

			res, err := setup.RunConfigSetup(setup.ConfigSetupOptions{
				WorkspaceDir: targetDir,
				Profile:      setup.Profile(profile),
				OwnerGroup:   ownerGroup,
			})
			if err != nil {
				return err
			}

			if !GlobalJSONOutput {
				fmt.Printf("🚀 Initialized Gyrus workspace configuration successfully!\n")
				fmt.Printf("   - Configuration: %s (Profile: %s)\n", res.ConfigFile, profile)
				fmt.Printf("   - Storage Root:  %s\n", res.StorageDir)
				fmt.Printf("   - Default Owner: %s\n", res.OwnerGroup)
			} else {
				fmt.Printf("{\"status\":\"configured\",\"config_file\":\"%s\",\"profile\":\"%s\",\"storage\":\"%s\",\"owner_group\":\"%s\"}\n",
					res.ConfigFile, profile, res.StorageDir, res.OwnerGroup)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&profile, "profile", "p", "local", "Configuration profile: local, git, blob, s3, azure, gcs, postgres, vector")
	cmd.Flags().StringVarP(&ownerGroup, "owner-group", "o", "armckinney", "Default owner group for context documents")

	return cmd
}

// NewInitClientCmd constructs the 'init client' Cobra subcommand.
func NewInitClientCmd(application *app.App) *cobra.Command {
	var (
		target    string
		mode      string
		image     string
		global    bool
		pluginDir string
	)

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Install Agent Plugin and configure MCP across target AI tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			if target == "" {
				return fmt.Errorf("--target is required. Valid targets: antigravity, claude, codex, copilot")
			}

			clientTarget, err := setup.ValidateClientTarget(target)
			if err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			targetDir := cwd
			if storageFlag, _ := cmd.Flags().GetString("storage-path"); storageFlag != "" {
				targetDir = filepath.Dir(storageFlag)
			} else if GlobalStoragePath != "" {
				targetDir = filepath.Dir(GlobalStoragePath)
			}

			// Determine execution mode fallback if docker is not installed
			resolvedMode := mode
			if !cmd.Flags().Changed("mode") {
				if _, err := exec.LookPath("docker"); err != nil {
					resolvedMode = string(setup.MCPModeLocal)
				}
			}

			// Resolve binary path if gyrus is not in PATH
			binaryCmd := "gyrus"
			if _, err := exec.LookPath("gyrus"); err != nil {
				if self, err := os.Executable(); err == nil && filepath.IsAbs(self) {
					binaryCmd = self
				}
			}

			res, err := setup.RunClientSetup(setup.ClientSetupOptions{
				WorkspaceDir:   targetDir,
				Target:         clientTarget,
				Mode:           setup.MCPMode(resolvedMode),
				Global:         global,
				ContainerImage: image,
				BinaryCmd:      binaryCmd,
				PluginDir:      pluginDir,
			})
			if err != nil {
				return err
			}

			// If targeting Antigravity and agy CLI is present, register into Antigravity plugin registry
			agyRegistered := false
			if clientTarget == setup.ClientTargetAntigravity && res.PluginDir != "" {
				agyRegistered = registerWithAntigravity(res.PluginDir)
			}

			if !GlobalJSONOutput {
				fmt.Printf("🚀 Equipped Gyrus for client '%s'!\n", clientTarget)
				if res.PluginDir != "" {
					fmt.Printf("   - Agent Plugin:  %s (%d files written)\n", res.PluginDir, len(res.PluginFiles))
				}
				if agyRegistered {
					fmt.Printf("   - Antigravity:   Registered in plugin registry (agy plugin list)\n")
				}
				modeDesc := fmt.Sprintf("Mode: %s", resolvedMode)
				if global {
					modeDesc += ", Global: ~"
				}
				fmt.Printf("   - MCP Servers:   Registered for target '%s' (%d files updated, %s)\n", clientTarget, len(res.InstalledMCP), modeDesc)
				for _, f := range res.InstalledMCP {
					fmt.Printf("      • %s\n", f)
				}
			} else {
				fmt.Printf("{\"status\":\"equipped\",\"target\":\"%s\",\"mode\":\"%s\",\"plugin_dir\":\"%s\",\"plugin_files_count\":%d,\"mcp_files_count\":%d,\"agy_registered\":%t}\n",
					clientTarget, resolvedMode, res.PluginDir, len(res.PluginFiles), len(res.InstalledMCP), agyRegistered)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target agent tool (required): antigravity, claude, codex, copilot")
	cmd.Flags().StringVarP(&mode, "mode", "m", "container", "MCP execution mode: container (containerized stdio via Docker), local (local binary)")
	cmd.Flags().StringVar(&image, "mcp-container-image", "ghcr.io/armckinney/gyrus:latest", "Container image for containerized stdio MCP execution")
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Register MCP servers and plugin globally in user home directory (~)")
	cmd.Flags().StringVar(&pluginDir, "plugin-dir", "", "Custom destination directory for Agent Plugin bundle")

	return cmd
}

// NewMCPCmd constructs the 'mcp' parent Cobra command and its subcommands.
func NewMCPCmd(application *app.App) *cobra.Command {
	var mcpTarget string

	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol (MCP) server and integration commands",
	}

	mcpServeCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start embedded Gyrus MCP stdio server",
		RunE: func(cmd *cobra.Command, args []string) error {
			srv, err := mcp.NewServer(application.StorageRoot())
			if err != nil {
				return fmt.Errorf("failed to start MCP server: %w", err)
			}
			return srv.ServeStdio(context.Background())
		},
	}

	mcpSetupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Register Gyrus stdio MCP server in agent tools (Claude, Antigravity, Codex, Copilot)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if mcpTarget == "" {
				return fmt.Errorf("--target is required. Valid targets: antigravity, claude, codex, copilot")
			}

			validTarget, err := setup.ValidateClientTarget(mcpTarget)
			if err != nil {
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			files, err := setup.RegisterMCPServer(cwd, setup.MCPTarget(validTarget), setup.MCPModeLocal, false, "gyrus", "")
			if err != nil {
				return err
			}

			if !GlobalJSONOutput {
				fmt.Printf("✅ Registered Gyrus stdio MCP server across agent target (%s):\n", validTarget)
				for _, f := range files {
					fmt.Printf("   - %s\n", f)
				}
			} else {
				fmt.Printf("{\"status\":\"success\",\"mcp_target\":\"%s\"}\n", validTarget)
			}
			return nil
		},
	}

	mcpSetupCmd.Flags().StringVarP(&mcpTarget, "target", "t", "antigravity", "Target agent tool: claude, antigravity, codex, copilot")
	mcpCmd.AddCommand(mcpServeCmd)
	mcpCmd.AddCommand(mcpSetupCmd)

	return mcpCmd
}

// registerWithAntigravity attempts to register the plugin using the agy CLI if present on the machine.
func registerWithAntigravity(pluginDir string) bool {
	agyPath := "agy"
	if _, err := exec.LookPath("agy"); err != nil {
		userHome, _ := os.UserHomeDir()
		candidates := []string{
			filepath.Join(userHome, ".gemini", "bin", "agy"),
			"/usr/local/bin/agy",
			"/root/.gemini/bin/agy",
		}
		found := false
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				agyPath = c
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	cmd := exec.Command(agyPath, "plugin", "install", pluginDir)
	return cmd.Run() == nil
}
