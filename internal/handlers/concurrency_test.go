package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler_Concurrency(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/concurrency", nil)
	w := httptest.NewRecorder()

	Concurrency(w, req)

	// Check StatusCode
	if w.Code != http.StatusOK {
		t.Errorf("got status %d; want %d", w.Code, http.StatusOK)
	}

	// Check Content-Type
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("got invalid content type %s; want application/json", contentType)
	}

	// Check Body
	var response struct {
		Count   int      `json:"count"`
		Results []string `json:"results"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("cannot decode response: %v", err)
	}

	expectedCount := 5
	if response.Count != expectedCount {
		t.Errorf("got count %d; want %d", response.Count, expectedCount)
	}

	if len(response.Results) != expectedCount {
		t.Errorf("got results length %d; want %d", len(response.Results), expectedCount)
	}

	// Since concurrency makes order non-deterministic, we just check presence
	// Or simplistic check: verify one known item exists or that all look like "taskX processed"
	for _, res := range response.Results {
		if len(res) < 14 { // "taskX processed" is at least ~15 chars
			t.Errorf("result string too short, likely malformed: %s", res)
		}
	}
}
