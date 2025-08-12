# 🚀 Рекомендации по производительности Cadence в CloudBridge Relay

## 📊 Обзор

Данный документ содержит рекомендации по оптимизации производительности Cadence Workflow в CloudBridge Relay, включая профилирование, мониторинг и отладку.

## 🔧 Профилирование

### Включение профилирования

Профилирование включается через конфигурацию:

```yaml
cadence:
  profiling:
    enabled: true                   # Включить профилирование
    port: ":6060"                  # Порт для pprof сервера
    enableMemoryProfiling: true     # Профилирование памяти
    enableCPUProfiling: true        # Профилирование CPU
    enableGoroutineProfiling: true  # Профилирование горутин
```

### Доступные endpoints

После запуска worker'а доступны следующие endpoints для профилирования:

#### Стандартные pprof endpoints:
- `http://localhost:6060/debug/pprof/` - общая информация
- `http://localhost:6060/debug/pprof/heap` - heap профиль
- `http://localhost:6060/debug/pprof/goroutine` - профиль горутин
- `http://localhost:6060/debug/pprof/allocs` - профиль аллокаций
- `http://localhost:6060/debug/pprof/block` - профиль блокировок
- `http://localhost:6060/debug/pprof/mutex` - профиль мьютексов
- `http://localhost:6060/debug/pprof/profile` - CPU профиль (30 секунд)

#### Кастомные endpoints:
- `http://localhost:6060/debug/cadence/stats` - статистика Cadence worker'а
- `http://localhost:6060/debug/cadence/workflows` - информация о зарегистрированных workflows
- `http://localhost:6060/debug/cadence/activities` - информация о зарегистрированных activities

### Использование профилирования

#### Анализ памяти:
```bash
# Получить heap профиль
go tool pprof http://localhost:6060/debug/pprof/heap

# Визуализировать граф потребления памяти
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/heap
```

#### Анализ CPU:
```bash
# Получить CPU профиль (30 секунд)
go tool pprof http://localhost:6060/debug/pprof/profile

# Анализ горутин
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

#### Анализ блокировок:
```bash
# Анализ блокировок
go tool pprof http://localhost:6060/debug/pprof/block

# Анализ мьютексов
go tool pprof http://localhost:6060/debug/pprof/mutex
```

## 🏥 Health Checks

### Конфигурация health checks

```yaml
cadence:
  health:
    enabled: true                   # Включить health endpoints
    port: ":8080"                  # Порт для health checks
    readinessTimeout: "10s"        # Timeout для readiness probe
    livenessTimeout: "5s"          # Timeout для liveness probe
    enableDetailedHealth: true      # Детальные health checks
```

### Доступные health endpoints

- `http://localhost:8080/health` - общий health check
- `http://localhost:8080/readiness` - readiness probe для Kubernetes
- `http://localhost:8080/liveness` - liveness probe для Kubernetes

### Интеграция с Kubernetes

```yaml
# deployment.yaml
spec:
  containers:
  - name: cadence-worker
    image: cloudbridge-relay:latest
    ports:
    - containerPort: 8080
      name: health
    livenessProbe:
      httpGet:
        path: /liveness
        port: health
      initialDelaySeconds: 30
      periodSeconds: 10
      timeoutSeconds: 5
    readinessProbe:
      httpGet:
        path: /readiness
        port: health
      initialDelaySeconds: 5
      periodSeconds: 5
      timeoutSeconds: 10
```

## 📊 Мониторинг производительности

### Ключевые метрики

#### Workflow метрики:
- `cadence_workflow_executions_total` - общее количество выполнений workflow
- `cadence_workflow_duration_seconds` - длительность выполнения workflow
- `cadence_workflow_errors_total` - количество ошибок workflow

#### Activity метрики:
- `cadence_activity_executions_total` - общее количество выполнений activity
- `cadence_activity_duration_seconds` - длительность выполнения activity
- `cadence_activity_errors_total` - количество ошибок activity

