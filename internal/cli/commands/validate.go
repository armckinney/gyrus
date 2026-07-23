package commands

import (
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/okf"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
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
			// Try JSON parse fallback
			var jsonErr error
			doc, jsonErr = okf.ParseJSON(data)
			if jsonErr != nil {
				return fmt.Errorf("failed parsing OKF document: %w", err)
			}
		}

		if err := okf.Validate(doc); err != nil {
			return err
		}

		if !cli.GlobalJSONOutput {
			fmt.Printf("✓ Validation successful for OKF document '%s' (type: %s, category: %s)\n", doc.ID, doc.Type, doc.Category)
		}

		return nil
	},
}

func init() {
	cli.RootCmd.AddCommand(validateCmd)
}
