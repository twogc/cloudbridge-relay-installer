#!/bin/bash

set -euo pipefail

# Variables
NGINX_CONFIG_FILE="nginx/edge.2gc.ru-fixed.conf"
NGINX_SITES_AVAILABLE="/etc/nginx/sites-available/edge.2gc.ru.conf"
NGINX_SITES_ENABLED="/etc/nginx/sites-enabled/edge.2gc.ru.conf"

# --- Functions ---
print_status() {
    echo -e "\033[0;34m=== $1 ===\033[0m"
}

print_info() {
    echo -e "\033[0;32m[INFO]\033[0m $1"
}

print_warning() {
    echo -e "\033[1;33m[WARNING]\033[0m $1"
}

print_error() {
    echo -e "\033[0;31m[ERROR]\033[0m $1"
    exit 1
}

# --- Main Script ---
print_status "Fixing Nginx Configuration for edge.2gc.ru"

# 1. Backup existing configuration
print_info "Backing up existing Nginx configuration..."
if [ -f "${NGINX_SITES_AVAILABLE}" ]; then
    sudo cp "${NGINX_SITES_AVAILABLE}" "${NGINX_SITES_AVAILABLE}.backup.$(date +%Y%m%d_%H%M%S)"
    print_info "Backup created: ${NGINX_SITES_AVAILABLE}.backup.$(date +%Y%m%d_%H%M%S)"
else
    print_warning "No existing configuration found to backup"
fi

# 2. Copy new configuration
print_info "Copying new Nginx configuration..."
sudo cp "${NGINX_CONFIG_FILE}" "${NGINX_SITES_AVAILABLE}" || print_error "Failed to copy Nginx config"

# 3. Create symlink to sites-enabled
print_info "Creating symlink to /etc/nginx/sites-enabled/..."
sudo ln -sf "${NGINX_SITES_AVAILABLE}" "${NGINX_SITES_ENABLED}" || print_error "Failed to create symlink"

# 4. Test Nginx configuration
print_info "Testing Nginx configuration..."
sudo nginx -t || print_error "Nginx configuration test failed"

# 5. Reload Nginx
print_info "Reloading Nginx..."
sudo systemctl reload nginx || print_error "Failed to reload Nginx"

print_info "Nginx configuration updated successfully!"
print_info "New configuration uses local ports:"
print_info "  - 127.0.0.1:5550 (Main server)"
print_info "  - 127.0.0.1:5552 (P2P API)"
print_info "  - 127.0.0.1:5551 (Metrics)"

# 6. Test endpoints
print_info "Testing endpoints..."
sleep 2

# Test health endpoint
print_info "Testing health endpoint..."
if curl -f https://edge.2gc.ru/health >/dev/null 2>&1; then
    print_info "✅ Health endpoint is working"
else
    print_warning "⚠️  Health endpoint test failed (this might be expected if CloudBridge Relay is not running locally)"
fi

# Test API endpoint
print_info "Testing API endpoint..."
if curl -f https://edge.2gc.ru/api/v1/ >/dev/null 2>&1; then
    print_info "✅ API endpoint is working"
else
    print_warning "⚠️  API endpoint test failed (this might be expected if authentication is required)"
fi

print_info "Nginx configuration fixed! Now you need to:"
print_info "1. Deploy CloudBridge Relay locally on edge.2gc.ru server"
print_info "2. Make sure it's running on ports 5550, 5551, 5552"
print_info "3. Test the endpoints again"
