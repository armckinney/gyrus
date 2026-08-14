package postgres

import (
	"context"

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

// GraphStore implements gyrus.GraphStore over PostgreSQL.
type GraphStore struct {
	pool PgxPool
}

// NewGraphStore creates a new GraphStore wrapping an open PostgreSQL connection pool.
func NewGraphStore(pool PgxPool) *GraphStore {
	return &GraphStore{pool: pool}
}

// UpsertEdges inserts or updates directed relationship edges in PostgreSQL.
func (g *GraphStore) UpsertEdges(ctx context.Context, edges []gyrus.DocumentEdge) error {
	if len(edges) == 0 {
		return nil
	}

	tx, err := g.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO document_edges (from_document_id, to_document_id, relationship_type, created_by, created_at)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT(from_document_id, to_document_id, relationship_type) DO UPDATE SET
		created_by=EXCLUDED.created_by,
		created_at=EXCLUDED.created_at;
	`

	for _, e := range edges {
		_, err := tx.Exec(ctx, query, e.FromDocumentID, e.ToDocumentID, string(e.RelationshipType), e.CreatedBy, e.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// DeleteEdges removes a directed relationship edge between two documents.
func (g *GraphStore) DeleteEdges(ctx context.Context, fromID string, toID string, relType gyrus.RelationshipType) error {
	query := "DELETE FROM document_edges WHERE from_document_id = $1 AND to_document_id = $2 AND relationship_type = $3"
	_, err := g.pool.Exec(ctx, query, fromID, toID, string(relType))
	return err
}

// Neighbors retrieves neighboring edges for a document matching the edge filter.
func (g *GraphStore) Neighbors(ctx context.Context, id string, filter gyrus.EdgeFilter) ([]gyrus.DocumentEdge, error) {
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
		args = append(args, string(filter.RelationshipType))
	}

	if dir == "outgoing" {
		rows, err = g.pool.Query(ctx, "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE from_document_id = $1"+whereClause, args...)
	} else if dir == "incoming" {
		rows, err = g.pool.Query(ctx, "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE to_document_id = $1"+whereClause, args...)
	} else {
		if filter.RelationshipType != "" {
			rows, err = g.pool.Query(ctx, "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE (from_document_id = $1 OR to_document_id = $1) AND relationship_type = $2", id, string(filter.RelationshipType))
		} else {
			rows, err = g.pool.Query(ctx, "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE from_document_id = $1 OR to_document_id = $1", id)
		}
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []gyrus.DocumentEdge
	for rows.Next() {
		var e gyrus.DocumentEdge
		var relStr string

		if err := rows.Scan(&e.FromDocumentID, &e.ToDocumentID, &relStr, &e.CreatedBy, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.RelationshipType = gyrus.RelationshipType(relStr)
		edges = append(edges, e)
	}

	return edges, rows.Err()
}

// Traverse executes a recursive CTE graph traversal in PostgreSQL.
func (g *GraphStore) Traverse(ctx context.Context, query gyrus.GraphQuery) ([]gyrus.GraphPath, error) {
	if query.MaxDepth <= 0 {
		query.MaxDepth = 1
	}

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
	rows, err := g.pool.Query(ctx, sql, query.StartID, query.MaxDepth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []gyrus.DocumentEdge
	for rows.Next() {
		var e gyrus.DocumentEdge
		var relStr string
		if err := rows.Scan(&e.FromDocumentID, &e.ToDocumentID, &relStr); err != nil {
			return nil, err
		}
		e.RelationshipType = gyrus.RelationshipType(relStr)
		edges = append(edges, e)
	}

	if len(edges) == 0 {
		return []gyrus.GraphPath{{
			Nodes: []string{query.StartID},
			Edges: nil,
		}}, nil
	}

	nodesMap := make(map[string]bool)
	nodesMap[query.StartID] = true
	for _, e := range edges {
		nodesMap[e.FromDocumentID] = true
		nodesMap[e.ToDocumentID] = true
	}

	var nodes []string
	for n := range nodesMap {
		nodes = append(nodes, n)
	}

	return []gyrus.GraphPath{{
		Nodes: nodes,
		Edges: edges,
	}}, nil
}
