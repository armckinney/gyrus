package cli

import (
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/spf13/cobra"
)

var (
	GlobalStoragePath string
	GlobalJSONOutput  bool
	GlobalVerbose     bool
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
			if GlobalStoragePath != "" {
				return application.Reset(GlobalStoragePath)
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&GlobalStoragePath, "storage-path", "", "Path to storage root directory (overrides GYRUS_STORAGE_PATH env)")
	rootCmd.PersistentFlags().BoolVar(&GlobalJSONOutput, "json", false, "Output results as formatted JSON")
	rootCmd.PersistentFlags().BoolVar(&GlobalVerbose, "verbose", false, "Enable verbose debug logging")

	rootCmd.AddCommand(NewCreateCmd(application))
	rootCmd.AddCommand(NewGetCmd(application))
	rootCmd.AddCommand(NewUpdateCmd(application))
	rootCmd.AddCommand(NewSearchCmd(application))
	rootCmd.AddCommand(NewSuggestContextCmd(application))
	rootCmd.AddCommand(NewLinkCmd(application))
	rootCmd.AddCommand(NewUnlinkCmd(application))
	rootCmd.AddCommand(NewSyncCmd(application))
	rootCmd.AddCommand(NewArchiveCmd(application))
	rootCmd.AddCommand(NewValidateCmd(application))
	rootCmd.AddCommand(NewSchemaCmd(application))
	rootCmd.AddCommand(NewInitCmd(application))
	rootCmd.AddCommand(NewMCPCmd(application))

	return rootCmd, nil
}

// Execute runs the Cobra CLI command router and exits with the appropriate programmatic code.
func Execute() {
	rootCmd, err := BuildRootCmd(GlobalStoragePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing application: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		exitCode := MapErrorToExitCode(err)
		if !GlobalJSONOutput {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(exitCode)
	}
}
