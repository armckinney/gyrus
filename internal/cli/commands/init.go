package commands

import (
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/setup"
	"github.com/spf13/cobra"
)

var (
	initProfile     string
	initOwnerGroup  string
	initMCPTarget   string
	initSkillTarget string
	initNoMCP       bool
	initNoSkill     bool
	initNoConfig    bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Gyrus storage, configuration, agent skill, and MCP server in workspace",
	Long: `Initialize Gyrus in your repository workspace.

Flags allow custom-tailoring your setup:
  - Select specific storage profiles (local, git, blob, postgres, vector)
  - Target specific agent tools for MCP registration (claude, antigravity, codex, copilot, all)
  - Skip MCP setup (--no-mcp) to configure CLI-only environments
  - Skip Agent Skill equipping (--no-skill)
  - Skip .gyrus.yaml generation (--no-config)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		res, err := setup.RunSetup(setup.SetupOptions{
			WorkspaceDir: cwd,
			Profile:      setup.Profile(initProfile),
			OwnerGroup:   initOwnerGroup,
			MCPTarget:    setup.MCPTarget(initMCPTarget),
			SkillTarget:  setup.SkillTarget(initSkillTarget),
			BinaryCmd:    "gyrus",
			SkipMCP:      initNoMCP,
			SkipSkill:    initNoSkill,
			SkipConfig:   initNoConfig,
		})
		if err != nil {
			return err
		}

		if !cli.GlobalJSONOutput {
			fmt.Printf("🚀 Initialized Gyrus workspace successfully!\n")
			if !initNoConfig {
				fmt.Printf("   - Configuration: %s (Profile: %s)\n", res.ConfigFile, initProfile)
			}
			fmt.Printf("   - Storage Root:  %s\n", res.StorageDir)
			if !initNoSkill {
				fmt.Printf("   - Agent Skill:   Equipped at .agents/skills/gyrus/ (Target: %s)\n", initSkillTarget)
			}
			if !initNoMCP {
				fmt.Printf("   - MCP Servers:   Registered for target '%s' (%d files updated)\n", initMCPTarget, len(res.InstalledMCP))
				for _, f := range res.InstalledMCP {
					fmt.Printf("      • %s\n", f)
				}
			} else {
				fmt.Printf("   - MCP Servers:   Skipped (--no-mcp set)\n")
			}
		} else {
			fmt.Printf("{\"status\":\"initialized\",\"profile\":\"%s\",\"storage\":\"%s\",\"mcp_skipped\":%v}\n", initProfile, res.StorageDir, initNoMCP)
		}
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&initProfile, "profile", "p", "local", "Configuration profile: local, git, blob, postgres, vector")
	initCmd.Flags().StringVarP(&initOwnerGroup, "owner-group", "o", "armckinney", "Default owner group for context documents")
	initCmd.Flags().StringVarP(&initMCPTarget, "mcp-target", "m", "all", "Target agent tool for MCP setup: claude, antigravity, codex, copilot, all")
	initCmd.Flags().StringVarP(&initSkillTarget, "skill-target", "s", "all", "Target agent tool for skill equipping: claude, antigravity, codex, copilot, all")
	initCmd.Flags().BoolVar(&initNoMCP, "no-mcp", false, "Skip MCP server registration (CLI-only setup)")
	initCmd.Flags().BoolVar(&initNoSkill, "no-skill", false, "Skip equipping agent skill files")
	initCmd.Flags().BoolVar(&initNoConfig, "no-config", false, "Skip generating .gyrus.yaml config file")
	cli.RootCmd.AddCommand(initCmd)
}
