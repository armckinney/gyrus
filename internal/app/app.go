package app

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/armckinney/gyrus/internal/config"
	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/provider"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// App is the central dependency injection service container for Gyrus.
type App struct {
	resolved *config.ResolvedConfig
	engine   *lifecycle.Engine
	store    gyrus.ConceptStore
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

// NewWithProject initializes an App container resolving configuration from an explicit project directory.
func NewWithProject(projectDir string) (*App, error) {
	resolved, err := config.LoadWithProject(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration for project '%s': %w", projectDir, err)
	}

	return &App{
		resolved: resolved,
	}, nil
}

// NewWithConfig initializes an App container resolving configuration from an explicit config file.
func NewWithConfig(configPath string) (*App, error) {
	resolved, err := config.LoadWithConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration from '%s': %w", configPath, err)
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
	a.store = nil
	return nil
}

// StorageRoot returns the resolved storage root directory path.
func (a *App) StorageRoot() string {
	if a.resolved == nil {
		return ""
	}
	return a.resolved.StorageRoot
}

// ProjectDir returns the root project directory for the application.
func (a *App) ProjectDir() string {
	if a == nil || a.resolved == nil {
		return "."
	}
	if a.resolved.SourcePath != "" && a.resolved.Source == config.SourceProject {
		return filepath.Dir(a.resolved.SourcePath)
	}
	if a.resolved.StorageRoot != "" {
		clean := filepath.Clean(a.resolved.StorageRoot)
		if filepath.Base(clean) == ".gyrus" {
			return filepath.Dir(clean)
		}
		return clean
	}
	return "."
}

// Config returns the resolved project or global configuration settings.
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

	store, err := a.conceptStoreLocked()
	if err != nil {
		return nil, err
	}

	indexer, err := provider.NewIndexStore(a.resolved.Config, store)
	if err != nil {
		return nil, fmt.Errorf("failed initializing index provider: %w", err)
	}

	graph, err := provider.NewGraphStore(a.resolved.Config, store, indexer)
	if err != nil {
		return nil, fmt.Errorf("failed initializing graph provider: %w", err)
	}

	search, err := provider.NewSearchProvider(a.resolved.Config, store, indexer)
	if err != nil {
		return nil, fmt.Errorf("failed initializing search provider: %w", err)
	}

	projectName := ""
	if a.resolved != nil && a.resolved.Config != nil {
		projectName = a.resolved.Config.ProjectName()
	}
	a.engine = lifecycle.NewEngineWithProject(store, search, indexer, graph, a.resolved.StorageRoot, projectName)
	return a.engine, nil
}

// ConceptStore returns the same storage instance used by the lifecycle engine.
// Provider-specific commands can use optional capabilities without adding them
// to the application container or the common ConceptStore contract.
func (a *App) ConceptStore() (gyrus.ConceptStore, error) {
	a.engineMu.Lock()
	defer a.engineMu.Unlock()
	return a.conceptStoreLocked()
}

// conceptStoreLocked requires engineMu. Storage may be requested before Engine
// so invocation setup can select a provider session before constructing adapters.
func (a *App) conceptStoreLocked() (gyrus.ConceptStore, error) {
	if a.store != nil {
		return a.store, nil
	}
	resolved, err := config.ResolveProviders(a.resolved.Config)
	if err != nil {
		return nil, err
	}
	store, err := provider.NewConceptStore(resolved, a.resolved.StorageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed initializing storage provider: %w", err)
	}
	a.resolved.Config = resolved
	a.store = store
	return store, nil
}
