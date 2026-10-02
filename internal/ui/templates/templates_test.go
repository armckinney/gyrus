package templates

import (
	"bytes"
	"testing"
)

func TestNewManager_RenderAllViews(t *testing.T) {
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	testCases := []struct {
		view   string
		isHTMX bool
		data   any
	}{
		{
			view:   "index.html",
			isHTMX: false,
			data: map[string]any{
				"Title":          "Index",
				"StorageBackend": "localfs",
				"TotalDocs":      0,
			},
		},
		{
			view:   "index.html",
			isHTMX: true,
			data: map[string]any{
				"FilterTitle": "Test",
			},
		},
		{
			view:   "doc.html",
			isHTMX: false,
			data: map[string]any{
				"Title":          "Doc Test",
				"StorageBackend": "postgres",
				"Doc": map[string]any{
					"ID":       "adr-001",
					"Title":    "ADR Title",
					"Category": "architecture",
					"Type":     "adr",
					"Status":   "accepted",
					"Version":  1,
				},
				"ContentHTML": "<h1>Hello</h1>",
			},
		},
		{
			view:   "search_results.html",
			isHTMX: true,
			data: map[string]any{
				"SearchQuery": "test",
			},
		},
		{
			view:   "graph.html",
			isHTMX: false,
			data: map[string]any{
				"Title":          "Graph",
				"StorageBackend": "localfs",
				"FocusDocID":     "adr-001",
			},
		},
	}

	for _, tc := range testCases {
		var buf bytes.Buffer
		if err := mgr.Render(&buf, tc.view, tc.data, tc.isHTMX); err != nil {
			t.Errorf("failed rendering %s (isHTMX=%v): %v", tc.view, tc.isHTMX, err)
		}
		if buf.Len() == 0 {
			t.Errorf("expected non-empty output for %s", tc.view)
		}
	}
}
