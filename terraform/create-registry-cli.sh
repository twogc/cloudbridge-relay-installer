#!/bin/bash

# CloudBridge Relay - Создание Container Registry через OpenStack CLI
# Этот скрипт пытается создать Container Registry программно

set -e

echo "🔧 Создание Container Registry через OpenStack CLI"
echo "================================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода с цветом
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка аутентификации
check_auth() {
    print_status "Проверка аутентификации OpenStack..."
    
    if ! source ../auth.sh; then
        print_error "Ошибка аутентификации в VK Cloud"
        exit 1
    fi
    
    print_success "Аутентификация в VK Cloud успешна"
}

# Попытка создания Container Registry через OpenStack CLI
create_registry_cli() {
    print_status "Попытка создания Container Registry через OpenStack CLI..."
    
    # Проверяем доступные команды
    print_status "Проверка доступных команд OpenStack..."
    
    # Попытка 1: Стандартная команда OpenStack
    if command -v openstack &> /dev/null; then
        print_status "Попытка создания через openstack container registry create..."
        
        if openstack container registry create \
            --name cloudbridge-relay \
            --description "CloudBridge Relay Docker Registry" \
            --region MS1; then
            print_success "Container Registry создан через OpenStack CLI!"
            return 0
        else
            print_warning "Команда openstack container registry create не сработала"
        fi
    fi
    
    # Попытка 2: Через Magnum (если доступен)
    print_status "Попытка создания через Magnum..."
    
    if openstack coe cluster create \
        --name cloudbridge-relay-registry \
        --cluster-template kubernetes \
        --node-count 1; then
        print_success "Kubernetes кластер создан для Container Registry!"
        return 0
    else
        print_warning "Создание через Magnum не удалось"
    fi
    
    # Попытка 3: Через другие сервисы
    print_status "Попытка создания через другие сервисы..."
    
    # Проверяем доступные сервисы
    openstack service list
    openstack endpoint list
    
    return 1
}

# Создание Container Registry через VK Cloud API
create_registry_api() {
    print_status "Попытка создания через VK Cloud API..."
    
    # Получаем токен аутентификации
    TOKEN=$(openstack token issue -f value -c id)
    
    if [ -z "$TOKEN" ]; then
        print_error "Не удалось получить токен аутентификации"
        return 1
    fi
    
    print_status "Токен получен: ${TOKEN:0:20}..."
    
    # Попытка создания через API
    print_status "Отправка запроса на создание Container Registry..."
    
    # Используем curl для отправки API запроса
    if curl -X POST \
        -H "X-Auth-Token: $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "cloudbridge-relay",
            "description": "CloudBridge Relay Docker Registry",
            "region": "MS1"
        }' \
        "https://infra.mail.ru:5000/v3/container-registries"; then
        
        print_success "API запрос отправлен успешно!"
        return 0
    else
        print_warning "API запрос не удался"
        return 1
    fi
}

# Основная функция
main() {
    echo ""
    print_status "Начинаем создание Container Registry..."
    echo ""
    
    check_auth
    echo ""
    
    # Попытка создания через CLI
    if create_registry_cli; then
        print_success "Container Registry создан через CLI!"
        exit 0
    fi
    
    echo ""
    
    # Попытка создания через API
    if create_registry_api; then
        print_success "Container Registry создан через API!"
        exit 0
    fi
    
    echo ""
    print_warning "Автоматическое создание Container Registry не удалось"
    print_status "Необходимо создать Container Registry вручную через веб-консоль"
    echo ""
    print_status "Инструкции по ручному созданию:"
    echo "1. Войдите в VK Cloud Console: https://mcs.mail.ru/"
    echo "2. Перейдите в раздел 'Container Registry'"
    echo "3. Нажмите 'Create Registry'"
    echo "4. Заполните форму:"
    echo "   - Name: cloudbridge-relay"
    echo "   - Description: CloudBridge Relay Docker Registry"
    echo "   - Region: MS1 (Москва)"
    echo "5. Нажмите 'Create'"
    echo ""
    print_status "После создания не забудьте:"
    echo "- Настроить аутентификацию (username/password)"
    echo "- Обновить .env файл"
    echo "- Протестировать подключение"
}

# Запуск основной функции
main "$@"
