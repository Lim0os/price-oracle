package persistence

import (
	"context"
	"fmt"

	"github.com/Lim0os/price-oracle/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

// PgxPool is the interface for pgxpool.Pool, allowing mocking in tests.
type PgxPool interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Close()
}

const insertRateQuery = `
		INSERT INTO rates (id, ask, bid, strategy, n, m, fetched_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`

// RateRepository persists rates to PostgreSQL.
type RateRepository struct {
	pool PgxPool
}

// NewRateRepository creates a new RateRepository.
func NewRateRepository(pool PgxPool) *RateRepository {
	return &RateRepository{pool: pool}
}

// Save persists a calculated rate into the rates table.
func (r *RateRepository) Save(ctx context.Context, rate *domain.Rate) error {
	_, err := r.pool.Exec(ctx, insertRateQuery,
		rate.ID,
		rate.Ask,
		rate.Bid,
		rate.Strategy,
		rate.N,
		rate.M,
		rate.FetchedAt,
	)
	if err != nil {
		return fmt.Errorf("save rate: %w", err)
	}

	return nil
}
