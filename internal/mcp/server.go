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

// NewServer initializes a new MCP stdio server using default configuration resolution.
func NewServer() (*Server, error) {
	application, err := app.New()
	if err != nil {
		return nil, fmt.Errorf("failed creating app container: %w", err)
	}
	return newServerFromApp(application)
}

// NewServerWithProject initializes an MCP stdio server using an explicit project directory.
func NewServerWithProject(projectDir string) (*Server, error) {
	application, err := app.NewWithProject(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed creating app container: %w", err)
	}
	return newServerFromApp(application)
}

// NewServerWithConfig initializes an MCP stdio server using an explicit config file path.
func NewServerWithConfig(configPath string) (*Server, error) {
	application, err := app.NewWithConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed creating app container: %w", err)
	}
	return newServerFromApp(application)
}

func newServerFromApp(application *app.App) (*Server, error) {
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

	tools.Register(mcpServer, engine)
	s.registerResources()
	s.registerPrompts()

	return s, nil
}

// MCPServer returns the underlying MCP server instance.
func (s *Server) MCPServer() *server.MCPServer {
	return s.mcpServer
}

// Engine returns the underlying lifecycle Engine instance.
func (s *Server) Engine() *lifecycle.Engine {
	return s.engine
}

// Start runs the MCP server over standard input/output.
func (s *Server) Start(ctx context.Context) error {
	return server.ServeStdio(s.mcpServer)
}

// ServeStdio runs the MCP server over standard input/output (alias for Start).
func (s *Server) ServeStdio(ctx context.Context) error {
	return s.Start(ctx)
}
