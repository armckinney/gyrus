package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/armckinney/gyrus/pkg/gyrus"
	_ "modernc.org/sqlite"
)

// Indexer handles document metadata indexing and FTS search over SQLite.
type Indexer struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewIndexer connects to SQLite DSN and runs migrations.
func NewIndexer(dsn string) (*Indexer, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	if err := Migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Indexer{db: db}, nil
}

// Close closes the underlying SQLite connection.
func (idx *Indexer) Close() error {
	return idx.db.Close()
}

func (idx *Indexer) Index(ctx context.Context, doc gyrus.Document) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	tagsJSON, _ := json.Marshal(doc.Tags)
	depsJSON, _ := json.Marshal(doc.Dependencies)

	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Upsert into documents_index
	queryIndex := `
	INSERT INTO documents_index (id, title, category, type, format, owner_group, version, status, last_modified_by, last_updated, tags, dependencies, file_path, checksum)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		title=excluded.title,
		category=excluded.category,
		type=excluded.type,
		format=excluded.format,
		owner_group=excluded.owner_group,
		version=excluded.version,
		status=excluded.status,
		last_modified_by=excluded.last_modified_by,
		last_updated=excluded.last_updated,
		tags=excluded.tags,
		dependencies=excluded.dependencies,
		file_path=excluded.file_path,
		checksum=excluded.checksum;
	`

	filePath := fmt.Sprintf("okf/%s/%s.md", doc.OwnerGroup, doc.ID)
	_, err = tx.ExecContext(ctx, queryIndex,
		doc.ID, doc.Title, string(doc.Category), string(doc.Type), doc.Format,
		doc.OwnerGroup, doc.Version, doc.Status, doc.LastModifiedBy,
		doc.LastUpdated.Format(time.RFC3339), string(tagsJSON), string(depsJSON),
		filePath, "sha256-mock",
	)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}

	// Delete existing FTS record and insert new one
	_, _ = tx.ExecContext(ctx, `DELETE FROM documents_fts WHERE id = ?`, doc.ID)

	tagsStr := strings.Join(doc.Tags, " ")
	queryFTS := `
	INSERT INTO documents_fts (id, title, content, tags, category, type, owner_group)
	VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	_, err = tx.ExecContext(ctx, queryFTS,
		doc.ID, doc.Title, doc.Content, tagsStr, string(doc.Category), string(doc.Type), doc.OwnerGroup,
	)
	if err != nil {
		return fmt.Errorf("failed to update fts index: %w", err)
	}

	return tx.Commit()
}

func (idx *Indexer) Remove(ctx context.Context, id string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, _ = tx.ExecContext(ctx, `DELETE FROM documents_index WHERE id = ?`, id)
	_, _ = tx.ExecContext(ctx, `DELETE FROM documents_fts WHERE id = ?`, id)

	return tx.Commit()
}

func (idx *Indexer) Search(ctx context.Context, q gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var sqlQuery strings.Builder
	var args []interface{}

	if q.Query != "" {
		sqlQuery.WriteString(`
		SELECT i.id, i.title, i.category, i.type, i.format, i.owner_group, i.version, i.status, i.last_modified_by, i.last_updated, i.tags, i.dependencies, fts.rank
		FROM documents_fts fts
		JOIN documents_index i ON fts.id = i.id
		WHERE documents_fts MATCH ?
		`)
		args = append(args, q.Query)
	} else {
		sqlQuery.WriteString(`
		SELECT id, title, category, type, format, owner_group, version, status, last_modified_by, last_updated, tags, dependencies, 0.0 as rank
		FROM documents_index
		WHERE 1=1
		`)
	}

	if q.Filter.Category != "" {
		sqlQuery.WriteString(" AND category = ?")
		args = append(args, string(q.Filter.Category))
	}
	if q.Filter.Type != "" {
		sqlQuery.WriteString(" AND type = ?")
		args = append(args, string(q.Filter.Type))
	}
	if q.Filter.Status != "" {
		sqlQuery.WriteString(" AND status = ?")
		args = append(args, q.Filter.Status)
	}
	if q.Filter.OwnerGroup != "" {
		sqlQuery.WriteString(" AND owner_group = ?")
		args = append(args, q.Filter.OwnerGroup)
	}

	sqlQuery.WriteString(" ORDER BY rank ASC")

	limit := q.MaxResults
	if limit <= 0 {
		limit = 20
	}
	sqlQuery.WriteString(fmt.Sprintf(" LIMIT %d", limit))

	rows, err := idx.db.QueryContext(ctx, sqlQuery.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("search query failed: %w", err)
	}
	defer rows.Close()

	var results []gyrus.SearchResult
	for rows.Next() {
		var doc gyrus.Document
		var catStr, typeStr, lastUpdatedStr, tagsJSON, depsJSON string
		var rank float64

		if err := rows.Scan(
			&doc.ID, &doc.Title, &catStr, &typeStr, &doc.Format, &doc.OwnerGroup,
			&doc.Version, &doc.Status, &doc.LastModifiedBy, &lastUpdatedStr,
			&tagsJSON, &depsJSON, &rank,
		); err != nil {
			return nil, err
		}

		doc.Category = gyrus.Category(catStr)
		doc.Type = gyrus.DocumentType(typeStr)
		_ = json.Unmarshal([]byte(tagsJSON), &doc.Tags)
		_ = json.Unmarshal([]byte(depsJSON), &doc.Dependencies)
		doc.LastUpdated, _ = time.Parse(time.RFC3339, lastUpdatedStr)

		results = append(results, gyrus.SearchResult{
			Document: doc,
			Score:    rank,
		})
	}

	return results, nil
}
