#!/bin/bash

# Демонстрационный скрипт Server B
# Получает данные и отправляет ответы для тестирования P2P mesh

echo "🖥️  ДЕМО: Server B - Получение и ответ"
echo "======================================"
echo ""

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Конфигурация
LISTEN_PORT="51821"
SERVER_ID="demo-server-b"

# Функция для отправки ответа
send_response() {
    local response="$1"
    local target_port="51820"
    
    echo -e "${MAGENTA}📤 Ответ: $response${NC}"
    echo "$response" | nc -u -w 2 localhost $target_port
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Ответ отправлен${NC}"
    else
        echo -e "${RED}❌ Ошибка отправки ответа${NC}"
    fi
    echo ""
}

# Обработка полученного сообщения
process_message() {
    local message="$1"
    local timestamp=$(date '+%H:%M:%S')
    
    echo -e "${MAGENTA}📥 [$timestamp] Получено: $message${NC}"
    
    # Пытаемся парсить JSON
    if echo "$message" | jq . >/dev/null 2>&1; then
        echo -e "${BLUE}📋 JSON данные получены${NC}"
        
        # Извлекаем тип сообщения
        local msg_type=$(echo "$message" | jq -r '.type // "unknown"')
        local server_id=$(echo "$message" | jq -r '.server_id // "unknown"')
        
        echo -e "${CYAN}🔍 Тип: $msg_type, От: $server_id${NC}"
        
        # Отправляем ответ в зависимости от типа
        case "$msg_type" in
            "wireguard_handshake")
                local response=$(cat <<EOF
{
    "type": "wireguard_response",
    "timestamp": $(date +%s),
    "peer_id": "$SERVER_ID",
    "server_id": "$SERVER_ID",
    "response_to": "$server_id",
    "status": "accepted",
    "message": "Handshake accepted by Server B"
}
EOF
)
                send_response "$response"
                ;;
            "peer_discovery")
                local response=$(cat <<EOF
{
    "type": "peer_discovery_response",
    "timestamp": $(date +%s),
    "server_id": "$SERVER_ID",
    "response_to": "$server_id",
    "status": "registered",
    "message": "Peer registered by Server B"
}
EOF
)
                send_response "$response"
                ;;
            "p2p_test")
                local sequence=$(echo "$message" | jq -r '.sequence // 1')
                local response=$(cat <<EOF
{
    "type": "p2p_test_response",
    "timestamp": $(date +%s),
    "sequence": $sequence,
    "server_id": "$SERVER_ID",
    "response_to": "$server_id",
    "data": "Response to test message $sequence from Server B",
    "echo": $(echo "$message" | jq -c '.data // "no data"')
}
EOF
)
                send_response "$response"
                ;;
            *)
                local response=$(cat <<EOF
{
    "type": "generic_response",
    "timestamp": $(date +%s),
    "server_id": "$SERVER_ID",
    "response_to": "$server_id",
    "message": "Message received and processed by Server B",
    "status": "ok"
}
EOF
)
                send_response "$response"
                ;;
        esac
    else
        echo -e "${YELLOW}📝 Простое текстовое сообщение${NC}"
        local response="Echo from Server B: $message"
        send_response "$response"
    fi
}

# Прослушивание UDP порта
listen_udp() {
    echo -e "${MAGENTA}👂 Прослушивание порта $LISTEN_PORT/udp${NC}"
    echo "Нажмите Ctrl+C для остановки"
    echo "----------------------------------------"
    
    # Создаем временный файл для логов
    local log_file="/tmp/server_b_demo_$(date +%s).log"
    
    # Запускаем netcat в фоне
    nc -u -l $LISTEN_PORT > "$log_file" 2>&1 &
    local nc_pid=$!
    
    echo -e "${GREEN}✅ Прослушивание запущено (PID: $nc_pid)${NC}"
    echo -e "${YELLOW}📝 Логи: $log_file${NC}"
    echo ""
    
    # Мониторим файл логов
    tail -f "$log_file" | while read line; do
        if [ -n "$line" ]; then
            process_message "$line"
        fi
    done
    
    # Очистка при выходе
    trap "kill $nc_pid 2>/dev/null; rm -f $log_file; echo -e '${GREEN}👋 Прослушивание остановлено${NC}'; exit" INT TERM
}

# Активная отправка данных
active_send() {
    echo -e "${YELLOW}📤 Активная отправка данных${NC}"
    echo "----------------------------------------"
    
    # Отправляем регистрацию
    local registration=$(cat <<EOF
{
    "type": "peer_discovery",
    "action": "register",
    "timestamp": $(date +%s),
    "server_id": "$SERVER_ID",
    "region": "demo-region-b",
    "capabilities": ["mesh", "quic", "wireguard"]
}
EOF
)
    
    echo -e "${MAGENTA}🔍 Отправка регистрации${NC}"
    echo "$registration" | nc -u -w 2 localhost 51820
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Регистрация отправлена${NC}"
    else
        echo -e "${RED}❌ Ошибка регистрации${NC}"
    fi
    echo ""
}

# Демонстрационный режим
demo_mode() {
    echo -e "${BLUE}🎬 Демонстрационный режим Server B${NC}"
    echo "Автоматически отвечает на все сообщения"
    echo "----------------------------------------"
    
    # Запускаем прослушивание в фоне
    nc -u -l $LISTEN_PORT > /tmp/demo_server_b.log 2>&1 &
    local nc_pid=$!
    
    echo -e "${GREEN}✅ Прослушивание запущено${NC}"
    echo ""
    
    # Отправляем регистрацию
    active_send
    
    echo -e "${YELLOW}📋 Мониторинг входящих сообщений:${NC}"
    tail -f /tmp/demo_server_b.log | while read line; do
        if [ -n "$line" ]; then
            process_message "$line"
        fi
    done
    
    # Очистка
    trap "kill $nc_pid 2>/dev/null; rm -f /tmp/demo_server_b.log; exit" INT TERM
}

# Главное меню
while true; do
    clear
    echo -e "${BLUE}🖥️  ДЕМО: Server B - Получение и ответ${NC}"
    echo "=========================================="
    echo -e "${MAGENTA}📍 Прослушивание: $LISTEN_PORT/udp${NC}"
    echo -e "${MAGENTA}🆔 Server ID: $SERVER_ID${NC}"
    echo ""
    
    echo -e "${YELLOW}Выберите действие:${NC}"
    echo "1) Запустить прослушивание UDP"
    echo "2) Активная отправка данных"
    echo "3) Демонстрационный режим"
    echo "4) Выход"
    echo ""
    read -p "Введите номер (1-4): " choice
    
    case $choice in
        1)
            listen_udp
            ;;
        2)
            active_send
            read -p "Нажмите Enter для продолжения..."
            ;;
        3)
            demo_mode
            ;;
        4)
            echo -e "${GREEN}👋 Завершение работы Server B${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Неверный выбор${NC}"
            sleep 2
            ;;
    esac
done


