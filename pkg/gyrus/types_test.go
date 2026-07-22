package gyrus_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"gopkg.in/yaml.v3"
)

func TestDocumentJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	doc := gyrus.Document{
		ID:             "adr-2026-001",
		Title:          "Use SQLite Edge Tables for Gyrus Relationships",
		Category:       gyrus.CategoryArchitecture,
		Type:           gyrus.TypeADR,
		Format:         "markdown",
		OwnerGroup:     "platform-engineering",
		Version:        1,
		Status:         "accepted",
		LastModifiedBy: "developer-joe",
		LastUpdated:    now,
		Tags:           []string{"storage", "sqlite"},
		Dependencies:   []string{"prd-context-manager"},
		Content:        "# Test Content",
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Failed to marshal document to JSON: %v", err)
	}

	var unmarshaled gyrus.Document
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal document from JSON: %v", err)
	}

	if unmarshaled.ID != doc.ID {
		t.Errorf("Expected ID %s, got %s", doc.ID, unmarshaled.ID)
	}
	if unmarshaled.Type != gyrus.TypeADR {
		t.Errorf("Expected Type %s, got %s", gyrus.TypeADR, unmarshaled.Type)
	}
}

func TestDocumentYAMLFrontmatterSerialization(t *testing.T) {
	doc := gyrus.Document{
		ID:             "prd-2026-001",
		Title:          "Context Manager PRD",
		Category:       gyrus.CategoryProduct,
		Type:           gyrus.TypePRD,
		Format:         "markdown",
		OwnerGroup:     "product",
		Version:        1,
		Status:         "draft",
		LastModifiedBy: "pm-jane",
		Tags:           []string{"prd", "specs"},
	}

	data, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("Failed to marshal document to YAML: %v", err)
	}

	var unmarshaled gyrus.Document
	if err := yaml.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal document from YAML: %v", err)
	}

	if unmarshaled.Title != doc.Title {
		t.Errorf("Expected Title %s, got %s", doc.Title, unmarshaled.Title)
	}
	if unmarshaled.Category != gyrus.CategoryProduct {
		t.Errorf("Expected Category %s, got %s", gyrus.CategoryProduct, unmarshaled.Category)
	}
}
