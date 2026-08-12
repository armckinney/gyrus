package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/mcp"
	"github.com/armckinney/gyrus/internal/setup"
	"github.com/spf13/cobra"
)

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
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			files, err := setup.RegisterMCPServer(cwd, setup.MCPTarget(mcpTarget), setup.MCPModeLocal, false, "gyrus", "")
			if err != nil {
				return err
			}

			if !GlobalJSONOutput {
				fmt.Printf("✅ Registered Gyrus stdio MCP server across agent targets (%s):\n", mcpTarget)
				for _, f := range files {
					fmt.Printf("   - %s\n", f)
				}
			} else {
				fmt.Printf("{\"status\":\"success\",\"mcp_target\":\"%s\"}\n", mcpTarget)
			}
			return nil
		},
	}

	mcpSetupCmd.Flags().StringVarP(&mcpTarget, "target", "t", "all", "Target agent tool: claude, antigravity, codex, copilot, all")
	mcpCmd.AddCommand(mcpServeCmd)
	mcpCmd.AddCommand(mcpSetupCmd)

	return mcpCmd
}
