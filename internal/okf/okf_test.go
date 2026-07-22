package okf_test

import (
	"testing"

	"github.com/armckinney/gyrus/internal/okf"
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

func TestParseMarkdownInvalidHeader(t *testing.T) {
	invalidMD := "# Just Markdown without frontmatter"
	_, err := okf.ParseMarkdown([]byte(invalidMD))
	if err == nil {
		t.Fatal("Expected error for missing frontmatter header, got nil")
	}
}

func TestValidateInvalidIDPattern(t *testing.T) {
	doc := &gyrus.Document{
		ID:         "INVALID ID WITH SPACES!",
		Title:      "Test Doc",
		Category:   gyrus.CategoryArchitecture,
		Type:       gyrus.TypeADR,
		OwnerGroup: "dev",
		Version:    1,
		Status:     "proposed",
	}

	err := okf.Validate(doc)
	if err == nil {
		t.Fatal("Expected validation error for invalid ID, got nil")
	}
}

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
