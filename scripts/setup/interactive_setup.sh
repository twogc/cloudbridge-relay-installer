#!/bin/bash

# CloudBridge Relay - Interactive Setup Script
# Интерактивная настройка с получением публичного ключа от пользователя

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Функции логирования
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Красивая заставка
show_banner() {
    echo -e "\033[1;36m"
    echo "╔═══════════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                           ║"
    echo "║    ██████╗ ██████╗  ██████╗ ██████╗ ██████╗ ██████╗ ██████╗ ██████╗       ║"
    echo "║   ██╔════╝██╔════╝ ██╔════╝██╔════╝██╔════╝██╔════╝██╔════╝██╔════╝       ║"
    echo "║   ██║     ██║  ███╗██║     ██║     ██║     ██║     ██║     ██║            ║"
    echo "║   ██║     ██║   ██║██║     ██║     ██║     ██║     ██║     ██║            ║"
    echo "║   ╚██████╗╚██████╔╝╚██████╗╚██████╗╚██████╗╚██████╗╚██████╗╚██████╗       ║"
    echo "║    ╚═════╝ ╚═════╝  ╚═════╝ ╚═════╝ ╚═════╝ ╚═════╝ ╚═════╝ ╚═════╝       ║"
    echo "║                                                                           ║"
    echo "║                    ██████╗██╗      ██████╗ ██████╗ ██████╗                ║"
    echo "║                   ██╔════╝██║     ██╔═══██╗██╔══██╗██╔══██╗               ║"
    echo "║                   ██║     ██║     ██║   ██║██████╔╝██║  ██║               ║"
    echo "║                   ██║     ██║     ██║   ██║██╔══██╗██║  ██║               ║"
    echo "║                   ╚██████╗███████╗╚██████╔╝██║  ██║██████╔╝               ║"
    echo "║                    ╚═════╝╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚═════╝                ║"
    echo "║                                                                           ║"
    echo "║                       ██████╗ ███████╗██╗      █████╗ ██╗   ██╗           ║"
    echo "║                      ██╔══██╗██╔════╝██║     ██╔══██╗╚██╗ ██╔╝            ║"
    echo "║                      ██████╔╝█████╗  ██║     ███████║ ╚████╔╝             ║"
    echo "║                      ██╔══██╗██╔══╝  ██║     ██╔══██║  ╚██╔╝              ║"
    echo "║                      ██║  ██║███████╗███████╗██║  ██║   ██║               ║"
    echo "║                      ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝  ╚═╝   ╚═╝               ║"
    echo "║                                                                           ║"
    echo "║                           Interactive Setup v2.0                          ║"
    echo "║                                                                           ║"
    echo "║  Step-by-Step Keycloak Setup  |  Automated Relay Configuration            ║"
    echo "║  Secure Key Management        |  Real-time Monitoring Setup               ║"
    echo "║  Interactive Workflow         |  Professional Deployment                  ║"
    echo "║                                                                           ║"
    echo "╚═══════════════════════════════════════════════════════════════════════════╝"
    echo -e "\033[0m"
    echo -e "\033[1;33m Starting Interactive CloudBridge Relay Setup...\033[0m"
    echo -e "\033[1;33m $(date '+%Y-%m-%d %H:%M:%S')\033[0m"
    echo ""
}

# Проверка зависимостей
check_dependencies() {
    log "Проверка зависимостей..."
    
    if ! command -v docker &> /dev/null; then
        error "Docker не установлен"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose не установлен"
        exit 1
    fi
    
    log "Все зависимости проверены"
}

# Создание необходимых директорий
create_directories() {
    log "Создание необходимых директорий..."
    
    mkdir -p keycloak/data
    mkdir -p keycloak/imports
    mkdir -p logs
    mkdir -p certs
    mkdir -p backups
    mkdir -p config
    
    # Устанавливаем правильные права
    chmod 755 keycloak/data
    chmod 755 keycloak/imports
    chmod 755 logs
    chmod 700 certs
    chmod 755 backups
    chmod 755 config
    
    log "Директории созданы"
}

# Генерация .env файла
generate_env() {
    log "Генерация .env файла..."
    
    if [[ ! -f ".env" ]]; then
        ./scripts/setup/generate_secrets.sh --env .env
        log ".env файл создан"
    else
        log ".env файл уже существует"
    fi
}

# Запуск Keycloak
start_keycloak() {
    log "Запуск Keycloak..."
    
    # Запускаем только Keycloak и PostgreSQL
    docker-compose up -d postgres keycloak
    
    log "Ожидание готовности Keycloak..."
    
    local max_attempts=60
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -f http://localhost:8080/ &> /dev/null; then
            log "Keycloak запущен и готов"
            return 0
        fi
        
        log "Попытка $attempt/$max_attempts - Keycloak еще не готов..."
        sleep 5
        ((attempt++))
    done
    
    error "Keycloak не запустился в течение 5 минут"
    return 1
}

