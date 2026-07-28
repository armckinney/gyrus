package provider

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/armckinney/gyrus/internal/provider/blob"
	"github.com/armckinney/gyrus/internal/provider/git"
	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/internal/provider/postgres"
	"github.com/armckinney/gyrus/internal/provider/sqlite"
	"github.com/armckinney/gyrus/internal/provider/vector"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// NewDocumentStore creates the appropriate gyrus.DocumentStore implementation
// based on the configuration settings in .gyrus.yaml.
func NewDocumentStore(cfg *localfs.Config, storageRoot string) (gyrus.DocumentStore, error) {
	if cfg == nil {
		cfg = &localfs.Config{}
	}

	prefix := cfg.StorageRoot
	if prefix == "" {
		prefix = cfg.Storage.Root
	}
	if prefix == "" {
		prefix = storageRoot
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

	case "s3", "aws_s3":
		bucketName := cfg.S3.BucketName
		if bucketName == "" {
			return nil, fmt.Errorf("s3 storage provider selected but s3.bucket_name is empty in config file (.gyrus.yaml)")
		}
		bucketURL := fmt.Sprintf("s3://%s", bucketName)
		if cfg.S3.Region != "" {
			bucketURL = fmt.Sprintf("s3://%s?region=%s", bucketName, cfg.S3.Region)
		}
		return blob.NewStore(context.Background(), bucketURL, prefix)

	case "azure", "azure_blob":
		container := cfg.AzureBlob.ContainerName
		if container == "" {
			container = cfg.Blob.ContainerName
		}
		if container == "" {
			return nil, fmt.Errorf("azure_blob storage provider selected but azure_blob.container_name is empty in config file (.gyrus.yaml)")
		}
		account := cfg.AzureBlob.StorageAccount
		if account == "" {
			account = cfg.Blob.StorageAccount
		}
		bucketURL := fmt.Sprintf("azblob://%s", container)
		if account != "" {
			bucketURL = fmt.Sprintf("azblob://%s?storage_account=%s", container, account)
		}
		return blob.NewStore(context.Background(), bucketURL, prefix)

	case "gcs", "gcp_gcs", "gcp":
		bucketName := cfg.GCS.BucketName
		if bucketName == "" {
			return nil, fmt.Errorf("gcs storage provider selected but gcs.bucket_name is empty in config file (.gyrus.yaml)")
		}
		bucketURL := fmt.Sprintf("gs://%s", bucketName)
		return blob.NewStore(context.Background(), bucketURL, prefix)

	case "blob":
		bucketURL := cfg.Blob.BucketURL
		if bucketURL == "" {
			if cfg.Blob.ContainerName != "" {
				container := cfg.Blob.ContainerName
				if cfg.Blob.StorageAccount != "" {
					bucketURL = fmt.Sprintf("azblob://%s?storage_account=%s", container, cfg.Blob.StorageAccount)
				} else {
					bucketURL = fmt.Sprintf("azblob://%s", container)
				}
			} else {
				bucketURL = "file://" + storageRoot
			}
		}
		if cfg.Blob.Prefix != "" {
			prefix = cfg.Blob.Prefix
		}
		return blob.NewStore(context.Background(), bucketURL, prefix)

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

// NewIndexStore creates the appropriate gyrus.IndexStore implementation
// based on index_provider in .gyrus.yaml.
func NewIndexStore(cfg *localfs.Config, storageRoot string) (gyrus.IndexStore, error) {
	if cfg == nil {
		cfg = &localfs.Config{}
	}

	switch cfg.IndexProvider {
	case "postgres":
		if cfg.Postgres.ConnectionString == "" {
			return nil, fmt.Errorf("postgres index provider selected but postgres.connection_string is empty")
		}
		return postgres.NewStore(context.Background(), cfg.Postgres.ConnectionString)

	default:
		dbPath := filepath.Join(storageRoot, "index.db")
		return sqlite.NewIndexer(dbPath)
	}
}

// NewGraphStore creates the appropriate gyrus.GraphStore implementation
// based on index_provider in .gyrus.yaml.
func NewGraphStore(cfg *localfs.Config, storageRoot string) (gyrus.GraphStore, error) {
	if cfg == nil {
		cfg = &localfs.Config{}
	}

	switch cfg.IndexProvider {
	case "postgres":
		if cfg.Postgres.ConnectionString == "" {
			return nil, fmt.Errorf("postgres index provider selected but postgres.connection_string is empty")
		}
		return postgres.NewStore(context.Background(), cfg.Postgres.ConnectionString)

	default:
		dbPath := filepath.Join(storageRoot, "index.db")
		return sqlite.NewIndexer(dbPath)
	}
}

type sqliteSearchAdapter struct {
	indexer *sqlite.Indexer
}

func (a *sqliteSearchAdapter) Search(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	return a.indexer.Search(ctx, gyrus.SearchQuery{
		Query:      query,
		Filter:     filter,
		MaxResults: 50,
	})
}

type postgresSearchAdapter struct {
	store *postgres.Store
}

func (a *postgresSearchAdapter) Search(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	return a.store.Search(ctx, gyrus.SearchQuery{
		Query:      query,
		Filter:     filter,
		MaxResults: 50,
	})
}

// NewSearchProvider creates the appropriate gyrus.SearchProvider implementation
// based on search_provider in .gyrus.yaml.
func NewSearchProvider(cfg *localfs.Config, storageRoot string) (gyrus.SearchProvider, error) {
	if cfg == nil {
		cfg = &localfs.Config{}
	}

	switch cfg.SearchProvider {
	case "vector":
		dbPath := filepath.Join(storageRoot, "index.db")
		sqliteStore, err := sqlite.NewIndexer(dbPath)
		if err != nil {
			return nil, err
		}
		var embedder vector.EmbeddingProvider
		if cfg.Vector.EmbeddingProvider == "openai" {
			embedder = vector.NewOpenAIEmbedder("")
		} else {
			endpoint := cfg.Vector.OllamaEndpoint
			if endpoint == "" {
				endpoint = "http://localhost:11434/api/embeddings"
			}
			model := cfg.Vector.Model
			if model == "" {
				model = "nomic-embed-text"
			}
			embedder = vector.NewOllamaEmbedder(endpoint, model)
		}
		adapter := &sqliteSearchAdapter{indexer: sqliteStore}
		vStore := vector.NewStore(adapter, embedder)

		// Load documents from storage into vector store for similarity scoring
		store, err := NewDocumentStore(cfg, storageRoot)
		if err == nil {
			if docs, err := sqliteStore.Search(context.Background(), gyrus.SearchQuery{MaxResults: 100}); err == nil {
				for _, res := range docs {
					if fullDoc, err := store.Get(context.Background(), res.Document.ID); err == nil {
						_ = vStore.AddDocument(context.Background(), fullDoc)
					}
				}
			}
		}
		return vStore, nil

	case "postgres_fts":
		if cfg.Postgres.ConnectionString == "" {
			return nil, fmt.Errorf("postgres_fts search provider selected but postgres.connection_string is empty")
		}
		store, err := postgres.NewStore(context.Background(), cfg.Postgres.ConnectionString)
		if err != nil {
			return nil, err
		}
		return &postgresSearchAdapter{store: store}, nil

	default: // "sqlite" or empty
		dbPath := filepath.Join(storageRoot, "index.db")
		indexer, err := sqlite.NewIndexer(dbPath)
		if err != nil {
			return nil, err
		}
		return &sqliteSearchAdapter{indexer: indexer}, nil
	}
}
