package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/spf13/cobra"
)

// NewSearchCmd constructs the 'search' Cobra command.
func NewSearchCmd(application *app.App) *cobra.Command {
	var (
		queryStr   string
		category   string
		docType    string
		status     string
		tag        string
		ownerGroup string
		maxResults int
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Execute search query over OKF documents and metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			filter := gyrus.SearchFilter{
				Category:   gyrus.Category(category),
				Type:       gyrus.DocumentType(docType),
				Status:     status,
				Tag:        tag,
				OwnerGroup: ownerGroup,
			}

			results, err := engine.Search(context.Background(), queryStr, filter)
			if err != nil {
				return err
			}

			if maxResults > 0 && len(results) > maxResults {
				results = results[:maxResults]
			}

			if GlobalJSONOutput {
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

	cmd.Flags().StringVar(&queryStr, "query", "", "Lexical search query text")
	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().StringVar(&docType, "type", "", "Filter by document type")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag")
	cmd.Flags().StringVar(&ownerGroup, "owner-group", "", "Filter by owner group")
	cmd.Flags().IntVar(&maxResults, "max-results", 10, "Maximum number of results to return")

	return cmd
}
