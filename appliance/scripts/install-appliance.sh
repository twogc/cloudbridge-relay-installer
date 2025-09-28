#!/bin/bash
# CloudBridge Appliance Installation Script

set -e

# Configuration
APP_NAME="cloudbridge-appliance"
APP_VERSION="1.0.0"
INSTALL_DIR="/opt/cloudbridge"
CONFIG_DIR="/etc/cloudbridge"
LOG_DIR="/var/log/cloudbridge"
DATA_DIR="/var/lib/cloudbridge"
SERVICE_USER="cloudbridge"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
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

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

# Check system requirements
check_requirements() {
    log_info "Checking system requirements..."
    
    # Check OS
    if [[ ! -f /etc/os-release ]]; then
        log_error "Cannot determine OS version"
        exit 1
    fi
    
    source /etc/os-release
    if [[ "$ID" != "ubuntu" && "$ID" != "debian" && "$ID" != "centos" && "$ID" != "rhel" ]]; then
        log_warning "Unsupported OS: $ID. Proceeding anyway..."
    fi
    
    # Check memory
    MEMORY_GB=$(free -g | awk '/^Mem:/{print $2}')
    if [[ $MEMORY_GB -lt 8 ]]; then
        log_warning "Low memory: ${MEMORY_GB}GB. Recommended: 16GB+"
    fi
    
    # Check disk space
    DISK_GB=$(df -BG / | awk 'NR==2{print $4}' | sed 's/G//')
    if [[ $DISK_GB -lt 50 ]]; then
        log_warning "Low disk space: ${DISK_GB}GB. Recommended: 100GB+"
    fi
    
    # Check network interfaces
    if ! ip link show | grep -q "state UP"; then
        log_warning "No active network interfaces found"
    fi
    
    log_success "System requirements check completed"
}

# Install dependencies
install_dependencies() {
    log_info "Installing dependencies..."
    
    # Update package lists
    if command -v apt-get &> /dev/null; then
        apt-get update
        apt-get install -y curl wget jq iptables iproute2 wireguard-tools
    elif command -v yum &> /dev/null; then
        yum update -y
        yum install -y curl wget jq iptables iproute2
    elif command -v dnf &> /dev/null; then
        dnf update -y
        dnf install -y curl wget jq iptables iproute2
    else
        log_error "Unsupported package manager"
        exit 1
    fi
    
    log_success "Dependencies installed"
}

# Install Docker
install_docker() {
    log_info "Installing Docker..."
    
    if command -v docker &> /dev/null; then
        log_info "Docker already installed"
        return
    fi
    
    # Install Docker
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
    
    # Start and enable Docker
    systemctl start docker
    systemctl enable docker
    
    # Install Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        chmod +x /usr/local/bin/docker-compose
    fi
    
    log_success "Docker installed"
}

# Create user and directories
setup_user_and_directories() {
    log_info "Setting up user and directories..."
    
    # Create user if not exists
    if ! id "$SERVICE_USER" &>/dev/null; then
        useradd -r -s /bin/false -d "$INSTALL_DIR" "$SERVICE_USER"
    fi
    
    # Create directories
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "$DATA_DIR"
    mkdir -p "$CONFIG_DIR/ssl"
    mkdir -p "$CONFIG_DIR/wireguard"
    
    # Set permissions
    chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"
    chown -R "$SERVICE_USER:$SERVICE_USER" "$LOG_DIR"
    chown -R "$SERVICE_USER:$SERVICE_USER" "$DATA_DIR"
    chmod 755 "$CONFIG_DIR"
    chmod 700 "$CONFIG_DIR/ssl"
    chmod 700 "$CONFIG_DIR/wireguard"
    
    log_success "User and directories created"
}

