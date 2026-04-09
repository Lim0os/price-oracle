package grpcserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Lim0os/price-oracle/internal/adapter/telemetry"
	"github.com/Lim0os/price-oracle/internal/app/query"
	"github.com/Lim0os/price-oracle/internal/domain"
	ratesv1 "github.com/Lim0os/price-oracle/proto/rates/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockHealthChecker is a test double for domain.HealthChecker.
type mockHealthChecker struct {
	err error
}

func (m *mockHealthChecker) Ping(_ context.Context) error {
	return m.err
}

func TestServer_GetRates_Success(t *testing.T) {
	fetcher := &mockOrderBookFetcher{
		book: &domain.OrderBook{
			Asks: []domain.OrderBookEntry{{Price: "1.001", Quantity: "10"}},
			Bids: []domain.OrderBookEntry{{Price: "0.999", Quantity: "10"}},
		},
	}
	repo := &mockRateRepo{}
	tracer := telemetry.NewTracer()

	handler := query.NewGetRatesHandler(fetcher, repo, "usdta7a5", tracer)
	srv := NewServer(handler, nil, nil)

	resp, err := srv.GetRates(context.Background(), &ratesv1.GetRatesRequest{Strategy: "topN", N: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Ask != "1.001" {
		t.Errorf("expected ask 1.001, got %s", resp.Ask)
	}
	if resp.Bid != "0.999" {
		t.Errorf("expected bid 0.999, got %s", resp.Bid)
	}
}

func TestServer_GetRates_InvalidStrategy(t *testing.T) {
	// Create handler that will fail with invalid strategy
	fetcher := &mockOrderBookFetcher{
		book: &domain.OrderBook{
			Asks: []domain.OrderBookEntry{{Price: "1.001", Quantity: "10"}},
			Bids: []domain.OrderBookEntry{{Price: "0.999", Quantity: "10"}},
		},
	}
	repo := &mockRateRepo{}
	tracer := telemetry.NewTracer()

	handler := query.NewGetRatesHandler(fetcher, repo, "usdta7a5", tracer)
	srv := NewServer(handler, nil, nil)

	_, err := srv.GetRates(context.Background(), &ratesv1.GetRatesRequest{Strategy: "unknown"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status, got %v", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %s", st.Code())
	}
}

func TestServer_GetRates_ExchangeError(t *testing.T) {
	fetcher := &mockOrderBookFetcher{err: errors.New("exchange error")}
	repo := &mockRateRepo{}
	tracer := telemetry.NewTracer()

	handler := query.NewGetRatesHandler(fetcher, repo, "usdta7a5", tracer)
	srv := NewServer(handler, nil, nil)

	_, err := srv.GetRates(context.Background(), &ratesv1.GetRatesRequest{Strategy: "topN", N: 1})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status, got %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected Internal, got %s", st.Code())
	}
}

func TestServer_Healthcheck_Serving(t *testing.T) {
	healthChecker := &mockHealthChecker{}
	srv := NewServer(nil, healthChecker, nil)

	resp, err := srv.Healthcheck(context.Background(), &ratesv1.HealthcheckRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "SERVING" {
		t.Errorf("expected SERVING, got %s", resp.Status)
	}
}

func TestServer_Healthcheck_NotServing(t *testing.T) {
	healthChecker := &mockHealthChecker{
		err: errors.New("database down"),
	}
	srv := NewServer(nil, healthChecker, nil)

	resp, err := srv.Healthcheck(context.Background(), &ratesv1.HealthcheckRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "NOT_SERVING" {
		t.Errorf("expected NOT_SERVING, got %s", resp.Status)
	}
}

func TestToGRPCError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode codes.Code
	}{
		{"nil error", nil, codes.OK},
		{"ErrInvalidStrategy", domain.ErrInvalidStrategy, codes.InvalidArgument},
		{"ErrEmptyOrderBook", domain.ErrEmptyOrderBook, codes.FailedPrecondition},
		{"ErrIndexOutOfBounds", domain.ErrIndexOutOfBounds, codes.OutOfRange},
		{"ErrInvalidRange", domain.ErrInvalidRange, codes.InvalidArgument},
		{"generic error", errors.New("something broke"), codes.Internal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := toGRPCError(tt.err)
			if tt.err == nil {
				if err != nil {
					t.Errorf("expected nil, got %v", err)
				}
				return
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected gRPC status, got %v", err)
			}
			if st.Code() != tt.expectedCode {
				t.Errorf("expected %s, got %s", tt.expectedCode, st.Code())
			}
		})
	}
}

// ─── Test doubles ────────────────────────────────────────────────────────────

type mockOrderBookFetcher struct {
	book *domain.OrderBook
	err  error
}

func (m *mockOrderBookFetcher) FetchOrderBook(_ context.Context, _ string) (*domain.OrderBook, error) {
	return m.book, m.err
}

type mockRateRepo struct {
	saveErr error
}

func (m *mockRateRepo) Save(_ context.Context, _ *domain.Rate) error {
	return m.saveErr
}

func TestNewRate_FetchedAtSet(t *testing.T) {
	now := time.Now()
	rate := domain.NewRate("1.0", "0.9", "topN", 1, 0, now)

	if rate.FetchedAt != now {
		t.Errorf("expected fetched_at %v, got %v", now, rate.FetchedAt)
	}
}
