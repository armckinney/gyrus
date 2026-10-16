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
		queryStr    string
		category    string
		conceptType string
		status      string
		tag         string
		tenant      string
		scope       string
		project     string
		maxResults  int
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Execute search query over OKF concepts and metadata",
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
			effectiveProject := ""
			if cmd.Flags().Changed("project") {
				effectiveProject, _ = cmd.Flags().GetString("project")
			}

			filter := gyrus.SearchFilter{
				Category: gyrus.Category(category),
				Type:     gyrus.ConceptType(conceptType),
				Status:   status,
				Tag:      tag,
				Tenant:   tenant,
				Scope:    effectiveScope,
				Project:  effectiveProject,
			}

			results, err := engine.SearchQuery(context.Background(), gyrus.SearchQuery{Query: queryStr, Filter: filter, MaxResults: maxResults})
			if err != nil {
				return err
			}

			if GlobalJSONOutput {
				data, _ := json.MarshalIndent(results, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("Search Results (%d matches):\n", len(results))
				for i, res := range results {
					doc := res.Concept
					scopeTag := doc.Scope
					if doc.Project != "" {
						scopeTag = "project:" + doc.Project
					}
					fmt.Printf("  %d. [%s] %s (%s, %s, status: %s, scope: %s, score: %.1f)\n",
						i+1, doc.Name, doc.Title, doc.Type, doc.Category, doc.Status, scopeTag, res.Score)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&queryStr, "query", "", "Lexical search query text")
	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().StringVar(&conceptType, "type", "", "Filter by concept type")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag")
	cmd.Flags().StringVar(&tenant, "tenant", "", "Filter by tenant")
	cmd.Flags().StringVar(&scope, "scope", "", "Filter by scope (project, reference, all)")
	cmd.Flags().StringVar(&project, "project", "", "Filter by project name")
	cmd.Flags().IntVar(&maxResults, "max-results", 0, "Maximum results (default: 10 for keyword search, unlimited for empty-query browsing)")

	return cmd
}
