# Multi-stage build for Enhanced QUIC
FROM golang:1.25.1-alpine AS builder

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

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o enhanced-quic .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S enhanced-quic && \
    adduser -u 1001 -S enhanced-quic -G enhanced-quic

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/enhanced-quic .

# Copy configuration files
COPY --from=builder /app/config ./config

# Change ownership
RUN chown -R enhanced-quic:enhanced-quic /app

# Switch to non-root user
USER enhanced-quic

# Expose ports
EXPOSE 8080 8081 8082

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command
ENTRYPOINT ["./enhanced-quic"]

# Default arguments
CMD ["--config", "/app/config/config.json"] 