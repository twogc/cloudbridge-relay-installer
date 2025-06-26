# CloudBridge Relay Makefile

# Variables
BINARY_NAME=relay
BUILD_DIR=build
DOCKER_IMAGE=cloudbridge-relay
DOCKER_TAG=latest
GO_VERSION=1.21
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

# Default target
.PHONY: all
all: clean build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/relay

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/relay
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/relay
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/relay
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/relay

# Run the application
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	go run ./cmd/relay

# Run with development configuration
.PHONY: run-dev
run-dev:
	@echo "Running $(BINARY_NAME) in development mode..."
	go run ./cmd/relay --config config/config.json --log-level debug

# Test the application
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run integration tests
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

# Run E2E tests
.PHONY: test-e2e
test-e2e:
	@echo "Running E2E tests..."
	go test -v ./tests/e2e/...

# Run all tests
.PHONY: test-all
test-all: test test-integration test-e2e

# Test new metrics
.PHONY: test-metrics
test-metrics:
	@echo "Testing new metrics functionality..."
	go test -v ./internal/metrics/...

# Test Keycloak integration
.PHONY: test-keycloak
test-keycloak:
	@echo "Testing Keycloak integration..."
	go test -v ./tests/integration/keycloak_integration_test.go

# Lint the code
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet the code
.PHONY: vet
vet:
	@echo "Vetting code..."
	go vet ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Docker run
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -d \
		--name relay \
		-p 8080:8080 \
		-p 8081:8081 \
		-v $(PWD)/certs:/etc/relay/certs \
		-v $(PWD)/logs:/var/log/relay \
		-e RELAY_JWT_SECRET=test-secret \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker stop
.PHONY: docker-stop
docker-stop:
	@echo "Stopping Docker container..."
	docker stop relay || true
	docker rm relay || true

# Docker Compose up
.PHONY: compose-up
compose-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Docker Compose down
.PHONY: compose-down
compose-down:
	@echo "Stopping services with Docker Compose..."
	docker-compose down

# Docker Compose logs
.PHONY: compose-logs
compose-logs:
	@echo "Showing Docker Compose logs..."
	docker-compose logs -f

# Keycloak integration
.PHONY: keycloak-up
keycloak-up:
	@echo "Starting Keycloak with Docker Compose..."
	docker-compose -f docker-compose.keycloak.yml up -d

.PHONY: keycloak-down
keycloak-down:
	@echo "Stopping Keycloak services..."
	docker-compose -f docker-compose.keycloak.yml down

.PHONY: keycloak-logs
keycloak-logs:
	@echo "Showing Keycloak logs..."
	docker-compose -f docker-compose.keycloak.yml logs -f

.PHONY: keycloak-setup
keycloak-setup:
	@echo "Setting up Keycloak for CloudBridge Relay..."
	@chmod +x scripts/setup_keycloak.sh
	./scripts/setup_keycloak.sh

.PHONY: keycloak-test
keycloak-test:
	@echo "Testing Keycloak integration..."
	@echo "Getting access token..."
	@curl -s -X POST http://localhost:8080/realms/cloudbridge/protocol/openid-connect/token \
		-H "Content-Type: application/x-www-form-urlencoded" \
		-d "grant_type=password" \
		-d "client_id=relay-client" \
		-d "client_secret=relay-client-secret" \
		-d "username=admin" \
		-d "password=admin123" | jq -r '.access_token' > /tmp/keycloak_token
	@echo "Token obtained and saved to /tmp/keycloak_token"
	@echo "Testing relay health endpoint..."
	@curl -s -H "Authorization: Bearer $(cat /tmp/keycloak_token)" \
		http://localhost:8081/health | jq .

# Log management
.PHONY: logs-rotate
logs-rotate:
	@echo "Rotating log files..."
	@chmod +x scripts/manage_logs.sh
	./scripts/manage_logs.sh rotate

.PHONY: logs-clean
logs-clean:
	@echo "Cleaning old log files..."
	@chmod +x scripts/manage_logs.sh
	./scripts/manage_logs.sh clean

.PHONY: logs-analyze
logs-analyze:
	@echo "Analyzing log files..."
	@chmod +x scripts/manage_logs.sh
	./scripts/manage_logs.sh analyze

.PHONY: logs-tail
logs-tail:
	@echo "Tailing log files..."
	@chmod +x scripts/manage_logs.sh
	./scripts/manage_logs.sh tail --follow

.PHONY: logs-search
logs-search:
	@echo "Searching logs for pattern: $(PATTERN)"
	@chmod +x scripts/manage_logs.sh
	./scripts/manage_logs.sh search "$(PATTERN)"

.PHONY: logs-stats
logs-stats:
	@echo "Showing log statistics..."
	@chmod +x scripts/manage_logs.sh
	./scripts/manage_logs.sh stats

