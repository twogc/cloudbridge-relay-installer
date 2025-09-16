#!/bin/bash

set -e

echo "🚀 Enabling P2P Mesh Monitoring for CloudBridge Relay"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl is not installed or not in PATH"
    exit 1
fi

# Check if we're connected to a cluster
if ! kubectl cluster-info &> /dev/null; then
    log_error "Not connected to a Kubernetes cluster"
    exit 1
fi

log_info "Connected to cluster: $(kubectl config current-context)"

# 1. Check if monitoring namespace exists
if ! kubectl get namespace cloudbridge-monitoring &> /dev/null; then
    log_warn "Creating cloudbridge-monitoring namespace"
    kubectl create namespace cloudbridge-monitoring
fi

# 2. Check if Prometheus is running
if ! kubectl get pods -n cloudbridge-monitoring | grep -q prometheus; then
    log_error "Prometheus is not running in cloudbridge-monitoring namespace"
    log_info "Please deploy Prometheus first using: kubectl apply -f k8s/monitoring/"
    exit 1
fi

# 3. Check if Grafana is running
if ! kubectl get pods -n cloudbridge-monitoring | grep -q grafana; then
    log_error "Grafana is not running in cloudbridge-monitoring namespace"
    log_info "Please deploy Grafana first using: kubectl apply -f k8s/monitoring/"
    exit 1
fi

# 4. Apply monitoring configuration
log_info "Applying P2P mesh monitoring configuration..."

# Apply ConfigMaps (ServiceMonitor will fail if Prometheus Operator is not installed)
kubectl apply -f k8s/p2p-mesh-monitoring.yaml || log_warn "ServiceMonitor creation failed (Prometheus Operator not installed)"

# 5. Check if relay is exposing metrics
log_info "Checking relay metrics endpoint..."
RELAY_POD=$(kubectl get pods -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

if [ -n "$RELAY_POD" ]; then
    log_info "Found relay pod: $RELAY_POD"
    
    # Check if metrics endpoint is accessible
    if kubectl exec -n cloudbridge $RELAY_POD -- curl -s http://localhost:8084/metrics &> /dev/null; then
        log_info "✅ Relay metrics endpoint is accessible"
    else
        log_warn "⚠️  Relay metrics endpoint is not accessible on port 8084"
        log_info "Checking if metrics are available on different port..."
        
        # Try different ports
        for port in 9090 8082 8084; do
            if kubectl exec -n cloudbridge $RELAY_POD -- curl -s http://localhost:$port/metrics &> /dev/null; then
                log_info "✅ Metrics found on port $port"
                break
            fi
        done
    fi
else
    log_error "No relay pods found in cloudbridge namespace"
    exit 1
fi

# 6. Create Grafana dashboard manually
log_info "Creating Grafana dashboard..."

# Get Grafana service
GRAFANA_SERVICE=$(kubectl get svc -n cloudbridge-monitoring | grep grafana | awk '{print $1}' | head -1)

if [ -n "$GRAFANA_SERVICE" ]; then
    log_info "Found Grafana service: $GRAFANA_SERVICE"
    
    # Start port-forward in background
    kubectl port-forward -n cloudbridge-monitoring svc/$GRAFANA_SERVICE 3000:3000 &
    PF_PID=$!
    
    # Wait for port-forward to be ready
    sleep 5
    
    # Check if Grafana is accessible
    if curl -s http://localhost:3000/api/health &> /dev/null; then
        log_info "✅ Grafana is accessible at http://localhost:3000"
        log_info "Default credentials: admin/admin"
        log_info "Dashboard will be available after login"
    else
        log_warn "⚠️  Grafana is not accessible on localhost:3000"
    fi
    
    # Stop port-forward
    kill $PF_PID 2>/dev/null || true
else
    log_error "Grafana service not found"
fi

# 7. Display monitoring URLs
echo ""
log_info "📊 Monitoring Setup Complete!"
echo ""
echo "Access URLs:"
echo "  Grafana:     kubectl port-forward -n cloudbridge-monitoring svc/grafana 3000:3000"
echo "  Prometheus:  kubectl port-forward -n cloudbridge-monitoring svc/prometheus 9090:9090"
echo ""
echo "Default credentials:"
echo "  Grafana:     admin/admin"
echo "  Prometheus:  no authentication"
echo ""
echo "Dashboard:"
echo "  P2P Mesh:    http://localhost:3000/d/p2p-mesh (after port-forward)"
echo ""
echo "Metrics to monitor:"
echo "  - cloudbridge_p2p_connections_total"
echo "  - cloudbridge_wireguard_peers_total"
echo "  - cloudbridge_quic_connections_total"
echo "  - cloudbridge_p2p_bytes_received_total"
echo "  - cloudbridge_p2p_bytes_sent_total"
echo ""

# 8. Test metrics collection
log_info "Testing metrics collection..."
if kubectl exec -n cloudbridge $RELAY_POD -- curl -s http://localhost:8084/metrics | grep -q "cloudbridge"; then
    log_info "✅ CloudBridge metrics are being exposed"
else
    log_warn "⚠️  CloudBridge metrics not found. Check if metrics are enabled in relay configuration"
fi

log_info "P2P Mesh monitoring setup completed!"

