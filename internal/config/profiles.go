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
	ProfileBlob     Profile = "blob"
	ProfileS3       Profile = "s3"
	ProfileAzure    Profile = "azure"
	ProfileGCS      Profile = "gcs"
	ProfilePostgres Profile = "postgres"
	ProfileVector   Profile = "vector"
)

// ConfigSetupOptions defines input options for project config initialization.
type ConfigSetupOptions struct {
	ProjectDir string
	Profile    Profile
	Tenant     string
}

// ConfigSetupResult contains status details from project configuration setup.
type ConfigSetupResult struct {
	ConfigFile string
	StorageDir string
	Profile    Profile
	Tenant     string
}

// GenerateConfigYaml generates the .gyrus.yaml config file content for a given profile using the nested schema.
func GenerateConfigYaml(profile Profile, tenant string) string {
	if tenant == "" {
		tenant = "root"
	}

	switch profile {
	case ProfileGit:
		return fmt.Sprintf(`# Gyrus Configuration - Git Profile
storage:
  provider: git
  root: .gyrus

index:
  provider: sqlite

graph:
  provider: sqlite

search:
  provider: sqlite

tenant: %s

git:
  repo_url: ""
  branch: main
`, tenant)

	case ProfileS3:
		return fmt.Sprintf(`# Gyrus Configuration - AWS S3 Profile
storage:
  provider: s3
  root: .gyrus

index:
  provider: sqlite

graph:
  provider: sqlite

search:
  provider: sqlite

tenant: %s

s3:
  bucket_name: "my-gyrus-bucket"
  region: "us-east-1"
`, tenant)

	case ProfileAzure:
		return fmt.Sprintf(`# Gyrus Configuration - Azure Blob Profile
storage:
  provider: azure_blob
  root: .gyrus

index:
  provider: sqlite

graph:
  provider: sqlite

search:
  provider: sqlite

tenant: %s

azure_blob:
  storage_account: "myaccountname"
  container_name: "my-gyrus-container"
`, tenant)

	case ProfileGCS:
		return fmt.Sprintf(`# Gyrus Configuration - Google Cloud Storage Profile
storage:
  provider: gcs
  root: .gyrus

index:
  provider: sqlite

graph:
  provider: sqlite

search:
  provider: sqlite

tenant: %s

gcs:
  bucket_name: "my-gyrus-gcs-bucket"
`, tenant)

	case ProfileBlob:
		return fmt.Sprintf(`# Gyrus Configuration - Cloud Blob Profile
storage:
  provider: blob
  root: .gyrus

index:
  provider: sqlite

graph:
  provider: sqlite

search:
  provider: sqlite

tenant: %s

blob:
  bucket_url: "file:///tmp/gyrus-blob-bucket"
`, tenant)

	case ProfilePostgres:
		return fmt.Sprintf(`# Gyrus Configuration - PostgreSQL Profile
storage:
  provider: postgres
  root: .gyrus

index:
  provider: postgres

graph:
  provider: postgres

search:
  provider: postgres_fts

tenant: %s

postgres:
  connection_string: "postgres://postgres:postgres@postgres:5432/gyrus?sslmode=disable"
`, tenant)

	case ProfileVector:
		return fmt.Sprintf(`# Gyrus Configuration - Semantic Vector Profile
storage:
  provider: localfs
  root: .gyrus

index:
  provider: sqlite

graph:
  provider: sqlite

search:
  provider: vector

tenant: %s

vector:
  embedding_provider: ollama
  model: nomic-embed-text
  ollama_endpoint: "http://localhost:11434"
`, tenant)

	default: // ProfileLocal
		return fmt.Sprintf(`# Gyrus Configuration - LocalFS Profile
storage:
  provider: localfs
  root: .gyrus

index:
  provider: sqlite

graph:
  provider: sqlite

search:
  provider: sqlite

tenant: %s
`, tenant)
	}
}

// WriteConfigFile writes .gyrus.yaml to the specified project directory.
func WriteConfigFile(dir string, profile Profile, tenant string) (string, error) {
	configPath := filepath.Join(dir, ".gyrus.yaml")
	content := GenerateConfigYaml(profile, tenant)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write .gyrus.yaml config file: %w", err)
	}
	return configPath, nil
}

// WriteGlobalConfigFile writes .gyrus.yaml to the user's home directory.
// Returns an error if the file already exists and force is false.
func WriteGlobalConfigFile(profile Profile, tenant string, force bool) (string, error) {
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
	return WriteConfigFile(homeDir, profile, tenant)
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
		opts.Tenant = "root"
	}

	cfgPath, err := WriteConfigFile(opts.ProjectDir, opts.Profile, opts.Tenant)
	if err != nil {
		return nil, err
	}

	return &ConfigSetupResult{
		ConfigFile: cfgPath,
		StorageDir: filepath.Join(opts.ProjectDir, ".gyrus"),
		Profile:    opts.Profile,
		Tenant:     opts.Tenant,
	}, nil
}
