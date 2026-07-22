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

// ImmutabilityError indicates an illegal attempt to mutate content on a locked/immutable document.
type ImmutabilityError struct {
	DocType gyrus.DocumentType
	Status  string
}

func (e *ImmutabilityError) Error() string {
	return fmt.Sprintf("cannot modify content of immutable document type '%s' in status '%s'", e.DocType, e.Status)
}

// IsImmutableType returns true if the document type is an immutable historical decision log.
func IsImmutableType(docType gyrus.DocumentType) bool {
	switch docType {
	case gyrus.TypeADR, gyrus.TypeImprovementProposal, gyrus.TypeReleaseNote:
		return true
	default:
		return false
	}
}

// IsLockedStatus returns true if the document status represents a finalized/locked state.
func IsLockedStatus(docType gyrus.DocumentType, status string, isExplicitlyImmutable bool) bool {
	if isExplicitlyImmutable {
		return status != "draft" && status != "proposed"
	}
	switch status {
	case "accepted", "rejected", "superseded", "deprecated", "approved", "abandoned", "implemented", "published", "archived":
		return true
	default:
		return false
	}
}

// ValidateMutation verifies whether content modifications are permitted for the given docType, status, and immutability flag.
func ValidateMutation(docType gyrus.DocumentType, currentStatus string, isExplicitlyImmutable bool, contentChanged bool) error {
	if contentChanged && (isExplicitlyImmutable || IsImmutableType(docType)) && IsLockedStatus(docType, currentStatus, isExplicitlyImmutable) {
		return &ImmutabilityError{
			DocType: docType,
			Status:  currentStatus,
		}
	}
	return nil
}
