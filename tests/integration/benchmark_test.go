package integration_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	graphsqlite "github.com/armckinney/gyrus/internal/provider/graph/sqlite"
	indexsqlite "github.com/armckinney/gyrus/internal/provider/index/sqlite"
	searchfts5 "github.com/armckinney/gyrus/internal/provider/search/fts5"
	"github.com/armckinney/gyrus/pkg/gyrus"
)

// -----------------------------------------------------------------------------
// [Test Level]: Performance Benchmark
// [Purpose]: Benchmarks high-throughput document indexing into SQLite FTS5 index.
// [Execution Surface]: Real In-Memory / Local SQLite Database
// [Assertions]: Measures ns/op and allocs/op for document indexing.
// -----------------------------------------------------------------------------
func BenchmarkDocumentIndexing(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench_index.db")

	indexer, err := indexsqlite.NewIndexer(dbPath)
	if err != nil {
		b.Fatalf("Failed creating indexer: %v", err)
	}
	defer indexer.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := gyrus.Document{
			ID:          fmt.Sprintf("bench-doc-%d", i),
			Title:       fmt.Sprintf("Benchmark Document %d", i),
			Category:    gyrus.CategoryTechnical,
			Type:        gyrus.TypeSpecification,
			OwnerGroup:  "benchmarks",
			Version:     1,
			Status:      "active",
			LastUpdated: time.Now(),
			Content:     "High performance benchmark document body content with indexing terms.",
		}
		if err := indexer.Index(ctx, doc); err != nil {
			b.Fatalf("Index failed at iteration %d: %v", i, err)
		}
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Performance Benchmark
// [Purpose]: Benchmarks lexical FTS5 search query latency against pre-populated corpus.
// [Execution Surface]: Real In-Memory / Local SQLite Database
// [Assertions]: Measures query latency and memory allocations during search execution.
// -----------------------------------------------------------------------------
func BenchmarkFTS5Search(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench_search.db")

	indexer, err := indexsqlite.NewIndexer(dbPath)
	if err != nil {
		b.Fatalf("Failed creating indexer: %v", err)
	}
	defer indexer.Close()

	ctx := context.Background()

	// Seed 500 documents
	for i := 0; i < 500; i++ {
		doc := gyrus.Document{
			ID:          fmt.Sprintf("seed-doc-%d", i),
			Title:       fmt.Sprintf("Seed Document %d", i),
			Category:    gyrus.CategoryArchitecture,
			Type:        gyrus.TypeADR,
			OwnerGroup:  "platform",
			Version:     1,
			Status:      "accepted",
			LastUpdated: time.Now(),
			Content:     "PostgreSQL and SQLite storage engine with FTS5 lexical ranking.",
		}
		_ = indexer.Index(ctx, doc)
	}

	searcher := searchfts5.NewSearchProvider(indexer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, err := searcher.Search(ctx, "PostgreSQL", gyrus.SearchFilter{})
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
// [Execution Surface]: Real In-Memory / Local SQLite Database
// [Assertions]: Measures traversal latency across chained graph nodes.
// -----------------------------------------------------------------------------
func BenchmarkGraphTraversal(b *testing.B) {
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench_graph.db")

	indexer, err := indexsqlite.NewIndexer(dbPath)
	if err != nil {
		b.Fatalf("Failed creating indexer: %v", err)
	}
	defer indexer.Close()

	graphStore := graphsqlite.NewGraphStore(indexer.DB())
	ctx := context.Background()

	// Seed a chain of 50 graph edges
	var edges []gyrus.DocumentEdge
	for i := 0; i < 50; i++ {
		edges = append(edges, gyrus.DocumentEdge{
			FromDocumentID:   fmt.Sprintf("node-%d", i),
			ToDocumentID:     fmt.Sprintf("node-%d", i+1),
			RelationshipType: gyrus.RelDependsOn,
			CreatedAt:        time.Now(),
		})
	}
	_ = graphStore.UpsertEdges(ctx, edges)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := graphStore.Traverse(ctx, gyrus.GraphQuery{
			StartID:  "node-0",
			MaxDepth: 10,
		})
		if err != nil {
			b.Fatalf("Traverse failed: %v", err)
		}
	}
}
