package cli

import (
	"errors"
	"strings"

	"github.com/armckinney/gyrus/internal/lifecycle"
	"github.com/armckinney/gyrus/internal/okf"
)

const (
	ExitSuccess          = 0
	ExitValidationError  = 1
	ExitTransitionError  = 2
	ExitAuthError        = 3
	ExitConcurrencyError = 4
	ExitStorageError     = 5
)

// MapErrorToExitCode inspects an error and returns the corresponding Gyrus programmatic exit code.
func MapErrorToExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}

	var valErr *okf.ValidationError
	if errors.As(err, &valErr) {
		return ExitValidationError
	}

	var transErr *lifecycle.TransitionError
	if errors.As(err, &transErr) {
		return ExitTransitionError
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "concurrency error") || strings.Contains(errMsg, "expected version") {
		return ExitConcurrencyError
	}
	if strings.Contains(errMsg, "permission denied") || strings.Contains(errMsg, "unauthorized") {
		return ExitAuthError
	}
	if strings.Contains(errMsg, "validation error") {
		return ExitValidationError
	}

	return ExitStorageError
}
