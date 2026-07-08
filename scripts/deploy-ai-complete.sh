#!/bin/bash

# Complete AI Service Deployment Script
# Deploys AI service with all features: monitoring, autoscaling, security

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="cloudbridge"
MONITORING_NAMESPACE="cloudbridge-monitoring"
SERVICE_NAME="ai-service"

echo -e "${PURPLE}Complete AI Service Deployment Script${NC}"
echo "================================================"
echo -e "${BLUE}This script will deploy AI service with:${NC}"
echo "  - AI Service with all endpoints"
echo "  - Monitoring and alerting (Prometheus + Grafana)"
echo "  - Autoscaling (HPA + VPA)"
echo "  - Security (NetworkPolicy + RBAC)"
echo "  - Testing and validation"
echo

# Function to print step header
print_step() {
    echo
    echo -e "${PURPLE}STEP $1: $2${NC}"
    echo "================================================"
}

# Function to check if namespace exists
check_namespace() {
    local namespace="$1"
    if kubectl get namespace $namespace >/dev/null 2>&1; then
        echo -e "${GREEN}Namespace $namespace exists${NC}"
        return 0
    else
        echo -e "${RED}Namespace $namespace not found${NC}"
        return 1
    fi
}

# Function to wait for deployment
wait_for_deployment() {
    local deployment="$1"
    local namespace="$2"
    local timeout="${3:-300}"
    
    echo -n "Waiting for $deployment to be ready... "
    
    if kubectl wait --for=condition=available --timeout=${timeout}s deployment/$deployment -n $namespace >/dev/null 2>&1; then
        echo -e "${GREEN}Ready${NC}"
        return 0
    else
        echo -e "${RED}Timeout${NC}"
        return 1
    fi
}

# Function to run script with error handling
run_script() {
    local script="$1"
    local description="$2"
    
    echo -e "${YELLOW}Running: $description${NC}"
    
    if [ -f "$script" ]; then
        if bash "$script"; then
            echo -e "${GREEN}$description completed successfully${NC}"
            return 0
        else
            echo -e "${RED}$description failed${NC}"
            return 1
        fi
    else
        echo -e "${RED}Script $script not found${NC}"
        return 1
    fi
}

# Function to test AI service endpoints
test_ai_service() {
    echo -e "${YELLOW}Testing AI Service endpoints...${NC}"
    
    # Get AI service pod
    POD_NAME=$(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=$SERVICE_NAME -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    
    if [ -z "$POD_NAME" ]; then
        echo -e "${RED}No AI service pod found${NC}"
        return 1
    fi
    
    # Test key endpoints
    local endpoints=(
        "/health:Health Check"
        "/api/v1/ai/status:AI Service Status"
        "/api/v1/ai/history/anomalies:Anomaly History"
        "/api/v1/ai/realtime/connections:Real-time Connections"
        "/api/v1/ai/alerts:Active Alerts"
    )
    
    local failed=0
    
    for endpoint_info in "${endpoints[@]}"; do
        IFS=':' read -r endpoint description <<< "$endpoint_info"
        
        echo -n "Testing $description... "
        
        if kubectl exec -n $NAMESPACE $POD_NAME -- wget -qO- --timeout=10 "http://localhost:8083$endpoint" >/dev/null 2>&1; then
            echo -e "${GREEN}OK${NC}"
        else
            echo -e "${RED}FAILED${NC}"
            failed=$((failed + 1))
        fi
    done
    
    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}All endpoint tests passed${NC}"
        return 0
    else
        echo -e "${RED}$failed endpoint tests failed${NC}"
        return 1
    fi
}

# Function to show access information
show_access_info() {
    echo
    echo -e "${PURPLE}ACCESS INFORMATION${NC}"
    echo "================================================"
    
    echo -e "${BLUE}AI Service:${NC}"
    echo "  kubectl port-forward -n $NAMESPACE svc/$SERVICE_NAME 8083:8083"
    echo "  http://localhost:8083"
    echo
    
    echo -e "${BLUE}Grafana Dashboard:${NC}"
    echo "  kubectl port-forward -n $MONITORING_NAMESPACE svc/grafana 3000:80"
    echo "  http://localhost:3000 (admin/admin)"
    echo
    
    echo -e "${BLUE}Prometheus:${NC}"
    echo "  kubectl port-forward -n $MONITORING_NAMESPACE svc/prometheus 9090:9090"
    echo "  http://localhost:9090"
    echo
    
    echo -e "${BLUE}API Documentation:${NC}"
    echo "  ./scripts/start-api-docs.sh"
    echo "  http://localhost:8080/ai-api-docs.html"
    echo
}

