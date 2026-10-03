package commands

import (
	"context"
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/spf13/cobra"
)

// NewSuggestContextCmd constructs the 'suggest-context' Cobra command.
func NewSuggestContextCmd(application *app.App) *cobra.Command {
	var (
		prompt    string
		maxTokens int
		scope     string
		workspace string
		category  string
		maxDocs   int
	)

	cmd := &cobra.Command{
		Use:   "suggest-context",
		Short: "Suggest and linearize relevant document context for an agent prompt",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			if maxDocs <= 0 {
				maxDocs = 5
			}
			_ = maxTokens

			effectiveScope := ""
			if cmd.Flags().Changed("scope") {
				effectiveScope, _ = cmd.Flags().GetString("scope")
			}
			effectiveWorkspace := ""
			if cmd.Flags().Changed("workspace") {
				effectiveWorkspace, _ = cmd.Flags().GetString("workspace")
			}

			filter := gyrus.SearchFilter{
				Category:  gyrus.Category(category),
				Scope:     effectiveScope,
				Workspace: effectiveWorkspace,
			}

			contextLayer, err := engine.SuggestContextWithFilter(context.Background(), prompt, filter, maxDocs)
			if err != nil {
				return err
			}

			if GlobalJSONOutput {
				fmt.Printf("{\"context\": %q}\n", contextLayer)
			} else {
				fmt.Print(contextLayer)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&prompt, "prompt", "", "Prompt context or task description (required)")
	cmd.Flags().StringVar(&scope, "scope", "", "Context scope filter (workspace, reference, all)")
	cmd.Flags().StringVar(&workspace, "workspace", "", "Explicit workspace name to scope context to")
	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().IntVar(&maxDocs, "max-docs", 5, "Maximum number of documents to include")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 4000, "Maximum token context budget")
	_ = cmd.MarkFlagRequired("prompt")

	return cmd
}
