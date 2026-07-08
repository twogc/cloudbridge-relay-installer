#!/bin/bash
# Setup GitHub Secrets for CloudBridge Relay CI/CD

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}CloudBridge Relay - GitHub Secrets Setup${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}GitHub CLI (gh) is not installed${NC}"
    echo -e "${YELLOW}Please install it first:${NC}"
    echo "  curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg"
    echo "  echo \"deb [arch=\$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main\" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null"
    echo "  sudo apt update && sudo apt install gh"
    exit 1
fi

# Check if user is authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${RED}Not authenticated with GitHub${NC}"
    echo -e "${YELLOW}Please authenticate first:${NC}"
    echo "  gh auth login"
    exit 1
fi

# Get repository info
REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
echo -e "${GREEN}Repository: ${REPO}${NC}"
echo ""

# Function to set secret
set_secret() {
    local secret_name=$1
    local secret_description=$2
    local secret_value=""
    
    echo -e "${YELLOW}Setting secret: ${secret_name}${NC}"
    echo -e "${BLUE}Description: ${secret_description}${NC}"
    
    # Read secret value
    read -s -p "Enter value for ${secret_name}: " secret_value
    echo ""
    
    if [ -n "$secret_value" ]; then
        echo "$secret_value" | gh secret set "$secret_name"
        echo -e "${GREEN}Secret ${secret_name} set successfully${NC}"
    else
        echo -e "${YELLOW} Skipping ${secret_name} (empty value)${NC}"
    fi
    echo ""
}

# Function to set optional secret
set_optional_secret() {
    local secret_name=$1
    local secret_description=$2
    
    echo -e "${YELLOW}Optional secret: ${secret_name}${NC}"
    echo -e "${BLUE}Description: ${secret_description}${NC}"
    read -p "Do you want to set this secret? (y/N): " -n 1 -r
    echo ""
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        set_secret "$secret_name" "$secret_description"
    else
        echo -e "${YELLOW} Skipping ${secret_name}${NC}"
        echo ""
    fi
}

echo -e "${BLUE}Setting up GitHub Secrets for CI/CD Pipeline${NC}"
echo ""

# Required secrets for basic CI/CD
echo -e "${GREEN}Required Secrets:${NC}"
set_secret "GITHUB_TOKEN" "GitHub token for registry access (usually auto-generated)"

# Optional secrets for advanced features
echo -e "${GREEN}Optional Secrets (for advanced features):${NC}"

# Container Registry secrets
set_optional_secret "REGISTRY_USERNAME" "Container registry username"
set_optional_secret "REGISTRY_PASSWORD" "Container registry password"
set_optional_secret "REGISTRY_IP" "Container registry IP address"

# VK Cloud secrets
set_optional_secret "VK_REGISTRY_USERNAME" "VK Cloud registry username"
set_optional_secret "VK_REGISTRY_PASSWORD" "VK Cloud registry password"
set_optional_secret "VK_CLOUD_OPENRC" "VK Cloud OpenStack credentials"

# Security scanning
set_optional_secret "SNYK_TOKEN" "Snyk security scanning token"
set_optional_secret "CODECOV_TOKEN" "Codecov coverage reporting token"

# Deployment secrets
set_optional_secret "KUBECONFIG" "Kubernetes configuration for deployment"
set_optional_secret "SSH_PRIVATE_KEY" "SSH private key for server deployment"
set_optional_secret "SSH_HOST" "SSH host for deployment"
set_optional_secret "SSH_USER" "SSH user for deployment"

# Monitoring secrets
set_optional_secret "GRAFANA_PASSWORD" "Grafana admin password"
set_optional_secret "PROMETHEUS_PASSWORD" "Prometheus basic auth password"

# Notification secrets
set_optional_secret "SLACK_WEBHOOK" "Slack webhook for notifications"
set_optional_secret "TEAMS_WEBHOOK" "Microsoft Teams webhook for notifications"
set_optional_secret "DISCORD_WEBHOOK" "Discord webhook for notifications"

echo -e "${GREEN}GitHub Secrets setup completed!${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "1. Push your code to trigger the CI/CD pipeline"
echo "2. Check the Actions tab to see the pipeline running"
echo "3. Monitor the build and deployment process"
echo ""
echo -e "${BLUE} Useful links:${NC}"
echo "• GitHub Actions: https://github.com/${REPO}/actions"
echo "• Secrets management: https://github.com/${REPO}/settings/secrets/actions"
echo "• Container registry: https://github.com/${REPO}/packages"
echo ""
echo -e "${GREEN}Your CI/CD pipeline is ready!${NC}"