# Function to show monitoring commands
show_monitoring_commands() {
    echo
    echo -e "${PURPLE}MONITORING COMMANDS${NC}"
    echo "================================================"
    
    echo -e "${BLUE}Check AI Service Status:${NC}"
    echo "  kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=$SERVICE_NAME"
    echo "  kubectl get svc -n $NAMESPACE $SERVICE_NAME"
    echo
    
    echo -e "${BLUE}Check Autoscaling:${NC}"
    echo "  kubectl get hpa -n $NAMESPACE"
    echo "  kubectl get vpa -n $NAMESPACE"
    echo
    
    echo -e "${BLUE}Check Security:${NC}"
    echo "  kubectl get networkpolicy -n $NAMESPACE"
    echo "  kubectl get role,rolebinding,serviceaccount -n $NAMESPACE"
    echo
    
    echo -e "${BLUE}Check Monitoring:${NC}"
    echo "  kubectl get pods -n $MONITORING_NAMESPACE"
    echo "  kubectl get servicemonitor -n $NAMESPACE"
    echo
}

# Main deployment process
main() {
    echo -e "${YELLOW}Starting complete AI service deployment...${NC}"
    echo
    
    # Step 1: Check prerequisites
    print_step "1" "Checking Prerequisites"
    
    echo -e "${YELLOW}Checking namespaces...${NC}"
    check_namespace $NAMESPACE
    check_namespace $MONITORING_NAMESPACE
    
    echo
    echo -e "${YELLOW}Checking Kubernetes cluster...${NC}"
    if kubectl cluster-info >/dev/null 2>&1; then
        echo -e "${GREEN}Kubernetes cluster is accessible${NC}"
    else
        echo -e "${RED}Cannot access Kubernetes cluster${NC}"
        exit 1
    fi
    
    # Step 2: Deploy AI Service
    print_step "2" "Deploying AI Service"
    
    echo -e "${YELLOW}Building and deploying AI service...${NC}"
    
    # Build Docker image
    echo -n "Building Docker image... "
    if docker build -f Dockerfile.ai-service -t cloudbridge/ai-service:latest . --no-cache >/dev/null 2>&1; then
        echo -e "${GREEN}Built${NC}"
    else
        echo -e "${RED}Failed${NC}"
        exit 1
    fi
    
    # Load image to Minikube
    echo -n "Loading image to Minikube... "
    if minikube image load cloudbridge/ai-service:latest >/dev/null 2>&1; then
        echo -e "${GREEN}Loaded${NC}"
    else
        echo -e "${RED}Failed${NC}"
        exit 1
    fi
    
    # Deploy AI service
    echo -n "Deploying AI service... "
    if kubectl apply -f k8s/ai-service/ -n $NAMESPACE >/dev/null 2>&1; then
        echo -e "${GREEN}Deployed${NC}"
    else
        echo -e "${RED}Failed${NC}"
        exit 1
    fi
    
    # Wait for deployment
    wait_for_deployment $SERVICE_NAME $NAMESPACE
    
    # Step 3: Setup Monitoring
    print_step "3" "Setting up Monitoring"
    
    run_script "./scripts/setup-ai-dashboard.sh" "AI Dashboard Setup"
    
    # Step 4: Setup Autoscaling
    print_step "4" "Setting up Autoscaling"
    
    run_script "./scripts/setup-ai-autoscaling.sh" "AI Autoscaling Setup"
    
    # Step 5: Setup Security
    print_step "5" "Setting up Security"
    
    run_script "./scripts/setup-ai-security.sh" "AI Security Setup"
    
    # Step 6: Testing
    print_step "6" "Testing AI Service"
    
    test_ai_service
    
    # Step 7: Final validation
    print_step "7" "Final Validation"
    
    echo -e "${YELLOW}Running comprehensive endpoint tests...${NC}"
    run_script "./scripts/test-ai-endpoints.sh" "AI Endpoints Testing"
    
    # Show results
    echo
    echo -e "${PURPLE}DEPLOYMENT COMPLETE!${NC}"
    echo "================================================"
    
    show_access_info
    show_monitoring_commands
    
    echo -e "${GREEN}AI Service is fully deployed and operational!${NC}"
    echo
    echo -e "${YELLOW}📝 Next Steps:${NC}"
    echo "  1. Access the AI service and test all endpoints"
    echo "  2. Configure Grafana dashboards and alerts"
    echo "  3. Monitor autoscaling behavior"
    echo "  4. Review security policies and RBAC"
    echo "  5. Set up production monitoring and alerting"
    echo
    echo -e "${BLUE}Your AI-powered CloudBridge system is ready!${NC}"
}

# Run main function
main "$@"