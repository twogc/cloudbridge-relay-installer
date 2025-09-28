#!/bin/bash
# Скрипт копирования основного кода в appliance

set -e

# Конфигурация
SOURCE_DIR="/home/ubuntu/cloudbridge-relay-installer"
TARGET_DIR="/home/ubuntu/cloudbridge-relay-installer/appliance"

# Цвета для вывода
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Создать структуру src в appliance
log_info "Создание структуры src в appliance..."
mkdir -p "$TARGET_DIR/src"

# Копировать основной код
log_info "Копирование основного кода..."

# Копировать relay сервер
if [[ -d "$SOURCE_DIR/cmd" ]]; then
    cp -r "$SOURCE_DIR/cmd" "$TARGET_DIR/src/"
    log_success "Скопирован cmd/ (relay сервер)"
fi

# Копировать internal
if [[ -d "$SOURCE_DIR/internal" ]]; then
    cp -r "$SOURCE_DIR/internal" "$TARGET_DIR/src/"
    log_success "Скопирован internal/ (внутренние компоненты)"
fi

# Копировать pkg
if [[ -d "$SOURCE_DIR/pkg" ]]; then
    cp -r "$SOURCE_DIR/pkg" "$TARGET_DIR/src/"
    log_success "Скопирован pkg/ (пакеты)"
fi

# Копировать конфигурации
if [[ -d "$SOURCE_DIR/config" ]]; then
    cp -r "$SOURCE_DIR/config" "$TARGET_DIR/src/"
    log_success "Скопирован config/ (конфигурации)"
fi

# Копировать клиент
if [[ -d "$SOURCE_DIR/test/relay-client" ]]; then
    cp -r "$SOURCE_DIR/test/relay-client" "$TARGET_DIR/src/"
    log_success "Скопирован relay-client/"
fi

# Копировать SaaS
if [[ -d "$SOURCE_DIR/test/cf-saas" ]]; then
    cp -r "$SOURCE_DIR/test/cf-saas" "$TARGET_DIR/src/"
    log_success "Скопирован cf-saas/"
fi

# Копировать Kubernetes манифесты
if [[ -d "$SOURCE_DIR/k8s" ]]; then
    cp -r "$SOURCE_DIR/k8s" "$TARGET_DIR/src/"
    log_success "Скопирован k8s/ (Kubernetes манифесты)"
fi

# Копировать go.mod и другие файлы
if [[ -f "$SOURCE_DIR/go.mod" ]]; then
    cp "$SOURCE_DIR/go.mod" "$TARGET_DIR/src/"
    log_success "Скопирован go.mod"
fi

if [[ -f "$SOURCE_DIR/go.sum" ]]; then
    cp "$SOURCE_DIR/go.sum" "$TARGET_DIR/src/"
    log_success "Скопирован go.sum"
fi

# Копировать Makefile
if [[ -f "$SOURCE_DIR/Makefile" ]]; then
    cp "$SOURCE_DIR/Makefile" "$TARGET_DIR/src/"
    log_success "Скопирован Makefile"
fi

log_success "Основной код успешно скопирован в appliance/src/"
echo
echo "📁 Структура appliance:"
echo "├── src/                    # Основной код CloudBridge"
echo "│   ├── cmd/               # Relay сервер"
echo "│   ├── internal/          # Внутренние компоненты"
echo "│   ├── pkg/               # Пакеты"
echo "│   ├── config/           # Конфигурации"
echo "│   ├── test/relay-client/ # Клиент"
echo "│   ├── test/cf-saas/      # SaaS"
echo "│   └── k8s/               # Kubernetes"
echo "├── docker/                # Docker образы"
echo "├── api/                   # Management API"
echo "├── pilot/                 # Пилотная программа"
echo "├── configs/               # Appliance конфигурации"
echo "├── scripts/               # Скрипты управления"
echo "└── docs/                  # Документация"
