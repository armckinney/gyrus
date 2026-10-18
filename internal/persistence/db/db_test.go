package db_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armckinney/gyrus/internal/persistence/db"
)

func TestObsoleteSQLiteIdentitySchemaRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	legacy, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer legacy.Close()
	if _, err := legacy.Exec("CREATE TABLE concepts(id TEXT PRIMARY KEY); INSERT INTO concepts VALUES ('legacy')"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.OpenSQLite(path); err == nil || !strings.Contains(err.Error(), "obsolete") {
		t.Fatalf("legacy schema reused: %v", err)
	}
	var id string
	if err := legacy.QueryRow("SELECT id FROM concepts").Scan(&id); err != nil || id != "legacy" {
		t.Fatalf("legacy data changed: %s %v", id, err)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies that OpenSQLite implements thread-safe connection pool sharing across services.
// [Execution Surface]: In-Memory SQLite Connection Registry (internal/persistence/db)
// [Assertions]: Repeated OpenSQLite calls for same DB path return identical *sql.DB pointer instance.
// -----------------------------------------------------------------------------
func TestOpenSQLitePoolSharing(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	pool1, err := db.OpenSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to open SQLite pool 1: %v", err)
	}

	pool2, err := db.OpenSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to open SQLite pool 2: %v", err)
	}

	if pool1 != pool2 {
		t.Errorf("Expected identical shared *sql.DB pool pointer for same dbPath, got different pointers")
	}

	if err := db.CloseSQLite(dbPath); err != nil {
		t.Errorf("Failed to close SQLite pool: %v", err)
	}
}
