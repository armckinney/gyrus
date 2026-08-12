package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	sqliteMu    sync.Mutex
	sqlitePools = make(map[string]*sql.DB)
)

// OpenSQLite returns a shared *sql.DB connection pool for the given dbPath.
func OpenSQLite(dbPath string) (*sql.DB, error) {
	sqliteMu.Lock()
	defer sqliteMu.Unlock()

	absPath, err := filepath.Abs(dbPath)
	if err == nil {
		dbPath = absPath
	}

	if pool, exists := sqlitePools[dbPath]; exists {
		if err := pool.Ping(); err == nil {
			return pool, nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for SQLite database: %w", err)
	}

	pool, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	pool.SetMaxOpenConns(10)
	pool.SetMaxIdleConns(5)

	sqlitePools[dbPath] = pool
	return pool, nil
}

// CloseSQLite closes a shared SQLite connection pool for the given dbPath.
func CloseSQLite(dbPath string) error {
	sqliteMu.Lock()
	defer sqliteMu.Unlock()

	absPath, err := filepath.Abs(dbPath)
	if err == nil {
		dbPath = absPath
	}

	if pool, exists := sqlitePools[dbPath]; exists {
		delete(sqlitePools, dbPath)
		return pool.Close()
	}
	return nil
}
