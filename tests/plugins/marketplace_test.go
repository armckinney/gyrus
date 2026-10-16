package plugins_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies that .claude-plugin/marketplace.json, .agents/plugins/marketplace.json,
//
//	and .github/plugin/marketplace.json are valid manifests pointing to packaging/plugins/gyrus.
//
// [Execution Surface]: Filesystem Repository Root
// [Assertions]: All marketplace manifests declare name 'armckinney' and resolve packaging/plugins/gyrus.
// -----------------------------------------------------------------------------
func TestMarketplaceManifests(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repository root: %v", err)
	}

	// 1. Claude Marketplace (.claude-plugin/marketplace.json)
	t.Run("ClaudeMarketplace", func(t *testing.T) {
		manifestPath := filepath.Join(repoRoot, ".claude-plugin", "marketplace.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", manifestPath, err)
		}

		var mp struct {
			Name    string `json:"name"`
			Plugins []struct {
				Name   string `json:"name"`
				Source string `json:"source"`
			} `json:"plugins"`
		}
		if err := json.Unmarshal(data, &mp); err != nil {
			t.Fatalf("Failed to parse %s as JSON: %v", manifestPath, err)
		}

		if mp.Name != "armckinney" {
			t.Errorf("Expected Claude marketplace name 'armckinney', got '%s'", mp.Name)
		}

		found := false
		for _, p := range mp.Plugins {
			if p.Name == "gyrus" {
				found = true
				if p.Source != "./packaging/plugins/gyrus" {
					t.Errorf("Expected source './packaging/plugins/gyrus', got '%s'", p.Source)
				}
				resolved := filepath.Join(repoRoot, p.Source)
				if fi, err := os.Stat(resolved); err != nil || !fi.IsDir() {
					t.Errorf("Plugin source directory does not exist: %s", resolved)
				}
			}
		}
		if !found {
			t.Errorf("Plugin 'gyrus' not found in Claude marketplace")
		}
	})

	// 2. OpenAI Codex Marketplace (.agents/plugins/marketplace.json)
	t.Run("CodexMarketplace", func(t *testing.T) {
		manifestPath := filepath.Join(repoRoot, ".agents", "plugins", "marketplace.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", manifestPath, err)
		}

		var mp struct {
			Name    string `json:"name"`
			Plugins []struct {
				Name   string `json:"name"`
				Source struct {
					Source string `json:"source"`
					Path   string `json:"path"`
				} `json:"source"`
			} `json:"plugins"`
		}
		if err := json.Unmarshal(data, &mp); err != nil {
			t.Fatalf("Failed to parse %s as JSON: %v", manifestPath, err)
		}

		if mp.Name != "armckinney" {
			t.Errorf("Expected Codex marketplace name 'armckinney', got '%s'", mp.Name)
		}

		found := false
		for _, p := range mp.Plugins {
			if p.Name == "gyrus" {
				found = true
				if p.Source.Path != "./packaging/plugins/gyrus" {
					t.Errorf("Expected source path './packaging/plugins/gyrus', got '%s'", p.Source.Path)
				}
				resolved := filepath.Join(repoRoot, p.Source.Path)
				if fi, err := os.Stat(resolved); err != nil || !fi.IsDir() {
					t.Errorf("Plugin source directory does not exist: %s", resolved)
				}
			}
		}
		if !found {
			t.Errorf("Plugin 'gyrus' not found in Codex marketplace")
		}
	})

	// 3. GitHub Copilot Marketplace (.github/plugin/marketplace.json)
	t.Run("CopilotMarketplace", func(t *testing.T) {
		manifestPath := filepath.Join(repoRoot, ".github", "plugin", "marketplace.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", manifestPath, err)
		}

		var mp struct {
			Name  string `json:"name"`
			Owner struct {
				Name string `json:"name"`
			} `json:"owner"`
			Interface struct {
				DisplayName string `json:"displayName"`
			} `json:"interface"`
			Plugins []struct {
				Name   string `json:"name"`
				Source string `json:"source"`
			} `json:"plugins"`
		}
		if err := json.Unmarshal(data, &mp); err != nil {
			t.Fatalf("Failed to parse %s as JSON: %v", manifestPath, err)
		}

		if mp.Name != "armckinney" {
			t.Errorf("Expected Copilot marketplace name 'armckinney', got '%s'", mp.Name)
		}
		if mp.Owner.Name != "Andrew McKinney" {
			t.Errorf("Expected Copilot marketplace owner name 'Andrew McKinney', got '%s'", mp.Owner.Name)
		}
		if mp.Interface.DisplayName != "armckinney-gyrus" {
			t.Errorf("Expected displayName 'armckinney-gyrus', got '%s'", mp.Interface.DisplayName)
		}

		found := false
		for _, p := range mp.Plugins {
			if p.Name == "gyrus" {
				found = true
				if p.Source != "./packaging/plugins/gyrus" {
					t.Errorf("Expected source './packaging/plugins/gyrus', got '%s'", p.Source)
				}
				resolved := filepath.Join(repoRoot, p.Source)
				if fi, err := os.Stat(resolved); err != nil || !fi.IsDir() {
					t.Errorf("Plugin source directory does not exist: %s", resolved)
				}
			}
		}
		if !found {
			t.Errorf("Plugin 'gyrus' not found in Copilot marketplace")
		}
	})

	// 4. Antigravity Plugins Configuration (.agents/plugins.json)
	t.Run("AntigravityPluginsConfig", func(t *testing.T) {
		configPath := filepath.Join(repoRoot, ".agents", "plugins.json")
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", configPath, err)
		}

		var cfg struct {
			Entries []struct {
				Path string `json:"path"`
			} `json:"entries"`
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			t.Fatalf("Failed to parse %s as JSON: %v", configPath, err)
		}

		found := false
		for _, e := range cfg.Entries {
			if e.Path == "packaging/plugins" {
				found = true
				resolved := filepath.Join(repoRoot, e.Path)
				if fi, err := os.Stat(resolved); err != nil || !fi.IsDir() {
					t.Errorf("Plugin repository directory does not exist: %s", resolved)
				}
			}
		}
		if !found {
			t.Errorf("packaging/plugins not found in .agents/plugins.json entries")
		}
	})
}
