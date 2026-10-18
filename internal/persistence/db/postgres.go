package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pgMu    sync.Mutex
	pgPools = make(map[string]*pgxpool.Pool)
)

// OpenPostgres returns a shared *pgxpool.Pool connection pool for the given connString.
func OpenPostgres(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	pgMu.Lock()
	defer pgMu.Unlock()

	if pool, exists := pgPools[connString]; exists {
		if err := pool.Ping(ctx); err == nil {
			return pool, nil
		}
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}
	var existing, marker bool
	if err := pool.QueryRow(ctx, "SELECT to_regclass('concepts') IS NOT NULL, to_regclass('gyrus_contract') IS NOT NULL").Scan(&existing, &marker); err != nil {
		pool.Close()
		return nil, err
	}
	if existing && !marker {
		pool.Close()
		return nil, fmt.Errorf("obsolete PostgreSQL identity schema: export legacy data and rebuild for qualified concept names")
	}
	if _, err := pool.Exec(ctx, "CREATE TABLE IF NOT EXISTS gyrus_contract (singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton), version INTEGER NOT NULL)"); err != nil {
		pool.Close()
		return nil, err
	}
	if !marker {
		if _, err := pool.Exec(ctx, "INSERT INTO gyrus_contract (version) VALUES ($1) ON CONFLICT (singleton) DO NOTHING", qualifiedIdentityVersion); err != nil {
			pool.Close()
			return nil, err
		}
	}
	var version int
	if err := pool.QueryRow(ctx, "SELECT version FROM gyrus_contract").Scan(&version); err != nil {
		pool.Close()
		return nil, err
	}
	if version != qualifiedIdentityVersion {
		pool.Close()
		return nil, fmt.Errorf("unsupported PostgreSQL identity schema version %d", version)
	}

	pgPools[connString] = pool
	return pool, nil
}

// ClosePostgres closes a shared PostgreSQL connection pool for the given connString.
func ClosePostgres(connString string) {
	pgMu.Lock()
	defer pgMu.Unlock()

	if pool, exists := pgPools[connString]; exists {
		delete(pgPools, connString)
		pool.Close()
	}
}
