#!/bin/bash

# 🐳 Скрипт автоматической настройки зеркал Docker Hub для России
# Автор: CloudBridge Team
# Версия: 1.0.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для вывода
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка прав root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "Этот скрипт должен быть запущен с правами root или через sudo"
        exit 1
    fi
}

# Проверка Docker
check_docker() {
    log_info "Проверка установки Docker..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker не установлен. Пожалуйста, установите Docker перед продолжением."
        log_info "Выполните: curl -fsSL https://get.docker.com | sh"
        exit 1
    fi
    
    if ! systemctl is-active --quiet docker; then
        log_warning "Docker не запущен. Запуск Docker..."
        systemctl start docker
    fi
    
    log_success "Docker установлен и запущен"
}

# Создание резервной копии
backup_config() {
    log_info "Создание резервной копии конфигурации..."
    
    if [ -f /etc/docker/daemon.json ]; then
        cp /etc/docker/daemon.json /etc/docker/daemon.json.backup.$(date +%Y%m%d_%H%M%S)
        log_success "Резервная копия создана: /etc/docker/daemon.json.backup.$(date +%Y%m%d_%H%M%S)"
    else
        log_info "Конфигурация Docker не найдена, создание новой"
    fi
}

# Настройка зеркал
setup_mirrors() {
    log_info "Настройка зеркал Docker Hub..."
    
    # Список зеркал
    cat > /etc/docker/daemon.json << 'EOF'
{
  "registry-mirrors": [
    "https://dockerhub.timeweb.cloud",
    "https://huecker.io",
    "https://mirror.gcr.io",
    "https://c.163.com",
    "https://registry.docker-cn.com",
    "https://daocloud.io",
    "https://cr.yandex/mirror",
    "https://noohub.ru"
  ]
}
EOF
    
    log_success "Конфигурация зеркал создана"
}

# Перезапуск Docker
restart_docker() {
    log_info "Перезапуск Docker..."
    
    systemctl restart docker
    
    # Ожидание запуска
    sleep 5
    
    if systemctl is-active --quiet docker; then
        log_success "Docker успешно перезапущен"
    else
        log_error "Ошибка при перезапуске Docker"
        exit 1
    fi
}

# Проверка настройки
verify_setup() {
    log_info "Проверка настройки зеркал..."
    
    if docker info | grep -q "Registry Mirrors"; then
        log_success "Зеркала Docker Hub настроены:"
        docker info | grep -A 10 "Registry Mirrors"
    else
        log_warning "Не удалось проверить настройку зеркал"
    fi
}

# Тестирование зеркал
test_mirrors() {
    log_info "Тестирование зеркал..."
    
    # Тестирование загрузки образа
    log_info "Тестирование загрузки образа alpine:latest..."
    if docker pull alpine:latest > /dev/null 2>&1; then
        log_success "Тест загрузки прошел успешно"
    else
        log_warning "Тест загрузки не прошел, но это может быть нормально"
    fi
    
    # Очистка тестового образа
    docker rmi alpine:latest > /dev/null 2>&1 || true
}

# Проверка статуса зеркал
check_mirror_status() {
    log_info "Проверка статуса зеркал..."
    
    local mirrors=(
        "dockerhub.timeweb.cloud"
        "huecker.io"
        "mirror.gcr.io"
        "c.163.com"
        "registry.docker-cn.com"
        "daocloud.io"
        "cr.yandex"
        "noohub.ru"
    )
    
    echo ""
    log_info "Статус зеркал:"
    echo "================"
    
    for mirror in "${mirrors[@]}"; do
        echo -n "Проверка $mirror... "
        if curl -s --connect-timeout 5 https://$mirror/v2/ > /dev/null 2>&1; then
            echo "✅ Доступно"
        else
            echo "❌ Недоступно"
        fi
    done
}

# Откат изменений
rollback() {
    log_warning "Откат изменений..."
    
    if [ -f /etc/docker/daemon.json.backup.* ]; then
        local backup_file=$(ls -t /etc/docker/daemon.json.backup.* | head -1)
        cp "$backup_file" /etc/docker/daemon.json
        systemctl restart docker
        log_success "Откат выполнен из файла: $backup_file"
    else
        log_error "Резервная копия не найдена"
    fi
}

# Показать помощь
show_help() {
    echo "🐳 Скрипт настройки зеркал Docker Hub для России"
    echo ""
    echo "Использование:"
    echo "  $0 [опции]"
    echo ""
    echo "Опции:"
    echo "  -h, --help     Показать эту справку"
    echo "  -t, --test     Только тестирование зеркал"
    echo "  -s, --status   Проверить статус зеркал"
    echo "  -r, --rollback Откатить изменения"
    echo ""
    echo "Примеры:"
    echo "  sudo $0                    # Настройка зеркал"
    echo "  sudo $0 --test             # Тестирование"
    echo "  sudo $0 --status           # Проверка статуса"
    echo "  sudo $0 --rollback         # Откат"
}

# Основная функция
main() {
    echo "🐳 Настройка зеркал Docker Hub для России"
    echo "=========================================="
    echo ""
    
    # Обработка аргументов
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -t|--test)
            check_docker
            test_mirrors
            exit 0
            ;;
        -s|--status)
            check_docker
            check_mirror_status
            exit 0
            ;;
        -r|--rollback)
            check_root
            rollback
            exit 0
            ;;
        "")
            # Продолжить с настройкой
            ;;
        *)
            log_error "Неизвестная опция: $1"
            show_help
            exit 1
            ;;
    esac
    
    check_root
    check_docker
    backup_config
    setup_mirrors
    restart_docker
    verify_setup
    test_mirrors
    check_mirror_status
    
    echo ""
    log_success "🎉 Настройка зеркал Docker Hub завершена!"
    echo ""
    log_info "Полезные команды:"
    echo "  docker info | grep -A 10 'Registry Mirrors'  # Проверка зеркал"
    echo "  docker pull alpine:latest                     # Тест загрузки"
    echo "  $0 --status                                    # Проверка статуса"
    echo "  $0 --rollback                                  # Откат изменений"
    echo ""
}

# Запуск скрипта
main "$@"
