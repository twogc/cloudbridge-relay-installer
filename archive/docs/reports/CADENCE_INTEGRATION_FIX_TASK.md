# Техническое задание: Исправление интеграции AI Service с Cadence

## Цель
Исправить интеграцию AI Service с Cadence для обеспечения полного потока данных и workflow управления.

## Текущие проблемы

### 1. Cadence не отвечает на HTTP запросы
- **Проблема**: AI Service не может подключиться к Cadence через HTTP
- **Порт**: 7934 (web API)
- **Ошибка**: Connection refused / timeout
- **Статус**: Cadence работает, но HTTP API недоступен

### 2. Отсутствует gRPC интеграция
- **Проблема**: AI Service пытается использовать HTTP, но Cadence работает через gRPC
- **Порт**: 7933 (gRPC)
- **Статус**: Требуется настройка gRPC клиента

## Диагностика

### 1. Проверить статус Cadence
```bash
# Проверить поды
kubectl get pods -n cloudbridge | grep cadence

# Проверить сервисы
kubectl get services -n cloudbridge | grep cadence

# Проверить логи Cadence
kubectl logs deployment/cadence -n cloudbridge --tail=50
```

### 2. Проверить порты Cadence
```bash
# Проверить открытые порты
kubectl exec deployment/cadence -n cloudbridge -- netstat -tlnp | grep LISTEN

# Проверить HTTP API
kubectl exec deployment/cadence -n cloudbridge -- curl -s http://localhost:7934/health

# Проверить gRPC
kubectl exec deployment/cadence -n cloudbridge -- curl -s http://localhost:7933/health
```

### 3. Проверить конфигурацию Cadence
```bash
# Проверить конфигурацию
kubectl exec deployment/cadence -n cloudbridge -- cat /etc/cadence/config/cadence.yaml

# Проверить переменные окружения
kubectl exec deployment/cadence -n cloudbridge -- env | grep -i cadence
```

## Требования к исправлению

### 1. Настройка HTTP API Cadence

#### Проверить конфигурацию Cadence:
```yaml
# /etc/cadence/config/cadence.yaml
services:
  frontend:
    rpc:
      grpcPort: 7933
      membershipPort: 7939
    pprof:
      port: 7934
    metrics:
      prometheus:
        timerType: histogram
        listenAddress: 0.0.0.0:7934
```

#### Убедиться, что HTTP API включен:
```yaml
services:
  frontend:
    http:
      port: 7934
      bindOnLocalHost: false
      bindOnIP: "0.0.0.0"
```

### 2. Настройка gRPC клиента в AI Service

#### Обновить конфигурацию AI Service:
```yaml
# configs/ai-config.yaml
cadence:
  domain: "cloudbridge"
  host: "cadence-service.cloudbridge.svc.cluster.local"
  port: 7933
  protocol: "grpc" # Изменить с http на grpc
  timeout: "30s"
```

#### Обновить код интеграции:
```go
// internal/ai/cadence/cadence_integration.go
func NewCadenceIntegration(logger *zap.Logger, host string, port int, domain string) *CadenceIntegration {
    // Настроить gRPC клиент вместо HTTP
    conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())
    if err != nil {
        logger.Error("Failed to connect to Cadence", zap.Error(err))
        return nil
    }

    client := workflowservice.NewWorkflowServiceClient(conn)

    return &CadenceIntegration{
        logger: logger,
        client: client,
        domain: domain,
    }
}
```

### 3. Добавить необходимые зависимости

#### Обновить go.mod:
```go
// go.mod
require (
    go.uber.org/cadence v0.20.0
    google.golang.org/grpc v1.50.1
    google.golang.org/protobuf v1.28.1
)
```

#### Добавить импорты:
```go
// internal/ai/cadence/cadence_integration.go
import (
    "go.uber.org/cadence/client"
    "go.uber.org/cadence/worker"
    "go.uber.org/cadence/workflow"
    "google.golang.org/grpc"
    "go.uber.org/cadence/.gen/go/cadence/workflowservice"
)
```

## План исправления

### 1. Диагностика (30 минут)
- [ ] Проверить статус Cadence
- [ ] Проверить логи Cadence
- [ ] Проверить конфигурацию Cadence
- [ ] Проверить порты и сеть

