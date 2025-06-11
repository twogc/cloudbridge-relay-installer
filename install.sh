# CloudBridge Relay Server Installer

<<<<<<< HEAD
This repository contains the installation script and configuration files for the CloudBridge Relay Server.
=======
# CloudBridge Relay Server Installer
# This script installs the CloudBridge Relay Server on Ubuntu Server 22.04 LTS

set -e
>>>>>>> d9da572 (feat: add systemd service and config files)

## Contents

- `install.sh` - Main installation script
- `systemd/cloudbridge-relay.service` - Systemd service file
- `config/config.yaml` - Default configuration file
- `LICENSE` - MIT License

## Quick Install

To install the CloudBridge Relay Server, run:

```bash
curl -sSL https://raw.githubusercontent.com/twogc/cloudbridge-relay-installer/main/install.sh | sudo bash
```

## Manual Installation

1. Clone this repository:
```bash
git clone https://github.com/twogc/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer
```

2. Make the installer executable:
```bash
chmod +x install.sh
```

3. Run the installer:
```bash
sudo ./install.sh
```

## What Gets Installed

The installer will:

1. Create necessary directories:
   - `/etc/cloudbridge-relay/` - Configuration files
   - `/var/lib/cloudbridge-relay/` - Data directory
   - `/var/log/cloudbridge-relay/` - Log files

<<<<<<< HEAD
2. Create a system user `cloudbridge-relay`
=======
# Create necessary directories
print_status "Creating directories..."
mkdir -p "${CONFIG_DIR}"
mkdir -p "${DATA_DIR}"
mkdir -p "${LOG_DIR}"
mkdir -p /etc/cloudbridge-relay
mkdir -p /etc/cloudbridge-relay/certs
>>>>>>> d9da572 (feat: add systemd service and config files)

3. Download and install the latest release of CloudBridge Relay Server

4. Install and configure the systemd service

5. Set up the default configuration

6. Start the service

## Post-Installation

After installation:

1. Check the service status:
```bash
systemctl status cloudbridge-relay
```

2. View the logs:
```bash
tail -f /var/log/cloudbridge-relay/relay.log
```

3. Edit the configuration:
```bash
sudo nano /etc/cloudbridge-relay/config.yaml
```

<<<<<<< HEAD
4. Restart the service after configuration changes:
```bash
sudo systemctl restart cloudbridge-relay
```

## Configuration

The default configuration is installed at `/etc/cloudbridge-relay/config.yaml`. You can modify it to:

- Change the listening port
- Enable/disable TLS
- Configure rate limiting
- Adjust logging settings
- Enable/disable metrics

## Security

The installer:
- Creates a dedicated system user
- Sets appropriate file permissions
- Configures systemd service with security hardening
- Uses TLS by default

## Support

For issues and feature requests, please visit:
https://github.com/twogc/cloudbridge-relay/issues

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
=======
# Print success message
echo
echo -e "${GREEN}CloudBridge Client v${CLIENT_VERSION} installed successfully!${NC}"
echo
echo "Service status:"
systemctl status cloudbridge-client --no-pager
echo
echo "Useful commands:"
echo "  systemctl status cloudbridge-client    # Check service status"
echo "  journalctl -u cloudbridge-client -f    # View logs"
echo "  systemctl restart cloudbridge-client   # Restart service"
echo
echo "Configuration:"
echo "  Config file: ${CONFIG_DIR}/config.yaml"
echo "  Log file: ${LOG_DIR}/client.log"
echo "  Data directory: ${DATA_DIR}"

# Download latest release
print_status "Downloading latest release..."
LATEST_VERSION=$(curl -s https://api.github.com/repos/twogc/cloudbridge-relay/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
RELEASE_URL="https://github.com/twogc/cloudbridge-relay/releases/download/${LATEST_VERSION}/cloudbridge-relay_${LATEST_VERSION}_linux_amd64.tar.gz"

curl -L -o /tmp/cloudbridge-relay.tar.gz "$RELEASE_URL"
tar xzf /tmp/cloudbridge-relay.tar.gz -C /tmp
mv /tmp/cloudbridge-relay /usr/local/bin/
chmod +x /usr/local/bin/cloudbridge-relay

# Copy systemd service file
print_status "Installing systemd service..."
cp "$(dirname "$0")/systemd/cloudbridge-relay.service" /etc/systemd/system/

# Copy config file
print_status "Installing configuration..."
cp "$(dirname "$0")/config/config.yaml" /etc/cloudbridge-relay/

# Set permissions
print_status "Setting permissions..."
chown -R cloudbridge-relay:cloudbridge-relay /var/lib/cloudbridge-relay
chown -R cloudbridge-relay:cloudbridge-relay /var/log/cloudbridge-relay
chown -R cloudbridge-relay:cloudbridge-relay /etc/cloudbridge-relay
chmod 640 /etc/cloudbridge-relay/config.yaml

# Reload systemd and enable service
print_status "Enabling and starting service..."
systemctl daemon-reload
systemctl enable cloudbridge-relay
systemctl start cloudbridge-relay

# Cleanup
rm -f /tmp/cloudbridge-relay.tar.gz

echo
echo "Installation complete!"
echo "The CloudBridge Relay Server is now running."
echo "You can check its status with: systemctl status cloudbridge-relay"
echo "Logs are available at: /var/log/cloudbridge-relay/relay.log"
echo "Configuration is at: /etc/cloudbridge-relay/config.yaml"
echo 
>>>>>>> d9da572 (feat: add systemd service and config files)
