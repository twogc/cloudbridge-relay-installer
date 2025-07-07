#!/bin/bash

# CloudBridge Admin UI Management Script
# Usage: ./manage_ui.sh [start|stop|restart|status|enable|disable]

SERVICE_NAME="ui-admin"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
PROJECT_DIR="/home/ubuntu/cloudbridge-relay-installer"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Function to install service
install_service() {
    print_status "Installing UI service..."
    
    # Copy service file
    cp "${PROJECT_DIR}/systemd/${SERVICE_NAME}.service" "$SERVICE_FILE"
    
    # Reload systemd
    systemctl daemon-reload
    
    # Enable service
    systemctl enable "${SERVICE_NAME}.service"
    
    print_status "Service installed and enabled"
}

# Function to start UI
start_ui() {
    print_status "Starting UI service..."
    
    if ! systemctl is-active --quiet "${SERVICE_NAME}.service"; then
        systemctl start "${SERVICE_NAME}.service"
        sleep 2
        
        if systemctl is-active --quiet "${SERVICE_NAME}.service"; then
            print_status "UI service started successfully"
        else
            print_error "Failed to start UI service"
            systemctl status "${SERVICE_NAME}.service"
            exit 1
        fi
    else
        print_warning "UI service is already running"
    fi
}

# Function to stop UI
stop_ui() {
    print_status "Stopping UI service..."
    
    if systemctl is-active --quiet "${SERVICE_NAME}.service"; then
        systemctl stop "${SERVICE_NAME}.service"
        print_status "UI service stopped"
    else
        print_warning "UI service is not running"
    fi
}

# Function to restart UI
restart_ui() {
    print_status "Restarting UI service..."
    systemctl restart "${SERVICE_NAME}.service"
    sleep 2
    
    if systemctl is-active --quiet "${SERVICE_NAME}.service"; then
        print_status "UI service restarted successfully"
    else
        print_error "Failed to restart UI service"
        systemctl status "${SERVICE_NAME}.service"
        exit 1
    fi
}

# Function to show status
show_status() {
    print_status "UI service status:"
    systemctl status "${SERVICE_NAME}.service" --no-pager -l
    
    echo ""
    print_status "Checking if UI is accessible on port 5173..."
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:5173 | grep -q "200"; then
        print_status "UI is accessible at http://localhost:5173"
    else
        print_warning "UI is not accessible on port 5173"
    fi
}

# Function to enable service
enable_service() {
    print_status "Enabling UI service for auto-start..."
    systemctl enable "${SERVICE_NAME}.service"
    print_status "UI service will start automatically on boot"
}

# Function to disable service
disable_service() {
    print_status "Disabling UI service auto-start..."
    systemctl disable "${SERVICE_NAME}.service"
    print_status "UI service will not start automatically on boot"
}

# Function to show logs
show_logs() {
    print_status "Showing UI service logs:"
    journalctl -u "${SERVICE_NAME}.service" -f
}

# Main script logic
case "$1" in
    install)
        check_root
        install_service
        ;;
    start)
        start_ui
        ;;
    stop)
        stop_ui
        ;;
    restart)
        restart_ui
        ;;
    status)
        show_status
        ;;
    enable)
        check_root
        enable_service
        ;;
    disable)
        check_root
        disable_service
        ;;
    logs)
        show_logs
        ;;
    *)
        echo "Usage: $0 {install|start|stop|restart|status|enable|disable|logs}"
        echo ""
        echo "Commands:"
        echo "  install   - Install and enable the UI service (requires sudo)"
        echo "  start     - Start the UI service"
        echo "  stop      - Stop the UI service"
        echo "  restart   - Restart the UI service"
        echo "  status    - Show service status and check accessibility"
        echo "  enable    - Enable auto-start on boot (requires sudo)"
        echo "  disable   - Disable auto-start on boot (requires sudo)"
        echo "  logs      - Show service logs"
        exit 1
        ;;
esac 