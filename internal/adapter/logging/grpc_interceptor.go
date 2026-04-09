package logging

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor that logs
// incoming requests with method name, duration, and status.
func UnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		status := "ok"
		if err != nil {
			status = "error"
		}

		logger.Info("gRPC request",
			Field("method", info.FullMethod),
			Field("status", status),
			zap.Duration("duration", duration),
		)

		if err != nil {
			logger.Error("gRPC request failed",
				Field("method", info.FullMethod),
				zap.Error(err),
			)
		}

		return resp, err
	}
}
