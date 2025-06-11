#!/bin/bash

# CloudBridge Relay Server Installer
# This script installs the CloudBridge Relay Server on Ubuntu Server 22.04 LTS

set -e

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# Create necessary directories
echo "Creating directories..."
mkdir -p /etc/cloudbridge-relay
mkdir -p /var/lib/cloudbridge-relay
mkdir -p /var/log/cloudbridge-relay
mkdir -p /etc/cloudbridge-relay/certs

# Create system user
echo "Creating system user..."
if ! id "cloudbridge-relay" &>/dev/null; then
    useradd -r -s /bin/false cloudbridge-relay
fi

# Download latest release
echo "Downloading latest release..."
LATEST_VERSION=$(curl -s https://api.github.com/repos/twogc/cloudbridge-relay/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
RELEASE_URL="https://github.com/twogc/cloudbridge-relay/releases/download/${LATEST_VERSION}/cloudbridge-relay_${LATEST_VERSION}_linux_amd64.tar.gz"

curl -L -o /tmp/cloudbridge-relay.tar.gz "$RELEASE_URL"
tar xzf /tmp/cloudbridge-relay.tar.gz -C /tmp
mv /tmp/cloudbridge-relay /usr/local/bin/
chmod +x /usr/local/bin/cloudbridge-relay

# Copy systemd service file
echo "Installing systemd service..."
cp "$(dirname "$0")/systemd/cloudbridge-relay.service" /etc/systemd/system/

# Copy config file
echo "Installing configuration..."
cp "$(dirname "$0")/config/config.yaml" /etc/cloudbridge-relay/

# Set permissions
echo "Setting permissions..."
chown -R cloudbridge-relay:cloudbridge-relay /var/lib/cloudbridge-relay
chown -R cloudbridge-relay:cloudbridge-relay /var/log/cloudbridge-relay
chown -R cloudbridge-relay:cloudbridge-relay /etc/cloudbridge-relay
chmod 640 /etc/cloudbridge-relay/config.yaml

# Reload systemd and enable service
echo "Enabling and starting service..."
systemctl daemon-reload
systemctl enable cloudbridge-relay
systemctl start cloudbridge-relay

# Cleanup
rm -f /tmp/cloudbridge-relay.tar.gz

echo "Installation complete!"
echo "The CloudBridge Relay Server is now running."
echo "You can check its status with: systemctl status cloudbridge-relay"
echo "Logs are available at: /var/log/cloudbridge-relay/relay.log"
echo "Configuration is at: /etc/cloudbridge-relay/config.yaml"