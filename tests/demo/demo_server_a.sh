#!/bin/bash

# Демонстрационный скрипт Server A
# Отправляет данные на порт 51820/udp для тестирования P2P mesh

echo "🖥️  ДЕМО: Server A - Отправка данных"
echo "===================================="
echo ""

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Конфигурация
TARGET_HOST="localhost"
TARGET_PORT="51820"
SERVER_ID="demo-server-a"

# Функция отправки сообщения
send_message() {
    local message="$1"
    local type="$2"
    
    echo -e "${CYAN}📤 Отправка: $type${NC}"
    echo "$message" | nc -u -w 2 $TARGET_HOST $TARGET_PORT
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Отправлено успешно${NC}"
    else
        echo -e "${RED}❌ Ошибка отправки${NC}"
    fi
    echo ""
}

# Отправка простого сообщения
send_simple() {
    read -p "Введите сообщение: " message
    send_message "$message" "простое сообщение"
}

# Отправка WireGuard handshake
send_handshake() {
    local handshake=$(cat <<EOF
{
    "type": "wireguard_handshake",
    "timestamp": $(date +%s),
    "peer_id": "demo-peer-a",
    "public_key": "demo-public-key-a",
    "server_id": "$SERVER_ID",
    "message": "Hello from Server A!"
}
EOF
)
    send_message "$handshake" "WireGuard handshake"
}

# Отправка P2P discovery
send_discovery() {
    local discovery=$(cat <<EOF
{
    "type": "peer_discovery",
    "action": "register",
    "timestamp": $(date +%s),
    "server_id": "$SERVER_ID",
    "region": "demo-region",
    "capabilities": ["p2p", "mesh", "quic"]
}
EOF
)
    send_message "$discovery" "P2P discovery"
}

# Отправка тестовых данных
send_test_data() {
    local sequence="$1"
    local test_data=$(cat <<EOF
{
    "type": "p2p_test",
    "timestamp": $(date +%s),
    "sequence": $sequence,
    "server_id": "$SERVER_ID",
    "data": "Test message $sequence from Server A",
    "target": "$TARGET_HOST:$TARGET_PORT"
}
EOF
)
    send_message "$test_data" "тестовые данные ($sequence)"
}

# Автоматическая отправка серии
auto_send() {
    echo -e "${YELLOW}🔄 Автоматическая отправка серии сообщений${NC}"
    echo "----------------------------------------"
    
    # Отправляем handshake
    send_handshake
    sleep 1
    
    # Отправляем discovery
    send_discovery
    sleep 1
    
    # Отправляем тестовые данные
    for i in {1..5}; do
        send_test_data $i
        sleep 2
    done
    
    echo -e "${GREEN}🎉 Автоматическая отправка завершена!${NC}"
    echo ""
}

# Интерактивный режим
interactive_mode() {
    echo -e "${YELLOW}🎮 Интерактивный режим${NC}"
    echo "Введите сообщения (пустая строка для выхода):"
    echo "----------------------------------------"
    
    while true; do
        read -p "> " message
        if [ -z "$message" ]; then
            break
        fi
        send_message "$message" "интерактивное сообщение"
    done
}

# Демонстрационный режим
demo_mode() {
    echo -e "${BLUE}🎬 Демонстрационный режим${NC}"
    echo "Показывает все типы сообщений автоматически"
    echo "----------------------------------------"
    
    echo -e "${YELLOW}Этап 1: WireGuard Handshake${NC}"
    send_handshake
    sleep 3
    
    echo -e "${YELLOW}Этап 2: P2P Discovery${NC}"
    send_discovery
    sleep 3
    
    echo -e "${YELLOW}Этап 3: Тестовые данные${NC}"
    for i in {1..3}; do
        send_test_data $i
        sleep 2
    done
    
    echo -e "${GREEN}🎉 Демонстрация завершена!${NC}"
    echo "Теперь запустите Server B для получения ответов"
    echo ""
}

# Главное меню
while true; do
    clear
    echo -e "${BLUE}🖥️  ДЕМО: Server A - Отправка данных${NC}"
    echo "========================================="
    echo -e "${CYAN}📍 Цель: $TARGET_HOST:$TARGET_PORT${NC}"
    echo -e "${CYAN}🆔 Server ID: $SERVER_ID${NC}"
    echo ""
    
    echo -e "${YELLOW}Выберите действие:${NC}"
    echo "1) Отправить простое сообщение"
    echo "2) Отправить WireGuard handshake"
    echo "3) Отправить P2P discovery"
    echo "4) Отправить тестовые данные"
    echo "5) Автоматическая отправка серии"
    echo "6) Интерактивный режим"
    echo "7) Демонстрационный режим"
    echo "8) Выход"
    echo ""
    read -p "Введите номер (1-8): " choice
    
    case $choice in
        1)
            send_simple
            read -p "Нажмите Enter для продолжения..."
            ;;
        2)
            send_handshake
            read -p "Нажмите Enter для продолжения..."
            ;;
        3)
            send_discovery
            read -p "Нажмите Enter для продолжения..."
            ;;
        4)
            read -p "Введите номер последовательности: " seq
            send_test_data "${seq:-1}"
            read -p "Нажмите Enter для продолжения..."
            ;;
        5)
            auto_send
            read -p "Нажмите Enter для продолжения..."
            ;;
        6)
            interactive_mode
            ;;
        7)
            demo_mode
            read -p "Нажмите Enter для продолжения..."
            ;;
        8)
            echo -e "${GREEN}👋 Завершение работы Server A${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Неверный выбор${NC}"
            sleep 2
            ;;
    esac
done


