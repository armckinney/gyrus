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
owner_group: platform-engineering
version: 1
status: proposed
last_modified_by: developer-joe
tags:
  - storage
  - sqlite
dependencies:
  - prd-context-engine
---

# SQLite Edge Graph Provider

This document specifies the SQLite edge graph schema.
`

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that valid OKF YAML frontmatter and markdown body are accurately parsed into a gyrus.Document struct.
// [Execution Surface]: In-Memory Package (internal/domain/okf)
// [Assertions]: Document fields (ID, Type, Dependencies) are extracted and validation passes.
// -----------------------------------------------------------------------------
func TestParseMarkdownValid(t *testing.T) {
	doc, err := okf.ParseMarkdown([]byte(sampleMarkdown))
	if err != nil {
		t.Fatalf("Unexpected parse error: %v", err)
	}

	if doc.ID != "adr-2026-sqlite-edges" {
		t.Errorf("Expected ID 'adr-2026-sqlite-edges', got '%s'", doc.ID)
	}
	if doc.Type != gyrus.TypeADR {
		t.Errorf("Expected Type 'adr', got '%s'", doc.Type)
	}
	if len(doc.Dependencies) != 1 || doc.Dependencies[0] != "prd-context-engine" {
		t.Errorf("Expected dependency 'prd-context-engine', got %v", doc.Dependencies)
	}

	if err := okf.Validate(doc); err != nil {
		t.Errorf("Document validation failed: %v", err)
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
// [Assertions]: Validate rejects nil documents, missing mandatory fields, invalid categories, invalid types, version < 1, and malformed IDs.
// -----------------------------------------------------------------------------
func TestValidateSchemaConstraintsMatrix(t *testing.T) {
	validDoc := func() *gyrus.Document {
		return &gyrus.Document{
			ID:         "valid-doc-id_01",
			Title:      "Valid Title",
			Category:   gyrus.CategoryArchitecture,
			Type:       gyrus.TypeADR,
			OwnerGroup: "engineering",
			Version:    1,
			Status:     "proposed",
		}
	}

	// 1. Nil Document
	if err := okf.Validate(nil); err == nil {
		t.Errorf("Expected error for nil document, got nil")
	}

	tests := []struct {
		name        string
		modify      func(d *gyrus.Document)
		shouldError bool
	}{
		{
			name:        "Valid Document",
			modify:      func(d *gyrus.Document) {},
			shouldError: false,
		},
		{
			name:        "Missing ID",
			modify:      func(d *gyrus.Document) { d.ID = "" },
			shouldError: true,
		},
		{
			name:        "Invalid ID with spaces",
			modify:      func(d *gyrus.Document) { d.ID = "invalid doc id" },
			shouldError: true,
		},
		{
			name:        "Invalid ID with uppercase letters",
			modify:      func(d *gyrus.Document) { d.ID = "INVALID-UPPER" },
			shouldError: true,
		},
		{
			name:        "Invalid ID with special characters",
			modify:      func(d *gyrus.Document) { d.ID = "doc#123!" },
			shouldError: true,
		},
		{
			name:        "Valid ID with dashes and underscores",
			modify:      func(d *gyrus.Document) { d.ID = "doc-123_valid" },
			shouldError: false,
		},
		{
			name:        "Missing Title",
			modify:      func(d *gyrus.Document) { d.Title = "" },
			shouldError: true,
		},
		{
			name:        "Missing OwnerGroup",
			modify:      func(d *gyrus.Document) { d.OwnerGroup = "" },
			shouldError: true,
		},
		{
			name:        "Invalid Category",
			modify:      func(d *gyrus.Document) { d.Category = "non-existent-category" },
			shouldError: true,
		},
		{
			name:        "Invalid Type",
			modify:      func(d *gyrus.Document) { d.Type = "non-existent-type" },
			shouldError: true,
		},
		{
			name:        "Version Zero",
			modify:      func(d *gyrus.Document) { d.Version = 0 },
			shouldError: true,
		},
		{
			name:        "Negative Version",
			modify:      func(d *gyrus.Document) { d.Version = -1 },
			shouldError: true,
		},
		{
			name:        "Missing Status",
			modify:      func(d *gyrus.Document) { d.Status = "" },
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := validDoc()
			tt.modify(doc)
			err := okf.Validate(doc)
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
// [Assertions]: Serializing a Document and re-parsing it yields an identical Document structure.
// -----------------------------------------------------------------------------
func TestSerializeAndReParseMarkdown(t *testing.T) {
	originalDoc := &gyrus.Document{
		ID:             "spec-01-schema",
		Title:          "OKF Schema Envelope",
		Category:       gyrus.CategoryTechnical,
		Type:           gyrus.TypeSpecification,
		OwnerGroup:     "architecture",
		Version:        2,
		Status:         "active",
		LastModifiedBy: "architect-alice",
		Tags:           []string{"okf", "schema"},
		Content:        "## Summary\n\nComplete schema details.",
	}

	serialized, err := okf.SerializeMarkdown(originalDoc)
	if err != nil {
		t.Fatalf("Failed to serialize Markdown: %v", err)
	}

	parsed, err := okf.ParseMarkdown(serialized)
	if err != nil {
		t.Fatalf("Failed to re-parse serialized Markdown: %v", err)
	}

	if parsed.ID != originalDoc.ID {
		t.Errorf("Expected ID %s, got %s", originalDoc.ID, parsed.ID)
	}
	if parsed.Content != originalDoc.Content {
		t.Errorf("Expected content:\n%s\nGot:\n%s", originalDoc.Content, parsed.Content)
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
owner_group: ops
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

	// Invalid docType
	if err := okf.ValidateSchemaContent("../bad", validContent); err == nil {
		t.Errorf("expected error for bad docType, got nil")
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
