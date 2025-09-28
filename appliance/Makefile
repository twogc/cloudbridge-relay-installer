# CloudBridge Appliance Makefile

.PHONY: help build test clean deps install

# Default target
help:
	@echo "CloudBridge Appliance Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  deps     - Install dependencies"
	@echo "  build    - Build appliance components"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  install  - Install appliance"
	@echo "  docker   - Build Docker images"
	@echo "  pilot    - Setup pilot program"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	cd api && go mod tidy
	cd api && go mod download

# Build appliance components
build: deps
	@echo "Building CloudBridge Appliance..."
	cd api && go build -o ../bin/appliance-manager ./appliance_manager.go ./server.go ./types.go

# Run tests
test: deps
	@echo "Running tests..."
	cd api && go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf packages/

# Install appliance
install: build
	@echo "Installing CloudBridge Appliance..."
	sudo ./scripts/install-appliance.sh

# Build Docker images
docker:
	@echo "Building Docker images..."
	cd docker && ./build-appliance.sh

# Setup pilot program
pilot:
	@echo "Setting up pilot program..."
	@echo "Pilot program documentation: pilot/pilot-program.md"
	@echo "Application template: pilot/pilot-application.json"
	@echo "Metrics template: pilot/pilot-metrics.json"

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@echo "Dependencies installed"
	@echo "Ready for development"

# Check code quality
lint:
	@echo "Running linters..."
	cd api && go vet ./...
	cd api && go fmt ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	@echo "Documentation available in docs/ directory"
