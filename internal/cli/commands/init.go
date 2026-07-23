package commands

import (
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Gyrus storage and config in local workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, err := localfs.ResolveStoragePath(cli.GlobalStoragePath)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create storage root directory: %w", err)
		}

		if !cli.GlobalJSONOutput {
			fmt.Printf("Initialized Gyrus storage at: %s\n", targetDir)
		}
		return nil
	},
}

func init() {
	cli.RootCmd.AddCommand(initCmd)
}
