package commands

import (
	"context"
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/spf13/cobra"
)

// NewLinkCmd constructs the 'link' Cobra command.
func NewLinkCmd(application *app.App) *cobra.Command {
	var relTypeStr string

	cmd := &cobra.Command{
		Use:   "link <from-id> <to-id>",
		Short: "Create a directed relationship edge between two documents",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fromID, toID := args[0], args[1]
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			relType := gyrus.RelationshipType(relTypeStr)
			if relType == "" {
				relType = gyrus.RelDependsOn
			}

			if err := engine.Link(context.Background(), fromID, toID, relType); err != nil {
				return err
			}

			if !GlobalJSONOutput {
				fmt.Printf("Linked '%s' -[%s]-> '%s'\n", fromID, relType, toID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&relTypeStr, "rel-type", "depends_on", "Relationship type (depends_on|supersedes|implements|mitigates)")
	return cmd
}

// NewUnlinkCmd constructs the 'unlink' Cobra command.
func NewUnlinkCmd(application *app.App) *cobra.Command {
	var relTypeStr string

	cmd := &cobra.Command{
		Use:   "unlink <from-id> <to-id>",
		Short: "Remove a directed relationship edge between two documents",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fromID, toID := args[0], args[1]
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			relType := gyrus.RelationshipType(relTypeStr)
			if relType == "" {
				relType = gyrus.RelDependsOn
			}

			if err := engine.Unlink(context.Background(), fromID, toID, relType); err != nil {
				return err
			}

			if !GlobalJSONOutput {
				fmt.Printf("Unlinked '%s' -[%s]-> '%s'\n", fromID, relType, toID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&relTypeStr, "rel-type", "depends_on", "Relationship type (depends_on|supersedes|implements|mitigates)")
	return cmd
}
