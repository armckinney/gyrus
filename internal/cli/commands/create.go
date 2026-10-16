package commands

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
		conceptType  string
		tenant       string
		status       string
		tags         string
		dependencies string
		content      string
		contentFile  string
		scope        string
		project      string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new OKF contract concept",
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
				if conceptType == string(gyrus.TypeADR) {
					docStatus = "proposed"
				}
			}

			docTenant := tenant
			if docTenant == "" {
				cfg := application.Config()
				if cfg != nil && cfg.Tenant != "" {
					docTenant = cfg.Tenant
				} else {
					docTenant = "root"
				}
			}

			effectiveScope := ""
			if cmd.Flags().Changed("scope") {
				effectiveScope, _ = cmd.Flags().GetString("scope")
			}
			effectiveProject := ""
			if cmd.Flags().Changed("project") {
				effectiveProject, _ = cmd.Flags().GetString("project")
			}

			doc := gyrus.Concept{
				ID:           id,
				Title:        title,
				Category:     gyrus.Category(category),
				Type:         gyrus.ConceptType(conceptType),
				Tenant:       docTenant,
				Version:      1,
				Status:       docStatus,
				Tags:         tagsList,
				Dependencies: depsList,
				Scope:        effectiveScope,
				Project:      effectiveProject,
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
				fmt.Printf("Created concept '%s' (v%d, %s)\n", ref.ID, ref.Version, ref.Status)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Concept ID (required, pattern ^[a-z0-9-_]+$)")
	cmd.Flags().StringVar(&title, "title", "", "Concept Title (required)")
	cmd.Flags().StringVar(&category, "category", "", "Concept Category (architecture|business-logic|product|operations|technical)")
	cmd.Flags().StringVar(&conceptType, "type", "", "Concept Type (adr|prd|guide|specification|...)")
	cmd.Flags().StringVar(&tenant, "tenant", "", "Tenant (defaults to configured tenant or root)")
	cmd.Flags().StringVar(&status, "status", "", "Concept Status (draft|proposed|active)")
	cmd.Flags().StringVar(&tags, "tags", "", "Comma-separated list of tags")
	cmd.Flags().StringVar(&dependencies, "dependencies", "", "Comma-separated list of dependent concept IDs")
	cmd.Flags().StringVar(&scope, "scope", "", "Target scope (project or reference)")
	cmd.Flags().StringVar(&project, "project", "", "Target project name (defaults to configured project)")
	cmd.Flags().StringVar(&content, "content", "", "Inline concept body content")
	cmd.Flags().StringVar(&contentFile, "content-file", "", "Path to file containing concept body content")

	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("title")
	_ = cmd.MarkFlagRequired("category")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}
