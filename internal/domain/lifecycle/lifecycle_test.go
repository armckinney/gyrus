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
// [Purpose]: Verifies the valid and invalid lifecycle state transition rules for general living concepts (PRD, Specs, Standards).
// [Execution Surface]: In-Memory Package (internal/domain/lifecycle)
// [Assertions]: Permitted general concept transitions succeed; illegal transitions return a TransitionError.
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
// [Purpose]: Verifies concept immutability enforcement for accepted ADRs and concepts marked with `immutable: true`.
// [Execution Surface]: In-Memory Package (internal/domain/lifecycle)
// [Assertions]: Mutating accepted ADRs or immutable-flagged concepts returns an ImmutabilityError; mutable concepts succeed.
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
		t.Error("Expected immutability error for custom concept with immutable: true flag, got nil")
	}
}

type mockSchemaStore struct {
	schemas map[string]string
}

func (m *mockSchemaStore) Create(ctx context.Context, doc gyrus.Concept) (gyrus.ConceptRef, error) {
	return gyrus.ConceptRef{}, nil
}
func (m *mockSchemaStore) Get(ctx context.Context, id string) (gyrus.Concept, error) {
	return gyrus.Concept{Tenant: "root"}, nil
}
func (m *mockSchemaStore) Update(ctx context.Context, id string, patch gyrus.ConceptPatch, expectedVersion int) (gyrus.ConceptRef, error) {
	return gyrus.ConceptRef{}, nil
}
func (m *mockSchemaStore) Delete(ctx context.Context, id string) error {
	return nil
}
func (m *mockSchemaStore) Archive(ctx context.Context, id string) error {
	return nil
}

func (m *mockSchemaStore) GetSchema(ctx context.Context, conceptType string) (string, error) {
	if content, ok := m.schemas[conceptType]; ok {
		return content, nil
	}
	return "", os.ErrNotExist
}
func (m *mockSchemaStore) SaveSchema(ctx context.Context, conceptType string, content string) error {
	m.schemas[conceptType] = content
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
func (m *mockSchemaStore) DeleteSchema(ctx context.Context, conceptType string) error {
	if _, ok := m.schemas[conceptType]; !ok {
		return fmt.Errorf("schema '%s' not found", conceptType)
	}
	delete(m.schemas, conceptType)
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
tenant: architecture-board
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
		t.Errorf("Expected error for non-existent conceptType, got nil")
	}
}

type mockStoreWithSearch struct {
	docs map[string]gyrus.Concept
}

func (m *mockStoreWithSearch) Create(ctx context.Context, doc gyrus.Concept) (gyrus.ConceptRef, error) {
	m.docs[doc.ID] = doc
	return gyrus.ConceptRef{ID: doc.ID}, nil
}
func (m *mockStoreWithSearch) Get(ctx context.Context, id string) (gyrus.Concept, error) {
	if doc, ok := m.docs[id]; ok {
		return doc, nil
	}
	return gyrus.Concept{Tenant: "root"}, fmt.Errorf("concept not found: %s", id)
}
func (m *mockStoreWithSearch) Update(ctx context.Context, id string, patch gyrus.ConceptPatch, expectedVersion int) (gyrus.ConceptRef, error) {
	return gyrus.ConceptRef{ID: id}, nil
}
func (m *mockStoreWithSearch) Delete(ctx context.Context, id string) error {
	delete(m.docs, id)
	return nil
}
func (m *mockStoreWithSearch) Archive(ctx context.Context, id string) error {
	return nil
}
func (m *mockStoreWithSearch) Search(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	var results []gyrus.SearchResult
	for _, doc := range m.docs {
		if filter.Scope == "reference" && doc.Scope != "reference" {
			continue
		}
		if filter.Scope != "all" && filter.Project != "" && doc.Scope == "project" && doc.Project != filter.Project {
			continue
		}
		score := 20.0
		if doc.Project == filter.Project && doc.Scope == "project" {
			score = 60.0
		}
		results = append(results, gyrus.SearchResult{
			Concept: doc,
			Score:   score,
		})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	return results, nil
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that SuggestContextWithFilter prioritizes project concepts, traverses OKF dependencies to include referenced concepts, and groups output into clean banners.
// [Assertions]: Output contains [PROJECT CONTEXT: gyrus] and [SHARED REFERENCE LAYER: root], with linked reference ADRs included.
// -----------------------------------------------------------------------------
func TestEngineSuggestContextWithProjectScopingAndDependencyExpansion(t *testing.T) {
	ctx := context.Background()

	mockStore := &mockStoreWithSearch{
		docs: make(map[string]gyrus.Concept),
	}

	wsDoc := gyrus.Concept{
		ID:       "ticket-gyrus-202",
		Title:    "Implement Project Scoping",
		Category: gyrus.CategoryProduct,
		Type:     gyrus.TypePRD,
		Tenant:   "root",
		Scope:    "project",
		Project:  "gyrus",
		Links:    []gyrus.ConceptLink{{Concept: "root/reference/adr-002-simplified-directory-topology", Rel: gyrus.RelDependsOn}},
		Content:  "Implementation details for scoping.", Name: "root/projects/gyrus/ticket-gyrus-202",
	}
	refDoc := gyrus.Concept{
		ID:       "adr-002-simplified-directory-topology",
		Title:    "Simplified Directory Topology",
		Category: gyrus.CategoryArchitecture,
		Type:     gyrus.TypeADR,
		Tenant:   "root",
		Scope:    "reference",
		Content:  "Architecture for .gyrus/concepts layout.", Name: "root/reference/adr-002-simplified-directory-topology",
	}
	foreignWsDoc := gyrus.Concept{
		ID:       "ticket-billing-001",
		Title:    "Billing Tasks",
		Category: gyrus.CategoryProduct,
		Type:     gyrus.TypePRD,
		Tenant:   "root",
		Scope:    "project",
		Project:  "billing",
		Content:  "Secret billing tasks.", Name: "root/projects/billing/ticket-billing-001",
	}

	mockStore.docs[wsDoc.ID] = wsDoc
	mockStore.docs[refDoc.ID] = refDoc
	mockStore.docs[foreignWsDoc.ID] = foreignWsDoc

	engine := lifecycle.NewEngineWithProject(mockStore, mockStore, nil, nil, "", "gyrus")

	output, err := engine.SuggestContext(ctx, "scoping", "", 5)
	if err != nil {
		t.Fatalf("SuggestContext failed: %v", err)
	}

	// 1. Verify project banner and project concept
	if !strings.Contains(output, "[PROJECT CONTEXT: gyrus]") {
		t.Errorf("Expected [PROJECT CONTEXT: gyrus] in output, got: %s", output)
	}
	if !strings.Contains(output, "root/projects/gyrus/ticket-gyrus-202") {
		t.Errorf("Expected ticket-gyrus-202 in output, got: %s", output)
	}

	// 2. Verify shared reference banner and expanded dependency adr-002
	if !strings.Contains(output, "[SHARED REFERENCE LAYER: root]") {
		t.Errorf("Expected [SHARED REFERENCE LAYER: root] in output, got: %s", output)
	}
	if !strings.Contains(output, "root/reference/adr-002-simplified-directory-topology") {
		t.Errorf("Expected adr-002-simplified-directory-topology in output, got: %s", output)
	}

	// 3. Verify foreign project concept is excluded
	if strings.Contains(output, "root/projects/billing/ticket-billing-001") {
		t.Errorf("Did not expect foreign project concept ticket-billing-001 in output, got: %s", output)
	}
}
