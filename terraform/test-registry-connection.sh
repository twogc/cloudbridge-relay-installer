#!/bin/bash

# CloudBridge Relay - Тест подключения к Container Registry
# Этот скрипт тестирует подключение к VK Cloud Container Registry

set -e

echo "🧪 Тест подключения к Container Registry"
echo "======================================="

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

# Проверка переменных окружения
check_environment() {
    print_status "Проверка переменных окружения..."
    
    if [ -z "$VK_REGISTRY_USERNAME" ]; then
        print_error "VK_REGISTRY_USERNAME не установлена"
        print_status "Установите переменную: export VK_REGISTRY_USERNAME=your-username"
        return 1
    fi
    
    if [ -z "$VK_REGISTRY_PASSWORD" ]; then
        print_error "VK_REGISTRY_PASSWORD не установлена"
        print_status "Установите переменную: export VK_REGISTRY_PASSWORD=your-password"
        return 1
    fi
    
    print_success "Переменные окружения настроены"
    echo "Username: $VK_REGISTRY_USERNAME"
    echo "Password: [скрыто]"
}

# Тест подключения к Container Registry
test_registry_login() {
    print_status "Тест подключения к registry.vkcs.ru..."
    
    if echo "$VK_REGISTRY_PASSWORD" | docker login registry.vkcs.ru -u "$VK_REGISTRY_USERNAME" --password-stdin; then
        print_success "Подключение к Container Registry успешно!"
        return 0
    else
        print_error "Ошибка подключения к Container Registry"
        return 1
    fi
}

# Тест создания тестового образа
test_image_creation() {
    print_status "Создание тестового образа..."
    
    # Создаем простой тестовый образ
    cat > /tmp/test.Dockerfile << 'EOF'
FROM alpine:latest
RUN echo "Hello from CloudBridge Relay test image" > /message.txt
CMD ["cat", "/message.txt"]
EOF
    
    print_status "Сборка тестового образа..."
    docker build -f /tmp/test.Dockerfile -t registry.vkcs.ru/cloudbridge-relay:test-connection /tmp
    
    print_success "Тестовый образ создан"
}

# Тест загрузки образа в Container Registry
test_image_push() {
    print_status "Тест загрузки образа в Container Registry..."
    
    if docker push registry.vkcs.ru/cloudbridge-relay:test-connection; then
        print_success "Образ успешно загружен в Container Registry!"
        return 0
    else
        print_error "Ошибка загрузки образа в Container Registry"
        return 1
    fi
}

# Тест скачивания образа из Container Registry
test_image_pull() {
    print_status "Тест скачивания образа из Container Registry..."
    
    # Удаляем локальный образ
    docker rmi registry.vkcs.ru/cloudbridge-relay:test-connection 2>/dev/null || true
    
    # Скачиваем образ заново
    if docker pull registry.vkcs.ru/cloudbridge-relay:test-connection; then
        print_success "Образ успешно скачан из Container Registry!"
        return 0
    else
        print_error "Ошибка скачивания образа из Container Registry"
        return 1
    fi
}

# Тест запуска скачанного образа
test_image_run() {
    print_status "Тест запуска скачанного образа..."
    
    if docker run --rm registry.vkcs.ru/cloudbridge-relay:test-connection; then
        print_success "Образ успешно запущен!"
        return 0
    else
        print_error "Ошибка запуска образа"
        return 1
    fi
}

# Очистка тестовых образов
cleanup_test_images() {
    print_status "Очистка тестовых образов..."
    
    # Удаляем тестовый образ
    docker rmi registry.vkcs.ru/cloudbridge-relay:test-connection 2>/dev/null || true
    
    # Удаляем временные файлы
    rm -f /tmp/test.Dockerfile
    
    print_success "Тестовые образы очищены"
}

# Основная функция тестирования
main() {
    echo ""
    print_status "Начинаем тестирование Container Registry..."
    echo ""
    
    # Проверяем переменные окружения
    if ! check_environment; then
        print_error "Переменные окружения не настроены. Запустите: source ../load-env.sh"
        exit 1
    fi
    
    # Тестируем подключение
    if ! test_registry_login; then
        print_error "Не удалось подключиться к Container Registry"
        exit 1
    fi
    
    # Тестируем создание образа
    test_image_creation
    
    # Тестируем загрузку образа
    if ! test_image_push; then
        print_error "Не удалось загрузить образ в Container Registry"
        exit 1
    fi
    
    # Тестируем скачивание образа
    if ! test_image_pull; then
        print_error "Не удалось скачать образ из Container Registry"
        exit 1
    fi
    
    # Тестируем запуск образа
    if ! test_image_run; then
        print_error "Не удалось запустить образ"
        exit 1
    fi
    
    # Очищаем тестовые образы
    cleanup_test_images
    
    echo ""
    print_success "Все тесты Container Registry пройдены успешно!"
    echo ""
    print_status "Container Registry готов к использованию для автоматизации!"
    echo ""
    print_status "Следующие шаги:"
    echo "1. Настройте GitHub Secrets с вашими учетными данными"
    echo "2. Обновите docker-compose.yml для использования VK Cloud Registry"
    echo "3. Протестируйте автоматизацию через GitHub Actions"
}

# Запуск основной функции
main "$@"
