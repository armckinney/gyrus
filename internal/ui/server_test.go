package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/armckinney/gyrus/internal/app"
)

func setupTestWorkspace(t *testing.T) (*app.App, string) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "gyrus-ui-server-test-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}

	cfgContent := `
version: 1
project: gyrus
storage:
  provider: localfs
  localfs:
    path: .gyrus/concepts
tenant: test-core
search:
  provider: sqlite_fts5
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".gyrus.yaml"), []byte(cfgContent), 0644); err != nil {
		t.Fatalf("failed writing config: %v", err)
	}

	application, err := app.NewWithProject(tmpDir)
	if err != nil {
		t.Fatalf("failed initializing app: %v", err)
	}

	return application, tmpDir
}

func TestServer_HandlerRouting(t *testing.T) {
	application, tmpDir := setupTestWorkspace(t)
	defer os.RemoveAll(tmpDir)

	server, err := NewServerFromApp(application, "127.0.0.1", 8080)
	if err != nil {
		t.Fatalf("failed creating server: %v", err)
	}

	handler := server.Handler()

	t.Run("Static JS asset route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/js/htmx.min.js", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for static asset, got %d", rec.Code)
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "javascript") {
			t.Errorf("expected javascript content-type, got %s", rec.Header().Get("Content-Type"))
		}
	})

	t.Run("Static CSS asset route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/css/app.css", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for static css, got %d", rec.Code)
		}
		if !strings.Contains(rec.Header().Get("Content-Type"), "css") {
			t.Errorf("expected css content-type, got %s", rec.Header().Get("Content-Type"))
		}
	})

	t.Run("Root path redirects or serves home", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for root, got %d", rec.Code)
		}
	})
}

func TestServer_StartAndShutdown(t *testing.T) {
	application, tmpDir := setupTestWorkspace(t)
	defer os.RemoveAll(tmpDir)

	server, err := NewServerFromApp(application, "127.0.0.1", 18888)
	if err != nil {
		t.Fatalf("failed creating server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() {
		errCh <- server.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("unexpected error on shutdown: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("server failed to shutdown within timeout")
	}
}
