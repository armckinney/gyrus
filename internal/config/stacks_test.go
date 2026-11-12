package config_test

import (
	"github.com/armckinney/gyrus/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestOnlyCompleteStacksResolve(t *testing.T) {
	for _, storage := range []string{"localfs", "git", "postgres"} {
		for _, index := range []string{"okf", "postgres"} {
			for _, graph := range []string{"okf", "postgres"} {
				for _, search := range []string{"okf", "postgres_fts"} {
					c := &config.Config{}
					c.Storage.Provider = storage
					c.Index.Provider = index
					c.Graph.Provider = graph
					c.Search.Provider = search
					want := storage != "postgres" && index == "okf" && graph == "okf" && search == "okf" || storage == "postgres" && index == "postgres" && graph == "postgres" && search == "postgres_fts"
					if _, err := config.ResolveProviders(c); (err == nil) != want {
						t.Errorf("%s/%s/%s/%s: %v", storage, index, graph, search, err)
					}
				}
			}
		}
	}
}

func TestRemovedConfigurationFailsWithoutFallback(t *testing.T) {
	for _, extra := range []string{"git:\n  branch: main\n", "git:\n  review_branch: work\n", "git:\n  base_branch: main\n", "git:\n  commit_strategy: branch\n", "git:\n  trunk_branch: bad..branch\n", "blob:\n  bucket_url: file:///tmp/test\n", "s3:\n  bucket_name: test\n", "unknown: true\n", "---\ntenant: one\n---\ntenant: two\n"} {
		t.Run(extra, func(t *testing.T) {
			root, home := t.TempDir(), t.TempDir()
			path := filepath.Join(root, ".gyrus.yaml")
			original := "storage:\n  provider: git\n  root: must-not-exist\n" + extra
			if err := os.WriteFile(path, []byte(original), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(home, ".gyrus.yaml"), []byte("tenant: fallback\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := config.LoadFrom(root, home); err == nil {
				t.Fatal("accepted invalid config or fell back")
			}
			if _, err := os.Stat(filepath.Join(root, "must-not-exist")); !os.IsNotExist(err) {
				t.Fatal("storage touched")
			}
			data, _ := os.ReadFile(path)
			if string(data) != original {
				t.Fatal("configuration modified")
			}
		})
	}
	for _, removed := range []string{"blob", "s3", "aws_s3", "azure", "azure_blob", "gcs", "gcp", "gcp_gcs"} {
		c := &config.Config{}
		c.Storage.Provider = removed
		if _, err := config.ResolveProviders(c); err == nil {
			t.Fatalf("accepted removed provider %s", removed)
		}
		if _, err := config.GenerateConfigYaml(config.Profile(removed), "root"); err == nil {
			t.Fatalf("accepted removed profile %s", removed)
		}
	}
}
