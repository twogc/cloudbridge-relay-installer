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

echo "🤖 Запуск ML API сервера для тестирования"
echo "=========================================="

# Проверка Python
if ! command -v python3 &>/dev/null; then
  err "Python 3 не установлен!"
  exit 1
fi

# Проверка pip
if ! command -v pip3 &>/dev/null; then
  err "pip3 не установлен!"
  exit 1
fi

# Установка зависимостей
log "Устанавливаю зависимости..."
pip3 install -r ml-api-requirements.txt

# Проверка порта
if lsof -Pi :8000 -sTCP:LISTEN -t >/dev/null ; then
  warn "Порт 8000 уже занят!"
  echo "Остановите процесс на порту 8000 или измените порт в ml-api-server.py"
  exit 1
fi

# Запуск сервера
log "Запускаю ML API сервер..."
python3 ml-api-server.py 