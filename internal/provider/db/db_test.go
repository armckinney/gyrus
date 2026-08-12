package db_test

import (
	"path/filepath"
	"testing"

	"github.com/armckinney/gyrus/internal/provider/db"
)

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
