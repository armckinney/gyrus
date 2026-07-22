package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/okf"
	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get document envelope or Markdown payload by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		storageRoot, err := localfs.ResolveStoragePath(cli.GlobalStoragePath)
		if err != nil {
			return err
		}

		store, err := localfs.NewStore(storageRoot)
		if err != nil {
			return err
		}

		doc, err := store.Get(context.Background(), id)
		if err != nil {
			return err
		}

		if cli.GlobalJSONOutput {
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

func init() {
	cli.RootCmd.AddCommand(getCmd)
}
