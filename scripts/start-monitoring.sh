#!/bin/bash

# Monitoring Services Startup Script
# Starts Grafana, Prometheus, and AI Service port-forwards

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 CloudBridge Monitoring Services Startup${NC}"
echo "================================================"

# Function to check if port is in use
check_port() {
    local port="$1"
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0  # Port is in use
    else
        return 1  # Port is free
    fi
}

# Function to kill existing port-forwards
cleanup_port_forwards() {
    echo -e "${YELLOW}🧹 Cleaning up existing port-forwards...${NC}"
    
    # Kill existing kubectl port-forward processes
    pkill -f "kubectl.*port-forward" 2>/dev/null || true
    
    # Wait a moment for cleanup
    sleep 2
}

# Function to start port-forward
start_port_forward() {
    local service="$1"
    local namespace="$2"
    local local_port="$3"
    local remote_port="$4"
    local description="$5"
    
    echo -n "Starting $description on port $local_port... "
    
    if check_port $local_port; then
        echo -e "${YELLOW}  Port $local_port already in use${NC}"
        return 1
    fi
    
    # Start port-forward in background
    kubectl port-forward -n $namespace svc/$service $local_port:$remote_port >/dev/null 2>&1 &
    local pid=$!
    
    # Wait a moment and check if it's still running
    sleep 2
    if kill -0 $pid 2>/dev/null; then
        echo -e "${GREEN}Started (PID: $pid)${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed to start${NC}"
        return 1
    fi
}

# Function to check service availability
check_service() {
    local url="$1"
    local name="$2"
    
    echo -n "Checking $name availability... "
    
    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "200\|302\|401"; then
        echo -e "${GREEN}Available${NC}"
        return 0
    else
        echo -e "${RED}❌ Not available${NC}"
        return 1
    fi
}

# Main function
main() {
    echo -e "${YELLOW}Starting CloudBridge Monitoring Services...${NC}"
    echo
    
    # Cleanup existing port-forwards
    cleanup_port_forwards
    
    # Check if services are running in Kubernetes
    echo -e "${BLUE}🔍 Checking Kubernetes services...${NC}"
    
    # Check Grafana
    if kubectl get pod -n cloudbridge-monitoring -l app.kubernetes.io/name=grafana --no-headers 2>/dev/null | grep -q Running; then
        echo -e "${GREEN}Grafana pod is running${NC}"
    else
        echo -e "${RED}❌ Grafana pod not found or not running${NC}"
        echo -e "${YELLOW}   Run: kubectl get pods -n cloudbridge-monitoring${NC}"
        exit 1
    fi
    
    # Check Prometheus
    if kubectl get pod -n cloudbridge-monitoring -l app.kubernetes.io/name=prometheus --no-headers 2>/dev/null | grep -q Running; then
        echo -e "${GREEN}Prometheus pod is running${NC}"
    else
        echo -e "${RED}❌ Prometheus pod not found or not running${NC}"
        echo -e "${YELLOW}   Run: kubectl get pods -n cloudbridge-monitoring${NC}"
        exit 1
    fi
    
    # Check AI Service
    if kubectl get pod -n cloudbridge -l app.kubernetes.io/name=ai-service --no-headers 2>/dev/null | grep -q Running; then
        echo -e "${GREEN}AI Service pod is running${NC}"
    else
        echo -e "${RED}❌ AI Service pod not found or not running${NC}"
        echo -e "${YELLOW}   Run: kubectl get pods -n cloudbridge${NC}"
        exit 1
    fi
    
    echo
    
    # Start port-forwards
    echo -e "${BLUE}🚀 Starting port-forwards...${NC}"
    
    start_port_forward "grafana" "cloudbridge-monitoring" "3000" "3000" "Grafana Dashboard"
    start_port_forward "prometheus" "cloudbridge-monitoring" "9090" "9090" "Prometheus"
    start_port_forward "ai-service" "cloudbridge" "8083" "8083" "AI Service API"
    
    echo
    
    # Wait for services to be ready
    echo -e "${YELLOW}⏳ Waiting for services to be ready...${NC}"
    sleep 5
    
    # Check service availability
    echo -e "${BLUE}🔍 Checking service availability...${NC}"
    
    check_service "http://localhost:3000" "Grafana"
    check_service "http://localhost:9090" "Prometheus"
    check_service "http://localhost:8083/health" "AI Service"
    
    echo
    
    # Show access information
    echo -e "${PURPLE}🌐 ACCESS INFORMATION${NC}"
    echo "================================================"
    
    echo -e "${BLUE}📊 Grafana Dashboard:${NC}"
    echo "   URL: http://localhost:3000"
    echo "   Login: admin"
    echo "   Password: admin"
    echo
    
    echo -e "${BLUE} Prometheus:${NC}"
    echo "   URL: http://localhost:9090"
    echo "   No authentication required"
    echo
    
    echo -e "${BLUE} AI Service API:${NC}"
    echo "   URL: http://localhost:8083"
    echo "   Health: http://localhost:8083/health"
    echo "   Status: http://localhost:8083/api/v1/ai/status"
    echo
    
    echo -e "${BLUE}📖 API Documentation:${NC}"
    echo "   Run: ./scripts/start-api-docs.sh"
    echo "   URL: http://localhost:8080/ai-api-docs.html"
    echo
    
    echo -e "${GREEN}🎉 All monitoring services are now accessible!${NC}"
    echo
    
    echo -e "${YELLOW}📝 Useful Commands:${NC}"
    echo "   Stop all port-forwards: pkill -f 'kubectl.*port-forward'"
    echo "   Check running port-forwards: ps aux | grep 'kubectl.*port-forward'"
    echo "   View Grafana logs: kubectl logs -n cloudbridge-monitoring -l app.kubernetes.io/name=grafana"
    echo "   View Prometheus logs: kubectl logs -n cloudbridge-monitoring -l app.kubernetes.io/name=prometheus"
    echo "   View AI Service logs: kubectl logs -n cloudbridge -l app.kubernetes.io/name=ai-service"
    echo
    
    echo -e "${BLUE}🚀 Next Steps:${NC}"
    echo "   1. Open Grafana at http://localhost:3000"
    echo "   2. Import AI Service dashboard"
    echo "   3. Configure alerts and notifications"
    echo "   4. Monitor AI service performance"
    echo "   5. Test AI API endpoints"
}

# Run main function
main "$@"


