#!/bin/bash

# Главный скрипт для запуска трех терминалов тестирования P2P mesh

echo "🚀 Запуск трех терминалов для тестирования P2P Mesh"
echo "=================================================="
echo ""

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Проверка зависимостей
check_dependencies() {
    echo -e "${YELLOW}🔍 Проверка зависимостей...${NC}"
    
    local missing_deps=()
    
    # Проверяем необходимые команды
    for cmd in docker nc jq curl netstat; do
        if ! command -v $cmd &> /dev/null; then
            missing_deps+=($cmd)
        fi
    done
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo -e "${RED}❌ Отсутствуют зависимости: ${missing_deps[*]}${NC}"
        echo "Установите недостающие пакеты и повторите попытку."
        exit 1
    fi
    
    echo -e "${GREEN}✅ Все зависимости установлены${NC}"
    echo ""
}

# Проверка статуса relay сервера
check_relay_status() {
    echo -e "${YELLOW}🔍 Проверка статуса Relay Server...${NC}"
    
    if ! docker ps | grep -q cloudbridge-relay; then
        echo -e "${RED}❌ Контейнер cloudbridge-relay не запущен${NC}"
        echo "Запустите relay сервер перед тестированием:"
        echo "  docker-compose up -d"
        exit 1
    fi
    
    if ! netstat -tuln | grep -q ":51820"; then
        echo -e "${RED}❌ Порт 51820/udp не открыт${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Relay Server готов к тестированию${NC}"
    echo ""
}

# Создание tmux сессии
create_tmux_session() {
    echo -e "${YELLOW}🖥️  Создание tmux сессии...${NC}"
    
    # Убиваем существующую сессию если есть
    tmux kill-session -t p2p-test 2>/dev/null
    
    # Создаем новую сессию
    tmux new-session -d -s p2p-test -n "relay-monitor"
    
    # Создаем окна для каждого терминала
    tmux new-window -t p2p-test -n "server-a"
    tmux new-window -t p2p-test -n "server-b"
    
    # Настраиваем layout
    tmux select-window -t p2p-test:relay-monitor
    
    echo -e "${GREEN}✅ Tmux сессия создана${NC}"
    echo ""
}

# Запуск терминалов
start_terminals() {
    echo -e "${YELLOW}🚀 Запуск терминалов...${NC}"
    
    # Делаем скрипты исполняемыми
    chmod +x terminal1_relay_monitor.sh
    chmod +x terminal2_server_a.sh
    chmod +x terminal3_server_b.sh
    
    # Запускаем терминал 1 (Relay Monitor)
    tmux send-keys -t p2p-test:relay-monitor "./terminal1_relay_monitor.sh" Enter
    
    # Запускаем терминал 2 (Server A)
    tmux send-keys -t p2p-test:server-a "./terminal2_server_a.sh" Enter
    
    # Запускаем терминал 3 (Server B)
    tmux send-keys -t p2p-test:server-b "./terminal3_server_b.sh" Enter
    
    echo -e "${GREEN}✅ Все терминалы запущены${NC}"
    echo ""
}

# Отображение инструкций
show_instructions() {
    echo -e "${BLUE}📋 Инструкции по использованию:${NC}"
    echo "============================================="
    echo ""
    echo -e "${YELLOW}🖥️  Терминал 1 (Relay Monitor):${NC}"
    echo "  - Мониторинг логов P2P mesh"
    echo "  - Отслеживание UDP трафика"
    echo "  - Просмотр метрик"
    echo "  - Анализ соединений"
    echo ""
    echo -e "${YELLOW}🖥️  Терминал 2 (Server A):${NC}"
    echo "  - Отправка данных на порт 51820/udp"
    echo "  - WireGuard handshake"
    echo "  - P2P discovery"
    echo "  - Тестовые сообщения"
    echo ""
    echo -e "${YELLOW}🖥️  Терминал 3 (Server B):${NC}"
    echo "  - Получение данных с порта 51821/udp"
    echo "  - Автоматические ответы"
    echo "  - Обработка различных типов сообщений"
    echo ""
    echo -e "${GREEN}🎮 Управление tmux:${NC}"
    echo "  Ctrl+b c     - Создать новое окно"
    echo "  Ctrl+b n     - Следующее окно"
    echo "  Ctrl+b p     - Предыдущее окно"
    echo "  Ctrl+b 0-9   - Переключиться на окно по номеру"
    echo "  Ctrl+b d     - Отключиться от сессии"
    echo "  tmux attach -t p2p-test  - Подключиться к сессии"
    echo ""
    echo -e "${YELLOW}🧪 Сценарии тестирования:${NC}"
    echo "1. В терминале 2 отправьте WireGuard handshake"
    echo "2. В терминале 3 запустите прослушивание"
    echo "3. В терминале 1 наблюдайте за логами"
    echo "4. Отправьте тестовые данные и проверьте ответы"
    echo ""
}

# Главная функция
main() {
    echo -e "${BLUE}🚀 Запуск трех терминалов для тестирования P2P Mesh${NC}"
    echo "========================================================"
    echo ""
    
    # Проверяем зависимости
    check_dependencies
    
    # Проверяем статус relay сервера
    check_relay_status
    
    # Создаем tmux сессию
    create_tmux_session
    
    # Запускаем терминалы
    start_terminals
    
    # Показываем инструкции
    show_instructions
    
    # Подключаемся к сессии
    echo -e "${GREEN}🎉 Готово! Подключение к tmux сессии...${NC}"
    echo ""
    tmux attach-session -t p2p-test
}

# Обработка аргументов командной строки
case "${1:-}" in
    "help"|"-h"|"--help")
        echo "Использование: $0 [команда]"
        echo ""
        echo "Команды:"
        echo "  (без аргументов) - Запустить все терминалы"
        echo "  help             - Показать эту справку"
        echo "  status           - Проверить статус relay сервера"
        echo "  cleanup          - Очистить tmux сессии"
        echo ""
        exit 0
        ;;
    "status")
        check_dependencies
        check_relay_status
        echo -e "${GREEN}✅ Все проверки пройдены${NC}"
        exit 0
        ;;
    "cleanup")
        echo -e "${YELLOW}🧹 Очистка tmux сессий...${NC}"
        tmux kill-session -t p2p-test 2>/dev/null
        echo -e "${GREEN}✅ Очистка завершена${NC}"
        exit 0
        ;;
    *)
        main
        ;;
esac


