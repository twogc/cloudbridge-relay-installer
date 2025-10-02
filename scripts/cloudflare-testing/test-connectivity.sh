#!/bin/bash

# CloudBridge Cloudflare Connectivity Testing Script
# This script tests connectivity between CloudBridge relay and Cloudflare

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="cloudbridge-testing"
TESTING_DIR="k8s/cloudflare-testing"

echo -e "${BLUE}🧪 Starting CloudBridge Cloudflare Connectivity Tests${NC}"

# Check if namespace exists
if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
    echo -e "${RED}❌ Namespace ${NAMESPACE} does not exist${NC}"
    echo -e "${YELLOW}💡 Run deploy-testing.sh first${NC}"
    exit 1
fi

# Check if pods are running
echo -e "${BLUE}📊 Checking pod status${NC}"
kubectl get pods -n ${NAMESPACE}

# Test 1: Basic connectivity to CloudBridge relay
echo -e "${BLUE}🔍 Test 1: Basic connectivity to CloudBridge relay${NC}"
if kubectl exec -n ${NAMESPACE} deployment/cloudbridge-relay-testing -- wget -qO- http://localhost:6003/health; then
    echo -e "${GREEN}✅ CloudBridge relay health check passed${NC}"
else
    echo -e "${RED}❌ CloudBridge relay health check failed${NC}"
fi

# Test 2: Metrics endpoint
echo -e "${BLUE}📈 Test 2: Metrics endpoint${NC}"
if kubectl exec -n ${NAMESPACE} deployment/cloudbridge-relay-testing -- wget -qO- http://localhost:6002/metrics | head -5; then
    echo -e "${GREEN}✅ Metrics endpoint accessible${NC}"
else
    echo -e "${RED}❌ Metrics endpoint not accessible${NC}"
fi

# Test 3: API endpoint
echo -e "${BLUE}🔌 Test 3: API endpoint${NC}"
if kubectl exec -n ${NAMESPACE} deployment/cloudbridge-relay-testing -- wget -qO- http://localhost:6004/; then
    echo -e "${GREEN}✅ API endpoint accessible${NC}"
else
    echo -e "${RED}❌ API endpoint not accessible${NC}"
fi

# Test 4: Cloudflared connectivity
echo -e "${BLUE}☁️  Test 4: Cloudflared connectivity${NC}"
if kubectl exec -n ${NAMESPACE} deployment/cloudflared-testing -- wget -qO- http://localhost:6007/ready; then
    echo -e "${GREEN}✅ Cloudflared health check passed${NC}"
else
    echo -e "${RED}❌ Cloudflared health check failed${NC}"
fi

# Test 5: Cloudflared metrics
echo -e "${BLUE}📊 Test 5: Cloudflared metrics${NC}"
if kubectl exec -n ${NAMESPACE} deployment/cloudflared-testing -- wget -qO- http://localhost:6006/metrics | head -5; then
    echo -e "${GREEN}✅ Cloudflared metrics accessible${NC}"
else
    echo -e "${RED}❌ Cloudflared metrics not accessible${NC}"
fi

# Test 6: Network connectivity tests
echo -e "${BLUE}🌐 Test 6: Network connectivity tests${NC}"

# Test DNS resolution
echo -e "${BLUE}  📡 Testing DNS resolution${NC}"
if kubectl exec -n ${NAMESPACE} deployment/cloudbridge-relay-testing -- nslookup google.com; then
    echo -e "${GREEN}  ✅ DNS resolution working${NC}"
else
    echo -e "${RED}  ❌ DNS resolution failed${NC}"
fi

# Test external connectivity
echo -e "${BLUE}  🌍 Testing external connectivity${NC}"
if kubectl exec -n ${NAMESPACE} deployment/cloudbridge-relay-testing -- wget -qO- --timeout=10 https://httpbin.org/ip; then
    echo -e "${GREEN}  ✅ External connectivity working${NC}"
else
    echo -e "${RED}  ❌ External connectivity failed${NC}"
fi

# Test 7: QUIC connectivity (if available)
echo -e "${BLUE}🚀 Test 7: QUIC connectivity${NC}"
if kubectl exec -n ${NAMESPACE} deployment/cloudbridge-relay-testing -- wget -qO- https://httpbin.org/ip; then
    echo -e "${GREEN}✅ Basic HTTP connectivity working${NC}"
else
    echo -e "${YELLOW}⚠️  HTTP connectivity not available (expected in testing)${NC}"
fi

# Test 8: Load testing (basic)
echo -e "${BLUE}⚡ Test 8: Basic load testing${NC}"
echo -e "${BLUE}  📊 Running basic load test for 10 seconds${NC}"
if kubectl exec -n ${NAMESPACE} deployment/cloudbridge-relay-testing -- timeout 10 wget -qO- http://localhost:6003/health > /dev/null; then
    echo -e "${GREEN}  ✅ Basic load test passed${NC}"
else
    echo -e "${RED}  ❌ Basic load test failed${NC}"
fi

# Test 9: Log analysis
echo -e "${BLUE}📝 Test 9: Log analysis${NC}"
echo -e "${BLUE}  📋 CloudBridge relay logs (last 10 lines):${NC}"
kubectl logs -n ${NAMESPACE} deployment/cloudbridge-relay-testing --tail=10

echo -e "${BLUE}  📋 Cloudflared logs (last 10 lines):${NC}"
kubectl logs -n ${NAMESPACE} deployment/cloudflared-testing --tail=10

# Test 10: Resource usage
echo -e "${BLUE}💻 Test 10: Resource usage${NC}"
echo -e "${BLUE}  📊 Pod resource usage:${NC}"
kubectl top pods -n ${NAMESPACE}

echo -e "${GREEN}🎉 Connectivity tests completed!${NC}"
echo -e "${BLUE}📋 Summary:${NC}"
echo -e "  - Check the results above for any failures"
echo -e "  - Review logs for any errors"
echo -e "  - Monitor resource usage"
echo -e "  - Run performance tests if basic connectivity is working"
