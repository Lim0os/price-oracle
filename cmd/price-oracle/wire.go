package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Lim0os/price-oracle/internal/adapter/exchange"
	loggingadapter "github.com/Lim0os/price-oracle/internal/adapter/logging"
	"github.com/Lim0os/price-oracle/internal/adapter/persistence"
	"github.com/Lim0os/price-oracle/internal/adapter/telemetry"
	"github.com/Lim0os/price-oracle/internal/app/query"
	"github.com/Lim0os/price-oracle/internal/config"
	grpcserver "github.com/Lim0os/price-oracle/internal/ports/grpc"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var configModule = fx.Module("config",
	fx.Provide(config.Load),
)

var metricsModule = fx.Module("metrics",
	fx.Provide(func() (*prometheus.Registry, *telemetry.Metrics) {
		registry := prometheus.NewRegistry()
		m := telemetry.NewMetrics(registry)
		return registry, m
	}),
)

var loggingModule = fx.Module("logging",
	fx.Provide(
		func(cfg *config.Config) (*zap.Logger, error) {
			return loggingadapter.NewLogger(cfg.LogLevel)
		},
	),
)

type TelemetryResult struct {
	fx.Out
	Tracer         *telemetry.Tracer
	TracerProvider *telemetry.TracerProviderWrapper
}

var telemetryModule = fx.Module("telemetry",
	fx.Provide(
		func(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) (TelemetryResult, error) {
			var result TelemetryResult
			if !cfg.OTELEnabled || cfg.OTELEndpoint == "" {
				log.Info("telemetry disabled, using no-op tracer")
				result.Tracer = telemetry.NewTracer()
				return result, nil
			}
			tp, err := telemetry.InitTracerProvider(context.Background(), cfg.OTELEndpoint, "price-oracle", cfg.OTELInsecure)
			if err != nil {
				log.Warn("failed to initialize telemetry, running without it", zap.Error(err))
				result.Tracer = telemetry.NewTracer()
				return result, nil
			}
			result.Tracer = telemetry.NewTracer()
			result.TracerProvider = &telemetry.TracerProviderWrapper{Provider: tp}
			lc.Append(fx.Hook{
				OnStop: func(_ context.Context) error {
					log.Info("shutting down telemetry")
					return telemetry.Shutdown(context.Background(), tp, 10)
				},
			})
			log.Info("telemetry initialized")
			return result, nil
		},
	),
)

type DatabaseResult struct {
	fx.Out
	Pool            persistence.PgxPool
	DBHealthChecker *persistence.DBHealthChecker
}

var databaseModule = fx.Module("database",
	fx.Provide(
		func(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) (DatabaseResult, error) {
			dbCfg := persistence.DBConfig{
				Host: cfg.DBHost, Port: cfg.DBPort, DBName: cfg.DBName,
				User: cfg.DBUser, Password: cfg.DBPassword, SSLMode: cfg.DBSSLMode,
			}
			pool, err := persistence.NewPool(context.Background(), dbCfg)
			if err != nil {
				return DatabaseResult{}, fmt.Errorf("initialize database: %w", err)
			}
			log.Info("database connected",
				loggingadapter.Field("host", cfg.DBHost),
				loggingadapter.Field("port", cfg.DBPort),
				loggingadapter.Field("db", cfg.DBName),
			)
			lc.Append(fx.Hook{
				OnStop: func(_ context.Context) error {
					log.Info("closing database connection")
					pool.Close()
					return nil
				},
			})
			return DatabaseResult{
				Pool:            pool,
				DBHealthChecker: persistence.NewDBHealthChecker(pool),
			}, nil
		},
	),
)

var exchangeModule = fx.Module("exchange",
	fx.Provide(
		func(cfg *config.Config) *exchange.GrinexClient {
			return exchange.NewGrinexClient(cfg.GrinexBaseURL, cfg.HTTPTimeout)
		},
	),
)

var persistenceModule = fx.Module("persistence",
	fx.Provide(
		func(db DatabaseResult) *persistence.RateRepository {
			return persistence.NewRateRepository(db.Pool)
		},
	),
)

var appModule = fx.Module("app",
	fx.Provide(
		func(
			fetcher *exchange.GrinexClient,
			repo *persistence.RateRepository,
			cfg *config.Config,
			tracer TelemetryResult,
		) *query.GetRatesHandler {
			return query.NewGetRatesHandler(fetcher, repo, cfg.GrinexSymbol, tracer.Tracer)
		},
	),
)

type GRPCServerDeps struct {
	fx.In
	GetRatesHandler *query.GetRatesHandler
	DBHealthChecker *persistence.DBHealthChecker
	Logger          *zap.Logger
}

var grpcModule = fx.Module("grpc",
	fx.Provide(func(deps GRPCServerDeps) *grpcserver.Server {
		return grpcserver.NewServer(deps.GetRatesHandler, deps.DBHealthChecker, deps.Logger)
	}),
)

type RunServerDeps struct {
	fx.In
	Server     *grpcserver.Server
	Cfg        *config.Config
	Logger     *zap.Logger
	Tracer     *telemetry.Tracer
	Shutdowner fx.Shutdowner
}

func runGRPCServer(deps RunServerDeps) error {
	addr := fmt.Sprintf("%s:%d", deps.Cfg.GRPCHost, deps.Cfg.GRPCPort)
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		loggingadapter.UnaryServerInterceptor(deps.Logger),
		telemetry.UnaryServerInterceptor(deps.Tracer),
	}
	go func() {
		opts := []grpc.ServerOption{grpc.ChainUnaryInterceptor(unaryInterceptors...)}
		if err := deps.Server.Start(addr, opts...); err != nil {
			deps.Logger.Error("gRPC server error", zap.Error(err))
		}
	}()
	deps.Logger.Info("gRPC server started", loggingadapter.Field("address", addr))
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		deps.Logger.Info("shutdown signal received", zap.String("signal", sig.String()))
		_ = deps.Shutdowner.Shutdown()
	}()
	return nil
}

type MetricsServerDeps struct {
	fx.In
	Server   *grpcserver.Server
	Registry *prometheus.Registry
	Logger   *zap.Logger
}

func runMetricsServer(deps MetricsServerDeps) error {
	metricsAddr := "0.0.0.0:9090"
	deps.Server.SetMetricsHandler(telemetry.MetricsHandler(deps.Registry))
	if err := deps.Server.StartMetrics(metricsAddr); err != nil {
		deps.Logger.Warn("metrics server failed to start", zap.Error(err))
		return nil
	}
	return nil
}
