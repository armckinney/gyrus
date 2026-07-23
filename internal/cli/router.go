package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	GlobalStoragePath string
	GlobalJSONOutput  bool
	GlobalVerbose     bool
)

// RootCmd is the top-level Cobra command for gyrus.
var RootCmd = &cobra.Command{
	Use:   "gyrus",
	Short: "Gyrus: Unified Context & Memory Engine",
	Long:  "Gyrus is a high-performance local-first memory and context engine for software development teams and AI agents.",
}

func init() {
	RootCmd.PersistentFlags().StringVar(&GlobalStoragePath, "storage-path", "", "Path to storage root directory (overrides GYRUS_STORAGE_PATH env)")
	RootCmd.PersistentFlags().BoolVar(&GlobalJSONOutput, "json", false, "Output results as formatted JSON")
	RootCmd.PersistentFlags().BoolVar(&GlobalVerbose, "verbose", false, "Enable verbose debug logging")
}

// Execute runs the Cobra CLI command router and exits with the appropriate programmatic code.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		exitCode := MapErrorToExitCode(err)
		if !GlobalJSONOutput {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(exitCode)
	}
}
