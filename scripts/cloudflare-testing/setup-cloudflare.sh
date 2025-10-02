#!/bin/bash

# CloudBridge Cloudflare Setup Script
# This script helps setup Cloudflare tunnel for testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="cloudbridge-testing"

echo -e "${BLUE}☁️  CloudBridge Cloudflare Setup${NC}"

# Check if cloudflared is installed locally
if ! command -v cloudflared &> /dev/null; then
    echo -e "${YELLOW}⚠️  cloudflared not found locally. Installing...${NC}"
    
    # Detect OS
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
        sudo dpkg -i cloudflared-linux-amd64.deb
        rm cloudflared-linux-amd64.deb
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        brew install cloudflared
    else
        echo -e "${RED}❌ Unsupported OS. Please install cloudflared manually${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✅ cloudflared is available${NC}"

# Step 1: Login to Cloudflare
echo -e "${BLUE}🔐 Step 1: Login to Cloudflare${NC}"
echo -e "${YELLOW}💡 You need to login to your Cloudflare account${NC}"
echo -e "${YELLOW}💡 This will open a browser window for authentication${NC}"
read -p "Press Enter to continue with cloudflared login..."

cloudflared tunnel login

# Step 2: Create tunnel
echo -e "${BLUE}🚇 Step 2: Create Cloudflare tunnel${NC}"
echo -e "${YELLOW}💡 Creating tunnel for CloudBridge testing${NC}"

TUNNEL_NAME="cloudbridge-testing-$(date +%s)"
TUNNEL_ID=$(cloudflared tunnel create ${TUNNEL_NAME} | grep -o 'Created tunnel [^.]*' | cut -d' ' -f3)

if [ -z "$TUNNEL_ID" ]; then
    echo -e "${RED}❌ Failed to create tunnel${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Tunnel created: ${TUNNEL_ID}${NC}"

# Step 3: Create tunnel configuration
echo -e "${BLUE}📝 Step 3: Create tunnel configuration${NC}"

cat > /tmp/cloudflared-config.yml << EOF
tunnel: ${TUNNEL_ID}
credentials-file: /etc/cloudflared/credentials/credentials.json

# Metrics and health check
metrics: 0.0.0.0:5556
no-autoupdate: true
loglevel: debug

# Ingress rules for testing
ingress:
  # Health check endpoint
  - hostname: health-test.cloudbridge-testing.example.com
    service: http://cloudbridge-relay-testing:5553/health
  
  # Main API endpoint
  - hostname: api-test.cloudbridge-testing.example.com
    service: http://cloudbridge-relay-testing:5554
  
  # Metrics endpoint
  - hostname: metrics-test.cloudbridge-testing.example.com
    service: http://cloudbridge-relay-testing:5552/metrics
  
  # QUIC transport endpoint
  - hostname: quic-test.cloudbridge-testing.example.com
    service: http://cloudbridge-relay-testing:9091
  
  # Catch-all rule
  - service: http_status:404

# Origin certificate (if using)
originRequest:
  noTLSVerify: true
  connectTimeout: 30s
  tlsTimeout: 10s
  tcpKeepAlive: 30s
  keepAliveConnections: 10
  keepAliveTimeout: 90s
  httpHostHeader: ""
  originServerName: ""
  caPool: ""
  noTLSVerify: false
  disableChunkedEncoding: false
  bastionMode: false
  proxyAddress: ""
  proxyPort: 0
  proxyType: ""
  http2Origin: true

# Retry configuration
retries: 5
grace-period: 30s

# Connection settings
protocol: auto
heartbeatCount: 5
heartbeatInterval: 5s

# Logging
logfile: /var/log/cloudflared.log
logfile-max-size: 10
logfile-max-backups: 3
logfile-max-age: 7
EOF

echo -e "${GREEN}✅ Configuration file created: /tmp/cloudflared-config.yml${NC}"

# Step 4: Update Kubernetes secrets
echo -e "${BLUE}🔧 Step 4: Update Kubernetes secrets${NC}"

# Get tunnel credentials
CREDENTIALS_FILE="$HOME/.cloudflared/${TUNNEL_ID}.json"

if [ ! -f "$CREDENTIALS_FILE" ]; then
    echo -e "${RED}❌ Tunnel credentials file not found: ${CREDENTIALS_FILE}${NC}"
    exit 1
fi

# Update the secret in Kubernetes
kubectl create secret generic cloudflared-credentials \
  --from-file=credentials.json=${CREDENTIALS_FILE} \
  --namespace=${NAMESPACE} \
  --dry-run=client -o yaml | kubectl apply -f -

echo -e "${GREEN}✅ Kubernetes secret updated${NC}"

# Step 5: Update ConfigMap with tunnel ID
echo -e "${BLUE}📋 Step 5: Update ConfigMap with tunnel ID${NC}"

# Update the cloudflared ConfigMap
kubectl patch configmap cloudflared-test-config -n ${NAMESPACE} --type merge -p "{\"data\":{\"config.yml\":\"$(cat /tmp/cloudflared-config.yml | sed 's/"/\\"/g' | tr '\n' '\\n')\"}}"

echo -e "${GREEN}✅ ConfigMap updated with tunnel ID: ${TUNNEL_ID}${NC}"

# Step 6: Restart cloudflared deployment
echo -e "${BLUE}🔄 Step 6: Restart cloudflared deployment${NC}"
kubectl rollout restart deployment/cloudflared-testing -n ${NAMESPACE}

# Wait for deployment to be ready
kubectl wait --for=condition=Available deployment/cloudflared-testing -n ${NAMESPACE} --timeout=300s

echo -e "${GREEN}✅ Cloudflared deployment restarted${NC}"

# Step 7: Display connection information
echo -e "${GREEN}🎉 Cloudflare setup completed!${NC}"
echo -e "${BLUE}📋 Connection Information:${NC}"
echo -e "  Tunnel ID: ${TUNNEL_ID}"
echo -e "  Tunnel Name: ${TUNNEL_NAME}"
echo -e "  Namespace: ${NAMESPACE}"
echo -e "  Configuration: /tmp/cloudflared-config.yml"

echo -e "${BLUE}🔗 Next steps:${NC}"
echo -e "  1. Configure DNS records in Cloudflare dashboard"
echo -e "  2. Set up custom domains for testing"
echo -e "  3. Run connectivity tests"
echo -e "  4. Monitor tunnel status"

echo -e "${YELLOW}💡 To check tunnel status:${NC}"
echo -e "  kubectl logs -n ${NAMESPACE} deployment/cloudflared-testing -f"

echo -e "${YELLOW}💡 To test connectivity:${NC}"
echo -e "  ./scripts/cloudflare-testing/test-connectivity.sh"
