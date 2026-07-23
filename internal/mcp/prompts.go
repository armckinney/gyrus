package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerPrompts() {
	adrPrompt := mcp.NewPrompt(
		"prepare-adr",
		mcp.WithPromptDescription("Generates an Architecture Design Record (ADR) structural template"),
		mcp.WithArgument("title", mcp.ArgumentDescription("Title of the ADR proposal"), mcp.RequiredArgument()),
	)
	s.mcpServer.AddPrompt(adrPrompt, s.handlePrepareADRPrompt)
}

func (s *Server) handlePrepareADRPrompt(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	title, _ := req.Params.Arguments["title"]
	content := fmt.Sprintf(`---
id: adr-2026-new-proposal
title: "%s"
category: architecture
type: adr
owner_group: platform
version: 1
status: proposed
tags:
  - architecture
dependencies: []
---

# %s

## Context & Problem Statement
Describe the decision context.

## Proposed Decision
Outline the architectural choice.

## Consequences
Detail positive and negative trade-offs.
`, title, title)

	return &mcp.GetPromptResult{
		Description: "ADR Template",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: content,
				},
			},
		},
	}, nil
}
