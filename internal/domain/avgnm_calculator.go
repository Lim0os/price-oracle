// Package domain contains business logic, value objects, and port interfaces.
package domain

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// AvgNMCalculator returns the arithmetic average of prices in range [N, M].
type AvgNMCalculator struct {
	N int
	M int
}

// NewAvgNMCalculator creates a calculator for the [N, M] average.
func NewAvgNMCalculator(n, m int) *AvgNMCalculator {
	return &AvgNMCalculator{N: n, M: m}
}

// Calculate returns average ask and bid over the [N, M] range.
func (c *AvgNMCalculator) Calculate(book *OrderBook) (ask, bid string, err error) {
	if err := validateOrderBook(book); err != nil {
		return "", "", err
	}

	if c.N < 1 || c.M < 1 {
		return "", "", fmt.Errorf("avgNM: %w: n and m must be >= 1", ErrInvalidRange)
	}
	if c.N > c.M {
		return "", "", fmt.Errorf("avgNM: %w: n=%d > m=%d", ErrInvalidRange, c.N, c.M)
	}

	askAvg, err := c.averagePrice(book.Asks)
	if err != nil {
		return "", "", fmt.Errorf("avgNM asks: %w", err)
	}
	bidAvg, err := c.averagePrice(book.Bids)
	if err != nil {
		return "", "", fmt.Errorf("avgNM bids: %w", err)
	}

	return askAvg.String(), bidAvg.String(), nil
}

func (c *AvgNMCalculator) averagePrice(entries []OrderBookEntry) (decimal.Decimal, error) {
	if len(entries) == 0 {
		return decimal.Zero, ErrEmptyOrderBook
	}

	startIdx := c.N - 1
	endIdx := c.M - 1

	if endIdx >= len(entries) {
		return decimal.Zero, fmt.Errorf("%w: range [%d, %d], entries len %d", ErrIndexOutOfBounds, c.N, c.M, len(entries))
	}

	sum := decimal.Zero
	count := 0
	for i := startIdx; i <= endIdx; i++ {
		price, err := parseDecimal(entries[i].Price)
		if err != nil {
			return decimal.Zero, err
		}
		sum = sum.Add(price)
		count++
	}

	if count == 0 {
		return decimal.Zero, ErrEmptyOrderBook
	}

	return sum.Div(decimal.NewFromInt(int64(count))), nil
}
