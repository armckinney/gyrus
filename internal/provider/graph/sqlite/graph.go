package sqlite

import (
	"context"
	"database/sql"

	"github.com/armckinney/gyrus/pkg/gyrus"
	_ "modernc.org/sqlite"
)

// GraphStore implements gyrus.GraphStore over SQLite.
type GraphStore struct {
	db *sql.DB
}

// NewGraphStore creates a new GraphStore wrapping an open SQLite database connection.
func NewGraphStore(db *sql.DB) *GraphStore {
	return &GraphStore{db: db}
}

// UpsertEdges inserts or updates directed relationship edges in SQLite.
func (g *GraphStore) UpsertEdges(ctx context.Context, edges []gyrus.DocumentEdge) error {
	if len(edges) == 0 {
		return nil
	}

	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	INSERT INTO document_edges (from_document_id, to_document_id, relationship_type, created_by, created_at)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(from_document_id, to_document_id, relationship_type) DO UPDATE SET
		created_by=excluded.created_by,
		created_at=excluded.created_at;
	`

	for _, e := range edges {
		_, err := tx.ExecContext(ctx, query, e.FromDocumentID, e.ToDocumentID, string(e.RelationshipType), e.CreatedBy, e.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeleteEdges removes a directed relationship edge between two documents.
func (g *GraphStore) DeleteEdges(ctx context.Context, fromID string, toID string, relType gyrus.RelationshipType) error {
	query := "DELETE FROM document_edges WHERE from_document_id = ? AND to_document_id = ? AND relationship_type = ?"
	_, err := g.db.ExecContext(ctx, query, fromID, toID, string(relType))
	return err
}

// Neighbors retrieves neighboring edges for a document matching the edge filter.
func (g *GraphStore) Neighbors(ctx context.Context, id string, filter gyrus.EdgeFilter) ([]gyrus.DocumentEdge, error) {
	var query string
	var args []any

	relFilter := ""
	if filter.RelationshipType != "" {
		relFilter = " AND relationship_type = ?"
	}

	switch filter.Direction {
	case "outgoing":
		query = "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE from_document_id = ?" + relFilter
		args = append(args, id)
		if filter.RelationshipType != "" {
			args = append(args, string(filter.RelationshipType))
		}
	case "incoming":
		query = "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE to_document_id = ?" + relFilter
		args = append(args, id)
		if filter.RelationshipType != "" {
			args = append(args, string(filter.RelationshipType))
		}
	default: // "both"
		if filter.RelationshipType != "" {
			query = "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE (from_document_id = ? OR to_document_id = ?) AND relationship_type = ?"
			args = append(args, id, id, string(filter.RelationshipType))
		} else {
			query = "SELECT from_document_id, to_document_id, relationship_type, created_by, created_at FROM document_edges WHERE from_document_id = ? OR to_document_id = ?"
			args = append(args, id, id)
		}
	}

	rows, err := g.db.QueryContext(ctx, query, args...)
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

	return edges, nil
}

// Traverse executes a BFS graph traversal from a start document up to MaxDepth.
func (g *GraphStore) Traverse(ctx context.Context, query gyrus.GraphQuery) ([]gyrus.GraphPath, error) {
	if query.MaxDepth <= 0 {
		query.MaxDepth = 1
	}

	visitedNodes := make(map[string]bool)
	visitedEdges := make(map[string]bool)
	var allEdges []gyrus.DocumentEdge

	visitedNodes[query.StartID] = true
	currentLevel := []string{query.StartID}

	for depth := 0; depth < query.MaxDepth && len(currentLevel) > 0; depth++ {
		var nextLevel []string

		for _, nodeID := range currentLevel {
			edges, err := g.Neighbors(ctx, nodeID, gyrus.EdgeFilter{Direction: "outgoing"})
			if err != nil {
				return nil, err
			}

			for _, edge := range edges {
				edgeKey := edge.FromDocumentID + "->" + edge.ToDocumentID + ":" + string(edge.RelationshipType)
				if !visitedEdges[edgeKey] {
					visitedEdges[edgeKey] = true
					allEdges = append(allEdges, edge)
				}

				if !visitedNodes[edge.ToDocumentID] {
					visitedNodes[edge.ToDocumentID] = true
					nextLevel = append(nextLevel, edge.ToDocumentID)
				}
			}
		}
		currentLevel = nextLevel
	}

	var nodes []string
	for n := range visitedNodes {
		nodes = append(nodes, n)
	}

	return []gyrus.GraphPath{{
		Nodes: nodes,
		Edges: allEdges,
	}}, nil
}
