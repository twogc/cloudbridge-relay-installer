#!/bin/bash

# CloudBridge Relay - Keycloak Setup Script
# Этот скрипт настраивает Keycloak для работы с CloudBridge Relay

set -e
set -x

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
CLIENT_SECRET="${CLIENT_SECRET:-relay-client-secret}"

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
    
    log_success "Dependencies check passed"
}

# Wait for Keycloak to be ready
wait_for_keycloak() {
    log_info "Waiting for Keycloak to be ready..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$KEYCLOAK_URL/realms/master" > /dev/null 2>&1; then
            log_success "Keycloak is ready"
            return 0
        fi
        
        log_info "Attempt $attempt/$max_attempts - Keycloak not ready yet, waiting..."
        sleep 5
        ((attempt++))
    done
    
    log_error "Keycloak failed to start within expected time"
    exit 1
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
    
    log_info "Admin token: $access_token"
    
    if [ "$access_token" = "null" ] || [ -z "$access_token" ]; then
        log_error "Failed to extract access token from response"
        echo "Response: $token_response"
        exit 1
    fi
    
    echo "$access_token"
}

# Create realm
create_realm() {
    local admin_token="$1"
    
    log_info "Creating realm '$REALM_NAME'..."
    
    local realm_data="{\"realm\": \"$REALM_NAME\", \"displayName\": \"CloudBridge\", \"enabled\": true}"
    echo "$realm_data" > realm.json
    log_info "Realm data: $realm_data"
    log_info "Содержимое realm.json: $(cat realm.json)"
    
    log_info "Admin token used for curl: $admin_token"
    
    local response=$(curl -s -w "%{http_code}" -X POST "$KEYCLOAK_URL/admin/realms" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d @realm.json)
    
    local http_code="${response: -3}"
    local response_body="${response%???}"
    
    log_info "HTTP Code: $http_code"
    log_info "Response: $response_body"
    
    rm -f realm.json
    
    if [ "$http_code" = "201" ]; then
        log_success "Realm '$REALM_NAME' created successfully"
    elif [ "$http_code" = "409" ]; then
        log_warning "Realm '$REALM_NAME' already exists"
    else
        log_error "Failed to create realm. HTTP code: $http_code"
        echo "Response: $response_body"
        exit 1
    fi
}

# Create client
create_client() {
    local admin_token="$1"
    
    log_info "Creating client '$CLIENT_ID'..."
    
    local client_data=$(cat <<EOF
{
    "clientId": "$CLIENT_ID",
    "enabled": true,
    "protocol": "openid-connect",
    "publicClient": false,
    "standardFlowEnabled": true,
    "directAccessGrantsEnabled": true,
    "serviceAccountsEnabled": true,
    "authorizationServicesEnabled": false,
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
    },
    "redirectUris": [
        "http://localhost:8081/*",
        "https://localhost:8081/*"
    ],
    "webOrigins": [
        "http://localhost:8081",
        "https://localhost:8081"
    ]
}
EOF
)
    
    local response=$(curl -s -w "%{http_code}" -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d "$client_data")
    
    local http_code="${response: -3}"
    local response_body="${response%???}"
    
    if [ "$http_code" = "201" ]; then
        log_success "Client '$CLIENT_ID' created successfully"
    elif [ "$http_code" = "409" ]; then
        log_warning "Client '$CLIENT_ID' already exists"
    else
        log_error "Failed to create client. HTTP code: $http_code"
        echo "Response: $response_body"
        exit 1
    fi
}

# Get client secret
get_client_secret() {
    local admin_token="$1"
    
    log_info "Getting client secret for '$CLIENT_ID'..."
    
    # Get client ID (UUID)
    local client_response=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients?clientId=$CLIENT_ID" \
        -H "Authorization: Bearer $admin_token")
    
    local client_uuid=$(echo "$client_response" | jq -r '.[0].id')
    
    if [ "$client_uuid" = "null" ] || [ -z "$client_uuid" ]; then
        log_error "Failed to get client UUID"
        exit 1
    fi
    
    # Get client secret
    local secret_response=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients/$client_uuid/client-secret" \
        -H "Authorization: Bearer $admin_token")
    
    local client_secret=$(echo "$secret_response" | jq -r '.value')
    
    if [ "$client_secret" = "null" ] || [ -z "$client_secret" ]; then
        log_error "Failed to get client secret"
        exit 1
    fi
    
    echo "$client_secret"
    log_success "Client secret obtained"
}

