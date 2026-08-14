package cli

import (
	"errors"
	"testing"

	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that MapErrorToExitCode translates domain and storage errors into the standardized POSIX exit code contract.
// [Execution Surface]: In-Memory Unit Test (internal/cli)
// [Assertions]: Validation error -> 1; Transition error -> 2; Concurrency error -> 4; Auth error -> 3; Storage error -> 5.
// -----------------------------------------------------------------------------
func TestMapErrorToExitCode(t *testing.T) {
	cases := []struct {
		err          error
		expectedCode int
	}{
		{nil, ExitSuccess},
		{&okf.ValidationError{Field: "id", Message: "invalid"}, ExitValidationError},
		{&lifecycle.TransitionError{DocType: gyrus.TypeADR, CurrentStatus: "accepted", NewStatus: "proposed"}, ExitTransitionError},
		{errors.New("concurrency error: expected version mismatch"), ExitConcurrencyError},
		{errors.New("permission denied: unauthorized group"), ExitAuthError},
		{errors.New("disk storage read error"), ExitStorageError},
	}

	for _, c := range cases {
		code := MapErrorToExitCode(c.err)
		if code != c.expectedCode {
			t.Errorf("For error '%v', expected exit code %d, got %d", c.err, c.expectedCode, code)
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that the root Cobra command builds successfully and renders CLI help output.
// [Execution Surface]: In-Memory Cobra Command Tree
// [Assertions]: BuildRootCmd succeeds and executing --help returns Exit Code 0.
// -----------------------------------------------------------------------------
func TestRootCommandHelp(t *testing.T) {
	rootCmd, err := BuildRootCmd("")
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("RootCmd --help failed: %v", err)
	}
}
