#!/bin/bash
set -e

# Проверяем, есть ли сертификаты для восстановления
if [ -f "scripts/maintenance/restore_certs.sh" ]; then
    echo -e "\033[1;34m[INFO]\033[0m Восстановление существующих сертификатов..."
    bash scripts/maintenance/restore_certs.sh
else
    echo -e "\033[1;34m[INFO]\033[0m Настройка сертификатов (заглушка)"
    # TODO: интегрировать scripts/setup_certs.sh и scripts/restore_certs.sh
fi

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

# Stop Docker containers
echo "Stopping Docker containers..."
cd "$(dirname "$0")/.."  # Go to project root
sudo docker-compose down

# Check if port 80 is free
if sudo lsof -i :80 > /dev/null; then
    echo "Error: Port 80 is still in use. Please stop the service using it."
    exit 1
fi

# Check if certificate already exists
if sudo test -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" && sudo test -f "/etc/letsencrypt/live/$DOMAIN/privkey.pem"; then
    echo "Certificate for $DOMAIN already exists. Checking if renewal is needed..."
    # Try to renew certificate (this will only renew if needed)
    sudo certbot renew --dry-run
    if [ $? -eq 0 ]; then
        echo "Certificate is valid and up to date."
    else
        echo "Certificate needs renewal. Attempting to renew..."
        sudo certbot certonly --standalone -d $DOMAIN --non-interactive --agree-tos --email admin@2gc.ru --force-renewal
    fi
else
    # Obtain new certificate
    echo "Obtaining new certificate for $DOMAIN..."
    if ! sudo certbot certonly --standalone -d $DOMAIN --non-interactive --agree-tos --email admin@2gc.ru; then
        echo "Error: Failed to obtain certificate"
        # Start Docker containers back
        sudo docker-compose up -d
        exit 1
    fi
fi

# Verify certificate files exist and copy if present
if sudo test -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" && sudo test -f "/etc/letsencrypt/live/$DOMAIN/privkey.pem"; then
    echo "Copying certificates..."
    sudo cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem $CERT_DIR/
    sudo cp /etc/letsencrypt/live/$DOMAIN/privkey.pem $CERT_DIR/
    # Set proper permissions
    sudo chown -R ubuntu:ubuntu $CERT_DIR
    sudo chmod 600 $CERT_DIR/*.pem
    echo "Certificates copied to $CERT_DIR."
else
    echo "Error: Certificate files not found after certbot operation."
    # Start Docker containers back
    sudo docker-compose up -d
    exit 1
fi

# Start Docker containers back
echo "Starting Docker containers..."
sudo docker-compose up -d

echo "Certificate setup completed successfully for $DOMAIN" 