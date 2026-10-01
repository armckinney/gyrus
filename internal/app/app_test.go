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
	os.Unsetenv("GYRUS_WORKSPACE")
	os.Unsetenv("GYRUS_CONFIG")
	os.Exit(m.Run())
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that App service container instantiates dependencies and provides access to Engine.
// [Execution Surface]: In-Memory Dependency Container (internal/app)
// [Assertions]: App initializes without error, engine creates document, and returns valid DocumentRef.
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

	doc := gyrus.Document{
		ID:         "test-app-doc",
		Title:      "Test App Document",
		Category:   gyrus.CategoryTechnical,
		Type:       gyrus.TypeSpecification,
		OwnerGroup: "engineering",
		Version:    1,
		Status:     "draft",
		Content:    "# Test App Document",
	}

	ref, err := engine.Create(context.Background(), doc)
	if err != nil {
		t.Fatalf("Failed creating document via App engine: %v", err)
	}

	if ref.ID != "test-app-doc" {
		t.Errorf("Expected ref ID 'test-app-doc', got '%s'", ref.ID)
	}
}
