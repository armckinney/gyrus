package config_test

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/config"
)

func TestEffectiveProviderDefaultsAndCompatibility(t *testing.T) {
	for _, storage := range []string{"", "localfs", "git", "blob", "s3", "aws_s3", "azure", "azure_blob", "gcs", "gcp", "gcp_gcs"} {
		t.Run(storage, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.Storage.Provider = storage
			resolved, err := config.ResolveProviders(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if resolved.Index.Provider != "okf" || resolved.Graph.Provider != "okf" || resolved.Search.Provider != "okf" {
				t.Fatalf("defaults: %+v", resolved)
			}
		})
	}
	cfg := &config.Config{}
	cfg.Storage.Provider = "postgres"
	resolved, err := config.ResolveProviders(cfg)
	if err != nil || resolved.IndexProvider() != "postgres" || resolved.SearchProvider() != "postgres_fts" {
		t.Fatalf("postgres defaults: %+v %v", resolved, err)
	}
	cfg = &config.Config{}
	cfg.Search.Provider = "vector"
	resolved, err = config.ResolveProviders(cfg)
	if err != nil || resolved.IndexProvider() != "okf" || resolved.GraphProvider() != "okf" {
		t.Fatalf("vector defaults: %+v %v", resolved, err)
	}
	for _, tc := range []struct {
		index, graph, search string
		valid                bool
	}{{"okf", "okf", "okf", true}, {"okf", "okf", "vector", true}, {"postgres", "okf", "okf", true}, {"okf", "okf", "sqlite_fts5", false}, {"okf", "sqlite", "okf", false}, {"sqlite", "okf", "okf", false}, {"okf", "okf", "sqlite", false}, {"sqlite", "sqlite", "unknown", false}} {
		cfg = &config.Config{}
		cfg.Index.Provider = tc.index
		cfg.Graph.Provider = tc.graph
		cfg.Search.Provider = tc.search
		_, err = config.ResolveProviders(cfg)
		if (err == nil) != tc.valid {
			t.Errorf("%+v: %v", tc, err)
		}
	}
	for _, profile := range []config.Profile{config.ProfileLocal, config.ProfileGit, config.ProfileBlob, config.ProfileS3, config.ProfileAzure, config.ProfileGCS, config.ProfilePostgres, config.ProfileVector} {
		var cfg config.Config
		if err := yaml.Unmarshal([]byte(config.GenerateConfigYaml(profile, "root", "demo")), &cfg); err != nil {
			t.Fatal(err)
		}
		if _, err := config.ResolveProviders(&cfg); err != nil {
			t.Fatalf("profile %s: %v", profile, err)
		}
		if profile != config.ProfilePostgres && profile != config.ProfileVector && cfg.IndexProvider() != "okf" {
			t.Fatalf("profile %s did not use OKF", profile)
		}
	}
}

func TestExplicitInvalidConfigDoesNotFallBack(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(name, []byte("index:\n  provider: okf\nsearch:\n  provider: sqlite_fts5\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GYRUS_CONFIG", name)
	if _, err := config.Load(); err == nil {
		t.Fatal("invalid explicit provider settings silently fell back")
	}
}
