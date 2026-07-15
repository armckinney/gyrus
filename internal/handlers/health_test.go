package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestHandler_Health(t *testing.T) {
	db := &mockDB{}
	handler := Health(db)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d; want %d", w.Code, http.StatusOK)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("cannot decode response: %v", err)
	}

	if response["status"] != "OK" {
		t.Errorf("got status %s; want %s", response["status"], "OK")
	}
}
