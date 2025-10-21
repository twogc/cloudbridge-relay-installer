#!/bin/bash

# CloudBridge Phase 2 Fixes Application Script
# This script applies all Phase 2 NetworkPolicy and configuration fixes

set -e

echo "=========================================="
echo "CloudBridge Phase 2 Fixes Application"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_DIR="/Users/maxlanies/Git/2GC/cloudbridge-relay-installer/oss-repositories"

# Function to print step
print_step() {
    echo ""
    echo -e "${BLUE}=== Step $1: $2 ===${NC}"
    echo ""
}

# Function to execute command with status
execute_with_status() {
    local description=$1
    shift
    local cmd="$@"

    echo -e "${YELLOW}→${NC} $description"

    if eval "$cmd"; then
        echo -e "${GREEN}✓${NC} Success"
        return 0
    else
        echo -e "${RED}✗${NC} Failed: $description"
        return 1
    fi
}

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}✗ kubectl not found. Please install kubectl first.${NC}"
    exit 1
fi

# Check if cluster is accessible
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}✗ Cannot connect to Kubernetes cluster. Please check your kubeconfig.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Connected to Kubernetes cluster${NC}"
kubectl cluster-info | head -1

print_step 1 "Create Namespaces (if not exist)"

execute_with_status "Create cloudbridge namespace" \
    "kubectl create namespace cloudbridge --dry-run=client -o yaml | kubectl apply -f -"

execute_with_status "Create cloudbridge-relay namespace" \
    "kubectl create namespace cloudbridge-relay --dry-run=client -o yaml | kubectl apply -f -"

execute_with_status "Create cloudbridge-monitoring namespace" \
    "kubectl create namespace cloudbridge-monitoring --dry-run=client -o yaml | kubectl apply -f -"

print_step 2 "Apply Namespace Labels"

execute_with_status "Label cloudbridge namespace" \
    "kubectl label namespace cloudbridge name=cloudbridge --overwrite"

execute_with_status "Label cloudbridge-relay namespace" \
    "kubectl label namespace cloudbridge-relay name=cloudbridge-relay --overwrite"

execute_with_status "Label cloudbridge-monitoring namespace" \
    "kubectl label namespace cloudbridge-monitoring name=cloudbridge-monitoring --overwrite"

execute_with_status "Label kube-system namespace" \
    "kubectl label namespace kube-system name=kube-system --overwrite"

print_step 3 "Apply NetworkPolicies"

echo "Applying Monitoring NetworkPolicies..."
execute_with_status "Apply monitoring security NetworkPolicies" \
    "kubectl apply -f $BASE_DIR/cloudbridge-monitoring/security/network-policies-fixed.yaml"

echo ""
echo "Applying Control-Plane NetworkPolicies..."
execute_with_status "Apply control-plane tenant isolation NetworkPolicies" \
    "kubectl apply -f $BASE_DIR/cloudbridge-scalable-relay/configs/tenant-isolation-network-policy-fixed.yaml"

echo ""
echo "Applying Edge PoPs NetworkPolicies..."
execute_with_status "Apply edge PoPs monitoring NetworkPolicies" \
    "kubectl apply -f $BASE_DIR/cloudbridge-edge-pops/k8s/security/monitoring-network-policy.yaml"

print_step 4 "Verify NetworkPolicy Application"

echo "Checking NetworkPolicies in cloudbridge-monitoring namespace:"
kubectl get networkpolicies -n cloudbridge-monitoring

echo ""
echo "Checking NetworkPolicies in cloudbridge namespace:"
kubectl get networkpolicies -n cloudbridge

echo ""
echo "Checking NetworkPolicies in cloudbridge-relay namespace:"
kubectl get networkpolicies -n cloudbridge-relay

print_step 5 "Update Prometheus Configuration"

echo -e "${YELLOW}Note:${NC} Prometheus configuration update requires pod restart to take effect."
echo ""

execute_with_status "Apply Prometheus HA configuration" \
    "kubectl apply -f $BASE_DIR/cloudbridge-monitoring/prometheus/prometheus-ha-resilient.yaml"

print_step 6 "Restart Prometheus Pods (if already running)"

if kubectl get deployment prometheus -n cloudbridge-monitoring &>/dev/null; then
    execute_with_status "Rollout restart Prometheus deployment" \
        "kubectl rollout restart deployment prometheus -n cloudbridge-monitoring"

    echo ""
    echo "Waiting for Prometheus rollout to complete..."
    kubectl rollout status deployment prometheus -n cloudbridge-monitoring --timeout=120s || true
elif kubectl get statefulset prometheus -n cloudbridge-monitoring &>/dev/null; then
    execute_with_status "Rollout restart Prometheus StatefulSet" \
        "kubectl rollout restart statefulset prometheus -n cloudbridge-monitoring"

    echo ""
    echo "Waiting for Prometheus rollout to complete..."
    kubectl rollout status statefulset prometheus -n cloudbridge-monitoring --timeout=120s || true
else
    echo -e "${YELLOW}⚠${NC} Prometheus deployment not found (will be created during full deployment)"
fi

print_step 7 "Verification"

echo "Verifying namespace labels:"
kubectl get namespace cloudbridge cloudbridge-relay cloudbridge-monitoring kube-system -o custom-columns=NAME:.metadata.name,LABEL:.metadata.labels.name

echo ""
echo "Verifying NetworkPolicies count:"
echo -e "cloudbridge-monitoring: $(kubectl get networkpolicies -n cloudbridge-monitoring --no-headers 2>/dev/null | wc -l) policies"
echo -e "cloudbridge:            $(kubectl get networkpolicies -n cloudbridge --no-headers 2>/dev/null | wc -l) policies"
echo -e "cloudbridge-relay:      $(kubectl get networkpolicies -n cloudbridge-relay --no-headers 2>/dev/null | wc -l) policies"

echo ""
echo "=========================================="
echo -e "${GREEN}✓ Phase 2 Fixes Applied Successfully!${NC}"
echo "=========================================="
echo ""

echo "Next Steps:"
echo ""
echo "1. Verify Prometheus can scrape edge PoPs:"
echo "   kubectl port-forward -n cloudbridge-monitoring svc/prometheus 9090:9090"
echo "   Open: http://localhost:9090/targets"
echo ""
echo "2. Check for 'edge-pops-relay' job in targets list"
echo ""
echo "3. Deploy edge PoPs if not already deployed:"
echo "   kubectl apply -f $BASE_DIR/cloudbridge-edge-pops/k8s/relay-mow/"
echo "   kubectl apply -f $BASE_DIR/cloudbridge-edge-pops/k8s/relay-fra/"
echo "   kubectl apply -f $BASE_DIR/cloudbridge-edge-pops/k8s/relay-ams/"
echo ""
echo "4. Verify metrics are being scraped:"
echo "   # In Prometheus UI, run query:"
echo "   # relay_active_connections{region=\"mow\"}"
echo ""
echo "For detailed verification, run:"
echo "   bash scripts/verify-phase2-fixes.sh"
echo ""
