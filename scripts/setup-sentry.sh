#!/bin/bash

# CloudBridge Relay - Sentry Setup Script
# Этот скрипт помогает настроить Sentry для CloudBridge Relay

set -e

echo "🚀 CloudBridge Relay - Sentry Setup"
echo "=================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
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

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker не установлен. Установите Docker для запуска Sentry."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose не установлен. Установите Docker Compose."
        exit 1
    fi
    
    log_success "Все зависимости установлены"
}

# Создание .env файла
create_env_file() {
    log_info "Создание .env файла для Sentry..."
    
    cat > .env.sentry << EOF
# Sentry Configuration
SENTRY_DSN=https://your-dsn@sentry.io/project-id
SENTRY_ENVIRONMENT=development
SENTRY_RELEASE=cloudbridge-relay@1.0.0
SENTRY_DEBUG=true
SENTRY_SERVER_NAME=cloudbridge-relay

# Sentry Server Configuration
SENTRY_SECRET_KEY=your-secret-key-here
SENTRY_POSTGRES_PASSWORD=your-postgres-password
SENTRY_EMAIL_PASSWORD=your-email-password

# CloudBridge Configuration
CLOUDBRIDGE_TENANT_ID=default
CLOUDBRIDGE_USER_ID=admin
CLOUDBRIDGE_CONNECTION_ID=main
EOF

    log_success ".env.sentry файл создан"
    log_warning "Не забудьте обновить значения в .env.sentry файле!"
}

# Настройка Sentry проекта
setup_sentry_project() {
    log_info "Настройка Sentry проекта..."
    
    echo ""
    echo "📋 Инструкции по настройке Sentry:"
    echo "1. Перейдите на https://sentry.io"
    echo "2. Создайте новый проект 'CloudBridge Relay'"
    echo "3. Выберите платформу 'Go'"
    echo "4. Скопируйте DSN из настроек проекта"
    echo "5. Обновите SENTRY_DSN в .env.sentry файле"
    echo ""
    
    read -p "Нажмите Enter когда настроите Sentry проект..."
}

# Запуск Sentry локально
start_sentry_local() {
    log_info "Запуск Sentry локально..."
    
    if [ ! -f "docker-compose.sentry.yml" ]; then
        log_error "Файл docker-compose.sentry.yml не найден"
        exit 1
    fi
    
    # Запуск Sentry
    docker-compose -f docker-compose.sentry.yml up -d
    
    log_success "Sentry запущен локально"
    log_info "Sentry Web UI: http://localhost:9000"
    log_info "MailHog (SMTP): http://localhost:8025"
}

# Настройка переменных окружения
setup_environment() {
    log_info "Настройка переменных окружения..."
    
    # Загрузка .env файла если существует
    if [ -f ".env.sentry" ]; then
        source .env.sentry
        log_success "Переменные окружения загружены из .env.sentry"
    else
        log_warning ".env.sentry файл не найден. Создайте его вручную."
    fi
    
    # Экспорт переменных
    export SENTRY_DSN
    export SENTRY_ENVIRONMENT
    export SENTRY_RELEASE
    export SENTRY_DEBUG
    export SENTRY_SERVER_NAME
    
    log_success "Переменные окружения настроены"
}

# Тестирование интеграции
test_integration() {
    log_info "Тестирование Sentry интеграции..."
    
    # Создание тестового файла
    cat > test_sentry.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/getsentry/sentry-go"
    "github.com/twogc/cloudbridge-relay-installer/internal/sentry"
)

func main() {
    // Инициализация Sentry
    sentryConfig := sentry.DefaultConfig()
    if err := sentry.Init(sentryConfig); err != nil {
        log.Fatalf("Failed to initialize Sentry: %v", err)
    }
    defer sentry.Flush(2 * time.Second)
    
    // Тестовая ошибка
    fmt.Println("Отправка тестовой ошибки в Sentry...")
    sentry.CaptureMessage("CloudBridge Relay - Test Error")
    
    // Тестовая паника
    fmt.Println("Отправка тестовой паники в Sentry...")
    defer func() {
        if r := recover(); r != nil {
        sentry.CaptureException(fmt.Errorf("test panic: %v", r))
    }()
    
    panic("Test panic for Sentry")
}
EOF

    log_info "Запуск тестового скрипта..."
    go run test_sentry.go
    
    # Удаление тестового файла
    rm test_sentry.go
    
    log_success "Тест Sentry интеграции завершен"
    log_info "Проверьте Sentry dashboard для подтверждения получения событий"
}

# Создание systemd сервиса
create_systemd_service() {
    log_info "Создание systemd сервиса для Sentry..."
    
    cat > /tmp/cloudbridge-sentry.service << EOF
[Unit]
Description=CloudBridge Relay with Sentry
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/cloudbridge-relay-installer
Environment=SENTRY_DSN=\${SENTRY_DSN}
Environment=SENTRY_ENVIRONMENT=\${SENTRY_ENVIRONMENT}
Environment=SENTRY_RELEASE=\${SENTRY_RELEASE}
ExecStart=/home/ubuntu/cloudbridge-relay-installer/cloudbridge-relay
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    log_success "Systemd сервис создан в /tmp/cloudbridge-sentry.service"
    log_info "Для установки выполните:"
    log_info "sudo cp /tmp/cloudbridge-sentry.service /etc/systemd/system/"
    log_info "sudo systemctl daemon-reload"
    log_info "sudo systemctl enable cloudbridge-sentry"
    log_info "sudo systemctl start cloudbridge-sentry"
}

# Основное меню
show_menu() {
    echo ""
    echo "🔧 Sentry Setup Menu:"
    echo "1. Проверить зависимости"
    echo "2. Создать .env файл"
    echo "3. Настроить Sentry проект"
    echo "4. Запустить Sentry локально"
    echo "5. Настроить переменные окружения"
    echo "6. Тестировать интеграцию"
    echo "7. Создать systemd сервис"
    echo "8. Выполнить все шаги"
    echo "0. Выход"
    echo ""
}

# Выполнение всех шагов
run_all_steps() {
    log_info "Выполнение всех шагов настройки Sentry..."
    
    check_dependencies
    create_env_file
    setup_sentry_project
    start_sentry_local
    setup_environment
    test_integration
    create_systemd_service
    
    log_success "Все шаги выполнены!"
    echo ""
    echo "🎉 Sentry интеграция готова!"
    echo "📊 Sentry Dashboard: http://localhost:9000"
    echo "📧 MailHog: http://localhost:8025"
    echo "🔧 Не забудьте обновить DSN в .env.sentry файле"
}

# Главная функция
main() {
    while true; do
        show_menu
        read -p "Выберите опцию (0-8): " choice
        
        case $choice in
            1) check_dependencies ;;
            2) create_env_file ;;
            3) setup_sentry_project ;;
            4) start_sentry_local ;;
            5) setup_environment ;;
            6) test_integration ;;
            7) create_systemd_service ;;
            8) run_all_steps ;;
            0) 
                log_info "Выход из скрипта"
                exit 0
                ;;
            *)
                log_error "Неверный выбор. Попробуйте снова."
                ;;
        esac
        
        echo ""
        read -p "Нажмите Enter для продолжения..."
    done
}

# Запуск скрипта
main "$@"



