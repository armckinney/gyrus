package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/spf13/cobra"
)

// hookDeadline bounds a hook lookup so the prompt is never held for long; the hook's own timeout is the backstop.
const hookDeadline = 4 * time.Second

// hookInput is the part of a Claude Code UserPromptSubmit payload that the hook reads.
type hookInput struct {
	Prompt string `json:"prompt"`
	Cwd    string `json:"cwd"`
}

// hookOutput is a UserPromptSubmit hook result that adds context beside the prompt.
type hookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// NewSuggestContextCmd constructs the 'suggest-context' Cobra command.
func NewSuggestContextCmd(application *app.App) *cobra.Command {
	var (
		prompt      string
		maxTokens   int
		scope       string
		project     string
		category    string
		maxConcepts int
		hook        bool
	)

	cmd := &cobra.Command{
		Use:   "suggest-context",
		Short: "Suggest and linearize relevant concept context for an agent prompt",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if hook {
				// A hook never fails the prompt: any problem is reported on stderr and the prompt goes on without context.
				if err := runPromptHook(cmd.InOrStdin(), cmd.OutOrStdout(), application, filter, maxConcepts, maxTokens); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "gyrus: suggest-context hook: %v\n", err)
				}
				return nil
			}
			if prompt == "" {
				return errors.New(`required flag(s) "prompt" not set`)
			}

			engine, err := application.Engine()
			if err != nil {
				return err
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

	cmd.Flags().StringVar(&prompt, "prompt", "", "Prompt context or task description (required unless --hook)")
	cmd.Flags().StringVar(&scope, "scope", "", "Context scope filter (project, reference, all)")
	cmd.Flags().StringVar(&project, "project", "", "Explicit project name to scope context to")
	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().IntVar(&maxConcepts, "max-concepts", 5, "Maximum number of concepts to include")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 4000, "Token budget, enforced: the layer holds the best-matching sections whose estimate fits; 0 returns whole concepts; a budget that fits no paragraph returns only the header, and one below the header's own size (about 13 tokens) returns nothing")
	cmd.Flags().BoolVar(&hook, "hook", false, "Run as a Claude Code UserPromptSubmit hook: read the hook payload on stdin, resolve config from its cwd, print the context as hook JSON (nothing when no concept matches)")

	return cmd
}

// runPromptHook answers one UserPromptSubmit hook call.
func runPromptHook(in io.Reader, out io.Writer, application *app.App, filter gyrus.SearchFilter, maxConcepts, maxTokens int) error {
	var payload hookInput
	if err := json.NewDecoder(in).Decode(&payload); err != nil {
		return fmt.Errorf("reading hook payload: %w", err)
	}
	if strings.TrimSpace(payload.Prompt) == "" {
		return nil
	}
	// The hook may start outside the session's directory; configuration follows the session unless pinned.
	if payload.Cwd != "" && os.Getenv("GYRUS_CONFIG") == "" {
		if err := os.Chdir(payload.Cwd); err != nil {
			return err
		}
		if err := application.Reset(); err != nil {
			return err
		}
	}
	engine, err := application.Engine()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), hookDeadline)
	defer cancel()
	layer, err := engine.SuggestContextWithBudget(ctx, payload.Prompt, filter, maxConcepts, maxTokens)
	if err != nil {
		return err
	}
	if !strings.Contains(layer, "\n--- [") {
		return nil
	}
	var result hookOutput
	result.HookSpecificOutput.HookEventName = "UserPromptSubmit"
	result.HookSpecificOutput.AdditionalContext = layer
	return json.NewEncoder(out).Encode(result)
}
