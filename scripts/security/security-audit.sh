#!/bin/bash

# CloudBridge Relay Security Audit Script
# This script performs comprehensive security checks on the staging environment

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
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

# Check if kubectl is available
check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    log_success "kubectl is available"
}

# Check if minikube is running
check_minikube() {
    if ! minikube status &> /dev/null; then
        log_error "minikube is not running"
        exit 1
    fi
    log_success "minikube is running"
}

# Security check 1: Pod Security Contexts
check_pod_security_contexts() {
    log_info "Checking pod security contexts..."
    
    local namespace="cloudbridge-staging"
    local pods=$(kubectl get pods -n $namespace -o jsonpath='{.items[*].metadata.name}')
    
    for pod in $pods; do
        log_info "Checking pod: $pod"
        
        # Check if pod runs as non-root
        local runAsUser=$(kubectl get pod $pod -n $namespace -o jsonpath='{.spec.securityContext.runAsUser}')
        if [[ "$runAsUser" == "1000" ]]; then
            log_success "Pod $pod runs as non-root user (1000)"
        else
            log_warning "Pod $pod does not run as non-root user: $runAsUser"
        fi
        
        # Check if pod has read-only root filesystem
        local readOnlyRootFS=$(kubectl get pod $pod -n $namespace -o jsonpath='{.spec.containers[0].securityContext.readOnlyRootFilesystem}')
        if [[ "$readOnlyRootFS" == "true" ]]; then
            log_success "Pod $pod has read-only root filesystem"
        else
            log_warning "Pod $pod does not have read-only root filesystem"
        fi
        
        # Check if pod has privileged access
        local privileged=$(kubectl get pod $pod -n $namespace -o jsonpath='{.spec.containers[0].securityContext.privileged}')
        if [[ "$privileged" == "true" ]]; then
            log_error "Pod $pod runs in privileged mode - SECURITY RISK!"
        else
            log_success "Pod $pod does not run in privileged mode"
        fi
    done
}

# Security check 2: Network Policies
check_network_policies() {
    log_info "Checking network policies..."
    
    local namespace="cloudbridge-staging"
    local policies=$(kubectl get networkpolicies -n $namespace -o jsonpath='{.items[*].metadata.name}')
    
    if [[ -z "$policies" ]]; then
        log_warning "No network policies found in namespace $namespace"
    else
        log_success "Network policies found: $policies"
    fi
}

# Security check 3: RBAC Configuration
check_rbac() {
    log_info "Checking RBAC configuration..."
    
    local namespace="cloudbridge-staging"
    
    # Check service accounts
    local serviceAccounts=$(kubectl get serviceaccounts -n $namespace -o jsonpath='{.items[*].metadata.name}')
    log_info "Service accounts in $namespace: $serviceAccounts"
    
    # Check roles
    local roles=$(kubectl get roles -n $namespace -o jsonpath='{.items[*].metadata.name}')
    if [[ -n "$roles" ]]; then
        log_success "Roles found: $roles"
    else
        log_warning "No roles found in namespace $namespace"
    fi
    
    # Check role bindings
    local roleBindings=$(kubectl get rolebindings -n $namespace -o jsonpath='{.items[*].metadata.name}')
    if [[ -n "$roleBindings" ]]; then
        log_success "Role bindings found: $roleBindings"
    else
        log_warning "No role bindings found in namespace $namespace"
    fi
}

# Security check 4: Secrets Management
check_secrets() {
    log_info "Checking secrets management..."
    
    local namespace="cloudbridge-staging"
    local secrets=$(kubectl get secrets -n $namespace -o jsonpath='{.items[*].metadata.name}')
    
    if [[ -n "$secrets" ]]; then
        log_success "Secrets found: $secrets"
        
        # Check if secrets are properly labeled
        for secret in $secrets; do
            local labels=$(kubectl get secret $secret -n $namespace -o jsonpath='{.metadata.labels}')
            if [[ "$labels" == "{}" ]]; then
                log_warning "Secret $secret has no labels"
            else
                log_success "Secret $secret has labels: $labels"
            fi
        done
    else
        log_warning "No secrets found in namespace $namespace"
    fi
}

