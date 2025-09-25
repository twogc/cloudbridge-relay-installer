#!/bin/bash

# Test script for CloudBridge Relay health endpoint
# Usage: ./test-health.sh [host] [port]

set -e

# Default values
RELAY_HOST="${1:-localhost}"
RELAY_PORT="${2:-3456}"
TIMEOUT=10

echo "🧪 Testing CloudBridge Relay health endpoint..."
echo "📍 Host: $RELAY_HOST"
echo "🔌 Port: $RELAY_PORT"
echo "⏱️  Timeout: ${TIMEOUT}s"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS")
            echo -e "${GREEN}$message${NC}"
            ;;
        "FAIL")
            echo -e "${RED}❌ $message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ️  $message${NC}"
            ;;
        "WARN")
            echo -e "${YELLOW}  $message${NC}"
            ;;
    esac
}

# Function to check if curl is available
check_curl() {
    if ! command -v curl &> /dev/null; then
        print_status "FAIL" "curl is not installed. Please install curl to run this test."
        exit 1
    fi
}

# Function to check if jq is available
check_jq() {
    if ! command -v jq &> /dev/null; then
        print_status "WARN" "jq is not installed. JSON responses will not be formatted."
        JQ_AVAILABLE=false
    else
        JQ_AVAILABLE=true
    fi
}

# Function to make HTTP request
make_request() {
    local method=$1
    local path=$2
    local expected_code=$3
    local description=$4
    
    print_status "INFO" "Testing $method $path - $description"
    
    # Make request with timeout
    response=$(curl -s -w "%{http_code}" --max-time $TIMEOUT \
        -H "User-Agent: CloudBridge-Health-Test/1.0" \
        "$RELAY_HOST:$RELAY_PORT$path" 2>/dev/null || echo "000")
    
    http_code="${response: -3}"
    body="${response%???}"
    
    # Handle curl errors
    if [ "$http_code" = "000" ]; then
        print_status "FAIL" "Connection failed or timeout"
        return 1
    fi
    
    # Check HTTP status code
    if [ "$http_code" -eq "$expected_code" ]; then
        print_status "PASS" "HTTP $http_code received"
        
        # Format JSON response if jq is available
        if [ "$JQ_AVAILABLE" = true ] && [ -n "$body" ]; then
            echo "Response:"
            echo "$body" | jq . 2>/dev/null || echo "$body"
        elif [ -n "$body" ]; then
            echo "Response: $body"
        fi
        echo ""
        return 0
    else
        print_status "FAIL" "Expected HTTP $expected_code, got $http_code"
        if [ -n "$body" ]; then
            echo "Response: $body"
        fi
        echo ""
        return 1
    fi
}

