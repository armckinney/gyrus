package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load resolves the full Gyrus configuration using the precedence chain:
//  1. Workspace config (.gyrus.yaml) — searched from current working directory upward
//  2. Global config (~/.gyrus.yaml)
//  3. Hardcoded defaults
//
// Full override semantics: if a workspace config file exists, the global
// config is ignored entirely.
func Load() (*ResolvedConfig, error) {
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	}
	home, _ := os.UserHomeDir()
	return LoadFrom(pwd, home)
}

// LoadFrom resolves configuration relative to explicit working and home directories.
func LoadFrom(workingDir, homeDir string) (*ResolvedConfig, error) {
	// 1. Try to find workspace config (walk workingDir -> filesystem root)
	if cfg, path, err := findWorkspaceConfig(workingDir); err == nil && cfg != nil {
		if err := validate(cfg, path); err != nil {
			return nil, err
		}
		storageRoot := resolveStorageRoot(cfg, filepath.Dir(path))
		return &ResolvedConfig{
			Config:      cfg,
			Source:      SourceWorkspace,
			SourcePath:  path,
			StorageRoot: storageRoot,
		}, nil
	}

	// 2. Try global config (~/.gyrus.yaml)
	if cfg, path, err := findGlobalConfig(homeDir); err == nil && cfg != nil {
		if err := validate(cfg, path); err != nil {
			return nil, err
		}
		storageRoot := resolveStorageRoot(cfg, homeDir)
		return &ResolvedConfig{
			Config:      cfg,
			Source:      SourceGlobal,
			SourcePath:  path,
			StorageRoot: storageRoot,
		}, nil
	}

	// 3. Fallback to defaults
	defaultRoot := filepath.Join(homeDir, DefaultStorageRoot)
	if homeDir == "" {
		if abs, err := filepath.Abs("./" + DefaultStorageRoot); err == nil {
			defaultRoot = abs
		} else {
			defaultRoot = "./" + DefaultStorageRoot
		}
	} else if abs, err := filepath.Abs(defaultRoot); err == nil {
		defaultRoot = abs
	}

	return &ResolvedConfig{
		Config:      &Config{},
		Source:      SourceDefault,
		SourcePath:  "",
		StorageRoot: defaultRoot,
	}, nil
}

// findWorkspaceConfig walks parent directories looking for .gyrus.yaml.
// If a file is malformed, it returns nil to allow fallback.
func findWorkspaceConfig(startDir string) (*Config, string, error) {
	absStart, err := filepath.Abs(startDir)
	if err != nil {
		absStart = startDir
	}

	curr := absStart
	for {
		candidate := filepath.Join(curr, GlobalConfigFileName)
		if cfg, path, err := loadConfigFile(candidate); err == nil && cfg != nil {
			return cfg, path, nil
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return nil, "", nil
}

// findGlobalConfig checks ~/.gyrus.yaml in the specified home directory.
func findGlobalConfig(homeDir string) (*Config, string, error) {
	if homeDir == "" {
		return nil, "", nil
	}
	candidate := filepath.Join(homeDir, GlobalConfigFileName)
	return loadConfigFile(candidate)
}

// loadConfigFile reads and unmarshals a single YAML config file.
func loadConfigFile(path string) (*Config, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, "", err
	}

	return &cfg, path, nil
}

// resolveStorageRoot determines the absolute storage root path.
// Relative paths resolve relative to baseDir.
func resolveStorageRoot(cfg *Config, baseDir string) string {
	rawPath := ""
	if cfg != nil {
		rawPath = cfg.StorageRoot()
	}
	if rawPath == "" {
		rawPath = DefaultStorageRoot
	}

	resolved, err := expandAndAbsRelative(rawPath, baseDir)
	if err != nil {
		return filepath.Join(baseDir, DefaultStorageRoot)
	}
	return resolved
}

func expandAndAbsRelative(path string, baseDir string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
		return filepath.Abs(path)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}
	return filepath.Abs(path)
}

func validate(cfg *Config, sourcePath string) error {
	if cfg == nil {
		return nil
	}
	if p := cfg.StorageProvider(); p != "" && !isValidProvider(p, ValidStorageProviders) {
		return fmt.Errorf("unknown storage provider '%s' in %s. Valid: %s",
			p, sourcePath, strings.Join(ValidStorageProviders, ", "))
	}
	if p := cfg.IndexProvider(); p != "" && !isValidProvider(p, ValidIndexProviders) {
		return fmt.Errorf("unknown index provider '%s' in %s. Valid: %s",
			p, sourcePath, strings.Join(ValidIndexProviders, ", "))
	}
	if p := cfg.SearchProvider(); p != "" && !isValidProvider(p, ValidSearchProviders) {
		return fmt.Errorf("unknown search provider '%s' in %s. Valid: %s",
			p, sourcePath, strings.Join(ValidSearchProviders, ", "))
	}
	if p := cfg.GraphProvider(); p != "" && !isValidProvider(p, ValidGraphProviders) {
		return fmt.Errorf("unknown graph provider '%s' in %s. Valid: %s",
			p, sourcePath, strings.Join(ValidGraphProviders, ", "))
	}
	return nil
}

func isValidProvider(target string, valid []string) bool {
	for _, v := range valid {
		if target == v {
			return true
		}
	}
	return false
}
