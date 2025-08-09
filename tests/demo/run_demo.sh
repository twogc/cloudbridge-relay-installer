#!/bin/bash

# Главный скрипт для запуска демонстрации P2P Mesh
# Запускает все три скрипта в отдельных терминалах

echo "🚀 ДЕМОНСТРАЦИЯ P2P Mesh с Prometheus"
echo "===================================="
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
    
    for cmd in docker nc jq curl netstat tmux; do
        if ! command -v $cmd &> /dev/null; then
            missing_deps+=($cmd)
        fi
    done
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo -e "${RED}❌ Отсутствуют зависимости: ${missing_deps[*]}${NC}"
        echo "Установите недостающие пакеты:"
        echo "  sudo apt update && sudo apt install -y docker.io netcat-openbsd jq curl net-tools tmux"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Все зависимости установлены${NC}"
    echo ""
}

# Проверка сервисов
check_services() {
    echo -e "${YELLOW}🔍 Проверка сервисов...${NC}"
    
    # Проверяем relay сервер
    if ! docker ps | grep -q cloudbridge-relay; then
        echo -e "${RED}❌ Relay сервер не запущен${NC}"
        echo "Запустите relay сервер:"
        echo "  docker-compose up -d"
        exit 1
    fi
    
    # Проверяем Prometheus
    if ! curl -s http://localhost:9090/api/v1/status/config >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Prometheus недоступен${NC}"
    else
        echo -e "${GREEN}✅ Prometheus доступен${NC}"
    fi
    
    # Проверяем порт 51820
    if ! netstat -tuln | grep -q ":51820"; then
        echo -e "${RED}❌ Порт 51820/udp не открыт${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ Все сервисы готовы${NC}"
    echo ""
}

# Создание tmux сессии
create_demo_session() {
    echo -e "${YELLOW}🖥️  Создание демо сессии...${NC}"
    
    # Убиваем существующую сессию
    tmux kill-session -t p2p-demo 2>/dev/null
    
    # Создаем новую сессию
    tmux new-session -d -s p2p-demo -n "monitor"
    
    # Создаем окна для каждого скрипта
    tmux new-window -t p2p-demo -n "server-a"
    tmux new-window -t p2p-demo -n "server-b"
    
    # Настраиваем layout
    tmux select-window -t p2p-demo:monitor
    
    echo -e "${GREEN}✅ Tmux сессия создана${NC}"
    echo ""
}

# Запуск скриптов
start_demo() {
    echo -e "${YELLOW}🚀 Запуск демонстрации...${NC}"
    
    # Делаем скрипты исполняемыми
    chmod +x demo_monitor.sh demo_server_a.sh demo_server_b.sh
    
    # Запускаем мониторинг
    tmux send-keys -t p2p-demo:monitor "./demo_monitor.sh" Enter
    
    # Запускаем Server A
    tmux send-keys -t p2p-demo:server-a "./demo_server_a.sh" Enter
    
    # Запускаем Server B
    tmux send-keys -t p2p-demo:server-b "./demo_server_b.sh" Enter
    
    echo -e "${GREEN}✅ Все скрипты запущены${NC}"
    echo ""
}

# Показать инструкции
show_instructions() {
    echo -e "${BLUE}📋 Инструкции по демонстрации:${NC}"
    echo "========================================="
    echo ""
    echo -e "${YELLOW}🖥️  Окна tmux:${NC}"
    echo "  monitor  - Мониторинг с Prometheus метриками"
    echo "  server-a - Отправка данных (Server A)"
    echo "  server-b - Получение и ответы (Server B)"
    echo ""
    echo -e "${YELLOW}🎮 Управление tmux:${NC}"
    echo "  Ctrl+b n     - Следующее окно"
    echo "  Ctrl+b p     - Предыдущее окно"
    echo "  Ctrl+b 0-2   - Переключиться на окно по номеру"
    echo "  Ctrl+b d     - Отключиться от сессии"
    echo "  tmux attach -t p2p-demo  - Подключиться к сессии"
    echo ""
    echo -e "${YELLOW}🧪 Сценарий демонстрации:${NC}"
    echo "1. В окне 'monitor' выберите '3) Мониторинг Prometheus метрик'"
    echo "2. В окне 'server-a' выберите '7) Демонстрационный режим'"
    echo "3. В окне 'server-b' выберите '3) Демонстрационный режим'"
    echo "4. Наблюдайте за обменом данными и метриками"
    echo ""
    echo -e "${YELLOW}📊 Что демонстрируется:${NC}"
    echo "• UDP обмен данными на порту 51820/udp"
    echo "• JSON сообщения между серверами"
    echo "• WireGuard handshake симуляция"
    echo "• P2P discovery механизм"
    echo "• Prometheus метрики в реальном времени"
    echo "• Автоматические ответы на сообщения"
    echo ""
}

# Главная функция
main() {
    echo -e "${BLUE}🚀 ДЕМОНСТРАЦИЯ P2P Mesh с Prometheus${NC}"
    echo "============================================="
    echo ""
    
    # Проверяем зависимости
    check_dependencies
    
    # Проверяем сервисы
    check_services
    
    # Создаем tmux сессию
    create_demo_session
    
    # Запускаем демо
    start_demo
    
    # Показываем инструкции
    show_instructions
    
    # Подключаемся к сессии
    echo -e "${GREEN}🎉 Готово! Подключение к демо сессии...${NC}"
    echo ""
    tmux attach-session -t p2p-demo
}

# Обработка аргументов
case "${1:-}" in
    "help"|"-h"|"--help")
        echo "Использование: $0 [команда]"
        echo ""
        echo "Команды:"
        echo "  (без аргументов) - Запустить полную демонстрацию"
        echo "  help             - Показать эту справку"
        echo "  status           - Проверить статус сервисов"
        echo "  cleanup          - Очистить tmux сессии"
        echo "  monitor          - Запустить только мониторинг"
        echo "  server-a         - Запустить только Server A"
        echo "  server-b         - Запустить только Server B"
        echo ""
        exit 0
        ;;
    "status")
        check_dependencies
        check_services
        echo -e "${GREEN}✅ Все проверки пройдены${NC}"
        exit 0
        ;;
    "cleanup")
        echo -e "${YELLOW}🧹 Очистка tmux сессий...${NC}"
        tmux kill-session -t p2p-demo 2>/dev/null
        echo -e "${GREEN}✅ Очистка завершена${NC}"
        exit 0
        ;;
    "monitor")
        chmod +x demo_monitor.sh
        ./demo_monitor.sh
        ;;
    "server-a")
        chmod +x demo_server_a.sh
        ./demo_server_a.sh
        ;;
    "server-b")
        chmod +x demo_server_b.sh
        ./demo_server_b.sh
        ;;
    *)
        main
        ;;
esac


