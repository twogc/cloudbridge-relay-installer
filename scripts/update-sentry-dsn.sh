#!/bin/bash

# ===========================================
# Update Sentry DSN Script
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

main() {
    echo "🔄 Обновление Sentry DSN"
    echo "======================="
    echo ""
    
    # Показываем текущий DSN
    log "Текущий DSN:"
    if [ -f ".env" ]; then
        grep "SENTRY_DSN=" .env || echo "DSN не найден в .env"
    else
        error ".env файл не найден"
        exit 1
    fi
    
    echo ""
    echo "📋 Инструкции:"
    echo "1. Откройте https://sentry.io/"
    echo "2. Создайте новый проект Go"
    echo "3. Скопируйте DSN из настроек проекта"
    echo "4. Введите новый DSN ниже"
    echo ""
    
    read -p "Введите новый DSN: " NEW_DSN
    
    if [ -z "$NEW_DSN" ]; then
        error "DSN не может быть пустым"
        exit 1
    fi
    
    # Валидация DSN
    if [[ ! "$NEW_DSN" =~ ^https://[a-f0-9]+@[a-z0-9]+\.ingest\.sentry\.io/[0-9]+$ ]]; then
        error "Неверный формат DSN"
        echo "Ожидаемый формат: https://KEY@HOST/PROJECT_ID"
        exit 1
    fi
    
    success "DSN валиден: $NEW_DSN"
    
    # Создаем резервную копию
    cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
    success "Создана резервная копия .env"
    
    # Обновляем DSN
    if grep -q "SENTRY_DSN=" .env; then
        sed -i "s|SENTRY_DSN=.*|SENTRY_DSN=$NEW_DSN|" .env
    else
        echo "SENTRY_DSN=$NEW_DSN" >> .env
    fi
    
    success "DSN обновлен в .env файле"
    
    # Тестируем новый DSN
    echo ""
    log "Тестирование нового DSN..."
    
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
	sentrygo.CaptureMessage("Test message from updated DSN")
	
	// Ждем отправки
	time.Sleep(3 * time.Second)
	
	fmt.Println("✅ Тест завершен!")
}
EOF

    # Запускаем тест
    source scripts/load-env.sh
    go run test-new-dsn.go
    
    # Удаляем тестовый файл
    rm test-new-dsn.go
    
    success "Тест нового DSN завершен"
    
    echo ""
    success "DSN успешно обновлен!"
    echo ""
    echo "📊 Проверьте Sentry dashboard для подтверждения получения событий"
    echo "🔗 URL: https://sentry.io/"
}

main "$@"



