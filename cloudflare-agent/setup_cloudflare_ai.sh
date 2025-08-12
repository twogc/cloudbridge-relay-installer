#!/bin/bash
set -e

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[INFO]${NC} $1"; }
err() { echo -e "${RED}[ERROR]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
info() { echo -e "${BLUE}[INFO]${NC} $1"; }

echo "🚀 Настройка Cloudflare AI для AI Monitoring Agent"
echo "=================================================="

# Проверка wrangler
if ! command -v wrangler &>/dev/null; then
  err "wrangler не установлен!"
  echo "Установите его: npm install -g wrangler"
  exit 1
fi

# Проверка аутентификации
if ! wrangler whoami &>/dev/null; then
  warn "Не авторизован в Cloudflare!"
  echo "Выполните: wrangler login"
  exit 1
fi

log "Проверяю доступные AI модели..."
wrangler ai list

echo ""
info "Для настройки Cloudflare AI выполните следующие шаги:"
echo ""
echo "1. 📝 Получите API токен:"
echo "   - Откройте https://dash.cloudflare.com/profile/api-tokens"
echo "   - Создайте новый токен с правами 'Cloudflare Workers AI:Edit'"
echo "   - Скопируйте токен (начинается с CF_)"
echo ""
echo "2. 🔐 Добавьте токен как секрет:"
echo "   wrangler secret put CLOUDFLARE_API_TOKEN --env production"
echo ""
echo "3. 🚀 Запустите деплой:"
echo "   ./start_production.sh"
echo ""

# Проверка существующих секретов
log "Проверяю существующие секреты..."
if wrangler secret list --env production 2>/dev/null | grep -q CLOUDFLARE_API_TOKEN; then
  log "✅ CLOUDFLARE_API_TOKEN уже настроен"
else
  warn "❌ CLOUDFLARE_API_TOKEN не найден"
fi

if wrangler secret list --env production 2>/dev/null | grep -q OPENAI_API_KEY; then
  log "✅ OPENAI_API_KEY найден (можно использовать как fallback)"
else
  warn "❌ OPENAI_API_KEY не найден"
fi

echo ""
log "Настройка завершена! Теперь можете запустить деплой." 