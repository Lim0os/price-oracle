# price-oracle

gRPC service for fetching USDT exchange rates from Grinex, calculating prices using configurable strategies, and persisting results to PostgreSQL.

## Architecture

Built with **Hexagonal Architecture (Ports & Adapters)** and dependency injection via [uber/fx](https://github.com/uber-go/fx).

```
cmd/price-oracle/          ── Application entry point (fx wiring)
internal/
  domain/                  ── Core: value objects, calculators, port interfaces
  app/
    query/                 ── Read use case (GetRates)
    command/               ── Write use case (SaveRate)
  adapter/
    exchange/              ── Grinex HTTP client (resty)
    persistence/           ── PostgreSQL repository (pgx)
    logging/               ── Zap logger + gRPC interceptor
    telemetry/             ── OpenTelemetry tracer + Prometheus metrics
  config/                  ── Configuration (env + flags)
  ports/grpc/              ── gRPC server and handlers
proto/rates/v1/            ── Protobuf service definition
migrations/                ── SQL migrations (golang-migrate)
```

## Quick Start

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- `protoc` and `protoc-gen-go` (for proto generation)
- `golang-migrate` CLI (for database migrations)

### Full Stack (Docker Compose)

```bash
# Start everything: postgres, app, prometheus, jaeger
make docker-up

# Apply database migrations
make migrate-up

# Follow application logs
make docker-logs

# Stop all services
make docker-down
```

### Run Locally (without Docker)

```bash
# 1. Start PostgreSQL (via Docker)
docker compose up -d postgres

# 2. Apply migrations
make migrate-up

# 3. Run the application
make run
```

### Build Binary

```bash
make build
./bin/price-oracle
```

### Build Docker Image

```bash
make docker-build
```

## Configuration

Priority: **CLI flag > Environment variable > Default value**

| Variable | Flag | Default | Description |
|---|---|---|---|
| `APP_ENV` | `--app-env` | `local` | Application environment |
| `GRPC_HOST` | `--grpc-host` | `0.0.0.0` | gRPC server host |
| `GRPC_PORT` | `--grpc-port` | `50051` | gRPC server port |
| `HTTP_TIMEOUT` | `--http-timeout` | `5s` | HTTP client timeout |
| `GRINEX_BASE_URL` | `--grinex-base-url` | `https://grinex.io` | Grinex API base URL |
| `GRINEX_SYMBOL` | `--grinex-symbol` | `usdta7a5` | Trading pair symbol |
| `DB_HOST` | `--db-host` | `localhost` | PostgreSQL host |
| `DB_PORT` | `--db-port` | `5432` | PostgreSQL port |
| `DB_NAME` | `--db-name` | `price_oracle` | Database name |
| `DB_USER` | `--db-user` | `postgres` | Database user |
| `DB_PASSWORD` | `--db-password` | `postgres` | Database password |
| `DB_SSLMODE` | `--db-sslmode` | `disable` | SSL mode |
| `OTEL_ENABLED` | `--otel-enabled` | `true` | Enable OpenTelemetry |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `--otel-exporter-otlp-endpoint` | — | OTLP exporter endpoint |
| `OTEL_INSECURE` | `--otel-insecure` | `false` | Use insecure OTLP connection |
| `LOG_LEVEL` | `--log-level` | `info` | Log level (debug/info/warn/error) |

Example:

```bash
# Via environment variables
GRPC_PORT=50052 DB_HOST=remote-db make run

# Via CLI flags (overrides env)
make run -- --grpc-port=50052 --db-host=remote-db
```

## gRPC API

### Service

```protobuf
service RatesService {
  rpc GetRates(GetRatesRequest) returns (GetRatesResponse);
  rpc Healthcheck(HealthcheckRequest) returns (HealthcheckResponse);
}
```

### GetRates

Calculate rates using a chosen strategy.

| Field | Type | Description |
|---|---|---|
| `strategy` | string | `topN` or `avgNM` |
| `n` | int32 | Position index (1-based) |
| `m` | int32 | End index for `avgNM` (ignored for `topN`) |

**Example requests:**

```bash
# Best ask/bid (1st position)
grpcurl -plaintext \
  -d '{"strategy":"topN","n":1}' \
  localhost:50051 rates.v1.RatesService/GetRates

# Average of positions 1-5
grpcurl -plaintext \
  -d '{"strategy":"avgNM","n":1,"m":5}' \
  localhost:50051 rates.v1.RatesService/GetRates
```

### Healthcheck

```bash
grpcurl -plaintext localhost:50051 rates.v1.RatesService/Healthcheck
```

Returns `SERVING` or `NOT_SERVING` based on database health.

## Strategies

### topN

Returns the price from the N-th position of the order book (1-based index).

```json
{"strategy": "topN", "n": 1}
```

### avgNM

Returns the arithmetic mean of prices in the range [N, M] (1-based, inclusive).

```json
{"strategy": "avgNM", "n": 1, "m": 5}
```

## Observability

### Endpoints

| Service | URL | Description |
|---|---|---|
| gRPC | `localhost:50051` | Main API |
| Prometheus metrics | `localhost:9090/metrics` | Metrics scrape endpoint |
| Prometheus UI | `http://localhost:9091` | Prometheus dashboard |
| Jaeger UI | `http://localhost:16686` | Distributed tracing UI |

### Prometheus Metrics

| Metric | Type | Labels | Description |
|---|---|---|---|
| `price_oracle_requests_total` | Counter | `method`, `status` | Total gRPC requests |
| `price_oracle_request_duration_seconds` | Histogram | `method` | Request duration |
| `price_oracle_external_requests_total` | Counter | `endpoint` | Grinex API requests |
| `price_oracle_external_request_errors_total` | Counter | `endpoint` | Failed Grinex requests |
| `price_oracle_db_write_errors_total` | Counter | — | Database write errors |

Example PromQL queries:

```promql
# Request rate by method
sum(rate(price_oracle_requests_total[1m])) by (method, status)

# P95 latency
histogram_quantile(0.95, sum(rate(price_oracle_request_duration_seconds_bucket[5m])) by (le))

# Error rate
sum(rate(price_oracle_requests_total{status="error"}[1m])) by (method)
```

## Makefile Commands

| Command | Description |
|---|---|
| `make build` | Build binary to `bin/price-oracle` |
| `make run` | Run application locally |
| `make test` | Run all tests with race detection and coverage |
| `make lint` | Run golangci-lint |
| `make proto` | Generate Go code from `.proto` files |
| `make docker-build` | Build Docker image |
| `make docker-up` | Start full stack (postgres + app + prometheus + jaeger) |
| `make docker-down` | Stop all containers |
| `make docker-logs` | Follow application logs |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-force VERSION=0` | Force migration version |
| `make clean` | Remove build artifacts |

## Database Migrations

Uses [golang-migrate](https://github.com/golang-migrate/migrate). Install:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Migrations are stored in `migrations/` and use the default connection:

```
postgres://postgres:postgres@localhost:5432/price_oracle?sslmode=disable
```

Override with:

```bash
make migrate-up DB_DSN="postgres://user:pass@host:5432/dbname?sslmode=require"
```

## Testing

```bash
make test
```

Coverage:

| Package | Coverage |
|---|---|
| `adapter/exchange` | 100% |
| `app/query` | 100% |
| `domain` | 83% |
| `ports/grpc` | 36% |
| `adapter/persistence` | 22% |

## Tech Stack

- **Go** 1.25+
- **gRPC** / Protocol Buffers
- **PostgreSQL** 16+ via `pgx`
- **resty** — HTTP client
- **zap** — structured logging
- **OpenTelemetry** — distributed tracing
- **Prometheus** — metrics
- **uber/fx** — dependency injection
- **shopspring/decimal** — precise decimal arithmetic
- **gofrs/uuid** — UUID v7 generation
- **golang-migrate** — database migrations
