package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/armckinney/gyrus/internal/app"
	"github.com/armckinney/gyrus/internal/domain/lifecycle"
	"github.com/armckinney/gyrus/internal/ui/handlers"
	"github.com/armckinney/gyrus/internal/ui/markdown"
	"github.com/armckinney/gyrus/internal/ui/static"
	"github.com/armckinney/gyrus/internal/ui/templates"
)

// Server coordinates the embedded HTTP dashboard and API services for Gyrus.
type Server struct {
	app        *app.App
	engine     *lifecycle.Engine
	host       string
	port       int
	httpServer *http.Server
	handlers   *handlers.Handlers
}

// NewServer initializes a Web UI server resolving configuration from default paths.
func NewServer(host string, port int) (*Server, error) {
	application, err := app.New()
	if err != nil {
		return nil, fmt.Errorf("failed creating app container: %w", err)
	}
	return NewServerFromApp(application, host, port)
}

// NewServerWithWorkspace initializes a Web UI server for an explicit workspace directory.
func NewServerWithWorkspace(workspaceDir string, host string, port int) (*Server, error) {
	application, err := app.NewWithWorkspace(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("failed creating app container for workspace '%s': %w", workspaceDir, err)
	}
	return NewServerFromApp(application, host, port)
}

// NewServerWithConfig initializes a Web UI server using an explicit configuration file.
func NewServerWithConfig(configPath string, host string, port int) (*Server, error) {
	application, err := app.NewWithConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed creating app container with config '%s': %w", configPath, err)
	}
	return NewServerFromApp(application, host, port)
}

// NewServerFromApp creates a Web UI server from an existing app.App container.
func NewServerFromApp(application *app.App, host string, port int) (*Server, error) {
	engine, err := application.Engine()
	if err != nil {
		return nil, fmt.Errorf("failed initializing lifecycle engine: %w", err)
	}

	tmplMgr, err := templates.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed initializing template manager: %w", err)
	}

	renderer := markdown.NewRenderer()
	h := handlers.New(application, engine, tmplMgr, renderer)

	if host == "" {
		host = "127.0.0.1"
	}
	if port <= 0 {
		port = 8080
	}

	return &Server{
		app:      application,
		engine:   engine,
		host:     host,
		port:     port,
		handlers: h,
	}, nil
}

// Handler constructs the HTTP request router for the Web UI.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Static asset routing
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(static.FS())))

	// API endpoints
	mux.HandleFunc("/api/graph", s.handlers.APIGraph)

	// Web UI pages & fragments
	mux.HandleFunc("/search", s.handlers.Search)
	mux.HandleFunc("/graph", s.handlers.Graph)
	mux.HandleFunc("/docs/", s.handlers.Doc)
	mux.HandleFunc("/docs", s.handlers.Home)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.handlers.Home(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/docs/") {
			s.handlers.Doc(w, r)
			return
		}
		http.NotFound(w, r)
	})

	return mux
}

// Start launches the HTTP server and blocks until the context is canceled or a fatal error occurs.
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully stops the running HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
