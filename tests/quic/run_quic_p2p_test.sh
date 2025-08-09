#!/bin/bash

# Главный скрипт для тестирования QUIC → P2P Mesh
# Клиент → QUIC → Сервер A → P2P Mesh

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}🚀 QUIC → P2P Mesh Test Suite${NC}"
echo "=================================="

# Функция для проверки зависимостей
check_dependencies() {
    echo -e "${BLUE}🔍 Проверка зависимостей...${NC}"
    
    local deps=("docker" "curl" "jq" "tmux" "netstat")
    local missing=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing+=("$dep")
        fi
    done
    
    if [ ${#missing[@]} -ne 0 ]; then
        echo -e "${RED}❌ Отсутствуют зависимости: ${missing[*]}${NC}"
        echo -e "${YELLOW}Установите недостающие пакеты:${NC}"
        echo "sudo apt update && sudo apt install -y ${missing[*]}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Все зависимости установлены${NC}"
}

# Функция для проверки статуса relay сервера
check_relay_status() {
    echo -e "${BLUE}🔍 Проверка статуса relay сервера...${NC}"
    
    if ! docker ps | grep -q cloudbridge-relay; then
        echo -e "${RED}❌ Relay сервер не запущен${NC}"
        echo -e "${YELLOW}Запустите relay сервер:${NC}"
        echo "docker-compose up -d"
        exit 1
    fi
    
    if ! curl -s "http://localhost:8081/health" > /dev/null; then
        echo -e "${RED}❌ Relay сервер недоступен${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Relay сервер работает${NC}"
}

# Функция для создания tmux сессии
create_tmux_session() {
    local session_name="quic-p2p-test"
    
    echo -e "${BLUE}🖥️  Создание tmux сессии...${NC}"
    
    # Удаляем существующую сессию если есть
    tmux kill-session -t "$session_name" 2>/dev/null || true
    
    # Создаем новую сессию
    tmux new-session -d -s "$session_name" -n "monitor"
    tmux new-window -t "$session_name" -n "client"
    tmux new-window -t "$session_name" -n "server-a"
    
    echo -e "${GREEN}✅ Tmux сессия создана: ${session_name}${NC}"
}

# Функция для запуска монитора
start_monitor() {
    echo -e "${BLUE}📊 Запуск монитора...${NC}"
    tmux send-keys -t "quic-p2p-test:monitor" "chmod +x test_quic_to_p2p_monitor.sh && ./test_quic_to_p2p_monitor.sh" Enter
    echo -e "${GREEN}✅ Монитор запущен${NC}"
}

# Функция для запуска клиента
start_client() {
    echo -e "${BLUE}👤 Запуск QUIC клиента...${NC}"
    tmux send-keys -t "quic-p2p-test:client" "chmod +x test_quic_to_p2p_client.sh && ./test_quic_to_p2p_client.sh" Enter
    echo -e "${GREEN}✅ QUIC клиент запущен${NC}"
}

# Функция для запуска сервера A
start_server_a() {
    echo -e "${BLUE}🖥️  Запуск сервера A...${NC}"
    tmux send-keys -t "quic-p2p-test:server-a" "chmod +x test_quic_to_p2p_server_a.sh && ./test_quic_to_p2p_server_a.sh" Enter
    echo -e "${GREEN}✅ Сервер A запущен${NC}"
}

# Функция для показа инструкций
show_instructions() {
    echo -e "\n${PURPLE}📋 Инструкции по использованию:${NC}"
    echo "=================================="
    echo -e "${YELLOW}1. Tmux сессия:${NC} tmux attach -t quic-p2p-test"
    echo -e "${YELLOW}2. Переключение между окнами:${NC} Ctrl+B, затем номер окна (0,1,2)"
    echo -e "${YELLOW}3. Окно 0 (monitor):${NC} Мониторинг всей системы"
    echo -e "${YELLOW}4. Окно 1 (client):${NC} QUIC клиент для подключения"
    echo -e "${YELLOW}5. Окно 2 (server-a):${NC} Сервер A для маршрутизации"
    echo ""
    echo -e "${YELLOW}Порядок тестирования:${NC}"
    echo "1. В окне client выберите '8' для автоматического теста"
    echo "2. В окне server-a выберите '10' для автоматического теста"
    echo "3. В окне monitor наблюдайте за процессом"
    echo ""
    echo -e "${YELLOW}Выход из tmux:${NC} Ctrl+B, затем D"
    echo -e "${YELLOW}Завершение сессии:${NC} tmux kill-session -t quic-p2p-test"
}

# Функция для автоматического теста
run_auto_test() {
    echo -e "${PURPLE}🧪 Запуск автоматического теста...${NC}"
    
    # Создаем сессию
    create_tmux_session
    
    # Запускаем компоненты
    start_monitor
    sleep 2
    start_server_a
    sleep 2
    start_client
    
    # Показываем инструкции
    show_instructions
    
    # Присоединяемся к сессии
    echo -e "${GREEN}🎯 Присоединяемся к tmux сессии...${NC}"
    tmux attach -t quic-p2p-test
}

# Функция для ручного запуска
run_manual() {
    echo -e "${PURPLE}🔧 Ручной запуск компонентов...${NC}"
    
    echo -e "${BLUE}Выберите компонент для запуска:${NC}"
    echo "1) Монитор"
    echo "2) QUIC клиент"
    echo "3) Сервер A"
    echo "4) Все компоненты в tmux"
    echo "5) Выход"
    echo -n "Введите номер (1-5): "
    
    read -r choice
    
    case $choice in
        1)
            chmod +x test_quic_to_p2p_monitor.sh
            ./test_quic_to_p2p_monitor.sh
            ;;
        2)
            chmod +x test_quic_to_p2p_client.sh
            ./test_quic_to_p2p_client.sh
            ;;
        3)
            chmod +x test_quic_to_p2p_server_a.sh
            ./test_quic_to_p2p_server_a.sh
            ;;
        4)
            run_auto_test
            ;;
        5)
            echo -e "${GREEN}👋 Выход${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Неверный выбор${NC}"
            ;;
    esac
}

