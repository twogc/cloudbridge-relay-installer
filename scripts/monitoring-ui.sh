#!/bin/bash

# CloudBridge Relay - Monitoring UI Access Script
# Скрипт для доступа к UI интерфейсам мониторинга

set -e

NAMESPACE_MONITORING="cloudbridge-monitoring"
NAMESPACE_CLOUDBRIDGE="cloudbridge"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Функция для вывода заголовка
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  CloudBridge Relay Monitoring UI${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
}

# Функция для вывода информации о сервисе
print_service_info() {
    local service_name=$1
    local port=$2
    local description=$3
    local url=$4
    
    echo -e "${GREEN}📊 $service_name${NC}"
    echo -e "   ${CYAN}Описание:${NC} $description"
    echo -e "   ${CYAN}Порт:${NC} $port"
    echo -e "   ${CYAN}URL:${NC} $url"
    echo ""
}

# Функция для проверки статуса подов
check_pods_status() {
    echo -e "${YELLOW}🔍 Проверка статуса подов...${NC}"
    echo ""
    
    # Проверка подов мониторинга
    echo -e "${BLUE}📊 Мониторинг (cloudbridge-monitoring):${NC}"
    kubectl get pods -n $NAMESPACE_MONITORING --no-headers | while read line; do
        pod_name=$(echo $line | awk '{print $1}')
        status=$(echo $line | awk '{print $3}')
        if [ "$status" = "Running" ]; then
            echo -e "   $pod_name - $status"
        else
            echo -e "   ❌ $pod_name - $status"
        fi
    done
    echo ""
    
    # Проверка подов CloudBridge
    echo -e "${BLUE}🚀 CloudBridge (cloudbridge):${NC}"
    kubectl get pods -n $NAMESPACE_CLOUDBRIDGE --no-headers | while read line; do
        pod_name=$(echo $line | awk '{print $1}')
        status=$(echo $line | awk '{print $3}')
        if [ "$status" = "Running" ]; then
            echo -e "   $pod_name - $status"
        else
            echo -e "   ❌ $pod_name - $status"
        fi
    done
    echo ""
}

# Функция для запуска port-forward
start_port_forward() {
    local service_name=$1
    local namespace=$2
    local local_port=$3
    local remote_port=$4
    local description=$5
    
    echo -e "${YELLOW}🚀 Запуск port-forward для $service_name...${NC}"
    
    # Проверяем, не запущен ли уже port-forward
    if pgrep -f "kubectl port-forward.*$service_name.*$local_port:$remote_port" > /dev/null; then
        echo -e "${YELLOW}  Port-forward для $service_name уже запущен${NC}"
        return 0
    fi
    
    # Запускаем port-forward в фоне
    kubectl port-forward -n $namespace svc/$service_name $local_port:$remote_port > /dev/null 2>&1 &
    local pid=$!
    
    # Ждем немного для инициализации
    sleep 2
    
    # Проверяем, что процесс запустился
    if kill -0 $pid 2>/dev/null; then
        echo -e "${GREEN}Port-forward для $service_name запущен (PID: $pid)${NC}"
        echo -e "   ${CYAN}Доступ:${NC} http://localhost:$local_port"
        echo -e "   ${CYAN}Описание:${NC} $description"
        echo ""
        return 0
    else
        echo -e "${RED}❌ Ошибка запуска port-forward для $service_name${NC}"
        return 1
    fi
}

# Функция для остановки всех port-forward
stop_all_port_forwards() {
    echo -e "${YELLOW}🛑 Остановка всех port-forward...${NC}"
    
    # Находим и останавливаем все процессы kubectl port-forward
    pids=$(pgrep -f "kubectl port-forward" || true)
    
    if [ -n "$pids" ]; then
        echo "$pids" | xargs kill -TERM 2>/dev/null || true
        sleep 2
        echo "$pids" | xargs kill -KILL 2>/dev/null || true
        echo -e "${GREEN}Все port-forward остановлены${NC}"
    else
        echo -e "${YELLOW}ℹ️  Нет активных port-forward${NC}"
    fi
    
    # Останавливаем Cadence Web UI
    echo -e "${YELLOW}🛑 Остановка Cadence Web UI...${NC}"
    cadence_web_pids=$(pgrep -f "next dev.*8088" || true)
    
    if [ -n "$cadence_web_pids" ]; then
        echo "$cadence_web_pids" | xargs kill -TERM 2>/dev/null || true
        sleep 2
        echo "$cadence_web_pids" | xargs kill -KILL 2>/dev/null || true
        echo -e "${GREEN}Cadence Web UI остановлен${NC}"
    else
        echo -e "${YELLOW}ℹ️  Cadence Web UI не запущен${NC}"
    fi
    
    # NodePort сервис для Cadence остается активным
    echo -e "${YELLOW}ℹ️  NodePort сервис для Cadence остается активным${NC}"
    
    echo ""
}

