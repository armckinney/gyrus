package install_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test / Packaging
// [Purpose]: Verifies that 'packaging/install.sh' has valid bash syntax and proper architecture detection.
// [Execution Surface]: Bash Script Static Verification & Syntax Validation
// [Assertions]: 'bash -n packaging/install.sh' passes with exit code 0; architecture detection covers amd64/arm64.
// -----------------------------------------------------------------------------
func TestInstallerScriptSyntaxAndArchitecture(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repo root: %v", err)
	}

	installScript := filepath.Join(repoRoot, "packaging", "install.sh")
	if _, err := os.Stat(installScript); os.IsNotExist(err) {
		t.Fatalf("Missing packaging/install.sh: %s", installScript)
	}

	// 1. Bash syntax check (bash -n)
	cmd := exec.Command("bash", "-n", installScript)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("install.sh syntax error: %v\nOutput: %s", err, string(out))
	}

	// 2. Inspect script contents for mandatory packaging rules
	data, err := os.ReadFile(installScript)
	if err != nil {
		t.Fatalf("Failed reading install.sh: %v", err)
	}

	content := string(data)
	requiredTokens := []string{
		"REPO=\"armckinney/gyrus\"",
		"BINARY=\"gyrus\"",
		"x86_64|amd64",
		"aarch64|arm64",
		"chmod +x",
	}

	for _, token := range requiredTokens {
		if !strings.Contains(content, token) {
			t.Errorf("install.sh missing required token: %s", token)
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test / Build
// [Purpose]: Verifies that Gyrus builds statically with CGO_ENABLED=0 for portable distribution.
// [Execution Surface]: Go Compiler Subprocess
// [Assertions]: Binary compiles cleanly without CGO dependencies and executes --help with Exit Code 0.
// -----------------------------------------------------------------------------
func TestStaticBinaryBuild(t *testing.T) {
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("Failed to resolve repo root: %v", err)
	}

	tempDir := t.TempDir()
	binPath := filepath.Join(tempDir, "gyrus_static")

	cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", binPath, "./cmd/gyrus")
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Static binary build failed: %v\nOutput: %s", err, string(out))
	}

	// Execute binary --help
	helpCmd := exec.Command(binPath, "--help")
	var helpOut bytes.Buffer
	helpCmd.Stdout = &helpOut
	if err := helpCmd.Run(); err != nil {
		t.Fatalf("Static binary execution failed: %v", err)
	}

	if !strings.Contains(helpOut.String(), "Gyrus is a high-performance local-first memory") {
		t.Errorf("Static binary output did not contain expected banner: %s", helpOut.String())
	}
}
