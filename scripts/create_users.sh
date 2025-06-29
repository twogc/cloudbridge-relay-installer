#!/bin/bash

# CloudBridge Relay - Keycloak User Creation Script
# Этот скрипт создает пользователей в Keycloak

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
    
    log_success "Dependencies check passed"
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

# Create user
create_user() {
    local admin_token="$1"
    local username="$2"
    local email="$3"
    local first_name="$4"
    local last_name="$5"
    local password="$6"
    local groups="$7"
    local roles="$8"
    local attributes="$9"
    
    log_info "Creating user '$username'..."
    
    local user_data=$(cat <<EOF
{
    "username": "$username",
    "email": "$email",
    "firstName": "$first_name",
    "lastName": "$last_name",
    "enabled": true,
    "emailVerified": true,
    "credentials": [
        {
            "type": "password",
            "value": "$password",
            "temporary": false
        }
    ],
    "groups": [$groups],
    "realmRoles": [$roles],
    "attributes": $attributes
}
EOF
)
    
    local response=$(curl -s -w "%{http_code}" -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json" \
        -d "$user_data")
    
    local http_code="${response: -3}"
    local response_body="${response%???}"
    
    if [ "$http_code" = "201" ]; then
        log_success "User '$username' created successfully"
        return 0
    elif [ "$http_code" = "409" ]; then
        log_warning "User '$username' already exists"
        return 0
    else
        log_error "Failed to create user '$username'. HTTP code: $http_code"
        echo "Response: $response_body"
        return 1
    fi
}

# Create admin user
create_admin_user() {
    local admin_token="$1"
    
    local groups='"relay-admins"'
    local roles='"relay-admin"'
    local attributes='{
        "relay.max_tunnels": ["100"],
        "relay.rate_limit": ["1000"],
        "relay.allowed_ips": ["*"]
    }'
    
    create_user "$admin_token" "admin" "admin@cloudbridge.local" "Admin" "User" "SecurePassword123!" "$groups" "$roles" "$attributes"
}

# Create regular user
create_regular_user() {
    local admin_token="$1"
    
    local groups='"relay-users"'
    local roles='"relay-user"'
    local attributes='{
        "relay.max_tunnels": ["10"],
        "relay.rate_limit": ["100"],
        "relay.allowed_ips": ["192.168.1.0/24"]
    }'
    
    create_user "$admin_token" "user" "user@cloudbridge.local" "Regular" "User" "UserPassword123!" "$groups" "$roles" "$attributes"
}

# Create monitor user
create_monitor_user() {
    local admin_token="$1"
    
    local groups='"relay-monitors"'
    local roles='"relay-monitor"'
    local attributes='{
        "relay.max_tunnels": ["0"],
        "relay.rate_limit": ["50"],
        "relay.allowed_ips": ["10.0.0.0/8"]
    }'
    
    create_user "$admin_token" "monitor" "monitor@cloudbridge.local" "Monitor" "User" "MonitorPassword123!" "$groups" "$roles" "$attributes"
}

# Create custom user
create_custom_user() {
    local admin_token="$1"
    local username="$2"
    local email="$3"
    local password="$4"
    local role="${5:-relay-user}"
    local max_tunnels="${6:-10}"
    local rate_limit="${7:-100}"
    local allowed_ips="${8:-192.168.1.0/24}"
    
    local groups
    case "$role" in
        "relay-admin")
            groups='"relay-admins"'
            ;;
        "relay-monitor")
            groups='"relay-monitors"'
            ;;
        *)
            groups='"relay-users"'
            ;;
    esac
    
    local roles="\"$role\""
    local attributes="{
        \"relay.max_tunnels\": [\"$max_tunnels\"],
        \"relay.rate_limit\": [\"$rate_limit\"],
        \"relay.allowed_ips\": [\"$allowed_ips\"]
    }"
    
    create_user "$admin_token" "$username" "$email" "$username" "User" "$password" "$groups" "$roles" "$attributes"
}

