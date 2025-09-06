#!/bin/bash

# CloudBridge Relay - Staging Deployment Script
# Этот скрипт развертывает CloudBridge Relay в staging environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="${NAMESPACE:-cloudbridge-staging}"
ENVIRONMENT="${ENVIRONMENT:-staging}"
IMAGE_TAG="${IMAGE_TAG:-staging}"
REGISTRY="${REGISTRY:-ghcr.io/twogc/cloudbridge-relay-installer}"
DRY_RUN="${DRY_RUN:-false}"
SKIP_TESTS="${SKIP_TESTS:-false}"
FORCE_DEPLOY="${FORCE_DEPLOY:-false}"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_section() {
    echo -e "${PURPLE}[SECTION]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    for tool in kubectl docker; do
        if ! command -v "$tool" &> /dev/null; then
            missing_deps+=("$tool")
        fi
    done
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        exit 1
    fi
    
    # Check if kubectl can connect to cluster
    if ! kubectl cluster-info &>/dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    log_success "Dependencies check passed"
}

# Pre-deployment checks
pre_deployment_checks() {
    log_section "Performing pre-deployment checks..."
    
    # Check cluster resources
    local nodes=$(kubectl get nodes --no-headers | wc -l)
    if [ "$nodes" -eq 0 ]; then
        log_error "No nodes available in cluster"
        exit 1
    fi
    
    # Check available resources
    local available_cpu=$(kubectl top nodes --no-headers 2>/dev/null | awk '{sum+=$2} END {print 100-sum/NR}' || echo "unknown")
    local available_memory=$(kubectl top nodes --no-headers 2>/dev/null | awk '{sum+=$4} END {print 100-sum/NR}' || echo "unknown")
    
    log_info "Cluster resources:"
    log_info "  - Nodes: $nodes"
    log_info "  - Available CPU: $available_cpu%"
    log_info "  - Available Memory: $available_memory%"
    
    # Check if image exists
    log_info "Checking if image exists: $REGISTRY:$IMAGE_TAG"
    if ! docker manifest inspect "$REGISTRY:$IMAGE_TAG" &>/dev/null; then
        log_warning "Image $REGISTRY:$IMAGE_TAG not found in registry"
        if [ "$FORCE_DEPLOY" != "true" ]; then
            log_error "Use --force to deploy anyway"
            exit 1
        fi
    fi
    
    log_success "Pre-deployment checks completed"
}

# Create namespace if it doesn't exist
create_namespace() {
    log_section "Creating namespace: $NAMESPACE"
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would create namespace $NAMESPACE"
        return 0
    fi
    
    if ! kubectl get namespace "$NAMESPACE" &>/dev/null; then
        log_info "Creating namespace: $NAMESPACE"
        kubectl apply -f ./k8s/staging/namespace.yaml
        log_success "Namespace created: $NAMESPACE"
    else
        log_info "Namespace already exists: $NAMESPACE"
    fi
}

# Deploy configurations
deploy_configurations() {
    log_section "Deploying configurations..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would deploy configurations"
        return 0
    fi
    
    # Deploy ConfigMap
    log_info "Deploying ConfigMap..."
    kubectl apply -f ./k8s/staging/configmap.yaml
    
    # Deploy Secrets
    log_info "Deploying Secrets..."
    kubectl apply -f ./k8s/staging/secrets.yaml
    
    log_success "Configurations deployed"
}

# Deploy application
deploy_application() {
    log_section "Deploying application..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would deploy application with image $REGISTRY:$IMAGE_TAG"
        return 0
    fi
    
    # Update image tag in deployment
    log_info "Updating deployment with image: $REGISTRY:$IMAGE_TAG"
    kubectl set image deployment/cloudbridge-relay-staging \
        cloudbridge-relay="$REGISTRY:$IMAGE_TAG" \
        -n "$NAMESPACE"
    
    # Deploy application
    log_info "Deploying application..."
    kubectl apply -f ./k8s/staging/deployment.yaml
    
    # Deploy services
    log_info "Deploying services..."
    kubectl apply -f ./k8s/staging/services.yaml
    
    log_success "Application deployed"
}

