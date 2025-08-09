#!/bin/bash

# Терминал 1: Relay Server Monitor
# Мониторинг P2P mesh соединений и логов

echo "🖥️  ТЕРМИНАЛ 1: Relay Server Monitor"
echo "====================================="
echo "📍 Мониторинг P2P mesh на порту 51820/udp"
echo ""

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Функция для отображения статуса
show_status() {
    echo -e "${BLUE}📊 Статус Relay Server:${NC}"
    echo "----------------------------------------"
    
    # Проверка контейнера
    if docker ps | grep -q cloudbridge-relay; then
        echo -e "${GREEN}✅ Контейнер cloudbridge-relay запущен${NC}"
    else
        echo -e "${RED}❌ Контейнер cloudbridge-relay не найден${NC}"
        return 1
    fi
    
    # Проверка порта
    if netstat -tuln | grep -q ":51820"; then
        echo -e "${GREEN}✅ Порт 51820/udp открыт${NC}"
    else
        echo -e "${RED}❌ Порт 51820/udp не найден${NC}"
    fi
    
    # Health check
    health=$(curl -s http://localhost:8081/health 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Health check: OK${NC}"
    else
        echo -e "${RED}❌ Health check: FAILED${NC}"
    fi
    
    echo ""
}

# Функция для мониторинга логов
monitor_logs() {
    echo -e "${YELLOW}📋 Мониторинг логов P2P mesh (Ctrl+C для остановки):${NC}"
    echo "----------------------------------------"
    docker logs -f cloudbridge-relay | grep -E "(p2p|wireguard|mesh|peer|discovery)" --color=always
}

# Функция для мониторинга трафика
monitor_traffic() {
    echo -e "${YELLOW}🌐 Мониторинг UDP трафика на порту 51820:${NC}"
    echo "----------------------------------------"
    sudo tcpdump -i any -n udp port 51820 -v
}

# Функция для мониторинга метрик
monitor_metrics() {
    echo -e "${YELLOW}📈 Мониторинг метрик:${NC}"
    echo "----------------------------------------"
    while true; do
        clear
        echo -e "${BLUE}📊 Метрики P2P Mesh - $(date)${NC}"
        echo "----------------------------------------"
        
        # P2P метрики
        metrics=$(curl -s http://localhost:9090/metrics 2>/dev/null | grep -E "(p2p|wireguard|mesh)" | head -10)
        if [ -n "$metrics" ]; then
            echo "$metrics"
        else
            echo "Метрики не найдены"
        fi
        
        echo ""
        echo -e "${YELLOW}Обновление через 5 секунд... (Ctrl+C для выхода)${NC}"
        sleep 5
    done
}

# Функция для анализа соединений
analyze_connections() {
    echo -e "${YELLOW}🔍 Анализ активных соединений:${NC}"
    echo "----------------------------------------"
    
    # Проверка WireGuard интерфейса
    echo "WireGuard интерфейсы:"
    docker exec cloudbridge-relay wg show 2>/dev/null || echo "WireGuard не активен"
    
    echo ""
    echo "Сетевые соединения:"
    netstat -tuln | grep 51820
    
    echo ""
    echo "Docker сетевые интерфейсы:"
    docker exec cloudbridge-relay ip addr show | grep -A 2 -B 2 wg
}

# Главное меню
while true; do
    clear
    echo -e "${BLUE}🖥️  ТЕРМИНАЛ 1: Relay Server Monitor${NC}"
    echo "============================================="
    echo ""
    show_status
    echo -e "${YELLOW}Выберите действие:${NC}"
    echo "1) Мониторинг логов P2P mesh"
    echo "2) Мониторинг UDP трафика"
    echo "3) Мониторинг метрик"
    echo "4) Анализ соединений"
    echo "5) Обновить статус"
    echo "6) Выход"
    echo ""
    read -p "Введите номер (1-6): " choice
    
    case $choice in
        1)
            monitor_logs
            ;;
        2)
            monitor_traffic
            ;;
        3)
            monitor_metrics
            ;;
        4)
            analyze_connections
            read -p "Нажмите Enter для продолжения..."
            ;;
        5)
            show_status
            read -p "Нажмите Enter для продолжения..."
            ;;
        6)
            echo -e "${GREEN}👋 Завершение мониторинга${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Неверный выбор${NC}"
            sleep 2
            ;;
    esac
done


