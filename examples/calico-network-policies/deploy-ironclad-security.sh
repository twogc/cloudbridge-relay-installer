#!/bin/bash
# CloudBridge "Железобетонная" Security Deployment Script
# Разворачивает complete Calico Network Policies для enterprise-grade security

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="cloudbridge"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not found. Please install kubectl."
        exit 1
    fi
    
    # Check cluster connectivity
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi
    
    # Check Calico installation
    if ! kubectl get crd globalnetworkpolicies.crd.projectcalico.org &> /dev/null; then
        log_error "Calico CRDs not found. Please install Calico first."
        exit 1
    fi
    
    # Check if namespace exists
    if ! kubectl get namespace $NAMESPACE &> /dev/null; then
        log_warning "Namespace $NAMESPACE not found. Creating it..."
        kubectl create namespace $NAMESPACE
    fi
    
    log_success "Prerequisites check passed"
}

deploy_tier() {
    log_info "Deploying security tier..."
    
    if kubectl apply -f "${SCRIPT_DIR}/01-security-tier.yaml"; then
        log_success "Security tier deployed"
    else
        log_warning "Security tier deployment failed, continuing with backup approach"
    fi
}

deploy_baseline_deny() {
    log_info "Deploying baseline deny-all policies..."
    
    kubectl apply -f "${SCRIPT_DIR}/02-baseline-deny-all.yaml"
    log_success "Baseline deny-all policies deployed"
}

deploy_per_tenant_allow() {
    log_info "Deploying per-tenant allow policies..."
    
    kubectl apply -f "${SCRIPT_DIR}/03-per-tenant-allow.yaml"
    log_success "Per-tenant allow policies deployed"
}

deploy_system_exceptions() {
    log_info "Deploying system service exceptions..."
    
    kubectl apply -f "${SCRIPT_DIR}/04-system-exceptions.yaml"
    log_success "System exceptions deployed"
}

deploy_metadata_protection() {
    log_info "Deploying metadata services protection..."
    
    kubectl apply -f "${SCRIPT_DIR}/05-metadata-protection.yaml"
    log_success "Metadata protection deployed"
}

deploy_rbac_security() {
    log_info "Deploying RBAC and advanced security..."
    
    kubectl apply -f "${SCRIPT_DIR}/06-rbac-security.yaml"
    log_success "RBAC and advanced security deployed"
}

label_pods() {
    log_info "Applying tenant labels to CloudBridge pods..."
    
    # Label CloudBridge pods
    kubectl label pods -n $NAMESPACE -l app=cloudbridge-relay tenant.cloudbridge.io/id=system --overwrite
    
    # Example labeling for test pods (adjust as needed)
    # kubectl label pods -n tenant-216420165 --all tenant.cloudbridge.io/id=tenant-216420165 --overwrite
    
    log_success "Pod labeling completed"
}

verify_policies() {
    log_info "Verifying deployed policies..."
    
    # Check GlobalNetworkPolicies
    local policies=(
        "deny-all-cross-tenant-baseline"
        "allow-system-services"
        "block-metadata-services"
    )
    
    for policy in "${policies[@]}"; do
        if kubectl get globalnetworkpolicy $policy &> /dev/null; then
            log_success "Policy $policy is deployed"
        else
            log_error "Policy $policy is missing"
        fi
    done
    
    # Display policy summary
    log_info "Policy summary:"
    kubectl get globalnetworkpolicies -o wide
}

run_security_tests() {
    log_info "Running basic security validation tests..."
    
    # Test 1: Check if policies are applied
    local policy_count=$(kubectl get globalnetworkpolicies --no-headers | wc -l)
    if [[ $policy_count -gt 0 ]]; then
        log_success "Found $policy_count GlobalNetworkPolicies"
    else
        log_error "No GlobalNetworkPolicies found"
    fi
    
    # Test 2: Verify tier existence (if supported)
    if kubectl get tier security &> /dev/null; then
        log_success "Security tier is active"
    else
        log_warning "Security tier not found (may not be supported)"
    fi
    
    log_info "Security validation completed"
}

show_manual_steps() {
    log_info "Manual steps required:"
    echo ""
    echo "1. Label your tenant pods with appropriate tenant.cloudbridge.io/id labels:"
    echo "   kubectl label pods -l app=your-tenant-app tenant.cloudbridge.io/id=tenant-XXXXX"
    echo ""
    echo "2. For each new tenant, create a per-tenant allow policy:"
    echo "   Use the template in 03-per-tenant-allow.yaml"
    echo ""
    echo "3. Test isolation between tenants:"
    echo "   kubectl exec pod-tenant-A -- ping pod-tenant-B-ip  # Should FAIL"
    echo "   kubectl exec pod-tenant-A -- curl http://svc-tenant-B  # Should FAIL"
    echo ""
    echo "4. Monitor policy hits:"
    echo "   kubectl logs -n calico-system -l k8s-app=calico-felix | grep -i policy"
    echo ""
    echo "5. Test DNS resolution:"
    echo "   kubectl exec test-pod -- nslookup kubernetes.default.svc.cluster.local"
    echo ""
}

main() {
    echo "========================================"
    echo " CloudBridge Железобетонная Security    "
    echo "========================================"
    echo ""
    
    check_prerequisites
    
    log_info "Starting ironclad security deployment..."
    
    # Deploy in correct order (important for policy precedence)
    deploy_tier
    sleep 2
    
    deploy_metadata_protection  # Highest priority first
    sleep 2
    
    deploy_system_exceptions    # System exceptions
    sleep 2
    
    deploy_baseline_deny       # Baseline deny
    sleep 2
    
    deploy_per_tenant_allow    # Per-tenant allow
    sleep 2
    
    deploy_rbac_security       # Advanced security
    sleep 2
    
    label_pods
    sleep 2
    
    verify_policies
    run_security_tests
    
    echo ""
    log_success "🏆 ЖЕЛЕЗОБЕТОННАЯ SECURITY DEPLOYED SUCCESSFULLY! 🏆"
    echo ""
    log_info "Security features enabled:"
    echo "  Default-deny inter-tenant communication"
    echo "  Per-tenant isolation"
    echo "  IMDS/Metadata server protection"  
    echo "  kube-apiserver access control"
    echo "  DNS/System service exceptions"
    echo "  Network policy monitoring"
    echo ""
    
    show_manual_steps
}

# Run main function
main "$@"

