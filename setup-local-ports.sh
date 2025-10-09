#!/bin/bash

# CloudBridge Relay Local Port Setup Script
# Настройка локальных портов для CloudBridge Relay

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== CloudBridge Relay Local Port Setup ===${NC}"

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
   print_error "This script must be run as root"
   exit 1
fi

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if minikube is running
if ! minikube status &> /dev/null; then
    print_error "Minikube is not running. Please start Minikube first."
    exit 1
fi

print_status "Setting up port forwarding from Minikube to local ports..."

# Kill existing port forwarding processes
print_status "Stopping existing port forwarding processes..."
pkill -f "kubectl port-forward" 2>/dev/null || true
sleep 2

# Start port forwarding for CloudBridge Relay
print_status "Starting port forwarding..."

# Port 5550 (Main server) - 32090 -> 5550
kubectl port-forward -n cloudbridge service/cloudbridge-relay-external 5550:5550 &
PF_PID_5550=$!

# Port 5551 (Metrics) - 32091 -> 5551  
kubectl port-forward -n cloudbridge service/cloudbridge-relay-external 5551:5551 &
PF_PID_5551=$!

# Port 5552 (P2P API) - 32092 -> 5552
kubectl port-forward -n cloudbridge service/cloudbridge-relay-external 5552:5552 &
PF_PID_5552=$!

# Wait for port forwarding to start
print_status "Waiting for port forwarding to start..."
sleep 5

# Test the ports
print_status "Testing local ports..."

# Test port 5550 (Main server)
if curl -f http://localhost:5550/health >/dev/null 2>&1; then
    print_status "Port 5550 (Main server) is working"
else
    print_warning "Port 5550 (Main server) test failed"
fi

# Test port 5551 (Metrics)
if curl -f http://localhost:5551/metrics >/dev/null 2>&1; then
    print_status "Port 5551 (Metrics) is working"
else
    print_warning "Port 5551 (Metrics) test failed"
fi

# Test port 5552 (P2P API)
if curl -f http://localhost:5552/api/v1/ >/dev/null 2>&1; then
    print_status "Port 5552 (P2P API) is working"
else
    print_warning "Port 5552 (P2P API) test failed"
fi

# Save PIDs to file for later cleanup
echo $PF_PID_5550 > /tmp/cloudbridge-pf-5550.pid
echo $PF_PID_5551 > /tmp/cloudbridge-pf-5551.pid
echo $PF_PID_5552 > /tmp/cloudbridge-pf-5552.pid

print_status "Port forwarding setup completed!"
print_status "Local ports:"
print_status "  - 5550: Main server (http://localhost:5550/health)"
print_status "  - 5551: Metrics (http://localhost:5551/metrics)"
print_status "  - 5552: P2P API (http://localhost:5552/api/v1/)"

print_status "To stop port forwarding, run:"
print_status "  pkill -f 'kubectl port-forward'"

echo -e "${GREEN}=== Setup Complete ===${NC}"
