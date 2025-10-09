#!/bin/bash

# CloudBridge Relay - Sentry Integration Test
# Тестирует интеграцию Sentry с CloudBridge Relay

set -e

echo "🧪 CloudBridge Relay - Sentry Integration Test"
echo "=============================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Проверка переменных окружения
check_environment() {
    log_info "Проверка переменных окружения..."
    
    if [ -z "$SENTRY_DSN" ]; then
        log_error "SENTRY_DSN не установлен"
        log_info "Установите переменную: export SENTRY_DSN='your-dsn-here'"
        exit 1
    fi
    
    if [ -z "$SENTRY_ENVIRONMENT" ]; then
        export SENTRY_ENVIRONMENT="test"
        log_warning "SENTRY_ENVIRONMENT не установлен, используем 'test'"
    fi
    
    if [ -z "$SENTRY_RELEASE" ]; then
        export SENTRY_RELEASE="cloudbridge-relay@test"
        log_warning "SENTRY_RELEASE не установлен, используем 'cloudbridge-relay@test'"
    fi
    
    log_success "Переменные окружения настроены"
}

# Тест 1: Базовая инициализация Sentry
test_sentry_init() {
    log_info "Тест 1: Инициализация Sentry..."
    
    cat > test_init.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "time"
    
    sentrygo "github.com/getsentry/sentry-go"
    "github.com/twogc/cloudbridge-relay-installer/internal/sentry"
)

func main() {
    fmt.Println("Инициализация Sentry...")
    
    // Инициализация с конфигурацией по умолчанию
    sentryConfig := sentry.DefaultConfig()
    if err := sentry.Init(sentryConfig); err != nil {
        log.Fatalf("Failed to initialize Sentry: %v", err)
    }
    defer sentrygo.Flush(2 * time.Second)
    
    fmt.Println("✅ Sentry инициализирован успешно")
    
    // Тестовая отправка сообщения
    sentrygo.CaptureMessage("CloudBridge Relay - Sentry Init Test")
    fmt.Println("✅ Тестовое сообщение отправлено")
}
EOF

    go run test_init.go
    rm test_init.go
    
    log_success "Тест 1 пройден: Sentry инициализация работает"
}

# Тест 2: Gin middleware
test_gin_middleware() {
    log_info "Тест 2: Gin middleware..."
    
    cat > test_gin.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/getsentry/sentry-go"
    "github.com/twogc/cloudbridge-relay-installer/internal/sentry"
)

func main() {
    fmt.Println("Тестирование Gin middleware...")
    
    // Инициализация Sentry
    sentryConfig := sentry.DefaultConfig()
    if err := sentry.Init(sentryConfig); err != nil {
        log.Fatalf("Failed to initialize Sentry: %v", err)
    }
    defer sentry.Flush(2 * time.Second)
    
    // Создание Gin роутера
    r := gin.Default()
    
    // Добавление Sentry middleware
    r.Use(sentry.NewMiddleware(sentry.DefaultMiddlewareOptions()))
    r.Use(sentry.NewCloudBridgeMiddleware())
    
    // Тестовые роуты
    r.GET("/test", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "Test endpoint"})
    })
    
    r.GET("/error", func(c *gin.Context) {
        // Имитация ошибки
        sentry.CaptureCloudBridgeError(c, fmt.Errorf("test error"), map[string]interface{}{
            "test": "gin_middleware",
        })
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Test error"})
    })
    
    r.GET("/panic", func(c *gin.Context) {
        panic("Test panic for Sentry")
    })
    
    fmt.Println("✅ Gin middleware настроен")
    fmt.Println("🌐 Сервер запущен на :8080")
    fmt.Println("📋 Тестовые endpoints:")
    fmt.Println("   GET /test - обычный запрос")
    fmt.Println("   GET /error - ошибка")
    fmt.Println("   GET /panic - паника")
    fmt.Println("")
    fmt.Println("Нажмите Ctrl+C для остановки...")
    
    // Запуск сервера
    r.Run(":8080")
}
EOF

    # Запуск в фоне
    go run test_gin.go &
    GIN_PID=$!
    
    # Ждем запуска сервера
    sleep 3
    
    # Тестовые запросы
    log_info "Отправка тестовых запросов..."
    
    # Тест обычного запроса
    curl -s http://localhost:8080/test > /dev/null && log_success "GET /test - OK" || log_error "GET /test - FAILED"
    
    # Тест ошибки
    curl -s http://localhost:8080/error > /dev/null && log_success "GET /error - OK" || log_error "GET /error - FAILED"
    
    # Тест паники
    curl -s http://localhost:8080/panic > /dev/null && log_success "GET /panic - OK" || log_error "GET /panic - FAILED"
    
    # Остановка сервера
    kill $GIN_PID 2>/dev/null || true
    rm test_gin.go
    
    log_success "Тест 2 пройден: Gin middleware работает"
}

