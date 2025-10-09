# 🚀 Sentry Quick Start для CloudBridge Relay

## ✅ **Sentry интеграция готова!**

### **📋 Что уже настроено:**

1. ✅ **Sentry SDK установлен** (`github.com/getsentry/sentry-go/gin`)
2. ✅ **Конфигурация создана** (`internal/sentry/`)
3. ✅ **Middleware интегрирован** в роутеры
4. ✅ **Переменные окружения** добавлены в `.env`
5. ✅ **Тестирование работает** ✅

### **🚀 Как использовать прямо сейчас:**

#### **1. Загрузка переменных окружения:**
```bash
# Загрузите переменные из .env файла
source scripts/load-env.sh
```

#### **2. Запуск сервера с Sentry:**
```bash
# Запуск с мониторингом ошибок
go run cmd/relay/main.go
```

#### **3. Отправка ошибок в коде:**
```go
// В любом обработчике Gin
import "github.com/twogc/cloudbridge-relay-installer/internal/sentry"

// Отправка CloudBridge ошибки
sentry.CaptureCloudBridgeError(c, err, map[string]interface{}{
    "operation": "p2p_connection",
    "peer_id": peerID,
})

// Отправка сетевой ошибки
sentry.CaptureNetworkError(c, err, map[string]interface{}{
    "protocol": "QUIC",
    "connection_type": "P2P",
})

// Отправка ошибки аутентификации
sentry.CaptureAuthError(c, err, map[string]interface{}{
    "token_type": "JWT",
    "user_id": userID,
})
```

### **📊 Мониторинг в реальном времени:**

#### **Sentry Dashboard:**
- **URL**: https://sentry.io (ваш проект)
- **Events**: Все ошибки в реальном времени
- **Performance**: Метрики производительности
- **Releases**: Отслеживание версий

#### **Ключевые метрики:**
- **Error Rate**: Количество ошибок в секунду
- **Response Time**: Время выполнения запросов
- **Memory Usage**: Использование памяти
- **User Impact**: Влияние на пользователей

### **🔧 Настройка для production:**

#### **Обновите .env файл:**
```bash
# Для production
SENTRY_ENVIRONMENT=production
SENTRY_DEBUG=false
SENTRY_TRACES_SAMPLE_RATE=0.1  # 10% в production
```

#### **Kubernetes deployment:**
```bash
# Применение конфигурации
kubectl apply -f k8s/sentry/sentry-configmap.yaml
kubectl apply -f k8s/sentry/sentry-secret.yaml

# Обновление deployment
kubectl patch deployment cloudbridge-relay -p '{"spec":{"template":{"spec":{"containers":[{"name":"cloudbridge-relay","envFrom":[{"configMapRef":{"name":"sentry-config"}},{"secretRef":{"name":"sentry-secret"}}]}]}}}}'
```

### **📈 Продвинутые возможности:**

#### **1. Custom теги:**
```go
// Установка тегов для контекста
sentry.SetTag("service", "cloudbridge-relay")
sentry.SetTag("component", "p2p-manager")
sentry.SetTag("tenant_id", tenantID)
```

#### **2. User context:**
```go
// Установка пользователя
sentry.SetUser(sentry.User{
    ID: userID,
    Username: "cloudbridge-user",
})
```

#### **3. Breadcrumbs:**
```go
// Добавление breadcrumbs для отслеживания
sentry.AddBreadcrumb(&sentry.Breadcrumb{
    Category: "p2p",
    Message: "P2P connection established",
    Level: sentry.LevelInfo,
    Data: map[string]interface{}{
        "peer_id": peerID,
        "protocol": "QUIC",
    },
})
```

### **🚨 Алерты и уведомления:**

#### **Настройка алертов в Sentry:**
1. Перейдите в **Alerts** → **Create Alert Rule**
2. Настройте условия:
   - **Error Rate** > 5%
   - **Response Time** > 1s
   - **Memory Usage** > 80%
3. Настройте уведомления:
   - Email
   - Slack
   - PagerDuty

### **🔍 Troubleshooting:**

#### **Частые проблемы:**

**Проблема**: Sentry не инициализируется
```bash
# Решение: Проверьте DSN
echo $SENTRY_DSN
# Должен быть: https://your-dsn@sentry.io/project-id
```

**Проблема**: Ошибки не отправляются
```bash
# Решение: Проверьте логи
go run cmd/relay/main.go 2>&1 | grep sentry
```

**Проблема**: Performance issues
```bash
# Решение: Уменьшите sampling rate в .env
SENTRY_TRACES_SAMPLE_RATE=0.01  # 1% вместо 10%
```

### **📚 Полезные команды:**

```bash
# Загрузка переменных окружения
source scripts/load-env.sh

# Тестирование Sentry
go run test-sentry-simple.go

# Проверка статуса
curl -s http://localhost:8080/health

# Просмотр логов
docker logs cloudbridge-relay | grep sentry
```

### **🎯 Результат:**

Теперь у вас есть:

✅ **Real-time error tracking** - мгновенные уведомления  
✅ **Performance monitoring** - отслеживание производительности  
✅ **User context** - контекст пользователей и сессий  
✅ **Custom metrics** - специализированные метрики CloudBridge  
✅ **Production-ready** конфигурация  

### **🚀 Следующие шаги:**

1. **Настройте алерты** в Sentry dashboard
2. **Интегрируйте с Grafana** для визуализации
3. **Настройте уведомления** (Slack, Email)
4. **Оптимизируйте sampling rates** для production
5. **Документируйте процессы** для команды

---

**🎉 Sentry интеграция готова к production использованию!** 

**📊 Перейдите на https://sentry.io и проверьте ваш dashboard для подтверждения получения событий!**



