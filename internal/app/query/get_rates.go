// Package query provides read use cases (queries) for the application layer.
package query

import (
	"context"
	"time"

	"github.com/Lim0os/price-oracle/internal/adapter/telemetry"
	"github.com/Lim0os/price-oracle/internal/app/command"
	"github.com/Lim0os/price-oracle/internal/domain"
	"go.opentelemetry.io/otel/codes"
)

// GetRatesRequest holds the parameters for rate calculation.
type GetRatesRequest struct {
	Strategy string
	N        int
	M        int
}

// GetRatesResponse holds the calculated rate result.
type GetRatesResponse struct {
	Ask       string
	Bid       string
	FetchedAt time.Time
}

// GetRatesHandler orchestrates rate fetching, calculation, and persistence.
type GetRatesHandler struct {
	fetcher domain.OrderBookFetcher
	repo    domain.RateRepository
	symbol  string
	tracer  *telemetry.Tracer
}

// NewGetRatesHandler creates a new GetRatesHandler.
func NewGetRatesHandler(fetcher domain.OrderBookFetcher, repo domain.RateRepository, symbol string, tracer *telemetry.Tracer) *GetRatesHandler {
	return &GetRatesHandler{
		fetcher: fetcher,
		repo:    repo,
		symbol:  symbol,
		tracer:  tracer,
	}
}

// Handle executes the full rate calculation pipeline.
func (h *GetRatesHandler) Handle(ctx context.Context, req GetRatesRequest) (*GetRatesResponse, error) {
	ctx, span := h.tracer.StartSpan(ctx, "GetRatesHandler.Handle",
		telemetry.AttrStrategy.String(req.Strategy),
		telemetry.AttrSymbol.String(h.symbol),
	)
	defer span.End()

	calculator, err := h.createCalculator(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	book, err := h.fetcher.FetchOrderBook(ctx, h.symbol)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "fetch order book: "+err.Error())
		return nil, err
	}

	ask, bid, err := calculator.Calculate(book)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "calculate rates: "+err.Error())
		return nil, err
	}

	fetchedAt := time.Now()
	rate := domain.NewRate(ask, bid, req.Strategy, req.N, req.M, fetchedAt)

	saveCmd := command.SaveRateCommand{Rate: rate}
	if err := command.NewSaveRateHandler(h.repo, h.tracer).Handle(ctx, saveCmd); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "save rate: "+err.Error())
		return nil, err
	}

	return &GetRatesResponse{
		Ask:       ask,
		Bid:       bid,
		FetchedAt: fetchedAt,
	}, nil
}

func (h *GetRatesHandler) createCalculator(req GetRatesRequest) (domain.StrategyCalculator, error) {
	switch req.Strategy {
	case "topN":
		return domain.NewTopNCalculator(req.N), nil
	case "avgNM":
		return domain.NewAvgNMCalculator(req.N, req.M), nil
	default:
		return nil, domain.ErrInvalidStrategy
	}
}
