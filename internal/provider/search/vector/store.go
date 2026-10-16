package vector

import (
	"context"
	"math"
	"sort"
	"strings"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

type EmbeddingProvider interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// VectorConcept stores a concept along with its vector embedding
type VectorConcept struct {
	gyrus.Concept
	Vector []float32
}

// Store implements gyrus.SearchProvider
type Store struct {
	LexicalProvider gyrus.SearchProvider
	Embedder        EmbeddingProvider
	Vectors         []VectorConcept // In-memory store for local cosine similarity
}

func NewStore(lexical gyrus.SearchProvider, embedder EmbeddingProvider) *Store {
	return &Store{
		LexicalProvider: lexical,
		Embedder:        embedder,
		Vectors:         make([]VectorConcept, 0),
	}
}

// AddConcept is used to index a concept in the vector store
func (s *Store) AddConcept(ctx context.Context, doc gyrus.Concept) error {
	if err := okf.SetName(&doc, ""); err != nil {
		return err
	}
	vec, err := s.Embedder.Embed(ctx, doc.Content)
	if err != nil {
		return err
	}
	s.Vectors = append(s.Vectors, VectorConcept{
		Concept: doc,
		Vector:  vec,
	})
	return nil
}

func (s *Store) Search(ctx context.Context, request gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	query, filter := request.Query, request.Filter
	query = strings.TrimSpace(query)
	if query == "" {
		if s.LexicalProvider != nil {
			return s.LexicalProvider.Search(ctx, request)
		}
		results := make([]gyrus.SearchResult, 0)
		for _, doc := range s.Vectors {
			if matchFilter(doc.Concept, filter) {
				results = append(results, gyrus.SearchResult{Concept: doc.Concept, MatchReason: "All Concepts"})
			}
		}
		if limit := request.Limit(); limit > 0 && len(results) > limit {
			results = results[:limit]
		}
		return results, nil
	}
	// 1. Lexical search
	var lexicalResults []gyrus.SearchResult
	if s.LexicalProvider != nil {
		res, err := s.LexicalProvider.Search(ctx, request)
		if err != nil {
			return nil, err
		}
		lexicalResults = res
	}

	// 2. Vector search
	queryVec, err := s.Embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	var vectorResults []gyrus.SearchResult
	for _, vDoc := range s.Vectors {
		if !matchFilter(vDoc.Concept, filter) {
			continue
		}

		sim := cosineSimilarity(queryVec, vDoc.Vector)
		vectorResults = append(vectorResults, gyrus.SearchResult{
			Concept:     vDoc.Concept,
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
	if limit := request.Limit(); limit > 0 && len(merged) > limit {
		merged = merged[:limit]
	}
	return merged, nil
}

func matchFilter(doc gyrus.Concept, filter gyrus.SearchFilter) bool {
	switch filter.Scope {
	case "all":
	case "reference":
		if doc.Scope != "reference" {
			return false
		}
	case "project":
		if doc.Scope != "project" || filter.Project != "" && doc.Project != filter.Project {
			return false
		}
	default:
		if doc.Scope != "reference" && (filter.Project == "" || doc.Project != filter.Project) {
			return false
		}
	}
	if filter.Category != "" && doc.Category != filter.Category {
		return false
	}
	if filter.Type != "" && doc.Type != filter.Type {
		return false
	}
	if filter.Status != "" && doc.Status != filter.Status {
		return false
	}
	if filter.Tenant != "" && doc.Tenant != filter.Tenant {
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
	docs := make(map[string]gyrus.Concept)
	reasons := make(map[string]string)

	for i, res := range lexical {
		rank := i + 1
		scores[res.Concept.Name] += 1.0 / float64(k+rank)
		docs[res.Concept.Name] = res.Concept
		reasons[res.Concept.Name] = "lexical"
	}

	for i, res := range vector {
		rank := i + 1
		scores[res.Concept.Name] += 1.0 / float64(k+rank)
		docs[res.Concept.Name] = res.Concept
		if existing, ok := reasons[res.Concept.Name]; ok {
			reasons[res.Concept.Name] = existing + ", semantic"
		} else {
			reasons[res.Concept.Name] = "semantic"
		}
	}

	var merged []gyrus.SearchResult
	for id, score := range scores {
		merged = append(merged, gyrus.SearchResult{
			Concept:     docs[id],
			Score:       score,
			MatchReason: "hybrid: " + reasons[id],
		})
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Score > merged[j].Score
	})

	return merged
}
