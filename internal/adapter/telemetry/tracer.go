package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracerName is the name of the application's OpenTelemetry tracer.
const TracerName = "github.com/Lim0os/price-oracle"

// Tracer wraps the OTel tracer with convenience methods.
type Tracer struct {
	trace.Tracer
}

// NewTracer creates a new application tracer.
func NewTracer() *Tracer {
	return &Tracer{
		Tracer: otel.Tracer(TracerName),
	}
}

// StartSpan starts a new span with optional attributes.
func (t *Tracer) StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	var options []trace.SpanStartOption
	if len(attrs) > 0 {
		options = append(options, trace.WithAttributes(attrs...))
	}
	return t.Start(ctx, name, options...)
}

// Common attribute keys for this application.
const (
	AttrStrategy = attribute.Key("app.strategy")
	AttrSymbol   = attribute.Key("app.symbol")
)