# Create roles
create_roles() {
    local admin_token="$1"
    
    log_info "Creating roles..."
    
    local roles=("relay-admin" "relay-user" "relay-readonly")
    
    for role in "${roles[@]}"; do
        log_info "Creating role '$role'..."
        
        local role_data=$(cat <<EOF
{
    "name": "$role",
    "description": "Role for CloudBridge Relay - $role",
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
            log_success "Role '$role' created successfully"
        elif [ "$http_code" = "409" ]; then
            log_warning "Role '$role' already exists"
        else
            log_error "Failed to create role '$role'. HTTP code: $http_code"
        fi
    done
}

# Create groups
create_groups() {
    local admin_token="$1"
    
    log_info "Creating groups..."
    
    local groups=("relay-admins" "relay-users" "relay-monitors")
    
    for group in "${groups[@]}"; do
        log_info "Creating group '$group'..."
        
        local group_data=$(cat <<EOF
{
    "name": "$group",
    "attributes": {
        "description": ["Group for CloudBridge Relay - $group"]
    }
}
EOF
)
        
        local response=$(curl -s -w "%{http_code}" -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/groups" \
            -H "Authorization: Bearer $admin_token" \
            -H "Content-Type: application/json" \
            -d "$group_data")
        
        local http_code="${response: -3}"
        
        if [ "$http_code" = "201" ]; then
            log_success "Group '$group' created successfully"
        elif [ "$http_code" = "409" ]; then
            log_warning "Group '$group' already exists"
        else
            log_error "Failed to create group '$group'. HTTP code: $http_code"
        fi
    done
}

# Create users
create_users() {
    local admin_token="$1"
    
    log_info "Creating users..."
    
    # Create admin user
    local admin_user_data=$(cat <<EOF
{
    "username": "admin",
    "email": "admin@cloudbridge.local",
    "firstName": "Admin",
    "lastName": "User",
    "enabled": true,
    "emailVerified": true,
    "credentials": [
        {
            "type": "password",
            "value": "admin123",
            "temporary": false
        }
    ],
    "groups": ["relay-admins"],
    "realmRoles": ["relay-admin"]
}
EOF
)
    
    local response=$(curl -s -w "%{http_code}" -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d "$admin_user_data")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "201" ]; then
        log_success "Admin user created successfully"
    elif [ "$http_code" = "409" ]; then
        log_warning "Admin user already exists"
    else
        log_error "Failed to create admin user. HTTP code: $http_code"
    fi
    
    # Create regular user
    local user_data=$(cat <<EOF
{
    "username": "user",
    "email": "user@cloudbridge.local",
    "firstName": "Regular",
    "lastName": "User",
    "enabled": true,
    "emailVerified": true,
    "credentials": [
        {
            "type": "password",
            "value": "user123",
            "temporary": false
        }
    ],
    "groups": ["relay-users"],
    "realmRoles": ["relay-user"]
}
EOF
)
    
    response=$(curl -s -w "%{http_code}" -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d "$user_data")
    
    http_code="${response: -3}"
    
    if [ "$http_code" = "201" ]; then
        log_success "Regular user created successfully"
    elif [ "$http_code" = "409" ]; then
        log_warning "Regular user already exists"
    else
        log_error "Failed to create regular user. HTTP code: $http_code"
    fi
}

# Get public key
get_public_key() {
    log_info "Getting public key..."
    
    local jwks_response=$(curl -s "$KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/certs")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to get JWKS"
        exit 1
    fi
    
    # Получаем base64-encoded ключ из x5c (полный сертификат)
    local base64_key=$(echo "$jwks_response" | jq -r '.keys[0].x5c[0]')
    
    if [ "$base64_key" = "null" ] || [ -z "$base64_key" ]; then
        log_error "Failed to extract public key"
        exit 1
    fi
    
    # Формируем полный PEM формат
    local pem_key="-----BEGIN PUBLIC KEY-----"
    pem_key+=$'\n'
    pem_key+="$base64_key"
    pem_key+=$'\n'
    pem_key+="-----END PUBLIC KEY-----"
    
    echo "$pem_key"
    log_success "Public key obtained in PEM format"
}

# Generate configuration
generate_config() {
    local client_secret="$1"
    local public_key="$2"
    
    log_info "Generating configuration..."
    
    cat > keycloak_config.json <<EOF
{
  "keycloak": {
    "enabled": true,
    "server_url": "$KEYCLOAK_URL",
    "realm": "$REALM_NAME",
    "client_id": "$CLIENT_ID",
    "client_secret": "$client_secret",
    "public_key": "$public_key",
    "token_validation": {
      "issuer": "$KEYCLOAK_URL/realms/$REALM_NAME",
      "audience": "$CLIENT_ID",
      "clock_skew": "30s"
    },
    "jwks_cache_ttl": "1h",
    "http_timeout": "30s"
  }
}
EOF
    
    log_success "Configuration saved to keycloak_config.json"
}

# Main function
main() {
    log_info "Starting Keycloak setup for CloudBridge Relay..."
    
    check_dependencies
    wait_for_keycloak
    
    local admin_token=$(get_admin_token)
    
    create_realm "$admin_token"
    create_client "$admin_token"
    local client_secret=$(get_client_secret "$admin_token")
    create_roles "$admin_token"
    create_groups "$admin_token"
    create_users "$admin_token"
    local public_key=$(get_public_key)
    
    generate_config "$client_secret" "$public_key"
    
    log_success "Keycloak setup completed successfully!"
    log_info ""
    log_info "Configuration summary:"
    log_info "  Keycloak URL: $KEYCLOAK_URL"
    log_info "  Realm: $REALM_NAME"
    log_info "  Client ID: $CLIENT_ID"
    log_info "  Client Secret: $client_secret"
    log_info ""
    log_info "Users created:"
    log_info "  admin/admin123 (relay-admin role)"
    log_info "  user/user123 (relay-user role)"
    log_info ""
    log_info "Configuration file: keycloak_config.json"
    log_info ""
    log_info "Next steps:"
    log_info "1. Copy keycloak_config.json to your relay config directory"
    log_info "2. Update your relay configuration to enable Keycloak"
    log_info "3. Restart the relay service"
}

# Run main function
main "$@"

# Напоминание о сертификатах для HTTPS Keycloak
if [[ ! -f "certs/fullchain.pem" || ! -f "certs/privkey.pem" ]]; then
    echo -e "${YELLOW}[WARNING] Для HTTPS Keycloak нужны certs/fullchain.pem и certs/privkey.pem!${NC}"
    echo -e "${YELLOW}[WARNING] Без них Keycloak не сможет работать по 443 (HTTPS).${NC}"
else
    echo -e "${GREEN}[INFO] SSL-сертификаты для Keycloak найдены. HTTPS будет работать.${NC}"
fi

# Напоминание о KC_HOSTNAME
if [[ -z "$KC_HOSTNAME" || "$KC_HOSTNAME" == "localhost" ]]; then
    echo -e "${YELLOW}[WARNING] Переменная KC_HOSTNAME не задана или равна 'localhost'."
    echo -e "${YELLOW}[WARNING] Для корректных редиректов укажите внешний домен, например: edge.2gc.ru${NC}"
else
    echo -e "${GREEN}[INFO] KC_HOSTNAME: $KC_HOSTNAME${NC}"
fi

    # Интегрируем существующий interactive_setup.sh
    if [ -f "scripts/interactive_setup.sh" ]; then
        echo -e "\033[1;34m[INFO]\033[0m Запуск интерактивной настройки Keycloak..."
        bash scripts/interactive_setup.sh
    else
        echo -e "\033[1;34m[INFO]\033[0m Интерактивная настройка Keycloak (заглушка)"
        # TODO: интегрировать scripts/interactive_setup.sh
    fi 