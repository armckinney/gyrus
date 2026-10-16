package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerResources() {
	conceptTemplate := mcp.NewResourceTemplate(
		"memory://concept/{id}",
		"OKF Concept Content",
		mcp.WithTemplateDescription("Reads an Open Knowledge Format contract concept by ID"),
		mcp.WithTemplateMIMEType("text/markdown"),
	)
	s.mcpServer.AddResourceTemplate(conceptTemplate, s.handleReadResource)

	schemaTemplate := mcp.NewResourceTemplate(
		"memory://schema/{concept_type}",
		"OKF Schema Template",
		mcp.WithTemplateDescription("Retrieves the OKF contract schema template for a concept type"),
		mcp.WithTemplateMIMEType("text/markdown"),
	)
	s.mcpServer.AddResourceTemplate(schemaTemplate, s.handleReadResource)

	schemasResource := mcp.NewResource(
		"memory://schemas",
		"OKF Available Schemas",
		mcp.WithResourceDescription("Lists all available schemas across persistence layer and embedded templates"),
		mcp.WithMIMEType("application/json"),
	)
	s.mcpServer.AddResource(schemasResource, s.handleReadResource)
}

func (s *Server) handleReadResource(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := req.Params.URI

	// 1. memory://schemas list
	if uri == "memory://schemas" {
		schemas, err := s.engine.ListSchemas(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed listing schemas: %w", err)
		}
		data, err := json.MarshalIndent(schemas, "", "  ")
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      uri,
				MIMEType: "application/json",
				Text:     string(data),
			},
		}, nil
	}

	// 2. memory://schema/{concept_type}
	if strings.HasPrefix(uri, "memory://schema/") {
		conceptType := strings.TrimPrefix(uri, "memory://schema/")
		if conceptType == "" {
			return nil, fmt.Errorf("invalid schema URI: %s", uri)
		}

		template, err := s.engine.GetSchema(ctx, conceptType)
		if err != nil {
			return nil, fmt.Errorf("schema not found for type '%s': %w", conceptType, err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      uri,
				MIMEType: "text/markdown",
				Text:     template,
			},
		}, nil
	}

	// 3. memory://concept/{id}
	if strings.HasPrefix(uri, "memory://concept/") {
		id := strings.TrimPrefix(uri, "memory://concept/")
		if id == "" {
			return nil, fmt.Errorf("invalid concept URI: %s", uri)
		}

		doc, err := s.engine.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("concept not found: %w", err)
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

	return nil, fmt.Errorf("unsupported resource URI: %s", uri)
}
