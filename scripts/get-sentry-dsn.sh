#!/bin/bash

# ===========================================
# Get Sentry DSN Script
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
    echo "🔍 Получение правильного DSN из Sentry"
    echo "====================================="
    echo ""
    echo "Для получения правильного DSN:"
    echo ""
    echo "1. Откройте Sentry Dashboard: https://sentry.io/"
    echo "2. Перейдите в ваш проект Go"
    echo "3. Откройте Settings → Projects → [Ваш проект]"
    echo "4. Перейдите в раздел 'Client Keys (DSN)'"
    echo "5. Скопируйте DSN (он должен выглядеть так):"
    echo "   https://YOUR_KEY@YOUR_ORG.ingest.sentry.io/PROJECT_ID"
    echo ""
    echo "6. Убедитесь, что проект активен и имеет правильные настройки"
    echo ""
    echo "Текущий DSN:"
    source scripts/load-env.sh 2>/dev/null || true
    echo "SENTRY_DSN: $SENTRY_DSN"
    echo ""
    echo "Если DSN неверный, обновите его в .env файле"
    echo ""
    echo "Альтернативно, создайте новый проект:"
    echo "1. Создайте новый проект в Sentry"
    echo "2. Выберите платформу 'Go'"
    echo "3. Скопируйте новый DSN"
    echo "4. Обновите .env файл"
}

main "$@"



