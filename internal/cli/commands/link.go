package commands

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/internal/provider/sqlite"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/spf13/cobra"
)

var (
	linkRelType string
)

var linkCmd = &cobra.Command{
	Use:   "link <from-id> <to-id>",
	Short: "Create a directed relationship edge between two documents",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fromID, toID := args[0], args[1]
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

		relType := gyrus.RelationshipType(linkRelType)
		if relType == "" {
			relType = gyrus.RelDependsOn
		}

		edge := gyrus.DocumentEdge{
			FromDocumentID:   fromID,
			ToDocumentID:     toID,
			RelationshipType: relType,
			CreatedBy:        "cli",
			CreatedAt:        time.Now().Truncate(time.Second),
		}

		if err := indexer.UpsertEdges(context.Background(), []gyrus.DocumentEdge{edge}); err != nil {
			return err
		}

		if !cli.GlobalJSONOutput {
			fmt.Printf("Linked '%s' -[%s]-> '%s'\n", fromID, relType, toID)
		}

		return nil
	},
}

var unlinkCmd = &cobra.Command{
	Use:   "unlink <from-id> <to-id>",
	Short: "Remove a directed relationship edge between two documents",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fromID, toID := args[0], args[1]
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

		relType := gyrus.RelationshipType(linkRelType)
		if relType == "" {
			relType = gyrus.RelDependsOn
		}

		if err := indexer.DeleteEdges(context.Background(), fromID, toID, relType); err != nil {
			return err
		}

		if !cli.GlobalJSONOutput {
			fmt.Printf("Unlinked '%s' -[%s]-> '%s'\n", fromID, relType, toID)
		}

		return nil
	},
}

func init() {
	linkCmd.Flags().StringVar(&linkRelType, "rel-type", "depends_on", "Relationship type (depends_on|supersedes|implements|mitigates)")
	unlinkCmd.Flags().StringVar(&linkRelType, "rel-type", "depends_on", "Relationship type (depends_on|supersedes|implements|mitigates)")

	cli.RootCmd.AddCommand(linkCmd)
	cli.RootCmd.AddCommand(unlinkCmd)
}
