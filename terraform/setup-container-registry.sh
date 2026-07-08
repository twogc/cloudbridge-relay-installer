#!/bin/bash

# CloudBridge Relay - Настройка Container Registry в VK Cloud
# Этот скрипт поможет настроить Container Registry для автоматизации

set -e

echo "Настройка Container Registry для CloudBridge Relay"
echo "=================================================="

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

# Проверка зависимостей
check_dependencies() {
    print_status "Проверка зависимостей..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker не установлен. Установите Docker и попробуйте снова."
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        print_warning "jq не установлен. Установка..."
        sudo apt update && sudo apt install -y jq
    fi
    
    print_success "Все зависимости установлены"
}

# Создание Container Registry через VK Cloud API
create_container_registry() {
    print_status "Создание Container Registry..."
    
    # Получаем токен аутентификации
    if [ -z "$OS_AUTH_TOKEN" ]; then
        print_status "Получение токена аутентификации..."
        source ../auth.sh
    fi
    
    # Создаем Container Registry
    print_status "Создание Container Registry 'cloudbridge-relay'..."
    
    # Используем VK Cloud API для создания Container Registry
    # Примечание: это может потребовать специальных прав доступа
    
    print_warning "Container Registry должен быть создан вручную через VK Cloud Console"
    print_status "Инструкции по созданию:"
    echo "1. Войдите в VK Cloud Console: https://mcs.mail.ru/"
    echo "2. Перейдите в раздел 'Container Registry'"
    echo "3. Нажмите 'Create Registry'"
    echo "4. Заполните форму:"
    echo "   - Name: cloudbridge-relay"
    echo "   - Description: CloudBridge Relay Docker Registry"
    echo "   - Region: MS1 (Москва)"
    echo "5. Нажмите 'Create'"
}

# Настройка аутентификации для Container Registry
setup_registry_auth() {
    print_status "Настройка аутентификации для Container Registry..."
    
    print_warning "После создания Container Registry выполните следующие шаги:"
    echo ""
    echo "1. В созданном registry перейдите в 'Settings'"
    echo "2. В разделе 'Access' создайте новый токен:"
    echo "   - Username: cloudbridge-user"
    echo "   - Password: [сгенерированный пароль]"
    echo "3. Скопируйте username и password"
    echo ""
    echo "4. Настройте GitHub Secrets:"
    echo "   - VK_REGISTRY_USERNAME = cloudbridge-user"
    echo "   - VK_REGISTRY_PASSWORD = [сгенерированный пароль]"
}

# Тест подключения к Container Registry
test_registry_connection() {
    print_status "Тест подключения к Container Registry..."
    
    if [ -z "$VK_REGISTRY_USERNAME" ] || [ -z "$VK_REGISTRY_PASSWORD" ]; then
        print_warning "Переменные VK_REGISTRY_USERNAME и VK_REGISTRY_PASSWORD не установлены"
        print_status "Установите их для тестирования:"
        echo "export VK_REGISTRY_USERNAME=your-username"
        echo "export VK_REGISTRY_PASSWORD=your-password"
        return 1
    fi
    
    print_status "Тест подключения к registry.vkcs.ru..."
    if echo "$VK_REGISTRY_PASSWORD" | docker login registry.vkcs.ru -u "$VK_REGISTRY_USERNAME" --password-stdin; then
        print_success "Подключение к Container Registry успешно!"
        return 0
    else
        print_error "Ошибка подключения к Container Registry"
        return 1
    fi
}

# Сборка и загрузка тестового образа
build_and_push_test_image() {
    print_status "Сборка и загрузка тестового образа..."
    
    if [ -z "$VK_REGISTRY_USERNAME" ] || [ -z "$VK_REGISTRY_PASSWORD" ]; then
        print_warning "Пропуск сборки образа - не настроена аутентификация"
        return 1
    fi
    
    # Переходим в корневую директорию проекта
    cd ..
    
    print_status "Сборка Docker образа..."
    docker build -f Dockerfile.relay -t registry.vkcs.ru/cloudbridge-relay:test .
    
    print_status "Загрузка образа в Container Registry..."
    docker push registry.vkcs.ru/cloudbridge-relay:test
    
    print_success "Тестовый образ успешно загружен!"
    
    # Возвращаемся в terraform директорию
    cd terraform
}

# Создание конфигурационных файлов
create_config_files() {
    print_status "Создание конфигурационных файлов..."
    
    # Создаем .env файл для локального тестирования
    if [ ! -f "../.env" ]; then
        cat > ../.env << 'EOF'
# CloudBridge Relay - Локальные переменные окружения
# Скопируйте этот файл и настройте под свои нужды

# VK Cloud Container Registry
VK_REGISTRY_USERNAME=your-username
VK_REGISTRY_PASSWORD=your-password

# VK Cloud OpenStack
VK_CLOUD_OPENRC=your-openrc-content

# Docker образ
DOCKER_IMAGE_NAME=cloudbridge-relay
DOCKER_IMAGE_TAG=latest
EOF
        print_success "Создан файл .env"
    else
        print_status "Файл .env уже существует"
    fi
    
    # Создаем скрипт для загрузки переменных окружения
    cat > ../load-env.sh << 'EOF'
#!/bin/bash
# Загрузка переменных окружения из .env файла

if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo "Переменные окружения загружены из .env файла"
else
    echo "Файл .env не найден"
    exit 1
fi
EOF
    
    chmod +x ../load-env.sh
    print_success "Создан скрипт load-env.sh"
}

# Основная функция
main() {
    echo ""
    print_status "Начинаем настройку Container Registry..."
    echo ""
    
    check_dependencies
    echo ""
    
    create_container_registry
    echo ""
    
    setup_registry_auth
    echo ""
    
    create_config_files
    echo ""
    
    print_status "Настройка завершена!"
    echo ""
    print_status "Следующие шаги:"
    echo "1. Создайте Container Registry в VK Cloud Console"
    echo "2. Настройте аутентификацию"
    echo "3. Обновите .env файл с вашими учетными данными"
    echo "4. Запустите: source ../load-env.sh"
    echo "5. Протестируйте подключение: ./test-registry-connection.sh"
    echo ""
    print_success "Container Registry готов к использованию!"
}

# Запуск основной функции
main "$@"
