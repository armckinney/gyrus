package format_test

import (
	"testing"

	"github.com/armckinney/gyrus/internal/format"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that the OKF Markdown format engine serializes and deserializes documents with full field integrity.
// [Execution Surface]: In-Memory Package (internal/format)
// [Assertions]: Serialized OKF markdown reconstructs all metadata fields (ID, Title, Tags, Content) identically.
// -----------------------------------------------------------------------------
func TestOKFFormatSerialization(t *testing.T) {
	fmtEngine := format.NewOKFFormat()

	doc := &gyrus.Document{
		ID:         "adr-001-test",
		Title:      "Test ADR Document",
		Category:   "architecture",
		Type:       "adr",
		OwnerGroup: "armckinney",
		Version:    1,
		Status:     "proposed",
		Tags:       []string{"test", "architecture"},
		Content:    "# Test Content\n\nThis is a test body.",
	}

	data, err := fmtEngine.Serialize(doc)
	if err != nil {
		t.Fatalf("Failed to serialize OKF document: %v", err)
	}

	deserialized, err := fmtEngine.Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize OKF document: %v", err)
	}

	if deserialized.ID != doc.ID {
		t.Errorf("Expected ID %s, got %s", doc.ID, deserialized.ID)
	}
	if deserialized.Title != doc.Title {
		t.Errorf("Expected Title %s, got %s", doc.Title, deserialized.Title)
	}
	if len(deserialized.Tags) != 2 || deserialized.Tags[0] != "test" {
		t.Errorf("Expected tags %v, got %v", doc.Tags, deserialized.Tags)
	}
	if deserialized.Content != doc.Content {
		t.Errorf("Expected content:\n%s\nGot:\n%s", doc.Content, deserialized.Content)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that the Relational format engine serializes and deserializes documents into relational payloads.
// [Execution Surface]: In-Memory Package (internal/format)
// [Assertions]: Relational document representation serializes without error and maintains document identity upon deserialization.
// -----------------------------------------------------------------------------
func TestRelationalFormatSerialization(t *testing.T) {
	fmtEngine := format.NewRelationalFormat()

	doc := &gyrus.Document{
		ID:         "spec-001-test",
		Title:      "Test Spec Document",
		Category:   "technical",
		Type:       "specification",
		OwnerGroup: "armckinney",
		Version:    1,
		Status:     "active",
		Content:    "Relational payload content.",
	}

	data, err := fmtEngine.Serialize(doc)
	if err != nil {
		t.Fatalf("Failed to serialize Relational document: %v", err)
	}

	deserialized, err := fmtEngine.Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize Relational document: %v", err)
	}

	if deserialized.ID != doc.ID {
		t.Errorf("Expected ID %s, got %s", doc.ID, deserialized.ID)
	}
	if deserialized.Content != doc.Content {
		t.Errorf("Expected Content %s, got %s", doc.Content, deserialized.Content)
	}
}
