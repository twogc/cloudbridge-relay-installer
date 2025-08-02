#!/bin/bash

# CloudBridge Relay - Restore Lynis Plugins
# Восстановление отключенных плагинов Lynis

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка прав root
if [[ $EUID -ne 0 ]]; then
    log_error "Этот скрипт должен быть запущен с правами root"
    exit 1
fi

# Пути к плагинам
PLUGINS_DIR="/usr/local/lynis/plugins"
DISABLED_DIR="/usr/local/lynis/plugins_disabled"

log_info "Восстановление отключенных плагинов Lynis"

# Проверяем существование директории с отключенными плагинами
if [[ ! -d "$DISABLED_DIR" ]]; then
    log_warn "Директория с отключенными плагинами не найдена: $DISABLED_DIR"
    exit 1
fi

# Восстанавливаем плагины
restored_count=0

if [[ -f "$DISABLED_DIR/plugin_systemd_phase1" ]]; then
    mv "$DISABLED_DIR/plugin_systemd_phase1" "$PLUGINS_DIR/"
    log_info "Восстановлен плагин: systemd"
    ((restored_count++))
fi

if [[ -f "$DISABLED_DIR/plugin_pam_phase1" ]]; then
    mv "$DISABLED_DIR/plugin_pam_phase1" "$PLUGINS_DIR/"
    log_info "Восстановлен плагин: pam"
    ((restored_count++))
fi

if [[ -f "$DISABLED_DIR/systemd" ]]; then
    mv "$DISABLED_DIR/systemd" "$PLUGINS_DIR/"
    log_info "Восстановлен плагин: systemd (альтернативный)"
    ((restored_count++))
fi

if [[ -f "$DISABLED_DIR/pam" ]]; then
    mv "$DISABLED_DIR/pam" "$PLUGINS_DIR/"
    log_info "Восстановлен плагин: pam (альтернативный)"
    ((restored_count++))
fi

# Проверяем, остались ли файлы в директории отключенных плагинов
if [[ -z "$(ls -A "$DISABLED_DIR" 2>/dev/null)" ]]; then
    rmdir "$DISABLED_DIR"
    log_info "Директория отключенных плагинов удалена (пустая)"
fi

if [[ $restored_count -gt 0 ]]; then
    log_info "Восстановлено плагинов: $restored_count"
    log_warn "ВНИМАНИЕ: После восстановления плагинов Lynis может снова зависать!"
    log_warn "Рекомендуется протестировать работу Lynis перед включением автоматического запуска"
else
    log_warn "Не найдено плагинов для восстановления"
fi

log_info "Восстановление завершено" 