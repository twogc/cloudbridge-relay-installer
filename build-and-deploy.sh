#!/bin/bash

set -e

echo "🚀 Building and deploying CloudBridge Relay to minikube"

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

# Check if minikube is running
print_status "Checking minikube status..."
if ! minikube status > /dev/null 2>&1; then
    print_error "Minikube is not running. Please start minikube first:"
    echo "  minikube start"
    exit 1
fi

print_success "Minikube is running"

# Set minikube docker environment
print_status "Setting minikube docker environment..."
eval $(minikube docker-env)

# Build Docker image
print_status "Building Docker image..."
docker build -t cloudbridge-relay:latest .

if [ $? -eq 0 ]; then
    print_success "Docker image built successfully"
else
    print_error "Failed to build Docker image"
    exit 1
fi

# Apply Kubernetes manifests
print_status "Applying Kubernetes manifests..."

# Create namespace
kubectl apply -f k8s/namespace.yaml
print_success "Namespace created"

# Apply ConfigMap
kubectl apply -f k8s/configmap.yaml
print_success "ConfigMap applied"

# Apply Deployment
kubectl apply -f k8s/deployment.yaml
print_success "Deployment applied"

# Apply Service
kubectl apply -f k8s/service.yaml
print_success "Service applied"

# Apply Ingress
kubectl apply -f k8s/ingress.yaml
print_success "Ingress applied"

# Wait for deployment to be ready
print_status "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/cloudbridge-relay -n cloudbridge-relay

if [ $? -eq 0 ]; then
    print_success "Deployment is ready"
else
    print_error "Deployment failed to become ready"
    exit 1
fi

# Get service information
print_status "Getting service information..."
SERVICE_IP=$(minikube service cloudbridge-relay-service -n cloudbridge-relay --url | head -1)
print_success "Service is available at: $SERVICE_IP"

# Show pod status
print_status "Pod status:"
kubectl get pods -n cloudbridge-relay

# Show service status
print_status "Service status:"
kubectl get services -n cloudbridge-relay

# Show ingress status
print_status "Ingress status:"
kubectl get ingress -n cloudbridge-relay

# Test health endpoint
print_status "Testing health endpoint..."
sleep 10
if curl -f "$SERVICE_IP/health" > /dev/null 2>&1; then
    print_success "Health check passed"
else
    print_warning "Health check failed, but deployment might still be starting"
fi

# Show logs
print_status "Showing recent logs:"
kubectl logs -n cloudbridge-relay deployment/cloudbridge-relay --tail=20

echo ""
print_success "🎉 CloudBridge Relay deployed successfully!"
echo ""
echo "📋 Access information:"
echo "  Service URL: $SERVICE_IP"
echo "  Health check: $SERVICE_IP/health"
echo "  API docs: $SERVICE_IP/docs"
echo ""
echo "🔧 Useful commands:"
echo "  View logs: kubectl logs -n cloudbridge-relay deployment/cloudbridge-relay -f"
echo "  Port forward: kubectl port-forward -n cloudbridge-relay service/cloudbridge-relay-service 8080:8080"
echo "  Delete deployment: kubectl delete -f k8s/"
echo ""
