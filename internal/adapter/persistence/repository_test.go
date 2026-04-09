package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/Lim0os/price-oracle/internal/domain"
	"github.com/gofrs/uuid/v5"
	"github.com/pashagolub/pgxmock/v4"
)

func TestRateRepository_Save_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	id, _ := uuid.NewV7()
	rate := &domain.Rate{
		ID:        id,
		Ask:       "1.001",
		Bid:       "0.999",
		Strategy:  "topN",
		N:         1,
		M:         0,
		FetchedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO rates").
		WithArgs(id, rate.Ask, rate.Bid, rate.Strategy, rate.N, rate.M, rate.FetchedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	repo := &RateRepository{pool: mock}
	err = repo.Save(context.Background(), rate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRateRepository_Save_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	id, _ := uuid.NewV7()
	rate := &domain.Rate{
		ID:        id,
		Ask:       "1.001",
		Bid:       "0.999",
		Strategy:  "topN",
		N:         1,
		M:         0,
		FetchedAt: time.Now(),
	}

	mock.ExpectExec("INSERT INTO rates").
		WithArgs(id, rate.Ask, rate.Bid, rate.Strategy, rate.N, rate.M, rate.FetchedAt).
		WillReturnError(context.DeadlineExceeded)

	repo := &RateRepository{pool: mock}
	err = repo.Save(context.Background(), rate)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestRateRepository_ImplementsInterface(_ *testing.T) {
	// Compile-time check: RateRepository must implement domain.RateRepository
	var _ domain.RateRepository = (*RateRepository)(nil)
}
