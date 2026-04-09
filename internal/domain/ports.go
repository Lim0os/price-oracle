package domain

import "context"

// OrderBookFetcher fetches order book data from an exchange.
type OrderBookFetcher interface {
	FetchOrderBook(ctx context.Context, symbol string) (*OrderBook, error)
}

// RateRepository persists rate data.
type RateRepository interface {
	Save(ctx context.Context, rate *Rate) error
}

// HealthChecker checks dependency health.
type HealthChecker interface {
	Ping(ctx context.Context) error
}
