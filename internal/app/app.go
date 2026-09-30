package app

import (
	"fmt"
	"sync"

	"github.com/armckinney/gyrus/internal/config"
	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/provider"
)

// App is the central dependency injection service container for Gyrus.
type App struct {
	resolved *config.ResolvedConfig
	engine   *lifecycle.Engine
	engineMu sync.Mutex
}

// New initializes a new App container resolving configuration via config.Load().
func New() (*App, error) {
	resolved, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return &App{
		resolved: resolved,
	}, nil
}

// Reset reloads configuration and resets cached engine.
func (a *App) Reset() error {
	a.engineMu.Lock()
	defer a.engineMu.Unlock()

	resolved, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	a.resolved = resolved
	a.engine = nil
	return nil
}

// StorageRoot returns the resolved storage root directory path.
func (a *App) StorageRoot() string {
	if a.resolved == nil {
		return ""
	}
	return a.resolved.StorageRoot
}

// Config returns the resolved workspace or global configuration settings.
func (a *App) Config() *config.Config {
	if a.resolved == nil {
		return nil
	}
	return a.resolved.Config
}

// ResolvedConfig returns the complete ResolvedConfig metadata container.
func (a *App) ResolvedConfig() *config.ResolvedConfig {
	return a.resolved
}

// Engine constructs and returns the cached lifecycle.Engine domain service.
func (a *App) Engine() (*lifecycle.Engine, error) {
	a.engineMu.Lock()
	defer a.engineMu.Unlock()

	if a.engine != nil {
		return a.engine, nil
	}

	store, err := provider.NewDocumentStore(a.resolved.Config, a.resolved.StorageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed initializing storage provider: %w", err)
	}

	search, err := provider.NewSearchProvider(a.resolved.Config, a.resolved.StorageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed initializing search provider: %w", err)
	}

	indexer, err := provider.NewIndexStore(a.resolved.Config, a.resolved.StorageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed initializing index provider: %w", err)
	}

	graph, err := provider.NewGraphStore(a.resolved.Config, a.resolved.StorageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed initializing graph provider: %w", err)
	}

	a.engine = lifecycle.NewEngine(store, search, indexer, graph, a.resolved.StorageRoot)
	return a.engine, nil
}
