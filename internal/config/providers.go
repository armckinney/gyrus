package config

import (
	"fmt"
	"strings"
)

func storageName(value string) string {
	switch value {
	case "":
		return DefaultStorageProvider
	case "aws_s3":
		return "s3"
	case "azure":
		return "azure_blob"
	case "gcp_gcs", "gcp":
		return "gcs"
	}
	return value
}
func (c *Config) effectiveIndex() string {
	if c != nil && c.Index.Provider != "" {
		return c.Index.Provider
	}
	if c != nil && c.Search.Provider == "vector" {
		return "sqlite"
	}
	if c != nil && storageName(c.Storage.Provider) == "postgres" {
		return "postgres"
	}
	return DefaultIndexProvider
}

// ResolveProviders normalizes defaults and validates compatibility before resource creation.
func ResolveProviders(input *Config) (*Config, error) {
	c := &Config{}
	if input != nil {
		*c = *input
	}
	c.Storage.Provider = storageName(c.Storage.Provider)
	c.Index.Provider = c.effectiveIndex()
	if c.Graph.Provider == "" {
		c.Graph.Provider = c.Index.Provider
	}
	if c.Search.Provider == "" {
		switch c.Index.Provider {
		case "sqlite":
			c.Search.Provider = "sqlite_fts5"
		case "postgres":
			c.Search.Provider = "postgres_fts"
		default:
			c.Search.Provider = "okf"
		}
	}
	if c.Search.Provider == "sqlite" {
		c.Search.Provider = "sqlite_fts5"
	}
	for _, field := range []struct {
		name, value string
		valid       []string
	}{
		{"storage", c.Storage.Provider, ValidStorageProviders}, {"index", c.Index.Provider, ValidIndexProviders}, {"graph", c.Graph.Provider, ValidGraphProviders}, {"search", c.Search.Provider, ValidSearchProviders},
	} {
		if !isValidProvider(field.value, field.valid) {
			return nil, fmt.Errorf("unknown %s provider '%s'. Valid: %s", field.name, field.value, strings.Join(field.valid, ", "))
		}
	}
	compatible := false
	switch c.Index.Provider {
	case "okf":
		compatible = c.Graph.Provider == "okf" && c.Search.Provider == "okf"
	case "sqlite":
		compatible = (c.Graph.Provider == "sqlite" || c.Graph.Provider == "okf") && (c.Search.Provider == "sqlite_fts5" || c.Search.Provider == "vector" || c.Search.Provider == "okf")
	case "postgres":
		compatible = (c.Graph.Provider == "postgres" || c.Graph.Provider == "okf") && (c.Search.Provider == "postgres_fts" || c.Search.Provider == "okf")
	}
	if !compatible {
		return nil, fmt.Errorf("incompatible providers: index=%s graph=%s search=%s", c.Index.Provider, c.Graph.Provider, c.Search.Provider)
	}
	if c.Storage.Provider == "postgres" && (c.Index.Provider != "postgres" || c.Graph.Provider != "postgres" || c.Search.Provider != "postgres_fts") {
		return nil, fmt.Errorf("PostgreSQL storage requires index=postgres graph=postgres search=postgres_fts; OKF requires scanner-capable storage")
	}
	return c, nil
}

func defaultResolvedConfig() *Config { cfg, _ := ResolveProviders(nil); return cfg }
