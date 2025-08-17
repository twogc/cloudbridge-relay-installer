#!/bin/bash

# Скрипт для проверки и настройки GitHub Secrets для CI/CD
# Использование: ./scripts/check-ci-secrets.sh

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
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

# Проверяем наличие gh CLI
check_gh_cli() {
    if ! command -v gh &> /dev/null; then
        log_error "GitHub CLI (gh) не установлен"
        log_info "Установите GitHub CLI: https://cli.github.com/"
        exit 1
    fi
    
    # Проверяем аутентификацию
    if ! gh auth status &> /dev/null; then
        log_error "GitHub CLI не аутентифицирован"
        log_info "Выполните: gh auth login"
        exit 1
    fi
    
    log_success "GitHub CLI готов к использованию"
}

# Получаем информацию о репозитории
get_repo_info() {
    log_info "Получение информации о репозитории..."
    
    # Получаем owner/repo из git remote
    REMOTE_URL=$(git remote get-url origin)
    if [[ $REMOTE_URL == *"github.com"* ]]; then
        REPO_PATH=$(echo "$REMOTE_URL" | sed 's/.*github\.com[:/]\([^/]*\/[^/]*\)\.git.*/\1/')
        OWNER=$(echo "$REPO_PATH" | cut -d'/' -f1)
        REPO=$(echo "$REPO_PATH" | cut -d'/' -f2)
    else
        log_error "Не удалось определить GitHub репозиторий"
        exit 1
    fi
    
    log_success "Репозиторий: $OWNER/$REPO"
}

# Проверяем существующие секреты
check_existing_secrets() {
    log_info "Проверка существующих секретов..."
    
    local secrets=("REGISTRY_IP" "REGISTRY_USERNAME" "REGISTRY_PASSWORD" "KUBECONFIG")
    local missing_secrets=()
    
    for secret in "${secrets[@]}"; do
        if gh secret list --repo "$OWNER/$REPO" | grep -q "^$secret"; then
            log_success "Секрет $secret существует"
        else
            log_warning "Секрет $secret отсутствует"
            missing_secrets+=("$secret")
        fi
    done
    
    if [ ${#missing_secrets[@]} -eq 0 ]; then
        log_success "Все необходимые секреты настроены"
        return 0
    else
        log_warning "Отсутствующие секреты: ${missing_secrets[*]}"
        return 1
    fi
}

# Настройка секретов
setup_secrets() {
    log_info "Настройка секретов..."
    
    # Запрашиваем значения секретов
    echo
    log_info "Введите значения для секретов (или нажмите Enter для пропуска):"
    echo
    
    # REGISTRY_IP
    read -p "REGISTRY_IP (IP адрес VK Cloud Registry): " REGISTRY_IP
    if [[ -n "$REGISTRY_IP" ]]; then
        if gh secret set REGISTRY_IP --body "$REGISTRY_IP" --repo "$OWNER/$REPO"; then
            log_success "Секрет REGISTRY_IP создан"
        else
            log_error "Ошибка создания секрета REGISTRY_IP"
        fi
    fi
    
    # REGISTRY_USERNAME
    read -p "REGISTRY_USERNAME (имя пользователя): " REGISTRY_USERNAME
    if [[ -n "$REGISTRY_USERNAME" ]]; then
        if gh secret set REGISTRY_USERNAME --body "$REGISTRY_USERNAME" --repo "$OWNER/$REPO"; then
            log_success "Секрет REGISTRY_USERNAME создан"
        else
            log_error "Ошибка создания секрета REGISTRY_USERNAME"
        fi
    fi
    
    # REGISTRY_PASSWORD
    read -s -p "REGISTRY_PASSWORD (пароль): " REGISTRY_PASSWORD
    echo
    if [[ -n "$REGISTRY_PASSWORD" ]]; then
        if gh secret set REGISTRY_PASSWORD --body "$REGISTRY_PASSWORD" --repo "$OWNER/$REPO"; then
            log_success "Секрет REGISTRY_PASSWORD создан"
        else
            log_error "Ошибка создания секрета REGISTRY_PASSWORD"
        fi
    fi
    
    # KUBECONFIG
    read -p "Путь к файлу KUBECONFIG (или нажмите Enter для пропуска): " KUBECONFIG_PATH
    if [[ -n "$KUBECONFIG_PATH" ]] && [[ -f "$KUBECONFIG_PATH" ]]; then
        KUBECONFIG_B64=$(base64 -w 0 < "$KUBECONFIG_PATH")
        if gh secret set KUBECONFIG --body "$KUBECONFIG_B64" --repo "$OWNER/$REPO"; then
            log_success "Секрет KUBECONFIG создан"
        else
            log_error "Ошибка создания секрета KUBECONFIG"
        fi
    elif [[ -n "$KUBECONFIG_PATH" ]]; then
        log_warning "Файл KUBECONFIG не найден: $KUBECONFIG_PATH"
    fi
}

# Основная функция
main() {
    log_info "Проверка и настройка GitHub Secrets для CI/CD"
    echo
    
    # Проверяем gh CLI
    check_gh_cli
    
    # Получаем информацию о репозитории
    get_repo_info
    
    # Проверяем существующие секреты
    if check_existing_secrets; then
        log_success "Все секреты уже настроены"
        echo
        log_info "Для обновления секретов используйте:"
        log_info "gh secret set SECRET_NAME --body 'value' --repo $OWNER/$REPO"
        exit 0
    fi
    
    # Настраиваем недостающие секреты
    setup_secrets
    
    echo
    log_success "Настройка секретов завершена"
    log_info "Теперь CI/CD должен работать корректно"
}

# Запуск скрипта
main "$@"
