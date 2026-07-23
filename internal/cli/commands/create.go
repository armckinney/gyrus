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
	createID           string
	createTitle        string
	createCategory     string
	createType         string
	createOwnerGroup   string
	createStatus       string
	createTags         string
	createDependencies string
	createContent      string
	createContentFile  string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new OKF contract document",
	RunE: func(cmd *cobra.Command, args []string) error {
		storageRoot, err := localfs.ResolveStoragePath(cli.GlobalStoragePath)
		if err != nil {
			return err
		}

		store, err := localfs.NewStore(storageRoot)
		if err != nil {
			return err
		}

		body := createContent
		if createContentFile != "" {
			data, err := os.ReadFile(createContentFile)
			if err != nil {
				return fmt.Errorf("failed reading content file '%s': %w", createContentFile, err)
			}
			body = string(data)
		}

		var tagsList []string
		if createTags != "" {
			tagsList = strings.Split(createTags, ",")
		}

		var depsList []string
		if createDependencies != "" {
			depsList = strings.Split(createDependencies, ",")
		}

		status := createStatus
		if status == "" {
			status = "draft"
			if createType == string(gyrus.TypeADR) {
				status = "proposed"
			}
		}

		doc := gyrus.Document{
			ID:           createID,
			Title:        createTitle,
			Category:     gyrus.Category(createCategory),
			Type:         gyrus.DocumentType(createType),
			OwnerGroup:   createOwnerGroup,
			Version:      1,
			Status:       status,
			Tags:         tagsList,
			Dependencies: depsList,
			Content:      body,
		}

		ref, err := store.Create(context.Background(), doc)
		if err != nil {
			return err
		}

		if cli.GlobalJSONOutput {
			data, _ := json.MarshalIndent(ref, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Created document '%s' (v%d, %s)\n", ref.ID, ref.Version, ref.Status)
		}

		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&createID, "id", "", "Document ID (required, pattern ^[a-z0-9-_]+$)")
	createCmd.Flags().StringVar(&createTitle, "title", "", "Document Title (required)")
	createCmd.Flags().StringVar(&createCategory, "category", "", "Document Category (architecture|business-logic|product|operations|technical)")
	createCmd.Flags().StringVar(&createType, "type", "", "Document Type (adr|prd|guide|specification|...)")
	createCmd.Flags().StringVar(&createOwnerGroup, "owner-group", "", "Owner Group (required)")
	createCmd.Flags().StringVar(&createStatus, "status", "", "Document Status (draft|proposed|active)")
	createCmd.Flags().StringVar(&createTags, "tags", "", "Comma-separated list of tags")
	createCmd.Flags().StringVar(&createDependencies, "dependencies", "", "Comma-separated list of dependent document IDs")
	createCmd.Flags().StringVar(&createContent, "content", "", "Inline document body content")
	createCmd.Flags().StringVar(&createContentFile, "content-file", "", "Path to file containing document body content")

	_ = createCmd.MarkFlagRequired("id")
	_ = createCmd.MarkFlagRequired("title")
	_ = createCmd.MarkFlagRequired("category")
	_ = createCmd.MarkFlagRequired("type")
	_ = createCmd.MarkFlagRequired("owner-group")

	cli.RootCmd.AddCommand(createCmd)
}
