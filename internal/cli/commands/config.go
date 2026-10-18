package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
	configCmd.AddCommand(NewConfigInitCmd(application))
	return configCmd
}

// NewConfigInitCmd constructs the 'config init' Cobra subcommand.
func NewConfigInitCmd(application *app.App) *cobra.Command {
	var (
		profile string
		tenant  string
		project string
		global  bool
		force   bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize .gyrus.yaml project or global configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if global {
				cfgPath, err := config.WriteGlobalConfigFile(config.Profile(profile), tenant, force, project)
				if err != nil {
					return err
				}

				homeDir, _ := os.UserHomeDir()
				storageDir := filepath.Join(homeDir, ".gyrus")

				if !GlobalJSONOutput {
					fmt.Printf("🚀 Initialized Gyrus global configuration successfully!\n")
					fmt.Printf("   - Configuration: %s (Profile: %s)\n", cfgPath, profile)
					fmt.Printf("   - Storage Root:  %s\n", storageDir)
					fmt.Printf("   - Default Tenant: %s\n", tenant)
					fmt.Printf("   - Default Project: %s\n", project)
				} else {
					fmt.Printf("{\"status\":\"configured\",\"config_file\":\"%s\",\"profile\":\"%s\",\"storage\":\"%s\",\"tenant\":\"%s\",\"project\":\"%s\",\"scope\":\"global\"}\n",
						cfgPath, profile, storageDir, tenant, project)
				}
				return nil
			}

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			res, err := config.RunConfigSetup(config.ConfigSetupOptions{
				ProjectDir: cwd,
				Profile:    config.Profile(profile),
				Tenant:     tenant,
				Project:    project,
			})
			if err != nil {
				return err
			}

			if !GlobalJSONOutput {
				fmt.Printf("🚀 Initialized Gyrus project configuration successfully!\n")
				fmt.Printf("   - Configuration: %s (Profile: %s)\n", res.ConfigFile, profile)
				fmt.Printf("   - Storage Root:  %s\n", res.StorageDir)
				fmt.Printf("   - Default Tenant: %s\n", res.Tenant)
				fmt.Printf("   - Default Project: %s\n", res.Project)
			} else {
				fmt.Printf("{\"status\":\"configured\",\"config_file\":\"%s\",\"profile\":\"%s\",\"storage\":\"%s\",\"tenant\":\"%s\",\"project\":\"%s\",\"scope\":\"project\"}\n",
					res.ConfigFile, profile, res.StorageDir, res.Tenant, res.Project)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&profile, "profile", "p", "local", "Configuration profile: local, git, blob, s3, azure, gcs, postgres")
	cmd.Flags().StringVarP(&tenant, "tenant", "t", "root", "Default tenant for context concepts")
	cmd.Flags().StringVarP(&project, "project", "P", "default", "Default project for context concepts")
	cmd.Flags().BoolVarP(&global, "global", "g", false, "Write configuration to global ~/.gyrus.yaml instead of project")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration file without prompting")

	return cmd
}
