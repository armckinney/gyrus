package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/spf13/cobra"
)

var (
	updateTitle           string
	updateStatus          string
	updateTags            string
	updateDependencies    string
	updateContent         string
	updateContentFile     string
	updateExpectedVersion int
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update existing document fields or body content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		storageRoot, err := localfs.ResolveStoragePath(cli.GlobalStoragePath)
		if err != nil {
			return err
		}

		store, err := localfs.NewStore(storageRoot)
		if err != nil {
			return err
		}

		patch := gyrus.DocumentPatch{}

		if cmd.Flags().Changed("title") {
			patch.Title = &updateTitle
		}
		if cmd.Flags().Changed("status") {
			patch.Status = &updateStatus
		}
		if cmd.Flags().Changed("tags") {
			tList := strings.Split(updateTags, ",")
			patch.Tags = &tList
		}
		if cmd.Flags().Changed("dependencies") {
			dList := strings.Split(updateDependencies, ",")
			patch.Dependencies = &dList
		}

		if updateContentFile != "" {
			data, err := os.ReadFile(updateContentFile)
			if err != nil {
				return fmt.Errorf("failed reading content file '%s': %w", updateContentFile, err)
			}
			bodyStr := string(data)
			patch.Content = &bodyStr
		} else if cmd.Flags().Changed("content") {
			patch.Content = &updateContent
		}

		ref, err := store.Update(context.Background(), id, patch, updateExpectedVersion)
		if err != nil {
			return err
		}

		if cli.GlobalJSONOutput {
			data, _ := json.MarshalIndent(ref, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Updated document '%s' (v%d, %s)\n", ref.ID, ref.Version, ref.Status)
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New Document Title")
	updateCmd.Flags().StringVar(&updateStatus, "status", "", "New Document Status")
	updateCmd.Flags().StringVar(&updateTags, "tags", "", "New comma-separated list of tags")
	updateCmd.Flags().StringVar(&updateDependencies, "dependencies", "", "New comma-separated list of dependent IDs")
	updateCmd.Flags().StringVar(&updateContent, "content", "", "New inline body content")
	updateCmd.Flags().StringVar(&updateContentFile, "content-file", "", "Path to file containing new body content")
	updateCmd.Flags().IntVar(&updateExpectedVersion, "expected-version", 0, "Expected document version for optimistic concurrency check")

	cli.RootCmd.AddCommand(updateCmd)
}
