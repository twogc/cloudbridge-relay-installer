#!/bin/bash

# Check if domain is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <domain>"
    echo "Example: $0 edge.2gc.ru"
    exit 1
fi

DOMAIN=$1
CERT_DIR="/etc/relay/certs"

# Create certificate directory if it doesn't exist
sudo mkdir -p $CERT_DIR

# Install certbot if not present
if ! command -v certbot &> /dev/null; then
    echo "Installing certbot..."
    sudo apt-get update
    sudo apt-get install -y certbot
fi

# Stop any service that might be using port 80
echo "Stopping services that might use port 80..."
sudo systemctl stop nginx || true
sudo systemctl stop apache2 || true

# Check if port 80 is free
if sudo lsof -i :80 > /dev/null; then
    echo "Error: Port 80 is still in use. Please stop the service using it."
    exit 1
fi

# Obtain certificate
echo "Obtaining certificate for $DOMAIN..."
if ! sudo certbot certonly --standalone -d $DOMAIN --non-interactive --agree-tos --email admin@2gc.ru; then
    echo "Error: Failed to obtain certificate"
    exit 1
fi

# Verify certificate files exist
if [ ! -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ] || [ ! -f "/etc/letsencrypt/live/$DOMAIN/privkey.pem" ]; then
    echo "Error: Certificate files not found"
    exit 1
fi

# Copy certificates to relay directory
echo "Copying certificates..."
sudo cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem $CERT_DIR/
sudo cp /etc/letsencrypt/live/$DOMAIN/privkey.pem $CERT_DIR/

# Set proper permissions
sudo chown -R ubuntu:ubuntu $CERT_DIR
sudo chmod 600 $CERT_DIR/*.pem

echo "Certificate setup completed successfully for $DOMAIN" 