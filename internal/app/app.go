package app

import (
	"fmt"
	"sync"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/provider"
	"github.com/armckinney/gyrus/internal/provider/storage/localfs"
)

// App is the central dependency injection service container for Gyrus.
type App struct {
	storageRoot string
	config      *localfs.Config
	engine      *lifecycle.Engine
	engineMu    sync.Mutex
}

// New initializes a new App container for the target workspace storageRoot.
func New(storageRoot string) (*App, error) {
	absRoot, err := localfs.ResolveStoragePath(storageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve storage path: %w", err)
	}

	cfg, _, err := localfs.LoadConfig(absRoot)
	if err != nil {
		cfg = &localfs.Config{}
	}

	return &App{
		storageRoot: absRoot,
		config:      cfg,
	}, nil
}

// StorageRoot returns the workspace storage root directory path.
func (a *App) StorageRoot() string {
	return a.storageRoot
}

// Config returns the workspace configuration settings.
func (a *App) Config() *localfs.Config {
	return a.config
}

// Engine constructs and returns the cached lifecycle.Engine domain service.
func (a *App) Engine() (*lifecycle.Engine, error) {
	a.engineMu.Lock()
	defer a.engineMu.Unlock()

	if a.engine != nil {
		return a.engine, nil
	}

	store, err := provider.NewDocumentStore(a.config, a.storageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed initializing storage provider: %w", err)
	}

	search, err := provider.NewSearchProvider(a.config, a.storageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed initializing search provider: %w", err)
	}

	indexer, err := provider.NewIndexStore(a.config, a.storageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed initializing index provider: %w", err)
	}

	graph, err := provider.NewGraphStore(a.config, a.storageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed initializing graph provider: %w", err)
	}

	a.engine = lifecycle.NewEngine(store, search, indexer, graph, a.storageRoot)
	return a.engine, nil
}