### 2. Исправление HTTP API (1 час)
- [ ] Обновить конфигурацию Cadence
- [ ] Перезапустить Cadence
- [ ] Проверить HTTP API
- [ ] Протестировать подключение

### 3. Настройка gRPC клиента (2 часа)
- [ ] Обновить зависимости в go.mod
- [ ] Изменить код интеграции с HTTP на gRPC
- [ ] Обновить конфигурацию AI Service
- [ ] Пересобрать AI Service

### 4. Тестирование (1 час)
- [ ] Проверить подключение AI Service к Cadence
- [ ] Протестировать workflow функции
- [ ] Проверить логи на ошибки
- [ ] Убедиться в работоспособности

## Конкретные задачи

### 1. Проверить конфигурацию Cadence
```bash
# Проверить текущую конфигурацию
kubectl exec deployment/cadence -n cloudbridge -- cat /etc/cadence/config/cadence.yaml

# Проверить, включен ли HTTP API
grep -A 10 -B 5 "http:" /etc/cadence/config/cadence.yaml
```

### 2. Обновить конфигурацию Cadence
```yaml
# Добавить в конфигурацию Cadence
services:
  frontend:
    http:
      port: 7934
      bindOnLocalHost: false
      bindOnIP: "0.0.0.0"
    rpc:
      grpcPort: 7933
      membershipPort: 7939
```

### 3. Обновить код AI Service
```go
// internal/ai/cadence/cadence_integration.go
func (ci *CadenceIntegration) HealthCheck(ctx context.Context) error {
    // Использовать gRPC клиент вместо HTTP
    req := &workflowservice.GetClusterInfoRequest{}
    _, err := ci.client.GetClusterInfo(ctx, req)
    return err
}
```

### 4. Обновить конфигурацию AI Service
```yaml
# configs/ai-config.yaml
cadence:
  domain: "cloudbridge"
  host: "cadence-service.cloudbridge.svc.cluster.local"
  port: 7933
  protocol: "grpc"
  timeout: "30s"
```

## Ожидаемые результаты

### 1. HTTP API Cadence
- `curl http://cadence-service:7934/health` возвращает статус
- AI Service может подключиться через HTTP
- Метрики Cadence доступны

### 2. gRPC интеграция
- AI Service подключается к Cadence через gRPC
- Workflow функции работают
- Cadence управляет AI процессами

### 3. Полный поток данных
- Relay → Prometheus → AI → Cadence
- AI Service создает workflow в Cadence
- Cadence управляет AI операциями

## Критерии приемки

1. **HTTP API работает**: `curl http://cadence-service:7934/health` возвращает статус
2. **gRPC подключение**: AI Service подключается к Cadence через gRPC
3. **Workflow функции**: AI Service может создавать и управлять workflow
4. **Логи без ошибок**: Нет ошибок подключения в логах AI Service
5. **Полная интеграция**: Cadence управляет AI процессами

## План внедрения

1. **Подготовка** (30 минут)
   - Создать backup текущей конфигурации
   - Подготовить новую конфигурацию
   - Протестировать в dev среде

2. **Исправление Cadence** (1 час)
   - Обновить конфигурацию Cadence
   - Перезапустить Cadence
   - Проверить HTTP API

3. **Обновление AI Service** (2 часа)
   - Обновить зависимости
   - Изменить код интеграции
   - Пересобрать и развернуть

4. **Тестирование** (1 час)
   - Проверить подключение
   - Протестировать workflow
   - Убедиться в работоспособности

## Контакты

- **Ответственный**: Разработчик Backend
- **Куратор**: AI Team Lead
- **Срок выполнения**: 1 рабочий день
- **Приоритет**: Высокий

## Полезные ссылки

- [Cadence HTTP API Documentation](https://cadenceworkflow.io/docs/operation-guide/tools/cadence-web/)
- [Cadence gRPC Client](https://cadenceworkflow.io/docs/go-client/)
- [Kubernetes Service Discovery](https://kubernetes.io/docs/concepts/services-networking/service/)

---

*Документ создан: 2025-09-22*
*Версия: 1.0*
*Статус: Готов к выполнению*
