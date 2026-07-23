package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxPool defines the interface for pgxpool to allow mocking.
type PgxPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

// Store implements DocumentStore, IndexStore, GraphStore, SearchProvider
type Store struct {
	pool PgxPool
}

// NewStore initializes the PostgreSQL store and runs DDL migrations.
func NewStore(ctx context.Context, connString string) (*Store, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	s := &Store{pool: pool}
	if err := s.initSchema(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) initSchema(ctx context.Context) error {
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

	CREATE TABLE IF NOT EXISTS document_edges (
		from_document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
		to_document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
		relationship_type TEXT NOT NULL,
		created_by TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (from_document_id, to_document_id, relationship_type)
	);
	CREATE INDEX IF NOT EXISTS idx_edges_to ON document_edges(to_document_id);

	CREATE TABLE IF NOT EXISTS documents_history (
		id TEXT NOT NULL,
		version INT NOT NULL,
		frontmatter JSONB NOT NULL,
		content TEXT NOT NULL,
		modified_by TEXT,
		modified_at TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (id, version)
	);
	`
	_, err := s.pool.Exec(ctx, ddl)
	return err
}

func (s *Store) Close() {
	s.pool.Close()
}

// --- DocumentStore ---

func (s *Store) Create(ctx context.Context, doc gyrus.Document) (gyrus.DocumentRef, error) {
	if doc.Version == 0 {
		doc.Version = 1
	}
	if doc.LastUpdated.IsZero() {
		doc.LastUpdated = time.Now()
	}

	fm, err := json.Marshal(doc)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		"INSERT INTO documents (id, frontmatter, content) VALUES ($1, $2, $3)",
		doc.ID, fm, doc.Content,
	)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO documents_history (id, version, frontmatter, content, modified_by, modified_at) VALUES ($1, $2, $3, $4, $5, $6)",
		doc.ID, doc.Version, fm, doc.Content, doc.LastModifiedBy, doc.LastUpdated,
	)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return gyrus.DocumentRef{}, err
	}

	return gyrus.DocumentRef{
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

func (s *Store) Get(ctx context.Context, id string) (gyrus.Document, error) {
	var fm []byte
	var content string

	err := s.pool.QueryRow(ctx, "SELECT frontmatter, content FROM documents WHERE id = $1", id).Scan(&fm, &content)
	if err != nil {
		if err == pgx.ErrNoRows {
			return gyrus.Document{}, fmt.Errorf("document not found: %s", id)
		}
		return gyrus.Document{}, err
	}

	var doc gyrus.Document
	if err := json.Unmarshal(fm, &doc); err != nil {
		return gyrus.Document{}, err
	}
	doc.Content = content

	return doc, nil
}

func (s *Store) Update(ctx context.Context, id string, patch gyrus.DocumentPatch, expectedVersion int) (gyrus.DocumentRef, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}
	defer tx.Rollback(ctx)

	var fmBytes []byte
	var content string

	err = tx.QueryRow(ctx, "SELECT frontmatter, content FROM documents WHERE id = $1 FOR UPDATE", id).Scan(&fmBytes, &content)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	var doc gyrus.Document
	if err := json.Unmarshal(fmBytes, &doc); err != nil {
		return gyrus.DocumentRef{}, err
	}

	if expectedVersion != 0 && doc.Version != expectedVersion {
		return gyrus.DocumentRef{}, fmt.Errorf("version mismatch: expected %d, got %d", expectedVersion, doc.Version)
	}

	if patch.Title != nil {
		doc.Title = *patch.Title
	}
	if patch.Status != nil {
		doc.Status = *patch.Status
	}
	if patch.Tags != nil {
		doc.Tags = *patch.Tags
	}
	if patch.Dependencies != nil {
		doc.Dependencies = *patch.Dependencies
	}
	if patch.Content != nil {
		doc.Content = *patch.Content
	}
	
	doc.Version++
	doc.LastUpdated = time.Now()

	newFmBytes, err := json.Marshal(doc)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	_, err = tx.Exec(ctx, "UPDATE documents SET frontmatter = $1, content = $2 WHERE id = $3", newFmBytes, doc.Content, id)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO documents_history (id, version, frontmatter, content, modified_by, modified_at) VALUES ($1, $2, $3, $4, $5, $6)",
		doc.ID, doc.Version, newFmBytes, doc.Content, doc.LastModifiedBy, doc.LastUpdated,
	)
	if err != nil {
		return gyrus.DocumentRef{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return gyrus.DocumentRef{}, err
	}

	return gyrus.DocumentRef{
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := s.pool.Exec(ctx, "DELETE FROM documents WHERE id = $1", id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("document not found: %s", id)
	}
	return nil
}

func (s *Store) Archive(ctx context.Context, id string) error {
	patch := gyrus.DocumentPatch{
		Status: func() *string { s := "archived"; return &s }(),
		Reason: "archived",
	}
	_, err := s.Update(ctx, id, patch, 0)
	return err
}

// --- IndexStore ---

func (s *Store) Index(ctx context.Context, doc gyrus.Document) error {
	// Our postgres implementation is the primary store, so index is essentially a no-op or upsert.
	// We'll perform an upsert just in case it's used to manually sync the index.
	fm, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	
	_, err = s.pool.Exec(ctx, `
		INSERT INTO documents (id, frontmatter, content)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET frontmatter = EXCLUDED.frontmatter, content = EXCLUDED.content
	`, doc.ID, fm, doc.Content)
	return err
}

func (s *Store) Remove(ctx context.Context, id string) error {
	return s.Delete(ctx, id)
}

func (s *Store) Sync(ctx context.Context, storageRoot string) (gyrus.SyncReport, error) {
	// Placeholder for IndexStore sync
	return gyrus.SyncReport{}, nil
}

func (s *Store) Search(ctx context.Context, query gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	// Implement generic search without FTS for IndexStore if needed, or route to FTS.
	// Let's use the SearchProvider FTS.
	return s.SearchFTS(ctx, query.Query, query.Filter)
}

// --- GraphStore ---

func (s *Store) UpsertEdges(ctx context.Context, edges []gyrus.DocumentEdge) error {
	if len(edges) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, e := range edges {
		_, err := tx.Exec(ctx, `
			INSERT INTO document_edges (from_document_id, to_document_id, relationship_type, created_by, created_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (from_document_id, to_document_id, relationship_type) 
			DO UPDATE SET created_by = EXCLUDED.created_by, created_at = EXCLUDED.created_at
		`, e.FromDocumentID, e.ToDocumentID, e.RelationshipType, e.CreatedBy, e.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Store) DeleteEdges(ctx context.Context, fromID string, toID string, relType gyrus.RelationshipType) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM document_edges 
		WHERE from_document_id = $1 AND to_document_id = $2 AND relationship_type = $3
	`, fromID, toID, relType)
	return err
}

func (s *Store) Neighbors(ctx context.Context, id string, filter gyrus.EdgeFilter) ([]gyrus.DocumentEdge, error) {
	var rows pgx.Rows
	var err error
	
	dir := filter.Direction
	if dir == "" {
		dir = "both"
	}
	
	args := []any{id}
	whereClause := ""
	if filter.RelationshipType != "" {
		whereClause = " AND relationship_type = $2"
		args = append(args, filter.RelationshipType)
	}

	if dir == "outgoing" {
		rows, err = s.pool.Query(ctx, "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE from_document_id = $1"+whereClause, args...)
	} else if dir == "incoming" {
		rows, err = s.pool.Query(ctx, "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE to_document_id = $1"+whereClause, args...)
	} else {
		// both
		if filter.RelationshipType != "" {
			rows, err = s.pool.Query(ctx, "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE (from_document_id = $1 OR to_document_id = $1) AND relationship_type = $2", id, filter.RelationshipType)
		} else {
			rows, err = s.pool.Query(ctx, "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE from_document_id = $1 OR to_document_id = $1", id)
		}
	}
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []gyrus.DocumentEdge
	for rows.Next() {
		var e gyrus.DocumentEdge
		if err := rows.Scan(&e.FromDocumentID, &e.ToDocumentID, &e.RelationshipType, &e.CreatedBy, &e.CreatedAt); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}

	return edges, rows.Err()
}

func (s *Store) Traverse(ctx context.Context, query gyrus.GraphQuery) ([]gyrus.GraphPath, error) {
	// A simple BFS for multi-hop traversal
	// For simplicity, returning just 1 path if maxDepth is 1.
	var paths []gyrus.GraphPath

	if query.MaxDepth <= 0 {
		return paths, nil
	}

	// This is a naive implementation. In a real scenario, use recursive CTEs.
	// Let's implement a recursive CTE in postgres for traversal.
	sql := `
	WITH RECURSIVE traverse AS (
		SELECT from_document_id, to_document_id, relationship_type, 1 AS depth
		FROM document_edges
		WHERE from_document_id = $1
		
		UNION
		
		SELECT e.from_document_id, e.to_document_id, e.relationship_type, t.depth + 1
		FROM document_edges e
		INNER JOIN traverse t ON e.from_document_id = t.to_document_id
		WHERE t.depth < $2
	)
	SELECT from_document_id, to_document_id, relationship_type FROM traverse;
	`
	rows, err := s.pool.Query(ctx, sql, query.StartID, query.MaxDepth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Collecting edges
	var edges []gyrus.DocumentEdge
	for rows.Next() {
		var e gyrus.DocumentEdge
		if err := rows.Scan(&e.FromDocumentID, &e.ToDocumentID, &e.RelationshipType); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}

	if len(edges) > 0 {
		nodes := make(map[string]bool)
		nodes[query.StartID] = true
		for _, e := range edges {
			nodes[e.FromDocumentID] = true
			nodes[e.ToDocumentID] = true
		}
		
		var nodeList []string
		for k := range nodes {
			nodeList = append(nodeList, k)
		}

		paths = append(paths, gyrus.GraphPath{
			Nodes: nodeList,
			Edges: edges,
		})
	} else {
		paths = append(paths, gyrus.GraphPath{
			Nodes: []string{query.StartID},
			Edges: nil,
		})
	}

	return paths, nil
}

// --- SearchProvider ---

// SearchFTS is the underlying implementation for SearchProvider.Search
func (s *Store) SearchFTS(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	whereClauses := []string{"1=1"}
	args := []any{}
	argCount := 1

	if filter.Category != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("frontmatter->>'category' = $%d", argCount))
		args = append(args, string(filter.Category))
		argCount++
	}
	if filter.Type != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("frontmatter->>'type' = $%d", argCount))
		args = append(args, string(filter.Type))
		argCount++
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("frontmatter->>'status' = $%d", argCount))
		args = append(args, filter.Status)
		argCount++
	}
	if filter.OwnerGroup != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("frontmatter->>'owner_group' = $%d", argCount))
		args = append(args, filter.OwnerGroup)
		argCount++
	}
	if filter.Tag != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("frontmatter->'tags' @> $%d", argCount))
		args = append(args, fmt.Sprintf(`["%s"]`, filter.Tag))
		argCount++
	}

	ftsClause := ""
	scoreSelect := "0::float8 AS score"
	if query != "" {
		queryTsQuery := strings.ReplaceAll(query, " ", " | ") // basic text search conversion
		ftsClause = fmt.Sprintf(" AND search_vector @@ to_tsquery('english', $%d)", argCount)
		scoreSelect = fmt.Sprintf("ts_rank_cd(search_vector, to_tsquery('english', $%d)) AS score", argCount)
		args = append(args, queryTsQuery)
		argCount++
	}

	whereSql := strings.Join(whereClauses, " AND ") + ftsClause

	sql := fmt.Sprintf(`
		SELECT frontmatter, content, %s
		FROM documents
		WHERE %s
		ORDER BY score DESC
		LIMIT 50
	`, scoreSelect, whereSql)

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []gyrus.SearchResult
	for rows.Next() {
		var fm []byte
		var content string
		var score float64

		if err := rows.Scan(&fm, &content, &score); err != nil {
			return nil, err
		}

		var doc gyrus.Document
		if err := json.Unmarshal(fm, &doc); err != nil {
			return nil, err
		}
		doc.Content = content

		results = append(results, gyrus.SearchResult{
			Document:    doc,
			Score:       score,
			MatchReason: "FTS Match",
		})
	}

	return results, nil
}

func (s *Store) SearchProviderSearch(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	return s.SearchFTS(ctx, query, filter)
}

// Wrapper for the exact interface method
func (s *Store) SearchDocuments(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	return s.SearchFTS(ctx, query, filter)
}
