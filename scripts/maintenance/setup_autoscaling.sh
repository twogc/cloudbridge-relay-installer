#!/bin/bash

# CloudBridge Relay - Auto-scaling Setup Script
# Этот скрипт настраивает HPA, VPA и Cluster Autoscaler

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="${NAMESPACE:-cloudbridge}"
ENABLE_HPA="${ENABLE_HPA:-true}"
ENABLE_VPA="${ENABLE_VPA:-true}"
ENABLE_CLUSTER_AUTOSCALER="${ENABLE_CLUSTER_AUTOSCALER:-true}"
CLOUD_PROVIDER="${CLOUD_PROVIDER:-aws}"  # aws, gcp, azure
CLUSTER_NAME="${CLUSTER_NAME:-cloudbridge-cluster}"

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
    
    for tool in kubectl; do
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

# Install metrics server
install_metrics_server() {
    log_section "Installing metrics server..."
    
    # Check if metrics server is already installed
    if kubectl get deployment metrics-server -n kube-system &>/dev/null; then
        log_info "Metrics server is already installed"
        return 0
    fi
    
    log_info "Installing metrics server..."
    kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
    
    # Wait for metrics server to be ready
    log_info "Waiting for metrics server to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/metrics-server -n kube-system
    
    log_success "Metrics server installed successfully"
}

# Install VPA
install_vpa() {
    log_section "Installing Vertical Pod Autoscaler..."
    
    # Check if VPA is already installed
    if kubectl get crd verticalpodautoscalers.autoscaling.k8s.io &>/dev/null; then
        log_info "VPA is already installed"
        return 0
    fi
    
    log_info "Installing VPA..."
    
    # Clone VPA repository
    local vpa_dir="/tmp/vpa"
    git clone https://github.com/kubernetes/autoscaler.git "$vpa_dir"
    
    # Install VPA
    kubectl apply -f "$vpa_dir/vertical-pod-autoscaler/hack/vpa-installer.yaml"
    
    # Wait for VPA to be ready
    log_info "Waiting for VPA to be ready..."
    kubectl wait --for=condition=available --timeout=300s deployment/vpa-admission-controller -n kube-system
    kubectl wait --for=condition=available --timeout=300s deployment/vpa-recommender -n kube-system
    kubectl wait --for=condition=available --timeout=300s deployment/vpa-updater -n kube-system
    
    # Cleanup
    rm -rf "$vpa_dir"
    
    log_success "VPA installed successfully"
}

# Setup HPA
setup_hpa() {
    log_section "Setting up Horizontal Pod Autoscaler..."
    
    # Apply HPA configurations
    if [ -f "./k8s/autoscaling/hpa.yaml" ]; then
        log_info "Applying HPA configurations..."
        kubectl apply -f ./k8s/autoscaling/hpa.yaml
        
        # Wait for HPA to be created
        log_info "Waiting for HPA resources to be created..."
        kubectl wait --for=condition=ready --timeout=60s hpa --all -n "$NAMESPACE" || true
        
        log_success "HPA configured successfully"
    else
        log_error "HPA configuration file not found: ./k8s/autoscaling/hpa.yaml"
        exit 1
    fi
}

# Setup VPA
setup_vpa() {
    log_section "Setting up Vertical Pod Autoscaler..."
    
    # Apply VPA configurations
    if [ -f "./k8s/autoscaling/vpa.yaml" ]; then
        log_info "Applying VPA configurations..."
        kubectl apply -f ./k8s/autoscaling/vpa.yaml
        
        log_success "VPA configured successfully"
    else
        log_error "VPA configuration file not found: ./k8s/autoscaling/vpa.yaml"
        exit 1
    fi
}

# Setup Cluster Autoscaler
setup_cluster_autoscaler() {
    log_section "Setting up Cluster Autoscaler..."
    
    case $CLOUD_PROVIDER in
        aws)
            setup_aws_cluster_autoscaler
            ;;
        gcp)
            setup_gcp_cluster_autoscaler
            ;;
        azure)
            setup_azure_cluster_autoscaler
            ;;
        *)
            log_error "Unsupported cloud provider: $CLOUD_PROVIDER"
            exit 1
            ;;
    esac
}

# Setup AWS Cluster Autoscaler
setup_aws_cluster_autoscaler() {
    log_info "Setting up AWS Cluster Autoscaler..."
    
    # Check if running on AWS
    if ! curl -s http://169.254.169.254/latest/meta-data/instance-id &>/dev/null; then
        log_warning "Not running on AWS, skipping AWS-specific configuration"
        return 0
    fi
    
    # Apply Cluster Autoscaler configuration
    if [ -f "./k8s/autoscaling/cluster-autoscaler.yaml" ]; then
        log_info "Applying Cluster Autoscaler configuration..."
        kubectl apply -f ./k8s/autoscaling/cluster-autoscaler.yaml
        
        # Wait for Cluster Autoscaler to be ready
        log_info "Waiting for Cluster Autoscaler to be ready..."
        kubectl wait --for=condition=available --timeout=300s deployment/cluster-autoscaler -n kube-system
        
        log_success "AWS Cluster Autoscaler configured successfully"
    else
        log_error "Cluster Autoscaler configuration file not found"
        exit 1
    fi
}