# Security check 5: Resource Limits
check_resource_limits() {
    log_info "Checking resource limits..."
    
    local namespace="cloudbridge-staging"
    local pods=$(kubectl get pods -n $namespace -o jsonpath='{.items[*].metadata.name}')
    
    for pod in $pods; do
        log_info "Checking resource limits for pod: $pod"
        
        # Check CPU limits
        local cpuLimit=$(kubectl get pod $pod -n $namespace -o jsonpath='{.spec.containers[0].resources.limits.cpu}')
        if [[ -n "$cpuLimit" ]]; then
            log_success "Pod $pod has CPU limit: $cpuLimit"
        else
            log_warning "Pod $pod has no CPU limit"
        fi
        
        # Check memory limits
        local memoryLimit=$(kubectl get pod $pod -n $namespace -o jsonpath='{.spec.containers[0].resources.limits.memory}')
        if [[ -n "$memoryLimit" ]]; then
            log_success "Pod $pod has memory limit: $memoryLimit"
        else
            log_warning "Pod $pod has no memory limit"
        fi
    done
}

# Security check 6: Image Security
check_image_security() {
    log_info "Checking image security..."
    
    local namespace="cloudbridge-staging"
    local pods=$(kubectl get pods -n $namespace -o jsonpath='{.items[*].metadata.name}')
    
    for pod in $pods; do
        local image=$(kubectl get pod $pod -n $namespace -o jsonpath='{.spec.containers[0].image}')
        log_info "Pod $pod uses image: $image"
        
        # Check if image has a specific tag (not latest)
        if [[ "$image" == *":latest" ]]; then
            log_warning "Pod $pod uses 'latest' tag - consider using specific version"
        else
            log_success "Pod $pod uses specific image tag"
        fi
        
        # Check image pull policy
        local imagePullPolicy=$(kubectl get pod $pod -n $namespace -o jsonpath='{.spec.containers[0].imagePullPolicy}')
        log_info "Pod $pod image pull policy: $imagePullPolicy"
    done
}

# Security check 7: Pod Disruption Budgets
check_pod_disruption_budgets() {
    log_info "Checking pod disruption budgets..."
    
    local namespace="cloudbridge-staging"
    local pdbs=$(kubectl get poddisruptionbudgets -n $namespace -o jsonpath='{.items[*].metadata.name}')
    
    if [[ -n "$pdbs" ]]; then
        log_success "Pod disruption budgets found: $pdbs"
    else
        log_warning "No pod disruption budgets found in namespace $namespace"
    fi
}

# Generate security report
generate_security_report() {
    log_info "Generating security audit report..."
    
    local report_file="security-audit-report-$(date +%Y%m%d_%H%M%S).txt"
    
    cat > "$report_file" << EOF
CloudBridge Relay Security Audit Report
======================================

Audit Date: $(date)
Namespace: cloudbridge-staging
Auditor: Security Audit Script

EXECUTIVE SUMMARY
================
This report provides a comprehensive security assessment of the CloudBridge Relay
staging environment deployed in Kubernetes.

SECURITY CHECKS PERFORMED
=========================
1. Pod Security Contexts
2. Network Policies
3. RBAC Configuration
4. Secrets Management
5. Resource Limits
6. Image Security
7. Pod Disruption Budgets

RECOMMENDATIONS
===============
1. Implement Network Policies for micro-segmentation
2. Use Pod Security Policies for additional security
3. Ensure all secrets are properly labeled and managed
4. Set resource limits for all containers
5. Use specific image tags instead of 'latest'
6. Implement Pod Disruption Budgets for high availability

NEXT STEPS
==========
1. Review and implement missing security controls
2. Regular security audits and updates
3. Monitor security events and alerts
4. Implement security scanning in CI/CD pipeline

EOF
    
    log_success "Security audit report generated: $report_file"
}

# Main execution
main() {
    log_info "Starting CloudBridge Relay Security Audit..."
    
    check_kubectl
    check_minikube
    
    check_pod_security_contexts
    check_network_policies
    check_rbac
    check_secrets
    check_resource_limits
    check_image_security
    check_pod_disruption_budgets
    
    generate_security_report
    
    log_success "Security audit completed successfully!"
}

# Run main function
main "$@"

