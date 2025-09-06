#!/bin/bash

# CloudBridge Relay - Disaster Recovery Script
# Этот скрипт выполняет восстановление системы после сбоев

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="${NAMESPACE:-cloudbridge}"
VELERO_NAMESPACE="${VELERO_NAMESPACE:-velero}"
BACKUP_NAME="${BACKUP_NAME:-}"
RECOVERY_MODE="${RECOVERY_MODE:-full}"  # full, partial, config-only
DRY_RUN="${DRY_RUN:-false}"
FORCE_RECOVERY="${FORCE_RECOVERY:-false}"

# Recovery time objectives
RTO_TARGET="${RTO_TARGET:-240}"  # 4 hours in minutes
RPO_TARGET="${RPO_TARGET:-60}"   # 1 hour in minutes

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_section() {
    echo -e "${PURPLE}[SECTION]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    for tool in kubectl velero curl jq; do
        if ! command -v "$tool" &> /dev/null; then
            missing_deps+=("$tool")
        fi
    done
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        exit 1
    fi
    
    # Check if kubectl can connect to cluster
    if ! kubectl cluster-info &>/dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    # Check if Velero is available
    if ! kubectl get deployment velero -n "$VELERO_NAMESPACE" &>/dev/null; then
        log_error "Velero is not installed or not running"
        exit 1
    fi
    
    log_success "Dependencies check passed"
}

# Assess damage and determine recovery strategy
assess_damage() {
    log_section "Assessing system damage..."
    
    local damage_level="unknown"
    local affected_components=()
    local issues=()
    
    # Check namespace status
    if ! kubectl get namespace "$NAMESPACE" &>/dev/null; then
        damage_level="critical"
        issues+=("Namespace $NAMESPACE does not exist")
    else
        # Check deployments
        local deployments=$(kubectl get deployments -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
        if [ "$deployments" -eq 0 ]; then
            damage_level="critical"
            issues+=("No deployments found in namespace $NAMESPACE")
        else
            # Check deployment health
            local unhealthy_deployments=$(kubectl get deployments -n "$NAMESPACE" --no-headers | grep -v "1/1" | wc -l)
            if [ "$unhealthy_deployments" -gt 0 ]; then
                damage_level="high"
                issues+=("$unhealthy_deployments unhealthy deployments")
            fi
        fi
        
        # Check services
        local services=$(kubectl get services -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
        if [ "$services" -eq 0 ]; then
            damage_level="high"
            issues+=("No services found in namespace $NAMESPACE")
        fi
        
        # Check persistent volumes
        local pvcs=$(kubectl get pvc -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
        if [ "$pvcs" -eq 0 ]; then
            damage_level="medium"
            issues+=("No persistent volume claims found")
        fi
    fi
    
    # Check Velero backups
    local available_backups=$(velero backup get --namespace="$VELERO_NAMESPACE" --no-headers 2>/dev/null | wc -l)
    if [ "$available_backups" -eq 0 ]; then
        damage_level="critical"
        issues+=("No Velero backups available")
    fi
    
    # Determine recovery strategy based on damage level
    case $damage_level in
        critical)
            RECOVERY_MODE="full"
            log_warning "Critical damage detected - full recovery required"
            ;;
        high)
            RECOVERY_MODE="full"
            log_warning "High damage detected - full recovery recommended"
            ;;
        medium)
            RECOVERY_MODE="partial"
            log_info "Medium damage detected - partial recovery sufficient"
            ;;
        *)
            RECOVERY_MODE="config-only"
            log_info "Minimal damage detected - configuration recovery sufficient"
            ;;
    esac
    
    log_info "Damage assessment completed:"
    log_info "  - Damage level: $damage_level"
    log_info "  - Recovery mode: $RECOVERY_MODE"
    log_info "  - Issues found: ${#issues[@]}"
    for issue in "${issues[@]}"; do
        log_info "    - $issue"
    done
    
    echo "$damage_level"
}

# Find latest backup
find_latest_backup() {
    log_section "Finding latest backup..."
    
    if [ -n "$BACKUP_NAME" ]; then
        log_info "Using specified backup: $BACKUP_NAME"
        echo "$BACKUP_NAME"
        return 0
    fi
    
    # Find latest successful backup
    local latest_backup=$(velero backup get --namespace="$VELERO_NAMESPACE" --no-headers | \
        grep "Completed" | head -1 | awk '{print $1}')
    
    if [ -z "$latest_backup" ]; then
        log_error "No completed backups found"
        exit 1
    fi
    
    log_info "Found latest backup: $latest_backup"
    echo "$latest_backup"
}

# Validate backup
validate_backup() {
    local backup_name="$1"
    
    log_section "Validating backup: $backup_name"
    
    # Check backup status
    local backup_status=$(velero backup get "$backup_name" --namespace="$VELERO_NAMESPACE" --no-headers | awk '{print $2}')
    
    if [ "$backup_status" != "Completed" ]; then
        log_error "Backup $backup_name is not completed (status: $backup_status)"
        exit 1
    fi
    
    # Get backup details
    local backup_info=$(velero backup describe "$backup_name" --namespace="$VELERO_NAMESPACE" --details)
    
    # Check if backup contains required namespaces
    if ! echo "$backup_info" | grep -q "Included namespaces: $NAMESPACE"; then
        log_warning "Backup may not contain namespace $NAMESPACE"
    fi
    
    # Check backup age
    local backup_time=$(velero backup get "$backup_name" --namespace="$VELERO_NAMESPACE" --no-headers | awk '{print $3, $4}')
    log_info "Backup created: $backup_time"
    
    log_success "Backup validation completed"
}

# Pre-recovery checks
pre_recovery_checks() {
    log_section "Performing pre-recovery checks..."
    
    # Check cluster resources
    local nodes=$(kubectl get nodes --no-headers | wc -l)
    if [ "$nodes" -eq 0 ]; then
        log_error "No nodes available in cluster"
        exit 1
    fi
    
    # Check available storage
    local available_storage=$(kubectl get nodes -o json | jq -r '.items[].status.allocatable."ephemeral-storage"' | head -1)
    log_info "Available storage: $available_storage"
    
    # Check network connectivity
    if ! kubectl get pods -n kube-system &>/dev/null; then
        log_error "Cannot access kube-system namespace"
        exit 1
    fi
    
    # Check Velero status
    if ! kubectl get deployment velero -n "$VELERO_NAMESPACE" &>/dev/null; then
        log_error "Velero is not running"
        exit 1
    fi
    
    log_success "Pre-recovery checks completed"
}

# Stop existing services
stop_existing_services() {
    log_section "Stopping existing services..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would stop existing services in namespace $NAMESPACE"
        return 0
    fi
    
    # Scale down deployments
    local deployments=$(kubectl get deployments -n "$NAMESPACE" --no-headers 2>/dev/null | awk '{print $1}' || true)
    for deployment in $deployments; do
        log_info "Scaling down deployment: $deployment"
        kubectl scale deployment "$deployment" --replicas=0 -n "$NAMESPACE"
    done
    
    # Wait for pods to terminate
    log_info "Waiting for pods to terminate..."
    kubectl wait --for=delete pod --all -n "$NAMESPACE" --timeout=300s || true
    
    log_success "Existing services stopped"
}

# Restore from backup
restore_from_backup() {
    local backup_name="$1"
    
    log_section "Restoring from backup: $backup_name"
    
    local restore_name="disaster-recovery-$(date +%Y%m%d-%H%M%S)"
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would restore backup $backup_name as $restore_name"
        return 0
    fi
    
    # Create restore
    log_info "Creating restore: $restore_name"
    velero restore create "$restore_name" \
        --from-backup "$backup_name" \
        --include-namespaces "$NAMESPACE" \
        --restore-pvs \
        --namespace="$VELERO_NAMESPACE"
    
    # Wait for restore to complete
    log_info "Waiting for restore to complete..."
    velero restore get "$restore_name" --namespace="$VELERO_NAMESPACE"
    
    # Check restore status
    local restore_status=$(velero restore get "$restore_name" --namespace="$VELERO_NAMESPACE" --no-headers | awk '{print $2}')
    
    if [ "$restore_status" = "Completed" ]; then
        log_success "Restore completed successfully"
    else
        log_error "Restore failed with status: $restore_status"
        exit 1
    fi
    
    echo "$restore_name"
}

# Restore configurations only
restore_configurations() {
    local backup_name="$1"
    
    log_section "Restoring configurations only..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would restore configurations from backup $backup_name"
        return 0
    fi
    
    # Extract and apply ConfigMaps
    log_info "Restoring ConfigMaps..."
    velero restore create "config-restore-$(date +%Y%m%d-%H%M%S)" \
        --from-backup "$backup_name" \
        --include-namespaces "$NAMESPACE" \
        --include-resources configmaps \
        --namespace="$VELERO_NAMESPACE"
    
    # Extract and apply Secrets
    log_info "Restoring Secrets..."
    velero restore create "secrets-restore-$(date +%Y%m%d-%H%M%S)" \
        --from-backup "$backup_name" \
        --include-namespaces "$NAMESPACE" \
        --include-resources secrets \
        --namespace="$VELERO_NAMESPACE"
    
    log_success "Configuration restore completed"
}

# Post-recovery validation
post_recovery_validation() {
    log_section "Performing post-recovery validation..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would perform post-recovery validation"
        return 0
    fi
    
    # Wait for deployments to be ready
    log_info "Waiting for deployments to be ready..."
    kubectl wait --for=condition=available --timeout=600s deployment --all -n "$NAMESPACE"
    
    # Check pod status
    local pods=$(kubectl get pods -n "$NAMESPACE" --no-headers | wc -l)
    local running_pods=$(kubectl get pods -n "$NAMESPACE" --no-headers | grep "Running" | wc -l)
    
    log_info "Pod status: $running_pods/$pods running"
    
    if [ "$running_pods" -ne "$pods" ]; then
        log_warning "Not all pods are running"
        kubectl get pods -n "$NAMESPACE"
    fi
    
    # Check service endpoints
    local services=$(kubectl get services -n "$NAMESPACE" --no-headers | wc -l)
    log_info "Services restored: $services"
    
    # Check persistent volumes
    local pvcs=$(kubectl get pvc -n "$NAMESPACE" --no-headers | wc -l)
    log_info "Persistent volume claims restored: $pvcs"
    
    # Health check endpoints
    log_info "Performing health checks..."
    local deployments=$(kubectl get deployments -n "$NAMESPACE" --no-headers | awk '{print $1}')
    for deployment in $deployments; do
        local service_name=$(kubectl get service -n "$NAMESPACE" -l app="$deployment" --no-headers | awk '{print $1}' | head -1)
        if [ -n "$service_name" ]; then
            local service_port=$(kubectl get service "$service_name" -n "$NAMESPACE" -o jsonpath='{.spec.ports[0].port}')
            if [ -n "$service_port" ]; then
                log_info "Checking health of $service_name:$service_port"
                kubectl run health-check --image=curlimages/curl --rm -i --restart=Never -- \
                    curl -f "http://$service_name:$service_port/health" || log_warning "Health check failed for $service_name"
            fi
        fi
    done
    
    log_success "Post-recovery validation completed"
}

# Generate recovery report
generate_recovery_report() {
    local backup_name="$1"
    local restore_name="$2"
    local damage_level="$3"
    local start_time="$4"
    local end_time="$5"
    
    log_section "Generating recovery report..."
    
    local report_file="/tmp/disaster-recovery-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" <<EOF
CloudBridge Relay Disaster Recovery Report
==========================================

Recovery Date: $(date)
Recovery Mode: $RECOVERY_MODE
Damage Level: $damage_level
Backup Used: $backup_name
Restore Name: $restore_name

Recovery Timeline:
- Start Time: $start_time
- End Time: $end_time
- Duration: $(($(date -d "$end_time" +%s) - $(date -d "$start_time" +%s))) seconds

Recovery Objectives:
- RTO Target: $RTO_TARGET minutes
- RPO Target: $RPO_TARGET minutes

Components Restored:
- Namespace: $NAMESPACE
- Deployments: $(kubectl get deployments -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
- Services: $(kubectl get services -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
- ConfigMaps: $(kubectl get configmaps -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
- Secrets: $(kubectl get secrets -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
- PVCs: $(kubectl get pvc -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)

Post-Recovery Status:
- Running Pods: $(kubectl get pods -n "$NAMESPACE" --no-headers | grep "Running" | wc -l)
- Total Pods: $(kubectl get pods -n "$NAMESPACE" --no-headers | wc -l)

Recovery Actions Taken:
1. Damage assessment completed
2. Latest backup identified: $backup_name
3. Backup validation completed
4. Pre-recovery checks completed
5. Existing services stopped
6. Restore from backup completed
7. Post-recovery validation completed

Recommendations:
- Monitor system health for 24 hours
- Verify all critical services are functioning
- Update monitoring and alerting configurations
- Schedule additional backup verification
- Review and update disaster recovery procedures

EOF
    
    log_success "Recovery report generated: $report_file"
    cat "$report_file"
}

# Main function
main() {
    local start_time=$(date)
    
    log_info "Starting CloudBridge Relay disaster recovery..."
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --backup-name)
                BACKUP_NAME="$2"
                shift 2
                ;;
            --recovery-mode)
                RECOVERY_MODE="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            --force)
                FORCE_RECOVERY="true"
                shift
                ;;
            --namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Confirmation prompt
    if [ "$DRY_RUN" != "true" ] && [ "$FORCE_RECOVERY" != "true" ]; then
        log_warning "This will perform disaster recovery on namespace: $NAMESPACE"
        read -p "Are you sure you want to continue? (yes/no): " -r
        if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
            log_info "Recovery cancelled by user"
            exit 0
        fi
    fi
    
    # Execute recovery process
    check_dependencies
    local damage_level=$(assess_damage)
    local backup_name=$(find_latest_backup)
    validate_backup "$backup_name"
    pre_recovery_checks
    
    local restore_name=""
    case $RECOVERY_MODE in
        full)
            stop_existing_services
            restore_name=$(restore_from_backup "$backup_name")
            ;;
        partial)
            restore_name=$(restore_from_backup "$backup_name")
            ;;
        config-only)
            restore_configurations "$backup_name"
            ;;
        *)
            log_error "Unknown recovery mode: $RECOVERY_MODE"
            exit 1
            ;;
    esac
    
    post_recovery_validation
    
    local end_time=$(date)
    generate_recovery_report "$backup_name" "$restore_name" "$damage_level" "$start_time" "$end_time"
    
    log_success "Disaster recovery completed successfully!"
    log_info ""
    log_info "Recovery Summary:"
    log_info "  - Mode: $RECOVERY_MODE"
    log_info "  - Backup: $backup_name"
    log_info "  - Restore: $restore_name"
    log_info "  - Duration: $(($(date -d "$end_time" +%s) - $(date -d "$start_time" +%s))) seconds"
}

# Run main function
main "$@"

