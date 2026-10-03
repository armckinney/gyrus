package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
		scope      string
		workspace  string
		maxResults int
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Execute search query over OKF documents and metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			if queryStr == "" && len(args) > 0 {
				queryStr = strings.Join(args, " ")
			}

			engine, err := application.Engine()
			if err != nil {
				return err
			}

			effectiveScope := ""
			if cmd.Flags().Changed("scope") {
				effectiveScope, _ = cmd.Flags().GetString("scope")
			}
			effectiveWorkspace := ""
			if cmd.Flags().Changed("workspace") {
				effectiveWorkspace, _ = cmd.Flags().GetString("workspace")
			}

			filter := gyrus.SearchFilter{
				Category:   gyrus.Category(category),
				Type:       gyrus.DocumentType(docType),
				Status:     status,
				Tag:        tag,
				OwnerGroup: ownerGroup,
				Scope:      effectiveScope,
				Workspace:  effectiveWorkspace,
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
					scopeTag := doc.Scope
					if doc.Workspace != "" {
						scopeTag = "workspace:" + doc.Workspace
					}
					fmt.Printf("  %d. [%s] %s (%s, %s, status: %s, scope: %s, score: %.1f)\n",
						i+1, doc.ID, doc.Title, doc.Type, doc.Category, doc.Status, scopeTag, res.Score)
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
	cmd.Flags().StringVar(&scope, "scope", "", "Filter by scope (workspace, reference, all)")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Filter by workspace name")
	cmd.Flags().IntVar(&maxResults, "max-results", 10, "Maximum number of results to return")

	return cmd
}
