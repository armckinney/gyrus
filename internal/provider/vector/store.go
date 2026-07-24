package vector

import (
	"context"
	"math"
	"sort"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

type EmbeddingProvider interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorDocument stores a document along with its vector embedding
type VectorDocument struct {
	gyrus.Document
	Vector []float32
}

// Store implements gyrus.SearchProvider
type Store struct {
	LexicalProvider gyrus.SearchProvider
	Embedder        EmbeddingProvider
	Vectors         []VectorDocument // In-memory store for local cosine similarity
}

func NewStore(lexical gyrus.SearchProvider, embedder EmbeddingProvider) *Store {
	return &Store{
		LexicalProvider: lexical,
		Embedder:        embedder,
		Vectors:         make([]VectorDocument, 0),
	}
}

// AddDocument is used to index a document in the vector store
func (s *Store) AddDocument(ctx context.Context, doc gyrus.Document) error {
	vec, err := s.Embedder.Embed(ctx, doc.Content)
	if err != nil {
		return err
	}
	s.Vectors = append(s.Vectors, VectorDocument{
		Document: doc,
		Vector:   vec,
	})
	return nil
}

func (s *Store) Search(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	// 1. Lexical search
	var lexicalResults []gyrus.SearchResult
	if s.LexicalProvider != nil {
		res, err := s.LexicalProvider.Search(ctx, query, filter)
		if err == nil {
			lexicalResults = res
		}
	}

	// 2. Vector search
	queryVec, err := s.Embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	var vectorResults []gyrus.SearchResult
	for _, vDoc := range s.Vectors {
		if !matchFilter(vDoc.Document, filter) {
			continue
		}
		
		sim := cosineSimilarity(queryVec, vDoc.Vector)
		vectorResults = append(vectorResults, gyrus.SearchResult{
			Document:    vDoc.Document,
			Score:       float64(sim),
			MatchReason: "semantic similarity",
		})
	}
	
	// sort vector results by score descending
	sort.Slice(vectorResults, func(i, j int) bool {
		return vectorResults[i].Score > vectorResults[j].Score
	})

	// 3. Reciprocal Rank Fusion (RRF)
	merged := rrf(lexicalResults, vectorResults)
	return merged, nil
}

func matchFilter(doc gyrus.Document, filter gyrus.SearchFilter) bool {
	if filter.Category != "" && doc.Category != filter.Category {
		return false
	}
	if filter.Type != "" && doc.Type != filter.Type {
		return false
	}
	if filter.Status != "" && doc.Status != filter.Status {
		return false
	}
	if filter.OwnerGroup != "" && doc.OwnerGroup != filter.OwnerGroup {
		return false
	}
	if filter.Tag != "" {
		found := false
		for _, t := range doc.Tags {
			if t == filter.Tag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func cosineSimilarity(a, b []float32) float32 {
	var dotProduct, normA, normB float32
	for i := 0; i < len(a) && i < len(b); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// rrf merges two lists of search results using Reciprocal Rank Fusion
func rrf(lexical, vector []gyrus.SearchResult) []gyrus.SearchResult {
	const k = 60
	scores := make(map[string]float64)
	docs := make(map[string]gyrus.Document)
	reasons := make(map[string]string)

	for i, res := range lexical {
		rank := i + 1
		scores[res.Document.ID] += 1.0 / float64(k+rank)
		docs[res.Document.ID] = res.Document
		reasons[res.Document.ID] = "lexical"
	}

	for i, res := range vector {
		rank := i + 1
		scores[res.Document.ID] += 1.0 / float64(k+rank)
		docs[res.Document.ID] = res.Document
		if existing, ok := reasons[res.Document.ID]; ok {
			reasons[res.Document.ID] = existing + ", semantic"
		} else {
			reasons[res.Document.ID] = "semantic"
		}
	}

	var merged []gyrus.SearchResult
	for id, score := range scores {
		merged = append(merged, gyrus.SearchResult{
			Document:    docs[id],
			Score:       score,
			MatchReason: "hybrid: " + reasons[id],
		})
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Score > merged[j].Score
	})

	return merged
}
