#!/bin/bash

# CloudBridge Relay - Security Scanning Script
# This script performs security scans on the system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="cloudbridge"
MONITORING_NAMESPACE="cloudbridge-monitoring"
SCAN_DIR="security/scanning"

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}$1${NC}"
}

log_warning() {
    echo -e "${YELLOW}  $1${NC}"
}

log_error() {
    echo -e "${RED}$1${NC}"
}

# Check if tools are installed
check_prerequisites() {
    log_info "Checking security scanning prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if trivy is installed
    if ! command -v trivy &> /dev/null; then
        log_warning "Trivy is not installed. Installing Trivy..."
        install_trivy
    fi
    
    # Check if kube-score is installed
    if ! command -v kube-score &> /dev/null; then
        log_warning "kube-score is not installed. Installing kube-score..."
        install_kube_score
    fi
    
    log_success "Prerequisites check passed"
}

# Install Trivy
install_trivy() {
    log_info "Installing Trivy..."
    
    # Download and install Trivy
    curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
    
    log_success "Trivy installed"
}

# Install kube-score
install_kube_score() {
    log_info "Installing kube-score..."
    
    # Download and install kube-score
    curl -L https://github.com/zegl/kube-score/releases/latest/download/kube-score_linux_amd64.tar.gz | tar xz
    sudo mv kube-score /usr/local/bin/
    
    log_success "kube-score installed"
}

# Scan Kubernetes manifests
scan_manifests() {
    log_info "Scanning Kubernetes manifests..."
    
    # Create scan directory
    mkdir -p $SCAN_DIR/results
    
    # Scan with kube-score
    log_info "Running kube-score analysis..."
    kube-score score k8s/ --output-format json > $SCAN_DIR/results/kube-score-results.json 2>/dev/null || true
    kube-score score k8s/ --output-format human > $SCAN_DIR/results/kube-score-results.txt 2>/dev/null || true
    
    # Scan with Trivy
    log_info "Running Trivy vulnerability scan..."
    trivy k8s cluster --format json --output $SCAN_DIR/results/trivy-k8s-results.json 2>/dev/null || true
    trivy k8s cluster --format table --output $SCAN_DIR/results/trivy-k8s-results.txt 2>/dev/null || true
    
    log_success "Manifest scanning completed"
}

# Scan Docker images
scan_images() {
    log_info "Scanning Docker images..."
    
    # List of images to scan
    images=(
        "cloudbridge-relay:latest"
        "cassandra:3.11"
        "ubercadence/server:latest"
        "prom/prometheus:latest"
        "grafana/grafana:latest"
        "prom/alertmanager:latest"
    )
    
    for image in "${images[@]}"; do
        log_info "Scanning image: $image"
        trivy image --format json --output $SCAN_DIR/results/trivy-image-${image//[\/:]/_}-results.json $image 2>/dev/null || true
        trivy image --format table --output $SCAN_DIR/results/trivy-image-${image//[\/:]/_}-results.txt $image 2>/dev/null || true
    done
    
    log_success "Image scanning completed"
}

# Check security policies
check_security_policies() {
    log_info "Checking security policies..."
    
    # Check Network Policies
    log_info "Checking Network Policies..."
    kubectl get networkpolicies -A > $SCAN_DIR/results/network-policies.txt 2>/dev/null || true
    
    # Check Pod Security Policies
    log_info "Checking Pod Security Policies..."
    kubectl get podsecuritypolicies > $SCAN_DIR/results/pod-security-policies.txt 2>/dev/null || true
    
    # Check RBAC
    log_info "Checking RBAC..."
    kubectl get clusterroles > $SCAN_DIR/results/cluster-roles.txt 2>/dev/null || true
    kubectl get clusterrolebindings > $SCAN_DIR/results/cluster-role-bindings.txt 2>/dev/null || true
    
    # Check Service Accounts
    log_info "Checking Service Accounts..."
    kubectl get serviceaccounts -A > $SCAN_DIR/results/service-accounts.txt 2>/dev/null || true
    
    log_success "Security policies check completed"
}

# Check resource usage
check_resource_usage() {
    log_info "Checking resource usage..."
    
    # Check node resources
    log_info "Checking node resources..."
    kubectl top nodes > $SCAN_DIR/results/node-resources.txt 2>/dev/null || true
    kubectl describe nodes > $SCAN_DIR/results/node-description.txt 2>/dev/null || true
    
    # Check pod resources
    log_info "Checking pod resources..."
    kubectl top pods -A > $SCAN_DIR/results/pod-resources.txt 2>/dev/null || true
    
    # Check resource quotas
    log_info "Checking resource quotas..."
    kubectl get resourcequotas -A > $SCAN_DIR/results/resource-quotas.txt 2>/dev/null || true
    kubectl get limitranges -A > $SCAN_DIR/results/limit-ranges.txt 2>/dev/null || true
    
    log_success "Resource usage check completed"
}

# Check security configurations
check_security_configs() {
    log_info "Checking security configurations..."
    
    # Check if pods are running as non-root
    log_info "Checking non-root containers..."
    kubectl get pods -A -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.securityContext.runAsUser}{"\t"}{.spec.containers[*].securityContext.runAsUser}{"\n"}{end}' > $SCAN_DIR/results/non-root-check.txt 2>/dev/null || true
    
    # Check if pods have read-only root filesystem
    log_info "Checking read-only root filesystem..."
    kubectl get pods -A -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].securityContext.readOnlyRootFilesystem}{"\n"}{end}' > $SCAN_DIR/results/readonly-fs-check.txt 2>/dev/null || true
    
    # Check if pods drop all capabilities
    log_info "Checking dropped capabilities..."
    kubectl get pods -A -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[*].securityContext.capabilities.drop}{"\n"}{end}' > $SCAN_DIR/results/capabilities-check.txt 2>/dev/null || true
    
    log_success "Security configurations check completed"
}

