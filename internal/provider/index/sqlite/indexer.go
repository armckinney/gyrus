package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/armckinney/gyrus/internal/domain/okf"
	"github.com/armckinney/gyrus/internal/provider/db"
	"github.com/armckinney/gyrus/pkg/gyrus"
	_ "modernc.org/sqlite"
)

// Indexer implements gyrus.IndexStore over SQLite.
type Indexer struct {
	dbPath string
	db     *sql.DB
}

// NewIndexer opens SQLite database at dbPath and runs schema migrations.
func NewIndexer(dbPath string) (*Indexer, error) {
	database, err := db.OpenSQLite(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite index database: %w", err)
	}

	idx := &Indexer{
		dbPath: dbPath,
		db:     database,
	}

	if err := idx.initSchema(); err != nil {
		return nil, err
	}

	return idx, nil
}

func (idx *Indexer) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS documents (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		category TEXT NOT NULL,
		type TEXT NOT NULL,
		owner_group TEXT NOT NULL,
		version INTEGER NOT NULL,
		status TEXT NOT NULL,
		last_modified_by TEXT,
		last_updated DATETIME,
		tags TEXT,
		dependencies TEXT,
		filepath TEXT,
		content TEXT
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS fts_documents USING fts5(
		id UNINDEXED,
		title,
		content,
		tags,
		category,
		type,
		tokenize = 'porter unicode61'
	);

	CREATE TABLE IF NOT EXISTS document_edges (
		from_document_id TEXT NOT NULL,
		to_document_id TEXT NOT NULL,
		relationship_type TEXT NOT NULL,
		created_by TEXT,
		created_at DATETIME,
		PRIMARY KEY (from_document_id, to_document_id, relationship_type)
	);

	CREATE INDEX IF NOT EXISTS idx_edges_from ON document_edges(from_document_id);
	CREATE INDEX IF NOT EXISTS idx_edges_to ON document_edges(to_document_id);
	`
	_, err := idx.db.Exec(schema)
	return err
}

// DB returns the underlying *sql.DB connection.
func (idx *Indexer) DB() *sql.DB {
	return idx.db
}

// Close closes the underlying database connection.
func (idx *Indexer) Close() error {
	return idx.db.Close()
}

// Index inserts or updates a document in the index and FTS tables.
func (idx *Indexer) Index(ctx context.Context, doc gyrus.Document) error {
	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tagsJSON, _ := json.Marshal(doc.Tags)
	depsJSON, _ := json.Marshal(doc.Dependencies)

	query := `
	INSERT INTO documents (id, title, category, type, owner_group, version, status, last_modified_by, last_updated, tags, dependencies, filepath, content)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		title=excluded.title,
		category=excluded.category,
		type=excluded.type,
		owner_group=excluded.owner_group,
		version=excluded.version,
		status=excluded.status,
		last_modified_by=excluded.last_modified_by,
		last_updated=excluded.last_updated,
		tags=excluded.tags,
		dependencies=excluded.dependencies,
		filepath=excluded.filepath,
		content=excluded.content;
	`

	_, err = tx.ExecContext(ctx, query,
		doc.ID, doc.Title, string(doc.Category), string(doc.Type), doc.OwnerGroup,
		doc.Version, doc.Status, doc.LastModifiedBy, doc.LastUpdated.Format(time.RFC3339),
		string(tagsJSON), string(depsJSON), "", doc.Content,
	)
	if err != nil {
		return err
	}

	_, _ = tx.ExecContext(ctx, "DELETE FROM fts_documents WHERE id = ?", doc.ID)

	tagsStr := strings.Join(doc.Tags, " ")
	_, err = tx.ExecContext(ctx, `
	INSERT INTO fts_documents (id, title, content, tags, category, type)
	VALUES (?, ?, ?, ?, ?, ?);
	`, doc.ID, doc.Title, doc.Content, tagsStr, string(doc.Category), string(doc.Type))
	if err != nil {
		return err
	}

	// Auto-extract dependencies into document_edges
	for _, depID := range doc.Dependencies {
		edgeQuery := `
		INSERT INTO document_edges (from_document_id, to_document_id, relationship_type, created_by, created_at)
		VALUES (?, ?, 'depends_on', ?, ?)
		ON CONFLICT(from_document_id, to_document_id, relationship_type) DO NOTHING;
		`
		_, _ = tx.ExecContext(ctx, edgeQuery, doc.ID, depID, doc.OwnerGroup, time.Now().Format(time.RFC3339))
	}

	return tx.Commit()
}

// Remove deletes a document from the index.
func (idx *Indexer) Remove(ctx context.Context, id string) error {
	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM documents WHERE id = ?", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM fts_documents WHERE id = ?", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM document_edges WHERE from_document_id = ? OR to_document_id = ?", id, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Sync scans storageRoot for Markdown files and updates the SQLite index incrementally.
func (idx *Indexer) Sync(ctx context.Context, storageRoot string) (gyrus.SyncReport, error) {
	var report gyrus.SyncReport

	err := filepath.Walk(storageRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		report.ScannedFiles++

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		doc, err := okf.ParseMarkdown(data)
		if err != nil {
			return nil
		}

		if err := idx.Index(ctx, *doc); err != nil {
			return nil
		}

		report.IndexedFiles++
		return nil
	})

	return report, err
}

// Search executes an FTS5 search query with metadata filtering.
func (idx *Indexer) Search(ctx context.Context, query gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	maxRes := query.MaxResults
	if maxRes <= 0 {
		maxRes = 50
	}

	var rows *sql.Rows
	var err error

	if query.Query == "" {
		// Pure metadata filter query over documents table
		var whereClauses []string
		var args []any

		if query.Filter.Category != "" {
			whereClauses = append(whereClauses, "category = ?")
			args = append(args, string(query.Filter.Category))
		}
		if query.Filter.Type != "" {
			whereClauses = append(whereClauses, "type = ?")
			args = append(args, string(query.Filter.Type))
		}
		if query.Filter.Status != "" {
			whereClauses = append(whereClauses, "status = ?")
			args = append(args, query.Filter.Status)
		}
		if query.Filter.OwnerGroup != "" {
			whereClauses = append(whereClauses, "owner_group = ?")
			args = append(args, query.Filter.OwnerGroup)
		}
		if query.Filter.Tag != "" {
			whereClauses = append(whereClauses, "tags LIKE ?")
			args = append(args, "%\""+query.Filter.Tag+"\"%")
		}

		whereSQL := ""
		if len(whereClauses) > 0 {
			whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
		}

		sqlQuery := fmt.Sprintf(`
		SELECT id, title, category, type, owner_group, version, status, last_modified_by, last_updated, tags, dependencies, content, 0.0 AS rank
		FROM documents
		%s
		LIMIT ?;
		`, whereSQL)

		args = append(args, maxRes)
		rows, err = idx.db.QueryContext(ctx, sqlQuery, args...)
		if err != nil {
			return nil, err
		}
	} else {
		// Tokenize search query for FTS OR matching if multiple words
		words := strings.Fields(query.Query)
		var validTokens []string
		for _, w := range words {
			cleaned := strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
					return r
				}
				return -1
			}, w)
			if len(cleaned) >= 2 {
				validTokens = append(validTokens, `"`+cleaned+`"`)
			}
		}

		ftsMatchQuery := query.Query
		if len(validTokens) > 1 {
			ftsMatchQuery = strings.Join(validTokens, " OR ")
		}

		var whereClauses []string
		var args []any

		whereClauses = append(whereClauses, "fts_documents MATCH ?")
		args = append(args, ftsMatchQuery)

		if query.Filter.Category != "" {
			whereClauses = append(whereClauses, "d.category = ?")
			args = append(args, string(query.Filter.Category))
		}
		if query.Filter.Type != "" {
			whereClauses = append(whereClauses, "d.type = ?")
			args = append(args, string(query.Filter.Type))
		}
		if query.Filter.Status != "" {
			whereClauses = append(whereClauses, "d.status = ?")
			args = append(args, query.Filter.Status)
		}
		if query.Filter.OwnerGroup != "" {
			whereClauses = append(whereClauses, "d.owner_group = ?")
			args = append(args, query.Filter.OwnerGroup)
		}
		if query.Filter.Tag != "" {
			whereClauses = append(whereClauses, "d.tags LIKE ?")
			args = append(args, "%\""+query.Filter.Tag+"\"%")
		}

		whereSQL := "WHERE " + strings.Join(whereClauses, " AND ")

		sqlQuery := fmt.Sprintf(`
		SELECT d.id, d.title, d.category, d.type, d.owner_group, d.version, d.status, d.last_modified_by, d.last_updated, d.tags, d.dependencies, d.content, fts.rank
		FROM fts_documents fts
		JOIN documents d ON fts.id = d.id
		%s
		ORDER BY fts.rank
		LIMIT ?;
		`, whereSQL)

		ftsArgs := append(args, maxRes)
		rows, err = idx.db.QueryContext(ctx, sqlQuery, ftsArgs...)
		if err != nil {
			// Fallback to title/content LIKE query
			likeQuery := "%" + query.Query + "%"
			fallbackClauses := []string{"(title LIKE ? OR content LIKE ?)"}
			fallbackArgs := []any{likeQuery, likeQuery}

			if query.Filter.Category != "" {
				fallbackClauses = append(fallbackClauses, "category = ?")
				fallbackArgs = append(fallbackArgs, string(query.Filter.Category))
			}
			if query.Filter.Type != "" {
				fallbackClauses = append(fallbackClauses, "type = ?")
				fallbackArgs = append(fallbackArgs, string(query.Filter.Type))
			}
			if query.Filter.Status != "" {
				fallbackClauses = append(fallbackClauses, "status = ?")
				fallbackArgs = append(fallbackArgs, query.Filter.Status)
			}
			if query.Filter.OwnerGroup != "" {
				fallbackClauses = append(fallbackClauses, "owner_group = ?")
				fallbackArgs = append(fallbackArgs, query.Filter.OwnerGroup)
			}
			if query.Filter.Tag != "" {
				fallbackClauses = append(fallbackClauses, "tags LIKE ?")
				fallbackArgs = append(fallbackArgs, "%\""+query.Filter.Tag+"\"%")
			}

			fallbackWhere := "WHERE " + strings.Join(fallbackClauses, " AND ")
			fallbackSQL := fmt.Sprintf(`
			SELECT id, title, category, type, owner_group, version, status, last_modified_by, last_updated, tags, dependencies, content, 0.0 AS rank
			FROM documents
			%s
			LIMIT ?;
			`, fallbackWhere)

			fallbackArgs = append(fallbackArgs, maxRes)
			rows, err = idx.db.QueryContext(ctx, fallbackSQL, fallbackArgs...)
			if err != nil {
				return nil, err
			}
		}
	}
	defer rows.Close()

	var results []gyrus.SearchResult
	for rows.Next() {
		var doc gyrus.Document
		var catStr, typeStr, lastUpdatedStr, tagsJSON, depsJSON string
		var rank float64

		err := rows.Scan(
			&doc.ID, &doc.Title, &catStr, &typeStr, &doc.OwnerGroup,
			&doc.Version, &doc.Status, &doc.LastModifiedBy, &lastUpdatedStr,
			&tagsJSON, &depsJSON, &doc.Content, &rank,
		)
		if err != nil {
			return nil, err
		}

		doc.Category = gyrus.Category(catStr)
		doc.Type = gyrus.DocumentType(typeStr)
		if lastUpdatedStr != "" {
			doc.LastUpdated, _ = time.Parse(time.RFC3339, lastUpdatedStr)
		}
		_ = json.Unmarshal([]byte(tagsJSON), &doc.Tags)
		_ = json.Unmarshal([]byte(depsJSON), &doc.Dependencies)

		results = append(results, gyrus.SearchResult{
			Document:    doc,
			Score:       -rank,
			MatchReason: "SQLite FTS Match",
		})
	}

	return results, nil
}
