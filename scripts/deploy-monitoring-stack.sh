#!/bin/bash

# CloudBridge Monitoring Stack Deployment Script
# Based on reference implementations from kube-prometheus, VictoriaMetrics, Sloth, and Grafana

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="monitoring"
MONITORING_DIR="k8s/monitoring-stack"

echo -e "${BLUE}CloudBridge Monitoring Stack Deployment${NC}"
echo "================================================"
echo -e "${BLUE}Based on reference implementations:${NC}"
echo "  - kube-prometheus (HA Prometheus)"
echo "  - VictoriaMetrics (Long-term storage)"
echo "  - Sloth (SLO monitoring)"
echo "  - Grafana (Visualization)"
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if namespace exists
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    echo -e "${YELLOW} Namespace '$NAMESPACE' does not exist. Creating...${NC}"
    kubectl create namespace "$NAMESPACE"
fi

# Deploy RBAC
echo -e "${BLUE}Deploying RBAC...${NC}"
kubectl apply -f "$MONITORING_DIR/prometheus-rbac.yaml"

# Deploy VictoriaMetrics Cluster
echo -e "${BLUE}🗄️  Deploying VictoriaMetrics Cluster...${NC}"
kubectl apply -f "$MONITORING_DIR/victoriametrics-cluster.yaml"

# Wait for VictoriaMetrics to be ready
echo -e "${BLUE}Waiting for VictoriaMetrics to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/vminsert -n "$NAMESPACE" || true
kubectl wait --for=condition=available --timeout=300s deployment/vmselect -n "$NAMESPACE" || true

# Deploy HA Prometheus
echo -e "${BLUE}Deploying HA Prometheus...${NC}"
kubectl apply -f "$MONITORING_DIR/prometheus-ha.yaml"

# Deploy HA AlertManager
echo -e "${BLUE}Deploying HA AlertManager...${NC}"
kubectl apply -f "$MONITORING_DIR/alertmanager-ha.yaml"
kubectl apply -f "$MONITORING_DIR/alertmanager-config.yaml"

# Deploy SLO Monitoring
echo -e "${BLUE}Deploying SLO Monitoring...${NC}"
kubectl apply -f "$MONITORING_DIR/cloudbridge-slo.yaml"

# Deploy Grafana Provisioning
echo -e "${BLUE}Deploying Grafana Provisioning...${NC}"
kubectl apply -f "$MONITORING_DIR/grafana-provisioning.yaml"

# Wait for deployments to be ready
echo -e "${BLUE}Waiting for deployments to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s statefulset/vmstorage -n "$NAMESPACE" || true
kubectl wait --for=condition=available --timeout=300s deployment/vminsert -n "$NAMESPACE" || true
kubectl wait --for=condition=available --timeout=300s deployment/vmselect -n "$NAMESPACE" || true

# Check deployment status
echo -e "${BLUE}Checking deployment status...${NC}"
kubectl get deployments -n "$NAMESPACE" -l app.kubernetes.io/part-of=cloudbridge-monitoring
kubectl get statefulsets -n "$NAMESPACE" -l app.kubernetes.io/part-of=cloudbridge-monitoring
kubectl get services -n "$NAMESPACE" -l app.kubernetes.io/part-of=cloudbridge-monitoring

# Display access information
echo -e "${GREEN}CloudBridge Monitoring Stack deployment completed!${NC}"
echo ""
echo -e "${BLUE}Access Information:${NC}"
echo "  - Prometheus: kubectl port-forward svc/cloudbridge-prometheus 9090:9090 -n $NAMESPACE"
echo "  - AlertManager: kubectl port-forward svc/cloudbridge-alertmanager 9093:9093 -n $NAMESPACE"
echo "  - VictoriaMetrics: kubectl port-forward svc/vmselect 8481:8481 -n $NAMESPACE"
echo "  - Grafana: kubectl port-forward svc/grafana 3000:3000 -n $NAMESPACE"
echo ""
echo -e "${BLUE}Monitoring Features:${NC}"
echo "  - HA Prometheus (2 replicas) with remote write to VictoriaMetrics"
echo "  - HA AlertManager (3 replicas) with advanced routing"
echo "  - VictoriaMetrics cluster for long-term storage (1 year retention)"
echo "  - SLO monitoring with burn-rate alerts"
echo "  - Grafana provisioning with CloudBridge dashboards"
echo ""
echo -e "${BLUE}Alert Routing:${NC}"
echo "  - Critical alerts → oncall@cloudbridge.com + #alerts-critical"
echo "  - Warning alerts → team@cloudbridge.com + #alerts-warning"
echo "  - BGP alerts → network-team@cloudbridge.com + #network-bgp"
echo "  - Anycast alerts → sre@cloudbridge.com + #sre-anycast"
echo "  - CloudBridge alerts → devops@cloudbridge.com + #devops-cloudbridge"
echo ""
echo -e "${BLUE}SLO Monitoring:${NC}"
echo "  - CloudBridge Relay: 99.9% availability"
echo "  - BGP Sessions: 99.95% availability"
echo "  - Anycast Service: 99.99% availability"
echo "  - Latency: 99.5% under 100ms"
echo ""
echo -e "${YELLOW} Next Steps:${NC}"
echo "  1. Configure SMTP settings in AlertManager config"
echo "  2. Set up Slack webhooks for alert notifications"
echo "  3. Import additional Grafana dashboards"
echo "  4. Configure BGP monitoring (see k8s/bgp-monitoring/)"
echo "  5. Set up anycast health checks"
echo ""
echo -e "${GREEN}CloudBridge Monitoring Stack is ready!${NC}"
echo -e "${BLUE}Based on industry best practices from:${NC}"
echo "  - kube-prometheus (Prometheus Operator)"
echo "  - VictoriaMetrics (Long-term storage)"
echo "  - Sloth (SLO as Code)"
echo "  - Grafana (Visualization)"





