#!/bin/bash

# CloudBridge Relay - Add Admin UI Role Script
# Этот скрипт добавляет роль admin-ui для UI приложения

set -e

# Load environment variables from .env file if it exists
if [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
fi

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

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" >&2
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" >&2
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
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
}

# Create admin-ui role
create_admin_ui_role() {
    local admin_token="$1"
    
    log_info "Creating role 'admin-ui'..."
    
    local role_data=$(cat <<EOF
{
    "name": "admin-ui",
    "description": "Role for CloudBridge Admin UI access",
    "composite": false,
    "clientRole": false
}
EOF
)
    
    local response=$(curl -s -w "%{http_code}" -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/roles" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d "$role_data")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "201" ]; then
        log_success "Role 'admin-ui' created successfully"
    elif [ "$http_code" = "409" ]; then
        log_warning "Role 'admin-ui' already exists"
    else
        log_error "Failed to create role 'admin-ui'. HTTP code: $http_code"
        exit 1
    fi
}

# Assign admin-ui role to admin user
assign_admin_ui_role() {
    local admin_token="$1"
    
    log_info "Assigning admin-ui role to admin user..."
    
    # Get user ID for admin user
    local user_response=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users?username=admin" \
        -H "Authorization: Bearer $admin_token")
    
    local user_id=$(echo "$user_response" | jq -r '.[0].id')
    
    if [ "$user_id" = "null" ] || [ -z "$user_id" ]; then
        log_error "Failed to get admin user ID"
        exit 1
    fi
    
    # Get role ID for admin-ui role
    local role_response=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/roles/admin-ui" \
        -H "Authorization: Bearer $admin_token")
    
    local role_id=$(echo "$role_response" | jq -r '.id')
    
    if [ "$role_id" = "null" ] || [ -z "$role_id" ]; then
        log_error "Failed to get admin-ui role ID"
        exit 1
    fi
    
    # Assign role to user
    local response=$(curl -s -w "%{http_code}" -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users/$user_id/role-mappings/realm" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d "[{\"id\":\"$role_id\",\"name\":\"admin-ui\"}]")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "204" ]; then
        log_success "Admin-ui role assigned to admin user successfully"
    elif [ "$http_code" = "409" ]; then
        log_warning "Admin-ui role already assigned to admin user"
    else
        log_error "Failed to assign admin-ui role to admin user. HTTP code: $http_code"
        exit 1
    fi
}

# Update client redirect URIs for UI
update_client_redirects() {
    local admin_token="$1"
    
    log_info "Updating client redirect URIs for UI..."
    
    # Get client ID (UUID)
    local client_response=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients?clientId=relay-client" \
        -H "Authorization: Bearer $admin_token")
    
    local client_uuid=$(echo "$client_response" | jq -r '.[0].id')
    
    if [ "$client_uuid" = "null" ] || [ -z "$client_uuid" ]; then
        log_error "Failed to get client UUID"
        exit 1
    fi
    
    # Update client with UI redirect URIs
    local client_update_data=$(cat <<EOF
{
    "redirectUris": [
        "http://localhost:8081/*",
        "https://localhost:8081/*",
        "http://localhost:5173/*",
        "https://localhost:5173/*"
    ],
    "webOrigins": [
        "http://localhost:8081",
        "https://localhost:8081",
        "http://localhost:5173",
        "https://localhost:5173"
    ]
}
EOF
)
    
    local response=$(curl -s -w "%{http_code}" -X PUT "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients/$client_uuid" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d "$client_update_data")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "204" ]; then
        log_success "Client redirect URIs updated successfully"
    else
        log_error "Failed to update client redirect URIs. HTTP code: $http_code"
        exit 1
    fi
}

# Main function
main() {
    log_info "Adding admin-ui role for CloudBridge Relay..."
    
    local admin_token=$(get_admin_token)
    
    create_admin_ui_role "$admin_token"
    assign_admin_ui_role "$admin_token"
    update_client_redirects "$admin_token"
    
    log_success "Admin UI role setup completed successfully!"
    log_info ""
    log_info "Configuration summary:"
    log_info "  Role 'admin-ui' created and assigned to admin user"
    log_info "  Client redirect URIs updated for UI (localhost:5173)"
    log_info ""
    log_info "You can now use admin/admin123 to access the UI with admin-ui role"
}

# Run main function
main "$@" 