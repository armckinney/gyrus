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
// [Purpose]: Verifies that when no project or global config exists, defaults are used.
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
	if rc.Config.StorageProvider() != "localfs" {
		t.Errorf("Expected localfs default storage provider, got '%s'", rc.Config.StorageProvider())
	}
	if rc.Config.TenantName() != "root" {
		t.Errorf("Expected default tenant 'root', got '%s'", rc.Config.TenantName())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that global ~/.gyrus.yaml is used when no project config exists.
// [Assertions]: Source is SourceGlobal, values loaded from ~/.gyrus.yaml.
// -----------------------------------------------------------------------------
func TestLoadGlobalOnly(t *testing.T) {
	tempWork := t.TempDir()
	tempHome := t.TempDir()

	globalYaml := `storage:
  provider: postgres
  root: .gyrus
tenant: global-ops
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
	if rc.Config.Tenant != "global-ops" {
		t.Errorf("Expected tenant 'global-ops', got '%s'", rc.Config.Tenant)
	}
	expectedRoot := filepath.Join(tempHome, ".gyrus")
	if rc.StorageRoot != expectedRoot {
		t.Errorf("Expected storage root '%s', got '%s'", expectedRoot, rc.StorageRoot)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that project .gyrus.yaml is loaded when present.
// [Assertions]: Source is SourceProject, values loaded from project .gyrus.yaml.
// -----------------------------------------------------------------------------
func TestLoadProjectOnly(t *testing.T) {
	tempWork := t.TempDir()
	tempHome := t.TempDir()

	wsYaml := `storage:
  provider: localfs
  root: .my-store
index:
  provider: okf
search:
  provider: okf
tenant: project-team
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing project config: %v", err)
	}

	rc, err := config.LoadFrom(tempWork, tempHome)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if rc.Source != config.SourceProject {
		t.Errorf("Expected source '%s', got '%s'", config.SourceProject, rc.Source)
	}
	if rc.Config.StorageProvider() != "localfs" {
		t.Errorf("Expected storage provider 'localfs', got '%s'", rc.Config.StorageProvider())
	}
	if rc.Config.Tenant != "project-team" {
		t.Errorf("Expected tenant 'project-team', got '%s'", rc.Config.Tenant)
	}
	expectedRoot := filepath.Join(tempWork, ".my-store")
	if rc.StorageRoot != expectedRoot {
		t.Errorf("Expected storage root '%s', got '%s'", expectedRoot, rc.StorageRoot)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies full override semantics: project config completely overrides global config.
// [Assertions]: Source is SourceProject, global-only values are NOT merged into project config.
// -----------------------------------------------------------------------------
func TestLoadProjectOverGlobal(t *testing.T) {
	tempWork := t.TempDir()
	tempHome := t.TempDir()

	// Global defines postgres and global-ops
	globalYaml := `storage:
  provider: postgres
tenant: global-ops
postgres:
  connection_string: postgres://localhost:5432/global
`
	if err := os.WriteFile(filepath.Join(tempHome, ".gyrus.yaml"), []byte(globalYaml), 0644); err != nil {
		t.Fatalf("Failed writing global config: %v", err)
	}

	// Project defines localfs and local-team (does NOT set postgres)
	wsYaml := `storage:
  provider: localfs
tenant: local-team
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing project config: %v", err)
	}

	rc, err := config.LoadFrom(tempWork, tempHome)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if rc.Source != config.SourceProject {
		t.Errorf("Expected source '%s', got '%s'", config.SourceProject, rc.Source)
	}
	if rc.Config.StorageProvider() != "localfs" {
		t.Errorf("Expected storage provider 'localfs', got '%s'", rc.Config.StorageProvider())
	}
	if rc.Config.Tenant != "local-team" {
		t.Errorf("Expected tenant 'local-team', got '%s'", rc.Config.Tenant)
	}
	// Full override: global postgres connection string is NOT merged into project
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
// [Purpose]: Verifies that malformed project config falls through to valid global config.
// [Assertions]: Source is SourceGlobal.
// -----------------------------------------------------------------------------
func TestLoadMalformedProject(t *testing.T) {
	tempWork := t.TempDir()
	tempHome := t.TempDir()

	badYaml := `: this is not valid: yaml : [unclosed`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(badYaml), 0644); err != nil {
		t.Fatalf("Failed writing bad project config: %v", err)
	}

	goodGlobal := `storage:
  provider: localfs
tenant: fallback-global
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
	if rc.Config.Tenant != "fallback-global" {
		t.Errorf("Expected tenant 'fallback-global', got '%s'", rc.Config.Tenant)
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
// [Purpose]: Verifies that GYRUS_PROJECT environment variable overrides working directory.
// [Assertions]: Load() resolves .gyrus.yaml from GYRUS_PROJECT directory.
// -----------------------------------------------------------------------------
func TestLoad_EnvGyrusProject(t *testing.T) {
	tempWork := t.TempDir()
	wsYaml := `storage:
  provider: localfs
  root: .gyrus
tenant: env-ws-group
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}

	t.Setenv("GYRUS_PROJECT", tempWork)

	rc, err := config.Load()
	if err != nil {
		t.Fatalf("Load with GYRUS_PROJECT failed: %v", err)
	}
	if rc.Source != config.SourceProject {
		t.Errorf("Expected source '%s', got '%s'", config.SourceProject, rc.Source)
	}
	if rc.Config.TenantName() != "env-ws-group" {
		t.Errorf("Expected tenant 'env-ws-group', got '%s'", rc.Config.TenantName())
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
tenant: custom-cfg-group
`
	if err := os.WriteFile(cfgPath, []byte(cfgYaml), 0644); err != nil {
		t.Fatalf("Failed writing custom config: %v", err)
	}

	t.Setenv("GYRUS_CONFIG", cfgPath)

	rc, err := config.Load()
	if err != nil {
		t.Fatalf("Load with GYRUS_CONFIG failed: %v", err)
	}
	if rc.Source != config.SourceProject {
		t.Errorf("Expected source '%s', got '%s'", config.SourceProject, rc.Source)
	}
	if rc.Config.TenantName() != "custom-cfg-group" {
		t.Errorf("Expected tenant 'custom-cfg-group', got '%s'", rc.Config.TenantName())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that LoadWithProject explicitly resolves the specified project.
// [Assertions]: LoadWithProject loads .gyrus.yaml from target directory.
// -----------------------------------------------------------------------------
func TestLoadWithProject(t *testing.T) {
	tempWork := t.TempDir()
	wsYaml := `storage:
  provider: localfs
  root: .gyrus
tenant: explicit-ws
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}

	rc, err := config.LoadWithProject(tempWork)
	if err != nil {
		t.Fatalf("LoadWithProject failed: %v", err)
	}
	if rc.Config.TenantName() != "explicit-ws" {
		t.Errorf("Expected tenant 'explicit-ws', got '%s'", rc.Config.TenantName())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that explicit project configuration is parsed and accessible.
// [Assertions]: Config.ProjectName() matches configured project or empty default.
// -----------------------------------------------------------------------------
func TestProjectConfiguration(t *testing.T) {
	tempWork := t.TempDir()
	wsYaml := `storage:
  provider: localfs
  root: .gyrus
project: gyrus-core
tenant: root
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}

	rc, err := config.LoadWithProject(tempWork)
	if err != nil {
		t.Fatalf("LoadWithProject failed: %v", err)
	}
	if rc.Config.ProjectName() != "gyrus-core" {
		t.Errorf("Expected project 'gyrus-core', got '%s'", rc.Config.ProjectName())
	}

	// Test empty project
	tempWork2 := t.TempDir()
	wsYaml2 := `storage:
  provider: localfs
`
	if err := os.WriteFile(filepath.Join(tempWork2, ".gyrus.yaml"), []byte(wsYaml2), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}
	rc2, err := config.LoadWithProject(tempWork2)
	if err != nil {
		t.Fatalf("LoadWithProject failed: %v", err)
	}
	if rc2.Config.ProjectName() != "" {
		t.Errorf("Expected empty project, got '%s'", rc2.Config.ProjectName())
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that RunConfigSetup creates .gyrus.yaml with designated profile and tenant.
// [Execution Surface]: In-Memory / Temporary Filesystem
// [Assertions]: Creates .gyrus.yaml in project and returns correct status fields.
// -----------------------------------------------------------------------------
func TestRunConfigSetup(t *testing.T) {
	tempDir := t.TempDir()

	res, err := config.RunConfigSetup(config.ConfigSetupOptions{
		ProjectDir: tempDir,
		Profile:    config.ProfileLocal,
		Tenant:     "test-group",
		Project:    "test-project",
	})
	if err != nil {
		t.Fatalf("RunConfigSetup failed: %v", err)
	}

	configPath := filepath.Join(tempDir, ".gyrus.yaml")
	if res.ConfigFile != configPath {
		t.Errorf("Expected config file %s, got %s", configPath, res.ConfigFile)
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected .gyrus.yaml to exist at %s", configPath)
	}

	expectedStorageDir := filepath.Join(tempDir, ".gyrus")
	if res.StorageDir != expectedStorageDir {
		t.Errorf("Expected storage root path %s, got %s", expectedStorageDir, res.StorageDir)
	}
	if res.Tenant != "test-group" {
		t.Errorf("Expected tenant 'test-group', got '%s'", res.Tenant)
	}
	if res.Project != "test-project" {
		t.Errorf("Expected project 'test-project', got '%s'", res.Project)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed reading config: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 || lines[1] != "tenant: test-group" || lines[2] != "project: test-project" {
		t.Errorf("Expected tenant and project at top of config, got:\n%s", string(data))
	}

	// Verify default project fallback
	tempDir2 := t.TempDir()
	res2, err := config.RunConfigSetup(config.ConfigSetupOptions{
		ProjectDir: tempDir2,
		Profile:    config.ProfileLocal,
	})
	if err != nil {
		t.Fatalf("RunConfigSetup with defaults failed: %v", err)
	}
	if res2.Tenant != config.DefaultTenant {
		t.Errorf("Expected default tenant '%s', got '%s'", config.DefaultTenant, res2.Tenant)
	}
	if res2.Project != config.DefaultProject {
		t.Errorf("Expected default project '%s', got '%s'", config.DefaultProject, res2.Project)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that WriteConfigFile generates correct configurations across all supported profiles.
// -----------------------------------------------------------------------------
func TestWriteConfigFileProfiles(t *testing.T) {
	profiles := []struct {
		profile config.Profile
		keyword string
	}{
		{config.ProfileLocal, "provider: localfs"},
		{config.ProfileGit, "provider: git"},
		{config.ProfileBlob, "provider: blob"},
		{config.ProfileS3, "provider: s3"},
		{config.ProfileAzure, "provider: azure_blob"},
		{config.ProfileGCS, "provider: gcs"},
		{config.ProfilePostgres, "provider: postgres"},
	}

	for _, p := range profiles {
		t.Run(string(p.profile), func(t *testing.T) {
			tempDir := t.TempDir()
			// Default project ("default")
			cfgPath, err := config.WriteConfigFile(tempDir, p.profile, "my-team")
			if err != nil {
				t.Fatalf("WriteConfigFile failed for %s: %v", p.profile, err)
			}

			data, err := os.ReadFile(cfgPath)
			if err != nil {
				t.Fatalf("Failed reading generated config: %v", err)
			}

			content := string(data)
			if !strings.Contains(content, p.keyword) {
				t.Errorf("Expected config to contain '%s', got:\n%s", p.keyword, content)
			}
			lines := strings.Split(content, "\n")
			if len(lines) < 3 || lines[1] != "tenant: my-team" || lines[2] != "project: default" {
				t.Errorf("Expected tenant and default project at top of config, got:\n%s", content)
			}

			// Explicit project ("my-app")
			tempDirCustom := t.TempDir()
			cfgPathCustom, err := config.WriteConfigFile(tempDirCustom, p.profile, "my-team", "my-app")
			if err != nil {
				t.Fatalf("WriteConfigFile with explicit project failed for %s: %v", p.profile, err)
			}
			dataCustom, err := os.ReadFile(cfgPathCustom)
			if err != nil {
				t.Fatalf("Failed reading generated config: %v", err)
			}
			linesCustom := strings.Split(string(dataCustom), "\n")
			if len(linesCustom) < 3 || linesCustom[1] != "tenant: my-team" || linesCustom[2] != "project: my-app" {
				t.Errorf("Expected tenant and custom project at top of config, got:\n%s", string(dataCustom))
			}
		})
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that WriteGlobalConfigFile writes ~/.gyrus.yaml and honors the force flag.
// [Assertions]: Initial write succeeds, second write fails without force, second write succeeds with force.
// -----------------------------------------------------------------------------
func TestWriteGlobalConfigFile(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	// 1. First write succeeds
	path, err := config.WriteGlobalConfigFile(config.ProfileLocal, "global-team", false)
	if err != nil {
		t.Fatalf("First WriteGlobalConfigFile failed: %v", err)
	}
	expectedPath := filepath.Join(tempHome, ".gyrus.yaml")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}

	// 2. Second write without force fails
	_, err = config.WriteGlobalConfigFile(config.ProfileLocal, "global-team", false)
	if err == nil {
		t.Fatalf("Expected error when overwriting without force, got nil")
	}

	// 3. Second write with force succeeds
	_, err = config.WriteGlobalConfigFile(config.ProfilePostgres, "global-team", true)
	if err != nil {
		t.Fatalf("Expected force overwrite to succeed, got: %v", err)
	}
	data, _ := os.ReadFile(expectedPath)
	if !strings.Contains(string(data), "provider: postgres") {
		t.Errorf("Expected overwritten config to contain postgres, got: %s", string(data))
	}
}
