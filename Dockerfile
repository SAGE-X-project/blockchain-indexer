# Multi-stage build for blockchain indexer
# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty) \
    -X main.commit=$(git rev-parse --short HEAD) \
    -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /build/bin/indexer ./cmd/indexer

# Stage 2: Runtime
FROM alpine:3.22

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create app user
RUN addgroup -g 1000 indexer && \
    adduser -D -u 1000 -G indexer indexer

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/indexer /app/indexer

# Copy example configurations
COPY --from=builder /build/config/*.example.yaml /app/config/

# Create data directory
RUN mkdir -p /app/data && chown -R indexer:indexer /app

# Switch to non-root user
USER indexer

# Expose ports
# 8080: HTTP (REST + GraphQL)
# 50051: gRPC
# 9091: Metrics
EXPOSE 8080 50051 9091

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set entrypoint
ENTRYPOINT ["/app/indexer"]

# Default command
CMD ["server", "--config", "/app/config/config.yaml"]
