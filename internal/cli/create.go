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

// NewCreateCmd constructs the 'create' Cobra command.
func NewCreateCmd(application *app.App) *cobra.Command {
	var (
		id           string
		title        string
		category     string
		docType      string
		ownerGroup   string
		status       string
		tags         string
		dependencies string
		content      string
		contentFile  string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new OKF contract document",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			body := content
			if contentFile != "" {
				data, err := os.ReadFile(contentFile)
				if err != nil {
					return fmt.Errorf("failed reading content file '%s': %w", contentFile, err)
				}
				body = string(data)
			}

			var tagsList []string
			if tags != "" {
				tagsList = strings.Split(tags, ",")
			}

			var depsList []string
			if dependencies != "" {
				depsList = strings.Split(dependencies, ",")
			}

			docStatus := status
			if docStatus == "" {
				docStatus = "draft"
				if docType == string(gyrus.TypeADR) {
					docStatus = "proposed"
				}
			}

			doc := gyrus.Document{
				ID:           id,
				Title:        title,
				Category:     gyrus.Category(category),
				Type:         gyrus.DocumentType(docType),
				OwnerGroup:   ownerGroup,
				Version:      1,
				Status:       docStatus,
				Tags:         tagsList,
				Dependencies: depsList,
				Content:      body,
			}

			ref, err := engine.Create(context.Background(), doc)
			if err != nil {
				return err
			}

			if GlobalJSONOutput {
				data, _ := json.MarshalIndent(ref, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("Created document '%s' (v%d, %s)\n", ref.ID, ref.Version, ref.Status)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Document ID (required, pattern ^[a-z0-9-_]+$)")
	cmd.Flags().StringVar(&title, "title", "", "Document Title (required)")
	cmd.Flags().StringVar(&category, "category", "", "Document Category (architecture|business-logic|product|operations|technical)")
	cmd.Flags().StringVar(&docType, "type", "", "Document Type (adr|prd|guide|specification|...)")
	cmd.Flags().StringVar(&ownerGroup, "owner-group", "", "Owner Group (required)")
	cmd.Flags().StringVar(&status, "status", "", "Document Status (draft|proposed|active)")
	cmd.Flags().StringVar(&tags, "tags", "", "Comma-separated list of tags")
	cmd.Flags().StringVar(&dependencies, "dependencies", "", "Comma-separated list of dependent document IDs")
	cmd.Flags().StringVar(&content, "content", "", "Inline document body content")
	cmd.Flags().StringVar(&contentFile, "content-file", "", "Path to file containing document body content")

	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("category")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("owner-group")

	return cmd
}