# Setup GCP Cluster Autoscaler
setup_gcp_cluster_autoscaler() {
    log_info "Setting up GCP Cluster Autoscaler..."
    
    # Check if running on GCP
    if ! curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/id &>/dev/null; then
        log_warning "Not running on GCP, skipping GCP-specific configuration"
        return 0
    fi
    
    log_warning "GCP Cluster Autoscaler setup not fully implemented"
    log_info "Please configure GCP Cluster Autoscaler manually using:"
    log_info "  gcloud container clusters update $CLUSTER_NAME --enable-autoscaling --min-nodes=1 --max-nodes=10"
}

# Setup Azure Cluster Autoscaler
setup_azure_cluster_autoscaler() {
    log_info "Setting up Azure Cluster Autoscaler..."
    
    log_warning "Azure Cluster Autoscaler setup not fully implemented"
    log_info "Please configure Azure Cluster Autoscaler manually using:"
    log_info "  az aks update --resource-group <resource-group> --name $CLUSTER_NAME --enable-cluster-autoscaler --min-count 1 --max-count 10"
}

# Create custom metrics
create_custom_metrics() {
    log_section "Creating custom metrics..."
    
    # Create custom metrics API
    cat > /tmp/custom-metrics-api.yaml <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: custom-metrics-api
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: custom-metrics-api
rules:
- apiGroups: [""]
  resources: ["nodes", "pods"]
  verbs: ["get", "list"]
- apiGroups: ["custom.metrics.k8s.io"]
  resources: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: custom-metrics-api
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: custom-metrics-api
subjects:
- kind: ServiceAccount
  name: custom-metrics-api
  namespace: kube-system
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: custom-metrics-api
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: custom-metrics-api
  template:
    metadata:
      labels:
        app: custom-metrics-api
    spec:
      serviceAccountName: custom-metrics-api
      containers:
      - name: custom-metrics-api
        image: k8s.gcr.io/custom-metrics-apiserver:v0.1.0
        args:
        - --secure-port=6443
        - --cert-dir=/tmp
        - --logtostderr=true
        - --v=4
        ports:
        - containerPort: 6443
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 200m
            memory: 256Mi
---
apiVersion: v1
kind: Service
metadata:
  name: custom-metrics-api
  namespace: kube-system
spec:
  selector:
    app: custom-metrics-api
  ports:
  - port: 443
    targetPort: 6443
EOF
    
    kubectl apply -f /tmp/custom-metrics-api.yaml
    rm -f /tmp/custom-metrics-api.yaml
    
    log_success "Custom metrics API created"
}

# Verify autoscaling setup
verify_autoscaling() {
    log_section "Verifying autoscaling setup..."
    
    # Check HPA
    if [ "$ENABLE_HPA" = "true" ]; then
        log_info "Checking HPA status..."
        kubectl get hpa -n "$NAMESPACE" || log_warning "No HPA resources found"
    fi
    
    # Check VPA
    if [ "$ENABLE_VPA" = "true" ]; then
        log_info "Checking VPA status..."
        kubectl get vpa -n "$NAMESPACE" || log_warning "No VPA resources found"
    fi
    
    # Check Cluster Autoscaler
    if [ "$ENABLE_CLUSTER_AUTOSCALER" = "true" ]; then
        log_info "Checking Cluster Autoscaler status..."
        kubectl get deployment cluster-autoscaler -n kube-system || log_warning "Cluster Autoscaler not found"
    fi
    
    # Check metrics server
    log_info "Checking metrics server status..."
    kubectl get deployment metrics-server -n kube-system || log_warning "Metrics server not found"
    
    log_success "Autoscaling verification completed"
}

