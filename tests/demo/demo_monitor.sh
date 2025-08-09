#!/bin/bash

# Демонстрационный скрипт мониторинга P2P Mesh с Prometheus
# Отслеживает все события и показывает работу системы

echo "🔍 ДЕМО: Мониторинг P2P Mesh с Prometheus"
echo "========================================="
echo ""

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Конфигурация Prometheus
PROMETHEUS_URL="http://localhost:9090"
GRAFANA_URL="http://localhost:3000"

# Проверка relay сервера
check_relay() {
    echo -e "${BLUE}📊 Проверка Relay Server:${NC}"
    if docker ps | grep -q cloudbridge-relay; then
        echo -e "${GREEN}✅ Relay сервер запущен${NC}"
    else
        echo -e "${RED}❌ Relay сервер не найден${NC}"
        return 1
    fi
    
    if netstat -tuln | grep -q ":51820"; then
        echo -e "${GREEN}✅ Порт 51820/udp открыт${NC}"
    else
        echo -e "${RED}❌ Порт 51820/udp не найден${NC}"
    fi
    echo ""
}

# Проверка Prometheus
check_prometheus() {
    echo -e "${BLUE}📊 Проверка Prometheus:${NC}"
    
    # Проверяем доступность Prometheus
    if curl -s "$PROMETHEUS_URL/api/v1/status/config" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Prometheus доступен${NC}"
    else
        echo -e "${RED}❌ Prometheus недоступен${NC}"
        return 1
    fi
    
    # Проверяем доступность Grafana
    if curl -s "$GRAFANA_URL/api/health" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Grafana доступен${NC}"
    else
        echo -e "${YELLOW}⚠️  Grafana недоступен${NC}"
    fi
    echo ""
}

