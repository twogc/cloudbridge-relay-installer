#!/bin/bash

# CloudBridge Magic WAN Connectivity Test
# This script tests connectivity to Cloudflare Magic WAN endpoints

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Magic WAN endpoints
MAGIC_WAN_PRIMARY="162.159.75.92"
MAGIC_WAN_SECONDARY="172.64.251.46"
MAGIC_WAN_PORT="443"

echo -e "${BLUE}🧪 CloudBridge Magic WAN Connectivity Test${NC}"

# Test 1: Direct connectivity to Magic WAN endpoints
echo -e "${BLUE}🌐 Test 1: Direct connectivity to Magic WAN endpoints${NC}"

# Test primary Magic WAN endpoint
echo -e "${BLUE}  🎯 Testing primary Magic WAN endpoint: ${MAGIC_WAN_PRIMARY}:${MAGIC_WAN_PORT}${NC}"
if timeout 10 nc -z ${MAGIC_WAN_PRIMARY} ${MAGIC_WAN_PORT} 2>/dev/null; then
    echo -e "${GREEN}  ✅ Primary Magic WAN endpoint: REACHABLE${NC}"
    PRIMARY_LATENCY=$(ping -c 3 ${MAGIC_WAN_PRIMARY} | tail -1 | awk -F '/' '{print $5}')
    echo -e "${BLUE}  📊 Primary latency: ${PRIMARY_LATENCY}ms${NC}"
else
    echo -e "${RED}  ❌ Primary Magic WAN endpoint: UNREACHABLE${NC}"
fi

# Test secondary Magic WAN endpoint
echo -e "${BLUE}  🎯 Testing secondary Magic WAN endpoint: ${MAGIC_WAN_SECONDARY}:${MAGIC_WAN_PORT}${NC}"
if timeout 10 nc -z ${MAGIC_WAN_SECONDARY} ${MAGIC_WAN_PORT} 2>/dev/null; then
    echo -e "${GREEN}  ✅ Secondary Magic WAN endpoint: REACHABLE${NC}"
    SECONDARY_LATENCY=$(ping -c 3 ${MAGIC_WAN_SECONDARY} | tail -1 | awk -F '/' '{print $5}')
    echo -e "${BLUE}  📊 Secondary latency: ${SECONDARY_LATENCY}ms${NC}"
else
    echo -e "${RED}  ❌ Secondary Magic WAN endpoint: UNREACHABLE${NC}"
fi

# Test 2: HTTP connectivity to Magic WAN
echo -e "${BLUE}🔗 Test 2: HTTP connectivity to Magic WAN${NC}"

# Test HTTPS connectivity to primary
echo -e "${BLUE}  🔒 Testing HTTPS to primary Magic WAN...${NC}"
if timeout 10 curl -k -s --connect-timeout 5 https://${MAGIC_WAN_PRIMARY}:${MAGIC_WAN_PORT} > /dev/null; then
    echo -e "${GREEN}  ✅ Primary HTTPS: CONNECTED${NC}"
else
    echo -e "${YELLOW}  ⚠️  Primary HTTPS: No response (may be normal for Magic WAN)${NC}"
fi

# Test HTTPS connectivity to secondary
echo -e "${BLUE}  🔒 Testing HTTPS to secondary Magic WAN...${NC}"
if timeout 10 curl -k -s --connect-timeout 5 https://${MAGIC_WAN_SECONDARY}:${MAGIC_WAN_PORT} > /dev/null; then
    echo -e "${GREEN}  ✅ Secondary HTTPS: CONNECTED${NC}"
else
    echo -e "${YELLOW}  ⚠️  Secondary HTTPS: No response (may be normal for Magic WAN)${NC}"
fi

# Test 3: CloudBridge Relay status
echo -e "${BLUE}🌉 Test 3: CloudBridge Relay status${NC}"

# Check if CloudBridge relay is running
echo -e "${BLUE}  📊 Checking CloudBridge relay pods...${NC}"
kubectl get pods -n cloudbridge-testing | grep cloudbridge-relay-magic-wan

# Test CloudBridge relay connectivity
echo -e "${BLUE}  🔌 Testing CloudBridge relay connectivity...${NC}"

# Start port forwarding
kubectl port-forward -n cloudbridge-testing svc/cloudbridge-relay-magic-wan 6003:6003 &
PF_PID=$!

# Wait for port forwarding
sleep 3