# Create autoscaling documentation
create_autoscaling_docs() {
    log_section "Creating autoscaling documentation..."
    
    local docs_file="./docs/autoscaling/AUTOSCALING_GUIDE.md"
    mkdir -p "$(dirname "$docs_file")"
    
    cat > "$docs_file" <<EOF
# CloudBridge Relay Auto-scaling Guide

## Overview

This guide covers the auto-scaling configuration for CloudBridge Relay, including:

- Horizontal Pod Autoscaler (HPA)
- Vertical Pod Autoscaler (VPA)
- Cluster Autoscaler
- Custom metrics

## Components

### Horizontal Pod Autoscaler (HPA)

HPA automatically scales the number of pods based on CPU, memory, and custom metrics.

**Configuration:**
- Min replicas: 2-3
- Max replicas: 10-20
- CPU threshold: 70-80%
- Memory threshold: 80-85%

**Services with HPA:**
- cloudbridge-relay (3-20 replicas)
- quic-service (2-15 replicas)
- wireguard-service (2-10 replicas)
- ai-service (1-8 replicas)
- cadence-service (2-12 replicas)
- quantum-service (1-6 replicas)
- api-service (2-15 replicas)

### Vertical Pod Autoscaler (VPA)

VPA automatically adjusts CPU and memory requests and limits based on usage patterns.

**Configuration:**
- Update mode: Auto (for stateless services)
- Update mode: Initial (for databases)
- CPU range: 50m-4000m
- Memory range: 64Mi-16Gi

### Cluster Autoscaler

Cluster Autoscaler automatically scales the number of nodes in the cluster.

**Configuration:**
- Min nodes: 1
- Max nodes: 100
- Scale-down delay: 10 minutes
- Scale-down utilization threshold: 50%

## Monitoring

### Metrics to Monitor

1. **HPA Metrics:**
   - Current replicas vs desired replicas
   - CPU and memory utilization
   - Custom metrics (connections, requests)

2. **VPA Metrics:**
   - Resource recommendations
   - Actual vs recommended resources
   - Update frequency

3. **Cluster Autoscaler Metrics:**
   - Node count
   - Pending pods
   - Scale-up/down events

### Alerts

Set up alerts for:
- HPA unable to scale due to resource constraints
- VPA recommendations not being applied
- Cluster Autoscaler failing to scale nodes
- High resource utilization without scaling

## Best Practices

1. **Resource Limits:**
   - Set appropriate CPU and memory limits
   - Use VPA recommendations as a starting point
   - Monitor and adjust based on actual usage

2. **Scaling Policies:**
   - Use gradual scaling policies
   - Set appropriate stabilization windows
   - Configure scale-down delays

3. **Custom Metrics:**
   - Implement custom metrics for business logic
   - Use Prometheus for metrics collection
   - Set appropriate thresholds

4. **Testing:**
   - Load test with auto-scaling enabled
   - Verify scaling behavior under different loads
   - Test failure scenarios

## Troubleshooting

### Common Issues

1. **HPA not scaling:**
   - Check metrics server is running
   - Verify resource requests are set
   - Check for resource constraints

2. **VPA not updating:**
   - Verify VPA is in Auto mode
   - Check for resource policy conflicts
   - Monitor VPA logs

3. **Cluster Autoscaler not working:**
   - Check node group configuration
   - Verify IAM permissions
   - Check for pod disruption budgets

### Useful Commands

\`\`\`bash
# Check HPA status
kubectl get hpa -n cloudbridge

# Check VPA status
kubectl get vpa -n cloudbridge

# Check Cluster Autoscaler logs
kubectl logs -n kube-system deployment/cluster-autoscaler

# Check metrics server
kubectl top nodes
kubectl top pods -n cloudbridge
\`\`\`

## Configuration Files

- HPA: \`k8s/autoscaling/hpa.yaml\`
- VPA: \`k8s/autoscaling/vpa.yaml\`
- Cluster Autoscaler: \`k8s/autoscaling/cluster-autoscaler.yaml\`

EOF
    
    log_success "Autoscaling documentation created: $docs_file"
}

# Main function
main() {
    log_info "Starting CloudBridge Relay auto-scaling setup..."
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --disable-hpa)
                ENABLE_HPA="false"
                shift
                ;;
            --disable-vpa)
                ENABLE_VPA="false"
                shift
                ;;
            --disable-cluster-autoscaler)
                ENABLE_CLUSTER_AUTOSCALER="false"
                shift
                ;;
            --cloud-provider)
                CLOUD_PROVIDER="$2"
                shift 2
                ;;
            --cluster-name)
                CLUSTER_NAME="$2"
                shift 2
                ;;
            --namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Execute setup
    check_dependencies
    install_metrics_server
    
    if [ "$ENABLE_VPA" = "true" ]; then
        install_vpa
        setup_vpa
    fi
    
    if [ "$ENABLE_HPA" = "true" ]; then
        setup_hpa
    fi
    
    if [ "$ENABLE_CLUSTER_AUTOSCALER" = "true" ]; then
        setup_cluster_autoscaler
    fi
    
    create_custom_metrics
    verify_autoscaling
    create_autoscaling_docs
    
    log_success "Auto-scaling setup completed successfully!"
    log_info ""
    log_info "Auto-scaling components enabled:"
    log_info "  - HPA: $ENABLE_HPA"
    log_info "  - VPA: $ENABLE_VPA"
    log_info "  - Cluster Autoscaler: $ENABLE_CLUSTER_AUTOSCALER"
    log_info "  - Cloud Provider: $CLOUD_PROVIDER"
    log_info ""
    log_info "Useful commands:"
    log_info "  - Check HPA: kubectl get hpa -n $NAMESPACE"
    log_info "  - Check VPA: kubectl get vpa -n $NAMESPACE"
    log_info "  - Check metrics: kubectl top pods -n $NAMESPACE"
    log_info "  - Check Cluster Autoscaler: kubectl get deployment cluster-autoscaler -n kube-system"
}

# Run main function
main "$@"

