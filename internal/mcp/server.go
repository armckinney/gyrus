package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
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

	// Pin an entire request to its selection, including multi-step link/archive
	// operations. A switch waits for in-flight tools, resources, and prompts.
	var selection sync.RWMutex
	mcpServer := server.NewMCPServer("gyrus-memory", "1.0.0",
		server.WithToolHandlerMiddleware(func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
			return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				if req.Params.Name == "gyrus_git_switch" {
					selection.Lock()
					defer selection.Unlock()
				} else {
					selection.RLock()
					defer selection.RUnlock()
				}
				return next(ctx, req)
			}
		}),
		server.WithResourceHandlerMiddleware(func(next server.ResourceHandlerFunc) server.ResourceHandlerFunc {
			return func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				selection.RLock()
				defer selection.RUnlock()
				return next(ctx, req)
			}
		}),
		server.WithPromptHandlerMiddleware(func(next server.PromptHandlerFunc) server.PromptHandlerFunc {
			return func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
				selection.RLock()
				defer selection.RUnlock()
				return next(ctx, req)
			}
		}),
	)

	s := &Server{
		mcpServer:   mcpServer,
		storageRoot: application.StorageRoot(),
		app:         application,
		engine:      engine,
	}

	tools.Register(mcpServer, engine)
	tools.RegisterGit(mcpServer, application.ConceptStore)
	s.registerResources()
	s.registerPrompts()

	return s, nil
}

// NewServerFromApp retains the supplied invocation's runtime branch selection.
func NewServerFromApp(application *app.App) (*Server, error) { return newServerFromApp(application) }

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
