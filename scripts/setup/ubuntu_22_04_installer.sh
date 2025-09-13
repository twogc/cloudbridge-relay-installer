#!/bin/bash

# CloudBridge Relay - Ubuntu 22.04 Automated Installer
# Модульная система установки с загрузкой из Git репозитория
# Автор: CloudBridge Team
# Версия: 2.0

set -e

# Конфигурация
MODULES_REPO="https://github.com/twogc/cloudbridge-install-modules.git"
MODULES_BRANCH="main"
MODULES_DIR="/tmp/cloudbridge-modules"
INSTALL_LOG="/var/log/cloudbridge-install.log"
CONFIG_FILE="/etc/cloudbridge/installer.conf"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Функции логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$INSTALL_LOG"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$INSTALL_LOG"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$INSTALL_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$INSTALL_LOG"
}

# Красивый баннер
show_banner() {
    clear
    echo -e "\033[1;36m"
    echo "╔═══════════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                           ║"
    echo "║                    CloudBridge Relay Installer v2.0                      ║"
    echo "║                                                                           ║"
    echo "║                    Ubuntu 22.04 Automated Setup                          ║"
    echo "║                                                                           ║"
    echo "║  🔒 Security First  |  🚀 High Performance  |  📊 Full Monitoring        ║"
    echo "║                                                                           ║"
    echo "╚═══════════════════════════════════════════════════════════════════════════╝"
    echo -e "\033[0m"
    echo -e "\033[1;33m$(date '+%Y-%m-%d %H:%M:%S')\033[0m"
    echo ""
}

# Проверка root прав
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "Этот скрипт должен быть запущен с правами root (sudo)"
        exit 1
    fi
}

# Проверка Ubuntu 22.04
check_ubuntu_version() {
    log_info "Проверка версии Ubuntu..."
    
    if [[ ! -f /etc/os-release ]]; then
        log_error "Не удалось определить версию ОС"
        exit 1
    fi
    
    source /etc/os-release
    
    if [[ "$NAME" != "Ubuntu" ]] || [[ "$VERSION_ID" != "22.04" ]]; then
        log_error "Этот скрипт предназначен только для Ubuntu 22.04"
        log_error "Обнаружено: $NAME $VERSION_ID"
        exit 1
    fi
    
    log_success "Ubuntu 22.04 обнаружена"
}

# Проверка интернет соединения
check_internet() {
    log_info "Проверка интернет соединения..."
    
    if ! ping -c 1 8.8.8.8 &> /dev/null; then
        log_error "Нет интернет соединения"
        exit 1
    fi
    
    log_success "Интернет соединение доступно"
}

# Установка базовых зависимостей
install_basic_deps() {
    log_info "Установка базовых зависимостей..."
    
    apt update
    apt install -y \
        curl \
        wget \
        git \
        jq \
        unzip \
        software-properties-common \
        apt-transport-https \
        ca-certificates \
        gnupg \
        lsb-release
    
    log_success "Базовые зависимости установлены"
}

# Загрузка модулей из Git
download_modules() {
    log_info "Загрузка модулей установки..."
    
    if [[ -d "$MODULES_DIR" ]]; then
        rm -rf "$MODULES_DIR"
    fi
    
    if ! git clone -b "$MODULES_BRANCH" "$MODULES_REPO" "$MODULES_DIR"; then
        log_error "Не удалось загрузить модули из репозитория"
        log_info "Попытка загрузки из альтернативного источника..."
        
        # Альтернативный способ загрузки (если основной не работает)
        if ! curl -L -o /tmp/modules.zip "https://github.com/cloudbridge-team/cloudbridge-install-modules/archive/refs/heads/main.zip"; then
            log_error "Не удалось загрузить модули"
            exit 1
        fi
        
        unzip -q /tmp/modules.zip -d /tmp
        mv /tmp/cloudbridge-install-modules-main "$MODULES_DIR"
    fi
    
    log_success "Модули загружены в $MODULES_DIR"
}

# Парсинг конфигурации модулей
parse_modules_config() {
    log_info "Парсинг конфигурации модулей..."
    
    if [[ ! -f "$MODULES_DIR/modules.json" ]]; then
        log_error "Файл modules.json не найден"
        exit 1
    fi
    
    # Читаем порядок модулей из JSON
    MODULES_ORDER=$(jq -r '.modules[] | .file' "$MODULES_DIR/modules.json")
    
    log_success "Конфигурация модулей загружена"
}