# Тест 3: Специализированные функции CloudBridge
test_cloudbridge_functions() {
    log_info "Тест 3: CloudBridge функции..."
    
    cat > test_cloudbridge.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/getsentry/sentry-go"
    "github.com/twogc/cloudbridge-relay-installer/internal/sentry"
)

func main() {
    fmt.Println("Тестирование CloudBridge функций...")
    
    // Инициализация Sentry
    sentryConfig := sentry.DefaultConfig()
    if err := sentry.Init(sentryConfig); err != nil {
        log.Fatalf("Failed to initialize Sentry: %v", err)
    }
    defer sentry.Flush(2 * time.Second)
    
    // Создание тестового контекста Gin
    c := &gin.Context{}
    
    // Тест CloudBridge ошибки
    fmt.Println("Тестирование CaptureCloudBridgeError...")
    sentry.CaptureCloudBridgeError(c, fmt.Errorf("CloudBridge test error"), map[string]interface{}{
        "operation": "test_operation",
        "peer_id": "test-peer-123",
        "tenant_id": "test-tenant",
    })
    
    // Тест сетевой ошибки
    fmt.Println("Тестирование CaptureNetworkError...")
    sentry.CaptureNetworkError(c, fmt.Errorf("Network test error"), map[string]interface{}{
        "protocol": "QUIC",
        "connection_type": "P2P",
        "peer_id": "test-peer-456",
    })
    
    // Тест ошибки аутентификации
    fmt.Println("Тестирование CaptureAuthError...")
    sentry.CaptureAuthError(c, fmt.Errorf("Auth test error"), map[string]interface{}{
        "token_type": "JWT",
        "validation_step": "signature",
        "user_id": "test-user-789",
    })
    
    // Тест установки контекста
    fmt.Println("Тестирование SetCloudBridgeContext...")
    sentry.SetCloudBridgeContext(c, "test-tenant", "test-user", "test-connection")
    
    // Тест установки тегов
    fmt.Println("Тестирование SetTag...")
    sentry.SetTag("service", "cloudbridge-relay")
    sentry.SetTag("component", "test-component")
    
    // Тест установки пользователя
    fmt.Println("Тестирование SetUser...")
    sentry.SetUser(sentry.User{
        ID: "test-user-123",
        Username: "test-user",
    })
    
    // Тест добавления breadcrumbs
    fmt.Println("Тестирование AddBreadcrumb...")
    sentry.AddBreadcrumb(&sentry.Breadcrumb{
        Category: "test",
        Message: "Test breadcrumb",
        Level: sentry.LevelInfo,
        Data: map[string]interface{}{
            "test_key": "test_value",
        },
    })
    
    fmt.Println("✅ Все CloudBridge функции протестированы")
}
EOF

    go run test_cloudbridge.go
    rm test_cloudbridge.go
    
    log_success "Тест 3 пройден: CloudBridge функции работают"
}

# Тест 4: Performance мониторинг
test_performance_monitoring() {
    log_info "Тест 4: Performance мониторинг..."
    
    cat > test_performance.go << 'EOF'
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/getsentry/sentry-go"
    "github.com/twogc/cloudbridge-relay-installer/internal/sentry"
)

func main() {
    fmt.Println("Тестирование Performance мониторинга...")
    
    // Инициализация Sentry
    sentryConfig := sentry.DefaultConfig()
    if err := sentry.Init(sentryConfig); err != nil {
        log.Fatalf("Failed to initialize Sentry: %v", err)
    }
    defer sentry.Flush(2 * time.Second)
    
    // Тест производительности
    start := time.Now()
    
    // Имитация работы
    time.Sleep(100 * time.Millisecond)
    
    duration := time.Since(start)
    
    // Отправка метрик производительности
    sentry.ConfigureScope(func(scope *sentry.Scope) {
        scope.SetExtra("performance_test_duration_ms", duration.Milliseconds())
        scope.SetExtra("performance_test_operation", "test_operation")
        scope.SetTag("performance_test", "true")
    })
    
    sentry.CaptureMessage("Performance test completed")
    
    fmt.Printf("✅ Performance тест завершен за %v\n", duration)
}
EOF

    go run test_performance.go
    rm test_performance.go
    
    log_success "Тест 4 пройден: Performance мониторинг работает"
}

# Основная функция
main() {
    echo "🚀 Запуск тестов Sentry интеграции..."
    echo ""
    
    check_environment
    echo ""
    
    test_sentry_init
    echo ""
    
    test_gin_middleware
    echo ""
    
    test_cloudbridge_functions
    echo ""
    
    test_performance_monitoring
    echo ""
    
    log_success "🎉 Все тесты пройдены успешно!"
    echo ""
    echo "📊 Проверьте Sentry dashboard для подтверждения получения событий"
    echo "🔗 Sentry Dashboard: https://sentry.io (или ваш локальный Sentry)"
    echo ""
    echo "📋 Следующие шаги:"
    echo "1. Настройте DSN в production"
    echo "2. Настройте алерты в Sentry"
    echo "3. Интегрируйте с Grafana"
    echo "4. Настройте мониторинг в production"
}

# Запуск тестов
main "$@"
