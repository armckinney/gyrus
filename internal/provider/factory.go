package provider

import (
	"context"
	"fmt"

	"github.com/armckinney/gyrus/internal/provider/blob"
	"github.com/armckinney/gyrus/internal/provider/git"
	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/internal/provider/postgres"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// NewDocumentStore creates the appropriate gyrus.DocumentStore implementation
// based on the configuration settings in .gyrus.yaml.
func NewDocumentStore(cfg *localfs.Config, storageRoot string) (gyrus.DocumentStore, error) {
	if cfg == nil {
		cfg = &localfs.Config{}
	}

	switch cfg.StorageProvider {
	case "git":
		if cfg.Git.RepoURL == "" {
			return nil, fmt.Errorf("git storage provider selected but git.repo_url is empty in config file (.gyrus.yaml)")
		}
		return git.NewStore(git.Options{
			RepoURL: cfg.Git.RepoURL,
			Branch:  cfg.Git.Branch,
		})

	case "blob":
		bucketURL := cfg.Blob.BucketURL
		if bucketURL == "" {
			bucketURL = "file://" + storageRoot
		}
		return blob.NewStore(context.Background(), bucketURL, cfg.Blob.Prefix)

	case "postgres":
		if cfg.Postgres.ConnectionString == "" {
			return nil, fmt.Errorf("postgres storage provider selected but postgres.connection_string is empty in config file (.gyrus.yaml)")
		}
		return postgres.NewStore(context.Background(), cfg.Postgres.ConnectionString)

	case "localfs", "":
		return localfs.NewStore(storageRoot)

	default:
		return nil, fmt.Errorf("unknown storage_provider: '%s' in configuration file", cfg.StorageProvider)
	}
}
