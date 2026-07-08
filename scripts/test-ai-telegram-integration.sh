#!/bin/bash

# Скрипт для тестирования интеграции AI отчетов в Telegram
# Использование: ./scripts/test-ai-telegram-integration.sh

set -e

echo "🤖 Тестирование интеграции AI отчетов в Telegram"
echo "=================================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для логирования
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Проверяем доступность AI сервиса
log "Проверка доступности AI сервиса..."
if curl -s -f "http://localhost:8083/health" > /dev/null; then
    success "AI сервис доступен"
else
    error "AI сервис недоступен на localhost:8083"
    log "Попробуйте запустить AI сервис или изменить URL"
    exit 1
fi

# Проверяем статус Telegram
log "Проверка статуса Telegram уведомлений..."
TELEGRAM_STATUS=$(curl -s "http://localhost:8083/api/v1/ai/telegram/status" | jq -r '.telegram.enabled // false')
if [ "$TELEGRAM_STATUS" = "true" ]; then
    success "Telegram уведомления настроены"
else
    warning "Telegram уведомления не настроены"
    log "Убедитесь, что установлены переменные TELEGRAM_BOT_TOKEN и TELEGRAM_CHAT_ID"
fi

# Тест 1: Генерация отчета с Cloudflare AI
log "Тест 1: Генерация отчета с Cloudflare AI..."
RESPONSE1=$(curl -s -X POST "http://localhost:8083/api/v1/ai/generate-report" \
    -H "Content-Type: application/json" \
    -d '{
        "report_type": "telegram",
        "format": "professional",
        "language": "russian",
        "use_cloudflare_ai": true,
        "include_metrics": true,
        "include_analysis": true,
        "include_recommendations": true,
        "include_forecast": true
    }')

if echo "$RESPONSE1" | jq -e '.message' > /dev/null; then
    success "Отчет с Cloudflare AI сгенерирован успешно"
    echo "$RESPONSE1" | jq '.message'
else
    warning "Ошибка при генерации отчета с Cloudflare AI"
    echo "$RESPONSE1"
fi

# Тест 2: Генерация отчета с fallback
log "Тест 2: Генерация отчета с fallback (без Cloudflare AI)..."
RESPONSE2=$(curl -s -X POST "http://localhost:8083/api/v1/ai/generate-report" \
    -H "Content-Type: application/json" \
    -d '{
        "report_type": "telegram",
        "format": "professional",
        "language": "russian",
        "use_cloudflare_ai": false,
        "include_metrics": true,
        "include_analysis": true,
        "include_recommendations": true,
        "include_forecast": true
    }')

if echo "$RESPONSE2" | jq -e '.message' > /dev/null; then
    success "Отчет с fallback сгенерирован успешно"
    echo "$RESPONSE2" | jq '.message'
else
    error "Ошибка при генерации отчета с fallback"
    echo "$RESPONSE2"
fi

# Тест 3: Отправка произвольного сообщения
log "Тест 3: Отправка произвольного сообщения в Telegram..."
RESPONSE3=$(curl -s -X POST "http://localhost:8083/api/v1/ai/telegram/send" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "🧪 Тестовое сообщение от CloudBridge AI Integration Test"
  }')

if echo "$RESPONSE3" | jq -e '.message' > /dev/null; then
    success "Тестовое сообщение отправлено"
    echo "$RESPONSE3" | jq '.message'
else
    warning "Ошибка при отправке тестового сообщения"
    echo "$RESPONSE3"
fi

# Тест 4: Проверка метрик
log "Тест 4: Проверка метрик системы..."
METRICS_RESPONSE=$(curl -s "http://localhost:8083/api/v1/ai/metrics")
if echo "$METRICS_RESPONSE" | jq -e '.metrics' > /dev/null; then
    success "Метрики получены успешно"
    echo "$METRICS_RESPONSE" | jq '.metrics | keys'
else
    warning "Ошибка при получении метрик"
    echo "$METRICS_RESPONSE"
fi

# Тест 5: Проверка статуса AI сервиса
log "Тест 5: Проверка статуса AI сервиса..."
STATUS_RESPONSE=$(curl -s "http://localhost:8083/api/v1/ai/status")
if echo "$STATUS_RESPONSE" | jq -e '.status' > /dev/null; then
    success "Статус AI сервиса получен"
    echo "$STATUS_RESPONSE" | jq '.status'
else
    warning "Ошибка при получении статуса AI сервиса"
    echo "$STATUS_RESPONSE"
fi

echo ""
echo "Тестирование завершено!"
echo ""
echo "Сводка:"
echo "- AI сервис: $(curl -s "http://localhost:8083/health" | jq -r '.status // "недоступен"')"
echo "- Telegram: $TELEGRAM_STATUS"
echo "- Cloudflare AI: $(if [ -n "$CLOUDFLARE_AI_TOKEN" ]; then echo "настроен"; else echo "не настроен"; fi)"
echo ""
echo " Для настройки Cloudflare AI установите переменные:"
echo "   export CLOUDFLARE_AI_TOKEN=your_token"
echo "   export CLOUDFLARE_ACCOUNT_ID=your_account_id"
echo ""
echo " Для настройки Telegram установите переменные:"
echo "   export TELEGRAM_BOT_TOKEN=your_bot_token"
echo "   export TELEGRAM_CHAT_ID=your_chat_id"