# Функция для очистки
cleanup() {
    echo -e "${BLUE}🧹 Очистка...${NC}"
    tmux kill-session -t quic-p2p-test 2>/dev/null || true
    echo -e "${GREEN}✅ Очистка завершена${NC}"
}

# Функция для показа статуса
show_status() {
    echo -e "${BLUE}📊 Статус системы...${NC}"
    
    # Relay сервер
    if docker ps | grep -q cloudbridge-relay; then
        echo -e "${GREEN}✅ Relay сервер: запущен${NC}"
    else
        echo -e "${RED}❌ Relay сервер: остановлен${NC}"
    fi
    
    # Tmux сессия
    if tmux has-session -t quic-p2p-test 2>/dev/null; then
        echo -e "${GREEN}✅ Tmux сессия: активна${NC}"
    else
        echo -e "${YELLOW}⚠️  Tmux сессия: не создана${NC}"
    fi
    
    # UDP порт
    if netstat -tuln | grep -q ":51820"; then
        echo -e "${GREEN}✅ UDP порт 51820: слушает${NC}"
    else
        echo -e "${RED}❌ UDP порт 51820: не слушает${NC}"
    fi
}

# Основное меню
show_main_menu() {
    echo -e "\n${BLUE}Выберите действие:${NC}"
    echo "1) Автоматический тест (все компоненты)"
    echo "2) Ручной запуск компонентов"
    echo "3) Показать статус"
    echo "4) Очистка tmux сессий"
    echo "5) Выход"
    echo -n "Введите номер (1-5): "
}

# Основной цикл
main() {
    # Проверяем зависимости
    check_dependencies
    
    # Проверяем статус relay сервера
    check_relay_status
    
    while true; do
        show_main_menu
        read -r choice
        
        case $choice in
            1)
                run_auto_test
                ;;
            2)
                run_manual
                ;;
            3)
                show_status
                ;;
            4)
                cleanup
                ;;
            5)
                echo -e "${GREEN}👋 Выход из теста${NC}"
                cleanup
                exit 0
                ;;
            *)
                echo -e "${RED}❌ Неверный выбор${NC}"
                ;;
        esac
        
        echo -e "\n${YELLOW}Нажмите Enter для продолжения...${NC}"
        read -r
    done
}

# Обработка сигналов
trap 'echo -e "\n${YELLOW}🛑 Получен сигнал прерывания${NC}"; cleanup; exit 0' INT TERM

# Запуск
main


