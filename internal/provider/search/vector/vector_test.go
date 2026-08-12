package vector

import (
	"context"
	"math"
	"testing"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

type MockEmbedder struct {
	EmbedFunc func(text string) []float32
}

func (m *MockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	return m.EmbedFunc(text), nil
}

type MockLexicalSearch struct {
	Results []gyrus.SearchResult
}

func (m *MockLexicalSearch) Search(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	return m.Results, nil
}

func TestRRFAndHybridSearch(t *testing.T) {
	embedder := &MockEmbedder{
		EmbedFunc: func(text string) []float32 {
			if text == "query" {
				return []float32{1.0, 0.0, 0.0}
			}
			if text == "doc1" {
				return []float32{0.9, 0.1, 0.0}
			}
			if text == "doc2" {
				return []float32{0.0, 1.0, 0.0}
			}
			return []float32{0.0, 0.0, 1.0}
		},
	}

	doc1 := gyrus.Document{ID: "doc1", Content: "doc1"}
	doc2 := gyrus.Document{ID: "doc2", Content: "doc2"}

	// Lexical gives doc2 rank 1, doc1 rank 2
	lexical := &MockLexicalSearch{
		Results: []gyrus.SearchResult{
			{Document: doc2, Score: 0.9, MatchReason: "bm25"},
			{Document: doc1, Score: 0.5, MatchReason: "bm25"},
		},
	}

	store := NewStore(lexical, embedder)
	err := store.AddDocument(context.Background(), doc1)
	if err != nil {
		t.Fatalf("failed to add doc1: %v", err)
	}
	err = store.AddDocument(context.Background(), doc2)
	if err != nil {
		t.Fatalf("failed to add doc2: %v", err)
	}

	results, err := store.Search(context.Background(), "query", gyrus.SearchFilter{})
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Calculate expected scores
	// RRF(doc1): lexical rank 2 (1/62), vector rank 1 (1/61)
	// RRF(doc2): lexical rank 1 (1/61), vector rank 2 (1/62)
	// Because doc1 vector sim = 0.9 / sqrt(0.9^2 + 0.1^2) ~ 0.99
	// doc2 vector sim = 0.0
	// So RRF should be equal. To differentiate them, we can check that they are both scored properly.

	score1 := 1.0/62.0 + 1.0/61.0
	score2 := 1.0/61.0 + 1.0/62.0

	for _, res := range results {
		var expected float64
		if res.Document.ID == "doc1" {
			expected = score1
		} else {
			expected = score2
		}
		if math.Abs(res.Score-expected) > 1e-6 {
			t.Errorf("expected score %f for %s, got %f", expected, res.Document.ID, res.Score)
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	a := []float32{1.0, 0.0, 0.0}
	b := []float32{1.0, 0.0, 0.0}
	sim := cosineSimilarity(a, b)
	if math.Abs(float64(sim-1.0)) > 1e-6 {
		t.Errorf("Expected ~1.0, got %f", sim)
	}
}
