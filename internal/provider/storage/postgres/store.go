package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/armckinney/gyrus/internal/domain/okf"
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

// Store implements ConceptStore, IndexStore, GraphStore, SearchProvider
type Store struct {
	pool       PgxPool
	connString string
}

// NewStore initializes the PostgreSQL store and runs DDL migrations.
func NewStore(ctx context.Context, connString string) (*Store, error) {
	pool, err := db.OpenPostgres(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres pool: %w", err)
	}

	s := &Store{pool: pool, connString: connString}
	if err := s.initSchema(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) initSchema(ctx context.Context) error {
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

	CREATE TABLE IF NOT EXISTS concept_edges (
		from_concept_id TEXT NOT NULL REFERENCES concepts(id) ON DELETE CASCADE,
		to_concept_id TEXT NOT NULL REFERENCES concepts(id) ON DELETE CASCADE,
		relationship_type TEXT NOT NULL,
		PRIMARY KEY (from_concept_id, to_concept_id, relationship_type)
	);
	CREATE INDEX IF NOT EXISTS idx_edges_to ON concept_edges(to_concept_id);

	CREATE TABLE IF NOT EXISTS concepts_history (
		id TEXT NOT NULL,
		version INT NOT NULL,
		frontmatter JSONB NOT NULL,
		content TEXT NOT NULL,
		modified_by TEXT,
		modified_at TIMESTAMPTZ DEFAULT NOW(),
		PRIMARY KEY (id, version)
	);

	CREATE TABLE IF NOT EXISTS schemas (
		id TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`
	_, err := s.pool.Exec(ctx, ddl)
	return err
}

func (s *Store) Close() {
	s.pool.Close()
}

// --- ConceptStore ---

func (s *Store) Create(ctx context.Context, doc gyrus.Concept) (gyrus.ConceptRef, error) {
	if err := okf.SetName(&doc, ""); err != nil {
		return gyrus.ConceptRef{}, err
	}

	if doc.Version == 0 {
		doc.Version = 1
	}
	if doc.LastUpdated.IsZero() {
		doc.LastUpdated = time.Now()
	}

	fm, err := okf.SerializeMetadata(doc)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		"INSERT INTO concepts (id, frontmatter, content) VALUES ($1, $2, $3)",
		doc.Name, fm, doc.Content,
	)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO concepts_history (id, version, frontmatter, content, modified_by, modified_at) VALUES ($1, $2, $3, $4, $5, $6)",
		doc.Name, doc.Version, fm, doc.Content, doc.LastModifiedBy, doc.LastUpdated,
	)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return gyrus.ConceptRef{}, err
	}

	return gyrus.ConceptRef{
		Name:        doc.Name,
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

func (s *Store) Get(ctx context.Context, id string) (gyrus.Concept, error) {
	if err := okf.ValidateName(id); err != nil {
		return gyrus.Concept{}, err
	}

	var fm []byte
	var content string

	err := s.pool.QueryRow(ctx, "SELECT frontmatter, content FROM concepts WHERE id = $1", id).Scan(&fm, &content)
	if err != nil {
		if err == pgx.ErrNoRows {
			return gyrus.Concept{}, fmt.Errorf("concept not found: %s", id)
		}
		return gyrus.Concept{}, err
	}

	var doc gyrus.Concept
	if err := json.Unmarshal(fm, &doc); err != nil {
		return gyrus.Concept{}, err
	}
	doc.Content = content
	if err := okf.SetName(&doc, ""); err != nil {
		return gyrus.Concept{}, err
	}

	return doc, nil
}

func (s *Store) Update(ctx context.Context, id string, patch gyrus.ConceptPatch, expectedVersion int) (gyrus.ConceptRef, error) {
	if err := okf.ValidateName(id); err != nil {
		return gyrus.ConceptRef{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}
	defer tx.Rollback(ctx)

	var fmBytes []byte
	var content string

	err = tx.QueryRow(ctx, "SELECT frontmatter, content FROM concepts WHERE id = $1 FOR UPDATE", id).Scan(&fmBytes, &content)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	var doc gyrus.Concept
	if err := json.Unmarshal(fmBytes, &doc); err != nil {
		return gyrus.ConceptRef{}, err
	}
	doc.Content = content
	if err := okf.SetName(&doc, ""); err != nil {
		return gyrus.ConceptRef{}, err
	}

	if expectedVersion != 0 && doc.Version != expectedVersion {
		return gyrus.ConceptRef{}, fmt.Errorf("version mismatch: expected %d, got %d", expectedVersion, doc.Version)
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
	if patch.Links != nil {
		doc.Links = *patch.Links
	}
	if patch.Content != nil {
		doc.Content = *patch.Content
	}

	doc.Version++
	doc.LastUpdated = time.Now()

	newFmBytes, err := okf.SerializeMetadata(doc)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	_, err = tx.Exec(ctx, "UPDATE concepts SET frontmatter = $1, content = $2 WHERE id = $3", newFmBytes, doc.Content, id)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	_, err = tx.Exec(ctx,
		"INSERT INTO concepts_history (id, version, frontmatter, content, modified_by, modified_at) VALUES ($1, $2, $3, $4, $5, $6)",
		doc.Name, doc.Version, newFmBytes, doc.Content, doc.LastModifiedBy, doc.LastUpdated,
	)
	if err != nil {
		return gyrus.ConceptRef{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return gyrus.ConceptRef{}, err
	}

	return gyrus.ConceptRef{
		Name:        doc.Name,
		ID:          doc.ID,
		Version:     doc.Version,
		Status:      doc.Status,
		LastUpdated: doc.LastUpdated,
	}, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	if err := okf.ValidateName(id); err != nil {
		return err
	}

	res, err := s.pool.Exec(ctx, "DELETE FROM concepts WHERE id = $1", id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("concept not found: %s", id)
	}
	return nil
}

func (s *Store) Archive(ctx context.Context, id string) error {
	if err := okf.ValidateName(id); err != nil {
		return err
	}

	patch := gyrus.ConceptPatch{
		Status: func() *string { s := "archived"; return &s }(),
		Reason: "archived",
	}
	_, err := s.Update(ctx, id, patch, 0)
	return err
}

// --- IndexStore ---

func (s *Store) Index(ctx context.Context, doc gyrus.Concept) error {
	if err := okf.SetName(&doc, ""); err != nil {
		return err
	}

	fm, err := okf.SerializeMetadata(doc)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO concepts (id, frontmatter, content)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET frontmatter = EXCLUDED.frontmatter, content = EXCLUDED.content
	`, doc.Name, fm, doc.Content)
	return err
}

func (s *Store) Remove(ctx context.Context, id string) error {
	if err := okf.ValidateName(id); err != nil {
		return err
	}

	return s.Delete(ctx, id)
}

func (s *Store) Sync(ctx context.Context, storageRoot string) (gyrus.SyncReport, error) {
	return gyrus.SyncReport{}, nil
}

func (s *Store) Search(ctx context.Context, query gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	return s.SearchFTS(ctx, query.Query, query.Filter)
}

// --- GraphStore ---

func (s *Store) UpsertEdges(ctx context.Context, edges []gyrus.ConceptEdge) error {
	for _, edge := range edges {
		if err := okf.ValidateEdge(edge); err != nil {
			return err
		}
	}

	if len(edges) == 0 {
		return nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, e := range edges {
		if err := okf.ValidateEdge(e); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO concept_edges (from_concept_id, to_concept_id, relationship_type)
			VALUES ($1, $2, $3)
			ON CONFLICT (from_concept_id, to_concept_id, relationship_type) 
			DO NOTHING
		`, e.FromConceptID, e.ToConceptID, e.RelationshipType)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Store) DeleteEdges(ctx context.Context, fromID string, toID string, relType gyrus.RelationshipType) error {
	if err := okf.ValidateEdge(gyrus.ConceptEdge{FromConceptID: fromID, ToConceptID: toID, RelationshipType: relType}); err != nil {
		return err
	}

	_, err := s.pool.Exec(ctx, `
		DELETE FROM concept_edges 
		WHERE from_concept_id = $1 AND to_concept_id = $2 AND relationship_type = $3
	`, fromID, toID, relType)
	return err
}

func (s *Store) Neighbors(ctx context.Context, id string, filter gyrus.EdgeFilter) ([]gyrus.ConceptEdge, error) {
	if err := okf.ValidateName(id); err != nil {
		return nil, err
	}

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
		rows, err = s.pool.Query(ctx, "SELECT from_concept_id, to_concept_id, relationship_type FROM concept_edges WHERE from_concept_id = $1"+whereClause, args...)
	} else if dir == "incoming" {
		rows, err = s.pool.Query(ctx, "SELECT from_concept_id, to_concept_id, relationship_type FROM concept_edges WHERE to_concept_id = $1"+whereClause, args...)
	} else {
		// both
		if filter.RelationshipType != "" {
			rows, err = s.pool.Query(ctx, "SELECT from_concept_id, to_concept_id, relationship_type FROM concept_edges WHERE (from_concept_id = $1 OR to_concept_id = $1) AND relationship_type = $2", id, filter.RelationshipType)
		} else {
			rows, err = s.pool.Query(ctx, "SELECT from_concept_id, to_concept_id, relationship_type FROM concept_edges WHERE from_concept_id = $1 OR to_concept_id = $1", id)
		}
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []gyrus.ConceptEdge
	for rows.Next() {
		var e gyrus.ConceptEdge
		if err := rows.Scan(&e.FromConceptID, &e.ToConceptID, &e.RelationshipType); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}

	return edges, rows.Err()
}

func (s *Store) Traverse(ctx context.Context, query gyrus.GraphQuery) ([]gyrus.GraphPath, error) {
	if err := okf.ValidateName(query.StartID); err != nil {
		return nil, err
	}

	var paths []gyrus.GraphPath

	if query.MaxDepth <= 0 {
		return paths, nil
	}

	sql := `
	WITH RECURSIVE traverse AS (
		SELECT from_concept_id, to_concept_id, relationship_type, 1 AS depth
		FROM concept_edges
		WHERE from_concept_id = $1
		
		UNION
		
		SELECT e.from_concept_id, e.to_concept_id, e.relationship_type, t.depth + 1
		FROM concept_edges e
		INNER JOIN traverse t ON e.from_concept_id = t.to_concept_id
		WHERE t.depth < $2
	)
	SELECT from_concept_id, to_concept_id, relationship_type FROM traverse;
	`
	rows, err := s.pool.Query(ctx, sql, query.StartID, query.MaxDepth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []gyrus.ConceptEdge
	for rows.Next() {
		var e gyrus.ConceptEdge
		if err := rows.Scan(&e.FromConceptID, &e.ToConceptID, &e.RelationshipType); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}

	if len(edges) > 0 {
		nodes := make(map[string]bool)
		nodes[query.StartID] = true
		for _, e := range edges {
			if err := okf.ValidateEdge(e); err != nil {
				return nil, err
			}
			nodes[e.FromConceptID] = true
			nodes[e.ToConceptID] = true
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
	if filter.Tenant != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("frontmatter->>'tenant' = $%d", argCount))
		args = append(args, filter.Tenant)
		argCount++
	}
	if filter.Tag != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("frontmatter->'tags' @> $%d", argCount))
		args = append(args, fmt.Sprintf(`["%s"]`, filter.Tag))
		argCount++
	}

	ftsClause := ""
	scoreSelect := "0::float8 AS score"
	if query != "" && okf.ValidateName(query) == nil {
		ftsClause = fmt.Sprintf(" AND id = $%d", argCount)
		scoreSelect = "1::float8 AS score"
		args = append(args, query)
		argCount++
	} else if query != "" {
		queryTsQuery := strings.ReplaceAll(query, " ", " | ") // basic text search conversion
		ftsClause = fmt.Sprintf(" AND search_vector @@ to_tsquery('english', $%d)", argCount)
		scoreSelect = fmt.Sprintf("ts_rank_cd(search_vector, to_tsquery('english', $%d)) AS score", argCount)
		args = append(args, queryTsQuery)
		argCount++
	}

	whereSql := strings.Join(whereClauses, " AND ") + ftsClause

	sql := fmt.Sprintf(`
		SELECT frontmatter, content, %s
		FROM concepts
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

		var doc gyrus.Concept
		if err := json.Unmarshal(fm, &doc); err != nil {
			return nil, err
		}
		doc.Content = content
		if err := okf.SetName(&doc, ""); err != nil {
			return nil, err
		}

		results = append(results, gyrus.SearchResult{
			Concept:     doc,
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
func (s *Store) SearchConcepts(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
	return s.SearchFTS(ctx, query, filter)
}

// --- SchemaStore ---

// GetSchema retrieves a persisted schema template for a concept type from PostgreSQL.
func (s *Store) GetSchema(ctx context.Context, conceptType string) (string, error) {
	var content string
	err := s.pool.QueryRow(ctx, "SELECT content FROM schemas WHERE id = $1", conceptType).Scan(&content)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("schema '%s' not found: %w", conceptType, err)
		}
		return "", err
	}
	return content, nil
}

// SaveSchema persists an OKF schema template to the schemas table in PostgreSQL.
func (s *Store) SaveSchema(ctx context.Context, conceptType string, content string) error {
	if err := okf.ValidateSchemaContent(conceptType, content); err != nil {
		return err
	}

	query := `
	INSERT INTO schemas (id, content, updated_at)
	VALUES ($1, $2, NOW())
	ON CONFLICT (id) DO UPDATE
	SET content = EXCLUDED.content, updated_at = NOW()
	`
	_, err := s.pool.Exec(ctx, query, conceptType, content)
	return err
}

// ListSchemas enumerates all persisted schema template types in PostgreSQL.
func (s *Store) ListSchemas(ctx context.Context) ([]string, error) {
	rows, err := s.pool.Query(ctx, "SELECT id FROM schemas ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		list = append(list, id)
	}
	return list, nil
}

// DeleteSchema removes a persisted schema template from PostgreSQL.
func (s *Store) DeleteSchema(ctx context.Context, conceptType string) error {
	tag, err := s.pool.Exec(ctx, "DELETE FROM schemas WHERE id = $1", conceptType)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("schema '%s' not found", conceptType)
	}
	return nil
}
