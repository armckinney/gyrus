package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/setup"
	"github.com/spf13/cobra"
)

// NewClientCmd constructs the 'client' parent Cobra command and its subcommands.
func NewClientCmd(application *app.App) *cobra.Command {
	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Agent client lifecycle commands (install, uninstall)",
	}

	clientCmd.AddCommand(NewClientInstallCmd(application))
	clientCmd.AddCommand(NewClientUninstallCmd(application))

	return clientCmd
}

// NewClientInstallCmd constructs the 'client install' Cobra subcommand.
func NewClientInstallCmd(application *app.App) *cobra.Command {
	var (
		target    string
		global    bool
		pluginDir string
	)

	cmd := &cobra.Command{
		Use:   "install",
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

			binaryCmd := "gyrus"
			if _, err := exec.LookPath("gyrus"); err != nil {
				if self, err := os.Executable(); err == nil && filepath.IsAbs(self) {
					binaryCmd = self
				}
			}

			res, err := setup.RunClientSetup(setup.ClientSetupOptions{
				WorkspaceDir: cwd,
				Target:       clientTarget,
				Mode:         setup.MCPModeLocal,
				Global:       global,
				BinaryCmd:    binaryCmd,
				PluginDir:    pluginDir,
			})
			if err != nil {
				return err
			}

			agyRegistered := false
			if clientTarget == setup.ClientTargetAntigravity && res.PluginDir != "" {
				agyRegistered = registerWithAntigravity(res.PluginDir)
			}

			// Automatic initial sync to hydrate the SQLite index
			syncedDocs := 0
			if engine, err := application.Engine(); err == nil {
				if report, err := engine.Sync(context.Background()); err == nil {
					syncedDocs = report.IndexedFiles + report.UnchangedFiles
				}
			}

			if !GlobalJSONOutput {
				fmt.Printf("🚀 Equipped Gyrus for client '%s'!\n", clientTarget)
				if res.PluginDir != "" {
					fmt.Printf("   - Agent Plugin:  %s (%d files written)\n", res.PluginDir, len(res.PluginFiles))
				}
				if agyRegistered {
					fmt.Printf("   - Antigravity:   Registered in plugin registry (agy plugin list)\n")
				}
				modeDesc := "Mode: local (native binary)"
				if global {
					modeDesc += ", Global: ~"
				}
				fmt.Printf("   - MCP Servers:   Registered for target '%s' (%d files updated, %s)\n", clientTarget, len(res.InstalledMCP), modeDesc)
				for _, f := range res.InstalledMCP {
					fmt.Printf("      • %s\n", f)
				}
				if syncedDocs > 0 {
					fmt.Printf("   - Index Status:  %d documents ready\n", syncedDocs)
				}
			} else {
				fmt.Printf("{\"status\":\"installed\",\"target\":\"%s\",\"mode\":\"local\",\"plugin_dir\":\"%s\",\"plugin_files_count\":%d,\"mcp_files_count\":%d,\"agy_registered\":%t,\"synced_docs\":%d}\n",
					clientTarget, res.PluginDir, len(res.PluginFiles), len(res.InstalledMCP), agyRegistered, syncedDocs)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target agent tool (required): antigravity, claude, codex, copilot")
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Register MCP servers and plugin globally in user home directory (~)")
	cmd.Flags().StringVar(&pluginDir, "plugin-dir", "", "Custom destination directory for Agent Plugin bundle")

	return cmd
}

// NewClientUninstallCmd constructs the 'client uninstall' Cobra subcommand.
func NewClientUninstallCmd(application *app.App) *cobra.Command {
	var (
		target    string
		global    bool
		pluginDir string
	)

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Agent Plugin and unregister MCP across target AI tools",
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

			res, err := setup.RunClientUninstall(setup.ClientUninstallOptions{
				WorkspaceDir: cwd,
				Target:       clientTarget,
				Global:       global,
				PluginDir:    pluginDir,
			})
			if err != nil {
				return err
			}

			agyDeregistered := false
			if clientTarget == setup.ClientTargetAntigravity {
				agyDeregistered = deregisterWithAntigravity()
			}

			if !GlobalJSONOutput {
				fmt.Printf("🗑️  Uninstalled Gyrus from client '%s'!\n", clientTarget)
				if res.PluginDir != "" {
					if res.PluginRemoved {
						fmt.Printf("   - Agent Plugin:  Removed (%s)\n", res.PluginDir)
					} else {
						fmt.Printf("   - Agent Plugin:  Not found or already removed (%s)\n", res.PluginDir)
					}
				}
				if agyDeregistered {
					fmt.Printf("   - Antigravity:   Deregistered from plugin registry (agy plugin uninstall gyrus)\n")
				}
				fmt.Printf("   - MCP Servers:   Cleaned for target '%s' (%d config files updated)\n", clientTarget, len(res.UnregisteredMCP))
				for _, f := range res.UnregisteredMCP {
					fmt.Printf("      • %s\n", f)
				}
			} else {
				fmt.Printf("{\"status\":\"uninstalled\",\"target\":\"%s\",\"plugin_dir\":\"%s\",\"plugin_removed\":%t,\"mcp_files_count\":%d,\"agy_deregistered\":%t}\n",
					clientTarget, res.PluginDir, res.PluginRemoved, len(res.UnregisteredMCP), agyDeregistered)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target agent tool (required): antigravity, claude, codex, copilot")
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Remove MCP servers and plugin globally from user home directory (~)")
	cmd.Flags().StringVar(&pluginDir, "plugin-dir", "", "Custom destination directory of Agent Plugin bundle to remove")

	return cmd
}

// deregisterWithAntigravity attempts to deregister the plugin using the agy CLI if present.
func deregisterWithAntigravity() bool {
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

	cmd := exec.Command(agyPath, "plugin", "uninstall", "gyrus")
	return cmd.Run() == nil
}
