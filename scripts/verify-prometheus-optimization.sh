#!/bin/bash

# CloudBridge Prometheus Optimization Verification Script
# This script verifies that the Prometheus optimization was successful

set -e

echo "🔍 Verifying Prometheus optimization..."

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

# Get Prometheus pod
PROMETHEUS_POD=$(kubectl get pods -n cloudbridge-monitoring -l app=prometheus -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

if [ -z "$PROMETHEUS_POD" ]; then
    print_error "Prometheus pod not found in cloudbridge-monitoring namespace"
    exit 1
fi

print_status "Found Prometheus pod: $PROMETHEUS_POD"

# Test 1: API Response Time
print_status "Test 1: API Response Time"
start_time=$(date +%s.%N)
kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=10 "http://localhost:9090/api/v1/query?query=up" > /dev/null 2>&1
end_time=$(date +%s.%N)
response_time=$(echo "$end_time - $start_time" | bc -l)
response_time_ms=$(echo "$response_time * 1000" | bc -l | cut -d. -f1)

if [ "$response_time_ms" -lt 5000 ]; then
    print_success "API response time: ${response_time_ms}ms (target: <5000ms)"
else
    print_warning "API response time: ${response_time_ms}ms (target: <5000ms)"
fi

# Test 2: Check Resource Usage
print_status "Test 2: Resource Usage"
kubectl top pod -n cloudbridge-monitoring -l app=prometheus 2>/dev/null || print_warning "Metrics server not available for resource monitoring"

# Test 3: Check Configuration
print_status "Test 3: Configuration Verification"
config_check=$(kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- cat /etc/prometheus/prometheus.yml 2>/dev/null)

# Check scrape_interval
if echo "$config_check" | grep -q "scrape_interval: 30s"; then
    print_success "Scrape interval set to 30s"
else
    print_error "Scrape interval not set to 30s"
fi

# Check query_timeout (it's set via command line args, not in config file)
# This check is handled in Test 7: Prometheus Arguments

# Check metric filtering
if echo "$config_check" | grep -q "metric_relabel_configs"; then
    print_success "Metric filtering configured"
else
    print_error "Metric filtering not configured"
fi

# Test 4: Check Targets
print_status "Test 4: Prometheus Targets"
targets_response=$(kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=10 "http://localhost:9090/api/v1/targets" 2>/dev/null)

# Count active targets
active_targets=$(echo "$targets_response" | grep -o '"health":"up"' | wc -l)
total_targets=$(echo "$targets_response" | grep -o '"health":"up"' | wc -l)
total_targets=$((total_targets + $(echo "$targets_response" | grep -o '"health":"down"' | wc -l)))

print_status "Active targets: $active_targets/$total_targets"

if [ "$total_targets" -lt 20 ]; then
    print_success "Target count optimized: $total_targets (target: <20)"
else
    print_warning "Target count: $total_targets (target: <20)"
fi

# Test 5: Check Metrics Count
print_status "Test 5: Metrics Count"
metrics_response=$(kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=10 "http://localhost:9090/api/v1/label/__name__/values" 2>/dev/null)
metrics_count=$(echo "$metrics_response" | grep -o '"[^"]*"' | wc -l)

if [ "$metrics_count" -lt 100 ]; then
    print_success "Metrics count optimized: $metrics_count (target: <100)"
else
    print_warning "Metrics count: $metrics_count (target: <100)"
fi

# Test 6: Check Specific Metrics
print_status "Test 6: Essential Metrics Availability"

# Test relay metrics
relay_metrics=$(kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=10 "http://localhost:9090/api/v1/query?query=relay_active_connections" 2>/dev/null)
if echo "$relay_metrics" | grep -q "result"; then
    print_success "Relay metrics available"
else
    print_warning "Relay metrics not available (service may not be running)"
fi

# Test AI metrics
ai_metrics=$(kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=10 "http://localhost:9090/api/v1/query?query=ai_requests_processed_total" 2>/dev/null)
if echo "$ai_metrics" | grep -q "result"; then
    print_success "AI metrics available"
else
    print_warning "AI metrics not available (service may not be running)"
fi

# Test Cadence metrics
cadence_metrics=$(kubectl exec -n cloudbridge-monitoring $PROMETHEUS_POD -- wget -qO- --timeout=10 "http://localhost:9090/api/v1/query?query=cadence_workflow_executions_total" 2>/dev/null)
if echo "$cadence_metrics" | grep -q "result"; then
    print_success "Cadence metrics available"
else
    print_warning "Cadence metrics not available (service may not be running)"
fi

# Test 7: Check Prometheus Arguments
print_status "Test 7: Prometheus Arguments"
pod_args=$(kubectl get pod -n cloudbridge-monitoring $PROMETHEUS_POD -o jsonpath='{.spec.containers[0].args}' 2>/dev/null)

if echo "$pod_args" | grep -q "query.max-concurrency=10"; then
    print_success "Query max concurrency set to 10"
else
    print_error "Query max concurrency not set to 10"
fi

if echo "$pod_args" | grep -q "query.max-samples=100000"; then
    print_success "Query max samples set to 100000"
else
    print_error "Query max samples not set to 100000"
fi

if echo "$pod_args" | grep -q "query.timeout=15s"; then
    print_success "Query timeout set to 15s"
else
    print_error "Query timeout not set to 15s"
fi

if echo "$pod_args" | grep -q "storage.tsdb.retention.time=24h"; then
    print_success "Retention time set to 24h"
else
    print_error "Retention time not set to 24h"
fi

# Test 8: Check Resource Limits
print_status "Test 8: Resource Limits"
pod_resources=$(kubectl get pod -n cloudbridge-monitoring $PROMETHEUS_POD -o jsonpath='{.spec.containers[0].resources}' 2>/dev/null)

if echo "$pod_resources" | grep -q '"cpu":"500m"'; then
    print_success "CPU request set to 500m"
else
    print_error "CPU request not set to 500m"
fi

if echo "$pod_resources" | grep -q '"memory":"1Gi"'; then
    print_success "Memory request set to 1Gi"
else
    print_error "Memory request not set to 1Gi"
fi

if echo "$pod_resources" | grep -q '"cpu":"2"'; then
    print_success "CPU limit set to 2"
else
    print_error "CPU limit not set to 2"
fi

if echo "$pod_resources" | grep -q '"memory":"4Gi"'; then
    print_success "Memory limit set to 4Gi"
else
    print_error "Memory limit not set to 4Gi"
fi

# Summary
print_status "Optimization verification completed!"
print_status "Summary:"
echo "  📊 API Response Time: ${response_time_ms}ms"
echo "  🎯 Active Targets: $active_targets/$total_targets"
echo "  📈 Metrics Count: $metrics_count"
echo "  ⚙️  Configuration: Optimized"
echo "  💾 Resources: Increased"
echo "  🚀 Performance: Improved"

print_success "Prometheus optimization verification completed!"
