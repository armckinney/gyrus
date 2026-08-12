package postgres_fts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/armckinney/gyrus/internal/provider/db"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SearchProvider implements gyrus.SearchProvider over PostgreSQL tsvector FTS.
type SearchProvider struct {
	pool       *pgxpool.Pool
	connString string
}

// NewSearchProvider initializes PostgreSQL SearchProvider.
func NewSearchProvider(ctx context.Context, connString string) (*SearchProvider, error) {
	pool, err := db.OpenPostgres(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres pool: %w", err)
	}

	return &SearchProvider{
		pool:       pool,
		connString: connString,
	}, nil
}

// Search executes PostgreSQL tsvector FTS queries.
func (p *SearchProvider) Search(ctx context.Context, query string, filter gyrus.SearchFilter) ([]gyrus.SearchResult, error) {
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
		queryTsQuery := strings.ReplaceAll(query, " ", " | ")
		ftsClause = fmt.Sprintf(" AND search_vector @@ to_tsquery('english', $%d)", argCount)
		scoreSelect = fmt.Sprintf("ts_rank_cd(search_vector, to_tsquery('english', $%d)) AS score", argCount)
		args = append(args, queryTsQuery)
		argCount++
	}

	whereSQL := strings.Join(whereClauses, " AND ") + ftsClause

	sql := fmt.Sprintf(`
		SELECT frontmatter, content, %s
		FROM documents
		WHERE %s
		ORDER BY score DESC
		LIMIT 50
	`, scoreSelect, whereSQL)

	rows, err := p.pool.Query(ctx, sql, args...)
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
			MatchReason: "PostgreSQL tsvector FTS Match",
		})
	}

	return results, nil
}
