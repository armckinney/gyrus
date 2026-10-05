package app

import (
	"fmt"
	"path/filepath"
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

// NewWithWorkspace initializes an App container resolving configuration from an explicit workspace directory.
func NewWithWorkspace(workspaceDir string) (*App, error) {
	resolved, err := config.LoadWithWorkspace(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration for workspace '%s': %w", workspaceDir, err)
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
	return nil
}

// StorageRoot returns the resolved storage root directory path.
func (a *App) StorageRoot() string {
	if a.resolved == nil {
		return ""
	}
	return a.resolved.StorageRoot
}

// WorkspaceDir returns the root workspace directory for the application.
func (a *App) WorkspaceDir() string {
	if a == nil || a.resolved == nil {
		return "."
	}
	if a.resolved.SourcePath != "" && a.resolved.Source == config.SourceWorkspace {
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

	wsName := ""
	if a.resolved != nil && a.resolved.Config != nil {
		wsName = a.resolved.Config.WorkspaceName()
	}
	a.engine = lifecycle.NewEngineWithWorkspace(store, search, indexer, graph, a.resolved.StorageRoot, wsName)
	return a.engine, nil
}
