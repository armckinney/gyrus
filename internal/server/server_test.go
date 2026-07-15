package server

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/template-go/internal/models"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		wantAddr string
	}{
		{
			name:     "Custom Port 8080",
			port:     8080,
			wantAddr: ":8080",
		},
		{
			name:     "Custom Port 4000",
			port:     4000,
			wantAddr: ":4000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer(tt.port, &mockDB{}, &mockRepo{})

			if srv.Addr != tt.wantAddr {
				t.Errorf("NewServer().Addr = %v, want %v", srv.Addr, tt.wantAddr)
			}
		})
	}
}

type mockDB struct{}

func (m *mockDB) Health() map[string]string {
	return map[string]string{"status": "OK"}
}

func (m *mockDB) Close() error {
	return nil
}

func (m *mockDB) GetDB() *sql.DB {
	return nil
}

type mockRepo struct{}

func (m *mockRepo) GetAll(ctx context.Context) ([]models.User, error) {
	return []models.User{}, nil
}

func (m *mockRepo) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	return user, nil
}

func TestRegisterRoutes(t *testing.T) {
	s := &Server{db: &mockDB{}, userRepo: &mockRepo{}}
	handler := s.RegisterRoutes()

	// Create a test server using the handler
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Test "/" route
	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("could not send GET request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("got status %d; want %d", res.StatusCode, http.StatusOK)
	}

	// Test "/health" route
	resHealth, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("could not send GET request: %v", err)
	}
	defer resHealth.Body.Close()

	if resHealth.StatusCode != http.StatusOK {
		t.Errorf("got status %d; want %d", resHealth.StatusCode, http.StatusOK)
	}
}
