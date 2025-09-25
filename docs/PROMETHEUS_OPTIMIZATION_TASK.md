# Техническое задание: Оптимизация Prometheus для CloudBridge AI

## 🎯 Цель
Оптимизировать Prometheus для эффективной работы с AI сервисом CloudBridge, устранить перегрузку и обеспечить стабильную передачу данных.

## 📊 Текущие проблемы

### 1. Перегрузка Prometheus
- **692 метрики** собираются каждые 15 секунд
- **251 метрика Kubernetes** (36% от общего количества)
- **36 подов** с аннотацией `prometheus.io/scrape`
- **2768 запросов/минуту** к различным источникам
- API запросы зависают с timeout

### 2. Неэффективное использование ресурсов
- CPU: 250m request, 1 limit
- Memory: 512Mi request, 2Gi limit
- Query timeout: 2m (слишком долго)

## 🎯 Требования к оптимизации

### 1. Сбор только необходимых данных

#### Обязательные метрики для AI сервиса:
```yaml
# CloudBridge Relay метрики
- relay_active_connections
- relay_requests_total
- relay_bandwidth_bytes
- relay_latency_seconds
- relay_errors_total
- relay_connections_total

# AI Service метрики
- ai_anomalies_detected_total
- ai_predictions_made_total
- ai_processing_time_milliseconds
- ai_requests_processed_total
- ai_routes_optimized_total
- ai_service_uptime_seconds

# Cadence метрики
- cadence_workflow_executions_total
- cadence_workflow_failures_total
- cadence_activity_executions_total

# Системные метрики (минимальный набор)
- up
- process_cpu_seconds_total
- process_memory_bytes
- go_goroutines
```

#### Исключить из сбора:
- Kubernetes API server метрики (кроме основных)
- Kubelet метрики (кроме системных)
- Etcd метрики
- CoreDNS метрики
- Метрики от тестовых подов (netshoot, tcpcheck, cleanup jobs)

### 2. Оптимизация конфигурации

#### Новая конфигурация prometheus.yml:
```yaml
global:
  scrape_interval: 30s          # Увеличить с 15s
  evaluation_interval: 30s      # Увеличить с 15s
  query_timeout: 15s            # Уменьшить с 2m

scrape_configs:
  # Prometheus itself
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]
    scrape_interval: 30s

  # CloudBridge Relay - только необходимые метрики
  - job_name: "cloudbridge-relay"
    static_configs:
      - targets: ["cloudbridge-relay.cloudbridge.svc.cluster.local:5551"]
    metrics_path: "/metrics"
    scrape_interval: 30s
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: 'relay_(active_connections|requests_total|bandwidth_bytes|latency_seconds|errors_total|connections_total)'
        action: keep

  # AI Service - только AI метрики
  - job_name: "ai-service"
    static_configs:
      - targets: ["ai-service.cloudbridge.svc.cluster.local:8083"]
    metrics_path: "/metrics"
    scrape_interval: 30s
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: 'ai_(anomalies_detected_total|predictions_made_total|processing_time_milliseconds|requests_processed_total|routes_optimized_total|service_uptime_seconds)'
        action: keep

  # Cadence - только workflow метрики
  - job_name: "cadence"
    static_configs:
      - targets: ["cadence-service.cloudbridge.svc.cluster.local:7934"]
    metrics_path: "/metrics"
    scrape_interval: 30s
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: 'cadence_(workflow_executions_total|workflow_failures_total|activity_executions_total)'
        action: keep

  # Системные метрики - только основные
  - job_name: "system-metrics"
    static_configs:
      - targets: ["localhost:9090"]
    metrics_path: "/metrics"
    scrape_interval: 60s
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: '(up|process_cpu_seconds_total|process_memory_bytes|go_goroutines)'
        action: keep
```

### 3. Увеличение ресурсов

#### Новые лимиты ресурсов:
```yaml
resources:
  requests:
    cpu: 500m        # Увеличить с 250m
    memory: 1Gi      # Увеличить с 512Mi
  limits:
    cpu: 2           # Увеличить с 1
    memory: 4Gi      # Увеличить с 2Gi
```

