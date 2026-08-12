package cli

import (
	"context"
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/spf13/cobra"
)

// NewArchiveCmd constructs the 'archive' Cobra command.
func NewArchiveCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <document-id>",
		Short: "Archive (delete) a document from storage and search index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			if err := engine.Archive(context.Background(), id); err != nil {
				return err
			}

			if GlobalJSONOutput {
				fmt.Printf("{\"archived\": true, \"id\": \"%s\"}\n", id)
			} else {
				fmt.Printf("Archived document '%s' from storage and index\n", id)
			}

			return nil
		},
	}
}
