package postgres

import (
	"context"
	"fmt"
	"github.com/armckinney/gyrus/internal/domain/okf"
	searchpostgres "github.com/armckinney/gyrus/internal/provider/search/postgres_fts"

	"github.com/armckinney/gyrus/internal/persistence/db"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PgxPool defines the interface for pgxpool to allow mocking.
type PgxPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

// Indexer implements gyrus.IndexStore over PostgreSQL.
type Indexer struct {
	pool       PgxPool
	connString string
}

// NewIndexer initializes PostgreSQL indexer and schema migrations.
func NewIndexer(ctx context.Context, connString string) (*Indexer, error) {
	pool, err := db.OpenPostgres(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres pool: %w", err)
	}

	idx := &Indexer{
		pool:       pool,
		connString: connString,
	}

	if err := idx.initSchema(ctx); err != nil {
		return nil, err
	}

	return idx, nil
}

func (idx *Indexer) initSchema(ctx context.Context) error {
	ddl := `
	CREATE TABLE IF NOT EXISTS concepts (
		id TEXT PRIMARY KEY,
		frontmatter JSONB NOT NULL,
		content TEXT NOT NULL,
		search_vector tsvector GENERATED ALWAYS AS (
			setweight(to_tsvector('english', coalesce(frontmatter->>'title', '')), 'A') ||
			setweight(to_tsvector('english', coalesce(content, '')), 'B') ||
			setweight(to_tsvector('english', coalesce((frontmatter->'tags')::text, '')), 'C')
		) STORED
	);

	CREATE INDEX IF NOT EXISTS idx_concepts_search ON concepts USING GIN (search_vector);
	CREATE INDEX IF NOT EXISTS idx_concepts_frontmatter ON concepts USING GIN (frontmatter);
	`
	_, err := idx.pool.Exec(ctx, ddl)
	return err
}

// Pool returns the underlying PgxPool connection.
func (idx *Indexer) Pool() PgxPool {
	return idx.pool
}

// Close closes the indexer connection pool.
func (idx *Indexer) Close() {
	db.ClosePostgres(idx.connString)
}

// Index inserts or updates a concept envelope in PostgreSQL.
func (idx *Indexer) Index(ctx context.Context, doc gyrus.Concept) error {
	if err := okf.SetName(&doc, ""); err != nil {
		return err
	}

	fm, err := okf.SerializeMetadata(doc)
	if err != nil {
		return err
	}

	_, err = idx.pool.Exec(ctx, `
		INSERT INTO concepts (id, frontmatter, content)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET frontmatter = EXCLUDED.frontmatter, content = EXCLUDED.content
	`, doc.Name, fm, doc.Content)
	return err
}

// Remove deletes a concept envelope from the index.
func (idx *Indexer) Remove(ctx context.Context, id string) error {
	if err := okf.ValidateName(id); err != nil {
		return err
	}

	_, err := idx.pool.Exec(ctx, "DELETE FROM concepts WHERE id = $1", id)
	return err
}

// Search executes an index search query.
func (idx *Indexer) Search(ctx context.Context, query gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	return searchpostgres.NewSearchProviderWithPool(idx.pool).Search(ctx, query)
}