### 4. Оптимизация параметров запуска

#### Новые аргументы запуска:
```bash
--query.max-concurrency=10
--query.max-samples=100000
--query.lookback-delta=5m
--query.timeout=15s
--storage.tsdb.retention.time=24h
```

## 🔄 Поток данных

### 1. Сбор данных
```
CloudBridge Relay (порт 5551) → Prometheus
AI Service (порт 8083) → Prometheus
Cadence (порт 7934) → Prometheus
```

### 2. Передача данных в AI
```
Prometheus → AI Service (через prometheus-proxy)
AI Service → Анализ и обработка
AI Service → Cadence (workflow)
```

### 3. Мониторинг
```
Prometheus → Grafana (визуализация)
AI Service → Собственные метрики
```

## Задачи разработчика

### 1. Создать оптимизированную конфигурацию
- [ ] Создать новый ConfigMap с оптимизированной конфигурацией
- [ ] Настроить фильтрацию метрик через `metric_relabel_configs`
- [ ] Увеличить интервалы сбора до 30s
- [ ] Уменьшить query timeout до 15s

### 2. Обновить ресурсы Prometheus
- [ ] Увеличить CPU request до 500m
- [ ] Увеличить Memory request до 1Gi
- [ ] Увеличить CPU limit до 2
- [ ] Увеличить Memory limit до 4Gi

### 3. Оптимизировать параметры запуска
- [ ] Добавить `--query.max-concurrency=10`
- [ ] Добавить `--query.max-samples=100000`
- [ ] Добавить `--query.timeout=15s`
- [ ] Уменьшить retention до 24h

### 4. Очистить ненужные аннотации
- [ ] Убрать `prometheus.io/scrape: "true"` с тестовых подов
- [ ] Оставить только для сервисов, которые должны предоставлять метрики
- [ ] Проверить все поды в namespace cloudbridge

### 5. Тестирование
- [ ] Проверить, что API запросы работают без timeout
- [ ] Убедиться, что AI сервис получает данные
- [ ] Проверить производительность Prometheus
- [ ] Протестировать запросы к метрикам

## 🎯 Ожидаемые результаты

### 1. Производительность
- API запросы работают без timeout
- Количество метрик уменьшено до ~50-100
- Частота сбора снижена до 30s
- Использование ресурсов оптимизировано

### 2. Функциональность
- AI сервис получает все необходимые данные
- Prometheus стабильно работает
- Мониторинг работает корректно
- Данные передаются в AI для анализа

### 3. Мониторинг
- Grafana отображает метрики
- AI сервис создает собственные метрики
- Cadence workflow отслеживается
- Системные метрики доступны

## 📝 Критерии приемки

1. **API работает**: `curl http://prometheus:9090/api/v1/query?query=up` возвращает результат за <5 секунд
2. **Метрики собираются**: AI сервис получает данные от Prometheus
3. **Ресурсы оптимизированы**: CPU <80%, Memory <70%
4. **Количество метрик**: <100 активных метрик
5. **Частота сбора**: 30s для всех целей

## 🚀 План внедрения

1. **Подготовка** (1 час)
   - Создать backup текущей конфигурации
   - Подготовить новую конфигурацию
   - Протестировать в dev среде

2. **Внедрение** (30 минут)
   - Обновить ConfigMap
   - Обновить ресурсы
   - Перезапустить Prometheus

3. **Тестирование** (30 минут)
   - Проверить API
   - Проверить сбор метрик
   - Проверить работу AI сервиса

4. **Мониторинг** (1 час)
   - Отслеживать производительность
   - Проверить логи
   - Убедиться в стабильности

## 📞 Контакты

- **Ответственный**: Разработчик DevOps
- **Куратор**: AI Team Lead
- **Срок выполнения**: 1 рабочий день
- **Приоритет**: Высокий

---

*Документ создан: 2025-09-22*  
*Версия: 1.0*  
*Статус: Готов к выполнению*
