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

func TestVectorIdentityIncludesTenantAndScope(t *testing.T) {
	store := NewStore(nil, &MockEmbedder{EmbedFunc: func(string) []float32 { return []float32{1, 0} }})
	for _, tenant := range []string{"one", "two"} {
		if err := store.AddConcept(context.Background(), gyrus.Concept{ID: "same", Tenant: tenant, Scope: "reference"}); err != nil {
			t.Fatal(err)
		}
	}
	results, err := store.Search(context.Background(), "same", gyrus.SearchFilter{})
	if err != nil || len(results) != 2 || results[0].Concept.Name == results[1].Concept.Name {
		t.Fatalf("qualified vector identities collapsed: %+v %v", results, err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that Hybrid Search fuses lexical and vector rankings using Reciprocal Rank Fusion (RRF).
// [Execution Surface]: In-Memory Vector Store + Mock Embedder
// [Assertions]: RRF scores are calculated accurately using standard constant k=60 across lexical and vector ranks.
// -----------------------------------------------------------------------------
func TestRRFAndHybridSearch(t *testing.T) {
	embedder := &MockEmbedder{
		EmbedFunc: func(text string) []float32 {
			if text == "query" {
				return []float32{1.0, 0.0, 0.0}
			}
			if text == "root/reference/doc1" {
				return []float32{0.9, 0.1, 0.0}
			}
			if text == "root/reference/doc2" {
				return []float32{0.0, 1.0, 0.0}
			}
			return []float32{0.0, 0.0, 1.0}
		},
	}

	doc1 := gyrus.Concept{ID: "doc1", Content: "root/reference/doc1", Tenant: "root", Name: "root/reference/doc1"}
	doc2 := gyrus.Concept{ID: "doc2", Content: "root/reference/doc2",

		// Lexical gives doc2 rank 1, doc1 rank 2
		Tenant: "root", Name: "root/reference/doc2"}

	lexical := &MockLexicalSearch{
		Results: []gyrus.SearchResult{
			{Concept: doc2, Score: 0.9, MatchReason: "bm25"},
			{Concept: doc1, Score: 0.5, MatchReason: "bm25"},
		},
	}

	store := NewStore(lexical, embedder)
	err := store.AddConcept(context.Background(), doc1)
	if err != nil {
		t.Fatalf("failed to add doc1: %v", err)
	}
	err = store.AddConcept(context.Background(), doc2)
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
	score1 := 1.0/62.0 + 1.0/61.0
	score2 := 1.0/61.0 + 1.0/62.0

	for _, res := range results {
		var expected float64
		if res.Concept.ID == "doc1" {
			expected = score1
		} else {
			expected = score2
		}
		if math.Abs(res.Score-expected) > 1e-6 {
			t.Errorf("expected score %f for %s, got %f", expected, res.Concept.ID, res.Score)
		}
	}
}
