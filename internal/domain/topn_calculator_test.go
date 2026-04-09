package domain

import (
	"errors"
	"testing"
)

func TestTopNCalculator_Success(t *testing.T) {
	calc := NewTopNCalculator(1)
	book := &OrderBook{
		Asks: []OrderBookEntry{{Price: "1.001", Quantity: "10"}, {Price: "1.002", Quantity: "20"}},
		Bids: []OrderBookEntry{{Price: "0.999", Quantity: "10"}, {Price: "0.998", Quantity: "20"}},
	}

	ask, bid, err := calc.Calculate(book)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ask != "1.001" {
		t.Errorf("expected ask 1.001, got %s", ask)
	}
	if bid != "0.999" {
		t.Errorf("expected bid 0.999, got %s", bid)
	}
}

func TestTopNCalculator_NthPosition(t *testing.T) {
	calc := NewTopNCalculator(2)
	book := &OrderBook{
		Asks: []OrderBookEntry{{Price: "1.001", Quantity: "10"}, {Price: "1.005", Quantity: "20"}},
		Bids: []OrderBookEntry{{Price: "0.999", Quantity: "10"}, {Price: "0.995", Quantity: "20"}},
	}

	ask, bid, err := calc.Calculate(book)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ask != "1.005" {
		t.Errorf("expected ask 1.005, got %s", ask)
	}
	if bid != "0.995" {
		t.Errorf("expected bid 0.995, got %s", bid)
	}
}

func TestTopNCalculator_IndexOutOfBounds(t *testing.T) {
	calc := NewTopNCalculator(5)
	book := &OrderBook{
		Asks: []OrderBookEntry{{Price: "1.001", Quantity: "10"}},
		Bids: []OrderBookEntry{{Price: "0.999", Quantity: "10"}},
	}

	_, _, err := calc.Calculate(book)
	if !errors.Is(err, ErrIndexOutOfBounds) {
		t.Errorf("expected ErrIndexOutOfBounds, got %v", err)
	}
}

func TestTopNCalculator_NLessThanOne(t *testing.T) {
	calc := NewTopNCalculator(0)
	book := &OrderBook{
		Asks: []OrderBookEntry{{Price: "1.001", Quantity: "10"}},
		Bids: []OrderBookEntry{{Price: "0.999", Quantity: "10"}},
	}

	_, _, err := calc.Calculate(book)
	if err == nil {
		t.Fatal("expected error for n < 1")
	}
}

func TestTopNCalculator_EmptyOrderBook(t *testing.T) {
	calc := NewTopNCalculator(1)
	book := &OrderBook{Asks: []OrderBookEntry{}, Bids: []OrderBookEntry{}}

	_, _, err := calc.Calculate(book)
	if !errors.Is(err, ErrEmptyOrderBook) {
		t.Errorf("expected ErrEmptyOrderBook, got %v", err)
	}
}

func TestTopNCalculator_NilOrderBook(t *testing.T) {
	calc := NewTopNCalculator(1)

	_, _, err := calc.Calculate(nil)
	if !errors.Is(err, ErrEmptyOrderBook) {
		t.Errorf("expected ErrEmptyOrderBook, got %v", err)
	}
}
