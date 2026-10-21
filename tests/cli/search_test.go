package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

func TestCLI_SearchLimitPassThrough(t *testing.T) {
	workspace := t.TempDir()
	run := func(args ...string) []byte {
		t.Helper()
		cmd := exec.Command(gyrusBinPath, args...)
		cmd.Dir = workspace
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
		return out
	}
	run("config", "init", "--profile", "local")
	const fixtureCount = 12
	for i := 0; i < fixtureCount; i++ {
		run("create", "--id", fmt.Sprintf("limit-%d", i), "--title", "Limit fixture", "--type", "guide", "--category", "technical", "--content", "limitmarker")
	}
	for _, tc := range []struct {
		query, limit string
		want         int
	}{
		{"limitmarker", "", 10}, {"limitmarker", "3", 3}, {"limitmarker", "20", fixtureCount}, {"", "", fixtureCount}, {"", "3", 3},
	} {
		args := []string{"search", "--query", tc.query, "--scope", "all", "--json"}
		if tc.limit != "" {
			args = append(args, "--max-results", tc.limit)
		}
		var results []gyrus.SearchResult
		if err := json.Unmarshal(run(args...), &results); err != nil {
			t.Fatal(err)
		}
		if len(results) != tc.want {
			t.Errorf("query=%q limit=%q: got %d, want %d", tc.query, tc.limit, len(results), tc.want)
		}
	}
	cmd := exec.Command(gyrusBinPath, "sync")
	cmd.Dir = workspace
	if err := cmd.Run(); err == nil {
		t.Fatal("removed sync command succeeded")
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus search' searches documents by keyword and metadata filters, returning both formatted text and pure JSON.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Results contain matched documents; filter combinations restrict results; --json flag outputs pure parseable JSON array.
// -----------------------------------------------------------------------------
func TestCLI_Search(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Init
	initCmd := exec.Command(gyrusBinPath, "config", "init", "--profile", "local")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	// Create 2 test documents
	doc1Cmd := exec.Command(gyrusBinPath, "create",
		"--id", "adr-search-001",
		"--title", "PostgreSQL Vector Store",
		"--category", "architecture",
		"--type", "adr",
		"--tenant", "platform-team",
		"--status", "accepted",
		"--tags", "postgres,vector,search",
		"--content", "Semantic vector search with pgvector extension.",
	)
	doc1Cmd.Dir = tempWorkspace
	if out, err := doc1Cmd.CombinedOutput(); err != nil {
		t.Fatalf("Create doc1 failed: %v\nOutput: %s", err, string(out))
	}

	doc2Cmd := exec.Command(gyrusBinPath, "create",
		"--id", "guide-search-001",
		"--title", "Testing Guide for Developers",
		"--category", "technical",
		"--type", "guide",
		"--tenant", "qa-team",
		"--status", "active",
		"--tags", "testing,guide,qa",
		"--content", "Comprehensive guidelines for writing Go unit and integration tests.",
	)
	doc2Cmd.Dir = tempWorkspace
	if out, err := doc2Cmd.CombinedOutput(); err != nil {
		t.Fatalf("Create doc2 failed: %v\nOutput: %s", err, string(out))
	}

	// 1. Text Search by Keyword
	searchCmd := exec.Command(gyrusBinPath, "search", "--query", "pgvector")
	searchCmd.Dir = tempWorkspace
	var searchOut bytes.Buffer
	searchCmd.Stdout = &searchOut
	if err := searchCmd.Run(); err != nil {
		t.Fatalf("Search keyword failed: %v", err)
	}

	if !strings.Contains(searchOut.String(), "platform-team/projects/default/adr-search-001") {
		t.Errorf("Expected search results to include 'adr-search-001', got:\n%s", searchOut.String())
	}

	// 2. Search with JSON output
	searchJSONCmd := exec.Command(gyrusBinPath, "search", "--query", "guidelines", "--json")
	searchJSONCmd.Dir = tempWorkspace
	var jsonOut bytes.Buffer
	searchJSONCmd.Stdout = &jsonOut
	if err := searchJSONCmd.Run(); err != nil {
		t.Fatalf("Search JSON failed: %v", err)
	}

	var results []gyrus.SearchResult
	if err := json.Unmarshal(jsonOut.Bytes(), &results); err != nil {
		t.Fatalf("Failed unmarshaling search JSON output: %v\nOutput:\n%s", err, jsonOut.String())
	}

	if len(results) != 1 || results[0].Concept.ID != "guide-search-001" {
		t.Errorf("Expected 1 result for 'guide-search-001', got %v", results)
	}

	// 3. Search with Category Filter
	filterCmd := exec.Command(gyrusBinPath, "search", "--category", "architecture", "--json")
	filterCmd.Dir = tempWorkspace
	var filterOut bytes.Buffer
	filterCmd.Stdout = &filterOut
	if err := filterCmd.Run(); err != nil {
		t.Fatalf("Filtered search failed: %v", err)
	}

	var filterResults []gyrus.SearchResult
	if err := json.Unmarshal(filterOut.Bytes(), &filterResults); err != nil {
		t.Fatalf("Failed unmarshaling filter JSON output: %v\nOutput:\n%s", err, filterOut.String())
	}

	if len(filterResults) != 1 || filterResults[0].Concept.ID != "adr-search-001" {
		t.Errorf("Expected 1 filtered result for 'adr-search-001', got %v", filterResults)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus suggest-context' surfaces relevant context documents matching a prompt task.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Context linearization returns relevant document text; --json returns pure context payload.
// -----------------------------------------------------------------------------
func TestCLI_SuggestContext(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Init
	initCmd := exec.Command(gyrusBinPath, "config", "init", "--profile", "local")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}

	// Create document
	createCmd := exec.Command(gyrusBinPath, "create",
		"--id", "spec-suggest-001",
		"--title", "Authentication & Security Model",
		"--category", "architecture",
		"--type", "specification",
		"--tenant", "security-team",
		"--status", "active",
		"--content", "All API endpoints must authenticate with OAuth2 bearer tokens.",
	)
	createCmd.Dir = tempWorkspace
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("Create failed: %v\nOutput: %s", err, string(out))
	}

	// 1. Text suggest-context
	suggestCmd := exec.Command(gyrusBinPath, "suggest-context", "--prompt", "How should we implement API OAuth2 authentication?")
	suggestCmd.Dir = tempWorkspace
	var suggestOut bytes.Buffer
	suggestCmd.Stdout = &suggestOut
	if err := suggestCmd.Run(); err != nil {
		t.Fatalf("Suggest context failed: %v", err)
	}

	if !strings.Contains(suggestOut.String(), "security-team/projects/default/spec-suggest-001") && !strings.Contains(suggestOut.String(), "OAuth2") {
		t.Errorf("Expected suggest-context output to contain spec-suggest-001 or OAuth2 context, got:\n%s", suggestOut.String())
	}

	// 2. JSON suggest-context
	jsonCmd := exec.Command(gyrusBinPath, "suggest-context", "--prompt", "OAuth2 security model", "--json")
	jsonCmd.Dir = tempWorkspace
	var jsonOut bytes.Buffer
	jsonCmd.Stdout = &jsonOut
	if err := jsonCmd.Run(); err != nil {
		t.Fatalf("Suggest context JSON failed: %v", err)
	}

	var jsonRes map[string]string
	if err := json.Unmarshal(jsonOut.Bytes(), &jsonRes); err != nil {
		t.Fatalf("Failed unmarshaling suggest-context JSON output: %v\nOutput:\n%s", err, jsonOut.String())
	}

	if _, exists := jsonRes["context"]; !exists {
		t.Errorf("Expected 'context' key in JSON response, got %v", jsonRes)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: E2E CLI Integration Test
// [Purpose]: Verifies 'gyrus suggest-context --hook' as a Claude Code UserPromptSubmit hook.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess, payload on stdin
// [Assertions]: Started outside the project, it resolves configuration from the payload's cwd and prints
// hookSpecificOutput.additionalContext within the budget; an empty prompt or a bad payload prints nothing and exits 0.
// -----------------------------------------------------------------------------
func TestCLI_SuggestContextHook(t *testing.T) {
	workspace := t.TempDir()
	initCmd := exec.Command(gyrusBinPath, "config", "init", "--profile", "local")
	initCmd.Dir = workspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Init failed: %v\nOutput: %s", err, string(out))
	}
	createCmd := exec.Command(gyrusBinPath, "create",
		"--id", "guide-queue-001",
		"--title", "Queue operations",
		"--category", "operations",
		"--type", "guide",
		"--tenant", "ops",
		"--status", "active",
		"--content", "# Queue\n\n## Restart\n\nRestart the queue with the restart script.\n\n## Unrelated\n\nNothing here.",
	)
	createCmd.Dir = workspace
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("Create failed: %v\nOutput: %s", err, string(out))
	}

	run := func(payload string) (string, string) {
		cmd := exec.Command(gyrusBinPath, "suggest-context", "--hook", "--max-tokens", "1500")
		cmd.Dir = t.TempDir()
		cmd.Stdin = strings.NewReader(payload)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("hook must exit 0, got %v\nstderr: %s", err, stderr.String())
		}
		return stdout.String(), stderr.String()
	}

	payload, _ := json.Marshal(map[string]string{"session_id": "s1", "hook_event_name": "UserPromptSubmit", "prompt": "how do I restart the queue?", "cwd": workspace})
	out, stderr := run(string(payload))
	var result struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("hook output is not JSON: %v\nstdout: %s\nstderr: %s", err, out, stderr)
	}
	if result.HookSpecificOutput.HookEventName != "UserPromptSubmit" || !strings.Contains(result.HookSpecificOutput.AdditionalContext, "Restart the queue") {
		t.Errorf("unexpected hook output: %+v", result)
	}

	if out, _ := run(`{"prompt": "  ", "cwd": "` + strings.ReplaceAll(workspace, `\`, `\`) + `"}`); out != "" {
		t.Errorf("an empty prompt must print nothing, got %q", out)
	}
	if out, stderr := run("not json"); out != "" || !strings.Contains(stderr, "suggest-context hook") {
		t.Errorf("a bad payload must print nothing and report on stderr, got %q / %q", out, stderr)
	}
}
