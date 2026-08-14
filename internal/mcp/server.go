package mcp

import (
	"context"
	"fmt"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/mcp/tools"
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

	tools.Register(s.mcpServer, s.engine)
	s.registerResources()
	s.registerPrompts()

	return s, nil
}

// MCPServer returns the underlying mark3labs MCPServer instance.
func (s *Server) MCPServer() *server.MCPServer {
	return s.mcpServer
}

// Engine returns the underlying lifecycle Engine service.
func (s *Server) Engine() *lifecycle.Engine {
	return s.engine
}

// ServeStdio starts serving MCP requests over stdio.
func (s *Server) ServeStdio(ctx context.Context) error {
	return server.ServeStdio(s.mcpServer)
}
