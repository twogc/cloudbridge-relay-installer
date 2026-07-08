#!/bin/bash
# CloudBridge Relay — interactive setup (Zitadel IAM)
# Keycloak flow removed; canonical auth is Zitadel.

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

show_banner() {
    echo -e "\033[1;36m"
    echo "╔═══════════════════════════════════════════════════════════════════════════╗"
    echo "║              CloudBridge Relay — Interactive Setup (Zitadel)              ║"
    echo "╚═══════════════════════════════════════════════════════════════════════════╝"
    echo -e "\033[0m"
}

check_dependencies() {
    log "Проверка зависимостей..."
    command -v docker >/dev/null || { error "Docker не установлен"; exit 1; }
    command -v docker-compose >/dev/null || command -v docker >/dev/null || {
        error "Docker Compose не установлен"; exit 1
    }
    log "Зависимости OK"
}

create_directories() {
    mkdir -p logs certs backups config
    chmod 755 logs backups config
    chmod 700 certs
    log "Директории созданы"
}

generate_env() {
    if [[ ! -f .env ]]; then
        if [[ -x ./scripts/setup/generate_secrets.sh ]]; then
            ./scripts/setup/generate_secrets.sh --env .env
            log ".env создан"
        else
            warn "generate_secrets.sh не найден — создайте .env вручную (ZITADEL_*)"
        fi
    else
        log ".env уже существует"
    fi
}

show_zitadel_hints() {
    echo -e "${BLUE}IAM: Zitadel (Keycloak больше не используется)${NC}"
    echo "  1. Задайте ZITADEL_DOMAIN / ZITADEL_CLIENT_ID / ZITADEL_CLIENT_SECRET в .env"
    echo "  2. Конфиг: configs/config.json (секция zitadel)"
    echo "  3. Документация: openwiki/architecture.md, docs/architecture/"
    echo
}

start_services() {
    if [[ -x ./scripts/utils/start_production.sh ]]; then
        log "Запуск production-окружения..."
        ./scripts/utils/start_production.sh "$@"
    else
        warn "start_production.sh не найден — запустите сервисы вручную"
    fi
}

main() {
    show_banner
    check_dependencies
    create_directories
    generate_env
    show_zitadel_hints
    echo -e "${YELLOW}Нажмите Enter для запуска сервисов (или Ctrl+C)...${NC}"
    read -r
    start_services
    log "Setup завершён. Auth provider: Zitadel."
}

main "$@"
