package config

// Config represents the unified Gyrus configuration schema.
// Uses nested YAML structure as the canonical format.
type Config struct {
	Storage struct {
		Provider string `yaml:"provider" json:"provider"`
		Root     string `yaml:"root" json:"root"`
	} `yaml:"storage" json:"storage"`
	Index struct {
		Provider string `yaml:"provider" json:"provider"`
	} `yaml:"index" json:"index"`
	Graph struct {
		Provider string `yaml:"provider" json:"provider"`
	} `yaml:"graph" json:"graph"`
	Search struct {
		Provider string `yaml:"provider" json:"provider"`
	} `yaml:"search" json:"search"`
	Project     string `yaml:"project" json:"project"`
	Tenant      string `yaml:"tenant" json:"tenant"`
	SchemasPath string `yaml:"schemas_path" json:"schemas_path"`
	Git         struct {
		RepoURL string `yaml:"repo_url" json:"repo_url"`
		Branch  string `yaml:"branch" json:"branch"`
	} `yaml:"git" json:"git"`
	Blob struct {
		BucketURL      string `yaml:"bucket_url" json:"bucket_url"`
		StorageAccount string `yaml:"storage_account" json:"storage_account"`
		ContainerName  string `yaml:"container_name" json:"container_name"`
		Prefix         string `yaml:"prefix" json:"prefix"`
	} `yaml:"blob" json:"blob"`
	S3 struct {
		BucketName string `yaml:"bucket_name" json:"bucket_name"`
		Region     string `yaml:"region" json:"region"`
	} `yaml:"s3" json:"s3"`
	AzureBlob struct {
		StorageAccount string `yaml:"storage_account" json:"storage_account"`
		ContainerName  string `yaml:"container_name" json:"container_name"`
	} `yaml:"azure_blob" json:"azure_blob"`
	GCS struct {
		BucketName string `yaml:"bucket_name" json:"bucket_name"`
	} `yaml:"gcs" json:"gcs"`
	Vector struct {
		EmbeddingProvider string `yaml:"embedding_provider" json:"embedding_provider"`
		Model             string `yaml:"model" json:"model"`
		OllamaEndpoint    string `yaml:"ollama_endpoint" json:"ollama_endpoint"`
	} `yaml:"vector" json:"vector"`
	Postgres struct {
		ConnectionString string `yaml:"connection_string" json:"connection_string"`
	} `yaml:"postgres" json:"postgres"`
	UI struct {
		Client  string `yaml:"client" json:"client"`
		Command string `yaml:"command" json:"command"`
	} `yaml:"ui" json:"ui"`
}

// StorageProvider returns the configured storage provider.
func (c *Config) StorageProvider() string {
	if c == nil {
		return DefaultStorageProvider
	}
	return storageName(c.Storage.Provider)
}

// IndexProvider returns the configured index provider.
func (c *Config) IndexProvider() string { return c.effectiveIndex() }

// SearchProvider returns the configured search provider.
func (c *Config) SearchProvider() string {
	if c != nil && c.Search.Provider != "" {
		return c.Search.Provider
	}
	switch c.effectiveIndex() {
	case "postgres":
		return "postgres_fts"
	}
	return "okf"
}

// GraphProvider returns the configured graph provider.
func (c *Config) GraphProvider() string {
	if c != nil && c.Graph.Provider != "" {
		return c.Graph.Provider
	}
	return c.effectiveIndex()
}

// StorageRoot returns the configured storage root path.
func (c *Config) StorageRoot() string {
	if c == nil {
		return ""
	}
	return c.Storage.Root
}

// ProjectName returns the explicitly configured project name, if any.
func (c *Config) ProjectName() string {
	if c == nil {
		return ""
	}
	return c.Project
}

// TenantName returns the configured default tenant, falling back to DefaultTenant.
func (c *Config) TenantName() string {
	if c == nil || c.Tenant == "" {
		return DefaultTenant
	}
	return c.Tenant
}

// UIClient returns the configured UI client harness (e.g. "antigravity", "claude").
func (c *Config) UIClient() string {
	if c == nil {
		return ""
	}
	return c.UI.Client
}

// UICommand returns any optional custom executable command for the UI agent.
func (c *Config) UICommand() string {
	if c == nil {
		return ""
	}
	return c.UI.Command
}

// ConfigSource indicates which configuration tier provided the resolved config.
type ConfigSource string

const (
	SourceDefault ConfigSource = "default"
	SourceGlobal  ConfigSource = "global"  // ~/.gyrus.yaml
	SourceProject ConfigSource = "project" // .gyrus.yaml in project
)

// ScopeType represents the retrieval and storage context boundary.
type ScopeType string

const (
	ScopeProject   ScopeType = "project"
	ScopeReference ScopeType = "reference"
	ScopeAll       ScopeType = "all"
)

// ResolvedConfig wraps the final Config with resolution metadata.
type ResolvedConfig struct {
	Config      *Config      `json:"config"`
	Source      ConfigSource `json:"source"`
	SourcePath  string       `json:"source_path"`
	StorageRoot string       `json:"storage_root"`
}
