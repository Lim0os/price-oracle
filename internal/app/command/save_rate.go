// Package command provides write use cases (commands) for the application layer.
package command

import (
	"context"

	"github.com/Lim0os/price-oracle/internal/adapter/telemetry"
	"github.com/Lim0os/price-oracle/internal/domain"
	"go.opentelemetry.io/otel/codes"
)

// SaveRateCommand holds the rate data to persist.
type SaveRateCommand struct {
	Rate *domain.Rate
}

// SaveRateHandler persists a calculated rate.
type SaveRateHandler struct {
	repo   domain.RateRepository
	tracer *telemetry.Tracer
}

// NewSaveRateHandler creates a new SaveRateHandler.
func NewSaveRateHandler(repo domain.RateRepository, tracer *telemetry.Tracer) *SaveRateHandler {
	return &SaveRateHandler{
		repo:   repo,
		tracer: tracer,
	}
}

// Handle persists the rate to the database.
func (h *SaveRateHandler) Handle(ctx context.Context, cmd SaveRateCommand) error {
	ctx, span := h.tracer.StartSpan(ctx, "SaveRateHandler.Handle")
	defer span.End()

	if err := h.repo.Save(ctx, cmd.Rate); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}
