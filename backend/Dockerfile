# =============================================================================
# Unified Backend Dockerfile
# Build all Go services in one image, share dependencies cache
# =============================================================================

# Build stage - shared for all services
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files first (cache layer)
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code (this layer changes frequently)
COPY . .

# Build all binaries with shared cache
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/ingestor ./cmd/ingestor && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/websocket ./cmd/websocket && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/ai-worker ./cmd/ai-worker

# =============================================================================
# Runtime stages - minimal images for each service
# =============================================================================

FROM alpine:latest AS base
RUN apk --no-cache add ca-certificates
WORKDIR /app

# API service
FROM base AS api
COPY --from=builder /out/api .
EXPOSE 8080
CMD ["./api"]

# Ingestor service (gRPC)
FROM base AS ingestor
COPY --from=builder /out/ingestor .
EXPOSE 50051
CMD ["./ingestor"]

# WebSocket service
FROM base AS websocket
COPY --from=builder /out/websocket .
EXPOSE 8081
CMD ["./websocket"]

# AI Worker service
FROM base AS ai-worker
COPY --from=builder /out/ai-worker .
CMD ["./ai-worker"]
