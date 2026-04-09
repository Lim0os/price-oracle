// Package telemetry provides OpenTelemetry initialization, tracing, and Prometheus metrics.
package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a gRPC interceptor that creates a tracing span for each request.
func UnaryServerInterceptor(tracer *Tracer) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return resp, err
	}
}