# Test health endpoint
if curl -s http://localhost:6003/health | jq -r '.status' | grep -q "healthy"; then
    echo -e "${GREEN}  ✅ CloudBridge relay: HEALTHY${NC}"
    HEALTH_RESPONSE=$(curl -s http://localhost:6003/health)
    echo -e "${BLUE}  📊 Uptime: $(echo $HEALTH_RESPONSE | jq -r '.uptime')${NC}"
    echo -e "${BLUE}  📊 Memory: $(echo $HEALTH_RESPONSE | jq -r '.metrics.memory_usage_mb')MB${NC}"
    echo -e "${BLUE}  📊 Connections: $(echo $HEALTH_RESPONSE | jq -r '.metrics.active_connections')${NC}"
else
    echo -e "${RED}  ❌ CloudBridge relay: UNHEALTHY${NC}"
fi

# Cleanup port forwarding
kill $PF_PID 2>/dev/null || true

# Test 4: Network routing to Magic WAN
echo -e "${BLUE}🛣️  Test 4: Network routing to Magic WAN${NC}"

# Test routing to primary
echo -e "${BLUE}  🎯 Testing routing to primary Magic WAN...${NC}"
PRIMARY_ROUTE=$(traceroute -m 5 ${MAGIC_WAN_PRIMARY} 2>/dev/null | tail -1)
echo -e "${BLUE}  📊 Primary route: ${PRIMARY_ROUTE}${NC}"

# Test routing to secondary
echo -e "${BLUE}  🎯 Testing routing to secondary Magic WAN...${NC}"
SECONDARY_ROUTE=$(traceroute -m 5 ${MAGIC_WAN_SECONDARY} 2>/dev/null | tail -1)
echo -e "${BLUE}  📊 Secondary route: ${SECONDARY_ROUTE}${NC}"

# Test 5: DNS resolution
echo -e "${BLUE}🌐 Test 5: DNS resolution${NC}"

# Test DNS resolution for Magic WAN IPs
echo -e "${BLUE}  🔍 Testing DNS resolution...${NC}"
if nslookup ${MAGIC_WAN_PRIMARY} > /dev/null 2>&1; then
    echo -e "${GREEN}  ✅ Primary IP DNS resolution: OK${NC}"
else
    echo -e "${YELLOW}  ⚠️  Primary IP DNS resolution: No reverse DNS${NC}"
fi

if nslookup ${MAGIC_WAN_SECONDARY} > /dev/null 2>&1; then
    echo -e "${GREEN}  ✅ Secondary IP DNS resolution: OK${NC}"
else
    echo -e "${YELLOW}  ⚠️  Secondary IP DNS resolution: No reverse DNS${NC}"
fi

# Test 6: Performance testing
echo -e "${BLUE}⚡ Test 6: Performance testing${NC}"

# Test bandwidth to Magic WAN
echo -e "${BLUE}  📈 Testing bandwidth to Magic WAN...${NC}"
if command -v iperf3 &> /dev/null; then
    echo -e "${BLUE}  🚀 Running iperf3 test to primary Magic WAN...${NC}"
    timeout 10 iperf3 -c ${MAGIC_WAN_PRIMARY} -p 5201 -t 5 2>/dev/null || echo -e "${YELLOW}  ⚠️  iperf3 test not available${NC}"
else
    echo -e "${YELLOW}  ⚠️  iperf3 not available for bandwidth testing${NC}"
fi

# Test 7: Magic WAN service discovery
echo -e "${BLUE}🔍 Test 7: Magic WAN service discovery${NC}"

# Test if Magic WAN responds to different protocols
echo -e "${BLUE}  🔍 Testing Magic WAN service discovery...${NC}"

# Test HTTP/2
echo -e "${BLUE}  📡 Testing HTTP/2 to Magic WAN...${NC}"
if timeout 10 curl -k -s --http2 https://${MAGIC_WAN_PRIMARY}:${MAGIC_WAN_PORT} > /dev/null; then
    echo -e "${GREEN}  ✅ HTTP/2 support: DETECTED${NC}"
else
    echo -e "${YELLOW}  ⚠️  HTTP/2 support: Not detected${NC}"
fi

# Test QUIC
echo -e "${BLUE}  🚀 Testing QUIC to Magic WAN...${NC}"
if timeout 10 curl -k -s --http3 https://${MAGIC_WAN_PRIMARY}:${MAGIC_WAN_PORT} > /dev/null; then
    echo -e "${GREEN}  ✅ QUIC support: DETECTED${NC}"
else
    echo -e "${YELLOW}  ⚠️  QUIC support: Not detected${NC}"
fi

# Final summary
echo -e "${GREEN}🎉 Magic WAN connectivity test completed!${NC}"
echo -e "${BLUE}📋 Test Summary:${NC}"
echo -e "  • Primary Magic WAN (${MAGIC_WAN_PRIMARY}): $(if timeout 5 nc -z ${MAGIC_WAN_PRIMARY} ${MAGIC_WAN_PORT} 2>/dev/null; then echo "✅ REACHABLE"; else echo "❌ UNREACHABLE"; fi)"
echo -e "  • Secondary Magic WAN (${MAGIC_WAN_SECONDARY}): $(if timeout 5 nc -z ${MAGIC_WAN_SECONDARY} ${MAGIC_WAN_PORT} 2>/dev/null; then echo "✅ REACHABLE"; else echo "❌ UNREACHABLE"; fi)"
echo -e "  • CloudBridge Relay: $(kubectl get pods -n cloudbridge-testing | grep cloudbridge-relay-magic-wan | grep Running > /dev/null && echo "✅ RUNNING" || echo "❌ NOT RUNNING")"

echo -e "${BLUE}🏗️  Architecture Status:${NC}"
echo -e "  Client (cloudflared) -> CloudBridge Relay (RU) -> Magic WAN (${MAGIC_WAN_PRIMARY}, ${MAGIC_WAN_SECONDARY})"

echo -e "${YELLOW}💡 Next steps:${NC}"
echo -e "  1. Ensure CloudBridge relay is running properly"
echo -e "  2. Configure real Cloudflare tunnel credentials"
echo -e "  3. Set up DNS records for your domain"
echo -e "  4. Test end-to-end connectivity"
echo -e "  5. Monitor performance and adjust settings"
