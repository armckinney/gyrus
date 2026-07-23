package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/internal/provider/sqlite"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/spf13/cobra"
)

var (
	suggestPrompt    string
	suggestMaxTokens int
)

var suggestCmd = &cobra.Command{
	Use:   "suggest-context",
	Short: "Suggest and linearize relevant document context for an agent prompt",
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

		q := gyrus.SearchQuery{
			Query:      suggestPrompt,
			MaxResults: 5,
		}

		results, err := indexer.Search(context.Background(), q)
		if err != nil {
			return err
		}

		store, err := localfs.NewStore(storageRoot)
		if err != nil {
			return err
		}

		var contextDocs []gyrus.Document
		for _, res := range results {
			if doc, err := store.Get(context.Background(), res.Document.ID); err == nil {
				contextDocs = append(contextDocs, doc)
			}
		}

		if cli.GlobalJSONOutput {
			data, _ := json.MarshalIndent(contextDocs, "", "  ")
			fmt.Println(string(data))
		} else {
			var sb strings.Builder
			sb.WriteString("=== SUGGESTED CONTEXT LAYER ===\n\n")
			for _, doc := range contextDocs {
				sb.WriteString(fmt.Sprintf("--- DOCUMENT: %s (%s) ---\n", doc.ID, doc.Title))
				sb.WriteString(doc.Content)
				sb.WriteString("\n\n")
			}
			fmt.Print(sb.String())
		}

		return nil
	},
}

func init() {
	suggestCmd.Flags().StringVar(&suggestPrompt, "prompt", "", "Prompt context or task description (required)")
	suggestCmd.Flags().IntVar(&suggestMaxTokens, "max-tokens", 4000, "Maximum token context budget")

	_ = suggestCmd.MarkFlagRequired("prompt")

	cli.RootCmd.AddCommand(suggestCmd)
}
