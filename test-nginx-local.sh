#!/bin/bash

# Local Nginx testing script for CloudBridge Relay
# Скрипт для локального тестирования Nginx конфигурации

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Local Nginx Testing for CloudBridge Relay ===${NC}"

# Function to print status
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   print_error "This script must be run as root for Nginx configuration"
   exit 1
fi

# Check if Nginx is installed
if ! command -v nginx &> /dev/null; then
    print_error "Nginx is not installed. Please install Nginx first."
    exit 1
fi

# Check if CloudBridge Relay is running
print_status "Checking if CloudBridge Relay is running..."

# Check if service is running on Minikube
if kubectl get pods -n cloudbridge | grep -q "cloudbridge-relay.*Running"; then
    print_status "CloudBridge Relay is running in Minikube"
    
    # Get Minikube IP
    MINIKUBE_IP=$(minikube ip)
    print_status "Minikube IP: ${MINIKUBE_IP}"
    
    # Test direct connection to Minikube
    print_status "Testing direct connection to Minikube..."
    
    # Test health endpoint
    if curl -f http://${MINIKUBE_IP}:32100/health >/dev/null 2>&1; then
        print_status "✓ Health endpoint accessible at http://${MINIKUBE_IP}:32100/health"
    else
        print_warning "✗ Health endpoint not accessible"
    fi
    
    # Test P2P API endpoint
    if curl -f http://${MINIKUBE_IP}:32099/api/v1/ >/dev/null 2>&1; then
        print_status "✓ P2P API endpoint accessible at http://${MINIKUBE_IP}:32099/api/v1/"
    else
        print_warning "✗ P2P API endpoint not accessible"
    fi
    
else
    print_warning "CloudBridge Relay is not running in Minikube"
    print_status "Please start Minikube and deploy CloudBridge Relay first"
    exit 1
fi

# Install local Nginx configuration
print_status "Installing local Nginx configuration..."

# Copy local test configuration
cp nginx/local-test.conf /etc/nginx/sites-available/cloudbridge-local

# Enable the site
ln -sf /etc/nginx/sites-available/cloudbridge-local /etc/nginx/sites-enabled/cloudbridge-local

# Remove default site if it exists
rm -f /etc/nginx/sites-enabled/default

# Test Nginx configuration
print_status "Testing Nginx configuration..."
nginx -t

if [ $? -ne 0 ]; then
    print_error "Nginx configuration test failed"
    exit 1
fi

# Reload Nginx
print_status "Reloading Nginx..."
systemctl reload nginx

# Wait for Nginx to start
sleep 2

# Test local Nginx endpoints
print_status "Testing local Nginx endpoints..."

# Test health endpoint through Nginx
if curl -f http://localhost:8080/health >/dev/null 2>&1; then
    print_status "✓ Health endpoint accessible through Nginx at http://localhost:8080/health"
    
    # Show health response
    echo -e "${BLUE}Health response:${NC}"
    curl -s http://localhost:8080/health | jq . 2>/dev/null || curl -s http://localhost:8080/health
    echo ""
else
    print_warning "✗ Health endpoint not accessible through Nginx"
fi

# Test P2P API endpoint through Nginx
if curl -f http://localhost:8080/api/v1/ >/dev/null 2>&1; then
    print_status "✓ P2P API endpoint accessible through Nginx at http://localhost:8080/api/v1/"
    
    # Show API response
    echo -e "${BLUE}API response:${NC}"
    curl -s http://localhost:8080/api/v1/ | jq . 2>/dev/null || curl -s http://localhost:8080/api/v1/
    echo ""
else
    print_warning "✗ P2P API endpoint not accessible through Nginx"
fi

# Test metrics endpoint through Nginx
if curl -f http://localhost:8080/metrics >/dev/null 2>&1; then
    print_status "✓ Metrics endpoint accessible through Nginx at http://localhost:8080/metrics"
else
    print_warning "✗ Metrics endpoint not accessible through Nginx"
fi

# Test with JWT token
print_status "Testing with JWT token..."

# Create test token
TEST_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2VkZ2UuMmdjLnJ1L29hdXRoL3YyIiwic3ViIjoic2VydmVyLXNlcnZlci1hLTEyMzQ1IiwiYXVkIjpbImNsb3VkYnJpZGdlLXJlbGF5Il0sImV4cCI6MTc2MDAxMDA1MiwiaWF0IjoxNzU5OTIzNjUyLCJqdGkiOiJqd3RfMTc1OTg3NTk3MTk1NF8xMGhnNjh4MHMiLCJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InJlbGF5OmNvbm5lY3QgcmVsYXk6bWFuYWdlIHRlbmFudDpyZWFkIHRlbmFudDp3cml0ZSIsInRlbmFudF9pZCI6InRlc3QtdGVuYW50LWRldiIsInNlcnZlcl9pZCI6InNlcnZlci1hLTEyMzQ1IiwiY29ubmVjdGlvbl90eXBlIjoicXVpYyIsInBlcm1pc3Npb25zIjpbInAycF9jb25uZWN0IiwicmVsYXk6bWFuYWdlIl19.5cHFXXRO2I_1AEZQ2uxQW8SytZeHddYiCBkYdHseAlo"

# Test POST request with token
echo -e "${BLUE}Testing POST request with JWT token:${NC}"
curl -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer ${TEST_TOKEN}" \
    -d '{"peer_id": "test-peer-nginx-local"}' \
    http://localhost:8080/api/v1/tenants/test-tenant-dev/peers/register \
    -w "\nHTTP Status: %{http_code}\n" \
    -s

echo ""

# Show Nginx status
print_status "Nginx status:"
systemctl status nginx --no-pager -l

# Show Nginx logs
print_status "Recent Nginx logs:"
tail -n 20 /var/log/nginx/access.log 2>/dev/null || echo "No access logs found"
tail -n 10 /var/log/nginx/error.log 2>/dev/null || echo "No error logs found"

print_status "Local Nginx testing completed!"
print_status "Nginx is running on port 8080"
print_status "Health endpoint: http://localhost:8080/health"
print_status "API endpoint: http://localhost:8080/api/v1/"
print_status "Metrics endpoint: http://localhost:8080/metrics"

echo -e "${GREEN}=== Test Summary ===${NC}"
echo -e "${GREEN}✓ Nginx configuration installed${NC}"
echo -e "${GREEN}✓ Nginx service reloaded${NC}"
echo -e "${GREEN}✓ Endpoints tested${NC}"
echo -e "${GREEN}✓ JWT token authentication tested${NC}"

echo -e "${BLUE}You can now test the external configuration by:${NC}"
echo -e "${BLUE}1. Updating your client config to use http://localhost:8080${NC}"
echo -e "${BLUE}2. Testing with curl commands${NC}"
echo -e "${BLUE}3. Running the client test script${NC}"
