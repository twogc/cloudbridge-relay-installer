#!/bin/bash

# VK Cloud Environment Setup Script
# This script sets up the environment for VK Cloud deployment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}$1${NC}"
}

log_warning() {
    echo -e "${YELLOW}  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# VK Cloud OpenStack credentials
export OS_AUTH_URL="https://infra.mail.ru:35357/v3/"
export OS_PROJECT_ID="8530029fa7b64199b2ad87c18d4620df"
export OS_REGION_NAME="RegionOne"
unset OS_PROJECT_NAME
unset OS_PROJECT_DOMAIN_ID

# unset v2.0 items in case set
unset OS_TENANT_ID
unset OS_TENANT_NAME

if [[ -z $OS_USERNAME ]] || [[ -z $OS_PASSWORD ]] || [[ "$OS_USERNAME" != "info@2gc.ru" ]]; then
    export OS_USERNAME="info@2gc.ru"
    export OS_USER_DOMAIN_NAME="users"
    
    # Set the OpenStack password directly
    export OS_PASSWORD="7Thbngy6"
fi

export OS_INTERFACE=public
export OS_IDENTITY_API_VERSION=3

# VK Cloud API Token
export VKCS_TOKEN="gAAAAABovHvYKO05TiBofWp8fu5Hr5trS8cEMOr5fwdr2PtYe5Na4_qzQvsyDpWFvoMXe_xob9Jx4w9DjsneYVCCCn5gpLbfWAAcs018YMKAhPT0eLbnZsSfv9gu5E_yE-fqXfFUH7TMoKv_hK06BGrcTSGrpJLyaGGPJyy7jXtb7EOgAjIpqJh0rU-jGQkKHyuU_h9jZXZL"

log_success "VK Cloud environment variables set"
log_info "OS_AUTH_URL: $OS_AUTH_URL"
log_info "OS_PROJECT_ID: $OS_PROJECT_ID"
log_info "OS_USERNAME: $OS_USERNAME"
log_info "OS_REGION_NAME: $OS_REGION_NAME"

# Install OpenStack CLI if not present
if ! command -v openstack &> /dev/null; then
    log_info "Installing OpenStack CLI..."
    sudo apt-get update
    sudo apt-get install -y python3-openstackclient
    log_success "OpenStack CLI installed"
else
    log_success "OpenStack CLI already installed"
fi

# Install VK Cloud Terraform provider
log_info "Installing VK Cloud Terraform provider..."
mkdir -p ~/.terraform.d/plugins/linux_amd64
cd ~/.terraform.d/plugins/linux_amd64

# Download VK Cloud Terraform provider
if [ ! -f "terraform-provider-vkcs" ]; then
    log_info "Downloading VK Cloud Terraform provider..."
    wget -O terraform-provider-vkcs https://github.com/vk-cs/terraform-provider-vkcs/releases/download/v0.12.1/terraform-provider-vkcs_linux_amd64
    chmod +x terraform-provider-vkcs
    log_success "VK Cloud Terraform provider downloaded"
else
    log_success "VK Cloud Terraform provider already exists"
fi

cd - > /dev/null

# Test OpenStack connection
log_info "Testing OpenStack connection..."
if openstack server list --limit 1 > /dev/null 2>&1; then
    log_success "OpenStack connection successful"
else
    log_error "OpenStack connection failed. Please check your credentials."
    exit 1
fi

log_success "VK Cloud environment setup completed!"
log_info "You can now run Terraform commands for VK Cloud deployment"