.PHONY: logs-compress
logs-compress:
	@echo "Compressing old log files..."
	@chmod +x scripts/manage_logs.sh
	./scripts/manage_logs.sh compress

.PHONY: logs-setup
logs-setup:
	@echo "Setting up log rotation..."
	@chmod +x scripts/manage_logs.sh
	./scripts/manage_logs.sh setup

# Monitoring and metrics
.PHONY: metrics-check
metrics-check:
	@echo "Checking Prometheus metrics..."
	@curl -s http://localhost:8081/metrics | grep -E "(relay_|keycloak_)" | head -20

.PHONY: metrics-keycloak
metrics-keycloak:
	@echo "Checking Keycloak-specific metrics..."
	@curl -s http://localhost:8081/metrics | grep "relay_keycloak_" | head -10

.PHONY: metrics-business
metrics-business:
	@echo "Checking business metrics..."
	@curl -s http://localhost:8081/metrics | grep "relay_business_" | head -10

.PHONY: metrics-technical
metrics-technical:
	@echo "Checking technical metrics..."
	@curl -s http://localhost:8081/metrics | grep "relay_technical_" | head -10

.PHONY: grafana-check
grafana-check:
	@echo "Checking Grafana availability..."
	@curl -s http://localhost:3000/api/health | jq .

.PHONY: prometheus-check
prometheus-check:
	@echo "Checking Prometheus availability..."
	@curl -s http://localhost:9090/-/healthy

.PHONY: alertmanager-check
alertmanager-check:
	@echo "Checking AlertManager availability..."
	@curl -s http://localhost:9093/-/healthy

# Health checks
.PHONY: health-check
health-check:
	@echo "Performing comprehensive health check..."
	@echo "1. Relay service..."
	@curl -s http://localhost:8081/health | jq . || echo "Relay service not responding"
	@echo "2. Prometheus..."
	@curl -s http://localhost:9090/-/healthy || echo "Prometheus not responding"
	@echo "3. Grafana..."
	@curl -s http://localhost:3000/api/health | jq . || echo "Grafana not responding"
	@echo "4. Keycloak..."
	@curl -s http://localhost:8080/realms/cloudbridge | jq . || echo "Keycloak not responding"

# Performance testing
.PHONY: perf-test
perf-test:
	@echo "Running performance tests..."
	@echo "Testing connection throughput..."
	@for i in {1..10}; do \
		curl -s -w "%{time_total}\n" -o /dev/null http://localhost:8081/health; \
	done | awk '{sum+=$1} END {print "Average response time:", sum/NR "s"}'

.PHONY: load-test
load-test:
	@echo "Running load test..."
	@echo "Simulating 100 concurrent requests..."
	@for i in {1..100}; do \
		curl -s -w "%{http_code}\n" -o /dev/null http://localhost:8081/health & \
	done; wait

# Security testing
.PHONY: security-test
security-test:
	@echo "Running security tests..."
	@echo "1. Testing invalid token..."
	@curl -s -H "Authorization: Bearer invalid-token" \
		http://localhost:8081/health | jq . || echo "Expected error for invalid token"
	@echo "2. Testing missing token..."
	@curl -s http://localhost:8081/health | jq . || echo "Expected error for missing token"

# Development utilities
.PHONY: dev-setup
dev-setup: deps logs-setup
	@echo "Setting up development environment..."
	@mkdir -p logs certs config
	@echo "Development environment ready"

.PHONY: dev-reset
dev-reset: compose-down keycloak-down clean
	@echo "Resetting development environment..."
	@rm -rf logs/* certs/* config/*
	@echo "Development environment reset"

.PHONY: dev-logs
dev-logs:
	@echo "Showing all development logs..."
	@docker-compose logs -f relay &
	@docker-compose -f docker-compose.keycloak.yml logs -f keycloak &
	@wait

# Documentation
.PHONY: docs-build
docs-build:
	@echo "Building documentation..."
	@mkdir -p docs/build
	@cp docs/*.md docs/build/
	@echo "Documentation built in docs/build/"

.PHONY: docs-serve
docs-serve:
	@echo "Serving documentation on http://localhost:8000"
	@cd docs && python3 -m http.server 8000

# Install as system service
.PHONY: install
install:
	@echo "Installing as system service..."
	sudo ./install.sh

# Help
.PHONY: help
help:
	@echo "CloudBridge Relay Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests"
	@echo "  docker-build   - Build Docker image"
	@echo "  compose-up     - Start all services"
	@echo "  keycloak-up    - Start Keycloak services"
	@echo "  logs-rotate    - Rotate log files"
	@echo "  logs-analyze   - Analyze log files"
	@echo "  metrics-check  - Check Prometheus metrics"
	@echo "  health-check   - Comprehensive health check"
	@echo "  dev-setup      - Setup development environment"
	@echo "  help           - Show this help" 