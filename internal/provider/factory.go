package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/armckinney/gyrus/internal/config"
	graphokf "github.com/armckinney/gyrus/internal/provider/graph/okf"
	graphpostgres "github.com/armckinney/gyrus/internal/provider/graph/postgres"
	indexokf "github.com/armckinney/gyrus/internal/provider/index/okf"
	indexpostgres "github.com/armckinney/gyrus/internal/provider/index/postgres"
	searchokf "github.com/armckinney/gyrus/internal/provider/search/okf"
	postgresfts "github.com/armckinney/gyrus/internal/provider/search/postgres_fts"
	vectorsearch "github.com/armckinney/gyrus/internal/provider/search/vector"
	"github.com/armckinney/gyrus/internal/provider/storage/blob"
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
	if cfg == nil {
		cfg = &config.Config{}
	}

	prefix := cfg.StorageRoot()
	if prefix == "" {
		prefix = config.DefaultStorageRoot
	}

	switch cfg.StorageProvider() {
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
		return localfs.NewStoreWithProject(storageRoot, cfg.ProjectName())

	default:
		return nil, fmt.Errorf("unknown storage_provider: '%s' in configuration file", cfg.StorageProvider())
	}
}

// NewIndexStore creates the configured index over the selected storage.
func NewIndexStore(cfg *config.Config, storageRoot string, store gyrus.ConceptStore) (gyrus.IndexStore, error) {
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
func NewGraphStore(cfg *config.Config, storageRoot string, store gyrus.ConceptStore, indexer gyrus.IndexStore) (gyrus.GraphStore, error) {
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
func NewSearchProvider(cfg *config.Config, storageRoot string, store gyrus.ConceptStore, indexer gyrus.IndexStore) (gyrus.SearchProvider, error) {
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
	case "vector":
		lexical, err := searchokf.NewSearchProvider(store)
		if err != nil {
			return nil, err
		}
		var embedder vectorsearch.EmbeddingProvider
		if cfg.Vector.EmbeddingProvider == "openai" {
			embedder = vectorsearch.NewOpenAIEmbedder("")
		} else {
			endpoint := cfg.Vector.OllamaEndpoint
			if endpoint == "" {
				endpoint = os.Getenv("OLLAMA_HOST")
			}
			if endpoint == "" {
				endpoint = "http://localhost:11434/api/embeddings"
			}
			model := cfg.Vector.Model
			if model == "" {
				model = "nomic-embed-text"
			}
			embedder = vectorsearch.NewOllamaEmbedder(endpoint, model)
		}
		v := vectorsearch.NewStore(lexical, embedder)
		scanner, ok := store.(gyrus.ConceptScanner)
		if !ok {
			return nil, fmt.Errorf("vector search requires scanner-capable storage")
		}
		report, err := scanner.Scan(context.Background(), func(doc gyrus.Concept) error {
			return v.AddConcept(context.Background(), doc)
		})
		if err != nil || len(report.Errors) > 0 {
			return nil, fmt.Errorf("initialize vector search: %v; collection errors: %v", err, report.Errors)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unknown search provider %q", cfg.SearchProvider())
	}
}
