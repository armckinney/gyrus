package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

			claudeRegistered := false
			if clientTarget == setup.ClientTargetClaude {
				claudeRegistered = registerWithClaude(binaryCmd)
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
				if claudeRegistered {
					fmt.Printf("   - Claude Code:   Registered via claude CLI (claude mcp list)\n")
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
				fmt.Printf("{\"status\":\"installed\",\"target\":\"%s\",\"mode\":\"local\",\"plugin_dir\":\"%s\",\"plugin_files_count\":%d,\"mcp_files_count\":%d,\"agy_registered\":%t,\"claude_registered\":%t,\"synced_docs\":%d}\n",
					clientTarget, res.PluginDir, len(res.PluginFiles), len(res.InstalledMCP), agyRegistered, claudeRegistered, syncedDocs)
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

			claudeDeregistered := false
			if clientTarget == setup.ClientTargetClaude {
				claudeDeregistered = deregisterWithClaude()
			}

			killedProcesses := terminateMCPServers()

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
					fmt.Printf("   - Antigravity:   Deregistered from plugin registry & MCP cleaned\n")
				}
				if claudeDeregistered {
					fmt.Printf("   - Claude Code:   Deregistered from claude CLI (claude mcp remove gyrus)\n")
				}
				fmt.Printf("   - MCP Servers:   Cleaned for target '%s' (%d config files updated)\n", clientTarget, len(res.UnregisteredMCP))
				for _, f := range res.UnregisteredMCP {
					fmt.Printf("      • %s\n", f)
				}
				if killedProcesses > 0 {
					fmt.Printf("   - Processes:     Terminated %d active 'gyrus mcp serve' process(es)\n", killedProcesses)
				}
			} else {
				fmt.Printf("{\"status\":\"uninstalled\",\"target\":\"%s\",\"plugin_dir\":\"%s\",\"plugin_removed\":%t,\"mcp_files_count\":%d,\"agy_deregistered\":%t,\"claude_deregistered\":%t,\"killed_processes\":%d}\n",
					clientTarget, res.PluginDir, res.PluginRemoved, len(res.UnregisteredMCP), agyDeregistered, claudeDeregistered, killedProcesses)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target agent tool (required): antigravity, claude, codex, copilot")
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Remove MCP servers and plugin globally from user home directory (~)")
	cmd.Flags().StringVar(&pluginDir, "plugin-dir", "", "Custom destination directory of Agent Plugin bundle to remove")

	return cmd
}

// findAgyPath searches PATH and standard installation paths for the agy CLI.
func findAgyPath() (string, bool) {
	if p, err := exec.LookPath("agy"); err == nil {
		return p, true
	}
	userHome, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(userHome, ".local", "bin", "agy"),
		filepath.Join(userHome, ".gemini", "bin", "agy"),
		"/usr/local/bin/agy",
		"/root/.local/bin/agy",
		"/root/.gemini/bin/agy",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, true
		}
	}
	return "", false
}

// registerWithAntigravity attempts to register the plugin using the agy CLI if present.
func registerWithAntigravity(pluginDir string) bool {
	agyPath, ok := findAgyPath()
	if !ok {
		return false
	}
	cmd := exec.Command(agyPath, "plugin", "install", pluginDir)
	return cmd.Run() == nil
}

// deregisterWithAntigravity attempts to deregister the plugin and clean MCP servers using the agy CLI.
func deregisterWithAntigravity() bool {
	agyPath, ok := findAgyPath()
	if !ok {
		return false
	}

	// 1. Uninstall plugin
	pluginCmd := exec.Command(agyPath, "plugin", "uninstall", "gyrus")
	pluginSuccess := pluginCmd.Run() == nil

	// 2. Remove any direct MCP servers (gyrus, gyrus_gyrus)
	_ = exec.Command(agyPath, "mcp", "remove", "gyrus").Run()
	_ = exec.Command(agyPath, "mcp", "remove", "gyrus_gyrus").Run()

	return pluginSuccess
}

// findClaudePath searches PATH and standard installation paths for the claude CLI.
func findClaudePath() (string, bool) {
	if p, err := exec.LookPath("claude"); err == nil {
		return p, true
	}
	userHome, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(userHome, ".local", "bin", "claude"),
		"/usr/local/bin/claude",
		filepath.Join(userHome, ".npm-global", "bin", "claude"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, true
		}
	}
	return "", false
}

// registerWithClaude registers gyrus as an MCP server with the claude CLI if installed.
func registerWithClaude(binaryCmd string) bool {
	claudePath, ok := findClaudePath()
	if !ok {
		return false
	}
	cmd := exec.Command(claudePath, "mcp", "add", "gyrus", "--", binaryCmd, "mcp", "serve")
	return cmd.Run() == nil
}

// deregisterWithClaude removes the gyrus MCP server from the claude CLI if installed.
func deregisterWithClaude() bool {
	claudePath, ok := findClaudePath()
	if !ok {
		return false
	}
	cmd := exec.Command(claudePath, "mcp", "remove", "gyrus")
	return cmd.Run() == nil
}

// terminateMCPServers finds and terminates active 'gyrus mcp serve' background processes.
func terminateMCPServers() int {
	currentPID := os.Getpid()
	killedCount := 0

	pgrepPath, err := exec.LookPath("pgrep")
	if err != nil {
		return 0
	}

	out, err := exec.Command(pgrepPath, "-f", "gyrus mcp serve").Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var pid int
		if _, err := fmt.Sscanf(line, "%d", &pid); err == nil && pid != currentPID && pid > 1 {
			if proc, err := os.FindProcess(pid); err == nil {
				_ = proc.Signal(os.Interrupt)
				killedCount++
			}
		}
	}
	return killedCount
}
