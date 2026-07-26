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
	ProfilePostgres Profile = "postgres"
	ProfileVector   Profile = "vector"
)

// GenerateConfigYaml generates the .gyrus.yaml config file content for a given profile.
func GenerateConfigYaml(profile Profile, ownerGroup string) string {
	if ownerGroup == "" {
		ownerGroup = "armckinney"
	}

	switch profile {
	case ProfileGit:
		return fmt.Sprintf(`# Gyrus CLI & MCP Configuration - Git Profile
storage_provider: git
index_provider: sqlite
search_provider: sqlite

storage_root: docs/.gyrus/docs
schemas_path: docs/.gyrus/schemas
default_owner_group: %s

git:
  repo_url: ""
  branch: main
`, ownerGroup)

	case ProfileBlob:
		return fmt.Sprintf(`# Gyrus CLI & MCP Configuration - Cloud Blob Profile
storage_provider: blob
index_provider: sqlite
search_provider: sqlite

storage_root: docs/.gyrus/docs
schemas_path: docs/.gyrus/schemas
default_owner_group: %s

blob:
  bucket_url: "file:///tmp/gyrus-blob-bucket"
  prefix: docs
`, ownerGroup)

	case ProfilePostgres:
		return fmt.Sprintf(`# Gyrus CLI & MCP Configuration - PostgreSQL Profile
storage_provider: postgres
index_provider: postgres
search_provider: postgres_fts

storage_root: docs/.gyrus/docs
schemas_path: docs/.gyrus/schemas
default_owner_group: %s

postgres:
  connection_string: "postgres://postgres:postgres@localhost:5432/gyrus?sslmode=disable"
`, ownerGroup)

	case ProfileVector:
		return fmt.Sprintf(`# Gyrus CLI & MCP Configuration - Semantic Vector Profile
storage_provider: localfs
index_provider: sqlite
search_provider: vector

storage_root: docs/.gyrus/docs
schemas_path: docs/.gyrus/schemas
default_owner_group: %s

vector:
  embedding_provider: ollama
  model: nomic-embed-text
  ollama_endpoint: "http://localhost:11434"
`, ownerGroup)

	default: // ProfileLocal
		return fmt.Sprintf(`# Gyrus CLI & MCP Configuration - LocalFS Profile
storage_provider: localfs
index_provider: sqlite
search_provider: sqlite

storage_root: docs/.gyrus/docs
schemas_path: docs/.gyrus/schemas
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
