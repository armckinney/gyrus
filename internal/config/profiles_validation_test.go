package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/config"
	"gopkg.in/yaml.v3"
)

func TestUnknownSearchProviderFailsBeforeStorage(t *testing.T) {
	for _, data := range []string{
		"search:\n  provider: unsupported\n",
	} {
		t.Run(data, func(t *testing.T) {
			root := t.TempDir()
			name := filepath.Join(root, ".gyrus.yaml")
			original := "storage:\n  provider: localfs\n  root: must-not-exist\n" + data
			if err := os.WriteFile(name, []byte(original), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := config.LoadWithConfig(name); err == nil {
				t.Fatal("accepted unknown search provider")
			}
			if _, err := os.Stat(filepath.Join(root, "must-not-exist")); !os.IsNotExist(err) {
				t.Fatalf("storage touched: %v", err)
			}
			current, err := os.ReadFile(name)
			if err != nil || string(current) != original {
				t.Fatalf("configuration changed: %v", err)
			}
		})
	}
}

func TestUnknownProfilesDoNotOverwriteConfiguration(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	name := filepath.Join(root, ".gyrus.yaml")
	const original = "tenant: preserved\n"
	if err := os.WriteFile(name, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	for _, profile := range []config.Profile{"unsupported", ""} {
		if _, err := config.GenerateConfigYaml(profile, "root"); err == nil {
			t.Fatalf("generated unsupported profile %q", profile)
		}
		if _, err := config.WriteConfigFile(root, profile, "root"); err == nil {
			t.Fatalf("wrote unsupported profile %q", profile)
		}
		if _, err := config.WriteGlobalConfigFile(profile, "root", true); err == nil {
			t.Fatalf("wrote global unsupported profile %q", profile)
		}
		current, err := os.ReadFile(name)
		if err != nil || string(current) != original {
			t.Fatalf("configuration changed: %v", err)
		}
	}
}

func TestSupportedProfileConfigurationsResolve(t *testing.T) {
	for _, profile := range []config.Profile{config.ProfileLocal, config.ProfileGit, config.ProfileBlob, config.ProfileS3, config.ProfileAzure, config.ProfileGCS, config.ProfilePostgres} {
		content, err := config.GenerateConfigYaml(profile, "root", "demo")
		if err != nil {
			t.Fatal(err)
		}
		var cfg config.Config
		if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
			t.Fatal(err)
		}
		if _, err := config.ResolveProviders(&cfg); err != nil {
			t.Fatal(err)
		}
	}
}
