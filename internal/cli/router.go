package cli

import (
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/cli/commands"
	"github.com/spf13/cobra"
)

// BuildRootCmd constructs the root Cobra command and explicitly registers subcommands.
func BuildRootCmd(storagePath string) (*cobra.Command, error) {
	application, err := app.New(storagePath)
	if err != nil {
		return nil, err
	}

	rootCmd := &cobra.Command{
		Use:   "gyrus",
		Short: "Gyrus: Unified Context & Memory Engine",
		Long:  "Gyrus is a high-performance local-first memory and context engine for software development teams and AI agents.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if commands.GlobalStoragePath != "" {
				return application.Reset(commands.GlobalStoragePath)
			}
			return nil
		},
	}

	commands.Register(rootCmd, application)
	return rootCmd, nil
}

// Execute runs the Cobra CLI command router and exits with the appropriate programmatic code.
func Execute() {
	rootCmd, err := BuildRootCmd(commands.GlobalStoragePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing application: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		exitCode := MapErrorToExitCode(err)
		if !commands.GlobalJSONOutput {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(exitCode)
	}
}
