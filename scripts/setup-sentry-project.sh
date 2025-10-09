#!/bin/bash

# ===========================================
# Sentry Project Setup Script
# ===========================================

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для логирования
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка зависимостей
check_dependencies() {
    log "Проверка зависимостей..."
    
    if ! command -v curl &> /dev/null; then
        error "curl не установлен"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        error "jq не установлен"
        exit 1
    fi
    
    success "Все зависимости установлены"
}

# Создание нового проекта в Sentry
create_sentry_project() {
    log "Создание нового проекта в Sentry..."
    
    echo ""
    echo "🔧 Настройка Sentry проекта"
    echo "=========================="
    echo ""
    echo "Для создания нового проекта в Sentry:"
    echo "1. Откройте https://sentry.io/"
    echo "2. Войдите в свой аккаунт"
    echo "3. Создайте новую организацию (если нужно)"
    echo "4. Создайте новый проект:"
    echo "   - Выберите 'Go' как платформу"
    echo "   - Название: 'cloudbridge-relay'"
    echo "   - Описание: 'CloudBridge Relay Error Monitoring'"
    echo ""
    echo "5. После создания проекта скопируйте DSN"
    echo "6. DSN будет выглядеть примерно так:"
    echo "   https://YOUR_KEY@YOUR_ORG.ingest.sentry.io/PROJECT_ID"
    echo ""
    
    read -p "Введите новый DSN из Sentry: " NEW_DSN
    
    if [ -z "$NEW_DSN" ]; then
        error "DSN не может быть пустым"
        exit 1
    fi
    
    # Валидация DSN
    if [[ ! "$NEW_DSN" =~ ^https://[a-f0-9]+@[a-z0-9]+\.ingest\.sentry\.io/[0-9]+$ ]]; then
        error "Неверный формат DSN"
        exit 1
    fi
    
    success "DSN валиден: $NEW_DSN"
    
    # Обновление .env файла
    log "Обновление .env файла..."
    
    # Создаем резервную копию
    cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
    
    # Обновляем DSN
    if grep -q "SENTRY_DSN=" .env; then
        sed -i "s|SENTRY_DSN=.*|SENTRY_DSN=$NEW_DSN|" .env
    else
        echo "SENTRY_DSN=$NEW_DSN" >> .env
    fi
    
    success "DSN обновлен в .env файле"
}

# Тестирование нового DSN
test_new_dsn() {
    log "Тестирование нового DSN..."
    
    # Загружаем переменные окружения
    source scripts/load-env.sh
    
    # Создаем простой тест
    cat > test-new-dsn.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"time"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/twogc/cloudbridge-relay-installer/internal/sentry"
)

func main() {
	fmt.Println("🧪 Тестирование нового DSN...")
	
	// Инициализация Sentry
	sentryConfig := sentry.DefaultConfig()
	fmt.Printf("DSN: %s\n", sentryConfig.DSN)
	
	if err := sentry.Init(sentryConfig); err != nil {
		log.Fatalf("Failed to initialize Sentry: %v", err)
	}
	defer sentrygo.Flush(5 * time.Second)
	
	// Отправляем тестовое сообщение
	fmt.Println("Отправка тестового сообщения...")
	sentrygo.CaptureMessage("Test message from new DSN")
	
	// Ждем отправки
	time.Sleep(3 * time.Second)
	
	fmt.Println("✅ Тест завершен!")
}
EOF

    # Запускаем тест
    go run test-new-dsn.go
    
    # Удаляем тестовый файл
    rm test-new-dsn.go
    
    success "Тест нового DSN завершен"
}

# Основная функция
main() {
    echo "🚀 Sentry Project Setup"
    echo "======================="
    echo ""
    
    # Проверка зависимостей
    check_dependencies
    
    # Создание проекта
    create_sentry_project
    
    # Тестирование
    test_new_dsn
    
    echo ""
    success "Настройка Sentry проекта завершена!"
    echo ""
    echo "📋 Следующие шаги:"
    echo "   1. Проверьте Sentry dashboard"
    echo "   2. Убедитесь, что события поступают"
    echo "   3. Настройте алерты при необходимости"
    echo ""
    echo "🔗 Sentry Dashboard: https://sentry.io/"
}

# Запуск основной функции
main "$@"



