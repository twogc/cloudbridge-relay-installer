# Исправление TLS проверок Blackbox Exporter

## Проблема

В логах Blackbox Exporter появляется ошибка:

```
time=2025-06-29T22:36:54.308Z level=ERROR source=tcp.go:131 msg="Error dialing TCP" module=tls_connect target=relay:8080 err="tls: failed to verify certificate: x509: certificate is valid for edge.2gc.ru, not relay"
```

## Причина

Blackbox Exporter пытается проверить TLS соединение с внутренним именем контейнера `relay`, но сертификат выдан для домена `edge.2gc.ru`. В Docker-среде это приводит к ошибке валидации сертификата.

## Решение

### 1. Автоматическое исправление

Запустите скрипт исправления:

```bash
./scripts/fix_blackbox_tls.sh
```

### 2. Ручное исправление

#### Шаг 1: Обновление конфигурации Blackbox

В файле `deploy/blackbox/blackbox.yml` добавлен новый модуль:

```yaml
# Модуль для внутренних TLS проверок в Docker-среде
tls_connect_internal:
  prober: tcp
  timeout: 5s
  tcp:
    tls: true
    tls_config:
      insecure_skip_verify: true
    preferred_ip_protocol: "ip4"
```

#### Шаг 2: Обновление конфигурации Prometheus

В файле `prometheus/prometheus.yml` изменен модуль для проверки Relay:

```yaml
- job_name: 'relay-tls'
  static_configs:
    - targets: ['relay:8080']
  metrics_path: '/probe'
  params:
    module: ['tls_connect_internal']  # Изменено с 'tls_connect'
```

#### Шаг 3: Перезапуск сервисов

```bash
# Остановка сервисов
docker-compose stop blackbox prometheus

# Запуск сервисов
docker-compose up -d blackbox prometheus
```

## Проверка исправления

### 1. Проверка статуса сервисов

```bash
# Проверка Blackbox
curl -f http://localhost:9115/-/healthy

# Проверка Prometheus
curl -f http://localhost:9090/-/healthy
```

### 2. Проверка метрик

```bash
# Проверка метрик Blackbox
curl http://localhost:9115/metrics | grep "probe_success" | grep "module=\"tls_connect_internal\""
```

### 3. Проверка в веб-интерфейсе

1. Откройте Prometheus: http://localhost:9090
2. Перейдите в Status -> Targets
3. Найдите job `relay-tls`
4. Статус должен быть "UP"

## Объяснение изменений

### Модуль `tls_connect_internal`

- **`insecure_skip_verify: true`** - пропускает проверку сертификата
- Подходит для внутренних проверок в Docker-среде
- Проверяет только доступность TLS соединения

### Безопасность

- Изменения применяются только для внутренних проверок
- Внешние проверки по-прежнему используют строгую валидацию сертификатов
- Не влияет на безопасность production окружения

## Альтернативные решения

### 1. Использование правильного домена

```yaml
- job_name: 'relay-tls-external'
  static_configs:
    - targets: ['edge.2gc.ru:8080']  # Реальный домен
  params:
    module: ['tls_connect']  # Строгая проверка
```

### 2. Настройка DNS в Docker

```yaml
# В docker-compose.yml
services:
  blackbox:
    extra_hosts:
      - "relay:172.18.0.3"  # IP адрес контейнера relay
```

### 3. Использование IP адреса

```yaml
- job_name: 'relay-tls'
  static_configs:
    - targets: ['172.18.0.3:8080']  # IP адрес контейнера
  params:
    module: ['tls_connect_internal']
```

## Мониторинг

После исправления в логах Blackbox Exporter должны появиться успешные проверки:

```
time=2025-06-29T22:36:54.304Z level=INFO source=handler.go:122 msg="Beginning probe" module=tls_connect_internal target=relay:8080
time=2025-06-29T22:36:54.308Z level=INFO source=tcp.go:131 msg="TCP connection established" module=tls_connect_internal target=relay:8080
time=2025-06-29T22:36:54.308Z level=INFO source=handler.go:135 msg="Probe succeeded" module=tls_connect_internal target=relay:8080 duration_seconds=0.004312962
```

## Устранение неполадок

### Проблема: Сервисы не запускаются

```bash
# Проверка логов
docker-compose logs blackbox
docker-compose logs prometheus

# Проверка конфигурации
docker exec blackbox cat /etc/blackbox_exporter/blackbox.yml
```

### Проблема: Метрики не появляются

```bash
# Проверка доступности Blackbox
curl http://localhost:9115/metrics

# Проверка конфигурации Prometheus
curl http://localhost:9090/api/v1/targets
```

### Проблема: Все еще есть ошибки

```bash
# Полная перезагрузка
docker-compose down
docker-compose up -d

# Проверка через 30 секунд
sleep 30
curl http://localhost:9115/metrics | grep "probe_success"
``` 