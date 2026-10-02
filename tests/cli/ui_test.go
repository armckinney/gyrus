package cli_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/armckinney/gyrus/internal/ui/handlers"
)

func TestUI_CLI_Help(t *testing.T) {
	cmd := exec.Command(gyrusBinPath, "ui", "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gyrus ui --help failed: %v, output: %s", err, string(out))
	}
	if !strings.Contains(string(out), "serve") {
		t.Errorf("expected 'serve' in gyrus ui --help output, got: %s", string(out))
	}

	cmdServe := exec.Command(gyrusBinPath, "ui", "serve", "--help")
	outServe, err := cmdServe.CombinedOutput()
	if err != nil {
		t.Fatalf("gyrus ui serve --help failed: %v, output: %s", err, string(outServe))
	}
	if !strings.Contains(string(outServe), "--port") || !strings.Contains(string(outServe), "--host") {
		t.Errorf("expected --port and --host flags in gyrus ui serve --help, got: %s", string(outServe))
	}
}

func TestUI_CLI_ServeE2E(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Initialize config in workspace
	initCmd := exec.Command(gyrusBinPath, "config", "init", "--profile", "local")
	initCmd.Dir = tempWorkspace
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("config init failed: %v, output: %s", err, string(out))
	}

	// Create test documents
	createCmd1 := exec.Command(gyrusBinPath, "create",
		"--id", "adr-e2e-001",
		"--title", "E2E Storage Architecture",
		"--category", "architecture",
		"--type", "adr",
		"--content", "ADR for E2E Web UI testing.",
	)
	createCmd1.Dir = tempWorkspace
	if out, err := createCmd1.CombinedOutput(); err != nil {
		t.Fatalf("create doc 1 failed: %v, output: %s", err, string(out))
	}

	createCmd2 := exec.Command(gyrusBinPath, "create",
		"--id", "prd-e2e-001",
		"--title", "E2E Web UI Roadmap",
		"--category", "product",
		"--type", "prd",
		"--content", "PRD for Web UI.",
	)
	createCmd2.Dir = tempWorkspace
	if out, err := createCmd2.CombinedOutput(); err != nil {
		t.Fatalf("create doc 2 failed: %v, output: %s", err, string(out))
	}

	// Link docs
	linkCmd := exec.Command(gyrusBinPath, "link",
		"prd-e2e-001", "adr-e2e-001",
		"--rel-type", "depends_on",
	)
	linkCmd.Dir = tempWorkspace
	if out, err := linkCmd.CombinedOutput(); err != nil {
		t.Fatalf("link docs failed: %v, output: %s", err, string(out))
	}


	testPort := 39182
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveCmd := exec.CommandContext(ctx, gyrusBinPath, "ui", "serve",
		"--workspace", tempWorkspace,
		"--port", fmt.Sprintf("%d", testPort),
		"--host", "127.0.0.1",
	)
	serveCmd.Dir = tempWorkspace

	if err := serveCmd.Start(); err != nil {
		t.Fatalf("failed starting gyrus ui serve: %v", err)
	}
	defer func() {
		cancel()
		_ = serveCmd.Wait()
	}()

	// Wait for server to become responsive
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", testPort)
	client := &http.Client{Timeout: 2 * time.Second}

	var ready bool
	for range 30 {
		time.Sleep(100 * time.Millisecond)
		resp, err := client.Get(baseURL + "/docs")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			ready = true
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	if !ready {
		t.Fatalf("server on %s did not become ready in time", baseURL)
	}

	// 1. Verify /docs contains created documents
	resp, err := client.Get(baseURL + "/docs")
	if err != nil {
		t.Fatalf("failed fetching /docs: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// 2. Verify static assets
	respHTMX, err := client.Get(baseURL + "/static/js/htmx.min.js")
	if err != nil {
		t.Fatalf("failed fetching htmx.min.js: %v", err)
	}
	respHTMX.Body.Close()
	if respHTMX.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for htmx.min.js, got %d", respHTMX.StatusCode)
	}

	respCyto, err := client.Get(baseURL + "/static/js/cytoscape.min.js")
	if err != nil {
		t.Fatalf("failed fetching cytoscape.min.js: %v", err)
	}
	respCyto.Body.Close()
	if respCyto.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for cytoscape.min.js, got %d", respCyto.StatusCode)
	}

	// 3. Verify /api/graph topology JSON
	respGraph, err := client.Get(baseURL + "/api/graph")
	if err != nil {
		t.Fatalf("failed fetching /api/graph: %v", err)
	}
	defer respGraph.Body.Close()
	if respGraph.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for /api/graph, got %d", respGraph.StatusCode)
	}

	var graphData handlers.GraphDataJSON
	if err := json.NewDecoder(respGraph.Body).Decode(&graphData); err != nil {
		t.Fatalf("failed parsing graph json: %v", err)
	}

	if len(graphData.Nodes) < 2 {
		t.Errorf("expected at least 2 nodes in graph, got %d", len(graphData.Nodes))
	}
	if len(graphData.Edges) < 1 {
		t.Errorf("expected at least 1 edge in graph, got %d", len(graphData.Edges))
	}

	// Verify edge details
	foundEdge := false
	for _, e := range graphData.Edges {
		if e.Source == "prd-e2e-001" && e.Target == "adr-e2e-001" && e.RelationshipType == "depends_on" {
			foundEdge = true
			break
		}
	}
	if !foundEdge {
		t.Errorf("expected prd-e2e-001 -> adr-e2e-001 edge in graph response, got %+v", graphData.Edges)
	}
}
