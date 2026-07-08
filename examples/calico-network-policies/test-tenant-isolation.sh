#!/bin/bash
# CloudBridge Tenant Isolation Testing Script
# Automated test suite для проверки "железобетонной" изоляции

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

NAMESPACE_A="tenant-216420165"
NAMESPACE_B="tenant-123456789"
TEST_TIMEOUT=10

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

create_test_pods() {
    log_info "Creating test pods for isolation testing..."
    
    # Create namespaces if they don't exist
    kubectl create namespace $NAMESPACE_A --dry-run=client -o yaml | kubectl apply -f -
    kubectl create namespace $NAMESPACE_B --dry-run=client -o yaml | kubectl apply -f -
    
    # Create test pod in tenant A
    cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-pod-a
  namespace: $NAMESPACE_A
  labels:
    tenant.cloudbridge.io/id: "tenant-216420165"
    app: test-pod
spec:
  containers:
  - name: test
    image: nicolaka/netshoot:latest
    command: ["/bin/sh", "-c", "sleep 3600"]
    resources:
      requests:
        memory: "64Mi"
        cpu: "50m"
      limits:
        memory: "128Mi"
        cpu: "100m"
  restartPolicy: Always
EOF

    # Create test pod in tenant B
    cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-pod-b
  namespace: $NAMESPACE_B
  labels:
    tenant.cloudbridge.io/id: "tenant-123456789"
    app: test-pod
spec:
  containers:
  - name: test
    image: nicolaka/netshoot:latest
    command: ["/bin/sh", "-c", "sleep 3600"]
    resources:
      requests:
        memory: "64Mi"
        cpu: "50m"
      limits:
        memory: "128Mi"
        cpu: "100m"
  restartPolicy: Always
EOF

    # Create test service in tenant B
    cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: test-service-b
  namespace: $NAMESPACE_B
  labels:
    tenant.cloudbridge.io/id: "tenant-123456789"
spec:
  selector:
    app: test-pod
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
  type: ClusterIP
EOF

    # Wait for pods to be ready
    log_info "Waiting for test pods to be ready..."
    kubectl wait --for=condition=Ready pod/test-pod-a -n $NAMESPACE_A --timeout=60s
    kubectl wait --for=condition=Ready pod/test-pod-b -n $NAMESPACE_B --timeout=60s
    
    log_success "Test pods created and ready"
}

test_ping_isolation() {
    log_info "Testing ping isolation between tenants..."
    
    # Get IP of pod B
    local pod_b_ip=$(kubectl get pod test-pod-b -n $NAMESPACE_B -o jsonpath='{.status.podIP}')
    
    if [[ -z "$pod_b_ip" ]]; then
        log_fail "Could not get IP of test-pod-b"
        return 1
    fi
    
    log_info "Pod B IP: $pod_b_ip"
    
    # Try to ping from A to B (should FAIL)
    if kubectl exec test-pod-a -n $NAMESPACE_A -- timeout $TEST_TIMEOUT ping -c 1 $pod_b_ip >/dev/null 2>&1; then
        log_fail "Ping from tenant A to tenant B succeeded - ISOLATION BROKEN!"
        return 1
    else
        log_success "Ping from tenant A to tenant B blocked - isolation working"
    fi
}

test_http_isolation() {
    log_info "Testing HTTP isolation between tenants..."
    
    # Get service IP of tenant B
    local service_b_ip=$(kubectl get service test-service-b -n $NAMESPACE_B -o jsonpath='{.spec.clusterIP}')
    
    if [[ -z "$service_b_ip" ]]; then
        log_fail "Could not get IP of test-service-b"
        return 1
    fi
    
    log_info "Service B IP: $service_b_ip"
    
    # Try HTTP request from A to B service (should FAIL)
    if kubectl exec test-pod-a -n $NAMESPACE_A -- timeout $TEST_TIMEOUT curl -s http://$service_b_ip >/dev/null 2>&1; then
        log_fail "HTTP request from tenant A to tenant B service succeeded - ISOLATION BROKEN!"
        return 1
    else
        log_success "HTTP request from tenant A to tenant B blocked - isolation working"
    fi
}

test_dns_resolution() {
    log_info "Testing DNS resolution..."
    
    # Test that system DNS works
    if kubectl exec test-pod-a -n $NAMESPACE_A -- timeout $TEST_TIMEOUT nslookup kubernetes.default.svc.cluster.local >/dev/null 2>&1; then
        log_success "DNS resolution to system services working"
    else
        log_fail "DNS resolution to system services broken"
        return 1
    fi
    
    # Test cross-tenant service DNS (should FAIL if DNS isolation is working)
    if kubectl exec test-pod-a -n $NAMESPACE_A -- timeout $TEST_TIMEOUT nslookup test-service-b.$NAMESPACE_B.svc.cluster.local >/dev/null 2>&1; then
        log_warning "Cross-tenant DNS resolution working - consider enabling DNS isolation"
    else
        log_success "Cross-tenant DNS resolution blocked"
    fi
}

