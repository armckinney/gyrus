package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/spf13/cobra"
)

// NewSchemaCmd constructs the 'schema' Cobra command with subcommands:
// get, set, list, and delete, with backward compatibility for 'schema <doc-type>'.
func NewSchemaCmd(application *app.App) *cobra.Command {
	schemaCmd := &cobra.Command{
		Use:   "schema [subcommand | <doc-type>]",
		Short: "Manage and inspect OKF contract schema templates in persistence layer",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			// If called with a single argument that isn't a known subcommand,
			// treat it as 'schema get <doc-type>' for backward compatibility.
			docType := args[0]
			return runSchemaGet(cmd, application, docType)
		},
	}

	// 1. gyrus schema get <doc-type>
	getCmd := &cobra.Command{
		Use:   "get <doc-type>",
		Short: "Print or retrieve the schema template for a document type",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSchemaGet(cmd, application, args[0])
		},
	}

	// 2. gyrus schema set <doc-type> [--content <string> | --content-file <path>]
	var contentFlag string
	var contentFileFlag string

	setCmd := &cobra.Command{
		Use:     "set <doc-type> [file]",
		Aliases: []string{"save", "put"},
		Short:   "Create or update a schema template in the persistence layer",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			docType := args[0]
			var payload string

			filePath := contentFileFlag
			if filePath == "" && len(args) == 2 {
				filePath = args[1]
			}

			if filePath != "" {
				data, err := os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("failed reading schema content file '%s': %w", filePath, err)
				}
				payload = string(data)
			} else if contentFlag != "" {
				payload = contentFlag
			} else {
				return fmt.Errorf("must specify schema content via argument, --content, or --content-file")
			}

			engine, err := application.Engine()
			if err != nil {
				return err
			}

			if err := engine.SaveSchema(cmd.Context(), docType, payload); err != nil {
				return fmt.Errorf("failed saving schema for type '%s': %w", docType, err)
			}

			if GlobalJSONOutput {
				res := map[string]string{
					"status":   "saved",
					"type":     docType,
					"location": fmt.Sprintf(".gyrus/schemas/%s.md", docType),
				}
				data, _ := json.MarshalIndent(res, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Schema for '%s' saved to persistence layer (.gyrus/schemas/%s.md)\n", docType, docType)
			}

			return nil
		},
	}
	setCmd.Flags().StringVar(&contentFlag, "content", "", "Inline Markdown schema template content")
	setCmd.Flags().StringVar(&contentFileFlag, "content-file", "", "Path to Markdown file containing schema template")

	// 3. gyrus schema list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available schemas across persistence layer and embedded templates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			schemas, err := engine.ListSchemas(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed listing schemas: %w", err)
			}

			if GlobalJSONOutput {
				data, _ := json.MarshalIndent(schemas, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%-24s %-12s %s\n", "DOCUMENT TYPE", "SOURCE", "PERSISTED LOCATION")
				for _, s := range schemas {
					loc := s.Location
					if loc == "" {
						loc = "-"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%-24s %-12s %s\n", s.Type, s.Source, loc)
				}
			}

			return nil
		},
	}

	// 4. gyrus schema delete <doc-type>
	deleteCmd := &cobra.Command{
		Use:     "delete <doc-type>",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete a custom schema template from the persistence layer",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			docType := args[0]
			engine, err := application.Engine()
			if err != nil {
				return err
			}

			if err := engine.DeleteSchema(cmd.Context(), docType); err != nil {
				return fmt.Errorf("failed deleting schema for type '%s': %w", docType, err)
			}

			if GlobalJSONOutput {
				res := map[string]string{
					"status": "deleted",
					"type":   docType,
				}
				data, _ := json.MarshalIndent(res, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "✓ Schema for '%s' deleted from persistence layer\n", docType)
			}

			return nil
		},
	}

	schemaCmd.AddCommand(getCmd)
	schemaCmd.AddCommand(setCmd)
	schemaCmd.AddCommand(listCmd)
	schemaCmd.AddCommand(deleteCmd)

	return schemaCmd
}

func runSchemaGet(cmd *cobra.Command, application *app.App, docType string) error {
	engine, err := application.Engine()
	if err != nil {
		return err
	}

	templateContent, err := engine.GetSchema(cmd.Context(), docType)
	if err != nil {
		return fmt.Errorf("failed retrieving template for type '%s': %w", docType, err)
	}

	if GlobalJSONOutput {
		res := map[string]string{
			"type":     docType,
			"template": templateContent,
		}
		data, _ := json.MarshalIndent(res, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	} else {
		fmt.Fprint(cmd.OutOrStdout(), templateContent)
	}

	return nil
}
