package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/config"
	"github.com/armckinney/gyrus/internal/setup"
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
	configCmd.AddCommand(NewConfigInitCmd(application))
	return configCmd
}

// NewConfigInitCmd constructs the 'config init' Cobra subcommand.
func NewConfigInitCmd(application *app.App) *cobra.Command {
	var (
		profile    string
		ownerGroup string
		global     bool
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize .gyrus.yaml workspace or global configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if global {
				cfgPath, err := setup.WriteGlobalConfigFile(setup.Profile(profile), ownerGroup, force)
				if err != nil {
					return err
				}

				homeDir, _ := os.UserHomeDir()
				storageDir := filepath.Join(homeDir, ".gyrus")

				if !GlobalJSONOutput {
					fmt.Printf("🚀 Initialized Gyrus global configuration successfully!\n")
					fmt.Printf("   - Configuration: %s (Profile: %s)\n", cfgPath, profile)
					fmt.Printf("   - Storage Root:  %s\n", storageDir)
					fmt.Printf("   - Default Owner: %s\n", ownerGroup)
				} else {
					fmt.Printf("{\"status\":\"configured\",\"config_file\":\"%s\",\"profile\":\"%s\",\"storage\":\"%s\",\"owner_group\":\"%s\",\"scope\":\"global\"}\n",
						cfgPath, profile, storageDir, ownerGroup)
				}
				return nil
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			res, err := setup.RunConfigSetup(setup.ConfigSetupOptions{
				WorkspaceDir: cwd,
				Profile:      setup.Profile(profile),
				OwnerGroup:   ownerGroup,
			})
			if err != nil {
				return err
			}

			if !GlobalJSONOutput {
				fmt.Printf("🚀 Initialized Gyrus workspace configuration successfully!\n")
				fmt.Printf("   - Configuration: %s (Profile: %s)\n", res.ConfigFile, profile)
				fmt.Printf("   - Storage Root:  %s\n", res.StorageDir)
				fmt.Printf("   - Default Owner: %s\n", res.OwnerGroup)
			} else {
				fmt.Printf("{\"status\":\"configured\",\"config_file\":\"%s\",\"profile\":\"%s\",\"storage\":\"%s\",\"owner_group\":\"%s\",\"scope\":\"workspace\"}\n",
					res.ConfigFile, profile, res.StorageDir, res.OwnerGroup)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&profile, "profile", "p", "local", "Configuration profile: local, git, blob, s3, azure, gcs, postgres, vector")
	cmd.Flags().StringVarP(&ownerGroup, "owner-group", "o", "root", "Default owner group for context documents")
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Write configuration to global ~/.gyrus.yaml instead of workspace")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration file without prompting")

	return cmd
}
