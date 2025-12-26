# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files first (for better layer caching)
COPY go.mod ./
# Copy go.sum if it exists
COPY go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/api ./cmd/api

# Install golang-migrate for migrations (use same version as in go.mod)
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates wget

# Copy binary from builder
COPY --from=builder /app/bin/api /app/bin/api

# Copy migrate binary from builder
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Copy migrations directory from builder
COPY --from=builder /app/migrations /app/migrations

# Create non-root user
RUN adduser -D -g '' appuser

# Create storage directory
RUN mkdir -p /app/storage && chown -R appuser:appuser /app

# Set working directory
WORKDIR /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./bin/api"]

