#!/bin/bash

# 🚀 Скрипт развертывания QUIC + ICE/STUN/TURN на продакшене
# Использование: ./deploy-to-production.sh [SERVER_IP]

set -euo pipefail

# Конфигурация
SERVER_IP="${1:-95.163.250.190}"
SERVER_USER="root"
PROJECT_DIR="/opt/cloudbridge-relay-installer"
SERVICE_NAME="cloudbridge-relay"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для вывода
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка подключения к серверу
check_server_connection() {
    log_info "Проверка подключения к серверу $SERVER_IP..."
    
    if ! ssh -o ConnectTimeout=10 -o BatchMode=yes $SERVER_USER@$SERVER_IP "echo 'Connection successful'" > /dev/null 2>&1; then
        log_error "Не удается подключиться к серверу $SERVER_IP"
        log_info "Убедитесь, что:"
        log_info "1. SSH ключ настроен для $SERVER_USER@$SERVER_IP"
        log_info "2. Сервер доступен по сети"
        log_info "3. SSH сервис запущен"
        exit 1
    fi
    
    log_success "Подключение к серверу установлено"
}

# Проверка текущего состояния сервера
check_current_status() {
    log_info "Проверка текущего состояния сервера..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        echo "=== Проверка запущенных сервисов ==="
        systemctl status cloudbridge-relay 2>/dev/null || echo "Сервис cloudbridge-relay не найден"
        
        echo "=== Проверка Docker контейнеров ==="
        docker ps -a | grep relay || echo "Docker контейнеры relay не найдены"
        
        echo "=== Проверка открытых портов ==="
        netstat -tlnp | grep -E ":(8080|8082|19302|9090|8443|9092)" || echo "Порты не открыты"
        
        echo "=== Проверка UFW статуса ==="
        ufw status || echo "UFW не настроен"
EOF
}

# Остановка старых сервисов
stop_old_services() {
    log_info "Остановка старых сервисов..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        # Остановить systemd сервис
        if systemctl is-active --quiet cloudbridge-relay; then
            echo "Останавливаем systemd сервис..."
            systemctl stop cloudbridge-relay
        fi
        
        # Остановить Docker контейнеры
        if docker ps -q --filter "name=relay" | grep -q .; then
            echo "Останавливаем Docker контейнеры..."
            docker stop $(docker ps -q --filter "name=relay")
        fi
        
        echo "Старые сервисы остановлены"
EOF
    
    log_success "Старые сервисы остановлены"
}

# Обновление кода на сервере
update_code() {
    log_info "Обновление кода на сервере..."
    
    ssh $SERVER_USER@$SERVER_IP << EOF
        # Создать директорию если не существует
        mkdir -p $PROJECT_DIR
        
        # Перейти в директорию
        cd $PROJECT_DIR
        
        # Клонировать или обновить репозиторий
        if [ -d ".git" ]; then
            echo "Обновляем существующий репозиторий..."
            git fetch origin
            git reset --hard origin/main
        else
            echo "Клонируем репозиторий..."
            git clone https://github.com/twogc/cloudbridge-relay-installer.git .
        fi
        
        # Проверить коммит
        echo "Текущий коммит:"
        git log --oneline -1
        
        # Проверить, что это правильный коммит
        if ! git log --oneline -1 | grep -q "341fa1f6"; then
            echo "ВНИМАНИЕ: Коммит не соответствует ожидаемому!"
            echo "Ожидается: 341fa1f6 🚀 Implement QUIC + ICE/STUN/TURN P2P architecture"
        fi
EOF
    
    log_success "Код обновлен"
}

# Установка Go 1.25
install_go() {
    log_info "Установка Go 1.25..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        # Проверить текущую версию Go
        if command -v go > /dev/null; then
            CURRENT_VERSION=$(go version | cut -d' ' -f3)
            echo "Текущая версия Go: $CURRENT_VERSION"
            
            if [[ "$CURRENT_VERSION" == "go1.25"* ]]; then
                echo "Go 1.25 уже установлен"
                exit 0
            fi
        fi
        
        # Установить Go 1.25
        echo "Устанавливаем Go 1.25..."
        wget -q https://go.dev/dl/go1.25.linux-amd64.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
        rm go1.25.linux-amd64.tar.gz
        
        # Добавить в PATH
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        export PATH=$PATH:/usr/local/go/bin
        
        # Проверить установку
        go version
EOF
    
    log_success "Go 1.25 установлен"
}

# Настройка окружения
setup_environment() {
    log_info "Настройка окружения..."
    
    ssh $SERVER_USER@$SERVER_IP << EOF
        cd $PROJECT_DIR
        
        # Создать .env файл
        cat > .env << 'ENVEOF'
RELAY_JWT_SECRET=fallback-key
RELAY_HTTP_PORT=8080
RELAY_P2P_PORT=8082
RELAY_QUIC_PORT=9090
RELAY_STUN_PORT=19302
RELAY_MASQUE_PORT=8443
RELAY_ENHANCED_QUIC_PORT=9092
ENVEOF
        
        echo "Файл .env создан"
        cat .env
EOF
    
    log_success "Окружение настроено"
}

