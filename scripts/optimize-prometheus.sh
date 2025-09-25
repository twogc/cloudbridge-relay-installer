#!/bin/bash

# CloudBridge Prometheus Optimization Script
# This script applies the optimized Prometheus configuration

set -e

echo "🚀 Starting Prometheus optimization..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is not installed or not in PATH"
    exit 1
fi

# Check if we can connect to the cluster
if ! kubectl cluster-info &> /dev/null; then
    print_error "Cannot connect to Kubernetes cluster"
    exit 1
fi

print_status "Connected to Kubernetes cluster"

# Create namespace if it doesn't exist
print_status "Creating cloudbridge-monitoring namespace if it doesn't exist..."
kubectl create namespace cloudbridge-monitoring --dry-run=client -o yaml | kubectl apply -f -

# Apply the optimized Prometheus configuration
print_status "Applying optimized Prometheus configuration..."
kubectl apply -f k8s/monitoring/prometheus-config.yaml

# Apply the optimized Prometheus deployment
print_status "Applying optimized Prometheus deployment..."
kubectl apply -f k8s/monitoring/prometheus-deployment.yaml

# Wait for the deployment to be ready
print_status "Waiting for Prometheus deployment to be ready..."
kubectl rollout status deployment/prometheus -n cloudbridge-monitoring --timeout=300s

# Check if Prometheus is running
print_status "Checking Prometheus pod status..."
kubectl get pods -n cloudbridge-monitoring -l app=prometheus

# Test Prometheus API
print_status "Testing Prometheus API..."
PROMETHEUS_POD=$(kubectl get pods -n cloudbridge-monitoring -l app=prometheus -o jsonpath='{.items[0].metadata.name}')

if [ -n "$PROMETHEUS_POD" ]; then
    print_status "Prometheus pod: $PROMETHEUS_POD"
    
    # Test basic API endpoint
    print_status "Testing basic API endpoint..."
    kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=10 "http://localhost:9090/api/v1/query?query=up" | head -c 200
    echo ""
    
    # Test metrics endpoint
    print_status "Testing metrics endpoint..."
    kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=10 "http://localhost:9090/metrics" | head -c 200
    echo ""
    
    print_success "Prometheus API is responding"
else
    print_error "Prometheus pod not found"
    exit 1
fi

# Check resource usage
print_status "Checking Prometheus resource usage..."
kubectl top pod -n cloudbridge-monitoring -l app=prometheus

# Check configuration
print_status "Checking Prometheus configuration..."
kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- cat /etc/prometheus/prometheus.yml | head -20

# Test specific metrics queries
print_status "Testing specific metrics queries..."

# Test relay metrics
print_status "Testing relay metrics query..."
kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=15 "http://localhost:9090/api/v1/query?query=relay_active_connections" | head -c 200
echo ""

# Test AI metrics
print_status "Testing AI metrics query..."
kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=15 "http://localhost:9090/api/v1/query?query=ai_requests_processed_total" | head -c 200
echo ""

# Test Cadence metrics
print_status "Testing Cadence metrics query..."
kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=15 "http://localhost:9090/api/v1/query?query=cadence_workflow_executions_total" | head -c 200
echo ""

# Check target status
print_status "Checking Prometheus targets..."
kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=10 "http://localhost:9090/api/v1/targets" | head -c 500
echo ""

print_success "Prometheus optimization completed successfully!"
print_status "Summary of optimizations applied:"
echo "  Increased scrape_interval to 30s"
echo "  Reduced query_timeout to 15s"
echo "  Added metric filtering for essential metrics only"
echo "  Increased CPU resources: 500m request, 2 limit"
echo "  Increased Memory resources: 1Gi request, 4Gi limit"
echo "  Optimized startup parameters"
echo "  Reduced retention to 24h"
echo "  Removed unnecessary Kubernetes monitoring"

print_status "You can now access Prometheus at:"
echo "  kubectl port-forward -n cloudbridge-monitoring svc/prometheus 9090:9090"
echo "  Then open: http://localhost:9090"
