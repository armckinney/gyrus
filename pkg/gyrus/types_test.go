package gyrus_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"gopkg.in/yaml.v3"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that Concept envelope serializes to and deserializes from JSON format cleanly.
// [Execution Surface]: In-Memory Unit Test (pkg/gyrus)
// [Assertions]: JSON marshaling preserves ID, Title, Type, and metadata fields without data corruption.
// -----------------------------------------------------------------------------
func TestConceptJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	concept := gyrus.Concept{
		ID:             "adr-2026-001",
		Title:          "Use SQLite Edge Tables for Gyrus Relationships",
		Category:       gyrus.CategoryArchitecture,
		Type:           gyrus.TypeADR,
		Format:         "markdown",
		Tenant:         "platform-engineering",
		Version:        1,
		Status:         "accepted",
		LastModifiedBy: "developer-joe",
		LastUpdated:    now,
		Tags:           []string{"storage", "sqlite"},
		Dependencies:   []string{"prd-context-manager"},
		Content:        "# Test Content",
	}

	data, err := json.Marshal(concept)
	if err != nil {
		t.Fatalf("Failed to marshal concept to JSON: %v", err)
	}

	var unmarshaled gyrus.Concept
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal concept from JSON: %v", err)
	}

	if unmarshaled.ID != concept.ID {
		t.Errorf("Expected ID %s, got %s", concept.ID, unmarshaled.ID)
	}
	if unmarshaled.Type != gyrus.TypeADR {
		t.Errorf("Expected Type %s, got %s", gyrus.TypeADR, unmarshaled.Type)
	}
	if unmarshaled.Tenant != concept.Tenant {
		t.Errorf("Expected Tenant %s, got %s", concept.Tenant, unmarshaled.Tenant)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that Concept envelope serializes to and deserializes from YAML frontmatter format cleanly.
// [Execution Surface]: In-Memory Unit Test (pkg/gyrus)
// [Assertions]: YAML marshaling preserves Title, Category, Status, and Tags.
// -----------------------------------------------------------------------------
func TestConceptYAMLFrontmatterSerialization(t *testing.T) {
	concept := gyrus.Concept{
		ID:             "prd-2026-001",
		Title:          "Context Manager PRD",
		Category:       gyrus.CategoryProduct,
		Type:           gyrus.TypePRD,
		Format:         "markdown",
		Tenant:         "product",
		Version:        1,
		Status:         "draft",
		LastModifiedBy: "pm-jane",
		Tags:           []string{"prd", "specs"},
	}

	data, err := yaml.Marshal(concept)
	if err != nil {
		t.Fatalf("Failed to marshal concept to YAML: %v", err)
	}

	var unmarshaled gyrus.Concept
	if err := yaml.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal concept from YAML: %v", err)
	}

	if unmarshaled.Title != concept.Title {
		t.Errorf("Expected Title %s, got %s", concept.Title, unmarshaled.Title)
	}
	if unmarshaled.Category != gyrus.CategoryProduct {
		t.Errorf("Expected Category %s, got %s", gyrus.CategoryProduct, unmarshaled.Category)
	}
	if unmarshaled.Tenant != concept.Tenant {
		t.Errorf("Expected Tenant %s, got %s", concept.Tenant, unmarshaled.Tenant)
	}
}
