#!/bin/bash
set -e

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log() { echo -e "${GREEN}[INFO]${NC} $1"; }
err() { echo -e "${RED}[ERROR]${NC} $1"; }

# Проверка node
if ! command -v node &>/dev/null; then
  err "Node.js не установлен!"
  exit 1
fi
NODE_VERSION=$(node -v | cut -d. -f1 | tr -d v)
if [ "$NODE_VERSION" -lt 20 ]; then
  err "Node.js >= 20.x.x required!"
  exit 1
fi

# Проверка wrangler
if ! command -v wrangler &>/dev/null; then
  log "Устанавливаю wrangler..."
  npm install -g wrangler
fi

# Проверка обязательных секретов
for secret in ML_API_KEY SIEM_API_KEY; do
  if ! wrangler secret list --env production | grep -q $secret; then
    err "Секрет $secret не найден для production!"
    echo "Добавьте его командой: wrangler secret put $secret --env production"
    exit 1
  fi
done

# Проверка AI токенов
AI_TOKEN_FOUND=false

if wrangler secret list --env production | grep -q CLOUDFLARE_API_TOKEN; then
  log "CLOUDFLARE_API_TOKEN найден - будет использоваться Cloudflare AI"
  AI_TOKEN_FOUND=true
fi

if wrangler secret list --env production | grep -q OPENAI_API_KEY; then
  log "OPENAI_API_KEY найден - можно использовать OpenAI"
  AI_TOKEN_FOUND=true
fi

if [ "$AI_TOKEN_FOUND" = false ]; then
  err "Не найден ни CLOUDFLARE_API_TOKEN, ни OPENAI_API_KEY!"
  echo "Добавьте один из них:"
  echo "  wrangler secret put CLOUDFLARE_API_TOKEN --env production"
  echo "  wrangler secret put OPENAI_API_KEY --env production"
  exit 1
fi

log "Собираю проект..."
npm run build

log "Деплой в production..."
wrangler deploy --env production

log "Готово!"
echo "Production URL: https://cloudbridge-ai-monitoring-agent-prod.<your-subdomain>.workers.dev" 