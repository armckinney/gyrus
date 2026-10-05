package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/config"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that when no workspace or global config exists, defaults are used.
// [Assertions]: Source is SourceDefault, StorageRoot points to ~/.gyrus.
// -----------------------------------------------------------------------------
func TestLoadDefaultsOnly(t *testing.T) {
	tempWork := t.TempDir()
	tempHome := t.TempDir()

	rc, err := config.LoadFrom(tempWork, tempHome)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if rc.Source != config.SourceDefault {
		t.Errorf("Expected source '%s', got '%s'", config.SourceDefault, rc.Source)
	}
	expectedRoot := filepath.Join(tempHome, ".gyrus")
	if rc.StorageRoot != expectedRoot {
		t.Errorf("Expected storage root '%s', got '%s'", expectedRoot, rc.StorageRoot)
	}
	if rc.Config.StorageProvider() != "" {
		t.Errorf("Expected empty default storage provider, got '%s'", rc.Config.StorageProvider())
	}
	if rc.Config.OwnerGroup() != "root" {
		t.Errorf("Expected default owner group 'root', got '%s'", rc.Config.OwnerGroup())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that global ~/.gyrus.yaml is used when no workspace config exists.
// [Assertions]: Source is SourceGlobal, values loaded from ~/.gyrus.yaml.
// -----------------------------------------------------------------------------
func TestLoadGlobalOnly(t *testing.T) {
	tempWork := t.TempDir()
	tempHome := t.TempDir()

	globalYaml := `storage:
  provider: postgres
  root: .gyrus
default_owner_group: global-ops
postgres:
  connection_string: postgres://user:pass@localhost:5432/gyrus
`
	if err := os.WriteFile(filepath.Join(tempHome, ".gyrus.yaml"), []byte(globalYaml), 0644); err != nil {
		t.Fatalf("Failed writing global config: %v", err)
	}

	rc, err := config.LoadFrom(tempWork, tempHome)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if rc.Source != config.SourceGlobal {
		t.Errorf("Expected source '%s', got '%s'", config.SourceGlobal, rc.Source)
	}
	if rc.Config.StorageProvider() != "postgres" {
		t.Errorf("Expected storage provider 'postgres', got '%s'", rc.Config.StorageProvider())
	}
	if rc.Config.DefaultOwnerGroup != "global-ops" {
		t.Errorf("Expected owner group 'global-ops', got '%s'", rc.Config.DefaultOwnerGroup)
	}
	expectedRoot := filepath.Join(tempHome, ".gyrus")
	if rc.StorageRoot != expectedRoot {
		t.Errorf("Expected storage root '%s', got '%s'", expectedRoot, rc.StorageRoot)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that workspace .gyrus.yaml is loaded when present.
// [Assertions]: Source is SourceWorkspace, values loaded from workspace .gyrus.yaml.
// -----------------------------------------------------------------------------
func TestLoadWorkspaceOnly(t *testing.T) {
	tempWork := t.TempDir()
	tempHome := t.TempDir()

	wsYaml := `storage:
  provider: localfs
  root: .my-store
index:
  provider: sqlite
search:
  provider: sqlite
default_owner_group: workspace-team
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing workspace config: %v", err)
	}

	rc, err := config.LoadFrom(tempWork, tempHome)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if rc.Source != config.SourceWorkspace {
		t.Errorf("Expected source '%s', got '%s'", config.SourceWorkspace, rc.Source)
	}
	if rc.Config.StorageProvider() != "localfs" {
		t.Errorf("Expected storage provider 'localfs', got '%s'", rc.Config.StorageProvider())
	}
	if rc.Config.DefaultOwnerGroup != "workspace-team" {
		t.Errorf("Expected owner group 'workspace-team', got '%s'", rc.Config.DefaultOwnerGroup)
	}
	expectedRoot := filepath.Join(tempWork, ".my-store")
	if rc.StorageRoot != expectedRoot {
		t.Errorf("Expected storage root '%s', got '%s'", expectedRoot, rc.StorageRoot)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies full override semantics: workspace config completely overrides global config.
// [Assertions]: Source is SourceWorkspace, global-only values are NOT merged into workspace config.
// -----------------------------------------------------------------------------
func TestLoadWorkspaceOverGlobal(t *testing.T) {
	tempWork := t.TempDir()
	tempHome := t.TempDir()

	// Global defines postgres and global-ops
	globalYaml := `storage:
  provider: postgres
default_owner_group: global-ops
postgres:
  connection_string: postgres://localhost:5432/global
`
	if err := os.WriteFile(filepath.Join(tempHome, ".gyrus.yaml"), []byte(globalYaml), 0644); err != nil {
		t.Fatalf("Failed writing global config: %v", err)
	}

	// Workspace defines localfs and local-team (does NOT set postgres)
	wsYaml := `storage:
  provider: localfs
default_owner_group: local-team
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing workspace config: %v", err)
	}

	rc, err := config.LoadFrom(tempWork, tempHome)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if rc.Source != config.SourceWorkspace {
		t.Errorf("Expected source '%s', got '%s'", config.SourceWorkspace, rc.Source)
	}
	if rc.Config.StorageProvider() != "localfs" {
		t.Errorf("Expected storage provider 'localfs', got '%s'", rc.Config.StorageProvider())
	}
	if rc.Config.DefaultOwnerGroup != "local-team" {
		t.Errorf("Expected owner group 'local-team', got '%s'", rc.Config.DefaultOwnerGroup)
	}
	// Full override: global postgres connection string is NOT merged into workspace
	if rc.Config.Postgres.ConnectionString != "" {
		t.Errorf("Expected empty postgres connection string due to full override, got '%s'", rc.Config.Postgres.ConnectionString)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that malformed global YAML falls back gracefully to defaults.
// [Assertions]: Source is SourceDefault, no fatal error.
// -----------------------------------------------------------------------------
func TestLoadMalformedGlobal(t *testing.T) {
	tempWork := t.TempDir()
	tempHome := t.TempDir()

	badYaml := `: this is not valid: yaml : [unclosed`
	if err := os.WriteFile(filepath.Join(tempHome, ".gyrus.yaml"), []byte(badYaml), 0644); err != nil {
		t.Fatalf("Failed writing bad global config: %v", err)
	}

	rc, err := config.LoadFrom(tempWork, tempHome)
	if err != nil {
		t.Fatalf("Expected graceful fallback, got error: %v", err)
	}

	if rc.Source != config.SourceDefault {
		t.Errorf("Expected source '%s', got '%s'", config.SourceDefault, rc.Source)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that malformed workspace config falls through to valid global config.
// [Assertions]: Source is SourceGlobal.
// -----------------------------------------------------------------------------
func TestLoadMalformedWorkspace(t *testing.T) {
	tempWork := t.TempDir()
	tempHome := t.TempDir()

	badYaml := `: this is not valid: yaml : [unclosed`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(badYaml), 0644); err != nil {
		t.Fatalf("Failed writing bad workspace config: %v", err)
	}

	goodGlobal := `storage:
  provider: localfs
default_owner_group: fallback-global
`
	if err := os.WriteFile(filepath.Join(tempHome, ".gyrus.yaml"), []byte(goodGlobal), 0644); err != nil {
		t.Fatalf("Failed writing good global config: %v", err)
	}

	rc, err := config.LoadFrom(tempWork, tempHome)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if rc.Source != config.SourceGlobal {
		t.Errorf("Expected source '%s', got '%s'", config.SourceGlobal, rc.Source)
	}
	if rc.Config.DefaultOwnerGroup != "fallback-global" {
		t.Errorf("Expected owner group 'fallback-global', got '%s'", rc.Config.DefaultOwnerGroup)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that unknown provider names fail validation with a descriptive error.
// [Assertions]: Returns error naming invalid provider and listing valid ones.
// -----------------------------------------------------------------------------
func TestLoadInvalidProvider(t *testing.T) {
	tempWork := t.TempDir()
	tempHome := t.TempDir()

	invalidYaml := `storage:
  provider: invalid_mongodb
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(invalidYaml), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}

	_, err := config.LoadFrom(tempWork, tempHome)
	if err == nil {
		t.Fatalf("Expected validation error for invalid provider, got nil")
	}

	if !strings.Contains(err.Error(), "unknown storage provider 'invalid_mongodb'") {
		t.Errorf("Expected error to mention unknown provider, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Valid: localfs") {
		t.Errorf("Expected error to list valid providers, got: %v", err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that relative storage.root paths resolve relative to config file directory.
// [Assertions]: StorageRoot is resolved accurately relative to config file directory.
// -----------------------------------------------------------------------------
func TestRelativePathResolution(t *testing.T) {
	tempWork := t.TempDir()
	subDir := filepath.Join(tempWork, "subproject")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed creating subdir: %v", err)
	}
	tempHome := t.TempDir()

	wsYaml := `storage:
  provider: localfs
  root: ./custom-store
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}

	// Load starting from subDir — should find parent's .gyrus.yaml and resolve relative to tempWork
	rc, err := config.LoadFrom(subDir, tempHome)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	expectedRoot := filepath.Join(tempWork, "custom-store")
	if rc.StorageRoot != expectedRoot {
		t.Errorf("Expected storage root '%s', got '%s'", expectedRoot, rc.StorageRoot)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that GYRUS_WORKSPACE environment variable overrides working directory.
// [Assertions]: Load() resolves .gyrus.yaml from GYRUS_WORKSPACE directory.
// -----------------------------------------------------------------------------
func TestLoad_EnvGyrusWorkspace(t *testing.T) {
	tempWork := t.TempDir()
	wsYaml := `storage:
  provider: localfs
  root: .gyrus
default_owner_group: env-ws-group
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}

	t.Setenv("GYRUS_WORKSPACE", tempWork)

	rc, err := config.Load()
	if err != nil {
		t.Fatalf("Load with GYRUS_WORKSPACE failed: %v", err)
	}
	if rc.Source != config.SourceWorkspace {
		t.Errorf("Expected source '%s', got '%s'", config.SourceWorkspace, rc.Source)
	}
	if rc.Config.OwnerGroup() != "env-ws-group" {
		t.Errorf("Expected owner group 'env-ws-group', got '%s'", rc.Config.OwnerGroup())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that GYRUS_CONFIG environment variable points directly to config file.
// [Assertions]: Load() resolves directly from GYRUS_CONFIG path.
// -----------------------------------------------------------------------------
func TestLoad_EnvGyrusConfig(t *testing.T) {
	tempWork := t.TempDir()
	cfgPath := filepath.Join(tempWork, "custom-config.yaml")
	cfgYaml := `storage:
  provider: localfs
  root: ./my-storage
default_owner_group: custom-cfg-group
`
	if err := os.WriteFile(cfgPath, []byte(cfgYaml), 0644); err != nil {
		t.Fatalf("Failed writing custom config: %v", err)
	}

	t.Setenv("GYRUS_CONFIG", cfgPath)

	rc, err := config.Load()
	if err != nil {
		t.Fatalf("Load with GYRUS_CONFIG failed: %v", err)
	}
	if rc.Source != config.SourceWorkspace {
		t.Errorf("Expected source '%s', got '%s'", config.SourceWorkspace, rc.Source)
	}
	if rc.Config.OwnerGroup() != "custom-cfg-group" {
		t.Errorf("Expected owner group 'custom-cfg-group', got '%s'", rc.Config.OwnerGroup())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that LoadWithWorkspace explicitly resolves the specified workspace.
// [Assertions]: LoadWithWorkspace loads .gyrus.yaml from target directory.
// -----------------------------------------------------------------------------
func TestLoadWithWorkspace(t *testing.T) {
	tempWork := t.TempDir()
	wsYaml := `storage:
  provider: localfs
  root: .gyrus
default_owner_group: explicit-ws
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}

	rc, err := config.LoadWithWorkspace(tempWork)
	if err != nil {
		t.Fatalf("LoadWithWorkspace failed: %v", err)
	}
	if rc.Config.OwnerGroup() != "explicit-ws" {
		t.Errorf("Expected owner group 'explicit-ws', got '%s'", rc.Config.OwnerGroup())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that explicit workspace configuration is parsed and accessible.
// [Assertions]: Config.WorkspaceName() matches configured workspace or empty default.
// -----------------------------------------------------------------------------
func TestWorkspaceConfiguration(t *testing.T) {
	tempWork := t.TempDir()
	wsYaml := `storage:
  provider: localfs
  root: .gyrus
workspace: gyrus-core
default_owner_group: root
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}

	rc, err := config.LoadWithWorkspace(tempWork)
	if err != nil {
		t.Fatalf("LoadWithWorkspace failed: %v", err)
	}
	if rc.Config.WorkspaceName() != "gyrus-core" {
		t.Errorf("Expected workspace 'gyrus-core', got '%s'", rc.Config.WorkspaceName())
	}

	// Test empty workspace
	tempWork2 := t.TempDir()
	wsYaml2 := `storage:
  provider: localfs
`
	if err := os.WriteFile(filepath.Join(tempWork2, ".gyrus.yaml"), []byte(wsYaml2), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}
	rc2, err := config.LoadWithWorkspace(tempWork2)
	if err != nil {
		t.Fatalf("LoadWithWorkspace failed: %v", err)
	}
	if rc2.Config.WorkspaceName() != "" {
		t.Errorf("Expected empty workspace, got '%s'", rc2.Config.WorkspaceName())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies automation config loading, defaults, and overrides.
// [Assertions]: Default values are true, explicit false values override defaults.
// -----------------------------------------------------------------------------
func TestAutomationConfiguration(t *testing.T) {
	// 1. Defaults when automation section is omitted
	tempWork1 := t.TempDir()
	wsYaml1 := `storage:
  provider: localfs
`
	if err := os.WriteFile(filepath.Join(tempWork1, ".gyrus.yaml"), []byte(wsYaml1), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}
	rc1, err := config.LoadWithWorkspace(tempWork1)
	if err != nil {
		t.Fatalf("LoadWithWorkspace failed: %v", err)
	}
	if !rc1.Config.HooksEnabled() {
		t.Errorf("Expected default HooksEnabled() to be true")
	}
	if !rc1.Config.AutoContext() {
		t.Errorf("Expected default AutoContext() to be true")
	}
	if !rc1.Config.ArchitecturalCheck() {
		t.Errorf("Expected default ArchitecturalCheck() to be true")
	}
	if len(rc1.Config.IgnoredPaths()) == 0 {
		t.Errorf("Expected default IgnoredPaths() to be non-empty")
	}

	// 2. Explicit overrides
	tempWork2 := t.TempDir()
	wsYaml2 := `storage:
  provider: localfs
automation:
  hooks_enabled: false
  auto_context: false
  architectural_check: false
  ignored_paths:
    - "custom/**"
`
	if err := os.WriteFile(filepath.Join(tempWork2, ".gyrus.yaml"), []byte(wsYaml2), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}
	rc2, err := config.LoadWithWorkspace(tempWork2)
	if err != nil {
		t.Fatalf("LoadWithWorkspace failed: %v", err)
	}
	if rc2.Config.HooksEnabled() {
		t.Errorf("Expected HooksEnabled() to be false when configured false")
	}
	if rc2.Config.AutoContext() {
		t.Errorf("Expected AutoContext() to be false when configured false")
	}
	if rc2.Config.ArchitecturalCheck() {
		t.Errorf("Expected ArchitecturalCheck() to be false when configured false")
	}
	if len(rc2.Config.IgnoredPaths()) != 1 || rc2.Config.IgnoredPaths()[0] != "custom/**" {
		t.Errorf("Expected IgnoredPaths() to contain 'custom/**', got %v", rc2.Config.IgnoredPaths())
	}
}
