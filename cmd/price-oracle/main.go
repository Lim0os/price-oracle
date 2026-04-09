// Package main is the entry point for the price-oracle gRPC service.
package main

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		configModule,
		metricsModule,
		loggingModule,
		telemetryModule,
		databaseModule,
		exchangeModule,
		persistenceModule,
		appModule,
		grpcModule,
		fx.Invoke(runGRPCServer),
		fx.Invoke(runMetricsServer),
	)
	app.Run()
}
