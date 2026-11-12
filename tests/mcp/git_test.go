package mcp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	storagegit "github.com/armckinney/gyrus/internal/provider/storage/git"
	"github.com/armckinney/gyrus/pkg/gyrus"
	gogit "github.com/go-git/go-git/v5"
)

func TestMCPRuntimeBranchWithExplicitConfiguration(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "remote.git")
	if _, err := gogit.PlainInit(remote, true); err != nil {
		t.Fatal(err)
	}
	store, err := storagegit.NewStore(storagegit.Options{RepoURL: remote})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = store.Create(context.Background(), gyrus.Concept{ID: "shared", Title: "Trunk", Type: gyrus.TypeGuide, Category: gyrus.CategoryTechnical, Tenant: "root", Status: "active"}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), ".gyrus.yaml")
	original := fmt.Sprintf("storage:\n  provider: git\ngit:\n  repo_url: %q\n  commit_strategy: direct\n", remote)
	if err = os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	client := startMCPClient(t, t.TempDir(), "--config", path, "--git-branch", "initiative/subprocess")
	client.Call(t, "initialize", map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{}, "clientInfo": map[string]any{"name": "test-runner"}})
	client.Notify(t, "notifications/initialized", map[string]any{})
	response := client.Call(t, "tools/call", map[string]any{"name": "gyrus_git_status", "arguments": map[string]any{}})
	var result struct {
		IsError bool `json:"isError"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err = json.Unmarshal(response.Result, &result); err != nil {
		t.Fatal(err)
	}
	if response.Error != nil || result.IsError || len(result.Content) != 1 || !strings.Contains(result.Content[0].Text, `"active_branch":"initiative/subprocess"`) {
		t.Fatalf("runtime branch was not forwarded: %+v", response)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != original {
		t.Fatal("MCP startup modified configuration")
	}
}
