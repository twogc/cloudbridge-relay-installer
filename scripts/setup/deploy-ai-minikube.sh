#!/bin/bash

# CloudBridge Relay AI - Minikube Deployment Script
# This script deploys the AI system to Minikube

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

# Check if Minikube is running
check_minikube() {
    log_info "Checking Minikube status..."
    if ! minikube status > /dev/null 2>&1; then
        log_error "Minikube is not running. Please start Minikube first:"
        log_info "minikube start --driver=docker --memory=4096 --cpus=2"
        exit 1
    fi
    log_success "Minikube is running"
}

# Set Docker environment for Minikube
setup_docker_env() {
    log_info "Setting up Docker environment for Minikube..."
    eval $(minikube docker-env)
    log_success "Docker environment configured for Minikube"
}

# Build AI Docker images
build_images() {
    log_info "Building AI Docker images..."
    
    # Build AI Service image
    log_info "Building AI Service image..."
    docker build -f Dockerfile.ai-service -t cloudbridge-ai:latest .
    if [ $? -eq 0 ]; then
        log_success "AI Service image built successfully"
    else
        log_error "Failed to build AI Service image"
        exit 1
    fi
    
    # Build AI Worker image
    log_info "Building AI Worker image..."
    docker build -f Dockerfile.ai-worker -t cloudbridge-ai-worker:latest .
    if [ $? -eq 0 ]; then
        log_success "AI Worker image built successfully"
    else
        log_error "Failed to build AI Worker image"
        exit 1
    fi
}

# Apply Kubernetes manifests
apply_manifests() {
    log_info "Applying Kubernetes manifests..."
    
    # Apply ConfigMaps
    log_info "Applying AI ConfigMap..."
    kubectl apply -f k8s/configmaps/ai-config.yaml
    if [ $? -eq 0 ]; then
        log_success "AI ConfigMap applied successfully"
    else
        log_error "Failed to apply AI ConfigMap"
        exit 1
    fi
    
    # Apply PersistentVolumeClaims
    log_info "Applying AI PersistentVolumeClaims..."
    kubectl apply -f k8s/persistentvolumeclaims/ai-pvc.yaml
    if [ $? -eq 0 ]; then
        log_success "AI PVCs applied successfully"
    else
        log_error "Failed to apply AI PVCs"
        exit 1
    fi
    
    # Apply Services
    log_info "Applying AI Services..."
    kubectl apply -f k8s/services/ai-service.yaml
    if [ $? -eq 0 ]; then
        log_success "AI Services applied successfully"
    else
        log_error "Failed to apply AI Services"
        exit 1
    fi
    
    # Apply Deployments
    log_info "Applying AI Service Deployment..."
    kubectl apply -f k8s/deployments/ai-service-deployment.yaml
    if [ $? -eq 0 ]; then
        log_success "AI Service Deployment applied successfully"
    else
        log_error "Failed to apply AI Service Deployment"
        exit 1
    fi
    
    log_info "Applying AI Worker Deployment..."
    kubectl apply -f k8s/deployments/ai-worker-deployment.yaml
    if [ $? -eq 0 ]; then
        log_success "AI Worker Deployment applied successfully"
    else
        log_error "Failed to apply AI Worker Deployment"
        exit 1
    fi
}

# Wait for deployments to be ready
wait_for_deployments() {
    log_info "Waiting for AI deployments to be ready..."
    
    # Wait for AI Service
    log_info "Waiting for AI Service pods..."
    kubectl wait --for=condition=available --timeout=300s deployment/ai-service -n cloudbridge
    if [ $? -eq 0 ]; then
        log_success "AI Service deployment is ready"
    else
        log_error "AI Service deployment failed to become ready"
        exit 1
    fi
    
    # Wait for AI Worker
    log_info "Waiting for AI Worker pods..."
    kubectl wait --for=condition=available --timeout=300s deployment/ai-worker -n cloudbridge
    if [ $? -eq 0 ]; then
        log_success "AI Worker deployment is ready"
    else
        log_error "AI Worker deployment failed to become ready"
        exit 1
    fi
}

# Check deployment status
check_status() {
    log_info "Checking AI deployment status..."
    
    echo ""
    log_info "AI Pods:"
    kubectl get pods -n cloudbridge -l app.kubernetes.io/component=ai
    
    echo ""
    log_info "AI Services:"
    kubectl get services -n cloudbridge -l app.kubernetes.io/component=ai
    
    echo ""
    log_info "AI PVCs:"
    kubectl get pvc -n cloudbridge -l app.kubernetes.io/component=ai
}

