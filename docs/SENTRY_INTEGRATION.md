# 🚨 Sentry Integration для CloudBridge Relay

## Обзор

CloudBridge Relay интегрирован с Sentry для комплексного мониторинга ошибок в реальном времени. Эта интеграция обеспечивает:

- **Real-time error tracking** - мгновенное обнаружение ошибок
- **Performance monitoring** - отслеживание производительности
- **User context** - контекст пользователей и сессий
- **Release tracking** - отслеживание версий
- **Custom dashboards** - специализированные дашборды

## 🚀 Быстрый старт

### 1. Настройка переменных окружения

```bash
export SENTRY_DSN="https://your-dsn@sentry.io/project-id"
export SENTRY_ENVIRONMENT="production"
export SENTRY_RELEASE="cloudbridge-relay@1.0.0"
export SENTRY_DEBUG="false"
```

### 2. Запуск с Sentry

```bash
# Запуск основного сервера
go run cmd/relay/main.go

# Запуск с Docker Compose (включая Sentry)
docker-compose -f docker-compose.sentry.yml up -d
```

## 📊 Возможности мониторинга

### **Error Tracking**
- Автоматическое отслеживание паник и ошибок
- Контекстная информация об ошибках
- Stack traces с исходным кодом
- Группировка похожих ошибок

### **Performance Monitoring**
- Время выполнения запросов
- Database query performance
- External API calls
- Memory и CPU usage

### **User Context**
- Tenant ID tracking
- User ID tracking
- Connection ID tracking
- Protocol information

### **Custom Metrics**
- P2P connection success rate
- QUIC protocol performance
- MASQUE tunnel efficiency
- Network latency metrics

## 🔧 Конфигурация

### **Основные настройки**

```yaml
# configs/sentry.yaml
sentry:
  dsn: "${SENTRY_DSN}"
  environment: "production"
  release: "cloudbridge-relay@1.0.0"
  traces_sample_rate: 0.1  # 10% в production
  enable_tracing: true
  enable_logs: true
```

### **Middleware настройки**

```go
// Настройка Sentry middleware
router.Use(sentry.NewMiddleware(&sentry.MiddlewareOptions{
    Repanic:         true,
    WaitForDelivery: false,
    Timeout:         30 * time.Second,
    EnableTracing:   true,
}))
```

## 📈 Использование в коде

### **Отправка ошибок**

```go
// В обработчиках Gin
func (h *Handler) HandleRequest(c *gin.Context) {
    if err := someOperation(); err != nil {
        // Отправка ошибки CloudBridge
        sentry.CaptureCloudBridgeError(c, err, map[string]interface{}{
            "operation": "handle_request",
            "peer_id": c.Param("peer_id"),
        })
        c.JSON(500, gin.H{"error": "Internal server error"})
        return
    }
}
```

### **Отправка сетевых ошибок**

```go
// Для сетевых операций
func (m *NetworkManager) Connect(peerID string) error {
    if err := m.establishConnection(peerID); err != nil {
        sentry.CaptureNetworkError(c, err, map[string]interface{}{
            "peer_id": peerID,
            "protocol": "QUIC",
            "connection_type": "P2P",
        })
        return err
    }
    return nil
}
```

### **Отправка ошибок аутентификации**

```go
// Для ошибок аутентификации
func (a *AuthManager) ValidateToken(token string) error {
    if err := a.validateJWT(token); err != nil {
        sentry.CaptureAuthError(c, err, map[string]interface{}{
            "token_type": "JWT",
            "validation_step": "signature",
        })
        return err
    }
    return nil
}
```

## 🎯 Специализированные функции

### **CloudBridge Context**

```go
// Установка контекста CloudBridge
sentry.SetCloudBridgeContext(c, tenantID, userID, connectionID)

// Установка тегов
sentry.SetTag("service", "cloudbridge-relay")
sentry.SetTag("component", "p2p-manager")

// Установка пользователя
sentry.SetUser(sentry.User{
    ID: userID,
    Username: "cloudbridge-user",
})
```

### **Custom Breadcrumbs**

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

## 📊 Дашборды и алерты

### **Sentry Dashboard**
- Error rate trends
- Performance metrics
- User impact analysis
- Release health

### **Grafana Integration**
- Custom dashboards
- Alert rules
- Performance monitoring
- Network metrics

### **Alert Rules**
```yaml
# monitoring/rules/sentry-alerts.yml
groups:
- name: sentry-alerts
  rules:
  - alert: HighErrorRate
    expr: rate(sentry_errors_total[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
```

## 🔒 Безопасность и приватность

### **Data Filtering**
```go
// Фильтрация чувствительных данных
sentryConfig.BeforeSend = func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
    // Удаляем чувствительные данные
    if event.User != nil {
        event.User.Email = ""
    }
    
    // Маскируем IP адреса
    if event.Request != nil {
        event.Request.Env["REMOTE_ADDR"] = "***.***.***.***"
    }
    
    return event
}
```

### **Environment Separation**
- Development: полное логирование
- Staging: ограниченное логирование
- Production: только критические ошибки

## 🚀 Deployment

### **Docker Compose**
```bash
# Запуск Sentry локально
docker-compose -f docker-compose.sentry.yml up -d

# Проверка статуса
docker-compose -f docker-compose.sentry.yml ps
```

### **Kubernetes**
```yaml
# k8s/sentry/sentry-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sentry
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sentry
  template:
    metadata:
      labels:
        app: sentry
    spec:
      containers:
      - name: sentry
        image: sentry:latest
        env:
        - name: SENTRY_DSN
          valueFrom:
            secretKeyRef:
              name: sentry-secret
              key: dsn
```

## 📚 Полезные ссылки

- [Sentry Go SDK Documentation](https://docs.sentry.io/platforms/go/)
- [Sentry Gin Integration](https://docs.sentry.io/platforms/go/guides/gin/)
- [CloudBridge Relay Documentation](../README.md)
- [Monitoring Setup Guide](../monitoring/README.md)

## 🆘 Troubleshooting

### **Частые проблемы**

1. **Sentry не инициализируется**
   ```bash
   # Проверьте DSN
   echo $SENTRY_DSN
   
   # Проверьте логи
   docker logs sentry-web
   ```

2. **Ошибки не отправляются**
   ```go
   // Проверьте инициализацию
   if err := sentry.Init(sentryConfig); err != nil {
       log.Printf("Sentry init failed: %v", err)
   }
   ```

3. **Performance issues**
   ```yaml
   # Уменьшите sample rate
   traces_sample_rate: 0.01  # 1%
   ```

---

**🎯 Sentry интеграция готова к использованию!** Теперь CloudBridge Relay имеет enterprise-уровень мониторинга ошибок.



