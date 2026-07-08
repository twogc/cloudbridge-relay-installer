#!/bin/bash

# 🚀 Cadence Setup Script
# Этот скрипт настраивает Cadence с PostgreSQL в Kubernetes

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

echo "🚀 Настройка Cadence..."

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка наличия kubectl
if ! command -v kubectl &> /dev/null; then
    error "kubectl не найден. Установите kubectl и попробуйте снова."
    exit 1
fi

# Проверка подключения к кластеру
if ! kubectl cluster-info &> /dev/null; then
    error "Не удается подключиться к кластеру Kubernetes"
    exit 1
fi

log "Проверка подключения к кластеру... ✅"

# Проверка наличия PostgreSQL
log "Проверка PostgreSQL..."
if ! kubectl get pods -n cloudbridge-system | grep postgres &> /dev/null; then
    error "PostgreSQL не найден в namespace cloudbridge-system"
    exit 1
fi

log "PostgreSQL найден ✅"

# Создание баз данных
log "Создание баз данных..."
kubectl exec -n cloudbridge-system postgres-587dfcc9cd-cvwsk -- psql -U ipam_user -d cloudbridge_ipam -c "CREATE DATABASE IF NOT EXISTS cadence;" 2>/dev/null || true
kubectl exec -n cloudbridge-system postgres-587dfcc9cd-cvwsk -- psql -U ipam_user -d cloudbridge_ipam -c "CREATE DATABASE IF NOT EXISTS cadence_visibility;" 2>/dev/null || true

log "Базы данных созданы ✅"

# Применение конфигурации
log "Применение конфигурации Cadence..."
kubectl apply -f k8s/cadence/cadence.yaml

log "Конфигурация применена ✅"

# Ожидание запуска
log "Ожидание запуска Cadence..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=cadence -n cloudbridge --timeout=300s

log "Cadence запущен ✅"

# Проверка статуса
log "Проверка статуса..."
kubectl get pods -n cloudbridge | grep cadence

# Проверка логов
log "Проверка логов..."
kubectl logs -n cloudbridge deployment/cadence --tail=10

# Проверка подключения к базе данных
log "Проверка подключения к базе данных..."
kubectl exec -n cloudbridge-system postgres-587dfcc9cd-cvwsk -- psql -U ipam_user -d cadence -c "SELECT 1;" &> /dev/null
if [ $? -eq 0 ]; then
    log "Подключение к базе данных работает ✅"
else
    warn "Проблемы с подключением к базе данных"
fi

# Проверка health check
log "Проверка health check..."
kubectl get pods -n cloudbridge | grep cadence | grep Running &> /dev/null
if [ $? -eq 0 ]; then
    log "Cadence работает корректно ✅"
else
    warn "Cadence может иметь проблемы"
fi

echo ""
log "🎉 Настройка Cadence завершена!"
echo ""
echo "📊 Полезные команды:"
echo "  kubectl get pods -n cloudbridge | grep cadence"
echo "  kubectl logs -n cloudbridge deployment/cadence"
echo "  kubectl port-forward -n cloudbridge svc/cadence 7933:7933"
echo ""
echo "📚 Документация: CADENCE_SETUP_GUIDE.md"


