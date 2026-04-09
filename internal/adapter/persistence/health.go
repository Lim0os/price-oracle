// Package persistence provides PostgreSQL adapters for rate storage and health checking.
package persistence

import (
	"context"
	"fmt"
)

// DBHealthChecker checks PostgreSQL connectivity.
type DBHealthChecker struct {
	pool PgxPool
}

// NewDBHealthChecker creates a new health checker.
func NewDBHealthChecker(pool PgxPool) *DBHealthChecker {
	return &DBHealthChecker{pool: pool}
}

// Ping checks if the database connection is alive.
func (h *DBHealthChecker) Ping(ctx context.Context) error {
	if err := h.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}