# Setup port forwarding
setup_port_forwarding() {
    log_info "Setting up port forwarding for AI Service..."
    
    # Kill existing port forwarding processes
    pkill -f "kubectl port-forward.*ai-service" || true
    
    # Start port forwarding in background
    kubectl port-forward -n cloudbridge service/ai-service-nodeport 8083:8083 &
    kubectl port-forward -n cloudbridge service/ai-service-nodeport 8084:8084 &
    
    log_success "Port forwarding setup complete"
    log_info "AI Service API: http://localhost:8083"
    log_info "AI Service Metrics: http://localhost:8084"
}

# Test AI Service
test_ai_service() {
    log_info "Testing AI Service..."
    
    # Wait a bit for port forwarding to be ready
    sleep 5
    
    # Test health endpoint
    log_info "Testing health endpoint..."
    if curl -s http://localhost:8083/health > /dev/null; then
        log_success "AI Service health check passed"
    else
        log_warning "AI Service health check failed (this might be normal during startup)"
    fi
    
    # Test AI status endpoint
    log_info "Testing AI status endpoint..."
    if curl -s http://localhost:8083/api/v1/ai/status > /dev/null; then
        log_success "AI Service status endpoint is accessible"
    else
        log_warning "AI Service status endpoint failed (this might be normal during startup)"
    fi
}

# Show useful commands
show_commands() {
    echo ""
    log_info "Useful commands:"
    echo ""
    echo "  # Check AI pods status:"
    echo "  kubectl get pods -n cloudbridge -l app.kubernetes.io/component=ai"
    echo ""
    echo "  # View AI Service logs:"
    echo "  kubectl logs -n cloudbridge -l app.kubernetes.io/name=ai-service"
    echo ""
    echo "  # View AI Worker logs:"
    echo "  kubectl logs -n cloudbridge -l app.kubernetes.io/name=ai-worker"
    echo ""
    echo "  # Access AI Service API:"
    echo "  curl http://localhost:8083/health"
    echo "  curl http://localhost:8083/api/v1/ai/status"
    echo ""
    echo "  # Access AI Service metrics:"
    echo "  curl http://localhost:8084/metrics"
    echo ""
    echo "  # Restart AI deployments:"
    echo "  kubectl rollout restart deployment/ai-service -n cloudbridge"
    echo "  kubectl rollout restart deployment/ai-worker -n cloudbridge"
    echo ""
    echo "  # Delete AI deployments:"
    echo "  kubectl delete -f k8s/deployments/ai-service-deployment.yaml"
    echo "  kubectl delete -f k8s/deployments/ai-worker-deployment.yaml"
    echo ""
}

# Main deployment function
main() {
    echo ""
    log_info "Starting CloudBridge Relay AI deployment to Minikube..."
    echo ""
    
    # Check prerequisites
    check_minikube
    setup_docker_env
    
    # Build and deploy
    build_images
    apply_manifests
    wait_for_deployments
    
    # Verify deployment
    check_status
    setup_port_forwarding
    test_ai_service
    
    # Show useful information
    show_commands
    
    echo ""
    log_success "AI deployment to Minikube completed successfully!"
    echo ""
    log_info "AI Service is now running at: http://localhost:8083"
    log_info "AI Metrics are available at: http://localhost:8084"
    echo ""
}

# Handle script arguments
case "${1:-}" in
    "status")
        check_status
        ;;
    "logs")
        kubectl logs -n cloudbridge -l app.kubernetes.io/component=ai -f
        ;;
    "restart")
        log_info "Restarting AI deployments..."
        kubectl rollout restart deployment/ai-service -n cloudbridge
        kubectl rollout restart deployment/ai-worker -n cloudbridge
        log_success "AI deployments restarted"
        ;;
    "delete")
        log_info "Deleting AI deployments..."
        kubectl delete -f k8s/deployments/ai-service-deployment.yaml
        kubectl delete -f k8s/deployments/ai-worker-deployment.yaml
        kubectl delete -f k8s/services/ai-service.yaml
        kubectl delete -f k8s/configmaps/ai-config.yaml
        kubectl delete -f k8s/persistentvolumeclaims/ai-pvc.yaml
        log_success "AI deployments deleted"
        ;;
    "test")
        test_ai_service
        ;;
    *)
        main
        ;;
esac