# Выполнение модуля
execute_module() {
    local module_file="$1"
    local module_path="$MODULES_DIR/$module_file"
    
    if [[ ! -f "$module_path" ]]; then
        log_error "Модуль $module_file не найден"
        return 1
    fi
    
    log_info "Выполнение модуля: $module_file"
    
    # Делаем модуль исполняемым
    chmod +x "$module_path"
    
    # Выполняем модуль
    if bash "$module_path"; then
        log_success "Модуль $module_file выполнен успешно"
        return 0
    else
        log_error "Ошибка выполнения модуля $module_file"
        return 1
    fi
}

# Основная функция установки
main_installation() {
    log_info "Начало установки CloudBridge Relay..."
    
    # Создаем директорию для логов
    mkdir -p "$(dirname "$INSTALL_LOG")"
    
    # Создаем директорию конфигурации
    mkdir -p "$(dirname "$CONFIG_FILE")"
    
    # Выполняем модули по порядку
    for module in $MODULES_ORDER; do
        if ! execute_module "$module"; then
            log_error "Установка прервана из-за ошибки в модуле $module"
            exit 1
        fi
    done
    
    log_success "Все модули выполнены успешно"
}

# Финальная настройка
final_setup() {
    log_info "Выполнение финальной настройки..."
    
    # Создаем файл конфигурации
    cat > "$CONFIG_FILE" << EOF
# CloudBridge Relay Installer Configuration
# Создано: $(date)

INSTALL_DATE=$(date +%Y-%m-%d)
INSTALL_VERSION=2.0
MODULES_REPO=$MODULES_REPO
MODULES_BRANCH=$MODULES_BRANCH

# Установленные компоненты:
# - Системные обновления
# - Базовая безопасность
# - Docker и Docker Compose
# - UFW файрвол
# - Keycloak
# - Lynis аудит безопасности
# - Мониторинг
EOF
    
    # Устанавливаем права
    chmod 600 "$CONFIG_FILE"
    
    log_success "Финальная настройка завершена"
}

# Показ информации о завершении
show_completion_info() {
    echo ""
    echo -e "\033[1;32m╔═══════════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                           ║"
    echo "║                    🎉 УСТАНОВКА ЗАВЕРШЕНА УСПЕШНО! 🎉                    ║"
    echo "║                                                                           ║"
    echo "╚═══════════════════════════════════════════════════════════════════════════╝\033[0m"
    echo ""
    echo -e "\033[1;33m📋 Что было установлено:\033[0m"
    echo "   • Системные обновления и базовые пакеты"
    echo "   • Docker и Docker Compose"
    echo "   • UFW файрвол с настройками безопасности"
    echo "   • CloudBridge Relay сервер"
    echo "   • Lynis для аудита безопасности"
    echo "   • Система мониторинга"
    echo "   • mTLS поддержка"
    echo "   • Мониторинг аномалий"
    echo ""
    echo -e "\033[1;33m📁 Важные файлы:\033[0m"
    echo "   • Лог установки: $INSTALL_LOG"
    echo "   • Конфигурация: $CONFIG_FILE"
    echo ""
    echo -e "\033[1;33m🔧 Следующие шаги:\033[0m"
    echo "   1. Запустите Relay сервер: relay-start"
    echo "   2. Проверьте статус: relay-status"
    echo "   3. Запустите аудит безопасности: lynis audit system"
    echo "   4. Проверьте мониторинг: relay-monitor"
    echo ""
    echo -e "\033[1;33m📞 Поддержка:\033[0m"
    echo "   • Документация: https://github.com/cloudbridge-team/docs"
    echo "   • Issues: https://github.com/cloudbridge-team/cloudbridge-relay/issues"
    echo ""
}

# Обработка сигналов
cleanup() {
    log_warning "Получен сигнал прерывания, очистка..."
    
    if [[ -d "$MODULES_DIR" ]]; then
        rm -rf "$MODULES_DIR"
    fi
    
    log_info "Очистка завершена"
    exit 1
}

# Установка обработчиков сигналов
trap cleanup SIGINT SIGTERM

# Главная функция
main() {
    show_banner
    check_root
    check_ubuntu_version
    check_internet
    install_basic_deps
    download_modules
    parse_modules_config
    main_installation
    final_setup
    show_completion_info
}

# Запуск
main "$@"
