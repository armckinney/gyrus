package commands

import (
	"fmt"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/okf"
	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
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

		if !cli.GlobalJSONOutput {
			fmt.Print(templateContent)
		}

		return nil
	},
}

func init() {
	cli.RootCmd.AddCommand(schemaCmd)
}
