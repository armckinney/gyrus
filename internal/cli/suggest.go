package cli

import (
	"context"
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/spf13/cobra"
)

// NewSuggestContextCmd constructs the 'suggest-context' Cobra command.
func NewSuggestContextCmd(application *app.App) *cobra.Command {
	var (
		prompt    string
		maxTokens int
	)

	cmd := &cobra.Command{
		Use:   "suggest-context",
		Short: "Suggest and linearize relevant document context for an agent prompt",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			_ = maxTokens
			contextLayer, err := engine.SuggestContext(context.Background(), prompt, "", 5)
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
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 4000, "Maximum token context budget")
	_ = cmd.MarkFlagRequired("prompt")

	return cmd
}
