#!/bin/bash

# CloudBridge Phase 2 Fixes Verification Script
# This script verifies that all Phase 2 NetworkPolicy fixes are properly applied

set -e

echo "=========================================="
echo "CloudBridge Phase 2 Verification Script"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
WARNINGS=0

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        ((FAILED++))
    fi
}

print_warning() {
    echo -e "${YELLOW}⚠ WARNING${NC}: $1"
    ((WARNINGS++))
}

print_info() {
    echo -e "ℹ INFO: $1"
}

echo "=== Step 1: Verify Namespace Labels ==="
echo ""

# Check if namespaces exist and have labels
check_namespace_label() {
    local ns=$1
    local expected_label=$2

    if kubectl get namespace "$ns" &>/dev/null; then
        label=$(kubectl get namespace "$ns" -o jsonpath='{.metadata.labels.name}' 2>/dev/null || echo "")
        if [ "$label" == "$expected_label" ]; then
            print_status 0 "Namespace $ns has correct label: name=$expected_label"
        else
            print_status 1 "Namespace $ns missing or incorrect label (expected: name=$expected_label, got: $label)"
            echo "  Fix: kubectl label namespace $ns name=$expected_label"
        fi
    else
        print_warning "Namespace $ns does not exist (will be created during deployment)"
    fi
}

check_namespace_label "cloudbridge-monitoring" "cloudbridge-monitoring"
check_namespace_label "cloudbridge-relay" "cloudbridge-relay"
check_namespace_label "cloudbridge" "cloudbridge"
check_namespace_label "kube-system" "kube-system"

echo ""
echo "=== Step 2: Verify NetworkPolicy Files ==="
echo ""

# Check if NetworkPolicy files exist and are valid YAML
check_file() {
    local file=$1
    local description=$2

    if [ -f "$file" ]; then
        if kubectl apply -f "$file" --dry-run=client &>/dev/null; then
            print_status 0 "$description exists and is valid YAML"
        else
            print_status 1 "$description has invalid YAML syntax"
        fi
    else
        print_status 1 "$description not found at: $file"
    fi
}

BASE_DIR="/Users/maxlanies/Git/2GC/cloudbridge-relay-installer/oss-repositories"

check_file "$BASE_DIR/cloudbridge-monitoring/security/network-policies-fixed.yaml" "Monitoring NetworkPolicy"
check_file "$BASE_DIR/cloudbridge-scalable-relay/configs/tenant-isolation-network-policy-fixed.yaml" "Control-plane NetworkPolicy"
check_file "$BASE_DIR/cloudbridge-edge-pops/k8s/security/monitoring-network-policy.yaml" "Edge PoPs NetworkPolicy"

echo ""
echo "=== Step 3: Verify NetworkPolicy Configuration ==="
echo ""

# Check for specific NetworkPolicy rules
check_netpol_rule() {
    local file=$1
    local pattern=$2
    local description=$3

    if [ -f "$file" ]; then
        if grep -q "$pattern" "$file"; then
            print_status 0 "$description configured in $(basename $file)"
        else
            print_status 1 "$description NOT found in $(basename $file)"
        fi
    else
        print_status 1 "File not found: $file"
    fi
}

# Monitoring NetworkPolicy checks
check_netpol_rule "$BASE_DIR/cloudbridge-monitoring/security/network-policies-fixed.yaml" \
    "port: 5551" \
    "Port 5551 (edge PoP metrics)"

check_netpol_rule "$BASE_DIR/cloudbridge-monitoring/security/network-policies-fixed.yaml" \
    "name: cloudbridge-relay" \
    "Cross-namespace to cloudbridge-relay"

# Edge PoPs NetworkPolicy checks
check_netpol_rule "$BASE_DIR/cloudbridge-edge-pops/k8s/security/monitoring-network-policy.yaml" \
    "allow-prometheus-scrape-relay" \
    "Allow Prometheus scrape policy"

check_netpol_rule "$BASE_DIR/cloudbridge-edge-pops/k8s/security/monitoring-network-policy.yaml" \
    "port: 5551" \
    "Metrics port 5551 ingress"

# Control-plane NetworkPolicy checks
check_netpol_rule "$BASE_DIR/cloudbridge-scalable-relay/configs/tenant-isolation-network-policy-fixed.yaml" \
    "allow-prometheus-scrape" \
    "Allow Prometheus scrape for relay"

check_netpol_rule "$BASE_DIR/cloudbridge-scalable-relay/configs/tenant-isolation-network-policy-fixed.yaml" \
    "allow-prometheus-scrape-ai" \
    "Allow Prometheus scrape for AI service"

