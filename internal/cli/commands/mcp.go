package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/mcp"
	"github.com/armckinney/gyrus/internal/setup"
	"github.com/spf13/cobra"
)

var mcpTarget string

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol (MCP) server and integration commands",
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start embedded Gyrus MCP stdio server",
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, err := mcp.NewServer(cli.GlobalStoragePath)
		if err != nil {
			return fmt.Errorf("failed to start MCP server: %w", err)
		}
		return srv.ServeStdio(context.Background())
	},
}

var mcpSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Register Gyrus stdio MCP server in agent tools (Claude, Antigravity, Codex, Copilot)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		files, err := setup.RegisterMCPServer(cwd, setup.MCPTarget(mcpTarget), "gyrus")
		if err != nil {
			return err
		}

		if !cli.GlobalJSONOutput {
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

func init() {
	mcpSetupCmd.Flags().StringVarP(&mcpTarget, "target", "t", "all", "Target agent tool: claude, antigravity, codex, copilot, all")
	mcpCmd.AddCommand(mcpServeCmd)
	mcpCmd.AddCommand(mcpSetupCmd)
	cli.RootCmd.AddCommand(mcpCmd)
}
