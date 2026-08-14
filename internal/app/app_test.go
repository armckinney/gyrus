package app_test

import (
	"context"
	"testing"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that App service container instantiates dependencies and provides access to Engine.
// [Execution Surface]: In-Memory Dependency Container (internal/app)
// [Assertions]: App initializes without error, engine creates document, and returns valid DocumentRef.
// -----------------------------------------------------------------------------
func TestAppContainerInstantiation(t *testing.T) {
	tempDir := t.TempDir()

	application, err := app.New(tempDir)
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