# Функция для показа статуса port-forward
show_port_forward_status() {
    echo -e "${YELLOW}📊 Статус port-forward:${NC}"
    
    local pids=$(pgrep -f "kubectl port-forward" || true)
    
    if [ -n "$pids" ]; then
        echo "$pids" | while read pid; do
            local cmd=$(ps -p $pid -o args= 2>/dev/null || echo "Процесс не найден")
            echo -e "   ${GREEN}PID $pid:${NC} $cmd"
        done
    else
        echo -e "   ${YELLOW}ℹ️  Нет активных port-forward${NC}"
    fi
    
    echo -e "${YELLOW}📊 Статус Cadence Web UI:${NC}"
    local cadence_web_pids=$(pgrep -f "next dev.*8088" || true)
    
    if [ -n "$cadence_web_pids" ]; then
        echo "$cadence_web_pids" | while read pid; do
            local cmd=$(ps -p $pid -o args= 2>/dev/null || echo "Процесс не найден")
            echo -e "   ${GREEN}PID $pid:${NC} $cmd"
        done
    else
        echo -e "   ${YELLOW}ℹ️  Cadence Web UI не запущен${NC}"
    fi
    
    echo -e "${YELLOW}📊 Статус NodePort для Cadence:${NC}"
    if kubectl get svc -n cloudbridge cadence-nodeport > /dev/null 2>&1; then
        local nodeport=$(kubectl get svc -n cloudbridge cadence-nodeport -o jsonpath='{.spec.ports[0].nodePort}')
        local minikube_ip=$(minikube ip 2>/dev/null || echo "unknown")
        echo -e "   ${GREEN}NodePort активен: $minikube_ip:$nodeport${NC}"
    else
        echo -e "   ${YELLOW}ℹ️  NodePort сервис не найден${NC}"
    fi
    
    echo ""
}

# Функция для открытия браузера
open_browser() {
    local url=$1
    local description=$2
    
    echo -e "${YELLOW}🌐 Открытие $description в браузере...${NC}"
    
    if command -v xdg-open > /dev/null; then
        xdg-open "$url" 2>/dev/null &
        echo -e "${GREEN}Браузер открыт для $description${NC}"
    elif command -v open > /dev/null; then
        open "$url" 2>/dev/null &
        echo -e "${GREEN}Браузер открыт для $description${NC}"
    else
        echo -e "${YELLOW}  Автоматическое открытие браузера не поддерживается${NC}"
        echo -e "   ${CYAN}Откройте вручную:${NC} $url"
    fi
    echo ""
}