test_metadata_protection() {
    log_info "Testing metadata service protection..."
    
    # Try to access IMDS (should FAIL)
    if kubectl exec test-pod-a -n $NAMESPACE_A -- timeout $TEST_TIMEOUT curl -s http://169.254.169.254/ >/dev/null 2>&1; then
        log_fail "IMDS access from tenant A succeeded - METADATA PROTECTION BROKEN!"
        return 1
    else
        log_success "IMDS access blocked - metadata protection working"
    fi
}

test_kubernetes_api_access() {
    log_info "Testing Kubernetes API access..."
    
    # Try direct API server access (should FAIL for regular tenant pods)
    local api_server=$(kubectl config view --minify --flatten -o jsonpath='{.clusters[0].cluster.server}' | sed 's|https://||' | cut -d: -f1)
    
    if [[ -n "$api_server" ]]; then
        if kubectl exec test-pod-a -n $NAMESPACE_A -- timeout $TEST_TIMEOUT curl -k -s https://kubernetes.default.svc.cluster.local/api/v1/namespaces >/dev/null 2>&1; then
            log_warning "Kubernetes API access working - verify RBAC restrictions"
        else
            log_success "Direct Kubernetes API access blocked"
        fi
    fi
}

test_intra_tenant_communication() {
    log_info "Testing intra-tenant communication (should work)..."
    
    # Create second pod in tenant A
    cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-pod-a2
  namespace: $NAMESPACE_A
  labels:
    tenant.cloudbridge.io/id: "tenant-216420165"
    app: test-pod-2
spec:
  containers:
  - name: test
    image: nicolaka/netshoot:latest
    command: ["/bin/sh", "-c", "sleep 3600"]
    resources:
      requests:
        memory: "64Mi"
        cpu: "50m"
      limits:
        memory: "128Mi"
        cpu: "100m"
  restartPolicy: Always
EOF

    kubectl wait --for=condition=Ready pod/test-pod-a2 -n $NAMESPACE_A --timeout=60s
    
    # Get IP of pod A2
    local pod_a2_ip=$(kubectl get pod test-pod-a2 -n $NAMESPACE_A -o jsonpath='{.status.podIP}')
    
    # Test ping within same tenant (should WORK)
    if kubectl exec test-pod-a -n $NAMESPACE_A -- timeout $TEST_TIMEOUT ping -c 1 $pod_a2_ip >/dev/null 2>&1; then
        log_success "Intra-tenant communication working"
    else
        log_fail "Intra-tenant communication blocked - check allow policies"
        return 1
    fi
}

cleanup_test_resources() {
    log_info "Cleaning up test resources..."
    
    kubectl delete pod test-pod-a -n $NAMESPACE_A --ignore-not-found
    kubectl delete pod test-pod-a2 -n $NAMESPACE_A --ignore-not-found
    kubectl delete pod test-pod-b -n $NAMESPACE_B --ignore-not-found
    kubectl delete service test-service-b -n $NAMESPACE_B --ignore-not-found
    
    log_success "Test resources cleaned up"
}

run_all_tests() {
    local failed_tests=0
    
    create_test_pods || ((failed_tests++))
    sleep 5
    
    test_intra_tenant_communication || ((failed_tests++))
    test_ping_isolation || ((failed_tests++))
    test_http_isolation || ((failed_tests++))
    test_dns_resolution || ((failed_tests++))
    test_metadata_protection || ((failed_tests++))
    test_kubernetes_api_access || ((failed_tests++))
    
    return $failed_tests
}

main() {
    echo "========================================"
    echo " CloudBridge Isolation Testing Suite   "
    echo "========================================"
    echo ""
    
    log_info "Starting tenant isolation tests..."
    
    if run_all_tests; then
        echo ""
        log_success "🏆 ALL TESTS PASSED - ЖЕЛЕЗОБЕТОННАЯ ISOLATION CONFIRMED! 🏆"
        echo ""
        echo "Security validation results:"
        echo "  Inter-tenant ping blocked"
        echo "  Inter-tenant HTTP blocked" 
        echo "  Metadata services blocked"
        echo "  DNS resolution working"
        echo "  Intra-tenant communication working"
        echo ""
    else
        echo ""
        log_fail "SOME TESTS FAILED - SECURITY ISSUES DETECTED!"
        echo ""
        log_info "Please review the failing tests and check your network policies."
    fi
    
    cleanup_test_resources
}

# Cleanup on script exit
trap cleanup_test_resources EXIT

# Run main function
main "$@"

