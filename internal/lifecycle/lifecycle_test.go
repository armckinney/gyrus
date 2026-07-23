package lifecycle_test

import (
	"testing"

	"github.com/armckinney/gyrus/internal/lifecycle"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

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
		{"proposed", "proposed"}, // Same status
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

func TestImprovementProposalTransitions(t *testing.T) {
	validCases := []struct {
		from string
		to   string
	}{
		{"draft", "reviewing"},
		{"reviewing", "approved"},
		{"approved", "implemented"},
		{"reviewing", "abandoned"},
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

func TestGeneralTransitions(t *testing.T) {
	validCases := []struct {
		from string
		to   string
	}{
		{"draft", "active"},
		{"active", "deprecated"},
		{"deprecated", "archived"},
		{"draft", "archived"},
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
