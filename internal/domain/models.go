package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// OrderBookEntry is a single price level in the order book.
type OrderBookEntry struct {
	Price    string
	Quantity string
}

// OrderBook holds asks and bids from the exchange.
type OrderBook struct {
	Asks []OrderBookEntry
	Bids []OrderBookEntry
}

// Rate is the calculated result with strategy parameters.
type Rate struct {
	ID        uuid.UUID
	Ask       string
	Bid       string
	Strategy  string
	N         int
	M         int
	FetchedAt time.Time
}

// NewRate creates a Rate with a generated UUID v7.
func NewRate(ask, bid, strategy string, n, m int, fetchedAt time.Time) *Rate {
	id, _ := uuid.NewV7()
	return &Rate{
		ID:        id,
		Ask:       ask,
		Bid:       bid,
		Strategy:  strategy,
		N:         n,
		M:         m,
		FetchedAt: fetchedAt,
	}
}
