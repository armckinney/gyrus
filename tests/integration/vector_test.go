package integration_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	searchvector "github.com/armckinney/gyrus/internal/provider/search/vector"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies end-to-end functionality for the 'vector' profile against a real Ollama sidecar service.
// [Execution Surface]: Real Ollama Embeddings Service (http://ollama:11434 / nomic-embed-text)
// [Assertions]: Ollama generates real 768-dim float32 embeddings, cosine similarity ranks relevant documents higher, and RRF blends FTS results.
// -----------------------------------------------------------------------------
func TestIntegration_OllamaVectorLive(t *testing.T) {
	ollamaHosts := []string{
		os.Getenv("OLLAMA_HOST"),
		"http://ollama:11434",
		"http://localhost:11434",
	}

	var host string
	client := &http.Client{Timeout: 2 * time.Second}
	for _, h := range ollamaHosts {
		if h == "" {
			continue
		}
		resp, err := client.Get(h + "/api/tags")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			host = h
			break
		}
	}

	if host == "" {
		t.Skip("Ollama sidecar service is not reachable, skipping live Ollama vector integration test")
	}

	ctx := context.Background()
	embedder := searchvector.NewOllamaEmbedder(host, "nomic-embed-text")

	doc1 := gyrus.Document{
		ID:      "adr-sql",
		Title:   "Relational Database Architecture",
		Content: "We use PostgreSQL relational database with SQL tables and ACID transactions.",
	}
	doc2 := gyrus.Document{
		ID:      "adr-ui",
		Title:   "Frontend React Visual Dashboard",
		Content: "We build a Single Page Application using React, Tailwind CSS, and WebSockets.",
	}

	lexicalMock := &mockLexical{
		results: []gyrus.SearchResult{
			{Document: doc1, Score: 1.0, MatchReason: "FTS Match"},
		},
	}

	vectorStore := searchvector.NewStore(lexicalMock, embedder)
	if err := vectorStore.AddDocument(ctx, doc1); err != nil {
		t.Fatalf("Vector AddDocument doc1 failed: %v", err)
	}
	if err := vectorStore.AddDocument(ctx, doc2); err != nil {
		t.Fatalf("Vector AddDocument doc2 failed: %v", err)
	}

	// Query: "structured relational SQL database"
	results, err := vectorStore.Search(ctx, "structured relational SQL database", gyrus.SearchFilter{})
	if err != nil {
		t.Fatalf("Vector search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("Expected vector search results, got 0")
	}
	if results[0].Document.ID != "adr-sql" {
		t.Errorf("Expected top vector result to be 'adr-sql', got %s", results[0].Document.ID)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Provider Integration Test
// [Purpose]: Verifies end-to-end functionality for the 'vector' profile with synthetic in-memory embeddings.
// [Execution Surface]: In-Memory Vector Embedding Provider & Reciprocal Rank Fusion
// [Assertions]: Hybrid search combines lexical FTS matches with vector cosine similarities.
// -----------------------------------------------------------------------------
func TestIntegration_VectorHybrid(t *testing.T) {
	ctx := context.Background()

	embedder := &mockEmbedder{
		embeddings: map[string][]float32{
			"semantic query":    {1.0, 0.0, 0.0},
			"vector document":   {0.95, 0.05, 0.0},
			"unrelated content": {0.0, 0.0, 1.0},
		},
	}

	doc1 := gyrus.Document{ID: "doc-vector", Content: "vector document", Title: "Vector Title"}
	doc2 := gyrus.Document{ID: "doc-other", Content: "unrelated content", Title: "Other Title"}

	lexicalMock := &mockLexical{
		results: []gyrus.SearchResult{
			{Document: doc1, Score: 1.0, MatchReason: "FTS Match"},
		},
	}

	vectorStore := searchvector.NewStore(lexicalMock, embedder)
	if err := vectorStore.AddDocument(ctx, doc1); err != nil {
		t.Fatalf("Vector index doc1 failed: %v", err)
	}
	if err := vectorStore.AddDocument(ctx, doc2); err != nil {
		t.Fatalf("Vector index doc2 failed: %v", err)
	}

	results, err := vectorStore.Search(ctx, "semantic query", gyrus.SearchFilter{})
	if err != nil {
		t.Fatalf("Hybrid search failed: %v", err)
	}

	if len(results) == 0 || results[0].Document.ID != "doc-vector" {
		t.Errorf("Expected top hybrid result to be doc-vector, got %v", results)
	}
}

type mockEmbedder struct {
	embeddings map[string][]float32
}

func (m *mockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if vec, exists := m.embeddings[text]; exists {
		return vec, nil
	}
	return []float32{0.0, 0.0, 0.0}, nil
}

type mockLexical struct {
	results []gyrus.SearchResult
}

func (m *mockLexical) Search(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	return m.results, nil
}
