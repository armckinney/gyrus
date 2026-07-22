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
