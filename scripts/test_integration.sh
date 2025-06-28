#!/bin/bash

# CloudBridge Relay - Keycloak Integration Test Script
# Этот скрипт тестирует интеграцию Keycloak с CloudBridge Relay

set -e

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
CLIENT_SECRET="${CLIENT_SECRET:-relay-client-secret}"
RELAY_URL="${RELAY_URL:-http://localhost:8080}"

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
    
    if ! command -v go &> /dev/null; then
        log_error "go is required but not installed"
        exit 1
    fi
    
    log_success "Dependencies check passed"
}

# Wait for services to be ready
wait_for_services() {
    log_info "Waiting for services to be ready..."
    
    # Wait for Keycloak
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$KEYCLOAK_URL/health" > /dev/null 2>&1; then
            log_success "Keycloak is ready"
            break
        fi
        
        log_info "Attempt $attempt/$max_attempts - Keycloak not ready yet, waiting..."
        sleep 5
        ((attempt++))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        log_error "Keycloak failed to start within expected time"
        exit 1
    fi
    
    # Wait for Relay
    attempt=1
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$RELAY_URL/health" > /dev/null 2>&1; then
            log_success "Relay is ready"
            break
        fi
        
        log_info "Attempt $attempt/$max_attempts - Relay not ready yet, waiting..."
        sleep 5
        ((attempt++))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        log_error "Relay failed to start within expected time"
        exit 1
    fi
}

# Get admin token
get_admin_token() {
    log_info "Getting admin token..."
    
    local token_response=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=password" \
        -d "client_id=admin-cli" \
        -d "username=admin" \
        -d "password=admin123")
    
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

# Test realm creation
test_realm() {
    local admin_token="$1"
    
    log_info "Testing realm creation..."
    
    local response=$(curl -s -w "%{http_code}" -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME" \
        -H "Authorization: Bearer $admin_token")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Realm '$REALM_NAME' exists"
        return 0
    else
        log_error "Realm '$REALM_NAME' not found"
        return 1
    fi
}

# Test client creation
test_client() {
    local admin_token="$1"
    
    log_info "Testing client creation..."
    
    local response=$(curl -s -w "%{http_code}" -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients?clientId=$CLIENT_ID" \
        -H "Authorization: Bearer $admin_token")
    
    local http_code="${response: -3}"
    local response_body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        local client_count=$(echo "$response_body" | jq 'length')
        if [ "$client_count" -gt 0 ]; then
            log_success "Client '$CLIENT_ID' exists"
            return 0
        fi
    fi
    
    log_error "Client '$CLIENT_ID' not found"
    return 1
}

# Test user authentication
test_user_auth() {
    log_info "Testing user authentication..."
    
    local token_response=$(curl -s -X POST "$KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=password" \
        -d "client_id=$CLIENT_ID" \
        -d "client_secret=$CLIENT_SECRET" \
        -d "username=user" \
        -d "password=user123")
    
    if [ $? -ne 0 ]; then
        log_error "Failed to authenticate user"
        return 1
    fi
    
    local access_token=$(echo "$token_response" | jq -r '.access_token')
    
    if [ "$access_token" = "null" ] || [ -z "$access_token" ]; then
        log_error "Failed to extract access token from user authentication"
        echo "Response: $token_response"
        return 1
    fi
    
    log_success "User authentication successful"
    echo "$access_token"
}

# Test relay integration
test_relay_integration() {
    local access_token="$1"
    
    log_info "Testing relay integration..."
    
    local response=$(curl -s -w "%{http_code}" -X GET "$RELAY_URL/health" \
        -H "Authorization: Bearer $access_token")
    
    local http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Relay integration successful"
        return 0
    else
        log_error "Relay integration failed. HTTP code: $http_code"
        return 1
    fi
}

# Run Go integration tests
run_go_tests() {
    log_info "Running Go integration tests..."
    
    if [ -d "tests/integration" ]; then
        cd tests/integration
        go test -v -run TestKeycloakIntegration
        if [ $? -eq 0 ]; then
            log_success "Go integration tests passed"
        else
            log_error "Go integration tests failed"
            return 1
        fi
        cd ../..
    else
        log_warning "Integration tests directory not found"
    fi
}

# Test JWKS endpoint
test_jwks() {
    log_info "Testing JWKS endpoint..."
    
    local response=$(curl -s -w "%{http_code}" -X GET "$KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/certs")
    
    local http_code="${response: -3}"
    local response_body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        local keys_count=$(echo "$response_body" | jq '.keys | length')
        if [ "$keys_count" -gt 0 ]; then
            log_success "JWKS endpoint working correctly ($keys_count keys)"
            return 0
        fi
    fi
    
    log_error "JWKS endpoint test failed"
    return 1
}

# Test OpenID Connect configuration
test_oidc_config() {
    log_info "Testing OpenID Connect configuration..."
    
    local response=$(curl -s -w "%{http_code}" -X GET "$KEYCLOAK_URL/realms/$REALM_NAME/.well-known/openid_configuration")
    
    local http_code="${response: -3}"
    local response_body="${response%???}"
    
    if [ "$http_code" = "200" ]; then
        local issuer=$(echo "$response_body" | jq -r '.issuer')
        if [ "$issuer" = "$KEYCLOAK_URL/realms/$REALM_NAME" ]; then
            log_success "OpenID Connect configuration correct"
            return 0
        fi
    fi
    
    log_error "OpenID Connect configuration test failed"
    return 1
}

# Main function
main() {
    log_info "Starting Keycloak integration tests..."
    
    check_dependencies
    wait_for_services
    
    local admin_token=$(get_admin_token)
    
    local tests_passed=0
    local tests_total=0
    
    # Test realm
    ((tests_total++))
    if test_realm "$admin_token"; then
        ((tests_passed++))
    fi
    
    # Test client
    ((tests_total++))
    if test_client "$admin_token"; then
        ((tests_passed++))
    fi
    
    # Test JWKS
    ((tests_total++))
    if test_jwks; then
        ((tests_passed++))
    fi
    
    # Test OpenID Connect configuration
    ((tests_total++))
    if test_oidc_config; then
        ((tests_passed++))
    fi
    
    # Test user authentication
    ((tests_total++))
    local user_token=$(test_user_auth)
    if [ $? -eq 0 ]; then
        ((tests_passed++))
        
        # Test relay integration
        ((tests_total++))
        if test_relay_integration "$user_token"; then
            ((tests_passed++))
        fi
    fi
    
    # Run Go tests
    ((tests_total++))
    if run_go_tests; then
        ((tests_passed++))
    fi
    
    # Summary
    log_info ""
    log_info "Test Summary:"
    log_info "  Tests passed: $tests_passed/$tests_total"
    
    if [ $tests_passed -eq $tests_total ]; then
        log_success "All integration tests passed!"
        exit 0
    else
        log_error "Some tests failed. Please check the configuration."
        exit 1
    fi
}

# Run main function
main "$@" 