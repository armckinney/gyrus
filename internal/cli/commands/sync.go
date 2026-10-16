package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/spf13/cobra"
)

// NewSyncCmd constructs the 'sync' Cobra command.
func NewSyncCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Scan concept storage and validate or rebuild its index",
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

			if len(report.Errors) > 0 {
				if !GlobalJSONOutput {
					for _, problem := range report.Errors {
						fmt.Fprintln(cmd.ErrOrStderr(), problem)
					}
				}
				return &okf.ValidationError{Field: "collection", Message: strings.Join(report.Errors, "; ")}
			}
			return nil
		},
	}
}
