package cli

import (
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/spf13/cobra"
)

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
