package lifecycle

import (
	"fmt"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

// TransitionError indicates an illegal state transition attempt.
type TransitionError struct {
	DocType       gyrus.DocumentType
	CurrentStatus string
	NewStatus     string
}

func (e *TransitionError) Error() string {
	return fmt.Sprintf("invalid lifecycle transition for type '%s': cannot transition from '%s' to '%s'", e.DocType, e.CurrentStatus, e.NewStatus)
}

// adrTransitions defines valid state transitions for ADR documents.
var adrTransitions = map[string]map[string]bool{
	"proposed": {
		"accepted": true,
		"rejected": true,
	},
	"accepted": {
		"superseded": true,
		"deprecated": true,
	},
}

// ipTransitions defines valid state transitions for Improvement Proposals.
var ipTransitions = map[string]map[string]bool{
	"draft": {
		"reviewing": true,
		"abandoned": true,
	},
	"reviewing": {
		"approved":  true,
		"rejected":  true,
		"abandoned": true,
	},
	"approved": {
		"implemented": true,
		"abandoned":   true,
	},
}

// generalTransitions defines default state transitions for standard document types.
var generalTransitions = map[string]map[string]bool{
	"draft": {
		"active":   true,
		"archived": true,
	},
	"active": {
		"deprecated": true,
		"archived":   true,
	},
	"deprecated": {
		"archived": true,
	},
}

// ValidateTransition verifies whether moving from currentStatus to newStatus is permitted for docType.
func ValidateTransition(docType gyrus.DocumentType, currentStatus, newStatus string) error {
	if currentStatus == newStatus {
		return nil // No-op status update is always valid
	}

	var transitions map[string]map[string]bool

	switch docType {
	case gyrus.TypeADR:
		transitions = adrTransitions
	case gyrus.TypeImprovementProposal:
		transitions = ipTransitions
	default:
		transitions = generalTransitions
	}

	allowedNewStatuses, exists := transitions[currentStatus]
	if !exists || !allowedNewStatuses[newStatus] {
		return &TransitionError{
			DocType:       docType,
			CurrentStatus: currentStatus,
			NewStatus:     newStatus,
		}
	}

	return nil
}
