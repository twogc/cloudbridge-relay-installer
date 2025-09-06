#!/bin/bash

# CloudBridge Relay - Performance Analysis Script
# This script analyzes system performance and resource usage

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
RESULTS_DIR="performance-results"

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
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
    
    # Check if metrics-server is available
    if ! kubectl top nodes &> /dev/null; then
        log_warning "Metrics server is not available. Some metrics may not be available."
    fi
    
    log_success "Prerequisites check passed"
}

# Create results directory
create_results_dir() {
    log_info "Creating results directory..."
    
    mkdir -p $RESULTS_DIR
    rm -rf $RESULTS_DIR/*
    
    log_success "Results directory created"
}

# Analyze system resources
analyze_system_resources() {
    log_info "Analyzing system resources..."
    
    # Node resources
    log_info "Collecting node resource information..."
    kubectl get nodes -o wide > $RESULTS_DIR/nodes-info.txt
    kubectl describe nodes > $RESULTS_DIR/nodes-description.txt
    
    # Node resource usage
    if kubectl top nodes &> /dev/null; then
        kubectl top nodes > $RESULTS_DIR/nodes-resource-usage.txt
    else
        echo "Metrics server not available" > $RESULTS_DIR/nodes-resource-usage.txt
    fi
    
    # Pod resources
    log_info "Collecting pod resource information..."
    kubectl get pods -A -o wide > $RESULTS_DIR/pods-info.txt
    
    # Pod resource usage
    if kubectl top pods -A &> /dev/null; then
        kubectl top pods -A > $RESULTS_DIR/pods-resource-usage.txt
    else
        echo "Metrics server not available" > $RESULTS_DIR/pods-resource-usage.txt
    fi
    
    # Resource quotas and limits
    log_info "Collecting resource quotas and limits..."
    kubectl get resourcequotas -A > $RESULTS_DIR/resource-quotas.txt
    kubectl get limitranges -A > $RESULTS_DIR/limit-ranges.txt
    
    log_success "System resources analysis completed"
}

# Analyze CloudBridge services
analyze_cloudbridge_services() {
    log_info "Analyzing CloudBridge services..."
    
    # CloudBridge Relay
    log_info "Analyzing CloudBridge Relay service..."
    kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cloudbridge-relay -o wide > $RESULTS_DIR/relay-pods.txt
    kubectl describe pods -n $NAMESPACE -l app.kubernetes.io/name=cloudbridge-relay > $RESULTS_DIR/relay-pods-description.txt
    
    # Cassandra
    log_info "Analyzing Cassandra service..."
    kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cassandra -o wide > $RESULTS_DIR/cassandra-pods.txt
    kubectl describe pods -n $NAMESPACE -l app.kubernetes.io/name=cassandra > $RESULTS_DIR/cassandra-pods-description.txt
    
    # Cadence
    log_info "Analyzing Cadence service..."
    kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cadence -o wide > $RESULTS_DIR/cadence-pods.txt
    kubectl describe pods -n $NAMESPACE -l app.kubernetes.io/name=cadence > $RESULTS_DIR/cadence-pods-description.txt
    
    # Monitoring services
    log_info "Analyzing monitoring services..."
    kubectl get pods -n $MONITORING_NAMESPACE -o wide > $RESULTS_DIR/monitoring-pods.txt
    kubectl describe pods -n $MONITORING_NAMESPACE > $RESULTS_DIR/monitoring-pods-description.txt
    
    log_success "CloudBridge services analysis completed"
}

# Analyze resource requests and limits
analyze_resource_requests_limits() {
    log_info "Analyzing resource requests and limits..."
    
    # Get resource requests and limits for all pods
    kubectl get pods -A -o jsonpath='{range .items[*]}{.metadata.namespace}{"\t"}{.metadata.name}{"\t"}{.spec.containers[*].resources.requests.cpu}{"\t"}{.spec.containers[*].resources.requests.memory}{"\t"}{.spec.containers[*].resources.limits.cpu}{"\t"}{.spec.containers[*].resources.limits.memory}{"\n"}{end}' > $RESULTS_DIR/resource-requests-limits.txt
    
    # Calculate total resource requests
    log_info "Calculating total resource requests..."
    cat > $RESULTS_DIR/resource-summary.txt << EOF
# Resource Summary

## Total CPU Requests
$(kubectl get pods -A -o jsonpath='{range .items[*]}{.spec.containers[*].resources.requests.cpu}{"\n"}{end}' | grep -v '^$' | sed 's/m//g' | awk '{sum += $1} END {print sum "m"}')

## Total Memory Requests
$(kubectl get pods -A -o jsonpath='{range .items[*]}{.spec.containers[*].resources.requests.memory}{"\n"}{end}' | grep -v '^$' | sed 's/Mi//g' | awk '{sum += $1} END {print sum "Mi"}')

## Total CPU Limits
$(kubectl get pods -A -o jsonpath='{range .items[*]}{.spec.containers[*].resources.limits.cpu}{"\n"}{end}' | grep -v '^$' | sed 's/m//g' | awk '{sum += $1} END {print sum "m"}')

## Total Memory Limits
$(kubectl get pods -A -o jsonpath='{range .items[*]}{.spec.containers[*].resources.limits.memory}{"\n"}{end}' | grep -v '^$' | sed 's/Mi//g' | awk '{sum += $1} END {print sum "Mi"}')

EOF
    
    log_success "Resource requests and limits analysis completed"
}

# Analyze performance metrics
analyze_performance_metrics() {
    log_info "Analyzing performance metrics..."
    
    # Check if Prometheus is available
    if kubectl get pods -n $MONITORING_NAMESPACE -l app.kubernetes.io/name=prometheus &> /dev/null; then
        log_info "Prometheus is available, collecting metrics..."
        
        # Port forward to Prometheus
        kubectl port-forward -n $MONITORING_NAMESPACE service/prometheus 9090:9090 &
        PROMETHEUS_PID=$!
        sleep 5
        
        # Collect metrics
        curl -s "http://localhost:9090/api/v1/query?query=up" > $RESULTS_DIR/prometheus-up-metrics.json
        curl -s "http://localhost:9090/api/v1/query?query=rate(process_cpu_seconds_total[5m])" > $RESULTS_DIR/prometheus-cpu-metrics.json
        curl -s "http://localhost:9090/api/v1/query?query=process_resident_memory_bytes" > $RESULTS_DIR/prometheus-memory-metrics.json
        
        # Kill port forward
        kill $PROMETHEUS_PID 2>/dev/null || true
        
        log_success "Performance metrics collected"
    else
        log_warning "Prometheus is not available, skipping metrics collection"
        echo "Prometheus not available" > $RESULTS_DIR/prometheus-metrics.txt
    fi
}

# Analyze logs for performance issues
analyze_logs() {
    log_info "Analyzing logs for performance issues..."
    
    # CloudBridge Relay logs
    if kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cloudbridge-relay &> /dev/null; then
        kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=cloudbridge-relay --tail=100 > $RESULTS_DIR/relay-logs.txt
    fi
    
    # Cassandra logs
    if kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cassandra &> /dev/null; then
        kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=cassandra --tail=100 > $RESULTS_DIR/cassandra-logs.txt
    fi
    
    # Cadence logs
    if kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cadence &> /dev/null; then
        kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=cadence --tail=100 > $RESULTS_DIR/cadence-logs.txt
    fi
    
    log_success "Logs analysis completed"
}

# Generate performance report
generate_performance_report() {
    log_info "Generating performance report..."
    
    cat > $RESULTS_DIR/performance-report.md << EOF
# CloudBridge Relay - Performance Analysis Report

## Analysis Date
$(date)

## System Overview
$(cat $RESULTS_DIR/nodes-info.txt)

## Resource Usage Summary
$(cat $RESULTS_DIR/resource-summary.txt)

## Node Resource Usage
$(cat $RESULTS_DIR/nodes-resource-usage.txt)

## Pod Resource Usage
$(cat $RESULTS_DIR/pods-resource-usage.txt)

## CloudBridge Services Status
- **Relay Service**: $(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cloudbridge-relay --no-headers | wc -l) pods
- **Cassandra**: $(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cassandra --no-headers | wc -l) pods
- **Cadence**: $(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cadence --no-headers | wc -l) pods
- **Monitoring**: $(kubectl get pods -n $MONITORING_NAMESPACE --no-headers | wc -l) pods

## Resource Requests and Limits
$(cat $RESULTS_DIR/resource-requests-limits.txt)

## Recommendations
1. **CPU Usage**: Monitor CPU usage and consider increasing limits if consistently high
2. **Memory Usage**: Monitor memory usage and consider increasing limits if consistently high
3. **Storage**: Monitor disk usage and consider increasing storage if needed
4. **Network**: Monitor network usage and consider optimizing if needed
5. **Scaling**: Consider horizontal pod autoscaling if load is high

## Files Generated
$(ls -la $RESULTS_DIR/)

EOF
    
    log_success "Performance report generated: $RESULTS_DIR/performance-report.md"
}

# Show performance summary
show_performance_summary() {
    log_info "Performance Analysis Summary:"
    echo ""
    
    echo "📊 System Resources:"
    if [ -f "$RESULTS_DIR/nodes-resource-usage.txt" ]; then
        cat $RESULTS_DIR/nodes-resource-usage.txt
    fi
    
    echo ""
    echo "📊 Pod Resources:"
    if [ -f "$RESULTS_DIR/pods-resource-usage.txt" ]; then
        cat $RESULTS_DIR/pods-resource-usage.txt
    fi
    
    echo ""
    echo "📊 Resource Summary:"
    if [ -f "$RESULTS_DIR/resource-summary.txt" ]; then
        cat $RESULTS_DIR/resource-summary.txt
    fi
    
    echo ""
    echo "📊 Service Status:"
    echo "• CloudBridge Relay: $(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cloudbridge-relay --no-headers | wc -l) pods"
    echo "• Cassandra: $(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cassandra --no-headers | wc -l) pods"
    echo "• Cadence: $(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=cadence --no-headers | wc -l) pods"
    echo "• Monitoring: $(kubectl get pods -n $MONITORING_NAMESPACE --no-headers | wc -l) pods"
    
    echo ""
    echo "📋 Files Generated:"
    ls -la $RESULTS_DIR/
}

# Cleanup old results
cleanup() {
    log_warning "Cleaning up old performance results..."
    
    rm -rf $RESULTS_DIR/*
    
    log_success "Cleanup completed"
}

# Main function
main() {
    echo "=== 📊 CloudBridge Relay - Performance Analysis ==="
    echo ""
    
    case "${1:-analyze}" in
        "analyze")
            check_prerequisites
            create_results_dir
            analyze_system_resources
            analyze_cloudbridge_services
            analyze_resource_requests_limits
            analyze_performance_metrics
            analyze_logs
            generate_performance_report
            show_performance_summary
            log_success "Performance analysis completed successfully!"
            ;;
        "summary")
            show_performance_summary
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [analyze|summary|cleanup|help]"
            echo ""
            echo "Commands:"
            echo "  analyze  - Run full performance analysis (default)"
            echo "  summary  - Show performance summary"
            echo "  cleanup  - Clean up old results"
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

