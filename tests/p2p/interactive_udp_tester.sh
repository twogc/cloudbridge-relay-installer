#!/bin/bash

# Интерактивный тестер UDP соединения для P2P mesh

HOST=${1:-"localhost"}
PORT=${2:-"51820"}

echo "🎮 Интерактивный UDP тестер для P2P Mesh"
echo "📍 Цель: $HOST:$PORT"
echo ""

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Функция для отправки данных
send_udp() {
    local message="$1"
    echo -e "${BLUE}📤 Отправка: $message${NC}"
    
    # Отправляем через netcat
    echo "$message" | nc -u -w 2 $HOST $PORT
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Отправлено успешно${NC}"
    else
        echo -e "${RED}❌ Ошибка отправки${NC}"
    fi
    echo ""
}

# Функция для мониторинга трафика
monitor_traffic() {
    echo -e "${YELLOW}🔍 Мониторинг трафика на порту $PORT (5 секунд)...${NC}"
    timeout 5 sudo tcpdump -i any -n udp port $PORT 2>/dev/null || echo "Трафик не обнаружен"
    echo ""
}

# Основное меню
while true; do
    echo -e "${YELLOW}Выберите действие:${NC}"
    echo "1) Отправить простое сообщение"
    echo "2) Отправить JSON данные"
    echo "3) Отправить WireGuard handshake"
    echo "4) Отправить P2P discovery"
    echo "5) Мониторинг трафика"
    echo "6) Интерактивный режим"
    echo "7) Выход"
    echo ""
    read -p "Введите номер (1-7): " choice
    
    case $choice in
        1)
            read -p "Введите сообщение: " message
            send_udp "$message"
            ;;
        2)
            json_data='{"type":"test","timestamp":"'$(date +%s)'","message":"JSON test"}'
            send_udp "$json_data"
            ;;
        3)
            handshake='{"type":"wireguard_handshake","peer_id":"test-peer","public_key":"test-key"}'
            send_udp "$handshake"
            ;;
        4)
            discovery='{"type":"peer_discovery","action":"register","peer_id":"test-peer"}'
            send_udp "$discovery"
            ;;
        5)
            monitor_traffic
            ;;
        6)
            echo -e "${YELLOW}🎮 Интерактивный режим (Ctrl+C для выхода)${NC}"
            echo "Вводите сообщения:"
            while true; do
                read -p "> " message
                if [ -z "$message" ]; then
                    break
                fi
                send_udp "$message"
            done
            ;;
        7)
            echo -e "${GREEN}👋 До свидания!${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Неверный выбор${NC}"
            ;;
    esac
done


