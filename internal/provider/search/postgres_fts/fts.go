package postgres_fts

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/internal/persistence/db"
	"github.com/armckinney/gyrus/pkg/gyrus"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PgxPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Close()
}

type SearchProvider struct {
	pool PgxPool
}

func NewSearchProvider(ctx context.Context, connString string) (*SearchProvider, error) {
	pool, err := db.OpenPostgres(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("open postgres search pool: %w", err)
	}
	return NewSearchProviderWithPool(pool), nil
}

func NewSearchProviderWithPool(pool PgxPool) *SearchProvider {
	return &SearchProvider{pool: pool}
}

func (p *SearchProvider) Search(ctx context.Context, request gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	filter := request.Filter
	var conditions []string
	var args []any
	bind := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}
	for _, field := range []struct{ name, value string }{
		{"category", string(filter.Category)}, {"type", string(filter.Type)},
		{"status", filter.Status}, {"tenant", filter.Tenant},
	} {
		if field.value != "" {
			conditions = append(conditions, "frontmatter->>'"+field.name+"' = "+bind(field.value))
		}
	}
	if filter.Tag != "" {
		tag, err := json.Marshal([]string{filter.Tag})
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, "frontmatter->'tags' @> "+bind(string(tag))+"::jsonb")
	}
	const scope = "COALESCE(frontmatter->>'scope', 'reference')"
	const project = "COALESCE(frontmatter->>'project', '')"
	switch filter.Scope {
	case "all":
	case "reference":
		conditions = append(conditions, scope+" = 'reference'")
	case "project":
		conditions = append(conditions, scope+" = 'project'")
		if filter.Project != "" {
			conditions = append(conditions, project+" = "+bind(filter.Project))
		}
	case "":
		condition := scope + " = 'reference'"
		if filter.Project != "" {
			condition = "(" + condition + " OR " + project + " = " + bind(filter.Project) + ")"
		}
		conditions = append(conditions, condition)
	default:
		return nil, &okf.ValidationError{Field: "scope", Message: "must be project, reference, or all"}
	}

	query := strings.TrimSpace(request.Query)
	score := "0::float8"
	reason := "All Concepts"
	if query != "" && okf.ValidateName(query) == nil {
		conditions = append(conditions, "id = "+bind(query))
		score, reason = "1000::float8", "Exact ID Match"
	} else if query != "" {
		term := bind(strings.Join(strings.Fields(query), " OR "))
		expression := "websearch_to_tsquery('english', " + term + ")"
		conditions = append(conditions, "search_vector @@ "+expression)
		score, reason = "ts_rank_cd(search_vector, "+expression+")", "PostgreSQL tsvector FTS Match"
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}
	statement := "SELECT id, frontmatter, content, " + score + " AS score FROM concepts" + where + " ORDER BY score DESC, id ASC"
	request.Query = query
	if limit := request.Limit(); limit > 0 {
		statement += " LIMIT " + bind(limit)
	}
	rows, err := p.pool.Query(ctx, statement, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]gyrus.SearchResult, 0)
	for rows.Next() {
		var name, content string
		var metadata []byte
		var rank float64
		if err := rows.Scan(&name, &metadata, &content, &rank); err != nil {
			return nil, err
		}
		doc, err := okf.ParseJSON(metadata)
		if err != nil {
			return nil, err
		}
		tenant, project, id, err := okf.ParseName(name)
		if err != nil {
			return nil, err
		}
		if doc.ID != id || doc.Tenant != tenant {
			return nil, fmt.Errorf("concept metadata does not match qualified name %q", name)
		}
		doc.Name, doc.Project, doc.Scope, doc.Content = name, project, "reference", content
		if project != "" {
			doc.Scope = "project"
		}
		results = append(results, gyrus.SearchResult{Concept: *doc, Score: rank, MatchReason: reason})
	}
	return results, rows.Err()
}
