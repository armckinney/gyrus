package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/spf13/cobra"
)

// NewUpdateCmd constructs the 'update' Cobra command.
func NewUpdateCmd(application *app.App) *cobra.Command {
	var (
		title           string
		status          string
		tags            string
		dependencies    string
		content         string
		contentFile     string
		expectedVersion int
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update existing document fields or body content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			patch := gyrus.DocumentPatch{}

			if cmd.Flags().Changed("title") {
				patch.Title = &title
			}
			if cmd.Flags().Changed("status") {
				patch.Status = &status
			}
			if cmd.Flags().Changed("tags") {
				tList := strings.Split(tags, ",")
				patch.Tags = &tList
			}
			if cmd.Flags().Changed("dependencies") {
				dList := strings.Split(dependencies, ",")
				patch.Dependencies = &dList
			}

			if contentFile != "" {
				data, err := os.ReadFile(contentFile)
				if err != nil {
					return fmt.Errorf("failed reading content file '%s': %w", contentFile, err)
				}
				bodyStr := string(data)
				patch.Content = &bodyStr
			} else if cmd.Flags().Changed("content") {
				patch.Content = &content
			}

			ref, err := engine.Update(context.Background(), id, patch, expectedVersion)
			if err != nil {
				return err
			}

			if GlobalJSONOutput {
				data, _ := json.MarshalIndent(ref, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("Updated document '%s' (v%d, %s)\n", ref.ID, ref.Version, ref.Status)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New Document Title")
	cmd.Flags().StringVar(&status, "status", "", "New Document Status")
	cmd.Flags().StringVar(&tags, "tags", "", "New comma-separated list of tags")
	cmd.Flags().StringVar(&dependencies, "dependencies", "", "New comma-separated list of dependent IDs")
	cmd.Flags().StringVar(&content, "content", "", "New inline body content")
	cmd.Flags().StringVar(&contentFile, "content-file", "", "Path to file containing new body content")
	cmd.Flags().IntVar(&expectedVersion, "expected-version", 0, "Expected document version for optimistic concurrency check")

	return cmd
}
