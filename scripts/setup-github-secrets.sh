#!/usr/bin/env bash

# 🔐 Настройка GitHub Secrets для автоматизации Docker Registry
# Автор: VK Cloud API Manager
# Версия: 1.0.0

set -euo pipefail

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Функции логирования
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Конфигурация
REGISTRY_IP="${REGISTRY_IP:-95.163.250.128}"
REGISTRY_USERNAME="${REGISTRY_USERNAME:-cloudbridge}"
REGISTRY_PASSWORD="${REGISTRY_PASSWORD:-password}"
GITHUB_TOKEN="${GITHUB_TOKEN:-}"
GITHUB_REPO="${GITHUB_REPO:-}"

# Функция для проверки зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    local deps=("gh" "jq" "curl")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "Отсутствует зависимость: $dep"
            exit 1
        fi
    done
    log_success "Все зависимости установлены"
}

# Функция для проверки GitHub CLI
check_github_cli() {
    log_info "Проверка GitHub CLI..."
    
    if ! gh auth status &> /dev/null; then
        log_error "GitHub CLI не аутентифицирован"
        log_info "Выполните: gh auth login"
        exit 1
    fi
    
    log_success "GitHub CLI аутентифицирован"
}

# Функция для получения информации о репозитории
get_repo_info() {
    log_info "Получение информации о репозитории..."
    
    if [ -z "$GITHUB_REPO" ]; then
        # Получаем информацию о текущем репозитории
        GITHUB_REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
    fi
    
    log_success "Репозиторий: $GITHUB_REPO"
}

# Функция для создания секретов
create_secrets() {
    log_info "Создание GitHub Secrets..."
    
    # Список секретов для создания
    declare -A secrets=(
        ["REGISTRY_IP"]="$REGISTRY_IP"
        ["REGISTRY_USERNAME"]="$REGISTRY_USERNAME"
        ["REGISTRY_PASSWORD"]="$REGISTRY_PASSWORD"
    )
    
    for secret_name in "${!secrets[@]}"; do
        local secret_value="${secrets[$secret_name]}"
        
        log_info "Создание секрета: $secret_name"
        
        if gh secret set "$secret_name" --body "$secret_value" --repo "$GITHUB_REPO"; then
            log_success "Секрет $secret_name создан"
        else
            log_warning "Секрет $secret_name уже существует или ошибка создания"
        fi
    done
}

# Функция для создания KUBECONFIG секрета
create_kubeconfig_secret() {
    log_info "Создание KUBECONFIG секрета..."
    
    # Проверяем наличие kubeconfig файла
    local kubeconfig_file="$HOME/.kube/config"
    
    if [ ! -f "$kubeconfig_file" ]; then
        log_warning "Kubeconfig файл не найден: $kubeconfig_file"
        log_info "Создайте kubeconfig файл для VK Cloud Kubernetes"
        return 0
    fi
    
    # Кодируем kubeconfig в base64
    local kubeconfig_b64=$(base64 -w 0 < "$kubeconfig_file")
    
    if gh secret set "KUBECONFIG" --body "$kubeconfig_b64" --repo "$GITHUB_REPO"; then
        log_success "KUBECONFIG секрет создан"
    else
        log_warning "KUBECONFIG секрет уже существует или ошибка создания"
    fi
}

# Функция для проверки секретов
verify_secrets() {
    log_info "Проверка созданных секретов..."
    
    # Получаем список секретов
    local secrets_list=$(gh secret list --repo "$GITHUB_REPO" --json name)
    
    # Проверяем наличие необходимых секретов
    local required_secrets=("REGISTRY_IP" "REGISTRY_USERNAME" "REGISTRY_PASSWORD")
    
    for secret in "${required_secrets[@]}"; do
        if echo "$secrets_list" | jq -e ".[] | select(.name == \"$secret\")" > /dev/null; then
            log_success "Секрет $secret найден"
        else
            log_warning "Секрет $secret не найден"
        fi
    done
    
    # Проверяем KUBECONFIG
    if echo "$secrets_list" | jq -e '.[] | select(.name == "KUBECONFIG")' > /dev/null; then
        log_success "Секрет KUBECONFIG найден"
    else
        log_warning "Секрет KUBECONFIG не найден"
    fi
}

