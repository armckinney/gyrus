package gyrus

import "fmt"

// ConflictError reports a mutation rejected because its observed state changed.
type ConflictError struct {
	ConceptID string
	Operation string
	Err       error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("concurrency conflict during %s for concept %q: %v", e.Operation, e.ConceptID, e.Err)
}
func (e *ConflictError) Unwrap() error { return e.Err }

// UnsupportedConditionalMutationError identifies a driver without safe mutations.
type UnsupportedConditionalMutationError struct{ Driver, Operation string }

func (e *UnsupportedConditionalMutationError) Error() string {
	return fmt.Sprintf("%s does not support conditional %s", e.Driver, e.Operation)
}

// PartialGraphMutationError reports a failed batch without hiding committed sources.
type PartialGraphMutationError struct {
	AppliedSourceIDs []string
	FailedSourceID   string
	Err              error
}

func (e *PartialGraphMutationError) Error() string {
	return fmt.Sprintf("graph mutation failed for %q after applying sources %v: %v", e.FailedSourceID, e.AppliedSourceIDs, e.Err)
}
func (e *PartialGraphMutationError) Unwrap() error { return e.Err }
