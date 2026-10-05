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
		DSN      string `yaml:"dsn" json:"dsn"`
	} `yaml:"index" json:"index"`
	Graph struct {
		Provider string `yaml:"provider" json:"provider"`
	} `yaml:"graph" json:"graph"`
	Search struct {
		Provider string `yaml:"provider" json:"provider"`
	} `yaml:"search" json:"search"`
	Workspace         string `yaml:"workspace" json:"workspace"`
	DefaultOwnerGroup string `yaml:"default_owner_group" json:"default_owner_group"`
	SchemasPath       string `yaml:"schemas_path" json:"schemas_path"`
	Git               struct {
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
	Automation struct {
		HooksEnabled       *bool    `yaml:"hooks_enabled,omitempty" json:"hooks_enabled,omitempty"`
		AutoContext        *bool    `yaml:"auto_context,omitempty" json:"auto_context,omitempty"`
		ArchitecturalCheck *bool    `yaml:"architectural_check,omitempty" json:"architectural_check,omitempty"`
		IgnoredPaths       []string `yaml:"ignored_paths,omitempty" json:"ignored_paths,omitempty"`
	} `yaml:"automation,omitempty" json:"automation,omitempty"`
}

// StorageProvider returns the configured storage provider.
func (c *Config) StorageProvider() string {
	if c == nil {
		return ""
	}
	return c.Storage.Provider
}

// IndexProvider returns the configured index provider.
func (c *Config) IndexProvider() string {
	if c == nil {
		return ""
	}
	return c.Index.Provider
}

// SearchProvider returns the configured search provider.
func (c *Config) SearchProvider() string {
	if c == nil {
		return ""
	}
	return c.Search.Provider
}

// GraphProvider returns the configured graph provider.
func (c *Config) GraphProvider() string {
	if c == nil {
		return ""
	}
	return c.Graph.Provider
}

// StorageRoot returns the configured storage root path.
func (c *Config) StorageRoot() string {
	if c == nil {
		return ""
	}
	return c.Storage.Root
}

// WorkspaceName returns the explicitly configured workspace name, if any.
func (c *Config) WorkspaceName() string {
	if c == nil {
		return ""
	}
	return c.Workspace
}

// OwnerGroup returns the configured default owner group, falling back to DefaultOwnerGroup.
func (c *Config) OwnerGroup() string {
	if c == nil || c.DefaultOwnerGroup == "" {
		return DefaultOwnerGroup
	}
	return c.DefaultOwnerGroup
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

// HooksEnabled returns true unless explicitly disabled in config.
func (c *Config) HooksEnabled() bool {
	if c == nil || c.Automation.HooksEnabled == nil {
		return true
	}
	return *c.Automation.HooksEnabled
}

// AutoContext returns true unless explicitly disabled in config.
func (c *Config) AutoContext() bool {
	if c == nil || c.Automation.AutoContext == nil {
		return true
	}
	return *c.Automation.AutoContext
}

// ArchitecturalCheck returns true unless explicitly disabled in config.
func (c *Config) ArchitecturalCheck() bool {
	if c == nil || c.Automation.ArchitecturalCheck == nil {
		return true
	}
	return *c.Automation.ArchitecturalCheck
}

// IgnoredPaths returns configured paths to ignore during architectural check,
// falling back to DefaultIgnoredPaths if none specified.
func (c *Config) IgnoredPaths() []string {
	if c == nil || len(c.Automation.IgnoredPaths) == 0 {
		return DefaultIgnoredPaths
	}
	return c.Automation.IgnoredPaths
}

// ConfigSource indicates which configuration tier provided the resolved config.
type ConfigSource string

const (
	SourceDefault   ConfigSource = "default"
	SourceGlobal    ConfigSource = "global"    // ~/.gyrus.yaml
	SourceWorkspace ConfigSource = "workspace" // .gyrus.yaml in workspace
)

// ScopeType represents the retrieval and storage context boundary.
type ScopeType string

const (
	ScopeWorkspace ScopeType = "workspace"
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
