#!/bin/bash

# CloudBridge Relay - Debug Token Script
# Этот скрипт помогает отладить токен пользователя

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
REALM_NAME="${REALM_NAME:-cloudbridge}"
CLIENT_ID="${CLIENT_ID:-relay-client}"
KEYCLOAK_ADMIN="${KEYCLOAK_ADMIN:-admin}"
KEYCLOAK_ADMIN_PASSWORD="${KEYCLOAK_ADMIN_PASSWORD:-Hocmrt7HfWiPRGmV}"

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

# Decode JWT token
decode_jwt() {
    local token="$1"
    
    # Split token into parts
    IFS='.' read -r -a parts <<< "$token"
    
    if [ ${#parts[@]} -ne 3 ]; then
        log_error "Invalid JWT token format"
        return 1
    fi
    
    # Decode header
    echo "=== JWT Header ==="
    echo "${parts[0]}" | base64 -d 2>/dev/null | jq . 2>/dev/null || echo "Failed to decode header"
    echo ""
    
    # Decode payload
    echo "=== JWT Payload ==="
    echo "${parts[1]}" | base64 -d 2>/dev/null | jq . 2>/dev/null || echo "Failed to decode payload"
    echo ""
}

# Get user token
get_user_token() {
    local username="$1"
    local password="$2"
    
    log_info "Getting token for user '$username'..."
    
    local token_response=$(curl -s -X POST "$KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=password" \
        -d "client_id=$CLIENT_ID" \
        -d "username=$username" \
        -d "password=$password" \
        -d "scope=openid profile email roles")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to get token"
        return 1
    fi
    
    local access_token=$(echo "$token_response" | jq -r '.access_token')
    
    if [ "$access_token" = "null" ] || [ -z "$access_token" ]; then
        log_error "Failed to extract access token from response"
        echo "Response: $token_response"
        return 1
    fi
    
    echo "$access_token"
}

# Main function
main() {
    local username="${1:-admin}"
    local password="${2:-admin123}"
    
    log_info "Debugging token for user '$username'..."
    
    # Check dependencies
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed"
        exit 1
    fi
    
    # Get token
    local token=$(get_user_token "$username" "$password")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to get token"
        exit 1
    fi
    
    log_success "Token obtained successfully"
    echo ""
    
    # Decode token
    decode_jwt "$token"
    
    log_info "Token debugging completed"
    log_info "Check the payload section above for roles information"
    log_info "Look for 'realm_access.roles' or 'resource_access' sections"
}

# Run main function
main "$@" 