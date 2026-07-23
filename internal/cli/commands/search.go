package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/internal/provider/sqlite"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/spf13/cobra"
)

var (
	searchQueryStr   string
	searchCategory   string
	searchType       string
	searchStatus     string
	searchTag        string
	searchOwnerGroup string
	searchMaxResults int
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Execute FTS5 search query over OKF documents and metadata",
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
			Query: searchQueryStr,
			Filter: gyrus.SearchFilter{
				Category:   gyrus.Category(searchCategory),
				Type:       gyrus.DocumentType(searchType),
				Status:     searchStatus,
				Tag:        searchTag,
				OwnerGroup: searchOwnerGroup,
			},
			MaxResults: searchMaxResults,
		}

		results, err := indexer.Search(context.Background(), q)
		if err != nil {
			return err
		}

		if cli.GlobalJSONOutput {
			data, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Search Results (%d matches):\n", len(results))
			for i, res := range results {
				doc := res.Document
				fmt.Printf("  %d. [%s] %s (%s, %s, status: %s)\n",
					i+1, doc.ID, doc.Title, doc.Type, doc.Category, doc.Status)
			}
		}

		return nil
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchQueryStr, "query", "", "Lexical search query text")
	searchCmd.Flags().StringVar(&searchCategory, "category", "", "Filter by category")
	searchCmd.Flags().StringVar(&searchType, "type", "", "Filter by document type")
	searchCmd.Flags().StringVar(&searchStatus, "status", "", "Filter by status")
	searchCmd.Flags().StringVar(&searchTag, "tag", "", "Filter by tag")
	searchCmd.Flags().StringVar(&searchOwnerGroup, "owner-group", "", "Filter by owner group")
	searchCmd.Flags().IntVar(&searchMaxResults, "max-results", 10, "Maximum number of results to return")

	cli.RootCmd.AddCommand(searchCmd)
}
