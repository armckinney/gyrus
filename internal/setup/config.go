package setup

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

// GenerateConfigYaml generates the .gyrus.yaml config file content for a given profile using the nested schema.
func GenerateConfigYaml(profile Profile, ownerGroup string) string {
	if ownerGroup == "" {
		ownerGroup = "root"
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

default_owner_group: %s

git:
  repo_url: ""
  branch: main
`, ownerGroup)

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

default_owner_group: %s

s3:
  bucket_name: "my-gyrus-bucket"
  region: "us-east-1"
`, ownerGroup)

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

default_owner_group: %s

azure_blob:
  storage_account: "myaccountname"
  container_name: "my-gyrus-container"
`, ownerGroup)

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

default_owner_group: %s

gcs:
  bucket_name: "my-gyrus-gcs-bucket"
`, ownerGroup)

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

default_owner_group: %s

blob:
  bucket_url: "file:///tmp/gyrus-blob-bucket"
`, ownerGroup)

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

default_owner_group: %s

postgres:
  connection_string: "postgres://postgres:postgres@postgres:5432/gyrus?sslmode=disable"
`, ownerGroup)

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

default_owner_group: %s

vector:
  embedding_provider: ollama
  model: nomic-embed-text
  ollama_endpoint: "http://localhost:11434"
`, ownerGroup)

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

default_owner_group: %s
`, ownerGroup)
	}
}

// WriteConfigFile writes .gyrus.yaml to the specified workspace directory.
func WriteConfigFile(dir string, profile Profile, ownerGroup string) (string, error) {
	configPath := filepath.Join(dir, ".gyrus.yaml")
	content := GenerateConfigYaml(profile, ownerGroup)

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write .gyrus.yaml config file: %w", err)
	}
	return configPath, nil
}

// WriteGlobalConfigFile writes .gyrus.yaml to the user's home directory.
// Returns an error if the file already exists and force is false.
func WriteGlobalConfigFile(profile Profile, ownerGroup string, force bool) (string, error) {
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
	return WriteConfigFile(homeDir, profile, ownerGroup)
}