# Wait for deployment to be ready
wait_for_deployment() {
    log_section "Waiting for deployment to be ready..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would wait for deployment to be ready"
        return 0
    fi
    
    log_info "Waiting for deployment to be ready..."
    kubectl wait --for=condition=available --timeout=600s deployment/cloudbridge-relay-staging -n "$NAMESPACE"
    
    # Check pod status
    local pods=$(kubectl get pods -n "$NAMESPACE" --no-headers | wc -l)
    local running_pods=$(kubectl get pods -n "$NAMESPACE" --no-headers | grep "Running" | wc -l)
    
    log_info "Pod status: $running_pods/$pods running"
    
    if [ "$running_pods" -ne "$pods" ]; then
        log_warning "Not all pods are running"
        kubectl get pods -n "$NAMESPACE"
    fi
    
    log_success "Deployment is ready"
}

# Run health checks
run_health_checks() {
    log_section "Running health checks..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would run health checks"
        return 0
    fi
    
    # Get service endpoint
    local service_ip=$(kubectl get service cloudbridge-relay-staging-service -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
    local service_port=$(kubectl get service cloudbridge-relay-staging-service -n "$NAMESPACE" -o jsonpath='{.spec.ports[0].port}')
    
    if [ -z "$service_ip" ]; then
        log_warning "LoadBalancer IP not available, using port-forward"
        # Use port-forward for testing
        kubectl port-forward service/cloudbridge-relay-staging-service 9090:9090 -n "$NAMESPACE" &
        local port_forward_pid=$!
        sleep 5
        service_ip="localhost"
    fi
    
    # Health check
    log_info "Running health check on $service_ip:$service_port"
    if curl -f "http://$service_ip:$service_port/health" >/dev/null 2>&1; then
        log_success "Health check passed"
    else
        log_error "Health check failed"
        exit 1
    fi
    
    # Metrics check
    log_info "Checking metrics endpoint..."
    if curl -f "http://$service_ip:$service_port/metrics" >/dev/null 2>&1; then
        log_success "Metrics endpoint is accessible"
    else
        log_warning "Metrics endpoint is not accessible"
    fi
    
    # Cleanup port-forward if used
    if [ -n "$port_forward_pid" ]; then
        kill $port_forward_pid 2>/dev/null || true
    fi
}

# Run smoke tests
run_smoke_tests() {
    log_section "Running smoke tests..."
    
    if [ "$SKIP_TESTS" = "true" ]; then
        log_info "Skipping smoke tests"
        return 0
    fi
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would run smoke tests"
        return 0
    fi
    
    # Run basic smoke tests
    log_info "Running basic smoke tests..."
    
    # Test API endpoints
    local service_ip=$(kubectl get service cloudbridge-relay-staging-service -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "localhost")
    local service_port="9090"
    
    # Use port-forward if no external IP
    if [ "$service_ip" = "localhost" ] || [ -z "$service_ip" ]; then
        kubectl port-forward service/cloudbridge-relay-staging-service 9090:9090 -n "$NAMESPACE" &
        local port_forward_pid=$!
        sleep 5
        service_ip="localhost"
    fi
    
    # Test endpoints
    local endpoints=("/health" "/metrics" "/api/v1/status")
    local failed_tests=0
    
    for endpoint in "${endpoints[@]}"; do
        log_info "Testing endpoint: $endpoint"
        if curl -f "http://$service_ip:$service_port$endpoint" >/dev/null 2>&1; then
            log_success "✓ $endpoint - OK"
        else
            log_error "✗ $endpoint - FAILED"
            ((failed_tests++))
        fi
    done
    
    # Cleanup port-forward if used
    if [ -n "$port_forward_pid" ]; then
        kill $port_forward_pid 2>/dev/null || true
    fi
    
    if [ $failed_tests -gt 0 ]; then
        log_error "Smoke tests failed: $failed_tests tests failed"
        exit 1
    fi
    
    log_success "Smoke tests passed"
}

# Deploy monitoring
deploy_monitoring() {
    log_section "Deploying monitoring..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would deploy monitoring"
        return 0
    fi
    
    # Deploy Prometheus and Grafana for staging
    if [ -f "./k8s/monitoring/prometheus-deployment.yaml" ]; then
        log_info "Deploying Prometheus..."
        kubectl apply -f ./k8s/monitoring/prometheus-deployment.yaml
    fi
    
    if [ -f "./k8s/monitoring/grafana-deployment.yaml" ]; then
        log_info "Deploying Grafana..."
        kubectl apply -f ./k8s/monitoring/grafana-deployment.yaml
    fi
    
    log_success "Monitoring deployed"
}

# Deploy auto-scaling
deploy_autoscaling() {
    log_section "Deploying auto-scaling..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would deploy auto-scaling"
        return 0
    fi
    
    # Deploy HPA for staging
    if [ -f "./k8s/autoscaling/hpa.yaml" ]; then
        log_info "Deploying HPA..."
        kubectl apply -f ./k8s/autoscaling/hpa.yaml
    fi
    
    log_success "Auto-scaling deployed"
}

# Generate deployment report
generate_deployment_report() {
    log_section "Generating deployment report..."
    
    local report_file="/tmp/staging-deployment-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" <<EOF
CloudBridge Relay Staging Deployment Report
==========================================

Deployment Date: $(date)
Environment: $ENVIRONMENT
Namespace: $NAMESPACE
Image: $REGISTRY:$IMAGE_TAG
Registry: $REGISTRY

Deployment Status:
- Namespace: $NAMESPACE
- Deployment: cloudbridge-relay-staging
- Replicas: $(kubectl get deployment cloudbridge-relay-staging -n "$NAMESPACE" --no-headers | awk '{print $2}' 2>/dev/null || echo "N/A")
- Ready Replicas: $(kubectl get deployment cloudbridge-relay-staging -n "$NAMESPACE" --no-headers | awk '{print $4}' 2>/dev/null || echo "N/A")

Services:
- LoadBalancer: cloudbridge-relay-staging-service
- Internal: cloudbridge-relay-staging-internal
- Headless: cloudbridge-relay-staging-headless

Pods Status:
$(kubectl get pods -n "$NAMESPACE" --no-headers 2>/dev/null || echo "No pods found")

Services Status:
$(kubectl get services -n "$NAMESPACE" --no-headers 2>/dev/null || echo "No services found")

Endpoints:
- Health: http://<loadbalancer-ip>:9090/health
- Metrics: http://<loadbalancer-ip>:9090/metrics
- API: http://<loadbalancer-ip>:9090/api/v1/

Next Steps:
1. Run comprehensive testing suite
2. Monitor application logs
3. Verify all endpoints are accessible
4. Check monitoring dashboards
5. Run load tests

EOF
    
    log_success "Deployment report generated: $report_file"
    cat "$report_file"
}

# Main function
main() {
    log_info "Starting CloudBridge Relay staging deployment..."
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --image-tag)
                IMAGE_TAG="$2"
                shift 2
                ;;
            --registry)
                REGISTRY="$2"
                shift 2
                ;;
            --namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            --skip-tests)
                SKIP_TESTS="true"
                shift
                ;;
            --force)
                FORCE_DEPLOY="true"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Execute deployment
    check_dependencies
    pre_deployment_checks
    create_namespace
    deploy_configurations
    deploy_application
    wait_for_deployment
    run_health_checks
    run_smoke_tests
    deploy_monitoring
    deploy_autoscaling
    generate_deployment_report
    
    log_success "Staging deployment completed successfully!"
    log_info ""
    log_info "Deployment Summary:"
    log_info "  - Environment: $ENVIRONMENT"
    log_info "  - Namespace: $NAMESPACE"
    log_info "  - Image: $REGISTRY:$IMAGE_TAG"
    log_info "  - Status: Ready"
    log_info ""
    log_info "Useful commands:"
    log_info "  - Check pods: kubectl get pods -n $NAMESPACE"
    log_info "  - Check services: kubectl get services -n $NAMESPACE"
    log_info "  - View logs: kubectl logs -f deployment/cloudbridge-relay-staging -n $NAMESPACE"
    log_info "  - Port forward: kubectl port-forward service/cloudbridge-relay-staging-service 9090:9090 -n $NAMESPACE"
}

# Run main function
main "$@"

