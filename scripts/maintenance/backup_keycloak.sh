#!/bin/bash

# CloudBridge Relay - Keycloak Backup Script
# Этот скрипт выполняет резервное копирование Keycloak

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
KEYCLOAK_URL="${KEYCLOAK_URL:-http://localhost:8080}"
KEYCLOAK_ADMIN="${KEYCLOAK_ADMIN:-admin}"
KEYCLOAK_ADMIN_PASSWORD="${KEYCLOAK_ADMIN_PASSWORD:-admin123}"
REALM_NAME="${REALM_NAME:-cloudbridge}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
COMPRESS_BACKUP="${COMPRESS_BACKUP:-true}"
ENCRYPT_BACKUP="${ENCRYPT_BACKUP:-false}"
ENCRYPTION_PASSWORD="${ENCRYPTION_PASSWORD:-}"

# Functions
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

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed"
        exit 1
    fi
    
    if [ "$COMPRESS_BACKUP" = "true" ] && ! command -v gzip &> /dev/null; then
        log_error "gzip is required for compression but not installed"
        exit 1
    fi
    
    if [ "$ENCRYPT_BACKUP" = "true" ] && ! command -v openssl &> /dev/null; then
        log_error "openssl is required for encryption but not installed"
        exit 1
    fi
    
    log_success "Dependencies check passed"
}

# Create backup directory
create_backup_dir() {
    log_info "Creating backup directory..."
    
    if [ ! -d "$BACKUP_DIR" ]; then
        mkdir -p "$BACKUP_DIR"
        log_success "Backup directory created: $BACKUP_DIR"
    else
        log_info "Backup directory exists: $BACKUP_DIR"
    fi
}

# Get admin token
get_admin_token() {
    log_info "Getting admin token..."
    
    local token_response=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=password" \
        -d "client_id=admin-cli" \
        -d "username=$KEYCLOAK_ADMIN" \
        -d "password=$KEYCLOAK_ADMIN_PASSWORD")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to get admin token"
        exit 1
    fi
    
    local access_token=$(echo "$token_response" | jq -r '.access_token')
    
    if [ "$access_token" = "null" ] || [ -z "$access_token" ]; then
        log_error "Failed to extract access token from response"
        echo "Response: $token_response"
        exit 1
    fi
    
    echo "$access_token"
    log_success "Admin token obtained"
}

# Export realm
export_realm() {
    local admin_token="$1"
    local timestamp="$2"
    
    log_info "Exporting realm '$REALM_NAME'..."
    
    local realm_file="$BACKUP_DIR/realm_${REALM_NAME}_${timestamp}.json"
    
    local response=$(curl -s -w "%{http_code}" -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -o "$realm_file")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Realm exported to: $realm_file"
        echo "$realm_file"
    else
        log_error "Failed to export realm. HTTP code: $http_code"
        return 1
    fi
}

# Export users
export_users() {
    local admin_token="$1"
    local timestamp="$2"
    
    log_info "Exporting users from realm '$REALM_NAME'..."
    
    local users_file="$BACKUP_DIR/users_${REALM_NAME}_${timestamp}.json"
    
    local response=$(curl -s -w "%{http_code}" -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -o "$users_file")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Users exported to: $users_file"
        echo "$users_file"
    else
        log_error "Failed to export users. HTTP code: $http_code"
        return 1
    fi
}

# Export clients
export_clients() {
    local admin_token="$1"
    local timestamp="$2"
    
    log_info "Exporting clients from realm '$REALM_NAME'..."
    
    local clients_file="$BACKUP_DIR/clients_${REALM_NAME}_${timestamp}.json"
    
    local response=$(curl -s -w "%{http_code}" -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -o "$clients_file")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Clients exported to: $clients_file"
        echo "$clients_file"
    else
        log_error "Failed to export clients. HTTP code: $http_code"
        return 1
    fi
}