# Generate SSL certificates
generate_ssl_certificates() {
    log_info "Generating SSL certificates..."
    
    if [[ -f "$CONFIG_DIR/ssl/server.crt" ]]; then
        log_info "SSL certificates already exist"
        return
    fi
    
    # Generate private key
    openssl genrsa -out "$CONFIG_DIR/ssl/server.key" 2048
    
    # Generate certificate
    openssl req -new -x509 -key "$CONFIG_DIR/ssl/server.key" -out "$CONFIG_DIR/ssl/server.crt" -days 365 -subj "/C=RU/ST=Moscow/L=Moscow/O=CloudBridge/OU=IT/CN=cloudbridge.local"
    
    # Generate CA certificate
    cp "$CONFIG_DIR/ssl/server.crt" "$CONFIG_DIR/ssl/ca.crt"
    
    # Set permissions
    chown "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR/ssl"/*
    chmod 600 "$CONFIG_DIR/ssl/server.key"
    chmod 644 "$CONFIG_DIR/ssl/server.crt"
    chmod 644 "$CONFIG_DIR/ssl/ca.crt"
    
    log_success "SSL certificates generated"
}

# Generate WireGuard keys
generate_wireguard_keys() {
    log_info "Generating WireGuard keys..."
    
    if [[ -f "$CONFIG_DIR/wireguard/private.key" ]]; then
        log_info "WireGuard keys already exist"
        return
    fi
    
    # Generate private key
    wg genkey | tee "$CONFIG_DIR/wireguard/private.key" | wg pubkey > "$CONFIG_DIR/wireguard/public.key"
    
    # Set permissions
    chown "$SERVICE_USER:$SERVICE_USER" "$CONFIG_DIR/wireguard"/*
    chmod 600 "$CONFIG_DIR/wireguard/private.key"
    chmod 644 "$CONFIG_DIR/wireguard/public.key"
    
    log_success "WireGuard keys generated"
}

# Create systemd service
create_systemd_service() {
    log_info "Creating systemd service..."
    
    cat > /etc/systemd/system/cloudbridge-appliance.service << EOF
[Unit]
Description=CloudBridge Appliance
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=/usr/local/bin/docker-compose -f $INSTALL_DIR/docker-compose.appliance.yml up
ExecStop=/usr/local/bin/docker-compose -f $INSTALL_DIR/docker-compose.appliance.yml down
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable cloudbridge-appliance
    
    log_success "Systemd service created"
}

# Copy configuration files
copy_config_files() {
    log_info "Copying configuration files..."
    
    # Copy Docker Compose file
    cp docker-compose.appliance.yml "$INSTALL_DIR/"
    
    # Copy configuration
    cp appliance-config.yaml "$CONFIG_DIR/"
    
    # Copy scripts
    cp -r scripts "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/scripts"/*
    
    # Set permissions
    chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"
    
    log_success "Configuration files copied"
}

# Configure firewall
configure_firewall() {
    log_info "Configuring firewall..."
    
    # Check if ufw is available
    if command -v ufw &> /dev/null; then
        ufw allow 8444/tcp comment "CloudBridge gRPC"
        ufw allow 8080/tcp comment "CloudBridge HTTP"
        ufw allow 19302/udp comment "CloudBridge WireGuard"
        ufw allow 3478/udp comment "CloudBridge STUN/TURN"
        ufw allow 5553/udp comment "CloudBridge QUIC"
        log_success "UFW firewall configured"
    elif command -v firewall-cmd &> /dev/null; then
        firewall-cmd --permanent --add-port=8444/tcp
        firewall-cmd --permanent --add-port=8080/tcp
        firewall-cmd --permanent --add-port=19302/udp
        firewall-cmd --permanent --add-port=3478/udp
        firewall-cmd --permanent --add-port=5553/udp
        firewall-cmd --reload
        log_success "Firewalld configured"
    else
        log_warning "No firewall manager found. Please configure firewall manually"
    fi
}

# Start services
start_services() {
    log_info "Starting CloudBridge Appliance..."
    
    # Start Docker services
    cd "$INSTALL_DIR"
    docker-compose -f docker-compose.appliance.yml up -d
    
    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    sleep 30
    
    # Check health
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log_success "CloudBridge Appliance is running"
    else
        log_error "Health check failed. Check logs with: docker-compose logs"
        exit 1
    fi
}

# Display installation summary
display_summary() {
    log_success "CloudBridge Appliance installation completed!"
    echo
    echo "📋 Installation Summary:"
    echo "  • Installation directory: $INSTALL_DIR"
    echo "  • Configuration directory: $CONFIG_DIR"
    echo "  • Log directory: $LOG_DIR"
    echo "  • Data directory: $DATA_DIR"
    echo
    echo "🌐 Access Information:"
    echo "  • Admin interface: http://localhost:8080"
    echo "  • gRPC endpoint: localhost:8444"
    echo "  • Health check: http://localhost:8080/health"
    echo "  • Metrics: http://localhost:8080/metrics"
    echo
    echo "🔧 Management Commands:"
    echo "  • Start: systemctl start cloudbridge-appliance"
    echo "  • Stop: systemctl stop cloudbridge-appliance"
    echo "  • Status: systemctl status cloudbridge-appliance"
    echo "  • Logs: docker-compose -f $INSTALL_DIR/docker-compose.appliance.yml logs"
    echo
    echo "📚 Documentation:"
    echo "  • Configuration: $CONFIG_DIR/appliance-config.yaml"
    echo "  • Logs: $LOG_DIR/"
    echo "  • Data: $DATA_DIR/"
    echo
    echo "🔐 Security Notes:"
    echo "  • Change default passwords in configuration"
    echo "  • Update SSL certificates for production"
    echo "  • Configure firewall rules"
    echo "  • Enable audit logging"
    echo
    echo "✅ Installation completed successfully!"
}

# Main installation function
main() {
    log_info "Starting CloudBridge Appliance installation..."
    
    check_root
    check_requirements
    install_dependencies
    install_docker
    setup_user_and_directories
    generate_ssl_certificates
    generate_wireguard_keys
    create_systemd_service
    copy_config_files
    configure_firewall
    start_services
    display_summary
}

# Run main function
main "$@"
