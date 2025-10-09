#!/bin/bash

# CloudBridge Relay Production Deployment Script
# Скрипт для развертывания на внешнем сервере edge.2gc.ru

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="cloudbridge-relay"
PRODUCTION_TAG="production-$(date +%Y%m%d-%H%M%S)"
DOCKER_IMAGE="cloudbridge-relay:${PRODUCTION_TAG}"
SERVICE_NAME="cloudbridge-relay"
NGINX_CONFIG="/etc/nginx/sites-available/edge.2gc.ru"
SYSTEMD_SERVICE="/etc/systemd/system/cloudbridge-relay.service"

echo -e "${BLUE}=== CloudBridge Relay Production Deployment ===${NC}"
echo -e "${BLUE}Deployment started at: $(date)${NC}"
echo -e "${BLUE}Production tag: ${PRODUCTION_TAG}${NC}"

# Function to print status
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   print_error "This script must be run as root"
   exit 1
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Nginx is installed
if ! command -v nginx &> /dev/null; then
    print_error "Nginx is not installed. Please install Nginx first."
    exit 1
fi

print_status "Building production Docker image..."

# Build production image
docker build -f Dockerfile.production -t ${DOCKER_IMAGE} .

if [ $? -ne 0 ]; then
    print_error "Failed to build Docker image"
    exit 1
fi

print_status "Docker image built successfully: ${DOCKER_IMAGE}"

# Create application directory
print_status "Creating application directory..."
mkdir -p /opt/cloudbridge-relay/{data,certs,logs}
mkdir -p /var/log/cloudbridge

# Create application user
print_status "Creating application user..."
if ! id "cloudbridge" &>/dev/null; then
    useradd -r -s /bin/false -d /opt/cloudbridge-relay cloudbridge
fi

# Set permissions
chown -R cloudbridge:cloudbridge /opt/cloudbridge-relay
chown -R cloudbridge:cloudbridge /var/log/cloudbridge

# Copy configuration files
print_status "Installing configuration files..."

# Copy Nginx configuration
cp nginx/edge.2gc.ru.conf ${NGINX_CONFIG}
ln -sf ${NGINX_CONFIG} /etc/nginx/sites-enabled/edge.2gc.ru

# Copy systemd service
cp systemd/cloudbridge-relay.service ${SYSTEMD_SERVICE}

# Copy production configuration
cp configs/production.yaml /opt/cloudbridge-relay/config.yaml

print_status "Configuration files installed"

# Generate SSL certificates (if not exist)
if [ ! -f /etc/ssl/certs/edge.2gc.ru.crt ]; then
    print_warning "SSL certificates not found. Generating self-signed certificates..."
    
    # Create certificate directory
    mkdir -p /etc/ssl/certs /etc/ssl/private
    
    # Generate self-signed certificate
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout /etc/ssl/private/edge.2gc.ru.key \
        -out /etc/ssl/certs/edge.2gc.ru.crt \
        -subj "/C=RU/ST=Moscow/L=Moscow/O=2GC/OU=IT/CN=edge.2gc.ru"
    
    print_warning "Self-signed certificates generated. Please replace with real certificates for production use."
fi

# Stop existing service if running
print_status "Stopping existing service..."
systemctl stop ${SERVICE_NAME} 2>/dev/null || true
docker stop cloudbridge-relay 2>/dev/null || true
docker rm cloudbridge-relay 2>/dev/null || true

# Reload systemd
systemctl daemon-reload

# Start the service
print_status "Starting CloudBridge Relay service..."
systemctl enable ${SERVICE_NAME}
systemctl start ${SERVICE_NAME}

# Wait for service to start
print_status "Waiting for service to start..."
sleep 10

# Check service status
if systemctl is-active --quiet ${SERVICE_NAME}; then
    print_status "Service started successfully"
else
    print_error "Service failed to start"
    systemctl status ${SERVICE_NAME}
    exit 1
fi

# Test health endpoint
print_status "Testing health endpoint..."
for i in {1..30}; do
    if curl -f http://localhost:9090/health >/dev/null 2>&1; then
        print_status "Health check passed"
        break
    fi
    if [ $i -eq 30 ]; then
        print_error "Health check failed after 60 seconds"
        exit 1
    fi
    sleep 2
done

# Test API endpoint
print_status "Testing API endpoint..."
if curl -f http://localhost:8082/api/v1/ >/dev/null 2>&1; then
    print_status "API endpoint is accessible"
else
    print_warning "API endpoint test failed (this might be expected if authentication is required)"
fi

# Reload Nginx
print_status "Reloading Nginx..."
nginx -t
if [ $? -eq 0 ]; then
    systemctl reload nginx
    print_status "Nginx reloaded successfully"
else
    print_error "Nginx configuration test failed"
    exit 1
fi

# Test external access
print_status "Testing external access..."
if curl -f https://edge.2gc.ru/health >/dev/null 2>&1; then
    print_status "External health check passed"
else
    print_warning "External health check failed (this might be expected if DNS is not configured)"
fi

# Show service status
print_status "Service status:"
systemctl status ${SERVICE_NAME} --no-pager

# Show logs
print_status "Recent logs:"
journalctl -u ${SERVICE_NAME} --no-pager -n 20

print_status "Deployment completed successfully!"
print_status "Service: ${SERVICE_NAME}"
print_status "Docker image: ${DOCKER_IMAGE}"
print_status "Health endpoint: http://localhost:9090/health"
print_status "API endpoint: http://localhost:8082/api/v1/"
print_status "External endpoint: https://edge.2gc.ru/health"

echo -e "${GREEN}=== Deployment Summary ===${NC}"
echo -e "${GREEN}✓ Docker image built and deployed${NC}"
echo -e "${GREEN}✓ Systemd service configured and started${NC}"
echo -e "${GREEN}✓ Nginx reverse proxy configured${NC}"
echo -e "${GREEN}✓ SSL certificates configured${NC}"
echo -e "${GREEN}✓ Health checks passed${NC}"

echo -e "${BLUE}Next steps:${NC}"
echo -e "${BLUE}1. Configure real SSL certificates${NC}"
echo -e "${BLUE}2. Set up environment variables for Keycloak/Zitadel${NC}"
echo -e "${BLUE}3. Configure DNS to point edge.2gc.ru to this server${NC}"
echo -e "${BLUE}4. Test external connectivity${NC}"
