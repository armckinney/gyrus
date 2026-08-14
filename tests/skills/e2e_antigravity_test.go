package skills_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Live Integration Test (Optional/Explicit)
// [Purpose]: Executes live integration evaluation against Google Antigravity agent shell script.
// [Execution Surface]: Subprocess execution of tests/skills/e2e_antigravity_eval.sh
// [Assertions]: Evaluation script finishes with exit code 0 when RUN_INTEGRATION_TESTS=1 is set.
// -----------------------------------------------------------------------------
func TestAntigravityIntegration(t *testing.T) {
	if os.Getenv("RUN_AGENT_EVAL") == "" && os.Getenv("RUN_E2E_TESTS") == "" {
		t.Skip("Skipping live Antigravity agent eval script. Set RUN_AGENT_EVAL=1 to run.")
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
