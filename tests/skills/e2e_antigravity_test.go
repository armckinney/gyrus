package skills_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestAntigravityIntegration executes live integration evaluation against Google Antigravity.
// This test is SKIPPED by default during standard `go test ./...` and `make test`.
// To execute: set RUN_INTEGRATION_TESTS=1 or run `make test-integration`.
func TestAntigravityIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" && os.Getenv("RUN_E2E_TESTS") == "" {
		t.Skip("Skipping live integration test (marked for explicit execution only). Set RUN_INTEGRATION_TESTS=1 to run.")
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repo root: %v", err)
	}

	evalScript := filepath.Join(repoRoot, "tests", "skills", "e2e_antigravity_eval.sh")
	cmd := exec.Command("bash", evalScript)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Live integration evaluation script failed: %v\nOutput:\n%s", err, string(out))
	}
}
