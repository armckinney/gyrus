package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/okf"
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

// NewGetCmd constructs the 'get' Cobra command.
func NewGetCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get document envelope or Markdown payload by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			doc, err := engine.Get(context.Background(), id)
			if err != nil {
				return err
			}

			if GlobalJSONOutput {
				data, err := json.MarshalIndent(doc, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			} else {
				mdData, err := okf.SerializeMarkdown(&doc)
				if err != nil {
					return err
				}
				fmt.Print(string(mdData))
			}

			return nil
		},
	}
}

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

// NewArchiveCmd constructs the 'archive' Cobra command.
func NewArchiveCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <document-id>",
		Short: "Archive (delete) a document from storage and search index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			if err := engine.Archive(context.Background(), id); err != nil {
				return err
			}

			if GlobalJSONOutput {
				fmt.Printf("{\"archived\": true, \"id\": \"%s\"}\n", id)
			} else {
				fmt.Printf("Archived document '%s' from storage and index\n", id)
			}

			return nil
		},
	}
}
