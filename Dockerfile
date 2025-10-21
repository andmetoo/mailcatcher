# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -o mailcatcher ./cmd/mailcatcher

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/mailcatcher .

# Expose ports
EXPOSE 1025 8025

# Create non-root user
RUN addgroup -g 1000 mailcatcher && \
    adduser -D -u 1000 -G mailcatcher mailcatcher && \
    chown -R mailcatcher:mailcatcher /app

USER mailcatcher

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8025/api/v1/emails || exit 1

ENTRYPOINT ["/app/mailcatcher"]
CMD []
