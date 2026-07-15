package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/template-go/internal/models"
)

type mockUserRepo struct {
	users []models.User
	err   error
}

func (m *mockUserRepo) GetAll(ctx context.Context) ([]models.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.users, nil
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	if m.err != nil {
		return models.User{}, m.err
	}
	user.ID = 1 // Mock ID
	return user, nil
}

func TestGetAllUsers(t *testing.T) {
	tests := []struct {
		name           string
		mockUsers      []models.User
		mockErr        error
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "Success",
			mockUsers: []models.User{
				{ID: 1, Username: "test1", Email: "test1@example.com"},
				{ID: 2, Username: "test2", Email: "test2@example.com"},
			},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Repository Error",
			mockUsers:      nil,
			mockErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepo{
				users: tt.mockUsers,
				err:   tt.mockErr,
			}
			h := &UserHandler{Repo: mockRepo}

			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			w := httptest.NewRecorder()

			h.GetAll(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var users []models.User
				if err := json.NewDecoder(w.Body).Decode(&users); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if len(users) != tt.expectedCount {
					t.Errorf("expected %d users, got %d", tt.expectedCount, len(users))
				}

				if w.Header().Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
				}
			}
		})
	}
}
