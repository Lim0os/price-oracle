package exchange

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lim0os/price-oracle/internal/domain"
)

func TestGrinexClient_FetchOrderBook_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/spot/depth" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "usdta7a5" {
			t.Fatalf("unexpected symbol: %s", r.URL.Query().Get("symbol"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"asks":[{"price":"1.001","quantity":"10"},{"price":"1.002","quantity":"20"}],"bids":[{"price":"0.999","quantity":"10"},{"price":"0.998","quantity":"20"}]}`))
	}))
	defer server.Close()

	client := NewGrinexClient(server.URL, 5*time.Second)
	book, err := client.FetchOrderBook(context.Background(), "usdta7a5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(book.Asks) != 2 {
		t.Fatalf("expected 2 asks, got %d", len(book.Asks))
	}
	if len(book.Bids) != 2 {
		t.Fatalf("expected 2 bids, got %d", len(book.Bids))
	}

	if book.Asks[0].Price != "1.001" {
		t.Errorf("expected ask[0].price 1.001, got %s", book.Asks[0].Price)
	}
	if book.Bids[0].Price != "0.999" {
		t.Errorf("expected bid[0].price 0.999, got %s", book.Bids[0].Price)
	}
}

func TestGrinexClient_FetchOrderBook_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := NewGrinexClient(server.URL, 5*time.Second)
	_, err := client.FetchOrderBook(context.Background(), "usdta7a5")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestGrinexClient_FetchOrderBook_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	client := NewGrinexClient(server.URL, 5*time.Second)
	_, err := client.FetchOrderBook(context.Background(), "usdta7a5")
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestGrinexClient_FetchOrderBook_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	client := NewGrinexClient(server.URL, 1*time.Millisecond)
	_, err := client.FetchOrderBook(context.Background(), "usdta7a5")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestGrinexClient_FetchOrderBook_CtxCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	client := NewGrinexClient(server.URL, 5*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.FetchOrderBook(ctx, "usdta7a5")
	if err == nil {
		t.Fatal("expected error for canceled context")
	}
}

func TestGrinexClient_FetchOrderBook_ImplementsInterface(_ *testing.T) {
	// Compile-time check: GrinexClient must implement domain.OrderBookFetcher
	var _ domain.OrderBookFetcher = (*GrinexClient)(nil)
}
