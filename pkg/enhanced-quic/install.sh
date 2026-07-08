#!/bin/bash

# Enhanced QUIC Installer Script
# This script installs Enhanced QUIC on various platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_URL="https://github.com/twogc/enhanced-quic"
LATEST_RELEASE_URL="https://api.github.com/repos/twogc/enhanced-quic/releases/latest"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/enhanced-quic"
SERVICE_DIR="/etc/systemd/system"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
    esac
    
    echo "${OS}-${ARCH}"
}

# Download latest release
download_release() {
    local platform=$1
    local version=$2
    
    log_info "Downloading Enhanced QUIC ${version} for ${platform}..."
    
    local download_url="${REPO_URL}/releases/download/${version}/enhanced-quic-${platform}"
    local temp_file="/tmp/enhanced-quic-${platform}"
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$temp_file" "$download_url"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$temp_file" "$download_url"
    else
        log_error "Neither curl nor wget is installed. Please install one of them."
        exit 1
    fi
    
    chmod +x "$temp_file"
    echo "$temp_file"
}

# Get latest version
get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        curl -s "$LATEST_RELEASE_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$LATEST_RELEASE_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        log_error "Cannot determine latest version. Please specify version manually."
        exit 1
    fi
}

# Install binary
install_binary() {
    local binary_path=$1
    
    log_info "Installing Enhanced QUIC to ${INSTALL_DIR}..."
    
    if [ ! -d "$INSTALL_DIR" ]; then
        sudo mkdir -p "$INSTALL_DIR"
    fi
    
    sudo cp "$binary_path" "${INSTALL_DIR}/enhanced-quic"
    sudo chmod +x "${INSTALL_DIR}/enhanced-quic"
    
    log_success "Binary installed successfully!"
}

# Create configuration
create_config() {
    log_info "Creating configuration directory..."
    
    if [ ! -d "$CONFIG_DIR" ]; then
        sudo mkdir -p "$CONFIG_DIR"
    fi
    
    # Create default config if it doesn't exist
    if [ ! -f "${CONFIG_DIR}/config.json" ]; then
        cat > /tmp/enhanced-quic-config.json << 'EOF'
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "max_connections": 1000
  },
  "quic": {
    "enabled": true,
    "max_streams": 100,
    "keep_alive": true,
    "idle_timeout": "30s"
  },
  "enhanced": {
    "multi_tenancy": true,
    "ddos_protection": true,
    "geographic_routing": true,
    "metrics": true
  },
  "security": {
    "tls": {
      "enabled": true,
      "cert_file": "/etc/enhanced-quic/cert.pem",
      "key_file": "/etc/enhanced-quic/key.pem"
    }
  }
}
EOF
        sudo cp /tmp/enhanced-quic-config.json "${CONFIG_DIR}/config.json"
        rm /tmp/enhanced-quic-config.json
    fi
    
    log_success "Configuration created!"
}

# Create systemd service
create_service() {
    log_info "Creating systemd service..."
    
    cat > /tmp/enhanced-quic.service << EOF
[Unit]
Description=Enhanced QUIC Service
After=network.target

[Service]
Type=simple
User=enhanced-quic
Group=enhanced-quic
ExecStart=${INSTALL_DIR}/enhanced-quic --config ${CONFIG_DIR}/config.json
Restart=always
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
    
    sudo cp /tmp/enhanced-quic.service "$SERVICE_DIR/"
    rm /tmp/enhanced-quic.service
    
    # Create user if it doesn't exist
    if ! id "enhanced-quic" &>/dev/null; then
        sudo useradd -r -s /bin/false enhanced-quic
    fi
    
    sudo chown enhanced-quic:enhanced-quic "$CONFIG_DIR"
    
    log_success "Systemd service created!"
}

# Install Docker (if requested)
install_docker() {
    log_info "Installing via Docker..."
    
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    docker pull twogc/enhanced-quic:latest
    
    # Create docker-compose.yml
    cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  enhanced-quic:
    image: twogc/enhanced-quic:latest
    container_name: enhanced-quic
    ports:
      - "8080:8080"
      - "8081:8081"
      - "8082:8082"
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
    restart: unless-stopped
    environment:
      - ENVIRONMENT=production
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
EOF
    
    log_success "Docker installation ready! Run 'docker-compose up -d' to start."
}

# Main installation function
main() {
    echo "Enhanced QUIC Installer"
    echo "=========================="
    
    # Parse arguments
    VERSION=""
    DOCKER_MODE=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --docker)
                DOCKER_MODE=true
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --version VERSION    Install specific version"
                echo "  --docker             Install via Docker"
                echo "  --help               Show this help"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Check if running as root
    if [ "$EUID" -eq 0 ]; then
        log_warning "Running as root is not recommended for security reasons."
    fi
    
    # Install via Docker
    if [ "$DOCKER_MODE" = true ]; then
        install_docker
        exit 0
    fi
    
    # Get platform
    PLATFORM=$(detect_platform)
    log_info "Detected platform: $PLATFORM"
    
    # Get version
    if [ -z "$VERSION" ]; then
        VERSION=$(get_latest_version)
        log_info "Latest version: $VERSION"
    fi
    
    # Download and install
    BINARY_PATH=$(download_release "$PLATFORM" "$VERSION")
    install_binary "$BINARY_PATH"
    create_config
    create_service
    
    # Cleanup
    rm -f "$BINARY_PATH"
    
    log_success "Enhanced QUIC installation completed!"
    echo ""
    echo "Next steps:"
    echo "1. Edit configuration: sudo nano ${CONFIG_DIR}/config.json"
    echo "2. Start service: sudo systemctl enable enhanced-quic && sudo systemctl start enhanced-quic"
    echo "3. Check status: sudo systemctl status enhanced-quic"
    echo "4. View logs: sudo journalctl -u enhanced-quic -f"
}

# Run main function
main "$@" 