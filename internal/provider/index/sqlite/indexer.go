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
	// Migrate legacy document_edges table if it has from_id instead of from_document_id
	var hasFromID, hasFromDocID bool
	rows, err := idx.db.Query("PRAGMA table_info(document_edges)")
	if err == nil {
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue interface{}
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil {
				if name == "from_id" {
					hasFromID = true
				}
				if name == "from_document_id" {
					hasFromDocID = true
				}
			}
		}
		rows.Close()
	}
	if hasFromID && !hasFromDocID {
		migration := `
		ALTER TABLE document_edges RENAME TO old_document_edges;
		CREATE TABLE document_edges (
			from_document_id TEXT NOT NULL,
			to_document_id TEXT NOT NULL,
			relationship_type TEXT NOT NULL,
			created_by TEXT,
			created_at DATETIME,
			PRIMARY KEY (from_document_id, to_document_id, relationship_type)
		);
		INSERT INTO document_edges (from_document_id, to_document_id, relationship_type, created_by, created_at)
		SELECT from_id, to_id, relationship_type, created_by, created_at FROM old_document_edges;
		DROP TABLE old_document_edges;
		`
		if _, err := idx.db.Exec(migration); err != nil {
			return err
		}
	}

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
		content TEXT,
		scope TEXT DEFAULT 'reference',
		workspace TEXT DEFAULT ''
	);

	CREATE VIRTUAL TABLE IF NOT EXISTS fts_documents USING fts5(
		id,
		title,
		content,
		tags,
		category,
		type,
		scope,
		workspace,
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
	if _, err = idx.db.Exec(schema); err != nil {
		return err
	}

	// Migrate documents table if missing scope or workspace columns
	var hasScope, hasWorkspace bool
	docCols, err := idx.db.Query("PRAGMA table_info(documents)")
	if err == nil {
		for docCols.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue interface{}
			if err := docCols.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil {
				if name == "scope" {
					hasScope = true
				}
				if name == "workspace" {
					hasWorkspace = true
				}
			}
		}
		docCols.Close()
	}
	if !hasScope {
		_, _ = idx.db.Exec("ALTER TABLE documents ADD COLUMN scope TEXT DEFAULT 'reference';")
	}
	if !hasWorkspace {
		_, _ = idx.db.Exec("ALTER TABLE documents ADD COLUMN workspace TEXT DEFAULT '';")
	}

	// Migrate fts_documents schema if it lacks workspace column or had id UNINDEXED
	var existingFTS string
	_ = idx.db.QueryRow("SELECT sql FROM sqlite_master WHERE name = 'fts_documents'").Scan(&existingFTS)
	if existingFTS != "" && (!strings.Contains(existingFTS, "workspace") || strings.Contains(existingFTS, "id UNINDEXED")) {
		_, _ = idx.db.Exec("DROP TABLE IF EXISTS fts_documents;")
		_, _ = idx.db.Exec(`
		CREATE VIRTUAL TABLE fts_documents USING fts5(
			id,
			title,
			content,
			tags,
			category,
			type,
			scope,
			workspace,
			tokenize = 'porter unicode61'
		);
		`)
		_, _ = idx.db.Exec(`
		INSERT INTO fts_documents (id, title, content, tags, category, type, scope, workspace)
		SELECT id, title, content, tags, category, type, coalesce(scope, 'reference'), coalesce(workspace, '') FROM documents;
		`)
	}

	return nil
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

	scope := doc.Scope
	if scope == "" {
		if doc.Workspace != "" {
			scope = "workspace"
		} else {
			scope = "reference"
		}
	}
	workspace := doc.Workspace

	query := `
	INSERT INTO documents (id, title, category, type, owner_group, version, status, last_modified_by, last_updated, tags, dependencies, filepath, content, scope, workspace)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		content=excluded.content,
		scope=excluded.scope,
		workspace=excluded.workspace;
	`

	_, err = tx.ExecContext(ctx, query,
		doc.ID, doc.Title, string(doc.Category), string(doc.Type), doc.OwnerGroup,
		doc.Version, doc.Status, doc.LastModifiedBy, doc.LastUpdated.Format(time.RFC3339),
		string(tagsJSON), string(depsJSON), doc.FilePath, doc.Content, scope, workspace,
	)
	if err != nil {
		return err
	}

	_, _ = tx.ExecContext(ctx, "DELETE FROM fts_documents WHERE id = ?", doc.ID)

	tagsStr := strings.Join(doc.Tags, " ")
	_, err = tx.ExecContext(ctx, `
	INSERT INTO fts_documents (id, title, content, tags, category, type, scope, workspace)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`, doc.ID, doc.Title, doc.Content, tagsStr, string(doc.Category), string(doc.Type), scope, workspace)
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

