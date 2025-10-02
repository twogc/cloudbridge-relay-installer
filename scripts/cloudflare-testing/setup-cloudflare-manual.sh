#!/bin/bash

# CloudBridge Cloudflare Manual Setup Script
# This script helps setup Cloudflare tunnel manually for testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="cloudbridge-testing"

echo -e "${BLUE}☁️  CloudBridge Cloudflare Manual Setup${NC}"

# Check if cloudflared is installed
if ! command -v cloudflared &> /dev/null; then
    echo -e "${RED}❌ cloudflared not found. Please install it first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ cloudflared is available${NC}"

# Step 1: Create tunnel manually
echo -e "${BLUE}🚇 Step 1: Create Cloudflare tunnel manually${NC}"
echo -e "${YELLOW}💡 You need to create a tunnel manually in Cloudflare dashboard${NC}"
echo -e "${YELLOW}💡 Or use cloudflared tunnel create command${NC}"

# Instructions for manual setup
echo -e "${BLUE}📋 Manual Setup Instructions:${NC}"
echo -e "1. Go to Cloudflare Dashboard: https://dash.cloudflare.com/"
echo -e "2. Navigate to: Zero Trust > Access > Tunnels"
echo -e "3. Click 'Create a tunnel'"
echo -e "4. Name it: cloudbridge-testing"
echo -e "5. Copy the tunnel token"
echo -e "6. Update the credentials in Kubernetes"

echo -e "${BLUE}🔧 Alternative: Use cloudflared CLI${NC}"
echo -e "Run these commands:"
echo -e "  cloudflared tunnel login"
echo -e "  cloudflared tunnel create cloudbridge-testing"
echo -e "  cloudflared tunnel route dns cloudbridge-testing health-test.yourdomain.com"
echo -e "  cloudflared tunnel route dns cloudbridge-testing api-test.yourdomain.com"

# Step 2: Update Kubernetes secrets
echo -e "${BLUE}🔧 Step 2: Update Kubernetes secrets${NC}"
echo -e "${YELLOW}💡 After creating the tunnel, update the secret:${NC}"

cat << 'EOF'
# Update the secret with your tunnel credentials
kubectl create secret generic cloudflared-credentials \
  --from-file=credentials.json=/path/to/your/credentials.json \
  --namespace=cloudbridge-testing \
  --dry-run=client -o yaml | kubectl apply -f -

# Or update the existing secret
kubectl patch secret cloudflared-credentials -n cloudbridge-testing \
  --type='json' \
  -p='{"data":{"credentials.json":"'$(base64 -w 0 /path/to/your/credentials.json)'"}}'
EOF

# Step 3: Update ConfigMap with real tunnel ID
echo -e "${BLUE}📋 Step 3: Update ConfigMap with real tunnel ID${NC}"
echo -e "${YELLOW}💡 Update the tunnel ID in the ConfigMap:${NC}"

cat << 'EOF'
# Update the ConfigMap with your tunnel ID
kubectl patch configmap cloudflared-test-config -n cloudbridge-testing \
  --type merge -p '{"data":{"config.yml":"tunnel: YOUR_TUNNEL_ID\ncredentials-file: /etc/cloudflared/credentials/credentials.json\n..."}}'
EOF

# Step 4: Restart deployment
echo -e "${BLUE}🔄 Step 4: Restart deployment${NC}"
echo -e "${YELLOW}💡 After updating credentials and config:${NC}"

cat << 'EOF'
# Restart cloudflared deployment
kubectl rollout restart deployment/cloudflared-testing -n cloudbridge-testing

# Check status
kubectl get pods -n cloudbridge-testing
kubectl logs -n cloudbridge-testing deployment/cloudflared-testing -f
EOF

# Step 5: Test connectivity
echo -e "${BLUE}🧪 Step 5: Test connectivity${NC}"
echo -e "${YELLOW}💡 After tunnel is running:${NC}"

cat << 'EOF'
# Test the tunnel endpoints
curl https://health-test.yourdomain.com
curl https://api-test.yourdomain.com
curl https://metrics-test.yourdomain.com

# Or test locally
kubectl port-forward -n cloudbridge-testing svc/cloudbridge-relay-testing 6003:6003
curl http://localhost:6003/health
EOF

echo -e "${GREEN}🎉 Manual setup instructions completed!${NC}"
echo -e "${BLUE}📋 Next steps:${NC}"
echo -e "  1. Create tunnel in Cloudflare dashboard"
echo -e "  2. Update Kubernetes secrets"
echo -e "  3. Update ConfigMap with tunnel ID"
echo -e "  4. Restart deployment"
echo -e "  5. Test connectivity"

echo -e "${YELLOW}💡 For testing without real tunnel, you can:${NC}"
echo -e "  - Use local port forwarding"
echo -e "  - Test internal connectivity"
echo -e "  - Simulate Cloudflare behavior"
