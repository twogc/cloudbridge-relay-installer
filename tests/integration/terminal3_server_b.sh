#!/bin/bash

# Терминал 3: Server B - Получение и ответ
# Симуляция второго сервера в P2P mesh

echo "🖥️  ТЕРМИНАЛ 3: Server B - Data Receiver & Responder"
echo "===================================================="
echo "📍 Получение данных с порта 51820/udp и ответы"
echo ""

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Конфигурация
LISTEN_PORT=${1:-"51821"}
SERVER_ID="server-b-$(hostname)"

# Функция для прослушивания UDP порта
listen_udp() {
    echo -e "${MAGENTA}👂 [Server B] Прослушивание порта $LISTEN_PORT/udp${NC}"
    echo "----------------------------------------"
    
    # Создаем временный файл для логов
    local log_file="/tmp/server_b_udp_$(date +%s).log"
    
    # Запускаем netcat в фоне для прослушивания
    nc -u -l $LISTEN_PORT > "$log_file" 2>&1 &
    local nc_pid=$!
    
    echo -e "${GREEN}✅ [Server B] Прослушивание запущено (PID: $nc_pid)${NC}"
    echo -e "${YELLOW}📝 Логи сохраняются в: $log_file${NC}"
    echo ""
    
    # Мониторим файл логов
    tail -f "$log_file" | while read line; do
        if [ -n "$line" ]; then
            echo -e "${MAGENTA}📥 [Server B] Получено: $line${NC}"
            
            # Пытаемся парсить JSON
            if echo "$line" | jq . >/dev/null 2>&1; then
                echo -e "${BLUE}📋 [Server B] JSON данные получены${NC}"
                
                # Извлекаем тип сообщения
                local msg_type=$(echo "$line" | jq -r '.type // "unknown"')
                local server_id=$(echo "$line" | jq -r '.server_id // "unknown"')
                
                echo -e "${CYAN}🔍 [Server B] Тип: $msg_type, От: $server_id${NC}"
                
                # Отправляем ответ в зависимости от типа
                case "$msg_type" in
                    "wireguard_handshake")
                        send_wireguard_response "$server_id"
                        ;;
                    "peer_discovery")
                        send_discovery_response "$server_id"
                        ;;
                    "p2p_test")
                        send_test_response "$server_id" "$line"
                        ;;
                    *)
                        send_generic_response "$server_id"
                        ;;
                esac
            else
                echo -e "${YELLOW}📝 [Server B] Простое текстовое сообщение${NC}"
                send_simple_response "$line"
            fi
            
            echo ""
        fi
    done
    
    # Очистка при выходе
    trap "kill $nc_pid 2>/dev/null; rm -f $log_file; echo -e '${GREEN}👋 [Server B] Прослушивание остановлено${NC}'; exit" INT TERM
}

# Функция для отправки ответа на WireGuard handshake
send_wireguard_response() {
    local sender_id="$1"
    
    local response=$(cat <<EOF
{
    "type": "wireguard_response",
    "version": "1.0",
    "timestamp": $(date +%s),
    "peer_id": "$SERVER_ID",
    "public_key": "server-b-public-key",
    "endpoint": "localhost:$LISTEN_PORT",
    "capabilities": ["p2p", "mesh", "quic"],
    "server_id": "$SERVER_ID",
    "response_to": "$sender_id",
    "status": "accepted"
}
EOF
)
    
    echo -e "${MAGENTA}🔐 [Server B] Отправка WireGuard ответа${NC}"
    echo "$response" | nc -u -w 2 localhost 51820
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ [Server B] WireGuard ответ отправлен${NC}"
    else
        echo -e "${RED}❌ [Server B] Ошибка отправки WireGuard ответа${NC}"
    fi
}

# Функция для отправки ответа на P2P discovery
send_discovery_response() {
    local sender_id="$1"
    
    local response=$(cat <<EOF
{
    "type": "peer_discovery_response",
    "action": "acknowledge",
    "timestamp": $(date +%s),
    "peer_info": {
        "id": "$SERVER_ID",
        "region": "test-region-b",
        "capabilities": ["mesh", "quic", "wireguard"],
        "endpoints": ["localhost:$LISTEN_PORT"],
        "last_seen": $(date +%s),
        "server_type": "relay"
    },
    "response_to": "$sender_id",
    "status": "registered"
}
EOF
)
    
    echo -e "${MAGENTA}🔍 [Server B] Отправка discovery ответа${NC}"
    echo "$response" | nc -u -w 2 localhost 51820
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ [Server B] Discovery ответ отправлен${NC}"
    else
        echo -e "${RED}❌ [Server B] Ошибка отправки discovery ответа${NC}"
    fi
}

