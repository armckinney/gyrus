package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies that 'gyrus hook pre-invocation' outputs valid JSON containing
//            SDLC directives and context, and honors GYRUS_HOOKS_ENABLED=false.
// [Assertions]: Outputs valid JSON with injectSteps, or {} when disabled.
// -----------------------------------------------------------------------------
func TestHookPreInvocation(t *testing.T) {
	tempWork := t.TempDir()
	os.Setenv("GYRUS_WORKSPACE", tempWork)
	defer os.Unsetenv("GYRUS_WORKSPACE")

	// 1. When hooks disabled via env var
	os.Setenv("GYRUS_HOOKS_ENABLED", "false")
	rootCmd, err := BuildRootCmd()
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"hook", "pre-invocation"})

	// Intercept stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = rootCmd.Execute()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("hook pre-invocation returned error: %v", err)
	}

	var out bytes.Buffer
	_, _ = out.ReadFrom(r)
	trimmed := strings.TrimSpace(out.String())
	if trimmed != "{}" {
		t.Errorf("Expected '{}' when GYRUS_HOOKS_ENABLED=false, got: %s", trimmed)
	}

	// 2. When hooks enabled
	os.Unsetenv("GYRUS_HOOKS_ENABLED")
	rootCmd2, err := BuildRootCmd()
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}

	r2, w2, _ := os.Pipe()
	os.Stdout = w2

	rootCmd2.SetArgs([]string{"hook", "pre-invocation", "--prompt", "architecture"})
	err = rootCmd2.Execute()
	w2.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("hook pre-invocation returned error: %v", err)
	}

	var out2 bytes.Buffer
	_, _ = out2.ReadFrom(r2)
	trimmed2 := strings.TrimSpace(out2.String())

	// Must be valid JSON (either {} or {"injectSteps":[...]})
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed2), &jsonMap); err != nil {
		t.Fatalf("Expected valid JSON from pre-invocation hook, got: %s (err: %v)", trimmed2, err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Integration Test
// [Purpose]: Verifies that 'gyrus hook stop' outputs valid JSON and honors disabled settings.
// [Assertions]: Outputs valid JSON ({}) on clean working directory or when disabled.
// -----------------------------------------------------------------------------
func TestHookStop(t *testing.T) {
	tempWork := t.TempDir()
	os.Setenv("GYRUS_WORKSPACE", tempWork)
	defer os.Unsetenv("GYRUS_WORKSPACE")

	// 1. When hooks disabled
	os.Setenv("GYRUS_HOOKS_ENABLED", "false")
	rootCmd, err := BuildRootCmd()
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"hook", "stop"})
	err = rootCmd.Execute()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("hook stop returned error: %v", err)
	}

	var out bytes.Buffer
	_, _ = out.ReadFrom(r)
	trimmed := strings.TrimSpace(out.String())
	if trimmed != "{}" {
		t.Errorf("Expected '{}' when disabled, got: %s", trimmed)
	}

	// 2. When enabled on clean tree
	os.Unsetenv("GYRUS_HOOKS_ENABLED")
	rootCmd2, err := BuildRootCmd()
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}

	r2, w2, _ := os.Pipe()
	os.Stdout = w2

	rootCmd2.SetArgs([]string{"hook", "stop"})
	err = rootCmd2.Execute()
	w2.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("hook stop returned error: %v", err)
	}

	var out2 bytes.Buffer
	_, _ = out2.ReadFrom(r2)
	trimmed2 := strings.TrimSpace(out2.String())

	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed2), &jsonMap); err != nil {
		t.Fatalf("Expected valid JSON from stop hook, got: %s (err: %v)", trimmed2, err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that .gyrus.yaml automation toggles are honored by hook commands.
// [Assertions]: When automation.hooks_enabled=false, hook outputs {}.
// -----------------------------------------------------------------------------
func TestHookConfigToggle(t *testing.T) {
	tempWork := t.TempDir()
	wsYaml := `storage:
  provider: localfs
automation:
  hooks_enabled: false
`
	if err := os.WriteFile(filepath.Join(tempWork, ".gyrus.yaml"), []byte(wsYaml), 0644); err != nil {
		t.Fatalf("Failed writing config: %v", err)
	}

	os.Setenv("GYRUS_WORKSPACE", tempWork)
	defer os.Unsetenv("GYRUS_WORKSPACE")

	rootCmd, err := BuildRootCmd()
	if err != nil {
		t.Fatalf("BuildRootCmd failed: %v", err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"hook", "pre-invocation"})
	err = rootCmd.Execute()
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("hook pre-invocation failed: %v", err)
	}

	var out bytes.Buffer
	_, _ = out.ReadFrom(r)
	trimmed := strings.TrimSpace(out.String())
	if trimmed != "{}" {
		t.Errorf("Expected '{}' when automation.hooks_enabled=false in .gyrus.yaml, got: %s", trimmed)
	}
}
