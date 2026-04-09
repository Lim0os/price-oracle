// Package exchange provides an HTTP client for the Grinex exchange API.
package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Lim0os/price-oracle/internal/domain"
	"github.com/go-resty/resty/v2"
)

// GrinexClient implements domain.OrderBookFetcher using the Grinex REST API.
type GrinexClient struct {
	httpClient *resty.Client
	baseURL    string
}

// NewGrinexClient creates a new Grinex API client.
func NewGrinexClient(baseURL string, timeout time.Duration) *GrinexClient {
	return &GrinexClient{
		httpClient: resty.New().
			SetBaseURL(baseURL).
			SetTimeout(timeout).
			SetHeader("Content-Type", "application/json"),
		baseURL: baseURL,
	}
}

// FetchOrderBook retrieves the order book for the given symbol from Grinex API.
// It uses the /api/v1/spot/depth endpoint.
func (c *GrinexClient) FetchOrderBook(ctx context.Context, symbol string) (*domain.OrderBook, error) {
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetQueryParam("symbol", symbol).
		Get("/api/v1/spot/depth")
	if err != nil {
		return nil, fmt.Errorf("fetch order book: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("fetch order book: unexpected status code: %d, body: %s", resp.StatusCode(), resp.Body())
	}

	var depthResp DepthResponse
	if err := json.Unmarshal(resp.Body(), &depthResp); err != nil {
		return nil, fmt.Errorf("parse order book response: %w", err)
	}

	book := &domain.OrderBook{
		Asks: make([]domain.OrderBookEntry, 0, len(depthResp.Asks)),
		Bids: make([]domain.OrderBookEntry, 0, len(depthResp.Bids)),
	}

	for _, entry := range depthResp.Asks {
		book.Asks = append(book.Asks, domain.OrderBookEntry{
			Price:    entry.Price,
			Quantity: entry.Quantity,
		})
	}
	for _, entry := range depthResp.Bids {
		book.Bids = append(book.Bids, domain.OrderBookEntry{
			Price:    entry.Price,
			Quantity: entry.Quantity,
		})
	}

	return book, nil
}

// DepthResponse represents the JSON structure of Grinex depth API response.
type DepthResponse struct {
	Asks []DepthEntry `json:"asks"`
	Bids []DepthEntry `json:"bids"`
}

// DepthEntry represents a single price level in the order book.
type DepthEntry struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
}
