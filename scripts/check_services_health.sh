#!/bin/bash

# CloudBridge Services Health Check
# Week 1, Task 1.1: Проверка статуса всех сервисов
# Date: 2025-10-28

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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
    echo -e "${RED}[✗]${NC} $1"
}

# Service health checks
declare -A SERVICES=(
    ["VictoriaMetrics"]="http://localhost:8428/-/healthy"
    ["Prometheus"]="http://localhost:9090/-/healthy"
    ["Grafana"]="http://localhost:3000/api/health"
    ["Redis"]="localhost:6379"
    ["PostgreSQL"]="localhost:5432"
    ["Zitadel"]="http://localhost:8080/healthz"
    ["AI Service"]="http://localhost:8000/health"
    ["Relay"]="http://localhost:8081/healthz"
    ["DDoS Protection"]="http://localhost:8082/healthz"
    ["DNS Core"]="localhost:53"
)

main() {
    log_info "CloudBridge Services Health Check"
    log_info "=================================="
    echo ""

    local healthy=0
    local unhealthy=0
    local unknown=0

    # Check Docker containers
    log_info "Checking Docker containers..."
    local running_containers=$(docker ps --format "{{.Names}}" 2>/dev/null | wc -l)
    log_info "Running containers: $running_containers"
    echo ""

    # Check each service
    for service in "${!SERVICES[@]}"; do
        endpoint="${SERVICES[$service]}"
        check_service "$service" "$endpoint"
        status=$?

        if [ $status -eq 0 ]; then
            ((healthy++))
        elif [ $status -eq 1 ]; then
            ((unhealthy++))
        else
            ((unknown++))
        fi
    done

    # Print summary
    echo ""
    echo -e "${BLUE}=================================="
    echo "HEALTH CHECK SUMMARY"
    echo "==================================${NC}"
    echo -e "${GREEN}✓ Healthy: $healthy${NC}"
    echo -e "${RED}✗ Unhealthy: $unhealthy${NC}"
    echo -e "${YELLOW}? Unknown: $unknown${NC}"
    echo ""

    if [ $unhealthy -gt 0 ]; then
        log_warn "Some services are unhealthy. Check logs for details."
        echo "Use: docker-compose logs -f <service>"
        return 1
    fi

    log_success "All services are healthy!"
    return 0
}

# Check individual service
check_service() {
    local service_name=$1
    local endpoint=$2

    # Check if it's an HTTP endpoint
    if [[ $endpoint == http* ]]; then
        if curl -s -f "$endpoint" > /dev/null 2>&1; then
            log_success "$service_name is healthy"
            return 0
        else
            log_error "$service_name is unhealthy (HTTP check failed)"
            return 1
        fi
    fi

    # Check if it's a TCP endpoint (host:port)
    if [[ $endpoint == *:* ]]; then
        local host=$(echo "$endpoint" | cut -d: -f1)
        local port=$(echo "$endpoint" | cut -d: -f2)

        if timeout 2 bash -c "</dev/tcp/$host/$port" 2>/dev/null; then
            log_success "$service_name is listening on $endpoint"
            return 0
        else
            log_warn "$service_name is not responding on $endpoint"
            return 2  # Unknown/not started yet
        fi
    fi

    log_warn "$service_name endpoint format unknown: $endpoint"
    return 2
}

# Run checks with retry logic
main "$@"

# If first check fails, wait and retry
if [ $? -ne 0 ]; then
    log_info ""
    log_info "Waiting 30 seconds for services to initialize..."
    sleep 30
    log_info "Running health check again..."
    main "$@"
fi
