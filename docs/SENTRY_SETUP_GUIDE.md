# 🚀 Sentry Setup Guide для CloudBridge Relay

## 📋 Пошаговое руководство по настройке Sentry

### **Шаг 1: Настройка Sentry проекта**

#### **1.1. Создание проекта в Sentry**
1. Перейдите на [https://sentry.io](https://sentry.io)
2. Войдите в аккаунт или создайте новый
3. Создайте новый проект:
   - **Platform**: Go
   - **Project Name**: CloudBridge Relay
   - **Team**: Выберите команду
4. Скопируйте **DSN** из настроек проекта

#### **1.2. Настройка переменных окружения**
```bash
# Основные переменные
export SENTRY_DSN="https://your-dsn@sentry.io/project-id"
export SENTRY_ENVIRONMENT="production"
export SENTRY_RELEASE="cloudbridge-relay@1.0.0"
export SENTRY_DEBUG="false"
export SENTRY_SERVER_NAME="cloudbridge-relay"
```

### **Шаг 2: Локальная настройка**

#### **2.1. Автоматическая настройка**
```bash
# Запуск интерактивного скрипта настройки
./scripts/setup-sentry.sh
```

#### **2.2. Ручная настройка**
```bash
# 1. Создание .env файла
cp configs/sentry.yaml .env.sentry

# 2. Редактирование конфигурации
nano .env.sentry

# 3. Запуск Sentry локально
docker-compose -f docker-compose.sentry.yml up -d
```

### **Шаг 3: Тестирование интеграции**

#### **3.1. Запуск тестов**
```bash
# Запуск всех тестов Sentry
./scripts/test-sentry-integration.sh
```

#### **3.2. Ручное тестирование**
```bash
# 1. Установка переменных
export SENTRY_DSN="your-dsn-here"

# 2. Запуск тестового сервера
go run cmd/relay/main.go

# 3. Отправка тестовых запросов
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/health
```

### **Шаг 4: Production развертывание**

#### **4.1. Kubernetes**
```bash
# 1. Применение конфигурации
kubectl apply -f k8s/sentry/sentry-configmap.yaml
kubectl apply -f k8s/sentry/sentry-secret.yaml

# 2. Обновление deployment
kubectl patch deployment cloudbridge-relay -p '{"spec":{"template":{"spec":{"containers":[{"name":"cloudbridge-relay","envFrom":[{"configMapRef":{"name":"sentry-config"}},{"secretRef":{"name":"sentry-secret"}}]}]}}}}'
```

#### **4.2. Docker**
```bash
# 1. Сборка образа
docker build -t cloudbridge-relay:sentry .

# 2. Запуск с переменными окружения
docker run -d \
  --name cloudbridge-relay \
  -e SENTRY_DSN="your-dsn-here" \
  -e SENTRY_ENVIRONMENT="production" \
  -p 8080:8080 \
  cloudbridge-relay:sentry
```

### **Шаг 5: Мониторинг и алерты**

#### **5.1. Настройка алертов в Sentry**
1. Перейдите в **Alerts** → **Create Alert Rule**
2. Настройте условия:
   - **Error Rate** > 5%
   - **Response Time** > 1s
   - **Memory Usage** > 80%
3. Настройте уведомления:
   - Email
   - Slack
   - PagerDuty

#### **5.2. Интеграция с Grafana**
```yaml
# grafana-datasource-sentry.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-datasource-sentry
data:
  sentry-datasource.yaml: |
    apiVersion: 1
    datasources:
    - name: Sentry
      type: sentry
      url: https://sentry.io
      access: proxy
      jsonData:
        orgSlug: your-org-slug
        projectSlug: cloudbridge-relay
```

### **Шаг 6: Оптимизация и настройка**

#### **6.1. Настройка sampling rates**
```yaml
# configs/sentry.yaml
sentry:
  traces_sample_rate: 0.1  # 10% в production
  sample_rate: 1.0        # 100% для ошибок
  max_breadcrumbs: 50      # Максимум breadcrumbs
```

#### **6.2. Фильтрация данных**
```go
// internal/sentry/config.go
sentryConfig.BeforeSend = func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
    // Фильтрация чувствительных данных
    if event.User != nil {
        event.User.Email = ""
    }
    
    // Маскировка IP адресов
    if event.Request != nil {
        event.Request.Env["REMOTE_ADDR"] = "***.***.***.***"
    }
    
    return event
}
```

### **Шаг 7: Troubleshooting**

#### **7.1. Частые проблемы**

**Проблема**: Sentry не инициализируется
```bash
# Решение: Проверьте DSN
echo $SENTRY_DSN
# Должен быть: https://your-dsn@sentry.io/project-id
```

**Проблема**: Ошибки не отправляются
```bash
# Решение: Проверьте логи
docker logs cloudbridge-relay
# Ищите: "Sentry initialization failed"
```

**Проблема**: Performance issues
```yaml
# Решение: Уменьшите sampling rate
traces_sample_rate: 0.01  # 1% вместо 10%
```

#### **7.2. Полезные команды**

```bash
# Проверка статуса Sentry
curl -s http://localhost:9000/api/0/health/

# Проверка логов
docker logs sentry-web
docker logs sentry-worker

# Проверка метрик
curl -s http://localhost:9090/metrics | grep sentry
```

### **Шаг 8: Мониторинг в production**

#### **8.1. Ключевые метрики**
- **Error Rate**: < 1%
- **Response Time**: < 200ms
- **Memory Usage**: < 80%
- **CPU Usage**: < 70%

#### **8.2. Алерты**
- **Critical**: Error rate > 5%
- **Warning**: Response time > 1s
- **Info**: Memory usage > 80%

### **🎯 Результат**

После выполнения всех шагов у вас будет:

✅ **Полноценный мониторинг ошибок** в реальном времени  
✅ **Performance метрики** для оптимизации  
✅ **Автоматические алерты** при проблемах  
✅ **Детальная аналитика** для debugging  
✅ **Production-ready** конфигурация  

### **📚 Дополнительные ресурсы**

- [Sentry Go SDK Documentation](https://docs.sentry.io/platforms/go/)
- [Sentry Gin Integration](https://docs.sentry.io/platforms/go/guides/gin/)
- [CloudBridge Relay Documentation](../README.md)
- [Monitoring Setup Guide](../monitoring/README.md)

---

**🚀 Sentry интеграция готова к production использованию!**



