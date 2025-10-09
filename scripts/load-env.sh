#!/bin/bash

# CloudBridge Relay - Load Environment Variables
# Загружает переменные из .env файла

set -e

# Цвета для вывода
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Проверка существования .env файла
if [ ! -f ".env" ]; then
    echo "❌ Файл .env не найден"
    exit 1
fi

log_info "Загрузка переменных из .env файла..."

# Загрузка переменных из .env файла
export $(grep -v '^#' .env | grep -v '^$' | xargs)

log_success "Переменные окружения загружены"

# Проверка ключевых переменных Sentry
if [ -z "$SENTRY_DSN" ]; then
    echo "⚠️  SENTRY_DSN не установлен"
else
    log_success "SENTRY_DSN: ${SENTRY_DSN:0:20}..."
fi

if [ -z "$SENTRY_ENVIRONMENT" ]; then
    echo "⚠️  SENTRY_ENVIRONMENT не установлен"
else
    log_success "SENTRY_ENVIRONMENT: $SENTRY_ENVIRONMENT"
fi

if [ -z "$SENTRY_RELEASE" ]; then
    echo "⚠️  SENTRY_RELEASE не установлен"
else
    log_success "SENTRY_RELEASE: $SENTRY_RELEASE"
fi

echo ""
echo "🚀 Переменные окружения готовы к использованию!"
echo "💡 Для применения в текущей сессии выполните: source scripts/load-env.sh"



