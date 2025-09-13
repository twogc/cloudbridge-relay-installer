# Multi-stage build for CloudBridge Relay Server
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the relay server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o relay-server ./cmd/relay/

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S relay && \
    adduser -u 1001 -S relay -G relay

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/relay-server .

# Copy configuration files
COPY --from=builder /app/configs/ ./configs/
COPY --from=builder /app/certs/ ./certs/

# Change ownership to non-root user
RUN chown -R relay:relay /app

# Switch to non-root user
USER relay

# Expose ports
EXPOSE 8080 8443 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the relay server
CMD ["./relay-server", "-config", "configs/config.json"]