check_netpol_rule "$BASE_DIR/cloudbridge-scalable-relay/configs/tenant-isolation-network-policy-fixed.yaml" \
    "allow-prometheus-scrape-cadence" \
    "Allow Prometheus scrape for Cadence"

echo ""
echo "=== Step 4: Verify Prometheus Configuration ==="
echo ""

PROM_CONFIG="$BASE_DIR/cloudbridge-monitoring/prometheus/prometheus-ha-resilient.yaml"

check_netpol_rule "$PROM_CONFIG" \
    "edge-pops-relay" \
    "Edge PoPs fallback scrape job"

check_netpol_rule "$PROM_CONFIG" \
    "port: 5551" \
    "Correct metrics port (5551) in scrape config"

check_netpol_rule "$PROM_CONFIG" \
    "quic-probe" \
    "QUIC probe scrape job"

echo ""
echo "=== Step 5: Verify Documentation ==="
echo ""

README_FILE="$BASE_DIR/cloudbridge-monitoring/README.md"

if grep -q "relay-mow:5551" "$README_FILE" 2>/dev/null; then
    print_status 0 "README.md updated with correct port 5551"
else
    if grep -q "relay-mow:8080" "$README_FILE" 2>/dev/null; then
        print_status 1 "README.md still shows incorrect port 8080"
    else
        print_warning "README.md relay example not found"
    fi
fi

echo ""
echo "=== Step 6: Check if Prometheus Operator is Installed ==="
echo ""

if kubectl get deployment prometheus-operator -n cloudbridge-monitoring &>/dev/null; then
    print_status 0 "Prometheus Operator is installed"
    print_info "ServiceMonitor will be used for edge PoP scraping"
elif kubectl get deployment prometheus-operator -n monitoring &>/dev/null; then
    print_status 0 "Prometheus Operator is installed (in monitoring namespace)"
    print_info "ServiceMonitor will be used for edge PoP scraping"
else
    print_warning "Prometheus Operator not found - fallback static config will be used"
    print_info "This is OK if you're using static scrape configs"
fi

echo ""
echo "=== Step 7: Verify ServiceMonitor Configuration ==="
echo ""

SERVICEMONITOR="$BASE_DIR/cloudbridge-edge-pops/k8s/monitoring/relay-servicemonitor.yaml"

if [ -f "$SERVICEMONITOR" ]; then
    print_status 0 "ServiceMonitor definition exists"

    if grep -q "port: metrics" "$SERVICEMONITOR"; then
        print_status 0 "ServiceMonitor uses 'metrics' port name"
    else
        print_status 1 "ServiceMonitor missing 'metrics' port reference"
    fi

    if grep -q "region" "$SERVICEMONITOR"; then
        print_status 0 "ServiceMonitor has region relabeling"
    else
        print_status 1 "ServiceMonitor missing region labels"
    fi
else
    print_status 1 "ServiceMonitor file not found"
fi

echo ""
echo "=========================================="
echo "VERIFICATION SUMMARY"
echo "=========================================="
echo ""
echo -e "${GREEN}Passed:${NC}   $PASSED"
echo -e "${RED}Failed:${NC}   $FAILED"
echo -e "${YELLOW}Warnings:${NC} $WARNINGS"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All critical checks passed!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Apply namespace labels:"
    echo "   kubectl label namespace cloudbridge-monitoring name=cloudbridge-monitoring --overwrite"
    echo "   kubectl label namespace cloudbridge-relay name=cloudbridge-relay --overwrite"
    echo "   kubectl label namespace cloudbridge name=cloudbridge --overwrite"
    echo ""
    echo "2. Apply NetworkPolicies:"
    echo "   kubectl apply -f $BASE_DIR/cloudbridge-monitoring/security/network-policies-fixed.yaml"
    echo "   kubectl apply -f $BASE_DIR/cloudbridge-scalable-relay/configs/tenant-isolation-network-policy-fixed.yaml"
    echo "   kubectl apply -f $BASE_DIR/cloudbridge-edge-pops/k8s/security/monitoring-network-policy.yaml"
    echo ""
    echo "3. Deploy/Update Prometheus:"
    echo "   kubectl apply -f $BASE_DIR/cloudbridge-monitoring/prometheus/prometheus-ha-resilient.yaml"
    echo ""
    echo "4. Verify Prometheus targets:"
    echo "   kubectl port-forward -n cloudbridge-monitoring svc/prometheus 9090:9090"
    echo "   Open: http://localhost:9090/targets"
    echo ""
    exit 0
else
    echo -e "${RED}✗ Some checks failed. Please review errors above.${NC}"
    echo ""
    exit 1
fi
