#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Version
VERSION="1.0.0"

# Default values
RELAY_VERSION="latest"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/cloudbridge-relay"
DATA_DIR="/var/lib/cloudbridge-relay"
LOG_DIR="/var/log/cloudbridge-relay"
SERVICE_DIR="/etc/systemd/system"
SERVICE_USER="cloudbridge-relay"
REPO_URL="https://github.com/twogc/2GC-CloudBridge-Tunnel"

# Print usage
usage() {
    echo "CloudBridge Relay Server Installer v${VERSION}"
    echo
    echo "Usage: $0 [options]"
    echo
    echo "Options:"
    echo "  -h, --help                 Show this help message"
    echo "  -v, --version              Show version"
    echo "  -d, --debug                Enable debug output"
    echo "  --relay-version VERSION    Install specific version (default: latest)"
    echo "  --install-dir DIR          Installation directory (default: ${INSTALL_DIR})"
    echo "  --config-dir DIR           Configuration directory (default: ${CONFIG_DIR})"
    echo "  --data-dir DIR             Data directory (default: ${DATA_DIR})"
    echo "  --log-dir DIR              Log directory (default: ${LOG_DIR})"
    echo
    echo "Example:"
    echo "  $0 --relay-version 1.0.0"
    echo
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--version)
            echo "CloudBridge Relay Server Installer v${VERSION}"
            exit 0
            ;;
        -d|--debug)
            set -x
            ;;
        --relay-version)
            RELAY_VERSION="$2"
            shift
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift
            ;;
        --config-dir)
            CONFIG_DIR="$2"
            shift
            ;;
        --data-dir)
            DATA_DIR="$2"
            shift
            ;;
        --log-dir)
            LOG_DIR="$2"
            shift
            ;;
        *)
            echo -e "${RED}Error: Unknown option $1${NC}"
            usage
            exit 1
            ;;
    esac
    shift
done

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: Please run as root${NC}"
    exit 1
fi

# Function to print status
print_status() {
    echo -e "${GREEN}==>${NC} $1"
}

# Function to print error
print_error() {
    echo -e "${RED}Error:${NC} $1"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

# Check system requirements
print_status "Checking system requirements..."

# Check if systemd is available
if ! command -v systemctl &> /dev/null; then
    print_error "systemd is not available. This installer requires systemd."
    exit 1
fi

# Check if curl is available
if ! command -v curl &> /dev/null; then
    print_error "curl is not available. Please install curl first."
    exit 1
fi

# Create necessary directories
print_status "Creating directories..."
mkdir -p "${CONFIG_DIR}"
mkdir -p "${DATA_DIR}"
mkdir -p "${LOG_DIR}"

# Create system user
print_status "Creating system user..."
if ! id "${SERVICE_USER}" &>/dev/null; then
    useradd -r -s /bin/false "${SERVICE_USER}" || {
        print_error "Failed to create system user"
        exit 1
    }
fi

# Download and install relay server
print_status "Downloading CloudBridge Relay Server..."
if [ "${RELAY_VERSION}" = "latest" ]; then
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/twogc/2GC-CloudBridge-Tunnel/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    RELAY_VERSION="${LATEST_VERSION}"
fi

DOWNLOAD_URL="https://github.com/twogc/2GC-CloudBridge-Tunnel/releases/download/${RELAY_VERSION}/cloudbridge-relay"
curl -L -o "${INSTALL_DIR}/cloudbridge-relay" "${DOWNLOAD_URL}" || {
    print_error "Failed to download relay server"
    exit 1
}

# Download systemd service file
print_status "Downloading systemd service file..."
curl -L -o "${SERVICE_DIR}/cloudbridge-relay.service" "${REPO_URL}/raw/main/systemd/cloudbridge-relay.service" || {
    print_error "Failed to download systemd service file"
    exit 1
}

# Download default configuration
print_status "Downloading default configuration..."
curl -L -o "${CONFIG_DIR}/config.yaml" "${REPO_URL}/raw/main/config/config.yaml" || {
    print_error "Failed to download configuration"
    exit 1
}

# Set permissions
print_status "Setting permissions..."
chown -R "${SERVICE_USER}:${SERVICE_USER}" "${DATA_DIR}" "${LOG_DIR}"
chmod 755 "${INSTALL_DIR}/cloudbridge-relay"
chmod 644 "${SERVICE_DIR}/cloudbridge-relay.service"
chmod 644 "${CONFIG_DIR}/config.yaml"

# Reload systemd
print_status "Reloading systemd..."
systemctl daemon-reload || {
    print_error "Failed to reload systemd"
    exit 1
}

# Enable and start service
print_status "Enabling and starting service..."
systemctl enable cloudbridge-relay || {
    print_error "Failed to enable service"
    exit 1
}

systemctl start cloudbridge-relay || {
    print_error "Failed to start service"
    exit 1
}

# Print success message
echo
echo -e "${GREEN}CloudBridge Relay Server v${RELAY_VERSION} installed successfully!${NC}"
echo
echo "Service status:"
systemctl status cloudbridge-relay --no-pager
echo
echo "Useful commands:"
echo "  systemctl status cloudbridge-relay    # Check service status"
echo "  journalctl -u cloudbridge-relay -f    # View logs"
echo "  systemctl restart cloudbridge-relay   # Restart service"
echo
echo "Configuration:"
echo "  Config file: ${CONFIG_DIR}/config.yaml"
echo "  Log file: ${LOG_DIR}/relay.log"
echo "  Data directory: ${DATA_DIR}"
echo 