// Sync scans storageRoot for Markdown files, updates the SQLite index incrementally,
// and prunes orphaned records that no longer exist in the storage root.
func (idx *Indexer) Sync(ctx context.Context, storageRoot string) (gyrus.SyncReport, error) {
	var report gyrus.SyncReport
	activeIDs := make(map[string]bool)

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

		cleanPath := filepath.ToSlash(path)
		doc.FilePath = cleanPath
		idxDocs := strings.Index(cleanPath, "/docs/")
		if idxDocs != -1 {
			sub := cleanPath[idxDocs+len("/docs/"):]
			parts := strings.Split(sub, "/")
			if len(parts) >= 3 && parts[1] == "reference" {
				doc.Scope = "reference"
				doc.Workspace = ""
			} else if len(parts) >= 3 && parts[1] == "workspaces" && parts[2] != "" {
				doc.Scope = "workspace"
				doc.Workspace = parts[2]
			} else {
				doc.Scope = "reference"
				doc.Workspace = ""
			}
		} else if strings.Contains(cleanPath, "/reference/") {
			doc.Scope = "reference"
			doc.Workspace = ""
		} else if idxWs := strings.LastIndex(cleanPath, "/workspaces/"); idxWs != -1 {
			rest := cleanPath[idxWs+len("/workspaces/"):]
			parts := strings.Split(rest, "/")
			if len(parts) >= 1 && parts[0] != "" {
				doc.Scope = "workspace"
				doc.Workspace = parts[0]
			} else {
				doc.Scope = "reference"
				doc.Workspace = ""
			}
		} else {
			doc.Scope = "reference"
			doc.Workspace = ""
		}

		activeIDs[doc.ID] = true

		if err := idx.Index(ctx, *doc); err != nil {
			return nil
		}

		report.IndexedFiles++
		return nil
	})
	if err != nil {
		return report, err
	}

	// Prune orphaned records from SQLite that are not present on disk
	rows, err := idx.db.QueryContext(ctx, "SELECT id FROM documents")
	if err == nil {
		defer rows.Close()
		var orphanedIDs []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err == nil {
				if !activeIDs[id] {
					orphanedIDs = append(orphanedIDs, id)
				}
			}
		}

		for _, id := range orphanedIDs {
			if err := idx.Remove(ctx, id); err == nil {
				report.RemovedFiles++
			}
		}
	}

	return report, nil
}

