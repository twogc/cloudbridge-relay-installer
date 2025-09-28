#!/bin/bash
# CloudBridge Appliance Management Script

set -e

# Configuration
APP_NAME="cloudbridge-appliance"
INSTALL_DIR="/opt/cloudbridge"
CONFIG_DIR="/etc/cloudbridge"
LOG_DIR="/var/log/cloudbridge"
DATA_DIR="/var/lib/cloudbridge"
SERVICE_NAME="cloudbridge-appliance"

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

# Show usage
show_usage() {
    echo "CloudBridge Appliance Management Script"
    echo
    echo "Usage: $0 <command> [options]"
    echo
    echo "Commands:"
    echo "  start       Start CloudBridge Appliance"
    echo "  stop        Stop CloudBridge Appliance"
    echo "  restart     Restart CloudBridge Appliance"
    echo "  status      Show status"
    echo "  logs        Show logs"
    echo "  health      Check health"
    echo "  metrics     Show metrics"
    echo "  config      Show configuration"
    echo "  backup      Create backup"
    echo "  restore     Restore from backup"
    echo "  update      Update appliance"
    echo "  uninstall  Uninstall appliance"
    echo
    echo "Options:"
    echo "  -f, --follow    Follow logs (for logs command)"
    echo "  -t, --tail      Number of lines to show (default: 100)"
    echo "  -h, --help      Show this help"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

# Start appliance
start_appliance() {
    log_info "Starting CloudBridge Appliance..."
    
    # Start systemd service
    systemctl start "$SERVICE_NAME"
    
    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    sleep 10
    
    # Check health
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log_success "CloudBridge Appliance started successfully"
    else
        log_warning "Health check failed. Check logs with: $0 logs"
    fi
}

# Stop appliance
stop_appliance() {
    log_info "Stopping CloudBridge Appliance..."
    
    # Stop systemd service
    systemctl stop "$SERVICE_NAME"
    
    log_success "CloudBridge Appliance stopped"
}

# Restart appliance
restart_appliance() {
    log_info "Restarting CloudBridge Appliance..."
    
    # Restart systemd service
    systemctl restart "$SERVICE_NAME"
    
    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    sleep 10
    
    # Check health
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log_success "CloudBridge Appliance restarted successfully"
    else
        log_warning "Health check failed. Check logs with: $0 logs"
    fi
}

# Show status
show_status() {
    log_info "CloudBridge Appliance Status:"
    echo
    
    # Systemd service status
    echo "📋 Service Status:"
    systemctl status "$SERVICE_NAME" --no-pager
    echo
    
    # Docker containers status
    echo "🐳 Docker Containers:"
    cd "$INSTALL_DIR"
    docker-compose -f docker-compose.appliance.yml ps
    echo
    
    # Health check
    echo "🏥 Health Check:"
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log_success "Appliance is healthy"
    else
        log_error "Appliance is not healthy"
    fi
    echo
    
    # Resource usage
    echo "📊 Resource Usage:"
    echo "CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}')%"
    echo "Memory: $(free -h | awk '/^Mem:/{print $3"/"$2}')"
    echo "Disk: $(df -h / | awk 'NR==2{print $3"/"$2" ("$5")"}')"
    echo
}

# Show logs
show_logs() {
    local follow=false
    local tail_lines=100
    
    # Parse options
    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--follow)
                follow=true
                shift
                ;;
            -t|--tail)
                tail_lines="$2"
                shift 2
                ;;
            *)
                shift
                ;;
        esac
    done
    
    log_info "Showing CloudBridge Appliance logs..."
    
    cd "$INSTALL_DIR"
    
    if [[ "$follow" == true ]]; then
        docker-compose -f docker-compose.appliance.yml logs -f --tail="$tail_lines"
    else
        docker-compose -f docker-compose.appliance.yml logs --tail="$tail_lines"
    fi
}

# Check health
check_health() {
    log_info "Checking CloudBridge Appliance health..."
    
    # Check if service is running
    if ! systemctl is-active --quiet "$SERVICE_NAME"; then
        log_error "Service is not running"
        return 1
    fi
    
    # Check HTTP health endpoint
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        log_success "HTTP health check passed"
    else
        log_error "HTTP health check failed"
        return 1
    fi
    
    # Check gRPC endpoint
    if nc -z localhost 8444; then
        log_success "gRPC endpoint is accessible"
    else
        log_error "gRPC endpoint is not accessible"
        return 1
    fi
    
    # Check WireGuard
    if nc -z localhost 19302; then
        log_success "WireGuard endpoint is accessible"
    else
        log_error "WireGuard endpoint is not accessible"
        return 1
    fi
    
    # Check STUN/TURN
    if nc -z localhost 3478; then
        log_success "STUN/TURN endpoint is accessible"
    else
        log_error "STUN/TURN endpoint is not accessible"
        return 1
    fi
    
    # Check QUIC
    if nc -z localhost 5553; then
        log_success "QUIC endpoint is accessible"
    else
        log_error "QUIC endpoint is not accessible"
        return 1
    fi
    
    log_success "All health checks passed"
}