# Generate security report
generate_report() {
    log_info "Generating security report..."
    
    cat > $SCAN_DIR/results/security-report.md << EOF
# CloudBridge Relay - Security Scan Report

## Scan Date
$(date)

## Summary
This report contains the results of security scans performed on the CloudBridge Relay system.

## Scans Performed
1. Kubernetes Manifest Analysis (kube-score)
2. Vulnerability Scanning (Trivy)
3. Security Policies Check
4. Resource Usage Analysis
5. Security Configuration Check

## Results
- **Manifest Analysis**: See kube-score-results.txt
- **Vulnerability Scan**: See trivy-k8s-results.txt
- **Network Policies**: See network-policies.txt
- **Pod Security Policies**: See pod-security-policies.txt
- **RBAC**: See cluster-roles.txt and cluster-role-bindings.txt
- **Resource Usage**: See node-resources.txt and pod-resources.txt
- **Security Configs**: See non-root-check.txt, readonly-fs-check.txt, capabilities-check.txt

## Recommendations
1. Review and address any high/critical vulnerabilities
2. Ensure all pods run as non-root users
3. Enable read-only root filesystem where possible
4. Drop unnecessary capabilities
5. Implement proper resource limits
6. Review and update security policies regularly

## Files Generated
$(ls -la $SCAN_DIR/results/)

EOF
    
    log_success "Security report generated: $SCAN_DIR/results/security-report.md"
}

# Show scan results
show_results() {
    log_info "Security Scan Results:"
    echo ""
    
    echo "Scan Results Summary:"
    echo "• Manifest Analysis: $SCAN_DIR/results/kube-score-results.txt"
    echo "• Vulnerability Scan: $SCAN_DIR/results/trivy-k8s-results.txt"
    echo "• Security Policies: $SCAN_DIR/results/network-policies.txt"
    echo "• Resource Usage: $SCAN_DIR/results/node-resources.txt"
    echo "• Security Report: $SCAN_DIR/results/security-report.md"
    echo ""
    
    echo "Quick Results:"
    if [ -f "$SCAN_DIR/results/kube-score-results.txt" ]; then
        echo "Kube-score issues found:"
        grep -c "CRITICAL\|WARNING" $SCAN_DIR/results/kube-score-results.txt 2>/dev/null || echo "0"
    fi
    
    if [ -f "$SCAN_DIR/results/trivy-k8s-results.txt" ]; then
        echo "Trivy vulnerabilities found:"
        grep -c "HIGH\|CRITICAL" $SCAN_DIR/results/trivy-k8s-results.txt 2>/dev/null || echo "0"
    fi
    
    echo ""
    echo "Resource Usage:"
    if [ -f "$SCAN_DIR/results/node-resources.txt" ]; then
        cat $SCAN_DIR/results/node-resources.txt
    fi
    
    echo ""
    echo "Pod Resources:"
    if [ -f "$SCAN_DIR/results/pod-resources.txt" ]; then
        cat $SCAN_DIR/results/pod-resources.txt
    fi
}

# Cleanup old results
cleanup() {
    log_warning "Cleaning up old scan results..."
    
    rm -rf $SCAN_DIR/results/*
    
    log_success "Cleanup completed"
}

# Main function
main() {
    echo "=== CloudBridge Relay - Security Scanning ==="
    echo ""
    
    case "${1:-scan}" in
        "scan")
            check_prerequisites
            scan_manifests
            scan_images
            check_security_policies
            check_resource_usage
            check_security_configs
            generate_report
            show_results
            log_success "Security scanning completed successfully!"
            ;;
        "results")
            show_results
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [scan|results|cleanup|help]"
            echo ""
            echo "Commands:"
            echo "  scan     - Run full security scan (default)"
            echo "  results  - Show scan results"
            echo "  cleanup  - Clean up old scan results"
            echo "  help     - Show this help message"
            ;;
        *)
            log_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"