// Search executes an FTS5 search query with ID boosting, workspace prioritization, title matching, and metadata filtering.
func (idx *Indexer) Search(ctx context.Context, query gyrus.SearchQuery) ([]gyrus.SearchResult, error) {
	maxRes := query.MaxResults
	if maxRes <= 0 {
		maxRes = 50
	}

	var rows *sql.Rows
	var err error

	rawQuery := strings.TrimSpace(query.Query)
	targetWS := query.Filter.Workspace

	// Build common scope filter condition
	buildScopeCondition := func(tablePrefix string) (string, []any) {
		colScope := "scope"
		colWS := "workspace"
		if tablePrefix != "" {
			colScope = tablePrefix + ".scope"
			colWS = tablePrefix + ".workspace"
		}

		switch query.Filter.Scope {
		case "reference":
			return colScope + " = 'reference'", nil
		case "workspace":
			if targetWS != "" {
				return colWS + " = ?", []any{targetWS}
			}
			return colScope + " = 'workspace'", nil
		case "all":
			return "", nil
		default:
			// Default scoping behavior
			if targetWS != "" {
				return "(" + colWS + " = ? OR " + colScope + " = 'reference')", []any{targetWS}
			}
			return colScope + " = 'reference'", nil
		}
	}

	if rawQuery == "" {
		// Pure metadata filter query over documents table
		var whereClauses []string
		var args []any

		if scopeCond, scopeArgs := buildScopeCondition(""); scopeCond != "" {
			whereClauses = append(whereClauses, scopeCond)
			args = append(args, scopeArgs...)
		}
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
		SELECT id, title, category, type, owner_group, version, status, last_modified_by, last_updated, tags, dependencies, content,
			COALESCE(scope, 'reference') AS scope, COALESCE(workspace, '') AS workspace, COALESCE(filepath, '') AS filepath,
			CASE
				WHEN workspace != '' AND workspace = ? THEN 60.0
				WHEN scope = 'reference' THEN 20.0
				ELSE 0.0
			END AS score,
			CASE
				WHEN workspace != '' AND workspace = ? THEN 'Workspace Document'
				WHEN scope = 'reference' THEN 'Reference Document'
				ELSE 'All Documents'
			END AS match_reason
		FROM documents
		%s
		ORDER BY score DESC, id ASC
		LIMIT ?;
		`, whereSQL)

		queryArgs := []any{targetWS, targetWS}
		queryArgs = append(queryArgs, args...)
		queryArgs = append(queryArgs, maxRes)

		rows, err = idx.db.QueryContext(ctx, sqlQuery, queryArgs...)
		if err != nil {
			return nil, err
		}
	} else {
		idExact := strings.ToLower(rawQuery)
		idLike := "%" + idExact + "%"
		titleLike := "%" + idExact + "%"

		words := strings.Fields(rawQuery)
		var validTokens []string
		for _, w := range words {
			cleaned := strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
					return r
				}
				return -1
			}, w)
			if len(cleaned) >= 1 {
				validTokens = append(validTokens, `"`+cleaned+`"*`)
			}
		}

		ftsMatchQuery := strings.Join(validTokens, " OR ")
		if ftsMatchQuery == "" {
			ftsMatchQuery = `"` + rawQuery + `"*`
		}

		var filterClauses []string
		var filterArgs []any

		if scopeCond, scopeArgs := buildScopeCondition("d"); scopeCond != "" {
			filterClauses = append(filterClauses, scopeCond)
			filterArgs = append(filterArgs, scopeArgs...)
		}
		if query.Filter.Category != "" {
			filterClauses = append(filterClauses, "d.category = ?")
			filterArgs = append(filterArgs, string(query.Filter.Category))
		}
		if query.Filter.Type != "" {
			filterClauses = append(filterClauses, "d.type = ?")
			filterArgs = append(filterArgs, string(query.Filter.Type))
		}
		if query.Filter.Status != "" {
			filterClauses = append(filterClauses, "d.status = ?")
			filterArgs = append(filterArgs, query.Filter.Status)
		}
		if query.Filter.OwnerGroup != "" {
			filterClauses = append(filterClauses, "d.owner_group = ?")
			filterArgs = append(filterArgs, query.Filter.OwnerGroup)
		}
		if query.Filter.Tag != "" {
			filterClauses = append(filterClauses, "d.tags LIKE ?")
			filterArgs = append(filterArgs, "%\""+query.Filter.Tag+"\"%")
		}

		filterSQL := ""
		if len(filterClauses) > 0 {
			filterSQL = "AND " + strings.Join(filterClauses, " AND ")
		}

		sqlQuery := fmt.Sprintf(`
		SELECT 
			d.id, d.title, d.category, d.type, d.owner_group, d.version, d.status, 
			d.last_modified_by, d.last_updated, d.tags, d.dependencies, d.content,
			COALESCE(d.scope, 'reference') AS scope, COALESCE(d.workspace, '') AS workspace, COALESCE(d.filepath, '') AS filepath,
			CASE 
				WHEN LOWER(d.id) = ? THEN 100.0
				WHEN d.workspace != '' AND d.workspace = ? THEN 60.0 + COALESCE(-fts.rank, 1.0)
				WHEN LOWER(d.id) LIKE ? THEN 50.0
				WHEN LOWER(d.title) LIKE ? THEN 30.0
				WHEN d.scope = 'reference' THEN 20.0 + COALESCE(-fts.rank, 1.0)
				ELSE COALESCE(-fts.rank, 1.0)
			END AS score,
			CASE
				WHEN LOWER(d.id) = ? THEN 'Exact ID Match'
				WHEN d.workspace != '' AND d.workspace = ? THEN 'Workspace Match'
				WHEN LOWER(d.id) LIKE ? THEN 'Document ID Match'
				WHEN LOWER(d.title) LIKE ? THEN 'Title Match'
				WHEN d.scope = 'reference' THEN 'Reference Match'
				ELSE 'SQLite FTS Match'
			END AS match_reason
		FROM documents d
		LEFT JOIN fts_documents fts ON d.id = fts.id AND fts.fts_documents MATCH ?
		WHERE (
			LOWER(d.id) = ?
			OR LOWER(d.id) LIKE ?
			OR LOWER(d.title) LIKE ?
			OR fts.id IS NOT NULL
		)
		%s
		ORDER BY score DESC
		LIMIT ?;
		`, filterSQL)

		queryArgs := []any{
			idExact, targetWS, idLike, titleLike,
			idExact, targetWS, idLike, titleLike,
			ftsMatchQuery,
			idExact, idLike, titleLike,
		}
		queryArgs = append(queryArgs, filterArgs...)
		queryArgs = append(queryArgs, maxRes)

		rows, err = idx.db.QueryContext(ctx, sqlQuery, queryArgs...)
		if err != nil {
			// Graceful fallback to pure LIKE query if FTS expression has syntax anomalies
			var fallbackFilterClauses []string
			var fallbackFilterArgs []any

			if scopeCond, scopeArgs := buildScopeCondition(""); scopeCond != "" {
				fallbackFilterClauses = append(fallbackFilterClauses, scopeCond)
				fallbackFilterArgs = append(fallbackFilterArgs, scopeArgs...)
			}
			if query.Filter.Category != "" {
				fallbackFilterClauses = append(fallbackFilterClauses, "category = ?")
				fallbackFilterArgs = append(fallbackFilterArgs, string(query.Filter.Category))
			}
			if query.Filter.Type != "" {
				fallbackFilterClauses = append(fallbackFilterClauses, "type = ?")
				fallbackFilterArgs = append(fallbackFilterArgs, string(query.Filter.Type))
			}
			if query.Filter.Status != "" {
				fallbackFilterClauses = append(fallbackFilterClauses, "status = ?")
				fallbackFilterArgs = append(fallbackFilterArgs, query.Filter.Status)
			}
			if query.Filter.OwnerGroup != "" {
				fallbackFilterClauses = append(fallbackFilterClauses, "owner_group = ?")
				fallbackFilterArgs = append(fallbackFilterArgs, query.Filter.OwnerGroup)
			}
			if query.Filter.Tag != "" {
				fallbackFilterClauses = append(fallbackFilterClauses, "tags LIKE ?")
				fallbackFilterArgs = append(fallbackFilterArgs, "%\""+query.Filter.Tag+"\"%")
			}

			fallbackFilterSQL := ""
			if len(fallbackFilterClauses) > 0 {
				fallbackFilterSQL = "AND " + strings.Join(fallbackFilterClauses, " AND ")
			}

			fallbackSQL := fmt.Sprintf(`
			SELECT 
				id, title, category, type, owner_group, version, status, 
				last_modified_by, last_updated, tags, dependencies, content,
				COALESCE(scope, 'reference') AS scope, COALESCE(workspace, '') AS workspace, COALESCE(filepath, '') AS filepath,
				CASE 
					WHEN LOWER(id) = ? THEN 100.0
					WHEN workspace != '' AND workspace = ? THEN 60.0
					WHEN LOWER(id) LIKE ? THEN 50.0
					WHEN LOWER(title) LIKE ? THEN 30.0
					WHEN scope = 'reference' THEN 20.0
					ELSE 1.0
				END AS score,
				CASE
					WHEN LOWER(id) = ? THEN 'Exact ID Match'
					WHEN workspace != '' AND workspace = ? THEN 'Workspace Match'
					WHEN LOWER(id) LIKE ? THEN 'Document ID Match'
					WHEN LOWER(title) LIKE ? THEN 'Title Match'
					WHEN scope = 'reference' THEN 'Reference Match'
					ELSE 'Content Match'
				END AS match_reason
			FROM documents
			WHERE (
				LOWER(id) = ?
				OR LOWER(id) LIKE ?
				OR LOWER(title) LIKE ?
				OR LOWER(content) LIKE ?
			)
			%s
			ORDER BY score DESC
			LIMIT ?;
			`, fallbackFilterSQL)

			fallbackArgs := []any{
				idExact, targetWS, idLike, titleLike,
				idExact, targetWS, idLike, titleLike,
				idExact, idLike, titleLike, "%" + idExact + "%",
			}
			fallbackArgs = append(fallbackArgs, fallbackFilterArgs...)
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
		var score float64
		var matchReason string

		err := rows.Scan(
			&doc.ID, &doc.Title, &catStr, &typeStr, &doc.OwnerGroup,
			&doc.Version, &doc.Status, &doc.LastModifiedBy, &lastUpdatedStr,
			&tagsJSON, &depsJSON, &doc.Content,
			&doc.Scope, &doc.Workspace, &doc.FilePath,
			&score, &matchReason,
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
			Score:       score,
			MatchReason: matchReason,
		})
	}

	return results, nil
}
