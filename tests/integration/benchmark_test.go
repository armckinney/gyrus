package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	graphokf "github.com/armckinney/gyrus/internal/provider/graph/okf"
	searchokf "github.com/armckinney/gyrus/internal/provider/search/okf"
	"github.com/armckinney/gyrus/internal/provider/storage/localfs"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Performance Benchmark
// [Purpose]: Benchmarks durable creation of OKF concepts.
// [Execution Surface]: Local filesystem with interprocess locking and atomic writes.
// [Assertions]: Measures ns/op and allocs/op for document indexing.
// -----------------------------------------------------------------------------
func BenchmarkDocumentIndexing(b *testing.B) {
	tempDir := b.TempDir()
	store, err := localfs.NewStore(tempDir)
	if err != nil {
		b.Fatalf("Failed creating indexer: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := gyrus.Concept{
			ID:          fmt.Sprintf("bench-doc-%d", i),
			Title:       fmt.Sprintf("Benchmark Document %d", i),
			Category:    gyrus.CategoryTechnical,
			Type:        gyrus.TypeSpecification,
			Tenant:      "benchmarks",
			Version:     1,
			Status:      "active",
			LastUpdated: time.Now(),
			Content:     "High performance benchmark document body content with indexing terms.", Name: "benchmarks/reference/" +
				fmt.Sprintf("bench-doc-%d", i),
		}
		if _, err := store.Create(ctx, doc); err != nil {
			b.Fatalf("Index failed at iteration %d: %v", i, err)
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Performance Benchmark
// [Purpose]: Benchmarks OKF keyword retrieval over a pre-populated collection.
// [Execution Surface]: Authoritative local concept files.
// [Assertions]: Measures query latency and memory allocations during search execution.
// -----------------------------------------------------------------------------
func BenchmarkOKFSearch(b *testing.B) {
	tempDir := b.TempDir()
	store, err := localfs.NewStore(tempDir)
	if err != nil {
		b.Fatalf("Failed creating indexer: %v", err)
	}

	ctx := context.Background()

	// Seed 500 documents
	for i := 0; i < 500; i++ {
		doc := gyrus.Concept{
			ID:          fmt.Sprintf("seed-doc-%d", i),
			Title:       fmt.Sprintf("Seed Document %d", i),
			Category:    gyrus.CategoryArchitecture,
			Type:        gyrus.TypeADR,
			Tenant:      "platform",
			Version:     1,
			Status:      "accepted",
			LastUpdated: time.Now(),
			Content:     "PostgreSQL and SQLite storage engine with FTS5 lexical ranking.", Name: "platform/reference/" +
				fmt.Sprintf("seed-doc-%d", i),
		}
		if _, err := store.Create(ctx, doc); err != nil {
			b.Fatal(err)
		}
	}

	searcher, err := searchokf.NewSearchProvider(store)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := searcher.Search(ctx, gyrus.SearchQuery{Query: "PostgreSQL"})
		if err != nil {
			b.Fatalf("Search failed: %v", err)
		}
		if len(results) == 0 {
			b.Fatalf("Expected search results")
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Performance Benchmark
// [Purpose]: Benchmarks multi-hop graph traversal latency across directed edge relationships.
// [Execution Surface]: Persisted OKF typed links.
// [Assertions]: Measures traversal latency across chained graph nodes.
// -----------------------------------------------------------------------------
func BenchmarkGraphTraversal(b *testing.B) {
	tempDir := b.TempDir()
	store, err := localfs.NewStore(tempDir)
	if err != nil {
		b.Fatalf("Failed creating indexer: %v", err)
	}
	graphStore, err := graphokf.NewGraphStore(store)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	// Seed a chain of 50 graph edges
	var edges []gyrus.ConceptEdge
	for i := 0; i < 50; i++ {
		if _, err := store.Create(ctx, okfConcept(fmt.Sprintf("node-%d", i))); err != nil {
			b.Fatal(err)
		}
		edges = append(edges, gyrus.ConceptEdge{
			FromConceptID:    fmt.Sprintf("root/reference/node-%d", i),
			ToConceptID:      fmt.Sprintf("root/reference/node-%d", i+1),
			RelationshipType: gyrus.RelDependsOn,
		})
	}
	if err := graphStore.UpsertEdges(ctx, edges); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := graphStore.Traverse(ctx, gyrus.GraphQuery{
			StartID:  "root/reference/node-0",
			MaxDepth: 10,
		})
		if err != nil {
			b.Fatalf("Traverse failed: %v", err)
		}
	}
}
