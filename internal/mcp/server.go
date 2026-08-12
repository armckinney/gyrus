package mcp

import (
	"context"
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/mark3labs/mcp-go/server"
)

// Server encapsulates the Gyrus MCP Server.
type Server struct {
	mcpServer   *server.MCPServer
	storageRoot string
	app         *app.App
	engine      *lifecycle.Engine
}

// NewServer initializes a new MCP stdio server targeting storageRoot.
func NewServer(storageRoot string) (*Server, error) {
	application, err := app.New(storageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed creating app container: %w", err)
	}

	engine, err := application.Engine()
	if err != nil {
		return nil, fmt.Errorf("failed initializing lifecycle engine: %w", err)
	}

	mcpServer := server.NewMCPServer("gyrus-memory", "1.0.0")

	s := &Server{
		mcpServer:   mcpServer,
		storageRoot: application.StorageRoot(),
		app:         application,
		engine:      engine,
	}

	s.registerTools()
	s.registerResources()
	s.registerPrompts()

	return s, nil
}

// ServeStdio starts serving MCP requests over stdio.
func (s *Server) ServeStdio(ctx context.Context) error {
	return server.ServeStdio(s.mcpServer)
}
