#!/bin/bash

# CloudBridge Relay - Kubernetes Deployment Script
# This script deploys CloudBridge Relay Server to Kubernetes

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
K8S_DIR="k8s"

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

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if kubectl can connect to cluster
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi
    
    # Check if k8s directory exists
    if [ ! -d "$K8S_DIR" ]; then
        log_error "Kubernetes manifests directory '$K8S_DIR' not found."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

create_namespaces() {
    log_info "Creating namespaces..."
    
    kubectl apply -f $K8S_DIR/namespaces/cloudbridge-namespace.yaml
    
    log_success "Namespaces created"
}

create_configmaps() {
    log_info "Creating ConfigMaps..."
    
    kubectl apply -f $K8S_DIR/configmaps/cloudbridge-config.yaml
    
    log_success "ConfigMaps created"
}

create_secrets() {
    log_info "Creating Secrets..."
    
    kubectl apply -f $K8S_DIR/secrets/cloudbridge-secrets.yaml
    
    log_success "Secrets created"
}

create_persistent_volumes() {
    log_info "Creating Persistent Volumes..."
    
    kubectl apply -f $K8S_DIR/persistent-volumes/cloudbridge-pv.yaml
    
    log_success "Persistent Volumes created"
}

deploy_services() {
    log_info "Deploying services..."
    
    # Deploy Cassandra first (StatefulSet)
    kubectl apply -f $K8S_DIR/deployments/cassandra-deployment.yaml
    kubectl apply -f $K8S_DIR/services/cloudbridge-services.yaml
    
    # Wait for Cassandra to be ready
    log_info "Waiting for Cassandra to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=cassandra -n $NAMESPACE --timeout=300s
    
    # Deploy Cadence
    kubectl apply -f $K8S_DIR/deployments/cadence-deployment.yaml
    
    # Wait for Cadence to be ready
    log_info "Waiting for Cadence to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=cadence -n $NAMESPACE --timeout=300s
    
    # Deploy CloudBridge Relay
    kubectl apply -f $K8S_DIR/deployments/cloudbridge-relay-deployment.yaml
    
    # Wait for CloudBridge Relay to be ready
    log_info "Waiting for CloudBridge Relay to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=cloudbridge-relay -n $NAMESPACE --timeout=300s
    
    log_success "Services deployed"
}

deploy_monitoring() {
    log_info "Deploying monitoring stack..."
    
    # Deploy Prometheus
    kubectl apply -f $K8S_DIR/monitoring/prometheus-deployment.yaml
    
    # Wait for Prometheus to be ready
    log_info "Waiting for Prometheus to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=prometheus -n $MONITORING_NAMESPACE --timeout=300s
    
    # Deploy Grafana
    kubectl apply -f $K8S_DIR/monitoring/grafana-deployment.yaml
    
    # Wait for Grafana to be ready
    log_info "Waiting for Grafana to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=grafana -n $MONITORING_NAMESPACE --timeout=300s
    
    log_success "Monitoring stack deployed"
}

deploy_ingress() {
    log_info "Deploying Ingress..."
    
    kubectl apply -f $K8S_DIR/ingress/cloudbridge-ingress.yaml
    
    log_success "Ingress deployed"
}

show_status() {
    log_info "Deployment Status:"
    echo ""
    
    echo "📊 Pods Status:"
    kubectl get pods -n $NAMESPACE
    kubectl get pods -n $MONITORING_NAMESPACE
    echo ""
    
    echo "🌐 Services Status:"
    kubectl get services -n $NAMESPACE
    kubectl get services -n $MONITORING_NAMESPACE
    echo ""
    
    echo "💾 Persistent Volumes Status:"
    kubectl get pv,pvc -n $NAMESPACE
    kubectl get pv,pvc -n $MONITORING_NAMESPACE
    echo ""
    
    echo "🔗 Ingress Status:"
    kubectl get ingress -n $NAMESPACE
    echo ""
}

show_access_info() {
    log_info "Access Information:"
    echo ""
    
    echo "🚀 CloudBridge Relay:"
    echo "   - Internal: http://cloudbridge-relay.$NAMESPACE.svc.cluster.local:9090"
    echo "   - External: https://cloudbridge.example.com"
    echo ""
    
    echo "📊 Grafana Dashboard:"
    echo "   - Internal: http://grafana.$MONITORING_NAMESPACE.svc.cluster.local:3000"
    echo "   - External: https://grafana.cloudbridge.example.com"
    echo "   - Username: admin"
    echo "   - Password: admin123"
    echo ""
    
    echo "📈 Prometheus:"
    echo "   - Internal: http://prometheus.$MONITORING_NAMESPACE.svc.cluster.local:9090"
    echo "   - External: https://prometheus.cloudbridge.example.com"
    echo ""
    
    echo "🗄️  Cassandra:"
    echo "   - Internal: cassandra-service.$NAMESPACE.svc.cluster.local:9042"
    echo ""
    
    echo "⚡ Cadence:"
    echo "   - Internal: cadence-service.$NAMESPACE.svc.cluster.local:7933"
    echo "   - Web UI: http://cadence-service.$NAMESPACE.svc.cluster.local:7934"
    echo ""
}

cleanup() {
    log_warning "Cleaning up deployment..."
    
    # Delete Ingress
    kubectl delete -f $K8S_DIR/ingress/cloudbridge-ingress.yaml --ignore-not-found=true
    
    # Delete monitoring
    kubectl delete -f $K8S_DIR/monitoring/grafana-deployment.yaml --ignore-not-found=true
    kubectl delete -f $K8S_DIR/monitoring/prometheus-deployment.yaml --ignore-not-found=true
    
    # Delete services
    kubectl delete -f $K8S_DIR/deployments/cloudbridge-relay-deployment.yaml --ignore-not-found=true
    kubectl delete -f $K8S_DIR/deployments/cadence-deployment.yaml --ignore-not-found=true
    kubectl delete -f $K8S_DIR/deployments/cassandra-deployment.yaml --ignore-not-found=true
    kubectl delete -f $K8S_DIR/services/cloudbridge-services.yaml --ignore-not-found=true
    
    # Delete persistent volumes
    kubectl delete -f $K8S_DIR/persistent-volumes/cloudbridge-pv.yaml --ignore-not-found=true
    
    # Delete secrets and configmaps
    kubectl delete -f $K8S_DIR/secrets/cloudbridge-secrets.yaml --ignore-not-found=true
    kubectl delete -f $K8S_DIR/configmaps/cloudbridge-config.yaml --ignore-not-found=true
    
    # Delete namespaces
    kubectl delete -f $K8S_DIR/namespaces/cloudbridge-namespace.yaml --ignore-not-found=true
    
    log_success "Cleanup completed"
}

# Main function
main() {
    echo "=== 🚀 CloudBridge Relay - Kubernetes Deployment ==="
    echo ""
    
    case "${1:-deploy}" in
        "deploy")
            check_prerequisites
            create_namespaces
            create_configmaps
            create_secrets
            create_persistent_volumes
            deploy_services
            deploy_monitoring
            deploy_ingress
            show_status
            show_access_info
            log_success "Deployment completed successfully!"
            ;;
        "status")
            show_status
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [deploy|status|cleanup|help]"
            echo ""
            echo "Commands:"
            echo "  deploy   - Deploy CloudBridge Relay to Kubernetes (default)"
            echo "  status   - Show deployment status"
            echo "  cleanup  - Remove all deployed resources"
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

