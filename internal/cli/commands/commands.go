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
	var gitBranch string
	rootCmd.PersistentFlags().StringVar(&gitBranch, "git-branch", "", "Runtime initiative branch (selects branching strategy; never saved)")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// MCP constructs its app from its own --config/--project flags.
		if cmd.Name() == "serve" && cmd.Parent() != nil && cmd.Parent().Name() == "mcp" {
			return nil
		}
		if cmd.Flags().Changed("git-branch") {
			return selectGitBranch(cmd.Context(), application, gitBranch)
		}
		return nil
	}
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

	// Project & Maintenance Commands
	rootCmd.AddCommand(NewValidateCmd(application))
	rootCmd.AddCommand(NewSchemaCmd(application))
	rootCmd.AddCommand(NewConfigCmd(application))
	rootCmd.AddCommand(NewGitCmd(application))

	// Setup & Server Commands
	rootCmd.AddCommand(NewMCPCmd(application))
	rootCmd.AddCommand(NewUICmd(application))
}
