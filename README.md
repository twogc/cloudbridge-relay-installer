# CloudBridge Relay Server Installer

This repository contains the installation script and configuration files for the CloudBridge Relay Server.

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

2. Create a system user `cloudbridge-relay`

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