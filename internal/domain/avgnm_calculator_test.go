package domain

import (
	"errors"
	"testing"
)

func TestAvgNMCalculator_Success(t *testing.T) {
	calc := NewAvgNMCalculator(1, 3)
	book := &OrderBook{
		Asks: []OrderBookEntry{{Price: "1.0", Quantity: "10"}, {Price: "2.0", Quantity: "20"}, {Price: "3.0", Quantity: "30"}},
		Bids: []OrderBookEntry{{Price: "0.9", Quantity: "10"}, {Price: "1.9", Quantity: "20"}, {Price: "2.9", Quantity: "30"}},
	}

	ask, bid, err := calc.Calculate(book)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// avg(1.0, 2.0, 3.0) = 2.0
	if ask != "2" {
		t.Errorf("expected ask 2, got %s", ask)
	}
	// avg(0.9, 1.9, 2.9) = 1.9
	if bid != "1.9" {
		t.Errorf("expected bid 1.9, got %s", bid)
	}
}

func TestAvgNMCalculator_RangeSubset(t *testing.T) {
	calc := NewAvgNMCalculator(2, 4)
	book := &OrderBook{
		Asks: []OrderBookEntry{{Price: "1", Quantity: "10"}, {Price: "2", Quantity: "20"}, {Price: "3", Quantity: "30"}, {Price: "4", Quantity: "40"}},
		Bids: []OrderBookEntry{{Price: "1", Quantity: "10"}, {Price: "2", Quantity: "20"}, {Price: "3", Quantity: "30"}, {Price: "4", Quantity: "40"}},
	}

	ask, bid, err := calc.Calculate(book)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// avg(2, 3, 4) = 3
	if ask != "3" {
		t.Errorf("expected ask 3, got %s", ask)
	}
	if bid != "3" {
		t.Errorf("expected bid 3, got %s", bid)
	}
}

func TestAvgNMCalculator_NGreaterThanM(t *testing.T) {
	calc := NewAvgNMCalculator(5, 2)
	book := &OrderBook{
		Asks: []OrderBookEntry{{Price: "1", Quantity: "10"}, {Price: "2", Quantity: "20"}},
		Bids: []OrderBookEntry{{Price: "1", Quantity: "10"}, {Price: "2", Quantity: "20"}},
	}

	_, _, err := calc.Calculate(book)
	if !errors.Is(err, ErrInvalidRange) {
		t.Errorf("expected ErrInvalidRange, got %v", err)
	}
}

func TestAvgNMCalculator_IndexOutOfBounds(t *testing.T) {
	calc := NewAvgNMCalculator(1, 5)
	book := &OrderBook{
		Asks: []OrderBookEntry{{Price: "1", Quantity: "10"}, {Price: "2", Quantity: "20"}},
		Bids: []OrderBookEntry{{Price: "1", Quantity: "10"}, {Price: "2", Quantity: "20"}},
	}

	_, _, err := calc.Calculate(book)
	if !errors.Is(err, ErrIndexOutOfBounds) {
		t.Errorf("expected ErrIndexOutOfBounds, got %v", err)
	}
}

func TestAvgNMCalculator_NLessThanOne(t *testing.T) {
	calc := NewAvgNMCalculator(0, 2)
	book := &OrderBook{
		Asks: []OrderBookEntry{{Price: "1", Quantity: "10"}, {Price: "2", Quantity: "20"}},
		Bids: []OrderBookEntry{{Price: "1", Quantity: "10"}, {Price: "2", Quantity: "20"}},
	}

	_, _, err := calc.Calculate(book)
	if !errors.Is(err, ErrInvalidRange) {
		t.Errorf("expected ErrInvalidRange, got %v", err)
	}
}

func TestAvgNMCalculator_EmptyOrderBook(t *testing.T) {
	calc := NewAvgNMCalculator(1, 1)
	book := &OrderBook{Asks: []OrderBookEntry{}, Bids: []OrderBookEntry{}}

	_, _, err := calc.Calculate(book)
	if !errors.Is(err, ErrEmptyOrderBook) {
		t.Errorf("expected ErrEmptyOrderBook, got %v", err)
	}
}

func TestAvgNMCalculator_NilOrderBook(t *testing.T) {
	calc := NewAvgNMCalculator(1, 1)

	_, _, err := calc.Calculate(nil)
	if !errors.Is(err, ErrEmptyOrderBook) {
		t.Errorf("expected ErrEmptyOrderBook, got %v", err)
	}
}

func TestAvgNMCalculator_SingleElement(t *testing.T) {
	calc := NewAvgNMCalculator(1, 1)
	book := &OrderBook{
		Asks: []OrderBookEntry{{Price: "5.5", Quantity: "10"}},
		Bids: []OrderBookEntry{{Price: "4.5", Quantity: "10"}},
	}

	ask, bid, err := calc.Calculate(book)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ask != "5.5" {
		t.Errorf("expected ask 5.5, got %s", ask)
	}
	if bid != "4.5" {
		t.Errorf("expected bid 4.5, got %s", bid)
	}
}
