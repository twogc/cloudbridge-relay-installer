#!/bin/bash

# CloudBridge Cloudflare Testing Deployment Script
# This script deploys the testing environment for Cloudflare integration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="cloudbridge-testing"
TESTING_DIR="k8s/cloudflare-testing"

echo -e "${BLUE}Starting CloudBridge Cloudflare Testing Deployment${NC}"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if we can connect to cluster
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}Cannot connect to Kubernetes cluster${NC}"
    exit 1
fi

echo -e "${GREEN}Kubernetes cluster connection verified${NC}"

# Create namespace
echo -e "${BLUE}📁 Creating namespace: ${NAMESPACE}${NC}"
kubectl apply -f ${TESTING_DIR}/namespace.yaml

# Wait for namespace to be ready
kubectl wait --for=condition=Active namespace/${NAMESPACE} --timeout=30s

# Create ConfigMaps
echo -e "${BLUE}Creating ConfigMaps${NC}"
kubectl apply -f ${TESTING_DIR}/cloudbridge-testing-configmap.yaml
kubectl apply -f ${TESTING_DIR}/cloudflared-configmap.yaml

# Create secrets for cloudflared (placeholder - needs to be filled with actual values)
echo -e "${YELLOW} Creating placeholder secrets for cloudflared${NC}"
kubectl create secret generic cloudflared-credentials \
  --from-literal=credentials.json='{"AccountTag":"PLACEHOLDER","TunnelSecret":"PLACEHOLDER","TunnelID":"PLACEHOLDER"}' \
  --namespace=${NAMESPACE} \
  --dry-run=client -o yaml | kubectl apply -f -

# Deploy cloudflared first
echo -e "${BLUE}☁️  Deploying cloudflared${NC}"
kubectl apply -f ${TESTING_DIR}/cloudflared-deployment.yaml

# Deploy CloudBridge relay
echo -e "${BLUE}Deploying CloudBridge relay${NC}"
kubectl apply -f ${TESTING_DIR}/cloudbridge-relay-testing-deployment.yaml

# Create services
echo -e "${BLUE}Creating services${NC}"
kubectl apply -f ${TESTING_DIR}/cloudbridge-testing-services.yaml

# Wait for deployments to be ready
echo -e "${BLUE}Waiting for deployments to be ready${NC}"
kubectl wait --for=condition=Available deployment/cloudflared-testing -n ${NAMESPACE} --timeout=300s
kubectl wait --for=condition=Available deployment/cloudbridge-relay-testing -n ${NAMESPACE} --timeout=300s

# Check pod status
echo -e "${BLUE}Checking pod status${NC}"
kubectl get pods -n ${NAMESPACE}

# Check services
echo -e "${BLUE}Checking services${NC}"
kubectl get services -n ${NAMESPACE}

# Display access information
echo -e "${GREEN}Testing environment deployed successfully!${NC}"
echo -e "${BLUE}Access Information:${NC}"
echo -e "  Namespace: ${NAMESPACE}"
echo -e "  CloudBridge Relay:"
echo -e "    - Main Server: kubectl port-forward svc/cloudbridge-relay-testing 6001:6001"
echo -e "    - Metrics: kubectl port-forward svc/cloudbridge-relay-testing 6002:6002"
echo -e "    - API: kubectl port-forward svc/cloudbridge-relay-testing 6003:6003"
echo -e "    - HTTP API: kubectl port-forward svc/cloudbridge-relay-testing 6004:6004"
echo -e "    - QUIC: kubectl port-forward svc/cloudbridge-relay-testing 6091:6091"
echo -e "  Cloudflared:"
echo -e "    - Metrics: kubectl port-forward svc/cloudflared-testing 6006:6006"
echo -e "    - Health: kubectl port-forward svc/cloudflared-testing 6007:6007"

echo -e "${YELLOW} Next steps:${NC}"
echo -e "  1. Update cloudflared credentials with actual values"
echo -e "  2. Configure Cloudflare tunnel"
echo -e "  3. Run connectivity tests"
echo -e "  4. Monitor metrics and logs"

echo -e "${GREEN}Deployment completed!${NC}"
