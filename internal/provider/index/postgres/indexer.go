package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/armckinney/gyrus/internal/provider/db"
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
	CREATE TABLE IF NOT EXISTS documents (
		id TEXT PRIMARY KEY,
		frontmatter JSONB NOT NULL,
		content TEXT NOT NULL,
		search_vector tsvector GENERATED ALWAYS AS (
			setweight(to_tsvector('english', coalesce(frontmatter->>'title', '')), 'A') ||
			setweight(to_tsvector('english', coalesce(content, '')), 'B') ||
			setweight(to_tsvector('english', coalesce((frontmatter->'tags')::text, '')), 'C')
		) STORED
	);

	CREATE INDEX IF NOT EXISTS idx_documents_search ON documents USING GIN (search_vector);
	CREATE INDEX IF NOT EXISTS idx_documents_frontmatter ON documents USING GIN (frontmatter);
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

// Index inserts or updates a document envelope in PostgreSQL.
func (idx *Indexer) Index(ctx context.Context, doc gyrus.Document) error {
	fm, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	_, err = idx.pool.Exec(ctx, `
		INSERT INTO documents (id, frontmatter, content)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET frontmatter = EXCLUDED.frontmatter, content = EXCLUDED.content
	`, doc.ID, fm, doc.Content)
	return err
}

// Remove deletes a document envelope from the index.
func (idx *Indexer) Remove(ctx context.Context, id string) error {
	_, err := idx.pool.Exec(ctx, "DELETE FROM documents WHERE id = $1", id)
	return err
}

// Sync is a no-op placeholder for remote database indexers.
func (idx *Indexer) Sync(ctx context.Context, storageRoot string) (gyrus.SyncReport, error) {
	return gyrus.SyncReport{}, nil
}

// Search executes an index search query.
func (idx *Indexer) Search(ctx context.Context, query gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	rows, err := idx.pool.Query(ctx, "SELECT frontmatter, content FROM documents WHERE id = $1", query.Query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []gyrus.SearchResult
	for rows.Next() {
		var fm []byte
		var content string
		if err := rows.Scan(&fm, &content); err != nil {
			return nil, err
		}

		var doc gyrus.Document
		if err := json.Unmarshal(fm, &doc); err != nil {
			return nil, err
		}
		doc.Content = content
		results = append(results, gyrus.SearchResult{Document: doc, Score: 1.0})
	}
	return results, nil
}
