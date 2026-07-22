package commands

import (
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema <doc-type>",
	Short: "Print frontmatter schema and template for a document type",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docType := args[0]
		templatePath := fmt.Sprintf("docs/specs/doc-types/%s.md", docType)

		data, err := os.ReadFile(templatePath)
		if err != nil {
			// Print default template fallback
			fallback := fmt.Sprintf(`---
id: %s-001
title: Example Title
category: architecture
type: %s
owner_group: platform
version: 1
status: draft
tags:
  - example
dependencies: []
---

# Title

Document description body here.
`, docType, docType)
			fmt.Print(fallback)
			return nil
		}

		if !cli.GlobalJSONOutput {
			fmt.Print(string(data))
		}

		return nil
	},
}

func init() {
	cli.RootCmd.AddCommand(schemaCmd)
}
