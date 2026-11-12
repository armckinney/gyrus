package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Profile represents a pre-configured storage and search setup matrix.
type Profile string

const (
	ProfileLocal    Profile = "local"
	ProfileGit      Profile = "git"
	ProfilePostgres Profile = "postgres"
)

// ConfigSetupOptions defines input options for project config initialization.
type ConfigSetupOptions struct {
	ProjectDir string
	Profile    Profile
	Tenant     string
	Project    string
}

// ConfigSetupResult contains status details from project configuration setup.
type ConfigSetupResult struct {
	ConfigFile string
	StorageDir string
	Profile    Profile
	Tenant     string
	Project    string
}

// GenerateConfigYaml generates the .gyrus.yaml config file content for a given profile using the nested schema.
func GenerateConfigYaml(profile Profile, tenant string, project ...string) (string, error) {
	if tenant == "" {
		tenant = DefaultTenant
	}
	proj := DefaultProject
	if len(project) > 0 && project[0] != "" {
		proj = project[0]
	}

	switch profile {
	case ProfileGit:
		return fmt.Sprintf(`# Gyrus Configuration - Git Profile
tenant: %s
project: %s

storage:
  provider: git

index:
  provider: okf

graph:
  provider: okf

search:
  provider: okf

git:
  repo_url: ""
  trunk_branch: main
  commit_strategy: direct
`, tenant, proj), nil

	case ProfilePostgres:
		return fmt.Sprintf(`# Gyrus Configuration - PostgreSQL Profile
tenant: %s
project: %s

storage:
  provider: postgres

index:
  provider: postgres

graph:
  provider: postgres

search:
  provider: postgres_fts

postgres:
  connection_string: "postgres://postgres:postgres@postgres:5432/gyrus?sslmode=disable"
`, tenant, proj), nil

	case ProfileLocal:
		return fmt.Sprintf(`# Gyrus Configuration - LocalFS Profile
tenant: %s
project: %s

storage:
  provider: localfs
  root: .gyrus

index:
  provider: okf

graph:
  provider: okf

search:
  provider: okf
`, tenant, proj), nil
	default:
		return "", fmt.Errorf("unsupported configuration profile %q: choose local, git, or postgres", profile)
	}
}

// WriteConfigFile writes .gyrus.yaml to the specified project directory.
func WriteConfigFile(dir string, profile Profile, tenant string, project ...string) (string, error) {
	proj := DefaultProject
	if len(project) > 0 && project[0] != "" {
		proj = project[0]
	}
	configPath := filepath.Join(dir, ".gyrus.yaml")
	content, err := GenerateConfigYaml(profile, tenant, proj)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write .gyrus.yaml config file: %w", err)
	}
	return configPath, nil
}

// WriteGlobalConfigFile writes .gyrus.yaml to the user's home directory.
// Returns an error if the file already exists and force is false.
func WriteGlobalConfigFile(profile Profile, tenant string, force bool, project ...string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	configPath := filepath.Join(homeDir, ".gyrus.yaml")
	if !force {
		if _, err := os.Stat(configPath); err == nil {
			return "", fmt.Errorf("global config already exists at %s. Use --force to overwrite", configPath)
		}
	}
	return WriteConfigFile(homeDir, profile, tenant, project...)
}

// RunConfigSetup generates .gyrus.yaml configuration for the target project.
func RunConfigSetup(opts ConfigSetupOptions) (*ConfigSetupResult, error) {
	if opts.ProjectDir == "" {
		opts.ProjectDir = "."
	}
	if opts.Profile == "" {
		opts.Profile = ProfileLocal
	}
	if opts.Tenant == "" {
		opts.Tenant = DefaultTenant
	}
	if opts.Project == "" {
		opts.Project = DefaultProject
	}

	cfgPath, err := WriteConfigFile(opts.ProjectDir, opts.Profile, opts.Tenant, opts.Project)
	if err != nil {
		return nil, err
	}

	return &ConfigSetupResult{
		ConfigFile: cfgPath,
		StorageDir: filepath.Join(opts.ProjectDir, ".gyrus"),
		Profile:    opts.Profile,
		Tenant:     opts.Tenant,
		Project:    opts.Project,
	}, nil
}
