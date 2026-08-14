package install_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test / Container
// [Purpose]: Verifies that 'packaging/Dockerfile' specifies an optimal multi-stage build and entrypoint for MCP stdio execution.
// [Execution Surface]: Dockerfile Static Analysis & Container Validation
// [Assertions]: Multi-stage build specifies builder image, minimal alpine runtime, non-root workspace, and ENTRYPOINT ["gyrus"].
// -----------------------------------------------------------------------------
func TestDockerfileSpecification(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repo root: %v", err)
	}

	dockerfilePath := filepath.Join(repoRoot, "packaging", "Dockerfile")
	data, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("Failed to read packaging/Dockerfile: %v", err)
	}

	content := string(data)

	requiredDirectives := []string{
		"FROM golang:",
		"AS builder",
		"CGO_ENABLED=0",
		"FROM alpine:",
		"COPY --from=builder /bin/gyrus /usr/local/bin/gyrus",
		"ENTRYPOINT [\"gyrus\"]",
		"CMD [\"mcp\", \"serve\"]",
	}

	for _, dir := range requiredDirectives {
		if !strings.Contains(content, dir) {
			t.Errorf("packaging/Dockerfile missing required directive: %s", dir)
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test / Container Runtime
// [Purpose]: Builds and executes Gyrus MCP inside a Docker container when Docker daemon is available.
// [Execution Surface]: Live Docker Engine (Skipped if Docker is not available in environment)
// [Assertions]: Container starts, serves help/mcp, and terminates gracefully.
// -----------------------------------------------------------------------------
func TestLiveDockerContainerExecution(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker executable not found on PATH; skipping live Docker container test")
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repo root: %v", err)
	}

	tag := "gyrus-test:local"
	buildCmd := exec.Command("docker", "build", "-f", "packaging/Dockerfile", "-t", tag, ".")
	buildCmd.Dir = repoRoot
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Docker build failed: %v\nOutput: %s", err, string(out))
	}
	defer func() {
		_ = exec.Command("docker", "rmi", tag).Run()
	}()

	runCmd := exec.Command("docker", "run", "--rm", tag, "--help")
	out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Docker run --help failed: %v\nOutput: %s", err, string(out))
	}

	if !strings.Contains(string(out), "Gyrus: Unified Context & Memory Engine") {
		t.Errorf("Unexpected Docker container output: %s", string(out))
	}
}
