package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_HelloWorld(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	HelloWorld(w, req)

	// Check StatusCode
	if w.Code != http.StatusOK {
		t.Errorf("got status %d; want %d", w.Code, http.StatusOK)
	}

	// Check Content-Type
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("got invalid content type %s; want application/json", contentType)
	}

	// Check Body
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("cannot decode response: %v", err)
	}

	expectedMsg := "Hello World"
	if response["message"] != expectedMsg {
		t.Errorf("got message %s; want %s", response["message"], expectedMsg)
	}
}
