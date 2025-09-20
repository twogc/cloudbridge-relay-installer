# Multi-stage build for CloudBridge Relay with Enhanced Components
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (skip tests to speed up build)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cloudbridge-relay ./cmd/relay

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk update && apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/cloudbridge-relay .

# Copy config files
COPY --from=builder /app/configs ./configs

# Create directories for enhanced components
RUN mkdir -p /app/data /app/certs && \
    chown -R appuser:appgroup /app/data /app/certs

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
# 8080 - Main TCP server
# 8082 - P2P API server  
# 8443 - MASQUE HTTP/3 server
# 5553 - P2P QUIC server
# 19302 - STUN server
EXPOSE 8080 8082 8443 5553/udp 19302/udp

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./cloudbridge-relay"]