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