# Функция для отправки ответа на тестовые данные
send_test_response() {
    local sender_id="$1"
    local original_data="$2"
    
    # Извлекаем sequence из оригинальных данных
    local sequence=$(echo "$original_data" | jq -r '.sequence // 1')
    
    local response=$(cat <<EOF
{
    "type": "p2p_test_response",
    "timestamp": $(date +%s),
    "sequence": $sequence,
    "server_id": "$SERVER_ID",
    "response_to": "$sender_id",
    "data": "Response to test message $sequence from Server B",
    "echo": $(echo "$original_data" | jq -c '.data // "no data"')
}
EOF
)
    
    echo -e "${MAGENTA}📊 [Server B] Отправка тестового ответа${NC}"
    echo "$response" | nc -u -w 2 localhost 51820
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ [Server B] Тестовый ответ отправлен${NC}"
    else
        echo -e "${RED}❌ [Server B] Ошибка отправки тестового ответа${NC}"
    fi
}

# Функция для отправки простого ответа
send_simple_response() {
    local original_message="$1"
    
    local response="Echo from Server B: $original_message"
    
    echo -e "${MAGENTA}📝 [Server B] Отправка простого ответа${NC}"
    echo "$response" | nc -u -w 2 localhost 51820
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ [Server B] Простой ответ отправлен${NC}"
    else
        echo -e "${RED}❌ [Server B] Ошибка отправки простого ответа${NC}"
    fi
}

# Функция для отправки общего ответа
send_generic_response() {
    local sender_id="$1"
    
    local response=$(cat <<EOF
{
    "type": "generic_response",
    "timestamp": $(date +%s),
    "server_id": "$SERVER_ID",
    "response_to": "$sender_id",
    "message": "Message received and processed by Server B",
    "status": "ok"
}
EOF
)
    
    echo -e "${MAGENTA}📨 [Server B] Отправка общего ответа${NC}"
    echo "$response" | nc -u -w 2 localhost 51820
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ [Server B] Общий ответ отправлен${NC}"
    else
        echo -e "${RED}❌ [Server B] Ошибка отправки общего ответа${NC}"
    fi
}

# Функция для активной отправки данных
active_send() {
    echo -e "${YELLOW}📤 [Server B] Активная отправка данных${NC}"
    echo "----------------------------------------"
    
    local target_host="localhost"
    local target_port="51820"
    
    # Отправляем регистрацию
    local registration=$(cat <<EOF
{
    "type": "peer_discovery",
    "action": "register",
    "peer_info": {
        "id": "$SERVER_ID",
        "region": "test-region-b",
        "capabilities": ["mesh", "quic", "wireguard"],
        "endpoints": ["localhost:$LISTEN_PORT"],
        "last_seen": $(date +%s),
        "server_type": "relay"
    }
}
EOF
)
    
    echo -e "${MAGENTA}🔍 [Server B] Отправка регистрации${NC}"
    echo "$registration" | nc -u -w 2 $target_host $target_port
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ [Server B] Регистрация отправлена${NC}"
    else
        echo -e "${RED}❌ [Server B] Ошибка регистрации${NC}"
    fi
    echo ""
}

# Функция для мониторинга трафика
monitor_traffic() {
    echo -e "${YELLOW}🌐 [Server B] Мониторинг UDP трафика${NC}"
    echo "----------------------------------------"
    sudo tcpdump -i any -n udp port $LISTEN_PORT -v
}

# Главное меню
while true; do
    clear
    echo -e "${BLUE}🖥️  ТЕРМИНАЛ 3: Server B - Data Receiver & Responder${NC}"
    echo "===================================================="
    echo -e "${MAGENTA}📍 Прослушивание: $LISTEN_PORT/udp${NC}"
    echo -e "${MAGENTA}🆔 Server ID: $SERVER_ID${NC}"
    echo ""
    
    echo -e "${YELLOW}Выберите действие:${NC}"
    echo "1) Запустить прослушивание UDP"
    echo "2) Активная отправка данных"
    echo "3) Мониторинг трафика"
    echo "4) Настройка порта"
    echo "5) Выход"
    echo ""
    read -p "Введите номер (1-5): " choice
    
    case $choice in
        1)
            listen_udp
            ;;
        2)
            active_send
            read -p "Нажмите Enter для продолжения..."
            ;;
        3)
            monitor_traffic
            ;;
        4)
            echo -e "${YELLOW}Настройка порта:${NC}"
            read -p "Новый порт [$LISTEN_PORT]: " new_port
            
            if [ -n "$new_port" ]; then
                LISTEN_PORT="$new_port"
            fi
            
            echo -e "${GREEN}✅ Порт обновлен: $LISTEN_PORT${NC}"
            sleep 2
            ;;
        5)
            echo -e "${GREEN}👋 [Server B] Завершение работы${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Неверный выбор${NC}"
            sleep 2
            ;;
    esac
done


