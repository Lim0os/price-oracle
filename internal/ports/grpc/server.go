// Package grpcserver implements the gRPC server and handlers for the price oracle.
package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/Lim0os/price-oracle/internal/app/query"
	"github.com/Lim0os/price-oracle/internal/domain"
	ratesv1 "github.com/Lim0os/price-oracle/proto/rates/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server handles gRPC requests for rate fetching and health checks.
type Server struct {
	ratesv1.UnimplementedRatesServiceServer
	getRatesHandler *query.GetRatesHandler
	dbHealthChecker domain.HealthChecker
	logger          *zap.Logger
	grpcServer      *grpc.Server
	healthServer    *health.Server
	metricsServer   *http.Server
	metricsHandler  http.Handler
}

// NewServer creates a gRPC server with the given dependencies.
func NewServer(
	getRatesHandler *query.GetRatesHandler,
	dbHealthChecker domain.HealthChecker,
	logger *zap.Logger,
) *Server {
	return &Server{
		getRatesHandler: getRatesHandler,
		dbHealthChecker: dbHealthChecker,
		logger:          logger,
		healthServer:    health.NewServer(),
	}
}

// Start starts the gRPC server on the given address.
func (s *Server) Start(addr string, opts ...grpc.ServerOption) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.grpcServer = grpc.NewServer(opts...)
	ratesv1.RegisterRatesServiceServer(s.grpcServer, s)
	healthpb.RegisterHealthServer(s.grpcServer, s.healthServer)
	reflection.Register(s.grpcServer)

	s.logger.Info("gRPC server started", zap.String("address", addr))

	if err := s.grpcServer.Serve(lis); err != nil {
		if !errors.Is(err, grpc.ErrServerStopped) {
			return fmt.Errorf("gRPC server error: %w", err)
		}
	}

	return nil
}

// Stop gracefully stops the gRPC server and metrics server.
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		s.logger.Info("gRPC server stopped gracefully")
	}
	if s.healthServer != nil {
		s.healthServer.Shutdown()
	}
	if s.metricsServer != nil {
		_ = s.metricsServer.Close()
		s.logger.Info("metrics server stopped")
	}
}

// StartMetrics starts the HTTP Prometheus metrics server.
func (s *Server) StartMetrics(addr string) error {
	if s.metricsHandler == nil {
		return fmt.Errorf("metrics handler not set")
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on metrics address %s: %w", addr, err)
	}

	s.metricsServer = &http.Server{Handler: s.metricsHandler}
	s.logger.Info("metrics server started", zap.String("address", addr))

	go func() {
		if err := s.metricsServer.Serve(lis); err != nil && err != http.ErrServerClosed {
			s.logger.Error("metrics server error", zap.Error(err))
		}
	}()

	return nil
}

// SetMetricsHandler sets the Prometheus metrics HTTP handler.
func (s *Server) SetMetricsHandler(handler http.Handler) {
	s.metricsHandler = handler
}

// GetRates calculates and returns rates based on the requested strategy.
func (s *Server) GetRates(ctx context.Context, req *ratesv1.GetRatesRequest) (*ratesv1.GetRatesResponse, error) {
	queryReq := query.GetRatesRequest{
		Strategy: req.GetStrategy(),
		N:        int(req.GetN()),
		M:        int(req.GetM()),
	}

	resp, err := s.getRatesHandler.Handle(ctx, queryReq)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &ratesv1.GetRatesResponse{
		Ask:       resp.Ask,
		Bid:       resp.Bid,
		FetchedAt: timestamppb.New(resp.FetchedAt),
	}, nil
}

// Healthcheck returns the health status of service dependencies.
func (s *Server) Healthcheck(ctx context.Context, _ *ratesv1.HealthcheckRequest) (*ratesv1.HealthcheckResponse, error) {
	if s.dbHealthChecker != nil {
		if err := s.dbHealthChecker.Ping(ctx); err != nil {
			if s.logger != nil {
				s.logger.Warn("healthcheck: database is not available", zap.Error(err))
			}
			return &ratesv1.HealthcheckResponse{Status: "NOT_SERVING"}, nil
		}
	}

	return &ratesv1.HealthcheckResponse{Status: "SERVING"}, nil
}

func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, domain.ErrInvalidStrategy):
		return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid strategy: %v", err))
	case errors.Is(err, domain.ErrEmptyOrderBook):
		return status.Error(codes.FailedPrecondition, fmt.Sprintf("empty order book: %v", err))
	case errors.Is(err, domain.ErrIndexOutOfBounds):
		return status.Error(codes.OutOfRange, fmt.Sprintf("index out of range: %v", err))
	case errors.Is(err, domain.ErrInvalidRange):
		return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid range: %v", err))
	default:
		return status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err))
	}
}
