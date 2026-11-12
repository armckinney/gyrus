package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

func storageName(value string) string {
	if value == "" {
		return DefaultStorageProvider
	}
	return value
}
func (c *Config) effectiveIndex() string {
	if c != nil && c.Index.Provider != "" {
		return c.Index.Provider
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
		case "postgres":
			c.Search.Provider = "postgres_fts"
		default:
			c.Search.Provider = "okf"
		}
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
	wantIndex, wantSearch := "okf", "okf"
	if c.Storage.Provider == "postgres" {
		wantIndex, wantSearch = "postgres", "postgres_fts"
	}
	if c.Index.Provider != wantIndex || c.Graph.Provider != wantIndex || c.Search.Provider != wantSearch {
		return nil, fmt.Errorf("%s storage requires index=%s graph=%s search=%s", c.Storage.Provider, wantIndex, wantIndex, wantSearch)
	}
	if c.Storage.Provider == "git" {
		if c.Git.TrunkBranch == "" {
			c.Git.TrunkBranch = "main"
		}
		if err := plumbing.NewBranchReferenceName(c.Git.TrunkBranch).Validate(); err != nil {
			return nil, fmt.Errorf("invalid git.trunk_branch: %w", err)
		}
		if c.Git.CommitStrategy == "" {
			c.Git.CommitStrategy = "direct"
		}
		if c.Git.CommitStrategy != "direct" && c.Git.CommitStrategy != "branching" {
			return nil, fmt.Errorf("invalid git.commit_strategy: choose direct or branching")
		}
		if strings.ContainsAny(c.Git.RepoURL, "\r\n") {
			return nil, fmt.Errorf("git.repo_url must not contain newlines")
		}
		if u, err := url.Parse(c.Git.RepoURL); err == nil && u.User != nil {
			if _, set := u.User.Password(); set {
				return nil, fmt.Errorf("git.repo_url must not contain a password; use Git credential helpers")
			}
		}
	}
	return c, nil
}

func defaultResolvedConfig() *Config { cfg, _ := ResolveProviders(nil); return cfg }