# Show metrics
show_metrics() {
    log_info "CloudBridge Appliance Metrics:"
    echo
    
    # Get metrics from Prometheus endpoint
    if curl -f http://localhost:8080/metrics > /dev/null 2>&1; then
        echo "📊 Prometheus Metrics:"
        curl -s http://localhost:8080/metrics | grep -E "(appliance_|cloudbridge_)" | head -20
        echo
    else
        log_warning "Metrics endpoint not accessible"
    fi
    
    # Get status from API
    if curl -f http://localhost:8080/api/v1/appliance/status > /dev/null 2>&1; then
        echo "📈 Appliance Status:"
        curl -s http://localhost:8080/api/v1/appliance/status | jq '.' 2>/dev/null || echo "JSON parsing failed"
        echo
    else
        log_warning "Status API not accessible"
    fi
}

# Show configuration
show_config() {
    log_info "CloudBridge Appliance Configuration:"
    echo
    
    # Show main configuration
    if [[ -f "$CONFIG_DIR/appliance-config.yaml" ]]; then
        echo "📋 Main Configuration:"
        cat "$CONFIG_DIR/appliance-config.yaml"
        echo
    else
        log_warning "Configuration file not found"
    fi
    
    # Show Docker Compose configuration
    if [[ -f "$INSTALL_DIR/docker-compose.appliance.yml" ]]; then
        echo "🐳 Docker Compose Configuration:"
        cat "$INSTALL_DIR/docker-compose.appliance.yml"
        echo
    else
        log_warning "Docker Compose file not found"
    fi
}

# Create backup
create_backup() {
    local backup_dir="/var/backups/cloudbridge"
    local backup_file="cloudbridge-backup-$(date +%Y%m%d-%H%M%S).tar.gz"
    
    log_info "Creating CloudBridge Appliance backup..."
    
    # Create backup directory
    mkdir -p "$backup_dir"
    
    # Create backup
    tar -czf "$backup_dir/$backup_file" \
        -C / "$CONFIG_DIR" \
        -C / "$DATA_DIR" \
        -C / "$LOG_DIR" \
        -C / "$INSTALL_DIR" \
        2>/dev/null || true
    
    # Set permissions
    chown -R root:root "$backup_dir"
    chmod 600 "$backup_dir/$backup_file"
    
    log_success "Backup created: $backup_dir/$backup_file"
    echo "Backup size: $(du -h "$backup_dir/$backup_file" | cut -f1)"
}

# Restore from backup
restore_backup() {
    local backup_file="$1"
    
    if [[ -z "$backup_file" ]]; then
        log_error "Backup file not specified"
        echo "Usage: $0 restore <backup_file>"
        exit 1
    fi
    
    if [[ ! -f "$backup_file" ]]; then
        log_error "Backup file not found: $backup_file"
        exit 1
    fi
    
    log_info "Restoring CloudBridge Appliance from backup..."
    
    # Stop appliance
    systemctl stop "$SERVICE_NAME"
    
    # Restore from backup
    tar -xzf "$backup_file" -C /
    
    # Start appliance
    systemctl start "$SERVICE_NAME"
    
    log_success "Restore completed"
}

# Update appliance
update_appliance() {
    local new_version="$1"
    
    if [[ -z "$new_version" ]]; then
        log_error "Version not specified"
        echo "Usage: $0 update <version>"
        exit 1
    fi
    
    log_info "Updating CloudBridge Appliance to version $new_version..."
    
    # Stop appliance
    systemctl stop "$SERVICE_NAME"
    
    # Create backup
    create_backup
    
    # Update Docker images
    cd "$INSTALL_DIR"
    docker-compose -f docker-compose.appliance.yml pull
    
    # Start appliance
    systemctl start "$SERVICE_NAME"
    
    log_success "Update completed"
}

# Uninstall appliance
uninstall_appliance() {
    log_warning "This will completely remove CloudBridge Appliance and all data!"
    read -p "Are you sure? (yes/no): " confirm
    
    if [[ "$confirm" != "yes" ]]; then
        log_info "Uninstall cancelled"
        exit 0
    fi
    
    log_info "Uninstalling CloudBridge Appliance..."
    
    # Stop and disable service
    systemctl stop "$SERVICE_NAME"
    systemctl disable "$SERVICE_NAME"
    
    # Remove systemd service
    rm -f /etc/systemd/system/cloudbridge-appliance.service
    systemctl daemon-reload
    
    # Remove Docker containers and images
    cd "$INSTALL_DIR"
    docker-compose -f docker-compose.appliance.yml down -v
    docker rmi cloudbridge-appliance:latest 2>/dev/null || true
    
    # Remove directories
    rm -rf "$INSTALL_DIR"
    rm -rf "$CONFIG_DIR"
    rm -rf "$LOG_DIR"
    rm -rf "$DATA_DIR"
    
    # Remove user
    userdel cloudbridge 2>/dev/null || true
    
    log_success "CloudBridge Appliance uninstalled"
}

# Main function
main() {
    local command="$1"
    shift
    
    case "$command" in
        start)
            check_root
            start_appliance
            ;;
        stop)
            check_root
            stop_appliance
            ;;
        restart)
            check_root
            restart_appliance
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$@"
            ;;
        health)
            check_health
            ;;
        metrics)
            show_metrics
            ;;
        config)
            show_config
            ;;
        backup)
            check_root
            create_backup
            ;;
        restore)
            check_root
            restore_backup "$@"
            ;;
        update)
            check_root
            update_appliance "$@"
            ;;
        uninstall)
            check_root
            uninstall_appliance
            ;;
        -h|--help|help)
            show_usage
            ;;
        *)
            log_error "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
