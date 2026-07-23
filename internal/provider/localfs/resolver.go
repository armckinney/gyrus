package localfs

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SchemasPath string `yaml:"schemas_path"`
	Storage     struct {
		Root string `yaml:"root"`
	} `yaml:"storage"`
}

// ResolveStoragePath evaluates storage root precedence hierarchy:
// 1. Explicit CLI flag argument (--storage-path)
// 2. GYRUS_STORAGE_PATH env var
// 3. Project config (.gyrus.yaml, .gyrus.yml, .gyrus/config.yaml, .gyrus/config.yml) in PWD or parent directories
// 4. User home config (~/.config/gyrus/config.yaml, ~/.gyrus.yaml)
// 5. Default ~/.gyrus/ fallback
func ResolveStoragePath(flagPath string) (string, error) {
	if flagPath != "" {
		return filepath.Abs(flagPath)
	}

	if envPath := os.Getenv("GYRUS_STORAGE_PATH"); envPath != "" {
		return filepath.Abs(envPath)
	}

	// Walk parent directories from PWD to find repository config file
	pwd, err := os.Getwd()
	if err == nil {
		curr := pwd
		for {
			configCandidates := []string{
				filepath.Join(curr, ".gyrus.yaml"),
				filepath.Join(curr, ".gyrus.yml"),
				filepath.Join(curr, ".gyrus", "config.yaml"),
				filepath.Join(curr, ".gyrus", "config.yml"),
			}

			for _, candidate := range configCandidates {
				if data, err := os.ReadFile(candidate); err == nil {
					var cfg Config
					if err := yaml.Unmarshal(data, &cfg); err == nil && cfg.Storage.Root != "" {
						return expandAndAbsRelative(cfg.Storage.Root, curr)
					}
				}
			}

			parent := filepath.Dir(curr)
			if parent == curr {
				break
			}
			curr = parent
		}
	}

	// Check user home config
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userCandidates := []string{
			filepath.Join(homeDir, ".config", "gyrus", "config.yaml"),
			filepath.Join(homeDir, ".config", "gyrus", "config.yml"),
			filepath.Join(homeDir, ".gyrus.yaml"),
			filepath.Join(homeDir, ".gyrus.yml"),
		}
		for _, userCfgPath := range userCandidates {
			if data, err := os.ReadFile(userCfgPath); err == nil {
				var cfg Config
				if err := yaml.Unmarshal(data, &cfg); err == nil && cfg.Storage.Root != "" {
					return expandAndAbsRelative(cfg.Storage.Root, homeDir)
				}
			}
		}

		// Fallback to ~/.gyrus/
		return filepath.Join(homeDir, ".gyrus"), nil
	}

	return filepath.Abs("./.gyrus")
}

// ResolveSchemasPath evaluates custom schemas directory path from config files.
func ResolveSchemasPath() (string, error) {
	pwd, err := os.Getwd()
	if err == nil {
		curr := pwd
		for {
			configCandidates := []string{
				filepath.Join(curr, ".gyrus.yaml"),
				filepath.Join(curr, ".gyrus.yml"),
				filepath.Join(curr, ".gyrus", "config.yaml"),
				filepath.Join(curr, ".gyrus", "config.yml"),
			}

			for _, candidate := range configCandidates {
				if data, err := os.ReadFile(candidate); err == nil {
					var cfg Config
					if err := yaml.Unmarshal(data, &cfg); err == nil && cfg.SchemasPath != "" {
						return expandAndAbsRelative(cfg.SchemasPath, curr)
					}
				}
			}

			parent := filepath.Dir(curr)
			if parent == curr {
				break
			}
			curr = parent
		}
	}

	return "", nil
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

