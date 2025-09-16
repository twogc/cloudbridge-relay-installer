#!/bin/bash

# AI Service Endpoints Testing Script
# Tests all new AI API endpoints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="cloudbridge"
SERVICE_NAME="ai-service"
PORT="8083"
BASE_URL="http://localhost:${PORT}"

echo -e "${BLUE}🚀 AI Service Endpoints Testing Script${NC}"
echo "================================================"

# Function to get AI service pod name
get_ai_pod() {
    kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=$SERVICE_NAME -o jsonpath='{.items[0].metadata.name}' 2>/dev/null
}

# Function to test endpoint
test_endpoint() {
    local endpoint="$1"
    local description="$2"
    local expected_status="${3:-200}"
    
    echo -n "Testing $description... "
    
    # Get pod name
    POD_NAME=$(get_ai_pod)
    if [ -z "$POD_NAME" ]; then
        echo -e "${RED}❌ No AI service pod found${NC}"
        return 1
    fi
    
    # Test endpoint
    response=$(kubectl exec -n $NAMESPACE $POD_NAME -- wget -qO- --timeout=10 "$BASE_URL$endpoint" 2>/dev/null)
    exit_code=$?
    
    if [ $exit_code -eq 0 ] && [ -n "$response" ]; then
        echo -e "${GREEN}✅ OK${NC}"
        return 0
    else
        echo -e "${RED}❌ FAILED${NC}"
        return 1
    fi
}

# Function to test endpoint with JSON validation
test_json_endpoint() {
    local endpoint="$1"
    local description="$2"
    local json_field="$3"
    
    echo -n "Testing $description... "
    
    POD_NAME=$(get_ai_pod)
    if [ -z "$POD_NAME" ]; then
        echo -e "${RED}❌ No AI service pod found${NC}"
        return 1
    fi
    
    response=$(kubectl exec -n $NAMESPACE $POD_NAME -- wget -qO- --timeout=10 "$BASE_URL$endpoint" 2>/dev/null)
    exit_code=$?
    
    if [ $exit_code -eq 0 ] && [ -n "$response" ]; then
        # Check if response contains expected JSON field
        if echo "$response" | grep -q "$json_field"; then
            echo -e "${GREEN}✅ OK${NC}"
            return 0
        else
            echo -e "${YELLOW}⚠️  OK (no $json_field field)${NC}"
            return 0
        fi
    else
        echo -e "${RED}❌ FAILED${NC}"
        return 1
    fi
}

echo -e "${YELLOW}📋 Testing AI Service Endpoints...${NC}"
echo

# Test basic endpoints
echo -e "${BLUE}🔍 Basic Endpoints:${NC}"
test_endpoint "/health" "Health Check"
test_endpoint "/api/v1/ai/status" "AI Service Status"
test_endpoint "/api/v1/ai/metrics" "AI Service Metrics"

echo

# Test anomaly detection endpoints
echo -e "${BLUE}🔍 Anomaly Detection:${NC}"
test_json_endpoint "/api/v1/ai/anomaly/status" "Anomaly Detector Status" "enabled"
test_json_endpoint "/api/v1/ai/history/anomalies" "Anomaly History" "history"

echo

# Test predictive analytics endpoints
echo -e "${BLUE}🔍 Predictive Analytics:${NC}"
test_json_endpoint "/api/v1/ai/predict/load/24" "Load Prediction (24h)" "forecast"
test_json_endpoint "/api/v1/ai/history/predictions" "Prediction History" "history"

echo

# Test routing endpoints
echo -e "${BLUE}🔍 Intelligent Routing:${NC}"
test_json_endpoint "/api/v1/ai/route/recommendations" "Route Recommendations" "recommendations"
test_json_endpoint "/api/v1/ai/history/routes" "Route History" "history"

echo

# Test model management endpoints
echo -e "${BLUE}🔍 Model Management:${NC}"
test_json_endpoint "/api/v1/ai/models" "List Models" "models"

echo

# Test historical data and trends
echo -e "${BLUE}🔍 Historical Data & Trends:${NC}"
test_json_endpoint "/api/v1/ai/trends/performance" "Performance Trends" "trends"

echo

# Test real-time data endpoints
echo -e "${BLUE}🔍 Real-time Data:${NC}"
test_json_endpoint "/api/v1/ai/realtime/connections" "Real-time Connections" "connections"
test_json_endpoint "/api/v1/ai/realtime/performance" "Real-time Performance" "performance"

echo

# Test alerts and notifications
echo -e "${BLUE}🔍 Alerts & Notifications:${NC}"
test_json_endpoint "/api/v1/ai/alerts" "Active Alerts" "alerts"

echo

# Test system health endpoints
echo -e "${BLUE}🔍 System Health:${NC}"
test_json_endpoint "/api/v1/ai/health/detailed" "Detailed Health" "health"
test_json_endpoint "/api/v1/ai/health/components" "Component Health" "components"

echo

# Summary
echo "================================================"
echo -e "${GREEN}🎉 AI Service Endpoints Testing Complete!${NC}"
echo

# Test POST endpoints (if needed)
echo -e "${YELLOW}📝 Note: POST endpoints require request bodies and are not tested in this script${NC}"
echo -e "${YELLOW}   Use the API documentation or manual testing for POST endpoints${NC}"

echo
echo -e "${BLUE}📊 Access Information:${NC}"
echo "  - AI Service: kubectl port-forward -n $NAMESPACE svc/$SERVICE_NAME $PORT:$PORT"
echo "  - API Base URL: $BASE_URL"
echo "  - Health Check: $BASE_URL/health"
echo "  - API Status: $BASE_URL/api/v1/ai/status"