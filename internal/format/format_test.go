package format_test

import (
	"testing"

	"github.com/armckinney/gyrus/internal/format"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that the OKF Markdown format engine serializes and deserializes concepts with full field integrity.
// [Execution Surface]: In-Memory Package (internal/format)
// [Assertions]: Serialized OKF markdown reconstructs all metadata fields (ID, Title, Tags, Content) identically.
// -----------------------------------------------------------------------------
func TestOKFFormatSerialization(t *testing.T) {
	fmtEngine := format.NewOKFFormat()

	concept := &gyrus.Concept{
		ID:       "adr-001-test",
		Title:    "Test ADR Concept",
		Category: "architecture",
		Type:     "adr",
		Tenant:   "root",
		Version:  1,
		Status:   "proposed",
		Tags:     []string{"test", "architecture"},
		Content:  "# Test Content\n\nThis is a test body.", Name: "root/reference/adr-001-test",
	}

	data, err := fmtEngine.Serialize(concept)
	if err != nil {
		t.Fatalf("Failed to serialize OKF concept: %v", err)
	}

	deserialized, err := fmtEngine.Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize OKF concept: %v", err)
	}

	if deserialized.ID != concept.ID {
		t.Errorf("Expected ID %s, got %s", concept.ID, deserialized.ID)
	}
	if deserialized.Title != concept.Title {
		t.Errorf("Expected Title %s, got %s", concept.Title, deserialized.Title)
	}
	if len(deserialized.Tags) != 2 || deserialized.Tags[0] != "test" {
		t.Errorf("Expected tags %v, got %v", concept.Tags, deserialized.Tags)
	}
	if deserialized.Tenant != concept.Tenant {
		t.Errorf("Expected Tenant %s, got %s", concept.Tenant, deserialized.Tenant)
	}
	if deserialized.Content != concept.Content {
		t.Errorf("Expected content:\n%s\nGot:\n%s", concept.Content, deserialized.Content)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that the Relational format engine serializes and deserializes concepts into relational payloads.
// [Execution Surface]: In-Memory Package (internal/format)
// [Assertions]: Relational concept representation serializes without error and maintains concept identity upon deserialization.
// -----------------------------------------------------------------------------
func TestRelationalFormatSerialization(t *testing.T) {
	fmtEngine := format.NewRelationalFormat()

	concept := &gyrus.Concept{
		ID:       "spec-001-test",
		Title:    "Test Spec Concept",
		Category: "technical",
		Type:     "specification",
		Tenant:   "root",
		Version:  1,
		Status:   "active",
		Content:  "Relational payload content.", Name: "root/reference/spec-001-test",
	}

	data, err := fmtEngine.Serialize(concept)
	if err != nil {
		t.Fatalf("Failed to serialize Relational concept: %v", err)
	}

	deserialized, err := fmtEngine.Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize Relational concept: %v", err)
	}

	if deserialized.ID != concept.ID {
		t.Errorf("Expected ID %s, got %s", concept.ID, deserialized.ID)
	}
	if deserialized.Tenant != concept.Tenant {
		t.Errorf("Expected Tenant %s, got %s", concept.Tenant, deserialized.Tenant)
	}
	if deserialized.Content != concept.Content {
		t.Errorf("Expected Content %s, got %s", concept.Content, deserialized.Content)
	}
}
