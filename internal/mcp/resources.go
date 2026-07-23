package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/armckinney/gyrus/internal/okf"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerResources() {
	docResource := mcp.NewResource(
		"memory://doc/{id}",
		"OKF Document Content",
		mcp.WithResourceDescription("Reads an Open Knowledge Format contract document by ID"),
		mcp.WithMIMEType("text/markdown"),
	)
	s.mcpServer.AddResource(docResource, s.handleReadResource)
}

func (s *Server) handleReadResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := req.Params.URI
	id := strings.TrimPrefix(uri, "memory://doc/")
	if id == "" {
		return nil, fmt.Errorf("invalid document URI: %s", uri)
	}

	doc, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	data, err := okf.SerializeMarkdown(&doc)
	if err != nil {
		return nil, err
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     string(data),
		},
	}, nil
}