# Export roles
export_roles() {
    local admin_token="$1"
    local timestamp="$2"
    
    log_info "Exporting roles from realm '$REALM_NAME'..."
    
    local roles_file="$BACKUP_DIR/roles_${REALM_NAME}_${timestamp}.json"
    
    local response=$(curl -s -w "%{http_code}" -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/roles" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -o "$roles_file")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Roles exported to: $roles_file"
        echo "$roles_file"
    else
        log_error "Failed to export roles. HTTP code: $http_code"
        return 1
    fi
}

# Export groups
export_groups() {
    local admin_token="$1"
    local timestamp="$2"
    
    log_info "Exporting groups from realm '$REALM_NAME'..."
    
    local groups_file="$BACKUP_DIR/groups_${REALM_NAME}_${timestamp}.json"
    
    local response=$(curl -s -w "%{http_code}" -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/groups" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -o "$groups_file")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Groups exported to: $groups_file"
        echo "$groups_file"
    else
        log_error "Failed to export groups. HTTP code: $http_code"
        return 1
    fi
}

# Compress backup
compress_backup() {
    local file="$1"
    
    if [ "$COMPRESS_BACKUP" = "true" ]; then
        log_info "Compressing backup file: $file"
        gzip "$file"
        log_success "Backup compressed: ${file}.gz"
        echo "${file}.gz"
    else
        echo "$file"
    fi
}

# Encrypt backup
encrypt_backup() {
    local file="$1"
    
    if [ "$ENCRYPT_BACKUP" = "true" ]; then
        if [ -z "$ENCRYPTION_PASSWORD" ]; then
            log_error "Encryption password is required when ENCRYPT_BACKUP=true"
            exit 1
        fi
        
        log_info "Encrypting backup file: $file"
        openssl enc -aes-256-cbc -salt -in "$file" -out "${file}.enc" -pass pass:"$ENCRYPTION_PASSWORD"
        rm "$file"
        log_success "Backup encrypted: ${file}.enc"
        echo "${file}.enc"
    else
        echo "$file"
    fi
}

# Cleanup old backups
cleanup_old_backups() {
    log_info "Cleaning up old backups (older than $BACKUP_RETENTION_DAYS days)..."
    
    local deleted_count=0
    
    find "$BACKUP_DIR" -name "*.json*" -type f -mtime +$BACKUP_RETENTION_DAYS | while read -r file; do
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
    local timestamp="$1"
    local files=("$@")
    
    local summary_file="$BACKUP_DIR/backup_summary_${timestamp}.txt"
    
    cat > "$summary_file" <<EOF
Keycloak Backup Summary
=======================

Backup Date: $(date)
Realm: $REALM_NAME
Keycloak URL: $KEYCLOAK_URL

Backup Files:
EOF
    
    for file in "${files[@]}"; do
        if [ -f "$file" ]; then
            local size=$(du -h "$file" | cut -f1)
            echo "  - $(basename "$file") ($size)" >> "$summary_file"
        fi
    done
    
    cat >> "$summary_file" <<EOF

Backup Configuration:
  - Compression: $COMPRESS_BACKUP
  - Encryption: $ENCRYPT_BACKUP
  - Retention: $BACKUP_RETENTION_DAYS days

EOF
    
    log_success "Backup summary created: $summary_file"
}

# Main function
main() {
    log_info "Starting Keycloak backup..."
    
    check_dependencies
    create_backup_dir
    
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local admin_token=$(get_admin_token)
    
    local backup_files=()
    
    # Export realm
    local realm_file=$(export_realm "$admin_token" "$timestamp")
    if [ $? -eq 0 ]; then
        backup_files+=("$realm_file")
    fi
    
    # Export users
    local users_file=$(export_users "$admin_token" "$timestamp")
    if [ $? -eq 0 ]; then
        backup_files+=("$users_file")
    fi
    
    # Export clients
    local clients_file=$(export_clients "$admin_token" "$timestamp")
    if [ $? -eq 0 ]; then
        backup_files+=("$clients_file")
    fi
    
    # Export roles
    local roles_file=$(export_roles "$admin_token" "$timestamp")
    if [ $? -eq 0 ]; then
        backup_files+=("$roles_file")
    fi
    
    # Export groups
    local groups_file=$(export_groups "$admin_token" "$timestamp")
    if [ $? -eq 0 ]; then
        backup_files+=("$groups_file")
    fi
    
    # Process backup files
    local processed_files=()
    for file in "${backup_files[@]}"; do
        local processed_file=$(compress_backup "$file")
        processed_file=$(encrypt_backup "$processed_file")
        processed_files+=("$processed_file")
    done
    
    # Cleanup old backups
    cleanup_old_backups
    
    # Create backup summary
    create_backup_summary "$timestamp" "${processed_files[@]}"
    
    log_success "Keycloak backup completed successfully!"
    log_info ""
    log_info "Backup files:"
    for file in "${processed_files[@]}"; do
        log_info "  - $file"
    done
    log_info ""
    log_info "Backup directory: $BACKUP_DIR"
}

# Run main function
main "$@" 