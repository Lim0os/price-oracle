package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Lim0os/price-oracle/internal/adapter/telemetry"
	"github.com/Lim0os/price-oracle/internal/domain"
)

// mockFetcher is a test double for domain.OrderBookFetcher.
type mockFetcher struct {
	book *domain.OrderBook
	err  error
}

func (m *mockFetcher) FetchOrderBook(_ context.Context, _ string) (*domain.OrderBook, error) {
	return m.book, m.err
}

// mockRepo is a test double for domain.RateRepository.
type mockRepo struct {
	saveErr error
}

func (m *mockRepo) Save(_ context.Context, _ *domain.Rate) error {
	return m.saveErr
}

func TestGetRatesHandler_Success_TopN(t *testing.T) {
	fetcher := &mockFetcher{
		book: &domain.OrderBook{
			Asks: []domain.OrderBookEntry{{Price: "1.001", Quantity: "10"}},
			Bids: []domain.OrderBookEntry{{Price: "0.999", Quantity: "10"}},
		},
	}
	repo := &mockRepo{}
	tracer := telemetry.NewTracer()

	handler := NewGetRatesHandler(fetcher, repo, "usdta7a5", tracer)

	resp, err := handler.Handle(context.Background(), GetRatesRequest{
		Strategy: "topN",
		N:        1,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Ask != "1.001" {
		t.Errorf("expected ask 1.001, got %s", resp.Ask)
	}
	if resp.Bid != "0.999" {
		t.Errorf("expected bid 0.999, got %s", resp.Bid)
	}
	if resp.FetchedAt.IsZero() {
		t.Error("expected non-zero fetched_at")
	}
}

func TestGetRatesHandler_InvalidStrategy(t *testing.T) {
	fetcher := &mockFetcher{
		book: &domain.OrderBook{
			Asks: []domain.OrderBookEntry{{Price: "1.001", Quantity: "10"}},
			Bids: []domain.OrderBookEntry{{Price: "0.999", Quantity: "10"}},
		},
	}
	repo := &mockRepo{}
	tracer := telemetry.NewTracer()

	handler := NewGetRatesHandler(fetcher, repo, "usdta7a5", tracer)

	_, err := handler.Handle(context.Background(), GetRatesRequest{
		Strategy: "unknown",
		N:        1,
	})

	if !errors.Is(err, domain.ErrInvalidStrategy) {
		t.Errorf("expected ErrInvalidStrategy, got %v", err)
	}
}

func TestGetRatesHandler_ExchangeError(t *testing.T) {
	expectedErr := errors.New("exchange unavailable")
	fetcher := &mockFetcher{err: expectedErr}
	repo := &mockRepo{}
	tracer := telemetry.NewTracer()

	handler := NewGetRatesHandler(fetcher, repo, "usdta7a5", tracer)

	_, err := handler.Handle(context.Background(), GetRatesRequest{
		Strategy: "topN",
		N:        1,
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected exchange error, got %v", err)
	}
}

func TestGetRatesHandler_RepositoryError(t *testing.T) {
	fetcher := &mockFetcher{
		book: &domain.OrderBook{
			Asks: []domain.OrderBookEntry{{Price: "1.001", Quantity: "10"}},
			Bids: []domain.OrderBookEntry{{Price: "0.999", Quantity: "10"}},
		},
	}
	expectedErr := errors.New("database write failed")
	repo := &mockRepo{saveErr: expectedErr}
	tracer := telemetry.NewTracer()

	handler := NewGetRatesHandler(fetcher, repo, "usdta7a5", tracer)

	_, err := handler.Handle(context.Background(), GetRatesRequest{
		Strategy: "topN",
		N:        1,
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected repository error, got %v", err)
	}
}

func TestGetRatesHandler_Success_AvgNM(t *testing.T) {
	fetcher := &mockFetcher{
		book: &domain.OrderBook{
			Asks: []domain.OrderBookEntry{{Price: "1", Quantity: "10"}, {Price: "3", Quantity: "20"}},
			Bids: []domain.OrderBookEntry{{Price: "2", Quantity: "10"}, {Price: "4", Quantity: "20"}},
		},
	}
	repo := &mockRepo{}
	tracer := telemetry.NewTracer()

	handler := NewGetRatesHandler(fetcher, repo, "usdta7a5", tracer)

	resp, err := handler.Handle(context.Background(), GetRatesRequest{
		Strategy: "avgNM",
		N:        1,
		M:        2,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// avg(1,3)=2, avg(2,4)=3
	if resp.Ask != "2" {
		t.Errorf("expected ask 2, got %s", resp.Ask)
	}
	if resp.Bid != "3" {
		t.Errorf("expected bid 3, got %s", resp.Bid)
	}
}

func TestGetRatesHandler_EmptyOrderBook(t *testing.T) {
	fetcher := &mockFetcher{
		book: &domain.OrderBook{
			Asks: []domain.OrderBookEntry{},
			Bids: []domain.OrderBookEntry{},
		},
	}
	repo := &mockRepo{}
	tracer := telemetry.NewTracer()

	handler := NewGetRatesHandler(fetcher, repo, "usdta7a5", tracer)

	_, err := handler.Handle(context.Background(), GetRatesRequest{
		Strategy: "topN",
		N:        1,
	})

	if !errors.Is(err, domain.ErrEmptyOrderBook) {
		t.Errorf("expected ErrEmptyOrderBook, got %v", err)
	}
}

func TestNewRate_UUIDGenerated(t *testing.T) {
	now := time.Now()
	rate := domain.NewRate("1.001", "0.999", "topN", 1, 0, now)

	if rate.ID.IsNil() {
		t.Error("expected non-nil UUID")
	}
	if rate.Ask != "1.001" {
		t.Errorf("expected ask 1.001, got %s", rate.Ask)
	}
	if rate.Strategy != "topN" {
		t.Errorf("expected strategy topN, got %s", rate.Strategy)
	}
	if rate.FetchedAt != now {
		t.Error("fetched_at mismatch")
	}
}
