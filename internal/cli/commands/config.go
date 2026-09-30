package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewConfigCmd constructs the 'config' parent Cobra command and subcommands.
func NewConfigCmd(application *app.App) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
	}

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Display the resolved Gyrus configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			rc := application.ResolvedConfig()
			if rc == nil {
				return fmt.Errorf("configuration not resolved")
			}

			if GlobalJSONOutput {
				payload := map[string]interface{}{
					"source":       rc.Source,
					"source_path":  rc.SourcePath,
					"storage_root": rc.StorageRoot,
					"config":       rc.Config,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(payload)
			}

			fmt.Printf("# Gyrus Resolved Configuration\n")
			if rc.Source == config.SourceDefault {
				fmt.Printf("# Source: defaults (no .gyrus.yaml or ~/.gyrus.yaml found)\n")
			} else {
				fmt.Printf("# Source: %s", rc.Source)
				if rc.SourcePath != "" {
					fmt.Printf(" (%s)", rc.SourcePath)
				}
				fmt.Printf("\n")
			}
			fmt.Printf("# Resolved Storage Root: %s\n\n", rc.StorageRoot)

			enc := yaml.NewEncoder(os.Stdout)
			enc.SetIndent(2)
			return enc.Encode(rc.Config)
		},
	}

	configCmd.AddCommand(showCmd)
	return configCmd
}
