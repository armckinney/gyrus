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
		prompt      string
		maxTokens   int
		scope       string
		project     string
		category    string
		maxConcepts int
	)

	cmd := &cobra.Command{
		Use:   "suggest-context",
		Short: "Suggest and linearize relevant concept context for an agent prompt",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			if maxConcepts <= 0 {
				maxConcepts = 5
			}
			if maxTokens < 0 {
				return fmt.Errorf("--max-tokens must be 0 (whole concepts) or more, got %d", maxTokens)
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
				Scope:    effectiveScope,
				Project:  effectiveProject,
			}

			contextLayer, err := engine.SuggestContextWithBudget(context.Background(), prompt, filter, maxConcepts, maxTokens)
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
	cmd.Flags().StringVar(&scope, "scope", "", "Context scope filter (project, reference, all)")
	cmd.Flags().StringVar(&project, "project", "", "Explicit project name to scope context to")
	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().IntVar(&maxConcepts, "max-concepts", 5, "Maximum number of concepts to include")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 4000, "Token budget, enforced: the layer holds the best-matching sections whose estimate fits; 0 returns whole concepts; a budget that fits no paragraph returns only the header, and one below the header's own size (about 13 tokens) returns nothing")
	_ = cmd.MarkFlagRequired("prompt")

	return cmd
}
