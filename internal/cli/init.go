package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/setup"
	"github.com/spf13/cobra"
)

// NewInitCmd constructs the 'init' Cobra command.
func NewInitCmd(application *app.App) *cobra.Command {
	var (
		profile     string
		ownerGroup  string
		mcpTarget   string
		mcpMode     string
		mcpImage    string
		global      bool
		skillTarget string
		noMCP       bool
		noSkill     bool
		noConfig    bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Gyrus storage, configuration, agent skill, and MCP server in workspace or globally",
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

			res, err := setup.RunSetup(setup.SetupOptions{
				WorkspaceDir:   targetDir,
				Profile:        setup.Profile(profile),
				OwnerGroup:     ownerGroup,
				MCPTarget:      setup.MCPTarget(mcpTarget),
				MCPMode:        setup.MCPMode(mcpMode),
				GlobalMCP:      global,
				ContainerImage: mcpImage,
				SkillTarget:    setup.SkillTarget(skillTarget),
				BinaryCmd:      "gyrus",
				SkipMCP:        noMCP,
				SkipSkill:      noSkill,
				SkipConfig:     noConfig,
			})
			if err != nil {
				return err
			}

			if !GlobalJSONOutput {
				fmt.Printf("🚀 Initialized Gyrus workspace successfully!\n")
				if !noConfig {
					fmt.Printf("   - Configuration: %s (Profile: %s)\n", res.ConfigFile, profile)
				}
				fmt.Printf("   - Storage Root:  %s\n", res.StorageDir)
				if !noSkill {
					fmt.Printf("   - Agent Skill:   Equipped at .agents/skills/gyrus/ (Target: %s)\n", skillTarget)
				}
				if !noMCP {
					modeDesc := fmt.Sprintf("Mode: %s", mcpMode)
					if global {
						modeDesc += ", Global: ~"
					}
					fmt.Printf("   - MCP Servers:   Registered for target '%s' (%d files updated, %s)\n", mcpTarget, len(res.InstalledMCP), modeDesc)
					for _, f := range res.InstalledMCP {
						fmt.Printf("      • %s\n", f)
					}
				} else {
					fmt.Printf("   - MCP Servers:   Skipped (--no-mcp set)\n")
				}
			} else {
				fmt.Printf("{\"status\":\"initialized\",\"profile\":\"%s\",\"storage\":\"%s\",\"mcp_skipped\":%v}\n", profile, res.StorageDir, noMCP)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&profile, "profile", "p", "local", "Configuration profile: local, git, blob, postgres, vector")
	cmd.Flags().StringVarP(&ownerGroup, "owner-group", "o", "armckinney", "Default owner group for context documents")
	cmd.Flags().StringVarP(&mcpTarget, "mcp-target", "m", "all", "Target agent tool for MCP setup: claude, antigravity, codex, copilot, all")
	cmd.Flags().StringVar(&mcpMode, "mcp-mode", "container", "MCP execution mode: container (containerized stdio via Docker), local (local binary)")
	cmd.Flags().StringVar(&mcpImage, "mcp-container-image", "ghcr.io/armckinney/gyrus:latest", "Container image for containerized stdio MCP execution")
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Register MCP servers globally in user home directory (~)")
	cmd.Flags().StringVarP(&skillTarget, "skill-target", "s", "all", "Target agent tool for skill equipping: claude, antigravity, codex, copilot, all")
	cmd.Flags().BoolVar(&noMCP, "no-mcp", false, "Skip MCP server registration (CLI-only setup)")
	cmd.Flags().BoolVar(&noSkill, "no-skill", false, "Skip equipping agent skill files")
	cmd.Flags().BoolVar(&noConfig, "no-config", false, "Skip generating .gyrus.yaml config file")

	return cmd
}
