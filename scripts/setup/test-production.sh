#!/bin/bash

# CloudBridge Relay - Production Testing Script
# This script tests the production deployment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="cloudbridge"
MONITORING_NAMESPACE="cloudbridge-monitoring"
TEST_RESULTS_DIR="test-results"

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}$1${NC}"
}

log_warning() {
    echo -e "${YELLOW}  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if curl is installed
    if ! command -v curl &> /dev/null; then
        log_error "curl is not installed. Please install curl first."
        exit 1
    fi
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        log_warning "jq is not installed. Installing jq..."
        sudo apt update && sudo apt install -y jq
    fi
    
    log_success "Prerequisites check passed"
}

# Create test results directory
create_test_results_dir() {
    log_info "Creating test results directory..."
    
    mkdir -p $TEST_RESULTS_DIR
    rm -rf $TEST_RESULTS_DIR/*
    
    log_success "Test results directory created"
}

# Test Kubernetes resources
test_kubernetes_resources() {
    log_info "Testing Kubernetes resources..."
    
    # Test namespaces
    log_info "Testing namespaces..."
    kubectl get namespace $NAMESPACE > $TEST_RESULTS_DIR/namespace-cloudbridge.txt 2>/dev/null || echo "Namespace $NAMESPACE not found" > $TEST_RESULTS_DIR/namespace-cloudbridge.txt
    kubectl get namespace $MONITORING_NAMESPACE > $TEST_RESULTS_DIR/namespace-monitoring.txt 2>/dev/null || echo "Namespace $MONITORING_NAMESPACE not found" > $TEST_RESULTS_DIR/namespace-monitoring.txt
    
    # Test pods
    log_info "Testing pods..."
    kubectl get pods -n $NAMESPACE > $TEST_RESULTS_DIR/pods-cloudbridge.txt 2>/dev/null || echo "No pods in $NAMESPACE namespace" > $TEST_RESULTS_DIR/pods-cloudbridge.txt
    kubectl get pods -n $MONITORING_NAMESPACE > $TEST_RESULTS_DIR/pods-monitoring.txt 2>/dev/null || echo "No pods in $MONITORING_NAMESPACE namespace" > $TEST_RESULTS_DIR/pods-monitoring.txt
    
    # Test services
    log_info "Testing services..."
    kubectl get services -n $NAMESPACE > $TEST_RESULTS_DIR/services-cloudbridge.txt 2>/dev/null || echo "No services in $NAMESPACE namespace" > $TEST_RESULTS_DIR/services-cloudbridge.txt
    kubectl get services -n $MONITORING_NAMESPACE > $TEST_RESULTS_DIR/services-monitoring.txt 2>/dev/null || echo "No services in $MONITORING_NAMESPACE namespace" > $TEST_RESULTS_DIR/services-monitoring.txt
    
    # Test deployments
    log_info "Testing deployments..."
    kubectl get deployments -n $NAMESPACE > $TEST_RESULTS_DIR/deployments-cloudbridge.txt 2>/dev/null || echo "No deployments in $NAMESPACE namespace" > $TEST_RESULTS_DIR/deployments-cloudbridge.txt
    kubectl get deployments -n $MONITORING_NAMESPACE > $TEST_RESULTS_DIR/deployments-monitoring.txt 2>/dev/null || echo "No deployments in $MONITORING_NAMESPACE namespace" > $TEST_RESULTS_DIR/deployments-monitoring.txt
    
    # Test persistent volumes
    log_info "Testing persistent volumes..."
    kubectl get pv,pvc -n $NAMESPACE > $TEST_RESULTS_DIR/volumes-cloudbridge.txt 2>/dev/null || echo "No volumes in $NAMESPACE namespace" > $TEST_RESULTS_DIR/volumes-cloudbridge.txt
    kubectl get pv,pvc -n $MONITORING_NAMESPACE > $TEST_RESULTS_DIR/volumes-monitoring.txt 2>/dev/null || echo "No volumes in $MONITORING_NAMESPACE namespace" > $TEST_RESULTS_DIR/volumes-monitoring.txt
    
    log_success "Kubernetes resources test completed"
}

# Test pod health
test_pod_health() {
    log_info "Testing pod health..."
    
    # Check pod status
    log_info "Checking pod status..."
    kubectl get pods -n $NAMESPACE -o json > $TEST_RESULTS_DIR/pods-status-cloudbridge.json 2>/dev/null || echo "{}" > $TEST_RESULTS_DIR/pods-status-cloudbridge.json
    kubectl get pods -n $MONITORING_NAMESPACE -o json > $TEST_RESULTS_DIR/pods-status-monitoring.json 2>/dev/null || echo "{}" > $TEST_RESULTS_DIR/pods-status-monitoring.json
    
    # Check for failed pods
    log_info "Checking for failed pods..."
    kubectl get pods -n $NAMESPACE --field-selector=status.phase!=Running > $TEST_RESULTS_DIR/failed-pods-cloudbridge.txt 2>/dev/null || echo "No failed pods in $NAMESPACE namespace" > $TEST_RESULTS_DIR/failed-pods-cloudbridge.txt
    kubectl get pods -n $MONITORING_NAMESPACE --field-selector=status.phase!=Running > $TEST_RESULTS_DIR/failed-pods-monitoring.txt 2>/dev/null || echo "No failed pods in $MONITORING_NAMESPACE namespace" > $TEST_RESULTS_DIR/failed-pods-monitoring.txt
    
    # Check pod logs for errors
    log_info "Checking pod logs for errors..."
    if kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cloudbridge-relay &> /dev/null; then
        kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=cloudbridge-relay --tail=50 > $TEST_RESULTS_DIR/relay-logs.txt 2>/dev/null || echo "No logs available" > $TEST_RESULTS_DIR/relay-logs.txt
    fi
    
    if kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cassandra &> /dev/null; then
        kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=cassandra --tail=50 > $TEST_RESULTS_DIR/cassandra-logs.txt 2>/dev/null || echo "No logs available" > $TEST_RESULTS_DIR/cassandra-logs.txt
    fi
    
    log_success "Pod health test completed"
}

# Test service connectivity
test_service_connectivity() {
    log_info "Testing service connectivity..."
    
    # Test CloudBridge Relay service
    log_info "Testing CloudBridge Relay service..."
    if kubectl get service cloudbridge-relay -n $NAMESPACE &> /dev/null; then
        # Port forward to test connectivity
        kubectl port-forward service/cloudbridge-relay 9090:9090 -n $NAMESPACE &
        PORT_FORWARD_PID=$!
        sleep 5
        
        # Test health endpoint
        if curl -f http://localhost:9090/health > $TEST_RESULTS_DIR/relay-health-test.txt 2>/dev/null; then
            echo "Health check: PASSED" >> $TEST_RESULTS_DIR/relay-health-test.txt
        else
            echo "Health check: FAILED" >> $TEST_RESULTS_DIR/relay-health-test.txt
        fi
        
        # Test metrics endpoint
        if curl -f http://localhost:8084/metrics > $TEST_RESULTS_DIR/relay-metrics-test.txt 2>/dev/null; then
            echo "Metrics endpoint: PASSED" >> $TEST_RESULTS_DIR/relay-metrics-test.txt
        else
            echo "Metrics endpoint: FAILED" >> $TEST_RESULTS_DIR/relay-metrics-test.txt
        fi
        
        # Kill port forward
        kill $PORT_FORWARD_PID 2>/dev/null || true
    else
        echo "CloudBridge Relay service not found" > $TEST_RESULTS_DIR/relay-health-test.txt
    fi
    
    # Test Grafana service
    log_info "Testing Grafana service..."
    if kubectl get service grafana -n $MONITORING_NAMESPACE &> /dev/null; then
        # Port forward to test connectivity
        kubectl port-forward service/grafana 3000:3000 -n $MONITORING_NAMESPACE &
        PORT_FORWARD_PID=$!
        sleep 5
        
        # Test Grafana endpoint
        if curl -f http://localhost:3000/api/health > $TEST_RESULTS_DIR/grafana-health-test.txt 2>/dev/null; then
            echo "Grafana health check: PASSED" >> $TEST_RESULTS_DIR/grafana-health-test.txt
        else
            echo "Grafana health check: FAILED" >> $TEST_RESULTS_DIR/grafana-health-test.txt
        fi
        
        # Kill port forward
        kill $PORT_FORWARD_PID 2>/dev/null || true
    else
        echo "Grafana service not found" > $TEST_RESULTS_DIR/grafana-health-test.txt
    fi
    
    # Test Prometheus service
    log_info "Testing Prometheus service..."
    if kubectl get service prometheus -n $MONITORING_NAMESPACE &> /dev/null; then
        # Port forward to test connectivity
        kubectl port-forward service/prometheus 9090:9090 -n $MONITORING_NAMESPACE &
        PORT_FORWARD_PID=$!
        sleep 5
        
        # Test Prometheus endpoint
        if curl -f http://localhost:9090/-/healthy > $TEST_RESULTS_DIR/prometheus-health-test.txt 2>/dev/null; then
            echo "Prometheus health check: PASSED" >> $TEST_RESULTS_DIR/prometheus-health-test.txt
        else
            echo "Prometheus health check: FAILED" >> $TEST_RESULTS_DIR/prometheus-health-test.txt
        fi
        
        # Kill port forward
        kill $PORT_FORWARD_PID 2>/dev/null || true
    else
        echo "Prometheus service not found" > $TEST_RESULTS_DIR/prometheus-health-test.txt
    fi
    
    log_success "Service connectivity test completed"
}

# Test resource usage
test_resource_usage() {
    log_info "Testing resource usage..."
    
    # Check node resources
    log_info "Checking node resources..."
    kubectl top nodes > $TEST_RESULTS_DIR/node-resources.txt 2>/dev/null || echo "Metrics server not available" > $TEST_RESULTS_DIR/node-resources.txt
    
    # Check pod resources
    log_info "Checking pod resources..."
    kubectl top pods -n $NAMESPACE > $TEST_RESULTS_DIR/pod-resources-cloudbridge.txt 2>/dev/null || echo "Metrics server not available" > $TEST_RESULTS_DIR/pod-resources-cloudbridge.txt
    kubectl top pods -n $MONITORING_NAMESPACE > $TEST_RESULTS_DIR/pod-resources-monitoring.txt 2>/dev/null || echo "Metrics server not available" > $TEST_RESULTS_DIR/pod-resources-monitoring.txt
    
    # Check resource quotas
    log_info "Checking resource quotas..."
    kubectl get resourcequotas -n $NAMESPACE > $TEST_RESULTS_DIR/resource-quotas-cloudbridge.txt 2>/dev/null || echo "No resource quotas in $NAMESPACE namespace" > $TEST_RESULTS_DIR/resource-quotas-cloudbridge.txt
    kubectl get resourcequotas -n $MONITORING_NAMESPACE > $TEST_RESULTS_DIR/resource-quotas-monitoring.txt 2>/dev/null || echo "No resource quotas in $MONITORING_NAMESPACE namespace" > $TEST_RESULTS_DIR/resource-quotas-monitoring.txt
    
    log_success "Resource usage test completed"
}

# Test security policies
test_security_policies() {
    log_info "Testing security policies..."
    
    # Check network policies
    log_info "Checking network policies..."
    kubectl get networkpolicies -n $NAMESPACE > $TEST_RESULTS_DIR/network-policies-cloudbridge.txt 2>/dev/null || echo "No network policies in $NAMESPACE namespace" > $TEST_RESULTS_DIR/network-policies-cloudbridge.txt
    kubectl get networkpolicies -n $MONITORING_NAMESPACE > $TEST_RESULTS_DIR/network-policies-monitoring.txt 2>/dev/null || echo "No network policies in $MONITORING_NAMESPACE namespace" > $TEST_RESULTS_DIR/network-policies-monitoring.txt
    
    # Check pod security policies
    log_info "Checking pod security policies..."
    kubectl get podsecuritypolicies > $TEST_RESULTS_DIR/pod-security-policies.txt 2>/dev/null || echo "No pod security policies found" > $TEST_RESULTS_DIR/pod-security-policies.txt
    
    # Check RBAC
    log_info "Checking RBAC..."
    kubectl get clusterroles | grep cloudbridge > $TEST_RESULTS_DIR/cluster-roles-cloudbridge.txt 2>/dev/null || echo "No CloudBridge cluster roles found" > $TEST_RESULTS_DIR/cluster-roles-cloudbridge.txt
    kubectl get clusterrolebindings | grep cloudbridge > $TEST_RESULTS_DIR/cluster-role-bindings-cloudbridge.txt 2>/dev/null || echo "No CloudBridge cluster role bindings found" > $TEST_RESULTS_DIR/cluster-role-bindings-cloudbridge.txt
    
    log_success "Security policies test completed"
}

# Generate test report
generate_test_report() {
    log_info "Generating test report..."
    
    cat > $TEST_RESULTS_DIR/test-report.md << EOF
# CloudBridge Relay - Production Test Report

## Test Date
$(date)

## Test Summary
This report contains the results of production deployment tests for CloudBridge Relay.

## Tests Performed
1. Kubernetes Resources Test
2. Pod Health Test
3. Service Connectivity Test
4. Resource Usage Test
5. Security Policies Test

## Test Results

### Namespaces
- **CloudBridge Namespace**: $(cat $TEST_RESULTS_DIR/namespace-cloudbridge.txt | head -1)
- **Monitoring Namespace**: $(cat $TEST_RESULTS_DIR/namespace-monitoring.txt | head -1)

### Pods Status
- **CloudBridge Pods**: $(cat $TEST_RESULTS_DIR/pods-cloudbridge.txt | wc -l) pods
- **Monitoring Pods**: $(cat $TEST_RESULTS_DIR/pods-monitoring.txt | wc -l) pods

### Services Status
- **CloudBridge Services**: $(cat $TEST_RESULTS_DIR/services-cloudbridge.txt | wc -l) services
- **Monitoring Services**: $(cat $TEST_RESULTS_DIR/services-monitoring.txt | wc -l) services

### Health Checks
- **Relay Health**: $(cat $TEST_RESULTS_DIR/relay-health-test.txt | tail -1)
- **Grafana Health**: $(cat $TEST_RESULTS_DIR/grafana-health-test.txt | tail -1)
- **Prometheus Health**: $(cat $TEST_RESULTS_DIR/prometheus-health-test.txt | tail -1)

### Resource Usage
- **Node Resources**: $(cat $TEST_RESULTS_DIR/node-resources.txt | wc -l) nodes
- **Pod Resources**: $(cat $TEST_RESULTS_DIR/pod-resources-cloudbridge.txt | wc -l) pods

### Security
- **Network Policies**: $(cat $TEST_RESULTS_DIR/network-policies-cloudbridge.txt | wc -l) policies
- **Pod Security Policies**: $(cat $TEST_RESULTS_DIR/pod-security-policies.txt | wc -l) policies

## Files Generated
$(ls -la $TEST_RESULTS_DIR/)

## Recommendations
1. Review any failed health checks
2. Monitor resource usage and adjust limits if needed
3. Verify security policies are properly applied
4. Check logs for any errors or warnings
5. Test all endpoints and functionality

EOF
    
    log_success "Test report generated: $TEST_RESULTS_DIR/test-report.md"
}

# Show test summary
show_test_summary() {
    log_info "Production Test Summary:"
    echo ""
    
    echo "📊 Test Results:"
    echo "• Namespaces: $(cat $TEST_RESULTS_DIR/namespace-cloudbridge.txt | head -1 | cut -d' ' -f1)"
    echo "• CloudBridge Pods: $(cat $TEST_RESULTS_DIR/pods-cloudbridge.txt | wc -l) pods"
    echo "• Monitoring Pods: $(cat $TEST_RESULTS_DIR/pods-monitoring.txt | wc -l) pods"
    echo "• CloudBridge Services: $(cat $TEST_RESULTS_DIR/services-cloudbridge.txt | wc -l) services"
    echo "• Monitoring Services: $(cat $TEST_RESULTS_DIR/services-monitoring.txt | wc -l) services"
    echo ""
    
    echo "🔍 Health Checks:"
    echo "• Relay Health: $(cat $TEST_RESULTS_DIR/relay-health-test.txt | tail -1)"
    echo "• Grafana Health: $(cat $TEST_RESULTS_DIR/grafana-health-test.txt | tail -1)"
    echo "• Prometheus Health: $(cat $TEST_RESULTS_DIR/prometheus-health-test.txt | tail -1)"
    echo ""
    
    echo "Resource Usage:"
    if [ -f "$TEST_RESULTS_DIR/node-resources.txt" ]; then
        cat $TEST_RESULTS_DIR/node-resources.txt
    fi
    echo ""
    
    echo "Pod Resources:"
    if [ -f "$TEST_RESULTS_DIR/pod-resources-cloudbridge.txt" ]; then
        cat $TEST_RESULTS_DIR/pod-resources-cloudbridge.txt
    fi
    echo ""
    
    echo "Files Generated:"
    ls -la $TEST_RESULTS_DIR/
}

# Cleanup test results
cleanup() {
    log_warning "Cleaning up test results..."
    
    rm -rf $TEST_RESULTS_DIR/*
    
    log_success "Cleanup completed"
}

# Main function
main() {
    echo "=== 🧪 CloudBridge Relay - Production Testing ==="
    echo ""
    
    case "${1:-test}" in
        "test")
            check_prerequisites
            create_test_results_dir
            test_kubernetes_resources
            test_pod_health
            test_service_connectivity
            test_resource_usage
            test_security_policies
            generate_test_report
            show_test_summary
            log_success "Production testing completed successfully!"
            ;;
        "summary")
            show_test_summary
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [test|summary|cleanup|help]"
            echo ""
            echo "Commands:"
            echo "  test     - Run full production test (default)"
            echo "  summary  - Show test summary"
            echo "  cleanup  - Clean up test results"
            echo "  help     - Show this help message"
            ;;
        *)
            log_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"



