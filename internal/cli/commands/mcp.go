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
		project string
		cfgPath string
	)

	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Model Context Protocol (MCP) server and integration commands",
	}

	mcpServeCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start embedded Gyrus MCP stdio server",
		RunE: func(cmd *cobra.Command, args []string) error {
			var selected *app.App
			var err error
			if cfgPath != "" {
				selected, err = app.NewWithConfig(cfgPath)
			} else if project != "" {
				selected, err = app.NewWithProject(project)
			} else {
				selected, err = app.New()
			}
			if err != nil {
				return fmt.Errorf("failed to start MCP server: %w", err)
			}
			if cmd.Flags().Changed("git-branch") {
				branch, _ := cmd.Flags().GetString("git-branch")
				if err := selectGitBranch(cmd.Context(), selected, branch); err != nil {
					return err
				}
			}
			srv, err := mcp.NewServerFromApp(selected)
			if err != nil {
				return err
			}
			return srv.ServeStdio(context.Background())
		},
	}

	mcpServeCmd.Flags().StringVarP(&project, "project", "p", "", "Project root directory for context resolution")
	mcpServeCmd.Flags().StringVarP(&cfgPath, "config", "c", "", "Explicit path to .gyrus.yaml configuration file")

	mcpCmd.AddCommand(mcpServeCmd)

	return mcpCmd
}
