package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const ddl = `
CREATE TABLE IF NOT EXISTS documents_index (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	category TEXT NOT NULL,
	type TEXT NOT NULL,
	format TEXT NOT NULL,
	owner_group TEXT NOT NULL,
	version INTEGER NOT NULL,
	status TEXT NOT NULL,
	last_modified_by TEXT NOT NULL,
	last_updated DATETIME NOT NULL,
	tags TEXT,
	dependencies TEXT,
	file_path TEXT NOT NULL,
	checksum TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_docs_category ON documents_index(category);
CREATE INDEX IF NOT EXISTS idx_docs_type ON documents_index(type);
CREATE INDEX IF NOT EXISTS idx_docs_status ON documents_index(status);
CREATE INDEX IF NOT EXISTS idx_docs_owner ON documents_index(owner_group);

CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
	id UNINDEXED,
	title,
	content,
	tags,
	category,
	type,
	owner_group
);

CREATE TABLE IF NOT EXISTS document_edges (
	from_id TEXT NOT NULL,
	to_id TEXT NOT NULL,
	relationship_type TEXT NOT NULL,
	created_by TEXT NOT NULL,
	created_at DATETIME NOT NULL,
	PRIMARY KEY (from_id, to_id, relationship_type)
);

CREATE INDEX IF NOT EXISTS idx_edges_from ON document_edges(from_id);
CREATE INDEX IF NOT EXISTS idx_edges_to ON document_edges(to_id);
`

// Migrate initializes the SQLite database schema DDL tables and indexes.
func Migrate(db *sql.DB) error {
	_, err := db.Exec(ddl)
	if err != nil {
		return fmt.Errorf("sqlite DDL migration failed: %w", err)
	}
	return nil
}
