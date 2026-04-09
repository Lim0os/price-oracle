// Package config provides application configuration loading from environment
// variables and CLI flags with flag precedence.
package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Config holds all application configuration.
type Config struct {
	AppEnv   string
	GRPCHost string
	GRPCPort int

	HTTPTimeout   time.Duration
	GrinexBaseURL string
	GrinexSymbol  string

	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	OTELEnabled  bool
	OTELEndpoint string
	OTELInsecure bool

	LogLevel string
}

// Load reads configuration from environment variables and CLI flags.
// Priority: CLI flag > env var > default value.
func Load() (*Config, error) {
	cfg := &Config{}

	cfg.AppEnv = "local"
	cfg.GRPCHost = "0.0.0.0"
	cfg.GRPCPort = 50051
	cfg.HTTPTimeout = 5 * time.Second
	cfg.GrinexBaseURL = "https://grinex.io"
	cfg.GrinexSymbol = "usdta7a5"
	cfg.DBHost = "localhost"
	cfg.DBPort = 5432
	cfg.DBName = "price_oracle"
	cfg.DBUser = "postgres"
	cfg.DBPassword = "postgres"
	cfg.DBSSLMode = "disable"
	cfg.OTELEnabled = true
	cfg.OTELInsecure = false
	cfg.LogLevel = "info"

	loadEnv(cfg)
	loadFlags(cfg)

	return cfg, nil
}

func loadEnv(cfg *Config) {
	if v := os.Getenv("APP_ENV"); v != "" {
		cfg.AppEnv = v
	}
	if v := os.Getenv("GRPC_HOST"); v != "" {
		cfg.GRPCHost = v
	}
	if v := os.Getenv("GRPC_PORT"); v != "" {
		cfg.GRPCPort = atoi(v, cfg.GRPCPort)
	}
	if v := os.Getenv("HTTP_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.HTTPTimeout = d
		}
	}
	if v := os.Getenv("GRINEX_BASE_URL"); v != "" {
		cfg.GrinexBaseURL = v
	}
	if v := os.Getenv("GRINEX_SYMBOL"); v != "" {
		cfg.GrinexSymbol = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.DBHost = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.DBPort = atoi(v, cfg.DBPort)
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.DBName = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.DBUser = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.DBPassword = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		cfg.DBSSLMode = v
	}
	if v := os.Getenv("OTEL_ENABLED"); v != "" {
		cfg.OTELEnabled = v == "true" || v == "1"
	}
	if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		cfg.OTELEndpoint = v
	}
	if v := os.Getenv("OTEL_INSECURE"); v != "" {
		cfg.OTELInsecure = v == "true" || v == "1"
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
}

func loadFlags(cfg *Config) {
	flag.StringVar(&cfg.AppEnv, "app-env", cfg.AppEnv, "Application environment")
	flag.StringVar(&cfg.GRPCHost, "grpc-host", cfg.GRPCHost, "gRPC server host")
	flag.IntVar(&cfg.GRPCPort, "grpc-port", cfg.GRPCPort, "gRPC server port")
	flag.DurationVar(&cfg.HTTPTimeout, "http-timeout", cfg.HTTPTimeout, "HTTP client timeout")
	flag.StringVar(&cfg.GrinexBaseURL, "grinex-base-url", cfg.GrinexBaseURL, "Grinex API base URL")
	flag.StringVar(&cfg.GrinexSymbol, "grinex-symbol", cfg.GrinexSymbol, "Trading pair symbol")
	flag.StringVar(&cfg.DBHost, "db-host", cfg.DBHost, "PostgreSQL host")
	flag.IntVar(&cfg.DBPort, "db-port", cfg.DBPort, "PostgreSQL port")
	flag.StringVar(&cfg.DBName, "db-name", cfg.DBName, "Database name")
	flag.StringVar(&cfg.DBUser, "db-user", cfg.DBUser, "Database user")
	flag.StringVar(&cfg.DBPassword, "db-password", cfg.DBPassword, "Database password")
	flag.StringVar(&cfg.DBSSLMode, "db-sslmode", cfg.DBSSLMode, "SSL mode")
	flag.BoolVar(&cfg.OTELEnabled, "otel-enabled", cfg.OTELEnabled, "Enable OpenTelemetry")
	flag.StringVar(&cfg.OTELEndpoint, "otel-exporter-otlp-endpoint", cfg.OTELEndpoint, "OTLP exporter endpoint")
	flag.BoolVar(&cfg.OTELInsecure, "otel-insecure", cfg.OTELInsecure, "Use insecure connection for OTLP")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Logging level")

	flag.Parse()
}

func atoi(s string, fallback int) int {
	var result int
	if _, err := fmt.Sscanf(s, "%d", &result); err != nil {
		return fallback
	}
	return result
}
