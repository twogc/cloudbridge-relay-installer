#!/bin/bash
set -euo pipefail

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода с цветом
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}[HEADER]${NC} $1"
}

# Функция для остановки всех port-forward процессов
stop_port_forwards() {
    print_status "Stopping all port-forward processes..."
    
    # Останавливаем процессы port-forward
    pkill -f "kubectl.*port-forward" || true
    
    # Ждем немного
    sleep 2
    
    print_status "All port-forward processes stopped"
}

# Функция для проверки доступности порта
check_port() {
    local port=$1
    if netstat -tuln 2>/dev/null | grep -q ":$port "; then
        return 0
    else
        return 1
    fi
}

# Функция для запуска port-forward
start_port_forward() {
    local namespace=$1
    local service=$2
    local local_port=$3
    local remote_port=$4
    local description=$5
    
    if check_port $local_port; then
        print_warning "Port $local_port is already in use"
        return 1
    fi
    
    print_status "Starting port-forward for $description..."
    print_status "Local: http://localhost:$local_port"
    print_status "Remote: $namespace/$service:$remote_port"
    
    kubectl -n $namespace port-forward service/$service $local_port:$remote_port &
    
    # Ждем немного и проверяем
    sleep 3
    if check_port $local_port; then
        print_status "Port-forward for $description started successfully"
        return 0
    else
        print_error "❌ Failed to start port-forward for $description"
        return 1
    fi
}

# Основная функция
main() {
    case "${1:-help}" in
        "start")
            print_header "Starting Zitadel port-forwards..."
            
            # Останавливаем существующие
            stop_port_forwards
            
            # Запускаем Zitadel UI
            start_port_forward "zitadel" "zitadel-login" "3000" "3000" "Zitadel UI"
            
            # Запускаем Zitadel API
            start_port_forward "zitadel" "zitadel" "8080" "8080" "Zitadel API"
            
            print_header "Port-forwards started!"
            echo ""
            print_status "🌐 Zitadel UI: http://zitadel.local:8080/ui/login"
            print_status "🔧 Zitadel API: http://zitadel.local:8080"
            echo ""
            print_status "Login credentials:"
            print_status "  Email: admin@local.test"
            print_status "  Password: jrcCaigzmqzHpRcxo61RDiS+xejxX4/K"
            echo ""
            print_status "To stop port-forwards, run: $0 stop"
            ;;
            
        "stop")
            stop_port_forwards
            ;;
            
        "status")
            print_header "Port-forward status:"
            echo ""
            
            if check_port 3000; then
                print_status "Port 3000 (Zitadel UI): ACTIVE"
            else
                print_warning "❌ Port 3000 (Zitadel UI): INACTIVE"
            fi
            
            if check_port 8080; then
                print_status "Port 8080 (Zitadel API): ACTIVE"
            else
                print_warning "❌ Port 8080 (Zitadel API): INACTIVE"
            fi
            
            if check_port 9091; then
                print_status "Port 9091 (CloudBridge Relay): ACTIVE"
            else
                print_warning "❌ Port 9091 (CloudBridge Relay): INACTIVE"
            fi
            ;;
            
        "cloudbridge")
            print_header "Starting CloudBridge Relay port-forwards..."
            
            start_port_forward "cloudbridge" "cloudbridge-relay-external" "9091" "9091" "CloudBridge Relay API"
            
            print_header "CloudBridge port-forwards started!"
            echo ""
            print_status "🚀 CloudBridge Relay API: http://localhost:9091"
            print_status "📊 Health check: http://localhost:9091/health"
            print_status "📈 Metrics: http://localhost:9091/metrics"
            ;;
            
        "all")
            print_header "Starting all port-forwards..."
            
            # Останавливаем существующие
            stop_port_forwards
            
            # Запускаем все сервисы
            start_port_forward "zitadel" "zitadel-login" "3000" "3000" "Zitadel UI"
            start_port_forward "zitadel" "zitadel" "8080" "8080" "Zitadel API"
            start_port_forward "cloudbridge" "cloudbridge-relay-external" "9091" "9091" "CloudBridge Relay API"
            
            print_header "All port-forwards started!"
            echo ""
            print_status "🌐 Zitadel UI: http://zitadel.local:8080/ui/login"
            print_status "🔧 Zitadel API: http://zitadel.local:8080"
            print_status "🚀 CloudBridge Relay API: http://localhost:9091"
            ;;
            
        "help"|*)
            echo "Usage: $0 {start|stop|status|cloudbridge|all|help}"
            echo ""
            echo "Commands:"
            echo "  start       - Start Zitadel port-forwards (UI + API)"
            echo "  stop        - Stop all port-forwards"
            echo "  status      - Show status of all port-forwards"
            echo "  cloudbridge - Start CloudBridge Relay port-forwards"
            echo "  all         - Start all port-forwards (Zitadel + CloudBridge)"
            echo "  help        - Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 start        # Start Zitadel UI and API"
            echo "  $0 cloudbridge  # Start CloudBridge Relay API"
            echo "  $0 all          # Start everything"
            echo "  $0 status       # Check what's running"
            echo "  $0 stop         # Stop everything"
            ;;
    esac
}

# Запуск основной функции
main "$@"
