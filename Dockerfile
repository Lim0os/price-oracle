# ─── Stage 1: Download dependencies ───────────────────────────────────────────
FROM golang:1.25-alpine AS deps

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# ─── Stage 2: Build the binary ────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency cache from deps stage
COPY --from=deps /go/pkg/mod /go/pkg/mod
COPY go.mod go.sum ./

# Ensure dependencies are in sync
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /app/price-oracle \
    ./cmd/price-oracle/

# ─── Stage 3: Minimal runtime image ───────────────────────────────────────────
FROM alpine:3.21 AS runtime

# Install CA certificates (for HTTPS) and tzdata (for timezones)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy only the binary from builder
COPY --from=builder /app/price-oracle .

# Copy migrations
COPY migrations/ ./migrations/

# Non-root user for security
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -D -s /bin/sh appuser

USER appuser

EXPOSE 50051

ENTRYPOINT ["./price-oracle"]
