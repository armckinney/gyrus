package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	storagegit "github.com/armckinney/gyrus/internal/provider/storage/git"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
)

func NewGitCmd(application *app.App) *cobra.Command {
	cmd := &cobra.Command{Use: "git", Short: "Inspect this invocation's Git branch selection"}
	cmd.AddCommand(&cobra.Command{Use: "status", Args: cobra.NoArgs, Short: "Show strategy, trunk, and runtime branch", RunE: func(cmd *cobra.Command, args []string) error {
		store, err := gitStore(application)
		if err != nil {
			return err
		}
		status := store.Status()
		if GlobalJSONOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(status)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Strategy: %s\nTrunk branch: %s\nActive branch: %s\nSelection required: %t\n", status.Strategy, status.TrunkBranch, status.ActiveBranch, status.SelectionRequired)
		return nil
	}})
	return cmd
}

func gitStore(application *app.App) (*storagegit.Store, error) {
	if application.Config().StorageProvider() != "git" {
		return nil, fmt.Errorf("Git operations require Git storage")
	}
	store, err := application.ConceptStore()
	if err != nil {
		return nil, err
	}
	git, ok := store.(*storagegit.Store)
	if !ok {
		return nil, fmt.Errorf("Git operations require Git storage")
	}
	return git, nil
}

// Runtime selection is a Git session operation, not factory configuration.
func selectGitBranch(ctx context.Context, application *app.App, branch string) error {
	if application.Config().StorageProvider() != "git" {
		return fmt.Errorf("--git-branch requires Git storage")
	}
	if branch == "" {
		return fmt.Errorf("initiative branch is required")
	}
	if err := plumbing.NewBranchReferenceName(branch).Validate(); err != nil {
		return fmt.Errorf("invalid initiative branch: %w", err)
	}
	if branch == application.Config().Git.TrunkBranch {
		return fmt.Errorf("initiative branch must differ from trunk")
	}
	store, err := gitStore(application)
	if err != nil {
		return err
	}
	_, err = store.Switch(ctx, "branching", branch)
	return err
}
