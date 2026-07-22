package mcp

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/armckinney/gyrus/internal/provider/localfs"
	"github.com/armckinney/gyrus/internal/provider/sqlite"
	"github.com/mark3labs/mcp-go/server"
)

// Server encapsulates the Gyrus MCP Server.
type Server struct {
	mcpServer   *server.MCPServer
	storageRoot string
	store       *localfs.Store
	indexer     *sqlite.Indexer
}

// NewServer initializes a new MCP stdio server targeting storageRoot.
func NewServer(storageRoot string) (*Server, error) {
	absRoot, err := localfs.ResolveStoragePath(storageRoot)
	if err != nil {
		return nil, fmt.Errorf("failed resolving storage root: %w", err)
	}

	store, err := localfs.NewStore(absRoot)
	if err != nil {
		return nil, fmt.Errorf("failed initializing localfs store: %w", err)
	}

	dbPath := filepath.Join(absRoot, "index.db")
	indexer, err := sqlite.NewIndexer(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed initializing sqlite indexer: %w", err)
	}

	mcpServer := server.NewMCPServer("gyrus-memory", "1.0.0")

	s := &Server{
		mcpServer:   mcpServer,
		storageRoot: absRoot,
		store:       store,
		indexer:     indexer,
	}

	s.registerTools()
	s.registerResources()
	s.registerPrompts()

	return s, nil
}

// ServeStdio starts serving MCP requests over stdio.
func (s *Server) ServeStdio(ctx context.Context) error {
	defer s.indexer.Close()
	return server.ServeStdio(s.mcpServer)
}
