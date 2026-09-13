package lifecycle_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies the valid and invalid lifecycle state transition rules for Architecture Design Records (ADR).
// [Execution Surface]: In-Memory Package (internal/domain/lifecycle)
// [Assertions]: Permitted ADR transitions succeed; illegal transitions return a TransitionError.
// -----------------------------------------------------------------------------
func TestADRTransitions(t *testing.T) {
	// Valid ADR transitions
	validCases := []struct {
		from string
		to   string
	}{
		{"proposed", "accepted"},
		{"proposed", "rejected"},
		{"accepted", "superseded"},
		{"accepted", "deprecated"},
		{"proposed", "proposed"}, // Same status idempotent
	}

	for _, c := range validCases {
		if err := lifecycle.ValidateTransition(gyrus.TypeADR, c.from, c.to); err != nil {
			t.Errorf("Expected valid ADR transition from '%s' to '%s', got error: %v", c.from, c.to, err)
		}
	}

	// Invalid ADR transitions
	invalidCases := []struct {
		from string
		to   string
	}{
		{"accepted", "proposed"},
		{"superseded", "accepted"},
		{"rejected", "proposed"},
		{"draft", "accepted"},
	}

	for _, c := range invalidCases {
		if err := lifecycle.ValidateTransition(gyrus.TypeADR, c.from, c.to); err == nil {
			t.Errorf("Expected transition error from '%s' to '%s', got nil", c.from, c.to)
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies the valid and invalid lifecycle state transition rules for Improvement Proposals (IP).
// [Execution Surface]: In-Memory Package (internal/domain/lifecycle)
// [Assertions]: Permitted IP transitions succeed; illegal transitions return a TransitionError.
// -----------------------------------------------------------------------------
func TestImprovementProposalTransitions(t *testing.T) {
	validCases := []struct {
		from string
		to   string
	}{
		{"draft", "reviewing"},
		{"reviewing", "approved"},
		{"approved", "implemented"},
		{"reviewing", "abandoned"},
		{"draft", "draft"},
	}

	for _, c := range validCases {
		if err := lifecycle.ValidateTransition(gyrus.TypeImprovementProposal, c.from, c.to); err != nil {
			t.Errorf("Expected valid IP transition from '%s' to '%s', got error: %v", c.from, c.to, err)
		}
	}

	invalidCases := []struct {
		from string
		to   string
	}{
		{"implemented", "draft"},
		{"approved", "reviewing"},
		{"draft", "implemented"},
	}

	for _, c := range invalidCases {
		if err := lifecycle.ValidateTransition(gyrus.TypeImprovementProposal, c.from, c.to); err == nil {
			t.Errorf("Expected transition error from '%s' to '%s', got nil", c.from, c.to)
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies the valid and invalid lifecycle state transition rules for general living documents (PRD, Specs, Standards).
// [Execution Surface]: In-Memory Package (internal/domain/lifecycle)
// [Assertions]: Permitted general document transitions succeed; illegal transitions return a TransitionError.
// -----------------------------------------------------------------------------
func TestGeneralTransitions(t *testing.T) {
	validCases := []struct {
		from string
		to   string
	}{
		{"draft", "active"},
		{"active", "deprecated"},
		{"deprecated", "archived"},
		{"draft", "archived"},
		{"active", "active"},
	}

	for _, c := range validCases {
		if err := lifecycle.ValidateTransition(gyrus.TypeSpecification, c.from, c.to); err != nil {
			t.Errorf("Expected valid specification transition from '%s' to '%s', got error: %v", c.from, c.to, err)
		}
	}

	invalidCases := []struct {
		from string
		to   string
	}{
		{"archived", "active"},
		{"deprecated", "active"},
		{"active", "draft"},
	}

	for _, c := range invalidCases {
		if err := lifecycle.ValidateTransition(gyrus.TypeSpecification, c.from, c.to); err == nil {
			t.Errorf("Expected transition error from '%s' to '%s', got nil", c.from, c.to)
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies document immutability enforcement for accepted ADRs and documents marked with `immutable: true`.
// [Execution Surface]: In-Memory Package (internal/domain/lifecycle)
// [Assertions]: Mutating accepted ADRs or immutable-flagged documents returns an ImmutabilityError; mutable documents succeed.
// -----------------------------------------------------------------------------
func TestValidateMutation(t *testing.T) {
	// Mutating accepted ADR content should fail (built-in immutable type)
	if err := lifecycle.ValidateMutation(gyrus.TypeADR, "accepted", false, true); err == nil {
		t.Error("Expected immutability error for accepted ADR content update, got nil")
	}

	// Mutating proposed ADR content should succeed
	if err := lifecycle.ValidateMutation(gyrus.TypeADR, "proposed", false, true); err != nil {
		t.Errorf("Expected valid mutation for proposed ADR, got error: %v", err)
	}

	// Mutating active living spec content should succeed (default mutable)
	if err := lifecycle.ValidateMutation(gyrus.TypeSpecification, "active", false, true); err != nil {
		t.Errorf("Expected valid mutation for active specification, got error: %v", err)
	}

	// Mutating active custom doc with immutable: true flag in frontmatter should fail!
	if err := lifecycle.ValidateMutation(gyrus.TypeFreeform, "active", true, true); err == nil {
		t.Error("Expected immutability error for custom document with immutable: true flag, got nil")
	}
}

type mockSchemaStore struct {
	schemas map[string]string
}

func (m *mockSchemaStore) Create(ctx context.Context, doc gyrus.Document) (gyrus.DocumentRef, error) {
	return gyrus.DocumentRef{}, nil
}
func (m *mockSchemaStore) Get(ctx context.Context, id string) (gyrus.Document, error) {
	return gyrus.Document{}, nil
}
func (m *mockSchemaStore) Update(ctx context.Context, id string, patch gyrus.DocumentPatch, expectedVersion int) (gyrus.DocumentRef, error) {
	return gyrus.DocumentRef{}, nil
}
func (m *mockSchemaStore) Delete(ctx context.Context, id string) error {
	return nil
}
func (m *mockSchemaStore) Archive(ctx context.Context, id string) error {
	return nil
}

func (m *mockSchemaStore) GetSchema(ctx context.Context, docType string) (string, error) {
	if content, ok := m.schemas[docType]; ok {
		return content, nil
	}
	return "", os.ErrNotExist
}
func (m *mockSchemaStore) SaveSchema(ctx context.Context, docType string, content string) error {
	m.schemas[docType] = content
	return nil
}
func (m *mockSchemaStore) ListSchemas(ctx context.Context) ([]string, error) {
	var list []string
	for k := range m.schemas {
		list = append(list, k)
	}
	sort.Strings(list)
	return list, nil
}
func (m *mockSchemaStore) DeleteSchema(ctx context.Context, docType string) error {
	if _, ok := m.schemas[docType]; !ok {
		return fmt.Errorf("schema '%s' not found", docType)
	}
	delete(m.schemas, docType)
	return nil
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that lifecycle.Engine executes schema resolution precedence (persisted > embedded > fallback), listing, saving, and deleting.
// [Execution Surface]: In-Memory Lifecycle Engine with Mock SchemaStore
// [Assertions]: Custom persisted schema overrides embedded template, deletion restores embedded fallback, list aggregates both sources.
// -----------------------------------------------------------------------------
func TestEngineSchemaOperations(t *testing.T) {
	ctx := context.Background()
	mock := &mockSchemaStore{
		schemas: make(map[string]string),
	}

	engine := lifecycle.NewEngine(mock, nil, nil, nil, "")

	// 1. GetSchema initially returns embedded ADR template
	adrTemplate, err := engine.GetSchema(ctx, "adr")
	if err != nil {
		t.Fatalf("GetSchema('adr') failed: %v", err)
	}
	if !strings.Contains(adrTemplate, "Architecture Decision Record") {
		t.Errorf("Expected embedded ADR template, got: %s", adrTemplate)
	}

	// 2. Save custom ADR schema
	customADR := `---
id: <unique-id>
title: <Custom ADR>
category: architecture
type: adr
owner_group: architecture-board
version: 1
status: proposed
---

# Custom ADR Template
`
	if err := engine.SaveSchema(ctx, "adr", customADR); err != nil {
		t.Fatalf("SaveSchema('adr') failed: %v", err)
	}

	// 3. GetSchema now returns persisted schema (precedence over embedded)
	fetched, err := engine.GetSchema(ctx, "adr")
	if err != nil {
		t.Fatalf("GetSchema('adr') after save failed: %v", err)
	}
	if fetched != customADR {
		t.Errorf("Expected custom persisted schema to take precedence over embedded")
	}

	// 4. ListSchemas reflects 'adr' as persisted and other types as embedded
	list, err := engine.ListSchemas(ctx)
	if err != nil {
		t.Fatalf("ListSchemas failed: %v", err)
	}
	var adrInfo *gyrus.SchemaInfo
	for i := range list {
		if list[i].Type == "adr" {
			adrInfo = &list[i]
			break
		}
	}
	if adrInfo == nil {
		t.Fatalf("Expected 'adr' in list, not found")
	}
	if adrInfo.Source != "persisted" {
		t.Errorf("Expected 'adr' source to be 'persisted', got '%s'", adrInfo.Source)
	}

	// 5. Delete custom ADR schema
	if err := engine.DeleteSchema(ctx, "adr"); err != nil {
		t.Fatalf("DeleteSchema('adr') failed: %v", err)
	}

	// 6. GetSchema now falls back to embedded template again
	reverted, err := engine.GetSchema(ctx, "adr")
	if err != nil {
		t.Fatalf("GetSchema('adr') after delete failed: %v", err)
	}
	if !strings.Contains(reverted, "Architecture Decision Record") {
		t.Errorf("Expected fallback to embedded template, got: %s", reverted)
	}

	// 7. GetSchema returns error for non-existent schema
	if _, err := engine.GetSchema(ctx, "nonexistent-doc-type"); err == nil {
		t.Errorf("Expected error for non-existent docType, got nil")
	}
}
