# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o relay ./cmd/relay

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/relay .

# Copy configuration
COPY config/config.json ./config/

# Create directory for logs and set permissions
RUN mkdir -p /var/log/relay && \
    chmod 777 /var/log/relay

# Set environment variables
ENV TZ=UTC

# Expose port
EXPOSE 8080

# Run the application
CMD ["./relay"] 