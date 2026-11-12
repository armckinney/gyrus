package provider

import (
	"context"
	"fmt"

	"github.com/armckinney/gyrus/internal/config"
	graphokf "github.com/armckinney/gyrus/internal/provider/graph/okf"
	graphpostgres "github.com/armckinney/gyrus/internal/provider/graph/postgres"
	indexokf "github.com/armckinney/gyrus/internal/provider/index/okf"
	indexpostgres "github.com/armckinney/gyrus/internal/provider/index/postgres"
	searchokf "github.com/armckinney/gyrus/internal/provider/search/okf"
	postgresfts "github.com/armckinney/gyrus/internal/provider/search/postgres_fts"
	"github.com/armckinney/gyrus/internal/provider/storage/git"
	"github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/armckinney/gyrus/internal/provider/storage/postgres"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// NewConceptStore creates the appropriate gyrus.ConceptStore implementation
// based on configuration settings in .gyrus.yaml.
func NewConceptStore(cfg *config.Config, storageRoot string) (gyrus.ConceptStore, error) {
	resolved, err := config.ResolveProviders(cfg)
	if err != nil {
		return nil, err
	}
	cfg = resolved

	switch cfg.StorageProvider() {
	case "git":
		if cfg.Git.RepoURL == "" {
			return nil, fmt.Errorf("git storage provider selected but git.repo_url is empty in config file (.gyrus.yaml)")
		}
		return git.NewStore(git.Options{
			RepoURL:        cfg.Git.RepoURL,
			TrunkBranch:    cfg.Git.TrunkBranch,
			CommitStrategy: cfg.Git.CommitStrategy,
		})

	case "postgres":
		if cfg.Postgres.ConnectionString == "" {
			return nil, fmt.Errorf("postgres storage provider selected but postgres.connection_string is empty in config file (.gyrus.yaml)")
		}
		return postgres.NewStore(context.Background(), cfg.Postgres.ConnectionString)

	case "localfs", "":
		return localfs.NewStoreWithProject(storageRoot, cfg.ProjectName())

	default:
		return nil, fmt.Errorf("unknown storage_provider: '%s' in configuration file", cfg.StorageProvider())
	}
}

// NewIndexStore creates the configured index over the selected storage.
func NewIndexStore(cfg *config.Config, store gyrus.ConceptStore) (gyrus.IndexStore, error) {
	cfg, err := config.ResolveProviders(cfg)
	if err != nil {
		return nil, err
	}
	switch cfg.IndexProvider() {
	case "okf":
		return indexokf.NewIndexer(store)
	case "postgres":
		if cfg.Postgres.ConnectionString == "" {
			return nil, fmt.Errorf("postgres index requires postgres.connection_string")
		}
		return indexpostgres.NewIndexer(context.Background(), cfg.Postgres.ConnectionString)
	default:
		return nil, fmt.Errorf("unknown index provider %q", cfg.IndexProvider())
	}
}

// NewGraphStore reuses database indexes or persists links in the selected store.
func NewGraphStore(cfg *config.Config, store gyrus.ConceptStore, indexer gyrus.IndexStore) (gyrus.GraphStore, error) {
	cfg, err := config.ResolveProviders(cfg)
	if err != nil {
		return nil, err
	}
	switch cfg.GraphProvider() {
	case "okf":
		return graphokf.NewGraphStore(store)
	case "postgres":
		idx, ok := indexer.(*indexpostgres.Indexer)
		if !ok {
			return nil, fmt.Errorf("postgres graph requires postgres index")
		}
		return graphpostgres.NewGraphStore(idx.Pool()), nil
	default:
		return nil, fmt.Errorf("unknown graph provider %q", cfg.GraphProvider())
	}
}

// NewSearchProvider shares the selected storage and index dependencies.
func NewSearchProvider(cfg *config.Config, store gyrus.ConceptStore, indexer gyrus.IndexStore) (gyrus.SearchProvider, error) {
	cfg, err := config.ResolveProviders(cfg)
	if err != nil {
		return nil, err
	}
	switch cfg.SearchProvider() {
	case "okf":
		return searchokf.NewSearchProvider(store)
	case "postgres_fts":
		idx, ok := indexer.(*indexpostgres.Indexer)
		if !ok {
			return nil, fmt.Errorf("postgres search requires postgres index")
		}
		return postgresfts.NewSearchProviderWithPool(idx.Pool()), nil
	default:
		return nil, fmt.Errorf("unknown search provider %q", cfg.SearchProvider())
	}
}
