package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/internal/provider/sqlite"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Re-index filesystem documents and extract dependency edges",
	RunE: func(cmd *cobra.Command, args []string) error {
		storageRoot, err := localfs.ResolveStoragePath(cli.GlobalStoragePath)
		if err != nil {
			return err
		}

		dbPath := filepath.Join(storageRoot, "index.db")
		indexer, err := sqlite.NewIndexer(dbPath)
		if err != nil {
			return err
		}
		defer indexer.Close()

		report, err := indexer.Sync(context.Background(), storageRoot)
		if err != nil {
			return err
		}

		if cli.GlobalJSONOutput {
			data, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Sync Completed: %d scanned, %d indexed, %d unchanged, %d removed\n",
				report.ScannedFiles, report.IndexedFiles, report.UnchangedFiles, report.RemovedFiles)
		}

		return nil
	},
}

func init() {
	cli.RootCmd.AddCommand(syncCmd)
}
