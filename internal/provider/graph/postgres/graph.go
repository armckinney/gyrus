package postgres

import (
	"context"
	"github.com/armckinney/gyrus/internal/domain/okf"

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
func (g *GraphStore) UpsertEdges(ctx context.Context, edges []gyrus.ConceptEdge) error {
	for _, edge := range edges {
		if err := okf.ValidateEdge(edge); err != nil {
			return err
		}
	}

	if len(edges) == 0 {
		return nil
	}

	tx, err := g.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO concept_edges (from_concept_id, to_concept_id, relationship_type)
	VALUES ($1, $2, $3)
	ON CONFLICT(from_concept_id, to_concept_id, relationship_type) DO NOTHING;
	`

	for _, e := range edges {
		if err := okf.ValidateEdge(e); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, query, e.FromConceptID, e.ToConceptID, string(e.RelationshipType))
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// DeleteEdges removes a directed relationship edge between two concepts.
func (g *GraphStore) DeleteEdges(ctx context.Context, fromID string, toID string, relType gyrus.RelationshipType) error {
	if err := okf.ValidateEdge(gyrus.ConceptEdge{FromConceptID: fromID, ToConceptID: toID, RelationshipType: relType}); err != nil {
		return err
	}

	query := "DELETE FROM concept_edges WHERE from_concept_id = $1 AND to_concept_id = $2 AND relationship_type = $3"
	_, err := g.pool.Exec(ctx, query, fromID, toID, string(relType))
	return err
}

// Neighbors retrieves neighboring edges for a concept matching the edge filter.
func (g *GraphStore) Neighbors(ctx context.Context, id string, filter gyrus.EdgeFilter) ([]gyrus.ConceptEdge, error) {
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
		args = append(args, string(filter.RelationshipType))
	}

	if dir == "outgoing" {
		rows, err = g.pool.Query(ctx, "SELECT from_concept_id, to_concept_id, relationship_type FROM concept_edges WHERE from_concept_id = $1"+whereClause, args...)
	} else if dir == "incoming" {
		rows, err = g.pool.Query(ctx, "SELECT from_concept_id, to_concept_id, relationship_type FROM concept_edges WHERE to_concept_id = $1"+whereClause, args...)
	} else {
		if filter.RelationshipType != "" {
			rows, err = g.pool.Query(ctx, "SELECT from_concept_id, to_concept_id, relationship_type FROM concept_edges WHERE (from_concept_id = $1 OR to_concept_id = $1) AND relationship_type = $2", id, string(filter.RelationshipType))
		} else {
			rows, err = g.pool.Query(ctx, "SELECT from_concept_id, to_concept_id, relationship_type FROM concept_edges WHERE from_concept_id = $1 OR to_concept_id = $1", id)
		}
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []gyrus.ConceptEdge
	for rows.Next() {
		var e gyrus.ConceptEdge
		var relStr string

		if err := rows.Scan(&e.FromConceptID, &e.ToConceptID, &relStr); err != nil {
			return nil, err
		}
		e.RelationshipType = gyrus.RelationshipType(relStr)
		edges = append(edges, e)
	}

	return edges, rows.Err()
}

// Traverse executes a recursive CTE graph traversal in PostgreSQL.
func (g *GraphStore) Traverse(ctx context.Context, query gyrus.GraphQuery) ([]gyrus.GraphPath, error) {
	if err := okf.ValidateName(query.StartID); err != nil {
		return nil, err
	}

	if query.MaxDepth <= 0 {
		query.MaxDepth = 1
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
	rows, err := g.pool.Query(ctx, sql, query.StartID, query.MaxDepth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []gyrus.ConceptEdge
	for rows.Next() {
		var e gyrus.ConceptEdge
		var relStr string
		if err := rows.Scan(&e.FromConceptID, &e.ToConceptID, &relStr); err != nil {
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
		nodesMap[e.FromConceptID] = true
		nodesMap[e.ToConceptID] = true
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
