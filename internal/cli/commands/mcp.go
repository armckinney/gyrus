package commands

import (
	"context"
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/mcp"
	"github.com/spf13/cobra"
)

// NewMCPCmd constructs the 'mcp' parent Cobra command and its subcommands.
func NewMCPCmd(application *app.App) *cobra.Command {
	var (
		workspace string
		cfgPath   string
	)

	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol (MCP) server and integration commands",
	}

	mcpServeCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start embedded Gyrus MCP stdio server",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				srv *mcp.Server
				err error
			)
			if cfgPath != "" {
				srv, err = mcp.NewServerWithConfig(cfgPath)
			} else if workspace != "" {
				srv, err = mcp.NewServerWithWorkspace(workspace)
			} else {
				srv, err = mcp.NewServer()
			}
			if err != nil {
				return fmt.Errorf("failed to start MCP server: %w", err)
			}
			return srv.ServeStdio(context.Background())
		},
	}

	mcpServeCmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace root directory for context resolution")
	mcpServeCmd.Flags().StringVarP(&cfgPath, "config", "c", "", "Explicit path to .gyrus.yaml configuration file")

	mcpCmd.AddCommand(mcpServeCmd)

	return mcpCmd
}
