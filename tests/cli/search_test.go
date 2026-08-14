package cli_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Product E2E Test
// [Purpose]: Verifies that 'gyrus search' searches documents by keyword and metadata filters, returning both formatted text and pure JSON.
// [Execution Surface]: Compiled ./gyrus CLI binary via os/exec subprocess
// [Assertions]: Results contain matched documents; filter combinations restrict results; --json flag outputs pure parseable JSON array.
// -----------------------------------------------------------------------------
func TestCLI_Search(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Init
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
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
		"--owner-group", "platform-team",
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
		"--owner-group", "qa-team",
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

	if !strings.Contains(searchOut.String(), "adr-search-001") {
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

	if len(results) != 1 || results[0].Document.ID != "guide-search-001" {
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

	if len(filterResults) != 1 || filterResults[0].Document.ID != "adr-search-001" {
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
	initCmd := exec.Command(gyrusBinPath, "init", "--profile", "local", "--no-mcp", "--no-skill")
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
		"--owner-group", "security-team",
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

	if !strings.Contains(suggestOut.String(), "spec-suggest-001") && !strings.Contains(suggestOut.String(), "OAuth2") {
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
