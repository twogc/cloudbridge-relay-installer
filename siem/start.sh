#!/bin/bash

# CloudBridge SIEM System Startup Script
# Version: 1.0.0

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    print_success "Docker is running"
}

# Function to check if Docker Compose is available
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose and try again."
        exit 1
    fi
    print_success "Docker Compose is available"
}

# Function to check system requirements
check_system_requirements() {
    print_status "Checking system requirements..."
    
    # Check available memory
    local mem_gb=$(free -g | awk '/^Mem:/{print $2}')
    if [ "$mem_gb" -lt 4 ]; then
        print_warning "System has less than 4GB RAM. SIEM may not perform optimally."
    else
        print_success "System has ${mem_gb}GB RAM"
    fi
    
    # Check available disk space
    local disk_gb=$(df -BG . | awk 'NR==2 {print $4}' | sed 's/G//')
    if [ "$disk_gb" -lt 20 ]; then
        print_warning "Less than 20GB free disk space available. Consider freeing up space."
    else
        print_success "System has ${disk_gb}GB free disk space"
    fi
}

# Function to create .env file if it doesn't exist
create_env_file() {
    if [ ! -f .env ]; then
        print_status "Creating .env file with default settings..."
        cat > .env << EOF
# Database
DATABASE_URL=postgresql://siem:password@localhost:5432/siem
REDIS_URL=redis://localhost:6379/0
ELASTICSEARCH_URL=http://localhost:9200

# Security
SECRET_KEY=$(openssl rand -hex 32)
JWT_SECRET=$(openssl rand -hex 32)

# External APIs
CLOUDFLARE_AGENT_URL=https://cloudbridge-ai-monitoring-agent.2gc.workers.dev
ML_API_URL=http://localhost:8000/detect
RELAY_METRICS_URL=http://localhost:9090/metrics

# Notifications (optional - uncomment and configure as needed)
# EMAIL_SMTP_SERVER=smtp.gmail.com
# EMAIL_SMTP_PORT=587
# EMAIL_USERNAME=your-email@gmail.com
# EMAIL_PASSWORD=your-app-password
# EMAIL_FROM=siem@yourcompany.com
# EMAIL_TO=admin@yourcompany.com

# TELEGRAM_BOT_TOKEN=your-bot-token
# TELEGRAM_CHAT_ID=your-chat-id

# SLACK_WEBHOOK_URL=https://hooks.slack.com/services/your/webhook/url
EOF
        print_success ".env file created"
    else
        print_status ".env file already exists"
    fi
}

# Function to increase system limits for Elasticsearch
setup_elasticsearch_limits() {
    print_status "Setting up Elasticsearch system limits..."
    
    # Check if vm.max_map_count is already set
    local current_limit=$(sysctl -n vm.max_map_count)
    if [ "$current_limit" -lt 262144 ]; then
        print_status "Increasing vm.max_map_count to 262144..."
        echo "vm.max_map_count=262144" | sudo tee -a /etc/sysctl.conf
        sudo sysctl -p
        print_success "System limits updated"
    else
        print_success "System limits are already properly configured"
    fi
}

# Function to start SIEM services
start_siem() {
    print_status "Starting SIEM services..."
    
    # Pull latest images
    print_status "Pulling Docker images..."
    docker-compose pull
    
    # Start services
    print_status "Starting services..."
    docker-compose up -d
    
    # Wait for services to be ready
    print_status "Waiting for services to be ready..."
    sleep 30
    
    # Check service health
    check_service_health
}

# Function to check service health
check_service_health() {
    print_status "Checking service health..."
    
    local services=("postgres" "redis" "elasticsearch" "siem-api" "siem-frontend")
    local healthy=true
    
    for service in "${services[@]}"; do
        if docker-compose ps "$service" | grep -q "Up"; then
            print_success "$service is running"
        else
            print_error "$service is not running"
            healthy=false
        fi
    done
    
    if [ "$healthy" = true ]; then
        print_success "All services are running"
    else
        print_warning "Some services may not be running properly. Check logs with: docker-compose logs"
    fi
}

# Function to initialize database
initialize_database() {
    print_status "Initializing database..."
    
    # Wait for PostgreSQL to be ready
    print_status "Waiting for PostgreSQL to be ready..."
    sleep 10
    
    # Run database migrations
    if docker-compose exec -T siem-api alembic upgrade head 2>/dev/null; then
        print_success "Database initialized"
    else
        print_warning "Database initialization failed. This may be normal if the service is still starting up."
    fi
}

# Function to show access information
show_access_info() {
    echo
    print_success "SIEM System is starting up!"
    echo
    echo "Access URLs:"
    echo "  Frontend:     http://localhost:3000"
    echo "  API Docs:     http://localhost:8000/docs"
    echo "  Grafana:      http://localhost:3001 (admin/admin)"
    echo "  Prometheus:   http://localhost:9090"
    echo
    echo "Useful commands:"
    echo "  View logs:    docker-compose logs -f"
    echo "  Stop system:  docker-compose down"
    echo "  Restart:      docker-compose restart"
    echo
    print_warning "Note: Services may take a few minutes to fully start up."
    echo
}

# Function to show help
show_help() {
    echo "CloudBridge SIEM System Startup Script"
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -s, --stop     Stop SIEM services"
    echo "  -r, --restart  Restart SIEM services"
    echo "  -l, --logs     Show service logs"
    echo "  -c, --clean    Clean up containers and volumes"
    echo
    echo "Examples:"
    echo "  $0              # Start SIEM system"
    echo "  $0 --stop       # Stop SIEM system"
    echo "  $0 --logs       # Show logs"
    echo
}

# Function to stop services
stop_services() {
    print_status "Stopping SIEM services..."
    docker-compose down
    print_success "Services stopped"
}

# Function to restart services
restart_services() {
    print_status "Restarting SIEM services..."
    docker-compose restart
    print_success "Services restarted"
}

# Function to show logs
show_logs() {
    print_status "Showing service logs..."
    docker-compose logs -f
}

# Function to clean up
clean_up() {
    print_warning "This will remove all containers and volumes. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        print_status "Cleaning up containers and volumes..."
        docker-compose down -v
        print_success "Cleanup completed"
    else
        print_status "Cleanup cancelled"
    fi
}

# Main script logic
main() {
    echo "=========================================="
    echo "  CloudBridge SIEM System Startup Script"
    echo "=========================================="
    echo
    
    # Parse command line arguments
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -s|--stop)
            stop_services
            exit 0
            ;;
        -r|--restart)
            restart_services
            exit 0
            ;;
        -l|--logs)
            show_logs
            exit 0
            ;;
        -c|--clean)
            clean_up
            exit 0
            ;;
        "")
            # Default action: start services
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
    
    # Start SIEM system
    print_status "Starting CloudBridge SIEM System..."
    
    # Check prerequisites
    check_docker
    check_docker_compose
    check_system_requirements
    
    # Setup environment
    create_env_file
    setup_elasticsearch_limits
    
    # Start services
    start_siem
    
    # Initialize database
    initialize_database
    
    # Show access information
    show_access_info
}

# Run main function
main "$@" 