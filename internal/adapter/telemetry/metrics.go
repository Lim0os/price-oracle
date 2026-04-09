package telemetry

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds Prometheus counters and histograms for the application.
type Metrics struct {
	RequestsTotal         *prometheus.CounterVec
	RequestDuration       *prometheus.HistogramVec
	ExternalRequestsTotal *prometheus.CounterVec
	ExternalErrorsTotal   *prometheus.CounterVec
	DBWriteErrorsTotal    *prometheus.CounterVec
}

// NewMetrics creates and registers all application Prometheus metrics.
func NewMetrics(registry *prometheus.Registry) *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "price_oracle_requests_total",
			Help: "Total number of gRPC requests",
		}, []string{"method", "status"}),

		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "price_oracle_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"method"}),

		ExternalRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "price_oracle_external_requests_total",
			Help: "Total number of external API requests (Grinex)",
		}, []string{"endpoint"}),

		ExternalErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "price_oracle_external_request_errors_total",
			Help: "Total number of failed external API requests",
		}, []string{"endpoint"}),

		DBWriteErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "price_oracle_db_write_errors_total",
			Help: "Total number of database write errors",
		}, nil),
	}

	registry.MustRegister(
		m.RequestsTotal,
		m.RequestDuration,
		m.ExternalRequestsTotal,
		m.ExternalErrorsTotal,
		m.DBWriteErrorsTotal,
	)

	return m
}

// MetricsHandler returns an HTTP handler serving Prometheus metrics.
func MetricsHandler(registry *prometheus.Registry) http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}