# Получение метрик из Prometheus
get_prometheus_metrics() {
    local query="$1"
    local result=$(curl -s "$PROMETHEUS_URL/api/v1/query?query=$query" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "$result" | jq -r '.data.result[]?.value[1] // "N/A"' 2>/dev/null
    else
        echo "N/A"
    fi
}

# Мониторинг UDP трафика
monitor_udp() {
    echo -e "${YELLOW}🌐 Мониторинг UDP трафика на порту 51820:${NC}"
    echo "Нажмите Ctrl+C для остановки"
    echo "----------------------------------------"
    
    sudo tcpdump -i any -n udp port 51820 -v | while read line; do
        timestamp=$(date '+%H:%M:%S')
        echo -e "${CYAN}[$timestamp]${NC} $line"
    done
}

# Мониторинг логов
monitor_logs() {
    echo -e "${YELLOW}📋 Мониторинг логов P2P mesh:${NC}"
    echo "Нажмите Ctrl+C для остановки"
    echo "----------------------------------------"
    
    docker logs -f cloudbridge-relay | grep -E "(p2p|wireguard|mesh|peer|discovery)" --color=always | while read line; do
        timestamp=$(date '+%H:%M:%S')
        echo -e "${GREEN}[$timestamp]${NC} $line"
    done
}

# Мониторинг Prometheus метрик
monitor_prometheus_metrics() {
    echo -e "${YELLOW}📈 Мониторинг Prometheus метрик:${NC}"
    echo "Обновление каждые 5 секунд"
    echo "----------------------------------------"
    
    while true; do
        clear
        echo -e "${BLUE}📊 Prometheus метрики P2P Mesh - $(date)${NC}"
        echo "============================================="
        
        # Health check
        health=$(curl -s http://localhost:8081/health 2>/dev/null)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ Relay Health: OK${NC}"
        else
            echo -e "${RED}❌ Relay Health: FAILED${NC}"
        fi
        
        # Prometheus health
        prom_health=$(curl -s "$PROMETHEUS_URL/api/v1/status/config" 2>/dev/null)
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ Prometheus Health: OK${NC}"
        else
            echo -e "${RED}❌ Prometheus Health: FAILED${NC}"
        fi
        
        echo ""
        echo -e "${CYAN}📊 P2P метрики:${NC}"
        echo "----------------------------------------"
        
        # P2P connections
        p2p_connections=$(get_prometheus_metrics "p2p_connections_total")
        echo -e "P2P Connections: ${MAGENTA}$p2p_connections${NC}"
        
        # P2P latency
        p2p_latency=$(get_prometheus_metrics "p2p_latency_seconds")
        echo -e "P2P Latency: ${MAGENTA}$p2p_latency${NC}"
        
        # P2P peer count
        p2p_peers=$(get_prometheus_metrics "p2p_active_peers")
        echo -e "Active Peers: ${MAGENTA}$p2p_peers${NC}"
        
        # WireGuard peers
        wg_peers=$(get_prometheus_metrics "wireguard_peers_total")
        echo -e "WireGuard Peers: ${MAGENTA}$wg_peers${NC}"
        
        # Relay connections
        relay_connections=$(get_prometheus_metrics "relay_connections_total")
        echo -e "Relay Connections: ${MAGENTA}$relay_connections${NC}"
        
        echo ""
        echo -e "${CYAN}📈 Системные метрики:${NC}"
        echo "----------------------------------------"
        
        # CPU usage
        cpu_usage=$(get_prometheus_metrics "rate(process_cpu_seconds_total[5m]) * 100")
        echo -e "CPU Usage: ${MAGENTA}${cpu_usage}%${NC}"
        
        # Memory usage
        memory_usage=$(get_prometheus_metrics "process_resident_memory_bytes / 1024 / 1024")
        echo -e "Memory Usage: ${MAGENTA}${memory_usage} MB${NC}"
        
        # Go routines
        go_routines=$(get_prometheus_metrics "go_goroutines")
        echo -e "Go Routines: ${MAGENTA}$go_routines${NC}"
        
        echo ""
        echo -e "${YELLOW}Обновление через 5 секунд... (Ctrl+C для выхода)${NC}"
        echo -e "${BLUE}Grafana: $GRAFANA_URL${NC}"
        sleep 5
    done
}

# Детальный анализ метрик
detailed_metrics() {
    echo -e "${YELLOW}🔍 Детальный анализ метрик:${NC}"
    echo "----------------------------------------"
    
    # Получаем все P2P метрики
    echo -e "${CYAN}P2P метрики:${NC}"
    curl -s "$PROMETHEUS_URL/api/v1/query?query=p2p_" | jq -r '.data.result[] | "\(.metric.__name__) = \(.value[1])"' 2>/dev/null || echo "Метрики не найдены"
    
    echo ""
    echo -e "${CYAN}WireGuard метрики:${NC}"
    curl -s "$PROMETHEUS_URL/api/v1/query?query=wireguard_" | jq -r '.data.result[] | "\(.metric.__name__) = \(.value[1])"' 2>/dev/null || echo "Метрики не найдены"
    
    echo ""
    echo -e "${CYAN}Relay метрики:${NC}"
    curl -s "$PROMETHEUS_URL/api/v1/query?query=relay_" | jq -r '.data.result[] | "\(.metric.__name__) = \(.value[1])"' 2>/dev/null || echo "Метрики не найдены"
    
    echo ""
    read -p "Нажмите Enter для продолжения..."
}

# Экспорт метрик в файл
export_metrics() {
    local filename="p2p_metrics_$(date +%Y%m%d_%H%M%S).json"
    
    echo -e "${YELLOW}📤 Экспорт метрик в $filename:${NC}"
    
    # Собираем все метрики
    curl -s "$PROMETHEUS_URL/api/v1/query?query={__name__=~\"(p2p_|wireguard_|relay_)\"}" > "$filename" 2>/dev/null
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Метрики экспортированы в $filename${NC}"
        echo -e "${CYAN}Количество метрик: $(jq '.data.result | length' "$filename" 2>/dev/null || echo "0")${NC}"
    else
        echo -e "${RED}❌ Ошибка экспорта метрик${NC}"
    fi
    
    echo ""
    read -p "Нажмите Enter для продолжения..."
}

# Главное меню
while true; do
    clear
    echo -e "${BLUE}🔍 ДЕМО: Мониторинг P2P Mesh с Prometheus${NC}"
    echo "============================================="
    echo ""
    check_relay
    check_prometheus
    echo -e "${YELLOW}Выберите режим мониторинга:${NC}"
    echo "1) Мониторинг UDP трафика"
    echo "2) Мониторинг логов"
    echo "3) Мониторинг Prometheus метрик"
    echo "4) Детальный анализ метрик"
    echo "5) Экспорт метрик"
    echo "6) Обновить статус"
    echo "7) Выход"
    echo ""
    read -p "Введите номер (1-7): " choice
    
    case $choice in
        1)
            monitor_udp
            ;;
        2)
            monitor_logs
            ;;
        3)
            monitor_prometheus_metrics
            ;;
        4)
            detailed_metrics
            ;;
        5)
            export_metrics
            ;;
        6)
            check_relay
            check_prometheus
            read -p "Нажмите Enter для продолжения..."
            ;;
        7)
            echo -e "${GREEN}👋 Завершение мониторинга${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Неверный выбор${NC}"
            sleep 2
            ;;
    esac
done