#### Worker метрики:
- `cadence_worker_connections` - количество подключений worker'а
- `cadence_worker_pollers` - количество poller'ов
- `cadence_worker_queue_depth` - глубина очереди задач

### Настройка алертов

```yaml
cadence:
  monitoring:
    alertThresholds:
      workflowFailureRate: 0.05     # 5% failure rate
      activityLatency: "30s"        # Максимальная задержка activity
      queueDepth: 1000              # Максимальная глубина очереди
```

### Prometheus правила алертинга

```yaml
# cadence-alerts.yml
groups:
- name: cadence
  rules:
  - alert: CadenceHighFailureRate
    expr: (rate(cadence_workflow_errors_total[5m]) / rate(cadence_workflow_executions_total[5m])) > 0.05
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High Cadence workflow failure rate"
      description: "Cadence workflow failure rate is {{ $value }}%"

  - alert: CadenceHighLatency
    expr: histogram_quantile(0.95, rate(cadence_activity_duration_seconds_bucket[5m])) > 30
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High Cadence activity latency"
      description: "95th percentile latency is {{ $value }}s"

  - alert: CadenceQueueDepthHigh
    expr: cadence_worker_queue_depth > 1000
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "Cadence worker queue depth too high"
      description: "Queue depth is {{ $value }}"
```

## 🧪 Benchmark тесты

### Запуск benchmark тестов

```bash
# Запуск всех benchmark тестов
go test -bench=. ./internal/cadence/workflows/

# Запуск конкретного benchmark'а
go test -bench=BenchmarkCloudBridgeMainWorkflow ./internal/cadence/workflows/

# Benchmark с профилированием памяти
go test -bench=BenchmarkMemoryUsage -benchmem ./internal/cadence/workflows/

# Benchmark с CPU профилем
go test -bench=. -cpuprofile=cpu.prof ./internal/cadence/workflows/

# Benchmark с memory профилем
go test -bench=. -memprofile=mem.prof ./internal/cadence/workflows/
```

### Анализ результатов

```bash
# Анализ CPU профиля
go tool pprof cpu.prof

# Анализ memory профиля
go tool pprof mem.prof

# Сравнение benchmark'ов
benchcmp old.txt new.txt
```

## 🚀 Интеграционные тесты

### Запуск интеграционных тестов

```bash
# Запуск с реальным Cadence сервером
CADENCE_HOST=localhost:7933 go test -tags=integration ./internal/cadence/integration_test/

# Запуск в Docker окружении
docker-compose up -d cadence
CADENCE_HOST=localhost:7933 go test -tags=integration ./internal/cadence/integration_test/
docker-compose down
```

### Docker Compose для тестирования

```yaml
# docker-compose.test.yml
version: '3.7'
services:
  cadence:
    image: ubercadence/server:0.21.3-auto-setup
    ports:
     - "7933:7933"
     - "7934:7934"
     - "7935:7935"
     - "7939:7939"
    environment:
      - "DB=cassandra"
      - "CASSANDRA_SEEDS=cassandra"
      - "DYNAMIC_CONFIG_FILE_PATH=config/dynamicconfig/development.yaml"
    depends_on:
      - cassandra
    volumes:
      - ./docker/dynamicconfig:/etc/cadence/config/dynamicconfig

  cassandra:
    image: cassandra:3.11
    ports:
      - "9042:9042"
    environment:
      - "CASSANDRA_SEEDS=cassandra"
```

## 🔧 Оптимизация производительности

### Настройки worker'а

```yaml
cadence:
  worker:
    maxConcurrentActivityExecutionSize: 1000    # Увеличить для CPU-intensive activities
    maxConcurrentWorkflowTaskExecutionSize: 1000 # Увеличить для большого количества workflows
    maxConcurrentLocalActivityExecutionSize: 1000 # Оптимизировать для local activities
    maxTaskListActivitiesPerSecond: 100000      # Ограничить throughput
    workerActivitiesPerSecond: 10000            # Ограничить per-worker throughput
```

### Оптимизация workflow'ов

