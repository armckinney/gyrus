package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestMain(m *testing.M) {
	os.Unsetenv("GYRUS_PROJECT")
	os.Unsetenv("GYRUS_CONFIG")
	os.Exit(m.Run())
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that App service container instantiates dependencies and provides access to Engine.
// [Execution Surface]: In-Memory Dependency Container (internal/app)
// [Assertions]: App initializes without error, engine creates concept, and returns valid ConceptRef.
// -----------------------------------------------------------------------------
func TestAppContainerInstantiation(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	wsYaml := "storage:\n  provider: localfs\n  root: .gyrus\n"
	if err := os.WriteFile(filepath.Join(tempDir, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing .gyrus.yaml: %v", err)
	}

	application, err := app.New()
	if err != nil {
		t.Fatalf("Failed creating app container: %v", err)
	}

	engine, err := application.Engine()
	if err != nil {
		t.Fatalf("Failed instantiating lifecycle engine: %v", err)
	}

	concept := gyrus.Concept{
		ID:       "test-app-concept",
		Title:    "Test App Concept",
		Category: gyrus.CategoryTechnical,
		Type:     gyrus.TypeSpecification,
		Tenant:   "engineering",
		Version:  1,
		Status:   "draft",
		Content:  "# Test App Concept", Name: "engineering/reference/test-app-concept",
	}

	ref, err := engine.Create(context.Background(), concept)
	if err != nil {
		t.Fatalf("Failed creating concept via App engine: %v", err)
	}

	if ref.ID != "test-app-concept" {
		t.Errorf("Expected ref ID 'test-app-concept', got '%s'", ref.ID)
	}
}

func TestAppNewWithProject(t *testing.T) {
	tempDir := t.TempDir()
	wsYaml := "project: custom-project\ntenant: custom-tenant\nstorage:\n  provider: localfs\n  root: .gyrus\n"
	if err := os.WriteFile(filepath.Join(tempDir, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing .gyrus.yaml: %v", err)
	}

	application, err := app.NewWithProject(tempDir)
	if err != nil {
		t.Fatalf("Failed creating app container with project: %v", err)
	}

	if application.ProjectDir() != tempDir {
		t.Errorf("Expected ProjectDir '%s', got '%s'", tempDir, application.ProjectDir())
	}
	if application.Config().ProjectName() != "custom-project" {
		t.Errorf("Expected ProjectName 'custom-project', got '%s'", application.Config().ProjectName())
	}
	if application.Config().TenantName() != "custom-tenant" {
		t.Errorf("Expected TenantName 'custom-tenant', got '%s'", application.Config().TenantName())
	}
}
