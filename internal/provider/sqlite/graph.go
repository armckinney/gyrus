package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/armckinney/gyrus/pkg/gyrus"
)

// GraphStore implements gyrus.GraphStore over SQLite's document_edges table.
func (idx *Indexer) UpsertEdges(ctx context.Context, edges []gyrus.DocumentEdge) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	INSERT INTO document_edges (from_id, to_id, relationship_type, created_by, created_at)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(from_id, to_id, relationship_type) DO UPDATE SET
		created_by=excluded.created_by,
		created_at=excluded.created_at;
	`

	for _, edge := range edges {
		createdAt := edge.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().Truncate(time.Second)
		}
		createdBy := edge.CreatedBy
		if createdBy == "" {
			createdBy = "system"
		}

		_, err := tx.ExecContext(ctx, query,
			edge.FromDocumentID, edge.ToDocumentID, string(edge.RelationshipType), createdBy, createdAt.Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("failed to upsert edge (%s -> %s): %w", edge.FromDocumentID, edge.ToDocumentID, err)
		}
	}

	return tx.Commit()
}

func (idx *Indexer) DeleteEdges(ctx context.Context, fromID string, toID string, relType gyrus.RelationshipType) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	query := `DELETE FROM document_edges WHERE from_id = ? AND to_id = ? AND relationship_type = ?`
	_, err := idx.db.ExecContext(ctx, query, fromID, toID, string(relType))
	if err != nil {
		return fmt.Errorf("failed to delete edge: %w", err)
	}
	return nil
}

func (idx *Indexer) Neighbors(ctx context.Context, id string, filter gyrus.EdgeFilter) ([]gyrus.DocumentEdge, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var query strings.Builder
	var args []interface{}

	direction := filter.Direction
	if direction == "" {
		direction = "both"
	}

	switch direction {
	case "outgoing":
		query.WriteString(`SELECT from_id, to_id, relationship_type, created_by, created_at FROM document_edges WHERE from_id = ?`)
		args = append(args, id)
	case "incoming":
		query.WriteString(`SELECT from_id, to_id, relationship_type, created_by, created_at FROM document_edges WHERE to_id = ?`)
		args = append(args, id)
	default:
		query.WriteString(`SELECT from_id, to_id, relationship_type, created_by, created_at FROM document_edges WHERE from_id = ? OR to_id = ?`)
		args = append(args, id, id)
	}

	if filter.RelationshipType != "" {
		query.WriteString(` AND relationship_type = ?`)
		args = append(args, string(filter.RelationshipType))
	}

	rows, err := idx.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query neighbors: %w", err)
	}
	defer rows.Close()

	var edges []gyrus.DocumentEdge
	for rows.Next() {
		var edge gyrus.DocumentEdge
		var relStr, createdAtStr string
		if err := rows.Scan(&edge.FromDocumentID, &edge.ToDocumentID, &relStr, &edge.CreatedBy, &createdAtStr); err != nil {
			return nil, err
		}
		edge.RelationshipType = gyrus.RelationshipType(relStr)
		edge.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		edges = append(edges, edge)
	}

	return edges, nil
}

func (idx *Indexer) Traverse(ctx context.Context, query gyrus.GraphQuery) ([]gyrus.GraphPath, error) {
	// Breadth-First Search traversal up to query.MaxDepth
	if query.MaxDepth <= 0 {
		query.MaxDepth = 3
	}

	visited := make(map[string]bool)
	queue := []string{query.StartID}
	visited[query.StartID] = true

	var allEdges []gyrus.DocumentEdge
	depth := 0

	for len(queue) > 0 && depth < query.MaxDepth {
		levelSize := len(queue)
		for i := 0; i < levelSize; i++ {
			curr := queue[0]
			queue = queue[1:]

			neighbors, err := idx.Neighbors(ctx, curr, gyrus.EdgeFilter{Direction: "outgoing"})
			if err != nil {
				return nil, err
			}

			for _, edge := range neighbors {
				allEdges = append(allEdges, edge)
				if !visited[edge.ToDocumentID] {
					visited[edge.ToDocumentID] = true
					queue = append(queue, edge.ToDocumentID)
				}
			}
		}
		depth++
	}

	pathNodes := make([]string, 0, len(visited))
	for node := range visited {
		pathNodes = append(pathNodes, node)
	}

	return []gyrus.GraphPath{
		{
			Nodes: pathNodes,
			Edges: allEdges,
		},
	}, nil
}
