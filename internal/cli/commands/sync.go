package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/spf13/cobra"
)

// NewSyncCmd constructs the 'sync' Cobra command.
func NewSyncCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Re-index filesystem documents and extract dependency edges",
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			report, err := engine.Sync(context.Background())
			if err != nil {
				return err
			}

			if GlobalJSONOutput {
				data, _ := json.MarshalIndent(report, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("Sync Completed: %d scanned, %d indexed, %d unchanged, %d removed\n",
					report.ScannedFiles, report.IndexedFiles, report.UnchangedFiles, report.RemovedFiles)
			}

			return nil
		},
	}
}

// NewValidateCmd constructs the 'validate' Cobra command.
func NewValidateCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "validate <file-path>",
		Short: "Validate an OKF Markdown file or JSON envelope schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed reading file '%s': %w", filePath, err)
			}

			doc, err := okf.ParseMarkdown(data)
			if err != nil {
				var jsonErr error
				doc, jsonErr = okf.ParseJSON(data)
				if jsonErr != nil {
					return fmt.Errorf("failed parsing OKF document: %w", err)
				}
			}

			if err := okf.Validate(doc); err != nil {
				return err
			}

			if !GlobalJSONOutput {
				fmt.Printf("✓ Validation successful for OKF document '%s' (type: %s, category: %s)\n", doc.ID, doc.Type, doc.Category)
			}

			return nil
		},
	}
}

// NewSchemaCmd constructs the 'schema' Cobra command.
func NewSchemaCmd(application *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "schema <doc-type>",
		Short: "Print frontmatter schema and template for a document type",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			docType := args[0]
			customSchemasDir, _ := localfs.ResolveSchemasPath()

			templateContent, err := okf.GetTemplate(docType, customSchemasDir)
			if err != nil {
				return fmt.Errorf("failed retrieving template for type '%s': %w", docType, err)
			}

			if !GlobalJSONOutput {
				fmt.Print(templateContent)
			}

			return nil
		},
	}
}
