package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/ui"
	"github.com/spf13/cobra"
)

// NewUICmd constructs the 'ui' parent Cobra command and its subcommands.
func NewUICmd(application *app.App) *cobra.Command {
	var (
		host      string
		port      int
		workspace string
		cfgPath   string
	)

	uiCmd := &cobra.Command{
		Use:   "ui",
		Short: "Web UI and interactive visualization commands",
	}

	uiServeCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start embedded Gyrus Web UI server",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				srv *ui.Server
				err error
			)
			if cfgPath != "" {
				srv, err = ui.NewServerWithConfig(cfgPath, host, port)
			} else if workspace != "" {
				srv, err = ui.NewServerWithWorkspace(workspace, host, port)
			} else {
				srv, err = ui.NewServer(host, port)
			}
			if err != nil {
				return fmt.Errorf("failed to initialize Web UI server: %w", err)
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			fmt.Fprintf(cmd.OutOrStdout(), "Gyrus Web UI listening on http://%s:%d\n", host, port)
			return srv.Start(ctx)
		},
	}

	uiServeCmd.Flags().StringVar(&host, "host", "127.0.0.1", "Host address to bind HTTP server")
	uiServeCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port number to listen on")
	uiServeCmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Workspace root directory for context resolution")
	uiServeCmd.Flags().StringVarP(&cfgPath, "config", "c", "", "Explicit path to .gyrus.yaml configuration file")

	uiCmd.AddCommand(uiServeCmd)

	return uiCmd
}
