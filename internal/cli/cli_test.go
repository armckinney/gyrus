package cli_test

import (
	"errors"
	"testing"

	"github.com/armckinney/gyrus/internal/cli"
	"github.com/armckinney/gyrus/internal/lifecycle"
	"github.com/armckinney/gyrus/internal/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestMapErrorToExitCode(t *testing.T) {
	cases := []struct {
		err          error
		expectedCode int
	}{
		{nil, cli.ExitSuccess},
		{&okf.ValidationError{Field: "id", Message: "invalid"}, cli.ExitValidationError},
		{&lifecycle.TransitionError{DocType: gyrus.TypeADR, CurrentStatus: "accepted", NewStatus: "proposed"}, cli.ExitTransitionError},
		{errors.New("concurrency error: expected version mismatch"), cli.ExitConcurrencyError},
		{errors.New("permission denied: unauthorized group"), cli.ExitAuthError},
		{errors.New("disk storage read error"), cli.ExitStorageError},
	}

	for _, c := range cases {
		code := cli.MapErrorToExitCode(c.err)
		if code != c.expectedCode {
			t.Errorf("For error '%v', expected exit code %d, got %d", c.err, c.expectedCode, code)
		}
	}
}

func TestRootCommandHelp(t *testing.T) {
	cli.RootCmd.SetArgs([]string{"--help"})
	if err := cli.RootCmd.Execute(); err != nil {
		t.Fatalf("RootCmd --help failed: %v", err)
	}
}
