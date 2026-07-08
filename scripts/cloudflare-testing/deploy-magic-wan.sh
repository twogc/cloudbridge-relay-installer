#!/bin/bash

# CloudBridge Magic WAN Integration Deployment Script
# This script deploys CloudBridge Relay as a proxy to Cloudflare Magic WAN

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

echo -e "${BLUE}Starting CloudBridge Magic WAN Integration Deployment${NC}"

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

# Create namespace if it doesn't exist
echo -e "${BLUE}📁 Ensuring namespace exists: ${NAMESPACE}${NC}"
kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -

# Deploy Magic WAN configuration
echo -e "${BLUE}Deploying Magic WAN configuration${NC}"
kubectl apply -f ${TESTING_DIR}/cloudbridge-magic-wan-config.yaml
kubectl apply -f ${TESTING_DIR}/cloudflared-magic-wan-config.yaml

# Create placeholder secrets for cloudflared
echo -e "${YELLOW} Creating placeholder secrets for cloudflared${NC}"
kubectl create secret generic cloudflared-credentials \
  --from-literal=credentials.json='{"AccountTag":"PLACEHOLDER","TunnelSecret":"PLACEHOLDER","TunnelID":"PLACEHOLDER"}' \
  --namespace=${NAMESPACE} \
  --dry-run=client -o yaml | kubectl apply -f -

# Deploy CloudBridge relay for Magic WAN
echo -e "${BLUE}Deploying CloudBridge relay for Magic WAN${NC}"
kubectl apply -f ${TESTING_DIR}/cloudbridge-magic-wan-deployment.yaml

# Create services
echo -e "${BLUE}Creating services${NC}"
kubectl apply -f ${TESTING_DIR}/cloudbridge-magic-wan-services.yaml

# Wait for deployments to be ready
echo -e "${BLUE}Waiting for deployments to be ready${NC}"
kubectl wait --for=condition=Available deployment/cloudbridge-relay-magic-wan -n ${NAMESPACE} --timeout=300s

# Check pod status
echo -e "${BLUE}Checking pod status${NC}"
kubectl get pods -n ${NAMESPACE}

# Check services
echo -e "${BLUE}Checking services${NC}"
kubectl get services -n ${NAMESPACE}

# Display access information
echo -e "${GREEN}Magic WAN integration deployed successfully!${NC}"
echo -e "${BLUE}Access Information:${NC}"
echo -e "  Namespace: ${NAMESPACE}"
echo -e "  CloudBridge Relay (Magic WAN):"
echo -e "    - Main Server: kubectl port-forward svc/cloudbridge-relay-magic-wan 6001:6001"
echo -e "    - Metrics: kubectl port-forward svc/cloudbridge-relay-magic-wan 6002:6002"
echo -e "    - API: kubectl port-forward svc/cloudbridge-relay-magic-wan 6003:6003"
echo -e "    - HTTP API: kubectl port-forward svc/cloudbridge-relay-magic-wan 6004:6004"
echo -e "    - HTTPS API: kubectl port-forward svc/cloudbridge-relay-magic-wan 6005:6005"
echo -e "    - QUIC: kubectl port-forward svc/cloudbridge-relay-magic-wan 6091:6091"

echo -e "${BLUE}External Access (if LoadBalancer is available):${NC}"
echo -e "  - HTTP: http://<EXTERNAL-IP>:80"
echo -e "  - HTTPS: https://<EXTERNAL-IP>:443"
echo -e "  - QUIC: <EXTERNAL-IP>:9091"

echo -e "${YELLOW} Next steps for Magic WAN integration:${NC}"
echo -e "  1. Set up real Cloudflare tunnel to Magic WAN"
echo -e "  2. Configure DNS records for your domain"
echo -e "  3. Update cloudflared credentials with real values"
echo -e "  4. Test client connections through CloudBridge Relay"
echo -e "  5. Monitor performance and adjust settings"

echo -e "${BLUE}🏗️  Architecture:${NC}"
echo -e "  Client (cloudflared) -> CloudBridge Relay (RU) -> Cloudflare Magic WAN"
echo -e "  CloudBridge Relay acts as a proxy/transmitter"

echo -e "${GREEN}Magic WAN integration deployment completed!${NC}"
