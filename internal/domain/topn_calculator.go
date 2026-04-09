package domain

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

// TopNCalculator returns the price from the N-th position (1-based index).
type TopNCalculator struct {
	N int
}

// NewTopNCalculator creates a calculator returning the N-th price (1-based).
func NewTopNCalculator(n int) *TopNCalculator {
	return &TopNCalculator{N: n}
}

// Calculate returns ask and bid from the N-th position.
func (c *TopNCalculator) Calculate(book *OrderBook) (ask, bid string, err error) {
	if err := validateOrderBook(book); err != nil {
		return "", "", err
	}

	if c.N < 1 {
		return "", "", fmt.Errorf("topN: %w: n must be >= 1", ErrIndexOutOfBounds)
	}

	idx := c.N - 1

	if idx >= len(book.Asks) {
		return "", "", fmt.Errorf("topN asks: %w: index %d, asks len %d", ErrIndexOutOfBounds, c.N, len(book.Asks))
	}
	if idx >= len(book.Bids) {
		return "", "", fmt.Errorf("topN bids: %w: index %d, bids len %d", ErrIndexOutOfBounds, c.N, len(book.Bids))
	}

	askPrice, err := parseDecimal(book.Asks[idx].Price)
	if err != nil {
		return "", "", fmt.Errorf("topN asks: %w", err)
	}
	bidPrice, err := parseDecimal(book.Bids[idx].Price)
	if err != nil {
		return "", "", fmt.Errorf("topN bids: %w", err)
	}

	return askPrice.String(), bidPrice.String(), nil
}

func parseDecimal(s string) (decimal.Decimal, error) {
	d, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to parse decimal value: %w", err)
	}
	return decimal.NewFromFloat(d), nil
}