# Function to validate health response
validate_health_response() {
    local body=$1
    
    if [ "$JQ_AVAILABLE" = true ]; then
        # Validate required fields
        local status=$(echo "$body" | jq -r '.status' 2>/dev/null)
        local version=$(echo "$body" | jq -r '.version' 2>/dev/null)
        local service=$(echo "$body" | jq -r '.service' 2>/dev/null)
        local timestamp=$(echo "$body" | jq -r '.timestamp' 2>/dev/null)
        local uptime=$(echo "$body" | jq -r '.uptime' 2>/dev/null)
        
        local errors=()
        
        if [ "$status" != "ok" ]; then
            errors+=("status should be 'ok', got '$status'")
        fi
        
        if [ "$service" != "cloudbridge-relay" ]; then
            errors+=("service should be 'cloudbridge-relay', got '$service'")
        fi
        
        if [ -z "$version" ] || [ "$version" = "null" ]; then
            errors+=("version field is missing or null")
        fi
        
        if [ -z "$timestamp" ] || [ "$timestamp" = "null" ]; then
            errors+=("timestamp field is missing or null")
        fi
        
        if [ -z "$uptime" ] || [ "$uptime" = "null" ]; then
            errors+=("uptime field is missing or null")
        fi
        
        # Check if timestamp is valid RFC3339 format
        if ! date -d "$timestamp" >/dev/null 2>&1; then
            errors+=("timestamp is not in valid RFC3339 format")
        fi
        
        if [ ${#errors[@]} -eq 0 ]; then
            print_status "PASS" "Health response validation passed"
        else
            print_status "FAIL" "Health response validation failed:"
            for error in "${errors[@]}"; do
                echo "  - $error"
            done
            return 1
        fi
    else
        print_status "WARN" "Skipping response validation (jq not available)"
    fi
}

# Function to test CORS
test_cors() {
    print_status "INFO" "Testing CORS support"
    
    # Test OPTIONS request
    response=$(curl -s -w "%{http_code}" --max-time $TIMEOUT \
        -H "Origin: http://example.com" \
        -H "Access-Control-Request-Method: GET" \
        -H "Access-Control-Request-Headers: Content-Type" \
        -X OPTIONS \
        "$RELAY_HOST:$RELAY_PORT/health" 2>/dev/null || echo "000")
    
    http_code="${response: -3}"
    
    if [ "$http_code" = "000" ]; then
        print_status "FAIL" "CORS OPTIONS request failed"
        return 1
    elif [ "$http_code" -eq 200 ]; then
        print_status "PASS" "CORS OPTIONS request successful"
    else
        print_status "FAIL" "CORS OPTIONS request returned $http_code"
        return 1
    fi
    
    echo ""
}

# Main test execution
main() {
    local test_results=()
    
    # Check prerequisites
    check_curl
    check_jq
    echo ""
    
    # Test 1: Health endpoint
    print_status "INFO" "=== Test 1: Health Endpoint ==="
    if make_request "GET" "/health" 200 "Health check endpoint"; then
        test_results+=("health:pass")
        # Validate response if we got a successful response
        if [ "$http_code" -eq 200 ] && [ -n "$body" ]; then
            validate_health_response "$body"
        fi
    else
        test_results+=("health:fail")
    fi
    
    # Test 2: Unknown path (404)
    print_status "INFO" "=== Test 2: Unknown Path ==="
    if make_request "GET" "/unknown" 404 "Unknown path should return 404"; then
        test_results+=("404:pass")
    else
        test_results+=("404:fail")
    fi
    
    # Test 3: Root path (404)
    print_status "INFO" "=== Test 3: Root Path ==="
    if make_request "GET" "/" 404 "Root path should return 404"; then
        test_results+=("root:pass")
    else
        test_results+=("root:fail")
    fi
    
    # Test 4: CORS support
    print_status "INFO" "=== Test 4: CORS Support ==="
    if test_cors; then
        test_results+=("cors:pass")
    else
        test_results+=("cors:fail")
    fi
    
    # Test 5: Connection timeout
    print_status "INFO" "=== Test 5: Connection Handling ==="
    print_status "INFO" "Testing with invalid host (should timeout)"
    response=$(curl -s -w "%{http_code}" --max-time 2 \
        "invalid-host:9999/health" 2>/dev/null || echo "000")
    http_code="${response: -3}"
    
    if [ "$http_code" = "000" ]; then
        print_status "PASS" "Invalid host correctly timed out"
        test_results+=("timeout:pass")
    else
        print_status "FAIL" "Invalid host should have timed out"
        test_results+=("timeout:fail")
    fi
    echo ""
    
    # Summary
    print_status "INFO" "=== Test Summary ==="
    local passed=0
    local total=0
    
    for result in "${test_results[@]}"; do
        local test_name="${result%:*}"
        local test_status="${result#*:}"
        total=$((total + 1))
        
        if [ "$test_status" = "pass" ]; then
            print_status "PASS" "$test_name"
            passed=$((passed + 1))
        else
            print_status "FAIL" "$test_name"
        fi
    done
    
    echo ""
    print_status "INFO" "Results: $passed/$total tests passed"
    
    if [ $passed -eq $total ]; then
        print_status "PASS" "All tests passed! Health endpoint is working correctly."
        exit 0
    else
        print_status "FAIL" "Some tests failed. Please check the relay server configuration."
        exit 1
    fi
}

# Run main function
main "$@" 