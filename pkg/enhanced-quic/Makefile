# Enhanced QUIC Package Makefile

.PHONY: help build test clean install docker-build docker-push npm-install npm-build npm-publish release

# Variables
BINARY_NAME=enhanced-quic
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Go build targets
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o ${BINARY_NAME} .

build-linux: ## Build for Linux
	@echo "Building ${BINARY_NAME} for Linux..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux-amd64 .

build-darwin: ## Build for macOS
	@echo "Building ${BINARY_NAME} for macOS..."
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-darwin-amd64 .

build-windows: ## Build for Windows
	@echo "Building ${BINARY_NAME} for Windows..."
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-windows-amd64.exe .

build-all: build-linux build-darwin build-windows ## Build for all platforms

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f ${BINARY_NAME}*
	rm -f coverage.out coverage.html

install: build ## Install the binary
	@echo "Installing ${BINARY_NAME}..."
	sudo cp ${BINARY_NAME} /usr/local/bin/

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t twogc/enhanced-quic:${VERSION} .
	docker tag twogc/enhanced-quic:${VERSION} twogc/enhanced-quic:latest

docker-push: docker-build ## Build and push Docker image
	@echo "Pushing Docker image..."
	docker push twogc/enhanced-quic:${VERSION}
	docker push twogc/enhanced-quic:latest

# NPM targets
npm-install: ## Install NPM dependencies
	@echo "Installing NPM dependencies..."
	npm install

npm-build: npm-install ## Build NPM package
	@echo "Building NPM package..."
	npm run build

npm-publish: npm-build ## Publish NPM package
	@echo "Publishing NPM package..."
	npm publish

# Release targets
release: build-all test lint ## Prepare release
	@echo "Preparing release ${VERSION}..."
	@echo "Built binaries:"
	@ls -la ${BINARY_NAME}*

release-tag: ## Create and push a new tag
	@echo "Creating tag ${VERSION}..."
	git tag ${VERSION}
	git push origin ${VERSION}

# Development targets
dev: ## Run in development mode
	@echo "Running in development mode..."
	go run . --config config/config.json

dev-docker: ## Run in Docker development mode
	@echo "Running in Docker development mode..."
	docker run -it --rm -p 8080:8080 -p 8081:8081 -p 8082:8082 \
		-v $(PWD)/config:/app/config \
		twogc/enhanced-quic:latest

# Documentation targets
docs: ## Generate documentation
	@echo "Generating documentation..."
	godoc -http=:6060

# Security targets
security-scan: ## Run security scan
	@echo "Running security scan..."
	gosec ./...

# Performance targets
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. ./...

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy 