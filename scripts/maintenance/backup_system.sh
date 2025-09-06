#!/bin/bash

# CloudBridge Relay - Comprehensive System Backup Script
# Этот скрипт выполняет полное резервное копирование всех компонентов системы

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/opt/cloudbridge/backups}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
COMPRESS_BACKUP="${COMPRESS_BACKUP:-true}"
ENCRYPT_BACKUP="${ENCRYPT_BACKUP:-false}"
ENCRYPTION_PASSWORD="${ENCRYPTION_PASSWORD:-}"
BACKUP_S3_BUCKET="${BACKUP_S3_BUCKET:-}"
BACKUP_GCS_BUCKET="${BACKUP_GCS_BUCKET:-}"

# Kubernetes configuration
KUBECONFIG="${KUBECONFIG:-~/.kube/config}"
NAMESPACE="${NAMESPACE:-cloudbridge}"

# Database configuration
CASSANDRA_HOST="${CASSANDRA_HOST:-cassandra}"
CASSANDRA_PORT="${CASSANDRA_PORT:-9042}"
CASSANDRA_KEYSPACE="${CASSANDRA_KEYSPACE:-cloudbridge}"

# Prometheus configuration
PROMETHEUS_HOST="${PROMETHEUS_HOST:-prometheus}"
PROMETHEUS_PORT="${PROMETHEUS_PORT:-9090}"

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
    
    # Required tools
    for tool in kubectl curl jq tar gzip; do
        if ! command -v "$tool" &> /dev/null; then
            missing_deps+=("$tool")
        fi
    done
    
    # Optional tools
    if [ "$COMPRESS_BACKUP" = "true" ] && ! command -v gzip &> /dev/null; then
        missing_deps+=("gzip")
    fi
    
    if [ "$ENCRYPT_BACKUP" = "true" ] && ! command -v openssl &> /dev/null; then
        missing_deps+=("openssl")
    fi
    
    if [ -n "$BACKUP_S3_BUCKET" ] && ! command -v aws &> /dev/null; then
        missing_deps+=("aws")
    fi
    
    if [ -n "$BACKUP_GCS_BUCKET" ] && ! command -v gsutil &> /dev/null; then
        missing_deps+=("gsutil")
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        exit 1
    fi
    
    log_success "Dependencies check passed"
}

# Create backup directory structure
create_backup_dirs() {
    log_info "Creating backup directory structure..."
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_path="$BACKUP_DIR/$timestamp"
    
    mkdir -p "$backup_path"/{kubernetes,cassandra,prometheus,configs,logs,velero}
    
    echo "$backup_path"
}

# Backup Kubernetes resources
backup_kubernetes() {
    local backup_path="$1"
    local k8s_backup_dir="$backup_path/kubernetes"
    
    log_section "Backing up Kubernetes resources..."
    
    # Backup all resources in namespace
    log_info "Backing up namespace resources..."
    kubectl get all -n "$NAMESPACE" -o yaml > "$k8s_backup_dir/all-resources.yaml"
    
    # Backup ConfigMaps
    log_info "Backing up ConfigMaps..."
    kubectl get configmaps -n "$NAMESPACE" -o yaml > "$k8s_backup_dir/configmaps.yaml"
    
    # Backup Secrets
    log_info "Backing up Secrets..."
    kubectl get secrets -n "$NAMESPACE" -o yaml > "$k8s_backup_dir/secrets.yaml"
    
    # Backup PersistentVolumes
    log_info "Backing up PersistentVolumes..."
    kubectl get pv -o yaml > "$k8s_backup_dir/persistent-volumes.yaml"
    
    # Backup PersistentVolumeClaims
    log_info "Backing up PersistentVolumeClaims..."
    kubectl get pvc -n "$NAMESPACE" -o yaml > "$k8s_backup_dir/persistent-volume-claims.yaml"
    
    # Backup Services
    log_info "Backing up Services..."
    kubectl get services -n "$NAMESPACE" -o yaml > "$k8s_backup_dir/services.yaml"
    
    # Backup Ingress
    log_info "Backing up Ingress..."
    kubectl get ingress -n "$NAMESPACE" -o yaml > "$k8s_backup_dir/ingress.yaml"
    
    # Backup NetworkPolicies
    log_info "Backing up NetworkPolicies..."
    kubectl get networkpolicies -n "$NAMESPACE" -o yaml > "$k8s_backup_dir/network-policies.yaml"
    
    # Backup RBAC
    log_info "Backing up RBAC resources..."
    kubectl get roles,rolebindings,clusterroles,clusterrolebindings -n "$NAMESPACE" -o yaml > "$k8s_backup_dir/rbac.yaml"
    
    log_success "Kubernetes backup completed"
}

# Backup Cassandra data
backup_cassandra() {
    local backup_path="$1"
    local cassandra_backup_dir="$backup_path/cassandra"
    
    log_section "Backing up Cassandra data..."
    
    # Get Cassandra pod name
    local cassandra_pod=$(kubectl get pods -n "$NAMESPACE" -l app=cassandra -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    
    if [ -z "$cassandra_pod" ]; then
        log_warning "Cassandra pod not found, skipping Cassandra backup"
        return 0
    fi
    
    log_info "Found Cassandra pod: $cassandra_pod"
    
    # Create snapshot
    log_info "Creating Cassandra snapshot..."
    kubectl exec -n "$NAMESPACE" "$cassandra_pod" -- nodetool snapshot "$CASSANDRA_KEYSPACE"
    
    # Get snapshot directory
    local snapshot_dir=$(kubectl exec -n "$NAMESPACE" "$cassandra_pod" -- find /var/lib/cassandra/data -name "snapshots" -type d | head -1)
    
    if [ -n "$snapshot_dir" ]; then
        log_info "Copying snapshot data..."
        kubectl cp "$NAMESPACE/$cassandra_pod:$snapshot_dir" "$cassandra_backup_dir/snapshots"
    fi
    
    # Backup schema
    log_info "Backing up Cassandra schema..."
    kubectl exec -n "$NAMESPACE" "$cassandra_pod" -- cqlsh -e "DESCRIBE KEYSPACE $CASSANDRA_KEYSPACE" > "$cassandra_backup_dir/schema.cql"
    
    log_success "Cassandra backup completed"
}

# Backup Prometheus data
backup_prometheus() {
    local backup_path="$1"
    local prometheus_backup_dir="$backup_path/prometheus"
    
    log_section "Backing up Prometheus data..."
    
    # Get Prometheus pod name
    local prometheus_pod=$(kubectl get pods -n "$NAMESPACE" -l app=prometheus -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    
    if [ -z "$prometheus_pod" ]; then
        log_warning "Prometheus pod not found, skipping Prometheus backup"
        return 0
    fi
    
    log_info "Found Prometheus pod: $prometheus_pod"
    
    # Backup Prometheus data directory
    log_info "Backing up Prometheus data directory..."
    kubectl exec -n "$NAMESPACE" "$prometheus_pod" -- tar -czf /tmp/prometheus-data.tar.gz -C /prometheus .
    kubectl cp "$NAMESPACE/$prometheus_pod:/tmp/prometheus-data.tar.gz" "$prometheus_backup_dir/prometheus-data.tar.gz"
    
    # Backup Prometheus configuration
    log_info "Backing up Prometheus configuration..."
    kubectl exec -n "$NAMESPACE" "$prometheus_pod" -- cat /etc/prometheus/prometheus.yml > "$prometheus_backup_dir/prometheus.yml"
    
    log_success "Prometheus backup completed"
}

# Backup configuration files
backup_configs() {
    local backup_path="$1"
    local configs_backup_dir="$backup_path/configs"
    
    log_section "Backing up configuration files..."
    
    # Backup local configuration files
    if [ -d "./configs" ]; then
        log_info "Backing up local configs directory..."
        cp -r ./configs "$configs_backup_dir/local"
    fi
    
    # Backup Terraform state
    if [ -d "./terraform" ]; then
        log_info "Backing up Terraform state..."
        cp -r ./terraform "$configs_backup_dir/terraform"
    fi
    
    # Backup Docker Compose files
    if [ -f "./docker-compose.yml" ]; then
        log_info "Backing up Docker Compose files..."
        cp docker-compose*.yml "$configs_backup_dir/" 2>/dev/null || true
    fi
    
    # Backup Kubernetes manifests
    if [ -d "./k8s" ]; then
        log_info "Backing up Kubernetes manifests..."
        cp -r ./k8s "$configs_backup_dir/k8s"
    fi
    
    log_success "Configuration backup completed"
}

# Backup logs
backup_logs() {
    local backup_path="$1"
    local logs_backup_dir="$backup_path/logs"
    
    log_section "Backing up system logs..."
    
    # Backup application logs
    if [ -d "./logs" ]; then
        log_info "Backing up application logs..."
        cp -r ./logs "$logs_backup_dir/application"
    fi
    
    # Backup system logs (if accessible)
    if [ -d "/var/log" ]; then
        log_info "Backing up system logs..."
        sudo cp -r /var/log "$logs_backup_dir/system" 2>/dev/null || log_warning "Could not backup system logs (permission denied)"
    fi
    
    # Backup Kubernetes logs
    log_info "Backing up Kubernetes logs..."
    kubectl logs -n "$NAMESPACE" --all-containers=true --previous=false > "$logs_backup_dir/kubernetes-logs.txt" 2>/dev/null || true
    
    log_success "Logs backup completed"
}

# Create Velero backup
create_velero_backup() {
    local backup_path="$1"
    local velero_backup_dir="$backup_path/velero"
    
    log_section "Creating Velero backup..."
    
    # Check if Velero is installed
    if ! kubectl get crd backupstoragelocations.velero.io &>/dev/null; then
        log_warning "Velero is not installed, skipping Velero backup"
        return 0
    fi
    
    local backup_name="cloudbridge-backup-$(date +%Y%m%d-%H%M%S)"
    
    log_info "Creating Velero backup: $backup_name"
    velero backup create "$backup_name" --include-namespaces "$NAMESPACE" --wait
    
    # Export backup information
    velero backup describe "$backup_name" --details > "$velero_backup_dir/backup-info.txt"
    
    log_success "Velero backup completed: $backup_name"
}

# Compress backup
compress_backup() {
    local backup_path="$1"
    
    if [ "$COMPRESS_BACKUP" = "true" ]; then
        log_info "Compressing backup..."
        tar -czf "${backup_path}.tar.gz" -C "$(dirname "$backup_path")" "$(basename "$backup_path")"
        rm -rf "$backup_path"
        log_success "Backup compressed: ${backup_path}.tar.gz"
        echo "${backup_path}.tar.gz"
    else
        echo "$backup_path"
    fi
}

# Encrypt backup
encrypt_backup() {
    local backup_file="$1"
    
    if [ "$ENCRYPT_BACKUP" = "true" ]; then
        if [ -z "$ENCRYPTION_PASSWORD" ]; then
            log_error "Encryption password is required when ENCRYPT_BACKUP=true"
            exit 1
        fi
        
        log_info "Encrypting backup..."
        openssl enc -aes-256-cbc -salt -in "$backup_file" -out "${backup_file}.enc" -pass pass:"$ENCRYPTION_PASSWORD"
        rm "$backup_file"
        log_success "Backup encrypted: ${backup_file}.enc"
        echo "${backup_file}.enc"
    else
        echo "$backup_file"
    fi
}

# Upload to cloud storage
upload_to_cloud() {
    local backup_file="$1"
    
    if [ -n "$BACKUP_S3_BUCKET" ]; then
        log_info "Uploading to S3: $BACKUP_S3_BUCKET"
        aws s3 cp "$backup_file" "s3://$BACKUP_S3_BUCKET/$(basename "$backup_file")"
        log_success "Uploaded to S3"
    fi
    
    if [ -n "$BACKUP_GCS_BUCKET" ]; then
        log_info "Uploading to GCS: $BACKUP_GCS_BUCKET"
        gsutil cp "$backup_file" "gs://$BACKUP_GCS_BUCKET/$(basename "$backup_file")"
        log_success "Uploaded to GCS"
    fi
}

# Cleanup old backups
cleanup_old_backups() {
    log_info "Cleaning up old backups (older than $BACKUP_RETENTION_DAYS days)..."
    
    local deleted_count=0
    
    find "$BACKUP_DIR" -maxdepth 1 -type d -name "20*" -mtime +$BACKUP_RETENTION_DAYS | while read -r dir; do
        rm -rf "$dir"
        log_info "Deleted old backup: $dir"
        ((deleted_count++))
    done
    
    find "$BACKUP_DIR" -maxdepth 1 -type f -name "*.tar.gz*" -mtime +$BACKUP_RETENTION_DAYS | while read -r file; do
        rm "$file"
        log_info "Deleted old backup: $file"
        ((deleted_count++))
    done
    
    if [ $deleted_count -gt 0 ]; then
        log_success "Cleaned up $deleted_count old backup files"
    else
        log_info "No old backup files to clean up"
    fi
}

# Create backup summary
create_backup_summary() {
    local backup_path="$1"
    local backup_file="$2"
    
    local summary_file="$BACKUP_DIR/backup_summary_$(date +%Y%m%d_%H%M%S).txt"
    
    cat > "$summary_file" <<EOF
CloudBridge Relay System Backup Summary
========================================

Backup Date: $(date)
Backup Path: $backup_path
Backup File: $backup_file

Components Backed Up:
- Kubernetes resources (deployments, services, configmaps, secrets)
- Cassandra data and schema
- Prometheus data and configuration
- Configuration files (configs, terraform, k8s manifests)
- System and application logs
- Velero backup (if available)

Backup Configuration:
- Compression: $COMPRESS_BACKUP
- Encryption: $ENCRYPT_BACKUP
- Retention: $BACKUP_RETENTION_DAYS days
- S3 Upload: ${BACKUP_S3_BUCKET:-"Not configured"}
- GCS Upload: ${BACKUP_GCS_BUCKET:-"Not configured"}

Backup Size: $(du -h "$backup_file" | cut -f1)

EOF
    
    log_success "Backup summary created: $summary_file"
}

# Main function
main() {
    log_info "Starting CloudBridge Relay system backup..."
    
    check_dependencies
    
    local backup_path=$(create_backup_dirs)
    local timestamp=$(date +%Y%m%d_%H%M%S)
    
    # Perform backups
    backup_kubernetes "$backup_path"
    backup_cassandra "$backup_path"
    backup_prometheus "$backup_path"
    backup_configs "$backup_path"
    backup_logs "$backup_path"
    create_velero_backup "$backup_path"
    
    # Process backup
    local backup_file=$(compress_backup "$backup_path")
    backup_file=$(encrypt_backup "$backup_file")
    
    # Upload to cloud storage
    upload_to_cloud "$backup_file"
    
    # Cleanup old backups
    cleanup_old_backups
    
    # Create summary
    create_backup_summary "$backup_path" "$backup_file"
    
    log_success "System backup completed successfully!"
    log_info ""
    log_info "Backup file: $backup_file"
    log_info "Backup size: $(du -h "$backup_file" | cut -f1)"
    log_info "Backup directory: $BACKUP_DIR"
}

# Run main function
main "$@"

