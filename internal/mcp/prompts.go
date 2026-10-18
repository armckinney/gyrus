package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/armckinney/gyrus/internal/format"
	"github.com/mark3labs/mcp-go/mcp"
	"gopkg.in/yaml.v3"
)

func (s *Server) registerPrompts() {
	adrPrompt := mcp.NewPrompt(
		"prepare-adr",
		mcp.WithPromptDescription("Generates an Architecture Decision Record from the active ADR schema template"),
		mcp.WithArgument("title", mcp.ArgumentDescription("Title of the ADR proposal"), mcp.RequiredArgument()),
	)
	s.mcpServer.AddPrompt(adrPrompt, s.handlePrepareADRPrompt)
}

func (s *Server) handlePrepareADRPrompt(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	title := req.Params.Arguments["title"]
	template, err := s.engine.GetSchema(ctx, "adr")
	if err != nil {
		return nil, err
	}
	template = strings.ReplaceAll(template, "\r\n", "\n")
	header, body, ok := strings.Cut(strings.TrimPrefix(template, "---\n"), "\n---")
	if !ok {
		return nil, fmt.Errorf("ADR schema is missing frontmatter")
	}
	var fields map[string]any
	if err := yaml.Unmarshal([]byte(header), &fields); err != nil {
		return nil, fmt.Errorf("parse ADR schema: %w", err)
	}
	fields["title"] = title
	metadata, err := format.MarshalFrontmatter(fields)
	if err != nil {
		return nil, err
	}
	body = strings.ReplaceAll(body, "[Short Title of Decision]", title)
	content := "---\n" + string(metadata) + "---" + body
	return &mcp.GetPromptResult{
		Description: "ADR Template",
		Messages: []mcp.PromptMessage{{
			Role:    mcp.RoleUser,
			Content: mcp.TextContent{Type: "text", Text: content},
		}},
	}, nil
}