# Сборка приложения
build_application() {
    log_info "Сборка приложения..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        cd /opt/cloudbridge-relay-installer
        
        # Установить PATH для Go
        export PATH=$PATH:/usr/local/go/bin
        
        # Собрать приложение
        echo "Собираем приложение..."
        go build -o cloudbridge-relay ./cmd/relay
        
        # Проверить результат
        if [ -f "cloudbridge-relay" ]; then
            echo "Приложение собрано успешно"
            ls -la cloudbridge-relay
            chmod +x cloudbridge-relay
        else
            echo "ОШИБКА: Приложение не собрано!"
            exit 1
        fi
EOF
    
    log_success "Приложение собрано"
}

# Настройка systemd сервиса
setup_systemd() {
    log_info "Настройка systemd сервиса..."
    
    ssh $SERVER_USER@$SERVER_IP << EOF
        # Создать systemd unit файл
        sudo cat > /etc/systemd/system/$SERVICE_NAME.service << 'UNITEOF'
[Unit]
Description=CloudBridge Relay Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$PROJECT_DIR
EnvironmentFile=$PROJECT_DIR/.env
ExecStart=$PROJECT_DIR/cloudbridge-relay
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
UNITEOF
        
        # Перезагрузить systemd
        sudo systemctl daemon-reload
        sudo systemctl enable $SERVICE_NAME
        
        echo "Systemd сервис настроен"
EOF
    
    log_success "Systemd сервис настроен"
}

# Настройка UFW firewall
setup_firewall() {
    log_info "Настройка UFW firewall..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        # Очистить старые правила
        sudo ufw --force reset
        
        # Разрешить SSH
        sudo ufw allow 22
        
        # Разрешить Zabbix
        sudo ufw allow 10050
        
        # Разрешить новые порты
        sudo ufw allow 8080/tcp  # HTTP API
        sudo ufw allow 8082/tcp  # P2P API
        sudo ufw allow 19302/udp # STUN Server
        sudo ufw allow 9090/udp  # QUIC Transport
        sudo ufw allow 8443/tcp  # MASQUE Proxy
        sudo ufw allow 9092/udp  # Enhanced QUIC
        
        # Включить firewall
        sudo ufw --force enable
        
        # Проверить статус
        sudo ufw status
EOF
    
    log_success "UFW firewall настроен"
}

# Запуск сервиса
start_service() {
    log_info "Запуск сервиса..."
    
    ssh $SERVER_USER@$SERVER_IP << EOF
        # Запустить сервис
        sudo systemctl start $SERVICE_NAME
        
        # Подождать немного
        sleep 5
        
        # Проверить статус
        sudo systemctl status $SERVICE_NAME --no-pager
        
        # Проверить порты
        echo "Проверка открытых портов:"
        netstat -tlnp | grep -E ":(8080|8082|19302|9090|8443|9092)" || echo "Порты не открыты"
EOF
    
    log_success "Сервис запущен"
}

# Тестирование endpoints
test_endpoints() {
    log_info "Тестирование endpoints..."
    
    # Тест HTTP API
    log_info "Тестирование HTTP API (8080)..."
    if curl -k -s --connect-timeout 10 "https://edge.2gc.ru:8080/health" | grep -q "ok"; then
        log_success "HTTP API работает"
    else
        log_warning "HTTP API не отвечает"
    fi
    
    # Тест P2P API
    log_info "Тестирование P2P API (8082)..."
    if curl -k -s --connect-timeout 10 "https://edge.2gc.ru:8082/health" | grep -q "ok"; then
        log_success "P2P API работает"
    else
        log_warning "P2P API не отвечает"
    fi
    
    # Тест STUN сервера
    log_info "Тестирование STUN сервера (19302)..."
    if echo -e "\x00\x01\x00\x00\x21\x12\xa4\x42\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" | nc -u -w 3 edge.2gc.ru 19302 | wc -c | grep -q "20"; then
        log_success "STUN сервер работает"
    else
        log_warning "STUN сервер не отвечает"
    fi
}

# Основная функция
main() {
    log_info "🚀 Начинаем развертывание QUIC + ICE/STUN/TURN на продакшене"
    log_info "Сервер: $SERVER_IP"
    log_info "Пользователь: $SERVER_USER"
    log_info "Директория: $PROJECT_DIR"
    
    # Проверки
    check_server_connection
    check_current_status
    
    # Подтверждение
    echo
    log_warning "ВНИМАНИЕ: Это развертывание заменит текущий relay сервер!"
    read -p "Продолжить? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Развертывание отменено"
        exit 0
    fi
    
    # Развертывание
    stop_old_services
    update_code
    install_go
    setup_environment
    build_application
    setup_systemd
    setup_firewall
    start_service
    
    # Тестирование
    test_endpoints
    
    log_success "🎉 Развертывание завершено!"
    log_info "Проверьте логи: ssh $SERVER_USER@$SERVER_IP 'journalctl -u $SERVICE_NAME -f'"
    log_info "Проверьте статус: ssh $SERVER_USER@$SERVER_IP 'systemctl status $SERVICE_NAME'"
}

# Запуск
main "$@"
