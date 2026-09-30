package commands

import (
	"github.com/armckinney/gyrus/internal/app"
	"github.com/spf13/cobra"
)

var (
	GlobalJSONOutput bool
	GlobalVerbose    bool
)

// Register registers all Gyrus CLI subcommands onto the root command.
func Register(rootCmd *cobra.Command, application *app.App) {
	rootCmd.PersistentFlags().BoolVar(&GlobalJSONOutput, "json", false, "Output results as formatted JSON")
	rootCmd.PersistentFlags().BoolVar(&GlobalVerbose, "verbose", false, "Enable verbose debug logging")

	// Document Lifecycle Commands
	rootCmd.AddCommand(NewCreateCmd(application))
	rootCmd.AddCommand(NewGetCmd(application))
	rootCmd.AddCommand(NewUpdateCmd(application))
	rootCmd.AddCommand(NewArchiveCmd(application))

	// Search & Context Commands
	rootCmd.AddCommand(NewSearchCmd(application))
	rootCmd.AddCommand(NewSuggestContextCmd(application))

	// Knowledge Graph Commands
	rootCmd.AddCommand(NewLinkCmd(application))
	rootCmd.AddCommand(NewUnlinkCmd(application))

	// Workspace & Maintenance Commands
	rootCmd.AddCommand(NewSyncCmd(application))
	rootCmd.AddCommand(NewValidateCmd(application))
	rootCmd.AddCommand(NewSchemaCmd(application))
	rootCmd.AddCommand(NewConfigCmd(application))

	// Setup & Server Commands
	rootCmd.AddCommand(NewInitCmd(application))
	rootCmd.AddCommand(NewMCPCmd(application))
}