# Инструкции по получению ключа
show_keycloak_instructions() {
    echo -e "\033[1;36m"
    echo "╔══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                              ║"
    echo "║                    Keycloak Public Key Setup                          ║"
    echo "║                                                                              ║"
    echo "╚══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "\033[0m"
    
    echo -e "\033[1;33m Пошаговая инструкция:\033[0m"
    echo ""
    echo -e "1. \033[1;36mОткройте Keycloak Admin Console:\033[0m"
    echo -e "   \033[1;33mhttp://localhost:8080\033[0m"
    echo ""
    echo -e "2. \033[1;36mВойдите в админку:\033[0m"
    echo -e "   Логин: \033[1;33madmin\033[0m"
    echo -e "   Пароль: \033[1;33madmin\033[0m (или из .env файла)"
    echo ""
    echo -e "3. \033[1;36mСоздайте или выберите realm:\033[0m"
    echo -e "   • Создайте новый realm 'cloudbridge'"
    echo -e "   • Или используйте существующий"
    echo ""
    echo -e "4. \033[1;36mПолучите публичный ключ:\033[0m"
    echo -e "   • Перейдите в \033[1;33mRealm Settings\033[0m"
    echo -e "   • Выберите вкладку \033[1;33mKeys\033[0m"
    echo -e "   • Найдите ключ \033[1;33mRS256\033[0m"
    echo -e "   • Скопируйте \033[1;33mPublic Key\033[0m"
    echo ""
    echo -e "5. \033[1;36mВставьте ключ в любом формате:\033[0m"
    echo -e "   \033[1;33m• PEM формат (с заголовками):\033[0m"
    echo -e "     -----BEGIN PUBLIC KEY-----"
    echo -e "     MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA..."
    echo -e "     -----END PUBLIC KEY-----"
    echo -e "   \033[1;33m• Или просто base64 (без заголовков):\033[0m"
    echo -e "     MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA..."
    echo ""
    echo -e "\033[1;32m Совет: Можно вставить ключ в любом формате - скрипт автоматически преобразует\033[0m"
    echo ""
    echo -e "\033[1;33m═══════════════════════════════════════════════════════════════════════════════\033[0m"
    echo -e "\033[1;33m Нажмите Enter для запуска установки...\033[0m"
    read -p ""
    echo ""
}

# Получение публичного ключа от пользователя
get_public_key() {
    echo -e "\033[1;36m"
    echo "╔══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                              ║"
    echo "║                     Введите публичный ключ Keycloak                          ║"
    echo "║                                                                              ║"
    echo "╚══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "\033[0m"
    
    echo -e "\033[1;33mВставьте публичный ключ (завершите ввод пустой строкой):\033[0m"
    echo ""
    
    local public_key=""
    local line=""
    
    while IFS= read -r line; do
        if [[ -z "$line" ]]; then
            break
        fi
        if [[ -z "$public_key" ]]; then
            public_key="$line"
        else
            public_key="$public_key"$'\n'"$line"
        fi
    done
    
    echo ""
    
    # Проверка и преобразование формата ключа
    if [[ ! "$public_key" =~ ^-----BEGIN\ PUBLIC\ KEY----- ]]; then
        # Если ключ в base64 формате, преобразуем в PEM
        if [[ "$public_key" =~ ^[A-Za-z0-9+/=]+$ ]]; then
            log "Обнаружен base64 ключ, преобразуем в PEM формат..."
            local pem_key="-----BEGIN PUBLIC KEY-----"
            pem_key+=$'\n'
            pem_key+="$public_key"
            pem_key+=$'\n'
            pem_key+="-----END PUBLIC KEY-----"
            public_key="$pem_key"
            log "Ключ преобразован в PEM формат"
        else
            error "Неверный формат ключа! Ключ должен быть в PEM формате или base64"
            return 1
        fi
    fi
    
    if [[ ! "$public_key" =~ -----END\ PUBLIC\ KEY-----$ ]]; then
        error "Неверный формат ключа! Ключ должен заканчиваться на '-----END PUBLIC KEY-----'"
        return 1
    fi
    
    # Сохранение ключа в файл
    echo "$public_key" > config/public_key.pem
    log "Публичный ключ сохранён в config/public_key.pem"
    
    return 0
}

# Запуск всех сервисов
start_all_services() {
    log "Запуск всех сервисов CloudBridge Relay..."
    
    # Запускаем production скрипт
    ./scripts/utils/start_production.sh
}

# Основная функция
main() {
    show_banner
    
    # Проверка зависимостей
    check_dependencies
    
    # Создание директорий
    create_directories
    
    # Генерация .env
    generate_env
    
    # Запуск Keycloak
    start_keycloak
    
    # Показ инструкций
    show_keycloak_instructions
    
    # Получение ключа от пользователя
    while ! get_public_key; do
        warn "Попробуйте ещё раз или нажмите Ctrl+C для выхода"
        echo ""
    done
    
    # Запуск всех сервисов
    start_all_services
    
    echo -e "\033[1;32m"
    echo "╔══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                              ║"
    echo "║                    Setup Completed Successfully!                      ║"
    echo "║                                                                              ║"
    echo "╚══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "\033[0m"
    
    log "CloudBridge Relay полностью настроен и запущен!"
    log "Публичный ключ Keycloak сохранён и используется в конфигурации."
}

# Запуск скрипта
main "$@" 