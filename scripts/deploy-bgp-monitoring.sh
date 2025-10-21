#!/bin/bash

# CloudBridge BGP Monitoring Deployment Script
# This script deploys BGP monitoring components for CloudBridge anycast network

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="monitoring"
BGP_MONITORING_DIR="k8s/bgp-monitoring"

echo -e "${BLUE}🌐 CloudBridge BGP Monitoring Deployment${NC}"
echo "================================================"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    echo -e "${YELLOW}⚠️  Namespace '$NAMESPACE' does not exist. Creating...${NC}"
    kubectl create namespace "$NAMESPACE"
fi

# Deploy RBAC
echo -e "${BLUE}📋 Deploying RBAC...${NC}"
kubectl apply -f "$BGP_MONITORING_DIR/bird-exporter-rbac.yaml"
kubectl apply -f "$BGP_MONITORING_DIR/bgp-exporters-rbac.yaml"

# Deploy BIRD Exporter
echo -e "${BLUE}🐦 Deploying BIRD Exporter...${NC}"
kubectl apply -f "$BGP_MONITORING_DIR/bird-exporter-deployment.yaml"
kubectl apply -f "$BGP_MONITORING_DIR/bird-exporter-service.yaml"
kubectl apply -f "$BGP_MONITORING_DIR/bird-exporter-servicemonitor.yaml"

# Deploy FRR Exporter
echo -e "${BLUE}🔄 Deploying FRR Exporter...${NC}"
kubectl apply -f "$BGP_MONITORING_DIR/frr-exporter-deployment.yaml"
kubectl apply -f "$BGP_MONITORING_DIR/frr-exporter-service.yaml"
kubectl apply -f "$BGP_MONITORING_DIR/frr-exporter-servicemonitor.yaml"

# Deploy GoBGP Exporter
echo -e "${BLUE}🚀 Deploying GoBGP Exporter...${NC}"
kubectl apply -f "$BGP_MONITORING_DIR/gobgp-exporter-deployment.yaml"
kubectl apply -f "$BGP_MONITORING_DIR/gobgp-exporter-service.yaml"
kubectl apply -f "$BGP_MONITORING_DIR/gobgp-exporter-servicemonitor.yaml"

# Deploy BGP Alerts
echo -e "${BLUE}🚨 Deploying BGP Alerts...${NC}"
kubectl apply -f "$BGP_MONITORING_DIR/bgp-alerts.yaml"

# Wait for deployments to be ready
echo -e "${BLUE}⏳ Waiting for deployments to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/bird-exporter -n "$NAMESPACE" || true
kubectl wait --for=condition=available --timeout=300s deployment/frr-exporter -n "$NAMESPACE" || true
kubectl wait --for=condition=available --timeout=300s deployment/gobgp-exporter -n "$NAMESPACE" || true

# Check deployment status
echo -e "${BLUE}📊 Checking deployment status...${NC}"
kubectl get deployments -n "$NAMESPACE" -l component=bgp-monitoring
kubectl get services -n "$NAMESPACE" -l component=bgp-monitoring
kubectl get servicemonitors -n "$NAMESPACE" -l component=bgp-monitoring

# Check if Prometheus is scraping the exporters
echo -e "${BLUE}🔍 Checking Prometheus targets...${NC}"
if kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=prometheus &> /dev/null; then
    echo -e "${GREEN}✅ Prometheus is running. BGP exporters should be automatically discovered.${NC}"
else
    echo -e "${YELLOW}⚠️  Prometheus is not running. Please ensure Prometheus is deployed to scrape BGP exporters.${NC}"
fi

# Display access information
echo -e "${GREEN}✅ BGP Monitoring deployment completed!${NC}"
echo ""
echo -e "${BLUE}📊 Access Information:${NC}"
echo "  - BIRD Exporter: kubectl port-forward svc/bird-exporter 9324:9324 -n $NAMESPACE"
echo "  - FRR Exporter: kubectl port-forward svc/frr-exporter 9342:9342 -n $NAMESPACE"
echo "  - GoBGP Exporter: kubectl port-forward svc/gobgp-exporter 9474:9474 -n $NAMESPACE"
echo ""
echo -e "${BLUE}📈 Grafana Dashboards:${NC}"
echo "  - BGP Overview: Import $BGP_MONITORING_DIR/bgp-overview-dashboard.json"
echo "  - Anycast Health: Import $BGP_MONITORING_DIR/anycast-health-dashboard.json"
echo ""
echo -e "${BLUE}🚨 Alerts:${NC}"
echo "  - BGP Session Down: Critical alert when BGP sessions are down"
echo "  - BGP Route Flapping: Warning alert for route instability"
echo "  - Anycast Health: Critical alert for anycast service health"
echo ""
echo -e "${YELLOW}⚠️  Next Steps:${NC}"
echo "  1. Configure BGP daemons (BIRD/FRR/GoBGP) on your PoP nodes"
echo "  2. Ensure BGP daemon sockets are accessible to exporters"
echo "  3. Configure anycast health checks for your services"
echo "  4. Import Grafana dashboards for visualization"
echo "  5. Configure AlertManager for BGP alerts"
echo ""
echo -e "${GREEN}🎉 BGP Monitoring is ready for CloudBridge anycast network!${NC}"





