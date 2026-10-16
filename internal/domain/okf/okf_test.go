package okf_test

import (
	"testing"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

const sampleMarkdown = `---
id: adr-2026-sqlite-edges
title: SQLite Edge Graph Provider
category: architecture
type: adr
tenant: platform-engineering
version: 1
status: proposed
last_modified_by: developer-joe
tags:
  - storage
  - sqlite
links:
  - concept: platform-engineering/reference/prd-context-engine
    rel: depends_on
---

# SQLite Edge Graph Provider

This concept specifies the SQLite edge graph schema.
`

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that valid OKF YAML frontmatter and markdown body are accurately parsed into a gyrus.Concept struct.
// [Execution Surface]: In-Memory Package (internal/domain/okf)
// [Assertions]: Concept fields (ID, Type, Dependencies, Tenant) are extracted and validation passes.
// -----------------------------------------------------------------------------
func TestParseMarkdownValid(t *testing.T) {
	concept, err := okf.ParseMarkdown([]byte(sampleMarkdown))
	if err != nil {
		t.Fatalf("Unexpected parse error: %v", err)
	}

	if concept.ID != "adr-2026-sqlite-edges" {
		t.Errorf("Expected ID 'adr-2026-sqlite-edges', got '%s'", concept.ID)
	}
	if concept.Type != gyrus.TypeADR {
		t.Errorf("Expected Type 'adr', got '%s'", concept.Type)
	}
	if concept.Tenant != "platform-engineering" {
		t.Errorf("Expected Tenant 'platform-engineering', got '%s'", concept.Tenant)
	}
	if len(concept.Links) != 1 || concept.Links[0].Concept != "platform-engineering/reference/prd-context-engine" {
		t.Errorf("Expected dependency 'prd-context-engine', got %v", concept.Links)
	}

	if err := okf.Validate(concept); err != nil {
		t.Errorf("Concept validation failed: %v", err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that parsing markdown lacking OKF YAML frontmatter delimiter returns an error.
// [Execution Surface]: In-Memory Package (internal/domain/okf)
// [Assertions]: ParseMarkdown returns a non-nil error when frontmatter is missing.
// -----------------------------------------------------------------------------
func TestParseMarkdownInvalidHeader(t *testing.T) {
	invalidMD := "# Just Markdown without frontmatter"
	_, err := okf.ParseMarkdown([]byte(invalidMD))
	if err == nil {
		t.Fatal("Expected error for missing frontmatter header, got nil")
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Exhaustively validates all mandatory OKF schema boundaries and constraints.
// [Execution Surface]: In-Memory Package (internal/domain/okf)
// [Assertions]: Validate rejects nil concepts, missing mandatory fields, invalid categories, invalid types, version < 1, and malformed IDs.
// -----------------------------------------------------------------------------
func TestValidateSchemaConstraintsMatrix(t *testing.T) {
	validConcept := func() *gyrus.Concept {
		return &gyrus.Concept{
			ID:       "valid-concept-id_01",
			Title:    "Valid Title",
			Category: gyrus.CategoryArchitecture,
			Type:     gyrus.TypeADR,
			Tenant:   "engineering",
			Version:  1,
			Status:   "proposed", Name: "engineering/reference/valid-concept-id_01",
		}
	}

	// 1. Nil Concept
	if err := okf.Validate(nil); err == nil {
		t.Errorf("Expected error for nil concept, got nil")
	}

	tests := []struct {
		name        string
		modify      func(d *gyrus.Concept)
		shouldError bool
	}{
		{
			name:        "Valid Concept",
			modify:      func(d *gyrus.Concept) {},
			shouldError: false,
		},
		{
			name:        "Missing ID",
			modify:      func(d *gyrus.Concept) { d.ID = "" },
			shouldError: true,
		},
		{
			name:        "Invalid ID with spaces",
			modify:      func(d *gyrus.Concept) { d.ID = "invalid concept id" },
			shouldError: true,
		},
		{
			name:        "Invalid ID with uppercase letters",
			modify:      func(d *gyrus.Concept) { d.ID = "INVALID-UPPER" },
			shouldError: true,
		},
		{
			name:        "Invalid ID with special characters",
			modify:      func(d *gyrus.Concept) { d.ID = "concept#123!" },
			shouldError: true,
		},
		{
			name:        "Valid ID with dashes and underscores",
			modify:      func(d *gyrus.Concept) { d.ID = "concept-123_valid" },
			shouldError: false,
		},
		{
			name:        "Missing Title",
			modify:      func(d *gyrus.Concept) { d.Title = "" },
			shouldError: true,
		},
		{
			name:        "Missing Tenant",
			modify:      func(d *gyrus.Concept) { d.Tenant = "" },
			shouldError: true,
		},
		{
			name:        "Invalid Category",
			modify:      func(d *gyrus.Concept) { d.Category = "non-existent-category" },
			shouldError: true,
		},
		{
			name:        "Invalid Type",
			modify:      func(d *gyrus.Concept) { d.Type = "non-existent-type" },
			shouldError: true,
		},
		{
			name:        "Version Zero",
			modify:      func(d *gyrus.Concept) { d.Version = 0 },
			shouldError: true,
		},
		{
			name:        "Negative Version",
			modify:      func(d *gyrus.Concept) { d.Version = -1 },
			shouldError: true,
		},
		{
			name:        "Missing Status",
			modify:      func(d *gyrus.Concept) { d.Status = "" },
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			concept := validConcept()
			tt.modify(concept)
			err := okf.Validate(concept)
			if tt.shouldError && err == nil {
				t.Errorf("Expected validation error for '%s', got nil", tt.name)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected validation error for '%s': %v", tt.name, err)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies round-trip fidelity between Markdown serialization and deserialization.
// [Execution Surface]: In-Memory Package (internal/domain/okf)
// [Assertions]: Serializing a Concept and re-parsing it yields an identical Concept structure.
// -----------------------------------------------------------------------------
func TestSerializeAndReParseMarkdown(t *testing.T) {
	originalConcept := &gyrus.Concept{
		ID:             "spec-01-schema",
		Title:          "OKF Schema Envelope",
		Category:       gyrus.CategoryTechnical,
		Type:           gyrus.TypeSpecification,
		Tenant:         "architecture",
		Version:        2,
		Status:         "active",
		LastModifiedBy: "architect-alice",
		Tags:           []string{"okf", "schema"},
		Content:        "## Summary\n\nComplete schema details.", Name: "architecture/reference/spec-01-schema",
	}

	serialized, err := okf.SerializeMarkdown(originalConcept)
	if err != nil {
		t.Fatalf("Failed to serialize Markdown: %v", err)
	}

	parsed, err := okf.ParseMarkdown(serialized)
	if err != nil {
		t.Fatalf("Failed to re-parse serialized Markdown: %v", err)
	}

	if parsed.ID != originalConcept.ID {
		t.Errorf("Expected ID %s, got %s", originalConcept.ID, parsed.ID)
	}
	if parsed.Tenant != originalConcept.Tenant {
		t.Errorf("Expected Tenant %s, got %s", originalConcept.Tenant, parsed.Tenant)
	}
	if parsed.Content != originalConcept.Content {
		t.Errorf("Expected content:\n%s\nGot:\n%s", originalConcept.Content, parsed.Content)
	}
}

func TestSchemaValidationAndListing(t *testing.T) {
	// 1. ListEmbeddedTemplates
	templates := okf.ListEmbeddedTemplates()
	if len(templates) == 0 {
		t.Fatalf("expected embedded templates, got none")
	}
	hasADR := false
	for _, tmpl := range templates {
		if tmpl == "adr" {
			hasADR = true
			break
		}
	}
	if !hasADR {
		t.Errorf("expected 'adr' in embedded templates list, got: %v", templates)
	}

	// 2. ValidateSchemaPath
	validPaths := []string{
		".gyrus/schemas/adr.md",
		"/path/to/project/.gyrus/schemas/custom.md",
		"prefix/.gyrus/schemas/runbook.md",
	}
	for _, p := range validPaths {
		if err := okf.ValidateSchemaPath(p); err != nil {
			t.Errorf("expected path %q to be valid, got: %v", p, err)
		}
	}

	invalidPaths := []string{
		"../schemas/adr.md",
		".gyrus/schemas/adr.txt",
		".gyrus/other/adr.md",
		".gyrus/schemas/Invalid_Name.md",
	}
	for _, p := range invalidPaths {
		if err := okf.ValidateSchemaPath(p); err == nil {
			t.Errorf("expected path %q to be invalid, got nil", p)
		}
	}

	// 3. ValidateSchemaContent
	validContent := `---
id: <unique-id>
title: <Title>
category: operations
type: runbook
tenant: ops
version: 1
status: active
---

# Runbook
`
	if err := okf.ValidateSchemaContent("runbook", validContent); err != nil {
		t.Errorf("expected content to be valid, got: %v", err)
	}

	// Mismatched type
	if err := okf.ValidateSchemaContent("different-type", validContent); err == nil {
		t.Errorf("expected error for mismatched type, got nil")
	}

	// Invalid empty content
	if err := okf.ValidateSchemaContent("runbook", ""); err == nil {
		t.Errorf("expected error for empty content, got nil")
	}

	// Invalid conceptType
	if err := okf.ValidateSchemaContent("../bad", validContent); err == nil {
		t.Errorf("expected error for bad conceptType, got nil")
	}

	// 4. GetTemplate returns error for non-existent schema
	if _, err := okf.GetTemplate("nonexistent-type", ""); err == nil {
		t.Errorf("expected error for non-existent template, got nil")
	}

	// GetTemplate succeeds for embedded schema
	if tmpl, err := okf.GetTemplate("adr", ""); err != nil || tmpl == "" {
		t.Errorf("expected embedded adr template, got err: %v", err)
	}
}
