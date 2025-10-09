#!/bin/bash

# CloudBridge Relay External Server Deployment
# Используем существующие конфигурации из папки k8s

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="cloudbridge-relay"
PRODUCTION_TAG="external-$(date +%Y%m%d-%H%M%S)"
DOCKER_IMAGE="cloudbridge-relay:${PRODUCTION_TAG}"
SERVICE_NAME="cloudbridge-relay"
NGINX_CONFIG="/etc/nginx/sites-available/edge.2gc.ru"
SYSTEMD_SERVICE="/etc/systemd/system/cloudbridge-relay.service"

echo -e "${BLUE}=== CloudBridge Relay External Deployment ===${NC}"
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

print_status "Building production Docker image using existing configurations..."

# Build production image using existing Dockerfile
docker build -t ${DOCKER_IMAGE} .

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

# Copy existing configuration
print_status "Installing configuration files..."

# Copy existing config from k8s/configmaps
cp k8s/configmaps/cloudbridge-config.yaml /tmp/cloudbridge-config.yaml

# Extract config.json from ConfigMap
print_status "Extracting configuration from Kubernetes ConfigMap..."
python3 -c "
import yaml
import json

with open('/tmp/cloudbridge-config.yaml', 'r') as f:
    config = yaml.safe_load(f)

config_json = config['data']['config.json']
with open('/opt/cloudbridge-relay/config.json', 'w') as f:
    f.write(config_json)

print('Configuration extracted successfully')
"

# Create Nginx configuration for external access
print_status "Creating Nginx configuration..."
cat > ${NGINX_CONFIG} << 'EOF'
# Nginx configuration for edge.2gc.ru
# CloudBridge Relay API Reverse Proxy

upstream cloudbridge_relay {
    server 127.0.0.1:5550;
    keepalive 32;
}

upstream cloudbridge_p2p_api {
    server 127.0.0.1:5552;
    keepalive 32;
}

upstream cloudbridge_metrics {
    server 127.0.0.1:5551;
    keepalive 32;
}

# Rate limiting
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=p2p:10m rate=5r/s;

server {
    listen 443 ssl http2;
    server_name edge.2gc.ru;

    # SSL Configuration (replace with real certificates)
    ssl_certificate /etc/ssl/certs/edge.2gc.ru.crt;
    ssl_certificate_key /etc/ssl/private/edge.2gc.ru.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Content-Security-Policy "default-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self';" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # API Health Check
    location /health {
        proxy_pass http://cloudbridge_relay;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # CORS headers
        add_header Access-Control-Allow-Origin "*" always;
        add_header Access-Control-Allow-Methods "GET, POST, OPTIONS" always;
        add_header Access-Control-Allow-Headers "Origin, Content-Type, Accept, Authorization, X-Requested-With" always;
        
        # Handle preflight requests
        if ($request_method = 'OPTIONS') {
            add_header Access-Control-Allow-Origin "*";
            add_header Access-Control-Allow-Methods "GET, POST, OPTIONS";
            add_header Access-Control-Allow-Headers "Origin, Content-Type, Accept, Authorization, X-Requested-With";
            add_header Access-Control-Max-Age 86400;
            add_header Content-Length 0;
            add_header Content-Type text/plain;
            return 204;
        }
    }

    # P2P API Endpoints
    location /api/v1/ {
        limit_req zone=p2p burst=20 nodelay;
        
        proxy_pass http://cloudbridge_p2p_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # CORS headers
        add_header Access-Control-Allow-Origin "*" always;
        add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS" always;
        add_header Access-Control-Allow-Headers "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-CSRF-Token" always;
        add_header Access-Control-Allow-Credentials "true" always;
        
        # Handle preflight requests
        if ($request_method = 'OPTIONS') {
            add_header Access-Control-Allow-Origin "*";
            add_header Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS";
            add_header Access-Control-Allow-Headers "Origin, Content-Type, Accept, Authorization, X-Requested-With, X-CSRF-Token";
            add_header Access-Control-Allow-Credentials "true";
            add_header Access-Control-Max-Age 86400;
            add_header Content-Length 0;
            add_header Content-Type text/plain;
            return 204;
        }
        
        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    # Metrics endpoint
    location /metrics {
        limit_req zone=api burst=10 nodelay;
        
        proxy_pass http://cloudbridge_metrics;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket support for P2P
    location /ws/ {
        proxy_pass http://cloudbridge_p2p_api;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Default location for other requests
    location / {
        return 404;
    }
}

# HTTP to HTTPS redirect
server {
    listen 80;
    server_name edge.2gc.ru;
    return 301 https://$server_name$request_uri;
}
EOF

# Enable Nginx site
ln -sf ${NGINX_CONFIG} /etc/nginx/sites-enabled/edge.2gc.ru

# Create systemd service
print_status "Creating systemd service..."
cat > ${SYSTEMD_SERVICE} << EOF
[Unit]
Description=CloudBridge Relay Server
Documentation=https://github.com/twogc/cloudbridge-relay-installer
After=network.target docker.service
Requires=docker.service

[Service]
Type=notify
NotifyAccess=all
TimeoutStartSec=0
Restart=always
RestartSec=5
StartLimitBurst=5
StartLimitIntervalSec=60

# Environment variables
Environment=RELAY_JWT_SECRET=85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=

# Working directory
WorkingDirectory=/opt/cloudbridge-relay

# User and group
User=cloudbridge
Group=cloudbridge

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/cloudbridge-relay/data /var/log/cloudbridge

# Resource limits
LimitNOFILE=65536
LimitNPROC=32768

# ExecStart command
ExecStart=/usr/bin/docker run --rm \\
    --name cloudbridge-relay \\
    --network host \\
    --user 1001:1001 \\
    -v /opt/cloudbridge-relay/data:/app/data \\
    -v /opt/cloudbridge-relay/certs:/app/certs \\
    -v /var/log/cloudbridge:/var/log \\
    -v /opt/cloudbridge-relay/config.json:/app/configs/config.json:ro \\
    -e RELAY_JWT_SECRET=85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE= \\
    ${DOCKER_IMAGE}

# ExecStop command
ExecStop=/usr/bin/docker stop cloudbridge-relay

# Health check
ExecStartPost=/bin/sleep 10
ExecStartPost=/bin/bash -c 'for i in {1..30}; do if curl -f http://localhost:5550/health >/dev/null 2>&1; then exit 0; fi; sleep 2; done; exit 1'

[Install]
WantedBy=multi-user.target
EOF

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
sleep 15

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
    if curl -f http://localhost:5550/health >/dev/null 2>&1; then
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
if curl -f http://localhost:5552/api/v1/ >/dev/null 2>&1; then
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
