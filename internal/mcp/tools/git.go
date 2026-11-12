package tools

import (
	"context"
	"encoding/json"
	"fmt"

	storagegit "github.com/armckinney/gyrus/internal/provider/storage/git"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func RegisterGit(s *server.MCPServer, storage func() (gyrus.ConceptStore, error)) {
	getStore := func() (*storagegit.Store, error) {
		store, err := storage()
		if err != nil {
			return nil, err
		}
		git, ok := store.(*storagegit.Store)
		if !ok {
			return nil, fmt.Errorf("Git operations require Git storage")
		}
		return git, nil
	}
	s.AddTool(mcp.NewTool("gyrus_git_status",
		mcp.WithDescription("Show this server session's Git strategy, trunk, and active initiative branch"),
		mcp.WithReadOnlyHintAnnotation(true),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		store, err := getStore()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, _ := json.Marshal(store.Status())
		return mcp.NewToolResultText(string(data)), nil
	})
	s.AddTool(mcp.NewTool("gyrus_git_switch",
		mcp.WithDescription("Switch only this server session's Git selection. Never saves configuration, pushes branches, or merges. Branching requires an initiative branch distinct from trunk."),
		mcp.WithString("strategy", mcp.Required(), mcp.Enum("direct", "branching")),
		mcp.WithString("branch", mcp.Description("Initiative branch; required for branching and omitted for direct")),
		mcp.WithDestructiveHintAnnotation(false),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		store, err := getStore()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		status, err := store.Switch(ctx, getArgString(req, "strategy"), getArgString(req, "branch"))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		data, _ := json.Marshal(status)
		return mcp.NewToolResultText(string(data)), nil
	})
}
