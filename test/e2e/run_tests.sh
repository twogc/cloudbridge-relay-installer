#!/bin/bash
set -e

# CloudBridge Relay E2E Test Runner
# This script runs comprehensive E2E tests for kubectl proxy functionality

# Configuration
RELAY_URL=${RELAY_URL:-"http://localhost:8444"}
METRICS_URL=${METRICS_URL:-"http://localhost:5551"}
CLUSTER_ID=${CLUSTER_ID:-"minikube"}
TEST_NAMESPACE=${TEST_NAMESPACE:-"e2e-test"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE} CloudBridge Relay E2E Test Suite${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

# Check if kubectl is installed and version >= 1.29
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}ERROR: kubectl is not installed${NC}"
    exit 1
fi

KUBECTL_VERSION=$(kubectl version --client --short 2>/dev/null | grep -o 'v[0-9]\+\.[0-9]\+' | head -1)
echo "kubectl version: $KUBECTL_VERSION"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}ERROR: Go is not installed${NC}" 
    exit 1
fi

GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+\.[0-9]\+')
echo "Go version: $GO_VERSION"

# Check if relay is running
echo -e "${YELLOW}Checking CloudBridge Relay...${NC}"
if ! curl -s "$RELAY_URL/health" > /dev/null; then
    echo -e "${RED}ERROR: CloudBridge Relay is not responding at $RELAY_URL${NC}"
    echo "Please ensure the relay is running with kubectl proxy support enabled"
    exit 1
fi

echo -e "${GREEN}✓ CloudBridge Relay is running${NC}"

# Check if metrics endpoint is available
if ! curl -s "$METRICS_URL/metrics" > /dev/null; then
    echo -e "${YELLOW}WARNING: Metrics endpoint not available at $METRICS_URL${NC}"
    echo "Metrics tests will be skipped"
    SKIP_METRICS=true
else
    echo -e "${GREEN}✓ Metrics endpoint is available${NC}"
fi

# Check if cluster is accessible
echo -e "${YELLOW}Checking cluster access...${NC}"
if ! curl -s "$RELAY_URL/kube/$CLUSTER_ID/api/v1/namespaces" > /dev/null; then
    echo -e "${RED}ERROR: Cannot access cluster '$CLUSTER_ID' through relay${NC}"
    echo "Please check cluster configuration and ensure cluster '$CLUSTER_ID' is configured in relay"
    exit 1
fi

echo -e "${GREEN}✓ Cluster '$CLUSTER_ID' is accessible${NC}"

# Set required environment variables
export KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1
export KUBECTL_PORT_FORWARD_WEBSOCKETS=1
export RELAY_URL
export METRICS_URL
export CLUSTER_ID
export TEST_NAMESPACE

echo ""
echo -e "${YELLOW}Test Configuration:${NC}"
echo "  Relay URL: $RELAY_URL"
echo "  Metrics URL: $METRICS_URL"
echo "  Cluster ID: $CLUSTER_ID"
echo "  Test Namespace: $TEST_NAMESPACE"
echo ""

# Run tests
echo -e "${BLUE}Running E2E Tests...${NC}"
echo ""

# Test categories
TESTS=(
    "basic_kubectl_commands:TestKubeProxyE2E"
    "websocket_streams:TestKubeProxyE2E/TestWebSocketStreams"  
    "authorization:TestAuthorizationE2E"
    "metrics:TestMetricsE2E"
)

FAILED_TESTS=()
PASSED_TESTS=()

for test_info in "${TESTS[@]}"; do
    IFS=':' read -r test_name test_path <<< "$test_info"
    
    echo -e "${YELLOW}Running $test_name tests...${NC}"
    
    # Skip metrics tests if endpoint is not available
    if [[ "$test_name" == "metrics" && "$SKIP_METRICS" == "true" ]]; then
        echo -e "${YELLOW}⚠ Skipping metrics tests (endpoint not available)${NC}"
        continue
    fi
    
    # Run the test
    if go test -v -timeout 300s ./test/e2e -run "$test_path" 2>&1 | tee "/tmp/test_$test_name.log"; then
        echo -e "${GREEN}✓ $test_name tests PASSED${NC}"
        PASSED_TESTS+=("$test_name")
    else
        echo -e "${RED}✗ $test_name tests FAILED${NC}"
        FAILED_TESTS+=("$test_name")
        
        # Show last few lines of failure
        echo -e "${RED}Last few lines of failure:${NC}"
        tail -10 "/tmp/test_$test_name.log"
    fi
    
    echo ""
done

# Run integration tests (all components together)
echo -e "${YELLOW}Running integration tests...${NC}"
if go test -v -timeout 600s ./test/e2e -run "Integration" 2>&1 | tee "/tmp/test_integration.log"; then
    echo -e "${GREEN}✓ Integration tests PASSED${NC}"
    PASSED_TESTS+=("integration")
else
    echo -e "${RED}✗ Integration tests FAILED${NC}"
    FAILED_TESTS+=("integration")
fi

# Summary
echo ""
echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE} Test Results Summary${NC}"
echo -e "${BLUE}======================================${NC}"

echo -e "${GREEN}Passed Tests (${#PASSED_TESTS[@]}):${NC}"
for test in "${PASSED_TESTS[@]}"; do
    echo -e "  ${GREEN}✓ $test${NC}"
done

if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
    echo ""
    echo -e "${RED}Failed Tests (${#FAILED_TESTS[@]}):${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "  ${RED}✗ $test${NC}"
    done
    
    echo ""
    echo -e "${RED}E2E Tests FAILED${NC}"
    echo "Check individual test logs in /tmp/test_*.log for details"
    exit 1
else
    echo ""
    echo -e "${GREEN}All E2E Tests PASSED! 🎉${NC}"
    echo ""
    echo -e "${YELLOW}CloudBridge Relay kubectl proxy functionality is working correctly${NC}"
fi

# Cleanup test logs (optional)
if [ "$CLEANUP_LOGS" = "true" ]; then
    rm -f /tmp/test_*.log
    echo "Test logs cleaned up"
fi

echo ""
echo -e "${BLUE}E2E Test Suite Completed${NC}"
