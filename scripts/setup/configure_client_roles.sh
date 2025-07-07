#!/bin/bash

# CloudBridge Relay - Configure Client Roles Script
# Этот скрипт настраивает клиент Keycloak для включения ролей в токен

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
CLIENT_ID="${CLIENT_ID:-relay-client}"

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

# Configure client to include roles in token
configure_client_roles() {
    local admin_token="$1"
    
    log_info "Configuring client to include roles in token..."
    
    # Get client ID (UUID)
    local client_response=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients?clientId=$CLIENT_ID" \
        -H "Authorization: Bearer $admin_token")
    
    local client_uuid=$(echo "$client_response" | jq -r '.[0].id')
    
    if [ "$client_uuid" = "null" ] || [ -z "$client_uuid" ]; then
        log_error "Failed to get client UUID"
        exit 1
    fi
    
    log_info "Client UUID: $client_uuid"
    
    # Get current client configuration
    local current_config=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients/$client_uuid" \
        -H "Authorization: Bearer $admin_token")
    
    log_info "Current client configuration retrieved"
    
    # Update client configuration to include roles in token
    local client_update_data=$(cat <<EOF
{
    "protocolMappers": [
        {
            "name": "realm roles",
            "protocol": "openid-connect",
            "protocolMapper": "oidc-usermodel-realm-role-mapper",
            "config": {
                "multivalued": "true",
                "userinfo.token.claim": "true",
                "id.token.claim": "true",
                "access.token.claim": "true",
                "claim.name": "realm_access.roles",
                "jsonType.label": "String"
            }
        },
        {
            "name": "client roles",
            "protocol": "openid-connect",
            "protocolMapper": "oidc-usermodel-client-role-mapper",
            "config": {
                "multivalued": "true",
                "userinfo.token.claim": "true",
                "id.token.claim": "true",
                "access.token.claim": "true",
                "claim.name": "resource_access.${CLIENT_ID}.roles",
                "jsonType.label": "String"
            }
        }
    ]
}
EOF
)
    
    # Update client
    local response=$(curl -s -w "%{http_code}" -X PUT "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients/$client_uuid" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d "$client_update_data")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "204" ]; then
        log_success "Client configured to include roles in token"
    else
        log_error "Failed to configure client. HTTP code: $http_code"
        echo "Response: ${response%???}"
        exit 1
    fi
}

# Enable client roles
enable_client_roles() {
    local admin_token="$1"
    
    log_info "Enabling client roles..."
    
    # Get client ID (UUID)
    local client_response=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients?clientId=$CLIENT_ID" \
        -H "Authorization: Bearer $admin_token")
    
    local client_uuid=$(echo "$client_response" | jq -r '.[0].id')
    
    # Enable client roles
    local client_update_data=$(cat <<EOF
{
    "clientId": "$CLIENT_ID",
    "enabled": true,
    "publicClient": true,
    "standardFlowEnabled": true,
    "directAccessGrantsEnabled": true,
    "serviceAccountsEnabled": false,
    "authorizationServicesEnabled": false,
    "implicitFlowEnabled": false,
    "fullScopeAllowed": true,
    "attributes": {
        "saml.assertion.signature": "false",
        "saml.force.post.binding": "false",
        "saml.multivalued.roles": "false",
        "saml.encrypt": "false",
        "saml.server.signature": "false",
        "saml.server.signature.keyinfo.ext": "false",
        "exclude.session.state.from.auth.response": "false",
        "saml_force_name_id_format": "false",
        "saml.client.signature": "false",
        "tls.client.certificate.bound.access.tokens": "false",
        "saml.authnstatement": "false",
        "display.on.consent.screen": "false",
        "saml.onetimeuse.condition": "false"
    }
}
EOF
)
    
    local response=$(curl -s -w "%{http_code}" -X PUT "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients/$client_uuid" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d "$client_update_data")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "204" ]; then
        log_success "Client roles enabled"
    else
        log_error "Failed to enable client roles. HTTP code: $http_code"
        echo "Response: ${response%???}"
        exit 1
    fi
}

# Main function
main() {
    log_info "Configuring client roles for CloudBridge Relay..."
    
    # Check dependencies
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed"
        exit 1
    fi
    
    # Get admin token
    local admin_token=$(get_admin_token)
    
    # Configure client
    configure_client_roles "$admin_token"
    enable_client_roles "$admin_token"
    
    log_success "Client roles configuration completed successfully!"
    log_info ""
    log_info "Configuration summary:"
    log_info "  Client '$CLIENT_ID' configured to include roles in token"
    log_info "  Realm roles will be included in 'realm_access.roles'"
    log_info "  Client roles will be included in 'resource_access.${CLIENT_ID}.roles'"
    log_info ""
    log_info "You may need to re-authenticate to get updated tokens with roles"
}

# Run main function
main "$@" 