# Функция для тестирования registry
test_registry() {
    log_info "Тестирование подключения к registry..."
    
    if curl -s --connect-timeout 10 "http://$REGISTRY_IP:5000/v2/" > /dev/null; then
        log_success "Registry доступен: $REGISTRY_IP:5000"
    else
        log_warning "Registry недоступен: $REGISTRY_IP:5000"
        log_info "Убедитесь, что registry запущен и доступен"
    fi
}

# Функция для вывода информации
print_info() {
    echo ""
    log_success "🎉 GitHub Secrets настроены!"
    echo ""
    echo "📋 Настроенные секреты:"
    echo "   REGISTRY_IP       = $REGISTRY_IP"
    echo "   REGISTRY_USERNAME = $REGISTRY_USERNAME"
    echo "   REGISTRY_PASSWORD = $REGISTRY_PASSWORD"
    echo "   KUBECONFIG        = [base64 encoded]"
    echo ""
    echo "🔧 Следующие шаги:"
    echo "   1. Убедитесь, что Docker Registry запущен"
    echo "   2. Проверьте доступность registry: http://$REGISTRY_IP:5000"
    echo "   3. Запустите GitHub Actions workflow"
    echo ""
    echo "🚀 Команды для тестирования:"
    echo "   # Тест подключения к registry"
    echo "   curl -u $REGISTRY_USERNAME:$REGISTRY_PASSWORD http://$REGISTRY_IP:5000/v2/_catalog"
    echo ""
    echo "   # Ручная сборка и отправка образа"
    echo "   ./scripts/docker-registry-automation.sh"
    echo ""
    echo "📚 Документация:"
    echo "   - GitHub Actions: .github/workflows/docker-build-push.yml"
    echo "   - Автоматизация: scripts/docker-registry-automation.sh"
    echo "   - Установка registry: scripts/install-vkcloud-registry.sh"
    echo ""
}

# Функция для удаления секретов
delete_secrets() {
    log_info "Удаление GitHub Secrets..."
    
    local secrets_to_delete=("REGISTRY_IP" "REGISTRY_USERNAME" "REGISTRY_PASSWORD" "KUBECONFIG")
    
    for secret in "${secrets_to_delete[@]}"; do
        log_info "Удаление секрета: $secret"
        
        if gh secret delete "$secret" --repo "$GITHUB_REPO" --yes; then
            log_success "Секрет $secret удален"
        else
            log_warning "Секрет $secret не найден или ошибка удаления"
        fi
    done
}

# Главная функция
main() {
    log_info "🔐 Настройка GitHub Secrets для VK Cloud Registry"
    
    check_dependencies
    check_github_cli
    get_repo_info
    create_secrets
    create_kubeconfig_secret
    verify_secrets
    test_registry
    print_info
}

# Обработка аргументов
case "${1:-}" in
    "help"|"-h"|"--help")
        echo "Использование: $0 [команда]"
        echo ""
        echo "Команды:"
        echo "  setup   - Настроить секреты (по умолчанию)"
        echo "  verify  - Проверить секреты"
        echo "  delete  - Удалить секреты"
        echo "  test    - Тестировать registry"
        echo "  help    - Показать эту справку"
        echo ""
        echo "Переменные окружения:"
        echo "  REGISTRY_IP       - IP адрес registry (по умолчанию: 95.163.250.128)"
        echo "  REGISTRY_USERNAME - Имя пользователя (по умолчанию: cloudbridge)"
        echo "  REGISTRY_PASSWORD - Пароль (по умолчанию: password)"
        echo "  GITHUB_REPO       - GitHub репозиторий (автоопределение)"
        echo ""
        echo "Примеры:"
        echo "  $0                    # Настроить секреты"
        echo "  $0 verify             # Проверить секреты"
        echo "  REGISTRY_IP=192.168.1.100 $0  # С кастомным IP"
        ;;
    "verify")
        check_dependencies
        check_github_cli
        get_repo_info
        verify_secrets
        test_registry
        ;;
    "delete")
        check_dependencies
        check_github_cli
        get_repo_info
        delete_secrets
        ;;
    "test")
        test_registry
        ;;
    "")
        main
        ;;
    *)
        log_error "Неизвестная команда: $1"
        echo "Используйте '$0 help' для справки"
        exit 1
        ;;
esac
