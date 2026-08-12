package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/spf13/cobra"
)

// NewGetCmd constructs the 'get' Cobra command.
func NewGetCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get document envelope or Markdown payload by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			doc, err := engine.Get(context.Background(), id)
			if err != nil {
				return err
			}

			if GlobalJSONOutput {
				data, err := json.MarshalIndent(doc, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			} else {
				mdData, err := okf.SerializeMarkdown(&doc)
				if err != nil {
					return err
				}
				fmt.Print(string(mdData))
			}

			return nil
		},
	}
}
