# Multi-stage build for CloudBridge Relay Server
FROM golang:1.25.1-bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata git && rm -rf /var/lib/apt/lists/*

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
FROM gcr.io/distroless/base-debian12:nonroot

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/relay-server .

# Copy configuration files
COPY --from=builder /app/configs/ ./configs/
COPY --from=builder /app/certs/ ./certs/

USER nonroot

# Expose ports
EXPOSE 8080 8443 9090 9091

# Distroless image does not include shell/wget; rely on k8s/externals for health checks

# Run the relay server
CMD ["./relay-server", "-config", "configs/config.json"]