# Test user authentication
test_user_auth() {
    local username="$1"
    local password="$2"
    
    log_info "Testing authentication for user '$username'..."
    
    local token_response=$(curl -s -X POST "$KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=password" \
        -d "client_id=$CLIENT_ID" \
        -d "username=$username" \
        -d "password=$password")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to authenticate user '$username'"
        return 1
    fi
    
    local access_token=$(echo "$token_response" | jq -r '.access_token')
    
    if [ "$access_token" = "null" ] || [ -z "$access_token" ]; then
        log_error "Failed to extract access token for user '$username'"
        echo "Response: $token_response"
        return 1
    fi
    
    log_success "User '$username' authentication successful"
    return 0
}

# List users
list_users() {
    local admin_token="$1"
    
    log_info "Listing users in realm '$REALM_NAME'..."
    
    local response=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users" \
        -H "Authorization: Bearer $admin_token" \
        -H "Content-Type: application/json")
    
    if [ $? -eq 0 ]; then
        local user_count=$(echo "$response" | jq 'length')
        log_success "Found $user_count users in realm '$REALM_NAME'"
        
        echo "$response" | jq -r '.[] | "  - \(.username) (\(.email))"'
    else
        log_error "Failed to list users"
        return 1
    fi
}

# Main function
main() {
    log_info "Starting Keycloak user creation..."
    
    check_dependencies
    
    local admin_token=$(get_admin_token)
    
    local users_created=0
    local users_total=0
    
    # Create admin user
    ((users_total++))
    if create_admin_user "$admin_token"; then
        ((users_created++))
    fi
    
    # Create regular user
    ((users_total++))
    if create_regular_user "$admin_token"; then
        ((users_created++))
    fi
    
    # Create monitor user
    ((users_total++))
    if create_monitor_user "$admin_token"; then
        ((users_created++))
    fi
    
    # Test user authentication
    log_info "Testing user authentication..."
    
    ((users_total++))
    if test_user_auth "admin" "SecurePassword123!"; then
        ((users_created++))
    fi
    
    ((users_total++))
    if test_user_auth "user" "UserPassword123!"; then
        ((users_created++))
    fi
    
    ((users_total++))
    if test_user_auth "monitor" "MonitorPassword123!"; then
        ((users_created++))
    fi
    
    # List users
    list_users "$admin_token"
    
    # Summary
    log_info ""
    log_info "User Creation Summary:"
    log_info "  Users created/tested: $users_created/$users_total"
    
    if [ $users_created -eq $users_total ]; then
        log_success "All users created and tested successfully!"
        log_info ""
        log_info "Default users:"
        log_info "  - admin/SecurePassword123! (relay-admin role)"
        log_info "  - user/UserPassword123! (relay-user role)"
        log_info "  - monitor/MonitorPassword123! (relay-monitor role)"
        log_info ""
        log_info "To create custom users, use:"
        log_info "  ./scripts/create_users.sh custom <username> <email> <password> [role] [max_tunnels] [rate_limit] [allowed_ips]"
        exit 0
    else
        log_error "Some user operations failed. Please check the configuration."
        exit 1
    fi
}

# Custom user creation
if [ "$1" = "custom" ]; then
    if [ $# -lt 5 ]; then
        log_error "Usage: $0 custom <username> <email> <password> [role] [max_tunnels] [rate_limit] [allowed_ips]"
        exit 1
    fi
    
    check_dependencies
    local admin_token=$(get_admin_token)
    
    local username="$2"
    local email="$3"
    local password="$4"
    local role="${5:-relay-user}"
    local max_tunnels="${6:-10}"
    local rate_limit="${7:-100}"
    local allowed_ips="${8:-192.168.1.0/24}"
    
    if create_custom_user "$admin_token" "$username" "$email" "$password" "$role" "$max_tunnels" "$rate_limit" "$allowed_ips"; then
        log_success "Custom user '$username' created successfully"
        if test_user_auth "$username" "$password"; then
            log_success "Custom user authentication test passed"
        fi
    else
        log_error "Failed to create custom user '$username'"
        exit 1
    fi
    
    exit 0
fi

# Run main function
main "$@" 