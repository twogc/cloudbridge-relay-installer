#!/bin/bash

# CloudBridge Services Startup Script
# Week 1, Task 1.1: Установка и запуск всех сервисов
# Date: 2025-10-28

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPOS_DIR="$PROJECT_ROOT/oss-repositories"

# Color codes for output
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
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Main execution
main() {
    log_info "CloudBridge Services Startup - Week 1, Task 1.1"
    log_info "=============================================="
    log_info ""

    # Check prerequisites
    check_prerequisites

    # Array of services to start
    declare -a SERVICES=(
        "cloudbridge-monitoring"           # VictoriaMetrics, Prometheus, Grafana
        "cloudbridge-control-plane"        # Zitadel, Cadence, PostgreSQL
        "cloudbridge-ai-service"           # AI models, FastAPI
        "cloudbridge-ddos-protection"      # DDoS ML protection
        "cloudbridge-dns-network"          # CoreDNS setup
        "cloudbridge-edge-pops"            # Edge PoP deployment
        "cloudbridge-scalable-relay"       # QUIC/MASQUE/P2P relay
        "cloudbridge-dashboard"            # Frontend + BFF
    )

    # Track started services
    STARTED_SERVICES=()
    FAILED_SERVICES=()

    # Start each service
    for service in "${SERVICES[@]}"; do
        start_service "$service"
    done

    # Summary
    print_summary
}

# Check Docker and Docker Compose
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    log_success "Docker found: $(docker --version)"

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi
    log_success "Docker Compose found: $(docker-compose --version)"

    # Check Docker daemon
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    log_success "Docker daemon is running"

    echo ""
}

# Start individual service
start_service() {
    local service_name=$1
    local service_path="$REPOS_DIR/$service_name"

    if [ ! -d "$service_path" ]; then
        log_warn "Service directory not found: $service_path"
        FAILED_SERVICES+=("$service_name (directory not found)")
        return
    fi

    if [ ! -f "$service_path/docker-compose.yml" ]; then
        log_warn "docker-compose.yml not found in $service_path"
        FAILED_SERVICES+=("$service_name (no docker-compose.yml)")
        return
    fi

    log_info "Starting $service_name..."

    cd "$service_path"

    # Try to start with docker-compose
    if docker-compose up -d 2>/dev/null; then
        log_success "$service_name started successfully"
        STARTED_SERVICES+=("$service_name")
    else
        log_error "Failed to start $service_name"
        FAILED_SERVICES+=("$service_name")
    fi

    sleep 2  # Wait between services
}

# Print summary
print_summary() {
    echo ""
    echo -e "${BLUE}=============================================="
    echo "STARTUP SUMMARY"
    echo "==============================================  ${NC}"
    echo ""

    echo -e "${GREEN}✓ Successfully Started (${#STARTED_SERVICES[@]}):${NC}"
    for service in "${STARTED_SERVICES[@]}"; do
        echo "  • $service"
    done

    if [ ${#FAILED_SERVICES[@]} -gt 0 ]; then
        echo ""
        echo -e "${RED}✗ Failed to Start (${#FAILED_SERVICES[@]}):${NC}"
        for service in "${FAILED_SERVICES[@]}"; do
            echo "  • $service"
        done
    fi

    echo ""
    echo -e "${YELLOW}Next Steps:${NC}"
    echo "1. Wait 30-60 seconds for services to fully initialize"
    echo "2. Run: docker ps --format 'table {{.Names}}\t{{.Status}}'"
    echo "3. Check health endpoints:"
    echo "   - VictoriaMetrics: http://localhost:8428"
    echo "   - Grafana: http://localhost:3000"
    echo "   - Prometheus: http://localhost:9090"
    echo ""
    echo "4. To view logs: docker-compose -f <service>/docker-compose.yml logs -f"
    echo ""

    if [ ${#FAILED_SERVICES[@]} -gt 0 ]; then
        log_warn "Some services failed to start. Review the errors above."
        exit 1
    fi
}

# Run main
main "$@"