#### 1. Используйте Local Activities для быстрых операций
```go
// Вместо обычного activity
err := workflow.ExecuteActivity(ctx, "FastActivity", input).Get(ctx, &result)

// Используйте local activity
err := workflow.ExecuteLocalActivity(ctx, FastActivity, input).Get(ctx, &result)
```

#### 2. Batch операции
```go
// Вместо множественных activity вызовов
for _, item := range items {
    workflow.ExecuteActivity(ctx, "ProcessItem", item)
}

// Используйте batch processing
workflow.ExecuteActivity(ctx, "ProcessItems", items)
```

#### 3. Parallel execution
```go
// Параллельное выполнение
selector := workflow.NewSelector(ctx)
for _, task := range tasks {
    future := workflow.ExecuteActivity(ctx, "ProcessTask", task)
    selector.AddFuture(future, func(f workflow.Future) {
        // Handle result
    })
}
```

### Оптимизация конфигурации

#### Timeout'ы
```yaml
cadence:
  workflows:
    # Оптимизировать timeout'ы под конкретные use cases
    executionTimeout: "1h"              # Для долгоживущих workflows
    activityStartToCloseTimeout: "5m"   # Для обычных activities
    workflowTaskTimeout: "10s"          # Для decision tasks
```

#### Retry политики
```yaml
cadence:
  retry:
    initialInterval: "1s"       # Начальная задержка
    maximumInterval: "10s"      # Максимальная задержка
    backoffCoefficient: 2.0     # Экспоненциальный backoff
    maximumAttempts: 3          # Максимум попыток
```

## 📈 Мониторинг ресурсов

### Системные метрики

```bash
# CPU usage
top -p $(pgrep cadence-worker)

# Memory usage
ps aux | grep cadence-worker

# Network connections
netstat -an | grep 7933

# File descriptors
lsof -p $(pgrep cadence-worker) | wc -l
```

### Cadence Web UI

Доступ к Cadence Web UI для мониторинга:
- URL: `http://localhost:8088`
- Domains: выберите ваш domain (`cloudbridge-relay`)
- Workflows: мониторинг выполняющихся workflows
- Task Lists: просмотр task lists и их статуса

## 🔍 Отладка проблем

### Типичные проблемы и решения

#### 1. Высокая latency activities
```bash
# Проверить CPU профиль
go tool pprof http://localhost:6060/debug/pprof/profile

# Проверить блокировки
go tool pprof http://localhost:6060/debug/pprof/block
```

#### 2. Memory leaks
```bash
# Анализ heap
go tool pprof http://localhost:6060/debug/pprof/heap

# Сравнение heap'ов во времени
go tool pprof -base heap1.prof heap2.prof
```

#### 3. Deadlocks
```bash
# Анализ горутин
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Детальный dump горутин
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

#### 4. Высокая нагрузка на Cadence сервер
- Проверить количество worker'ов
- Оптимизировать polling frequency
- Увеличить batch size для activities

### Логирование отладки

```yaml
cadence:
  logging:
    level: "debug"                  # Включить debug логи
    logWorkflowHistory: true        # Логировать историю workflows
    logActivityDetails: true        # Логировать детали activities
```

## 📝 Рекомендации по production

### 1. Мониторинг
- Настроить алерты на ключевые метрики
- Использовать distributed tracing
- Регулярно анализировать performance

### 2. Scaling
- Горизонтальное масштабирование worker'ов
- Load balancing между worker'ами
- Partition task lists по нагрузке

### 3. Maintenance
- Регулярные benchmark'и
- Профилирование в production (с осторожностью)
- Мониторинг resource usage

### 4. Security
- Отключить профилирование в production
- Ограничить доступ к health endpoints
- Использовать TLS для Cadence соединений

---

## 🔗 Полезные ссылки

- [Cadence Documentation](https://cadenceworkflow.io/)
- [Go pprof Guide](https://github.com/google/pprof/blob/master/doc/README.md)
- [Prometheus Monitoring](https://prometheus.io/docs/)
- [Kubernetes Health Checks](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
