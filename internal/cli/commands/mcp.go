package commands

import (
	"context"
	"fmt"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol (MCP) server commands",
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

func init() {
	mcpCmd.AddCommand(mcpServeCmd)
	cli.RootCmd.AddCommand(mcpCmd)
}
