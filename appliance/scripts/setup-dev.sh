#!/bin/bash
# CloudBridge Appliance Development Setup Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21+ first."
        echo "Download from: https://golang.org/dl/"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Found Go version: $GO_VERSION"
}

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_warning "Docker is not installed. Some features may not work."
    else
        log_info "Found Docker: $(docker --version)"
    fi
}

# Install API dependencies
install_api_deps() {
    log_info "Installing API dependencies..."
    
    cd api
    
    # Initialize go mod if not exists
    if [[ ! -f go.mod ]]; then
        go mod init cloudbridge-appliance-api
    fi
    
    # Add dependencies
    go get github.com/gorilla/mux@v1.8.1
    go get github.com/prometheus/client_golang@v1.17.0
    
    # Download dependencies
    go mod tidy
    go mod download
    
    log_success "API dependencies installed"
    
    cd ..
}

# Create development directories
create_dirs() {
    log_info "Creating development directories..."
    
    mkdir -p bin
    mkdir -p packages
    mkdir -p logs
    
    log_success "Directories created"
}

# Build API components
build_api() {
    log_info "Building API components..."
    
    cd api
    
    # Build main components
    go build -o ../bin/appliance-manager ./appliance_manager.go ./server.go ./types.go
    
    log_success "API components built"
    
    cd ..
}

# Run tests
run_tests() {
    log_info "Running tests..."
    
    cd api
    
    if go test -v ./...; then
        log_success "All tests passed"
    else
        log_warning "Some tests failed"
    fi
    
    cd ..
}

# Setup development environment
setup_dev_env() {
    log_info "Setting up development environment..."
    
    # Create .env file
    cat > .env << EOF
# CloudBridge Appliance Development Environment
APP_MODE=development
LOG_LEVEL=debug
API_HOST=localhost
API_PORT=8080
GRPC_PORT=8444
EOF
    
    log_success "Development environment configured"
}

# Main setup function
main() {
    log_info "Setting up CloudBridge Appliance development environment..."
    
    check_go
    check_docker
    create_dirs
    install_api_deps
    build_api
    run_tests
    setup_dev_env
    
    log_success "Development environment setup completed!"
    echo
    echo "📋 Next steps:"
    echo "  • Run 'make build' to build all components"
    echo "  • Run 'make test' to run tests"
    echo "  • Run 'make docker' to build Docker images"
    echo "  • Run 'make install' to install appliance"
    echo
    echo "📚 Documentation:"
    echo "  • API docs: docs/README.md"
    echo "  • Pricing: docs/pricing-guide.md"
    echo "  • Pilot program: pilot/pilot-program.md"
    echo
    echo "🔧 Development commands:"
    echo "  • make deps     - Install dependencies"
    echo "  • make build    - Build components"
    echo "  • make test     - Run tests"
    echo "  • make clean    - Clean artifacts"
    echo "  • make docker   - Build Docker images"
}

# Run main function
main "$@"
