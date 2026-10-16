package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load resolves the full Gyrus configuration using the precedence chain:
//  0. Explicit config file via GYRUS_CONFIG env var
//  1. Explicit project directory via GYRUS_PROJECT env var
//  2. Project config (.gyrus.yaml) — searched from current working directory upward
//  3. DevContainer auto-detection (/workspaces/*) if working directory is outside project
//  4. Global config (~/.gyrus.yaml)
//  5. Hardcoded defaults
//
// Full override semantics: if a project config file exists, the global
// config is ignored entirely.
func Load() (*ResolvedConfig, error) {
	// 0. Check GYRUS_CONFIG
	if envCfg := os.Getenv("GYRUS_CONFIG"); envCfg != "" {
		return LoadWithConfig(envCfg)
	}

	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	}
	inTemp := isTempDirectory(pwd)

	// 1. Check GYRUS_PROJECT (ignored if working directory is an isolated temporary directory and GYRUS_PROJECT points outside it)
	if envProject := os.Getenv("GYRUS_PROJECT"); envProject != "" {
		if !inTemp || isTempDirectory(envProject) {
			home, _ := os.UserHomeDir()
			res, err := LoadFrom(envProject, home)
			if err != nil {
				return nil, err
			}
			if res.Source == SourceProject {
				return res, nil
			}
		}
	}

	// 2. Search upward from working directory
	home, _ := os.UserHomeDir()
	res, err := LoadFrom(pwd, home)
	if err != nil {
		return nil, err
	}
	if res.Source == SourceProject {
		return res, nil
	}

	// 3. DevContainer auto-detection fallback if running inside container and not in a temp directory
	if isContainerEnvironment() && !inTemp {
		if ws := findDevcontainerProject(); ws != "" {
			if devRes, err := LoadFrom(ws, home); err == nil && devRes.Source == SourceProject {
				return devRes, nil
			}
		}
	}

	return res, nil
}

func isTempDirectory(path string) bool {
	tmp := os.TempDir()
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	if strings.HasPrefix(absPath, "/tmp") || strings.HasPrefix(absPath, "/private/tmp") || (tmp != "" && strings.HasPrefix(absPath, tmp)) {
		return true
	}
	return false
}

// LoadWithProject loads configuration with an explicit project directory.
func LoadWithProject(projectDir string) (*ResolvedConfig, error) {
	home, _ := os.UserHomeDir()
	return LoadFrom(projectDir, home)
}

// LoadWithConfig loads configuration from an explicit config file path.
func LoadWithConfig(configPath string) (*ResolvedConfig, error) {
	cfg, path, err := loadConfigFile(configPath)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("configuration file '%s' could not be loaded", configPath)
	}
	if err := validate(cfg, path); err != nil {
		return nil, err
	}
	return &ResolvedConfig{
		Config:      cfg,
		Source:      SourceProject,
		SourcePath:  path,
		StorageRoot: resolveStorageRoot(cfg, filepath.Dir(path)),
	}, nil
}

func isContainerEnvironment() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	if os.Getenv("REMOTE_CONTAINERS") == "true" || os.Getenv("CODESPACES") == "true" {
		return true
	}
	return false
}

func findDevcontainerProject() string {
	entries, err := os.ReadDir("/workspaces")
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			candidate := filepath.Join("/workspaces", e.Name())
			if _, err := os.Stat(filepath.Join(candidate, GlobalConfigFileName)); err == nil {
				return candidate
			}
		}
	}
	return ""
}

// LoadFrom resolves configuration relative to explicit working and home directories.
func LoadFrom(workingDir, homeDir string) (*ResolvedConfig, error) {
	// 1. Try to find project config (walk workingDir -> filesystem root)
	if cfg, path, err := findProjectConfig(workingDir); err == nil && cfg != nil {
		if err := validate(cfg, path); err != nil {
			return nil, err
		}
		storageRoot := resolveStorageRoot(cfg, filepath.Dir(path))
		return &ResolvedConfig{
			Config:      cfg,
			Source:      SourceProject,
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
		Config:      defaultResolvedConfig(),
		Source:      SourceDefault,
		SourcePath:  "",
		StorageRoot: defaultRoot,
	}, nil
}

// findProjectConfig walks parent directories looking for .gyrus.yaml.
// If a file is malformed, it returns nil to allow fallback.
func findProjectConfig(startDir string) (*Config, string, error) {
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
	resolved, err := ResolveProviders(cfg)
	if err != nil {
		return fmt.Errorf("%s: %w", sourcePath, err)
	}
	if cfg == nil {
		return nil
	}
	*cfg = *resolved
	if cfg.Index.Provider == "sqlite" && cfg.Index.DSN != "" {
		cfg.Index.DSN, err = expandAndAbsRelative(cfg.Index.DSN, filepath.Dir(sourcePath))
		if err != nil {
			return fmt.Errorf("resolve index.dsn: %w", err)
		}
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
