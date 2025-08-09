#!/bin/bash

# Терминал 2: Server A - Отправка данных
# Симуляция первого сервера в P2P mesh

echo "🖥️  ТЕРМИНАЛ 2: Server A - Data Sender"
echo "======================================"
echo "📍 Отправка данных на порт 51820/udp"
echo ""

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Конфигурация
TARGET_HOST=${1:-"localhost"}
TARGET_PORT=${2:-"51820"}
SERVER_ID="server-a-$(hostname)"

# Функция для отправки простого сообщения
send_simple_message() {
    local message="$1"
    echo -e "${CYAN}📤 [Server A] Отправка: $message${NC}"
    
    echo "$message" | nc -u -w 2 $TARGET_HOST $TARGET_PORT
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ [Server A] Отправлено успешно${NC}"
    else
        echo -e "${RED}❌ [Server A] Ошибка отправки${NC}"
    fi
    echo ""
}

# Функция для отправки WireGuard handshake
send_wireguard_handshake() {
    local peer_id="$1"
    local public_key="$2"
    
    local handshake_data=$(cat <<EOF
{
    "type": "wireguard_handshake",
    "version": "1.0",
    "timestamp": $(date +%s),
    "peer_id": "$peer_id",
    "public_key": "$public_key",
    "endpoint": "$TARGET_HOST:$TARGET_PORT",
    "capabilities": ["p2p", "mesh", "quic"],
    "server_id": "$SERVER_ID"
}
EOF
)
    
    echo -e "${CYAN}🔐 [Server A] Отправка WireGuard handshake${NC}"
    echo "$handshake_data" | nc -u -w 2 $TARGET_HOST $TARGET_PORT
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ [Server A] Handshake отправлен${NC}"
    else
        echo -e "${RED}❌ [Server A] Ошибка handshake${NC}"
    fi
    echo ""
}

# Функция для отправки P2P discovery
send_p2p_discovery() {
    local action="$1"
    
    local discovery_data=$(cat <<EOF
{
    "type": "peer_discovery",
    "action": "$action",
    "peer_info": {
        "id": "$SERVER_ID",
        "region": "test-region-a",
        "capabilities": ["mesh", "quic", "wireguard"],
        "endpoints": ["$TARGET_HOST:$TARGET_PORT"],
        "last_seen": $(date +%s),
        "server_type": "relay"
    }
}
EOF
)
    
    echo -e "${CYAN}🔍 [Server A] Отправка P2P discovery ($action)${NC}"
    echo "$discovery_data" | nc -u -w 2 $TARGET_HOST $TARGET_PORT
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ [Server A] Discovery отправлен${NC}"
    else
        echo -e "${RED}❌ [Server A] Ошибка discovery${NC}"
    fi
    echo ""
}

# Функция для отправки тестовых данных
send_test_data() {
    local sequence="$1"
    local data_type="$2"
    
    local test_data=$(cat <<EOF
{
    "type": "p2p_test",
    "timestamp": $(date +%s),
    "sequence": $sequence,
    "data_type": "$data_type",
    "server_id": "$SERVER_ID",
    "data": "Test message $sequence from Server A",
    "target": "$TARGET_HOST:$TARGET_PORT"
}
EOF
)
    
    echo -e "${CYAN}📊 [Server A] Отправка тестовых данных ($sequence/$data_type)${NC}"
    echo "$test_data" | nc -u -w 2 $TARGET_HOST $TARGET_PORT
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ [Server A] Тестовые данные отправлены${NC}"
    else
        echo -e "${RED}❌ [Server A] Ошибка отправки тестовых данных${NC}"
    fi
    echo ""
}

# Функция для автоматической отправки серии сообщений
auto_send_series() {
    local count="$1"
    local interval="$2"
    
    echo -e "${YELLOW}🔄 [Server A] Автоматическая отправка $count сообщений с интервалом $interval сек${NC}"
    echo "----------------------------------------"
    
    for i in $(seq 1 $count); do
        send_test_data $i "auto-series"
        sleep $interval
    done
    
    echo -e "${GREEN}🎉 [Server A] Автоматическая отправка завершена${NC}"
}

# Функция для интерактивной отправки
interactive_send() {
    echo -e "${YELLOW}🎮 [Server A] Интерактивный режим отправки${NC}"
    echo "Введите сообщения (пустая строка для выхода):"
    echo "----------------------------------------"
    
    local counter=1
    while true; do
        read -p "[Server A] > " message
        if [ -z "$message" ]; then
            break
        fi
        
        send_simple_message "$message"
        counter=$((counter + 1))
    done
}

# Главное меню
while true; do
    clear
    echo -e "${BLUE}🖥️  ТЕРМИНАЛ 2: Server A - Data Sender${NC}"
    echo "============================================="
    echo -e "${CYAN}📍 Цель: $TARGET_HOST:$TARGET_PORT${NC}"
    echo -e "${CYAN}🆔 Server ID: $SERVER_ID${NC}"
    echo ""
    
    echo -e "${YELLOW}Выберите действие:${NC}"
    echo "1) Отправить простое сообщение"
    echo "2) Отправить WireGuard handshake"
    echo "3) Отправить P2P discovery (register)"
    echo "4) Отправить P2P discovery (unregister)"
    echo "5) Отправить тестовые данные"
    echo "6) Автоматическая отправка серии"
    echo "7) Интерактивный режим"
    echo "8) Настройка цели"
    echo "9) Выход"
    echo ""
    read -p "Введите номер (1-9): " choice
    
    case $choice in
        1)
            read -p "Введите сообщение: " message
            send_simple_message "$message"
            read -p "Нажмите Enter для продолжения..."
            ;;
        2)
            read -p "Введите Peer ID: " peer_id
            read -p "Введите Public Key: " public_key
            send_wireguard_handshake "${peer_id:-test-peer-a}" "${public_key:-test-key-a}"
            read -p "Нажмите Enter для продолжения..."
            ;;
        3)
            send_p2p_discovery "register"
            read -p "Нажмите Enter для продолжения..."
            ;;
        4)
            send_p2p_discovery "unregister"
            read -p "Нажмите Enter для продолжения..."
            ;;
        5)
            read -p "Введите тип данных: " data_type
            send_test_data 1 "${data_type:-manual-test}"
            read -p "Нажмите Enter для продолжения..."
            ;;
        6)
            read -p "Количество сообщений: " count
            read -p "Интервал (сек): " interval
            auto_send_series "${count:-5}" "${interval:-2}"
            read -p "Нажмите Enter для продолжения..."
            ;;
        7)
            interactive_send
            ;;
        8)
            echo -e "${YELLOW}Настройка цели:${NC}"
            read -p "Новый хост [$TARGET_HOST]: " new_host
            read -p "Новый порт [$TARGET_PORT]: " new_port
            
            if [ -n "$new_host" ]; then
                TARGET_HOST="$new_host"
            fi
            if [ -n "$new_port" ]; then
                TARGET_PORT="$new_port"
            fi
            
            echo -e "${GREEN}✅ Цель обновлена: $TARGET_HOST:$TARGET_PORT${NC}"
            sleep 2
            ;;
        9)
            echo -e "${GREEN}👋 [Server A] Завершение работы${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Неверный выбор${NC}"
            sleep 2
            ;;
    esac
done


