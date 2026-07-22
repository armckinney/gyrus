package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/internal/provider/sqlite"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive <document-id>",
	Short: "Archive (delete) a document from storage and search index",
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

		if err := store.Archive(context.Background(), id); err != nil {
			return err
		}

		dbPath := filepath.Join(storageRoot, "index.db")
		if indexer, err := sqlite.NewIndexer(dbPath); err == nil {
			_ = indexer.Remove(context.Background(), id)
			indexer.Close()
		}

		if cli.GlobalJSONOutput {
			fmt.Printf("{\"archived\": true, \"id\": \"%s\"}\n", id)
		} else {
			fmt.Printf("Archived document '%s' from storage and index\n", id)
		}

		return nil
	},
}

func init() {
	cli.RootCmd.AddCommand(archiveCmd)
}
