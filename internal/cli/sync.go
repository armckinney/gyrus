package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/spf13/cobra"
)

// NewSyncCmd constructs the 'sync' Cobra command.
func NewSyncCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Re-index filesystem documents and extract dependency edges",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			report, err := engine.Sync(context.Background())
			if err != nil {
				return err
			}

			if GlobalJSONOutput {
				data, _ := json.MarshalIndent(report, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("Sync Completed: %d scanned, %d indexed, %d unchanged, %d removed\n",
					report.ScannedFiles, report.IndexedFiles, report.UnchangedFiles, report.RemovedFiles)
			}

			return nil
		},
	}
}
