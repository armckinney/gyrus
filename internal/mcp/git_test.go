package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	storagegit "github.com/armckinney/gyrus/internal/provider/storage/git"
	"github.com/armckinney/gyrus/pkg/gyrus"
	gogit "github.com/go-git/go-git/v5"
	sdk "github.com/mark3labs/mcp-go/mcp"
)

func callGitTool(t *testing.T, s *Server, name string, args map[string]any, wantError bool) string {
	t.Helper()
	request, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": map[string]any{"name": name, "arguments": args}})
	response := s.MCPServer().HandleMessage(context.Background(), request)
	data, _ := json.Marshal(response)
	var envelope struct {
		Result sdk.CallToolResult `json:"result"`
		Error  any                `json:"error"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error != nil || envelope.Result.IsError != wantError {
		t.Fatalf("%s: %s", name, data)
	}
	if len(envelope.Result.Content) == 0 {
		t.Fatalf("empty %s response", name)
	}
	return envelope.Result.Content[0].(sdk.TextContent).Text
}

func TestGitMCPSessionSwitching(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "remote.git")
	if _, err := gogit.PlainInit(remote, true); err != nil {
		t.Fatal(err)
	}
	trunk, err := storagegit.NewStore(storagegit.Options{RepoURL: remote})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = trunk.Create(context.Background(), gyrus.Concept{ID: "shared", Title: "trunk", Type: gyrus.TypeGuide, Category: gyrus.CategoryTechnical, Tenant: "root", Status: "active"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), ".gyrus.yaml")
	original := fmt.Sprintf("storage:\n  provider: git\ngit:\n  repo_url: %q\n  trunk_branch: main\n  commit_strategy: branching\n", remote)
	if err = os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	first, err := NewServerWithConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewServerWithConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range []*Server{first, second} {
		status := callGitTool(t, s, "gyrus_git_status", nil, false)
		if !strings.Contains(status, `"selection_required":true`) {
			t.Fatal(status)
		}
		for _, name := range []string{"gyrus_schema_list", "gyrus_schema_get", "gyrus_search"} {
			callGitTool(t, s, name, map[string]any{"type": "adr", "query": ""}, true)
		}
	}
	for i, s := range []*Server{first, second} {
		branch := fmt.Sprintf("initiative/%d", i)
		callGitTool(t, s, "gyrus_git_switch", map[string]any{"strategy": "branching", "branch": branch}, false)
		callGitTool(t, s, "gyrus_update", map[string]any{"id": "root/reference/shared", "title": branch, "expected_version": 1}, false)
	}
	for i, s := range []*Server{first, second} {
		result := callGitTool(t, s, "gyrus_get", map[string]any{"id": "root/reference/shared"}, false)
		if !strings.Contains(result, fmt.Sprintf("initiative/%d", i)) {
			t.Fatal(result)
		}
		callGitTool(t, s, "gyrus_git_switch", map[string]any{"strategy": "branching", "branch": "main"}, true)
	}
	// Exercise real server middleware concurrently with multi-step graph mutations.
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			callGitTool(t, first, "gyrus_git_status", nil, false)
			callGitTool(t, first, "gyrus_search", map[string]any{"query": ""}, false)
		}()
	}
	callGitTool(t, first, "gyrus_git_switch", map[string]any{"strategy": "direct"}, false)
	wg.Wait()
	result := callGitTool(t, first, "gyrus_get", map[string]any{"id": "root/reference/shared"}, false)
	if !strings.Contains(result, "title: trunk") {
		t.Fatal(result)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != original {
		t.Fatal("runtime switches changed configuration")
	}
	restarted, err := NewServerWithConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(callGitTool(t, restarted, "gyrus_git_status", nil, false), `"selection_required":true`) {
		t.Fatal("runtime branch persisted across restart")
	}
}