# Основная функция
main() {
    print_header
    
    case "${1:-}" in
        "start"|"")
            echo -e "${YELLOW}🚀 Запуск всех UI интерфейсов...${NC}"
            echo ""
            
            check_pods_status
            
            # Запускаем port-forward для всех сервисов
            start_port_forward "grafana" "$NAMESPACE_MONITORING" 3000 3000 "Grafana Dashboard"
            start_port_forward "prometheus-optimized-service" "$NAMESPACE_MONITORING" 9090 9090 "Prometheus Metrics"
            start_port_forward "cassandra-service" "$NAMESPACE_CLOUDBRIDGE" 9500 9500 "Cassandra Metrics"
            start_port_forward "cloudbridge-relay" "$NAMESPACE_CLOUDBRIDGE" 8084 8084 "CloudBridge Relay Metrics"
            
            echo -e "${YELLOW}🚀 Проверка Cadence Web UI в кластере...${NC}"
            echo -e "${CYAN}Cadence Web UI доступен на: http://192.168.49.2:30088${NC}"
            
            # Проверяем развертывание Cadence Web UI в кластере
            if kubectl get deployment -n cloudbridge cadence-web > /dev/null 2>&1; then
                echo -e "${GREEN}Cadence Web UI развернут в кластере${NC}"
                
                # Проверяем статус пода
                local cadence_web_pod=$(kubectl get pods -n cloudbridge -l app.kubernetes.io/name=cadence-web -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
                if [ -n "$cadence_web_pod" ]; then
                    local pod_status=$(kubectl get pod -n cloudbridge $cadence_web_pod -o jsonpath='{.status.phase}' 2>/dev/null || echo "Unknown")
                    if [ "$pod_status" = "Running" ]; then
                        echo -e "${GREEN}Cadence Web UI pod работает: $cadence_web_pod${NC}"
                    else
                        echo -e "${YELLOW}  Cadence Web UI pod не в состоянии Running: $pod_status${NC}"
                    fi
                fi
            else
                echo -e "${YELLOW}  Cadence Web UI не развернут, разворачиваем...${NC}"
                kubectl apply -f k8s/deployments/cadence-web-deployment.yaml > /dev/null 2>&1
                kubectl apply -f k8s/services/cadence-web-service.yaml > /dev/null 2>&1
                sleep 10
                echo -e "${GREEN}Cadence Web UI развернут${NC}"
            fi
            
            echo -e "${GREEN}🎉 Все UI интерфейсы запущены!${NC}"
            echo ""
            echo -e "${BLUE}Доступные интерфейсы:${NC}"
            print_service_info "Cadence Web UI" "30088" "Полноценный веб-интерфейс для Cadence" "http://192.168.49.2:30088"
            print_service_info "Grafana" "3000" "Дашборды и визуализация (включая Cadence)" "http://localhost:3000"
            print_service_info "Prometheus" "9090" "Метрики и алерты (включая Cadence)" "http://localhost:9090"
            print_service_info "Cassandra Metrics" "9500" "База данных метрики" "http://localhost:9500"
            print_service_info "CloudBridge Relay" "8084" "Основное приложение" "http://localhost:8084"
            echo -e "${YELLOW}Дополнительные инструменты:${NC}"
            echo -e "   ${CYAN}CLI:${NC} ./tools/cadence-web/cadence-monitor.sh"
            echo -e "   ${CYAN}Логи:${NC} kubectl logs -n cloudbridge cadence-xxx"
            
            echo -e "${YELLOW}Для остановки используйте: $0 stop${NC}"
            ;;
            
        "stop")
            stop_all_port_forwards
            ;;
            
        "status")
            check_pods_status
            show_port_forward_status
            ;;
            
        "open")
            echo -e "${YELLOW}🌐 Открытие всех интерфейсов в браузере...${NC}"
            echo ""
            
            # Запускаем port-forward если не запущены
            start_port_forward "grafana" "$NAMESPACE_MONITORING" 3000 3000 "Grafana Dashboard" > /dev/null
            start_port_forward "prometheus-optimized-service" "$NAMESPACE_MONITORING" 9090 9090 "Prometheus Metrics" > /dev/null
            
            # Проверяем NodePort сервис для Cadence
            if kubectl get svc -n cloudbridge cadence-nodeport > /dev/null 2>&1; then
                echo -e "${GREEN}NodePort сервис для Cadence найден${NC}"
            else
                kubectl apply -f k8s/services/cadence-nodeport.yaml > /dev/null 2>&1
                sleep 3
            fi
            
            # Запускаем Cadence Web UI если не запущен
            if ! pgrep -f "next dev.*8088" > /dev/null; then
                cd tools/cadence-web
                ./start-cadence-web.sh dev > /dev/null 2>&1 &
                cd ../..
                sleep 5
            fi
            
            # Открываем браузеры
            open_browser "http://192.168.49.2:30088" "Cadence Web UI"
            open_browser "http://localhost:3000" "Grafana Dashboard"
            open_browser "http://localhost:9090" "Prometheus Metrics"
            ;;
            
        "grafana")
            start_port_forward "grafana" "$NAMESPACE_MONITORING" 3000 3000 "Grafana Dashboard"
            open_browser "http://localhost:3000" "Grafana Dashboard"
            ;;
            
        "prometheus")
            start_port_forward "prometheus-optimized-service" "$NAMESPACE_MONITORING" 9090 9090 "Prometheus Metrics"
            open_browser "http://localhost:9090" "Prometheus Metrics"
            ;;
            
        "cadence")
            echo -e "${YELLOW}🚀 Проверка Cadence Web UI в кластере...${NC}"
            echo -e "${CYAN}Cadence Web UI доступен на: http://192.168.49.2:30088${NC}"
            echo ""
            
            # Проверяем развертывание Cadence Web UI в кластере
            if kubectl get deployment -n cloudbridge cadence-web > /dev/null 2>&1; then
                echo -e "${GREEN}Cadence Web UI развернут в кластере${NC}"
                
                # Проверяем статус пода
                local cadence_web_pod=$(kubectl get pods -n cloudbridge -l app.kubernetes.io/name=cadence-web -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
                if [ -n "$cadence_web_pod" ]; then
                    local pod_status=$(kubectl get pod -n cloudbridge $cadence_web_pod -o jsonpath='{.status.phase}' 2>/dev/null || echo "Unknown")
                    if [ "$pod_status" = "Running" ]; then
                        echo -e "${GREEN}Cadence Web UI pod работает: $cadence_web_pod${NC}"
                        echo -e "${CYAN}Доступ: http://192.168.49.2:30088${NC}"
                    else
                        echo -e "${YELLOW}  Cadence Web UI pod не в состоянии Running: $pod_status${NC}"
                    fi
                fi
            else
                echo -e "${YELLOW}  Cadence Web UI не развернут, разворачиваем...${NC}"
                kubectl apply -f k8s/deployments/cadence-web-deployment.yaml > /dev/null 2>&1
                kubectl apply -f k8s/services/cadence-web-service.yaml > /dev/null 2>&1
                sleep 10
                echo -e "${GREEN}Cadence Web UI развернут${NC}"
                echo -e "${CYAN}Доступ: http://192.168.49.2:30088${NC}"
            fi
            ;;
            
        "cassandra")
            start_port_forward "cassandra-service" "$NAMESPACE_CLOUDBRIDGE" 9500 9500 "Cassandra Metrics"
            open_browser "http://localhost:9500" "Cassandra Metrics"
            ;;
            
        "relay")
            start_port_forward "cloudbridge-relay" "$NAMESPACE_CLOUDBRIDGE" 8084 8084 "CloudBridge Relay Metrics"
            open_browser "http://localhost:8084" "CloudBridge Relay Metrics"
            ;;
            
        "help"|"-h"|"--help")
            echo -e "${BLUE}Использование: $0 [команда]${NC}"
            echo ""
            echo -e "${CYAN}Команды:${NC}"
            echo -e "  ${GREEN}start${NC}     - Запустить все UI интерфейсы (по умолчанию)"
            echo -e "  ${GREEN}stop${NC}      - Остановить все port-forward"
            echo -e "  ${GREEN}status${NC}    - Показать статус подов и port-forward"
            echo -e "  ${GREEN}open${NC}      - Открыть все интерфейсы в браузере"
            echo -e "  ${GREEN}grafana${NC}   - Запустить только Grafana"
            echo -e "  ${GREEN}prometheus${NC} - Запустить только Prometheus"
            echo -e "  ${GREEN}cadence${NC}   - Запустить только Cadence Web UI"
            echo -e "  ${GREEN}cassandra${NC} - Запустить только Cassandra Metrics"
            echo -e "  ${GREEN}relay${NC}     - Запустить только CloudBridge Relay"
            echo -e "  ${GREEN}help${NC}      - Показать эту справку"
            echo ""
            echo -e "${CYAN}Примеры:${NC}"
            echo -e "  $0                    # Запустить все интерфейсы"
            echo -e "  $0 start              # Запустить все интерфейсы"
            echo -e "  $0 open               # Открыть все в браузере"
            echo -e "  $0 grafana            # Только Grafana"
            echo -e "  $0 stop               # Остановить все"
            ;;
            
        *)
            echo -e "${RED}❌ Неизвестная команда: $1${NC}"
            echo -e "${YELLOW}Используйте '$0 help' для справки${NC}"
            exit 1
            ;;
    esac
}

# Обработка сигналов для корректного завершения
trap 'echo -e "\n${YELLOW}🛑 Получен сигнал завершения...${NC}"; stop_all_port_forwards; exit 0' INT TERM

# Запуск основной функции
main "$@"
