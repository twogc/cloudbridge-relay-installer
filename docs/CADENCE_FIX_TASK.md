# Задание: Исправление Cadence Workflow Engine

## Проблема
Cadence не запускается в Kubernetes из-за ошибок конфигурации. Текущий статус:
- Pod в состоянии `CrashLoopBackOff`
- Ошибка: `Must provide one of VisibilityStore and AdvancedVisibilityStore`
- Cadence пытается подключиться к Cassandra, но конфигурация указывает на SQLite

## Текущее состояние

### Файлы конфигурации:
- **Основная конфигурация**: `k8s/cadence/cadence-simple.yaml`
- **Альтернативные конфигурации**: 
  - `k8s/cadence/cadence-fixed.yaml`
  - `k8s/cadence/cadence-working.yaml`
  - `k8s/cadence/cadence-minimal.yaml`
  - `k8s/cadence/cadence-final-working.yaml`

### Интеграция в коде:
- **Основной сервер**: `cmd/relay/main.go` (строки 474-494)
- **Cadence client**: `internal/cadence/client.go`
- **P2P интеграция**: `internal/p2p/cadence_integration.go`
- **Workflows**: `internal/cadence/workflows/`

### Конфигурация в relay сервере:
- **Конфиг**: `configs/config.json` (строки 463-468)
- **Структура**: `config/config.go` (строки 280-286, 715-721)

## Задачи для исправления

### 1. Исправить конфигурацию Cadence
**Файл**: `k8s/cadence/cadence-simple.yaml`

**Проблема**: Неправильная конфигурация visibility store
**Решение**: 
- Использовать правильную конфигурацию для SQLite
- Добавить все необходимые переменные окружения
- Исправить настройки ringpop

### 2. Проверить интеграцию в коде
**Файлы для проверки**:
- `cmd/relay/main.go` - инициализация Cadence client
- `internal/cadence/client.go` - Cadence client
- `internal/p2p/cadence_integration.go` - P2P интеграция

### 3. Протестировать развертывание
**Команды для проверки**:
```bash
# Проверить статус
kubectl get pods -n cloudbridge | grep cadence

# Проверить логи
kubectl logs -n cloudbridge <cadence-pod-name> --tail=50

# Проверить сервис
kubectl get svc -n cloudbridge | grep cadence

# Проверить подключение
kubectl port-forward -n cloudbridge svc/cadence 7934:7934
curl http://localhost:7934/health
```

## Рекомендуемые шаги

### Шаг 1: Создать рабочую конфигурацию
1. Изучить документацию Cadence для SQLite
2. Создать правильную конфигурацию в `k8s/cadence/cadence-working.yaml`
3. Убедиться, что все переменные окружения настроены правильно

### Шаг 2: Проверить интеграцию
1. Убедиться, что Cadence client правильно инициализируется
2. Проверить, что P2P интеграция работает
3. Протестировать workflows

### Шаг 3: Развернуть и протестировать
1. Применить конфигурацию
2. Проверить, что pod запускается
3. Протестировать API endpoints

## Полезные ресурсы

### Документация Cadence:
- [Cadence Configuration](https://cadenceworkflow.io/docs/operation-guide/setup/#configuration)
- [SQLite Configuration](https://cadenceworkflow.io/docs/operation-guide/setup/#sqlite)
- [Docker Configuration](https://cadenceworkflow.io/docs/operation-guide/setup/#docker)

### Примеры конфигурации:
- `docker/compose/docker-compose-cadence.yml` - Docker Compose пример
- `k8s/cadence/cadence-final.yaml` - текущая конфигурация

## Ожидаемый результат

После исправления:
1. Cadence pod должен быть в состоянии `Running`
2. Health check должен проходить успешно
3. API endpoints должны отвечать
4. Интеграция с relay сервером должна работать

## Проверка работы

### 1. Проверить статус pod:
```bash
kubectl get pods -n cloudbridge | grep cadence
# Ожидаемый результат: cadence-xxx-xxx   1/1     Running   0          1m
```

### 2. Проверить логи:
```bash
kubectl logs -n cloudbridge <cadence-pod-name> --tail=20
# Ожидаемый результат: без ошибок, сервер запущен
```

### 3. Проверить health endpoint:
```bash
kubectl port-forward -n cloudbridge svc/cadence 7934:7934
curl http://localhost:7934/health
# Ожидаемый результат: {"status":"ok"}
```

### 4. Проверить интеграцию с relay сервером:
```bash
# Проверить, что relay сервер может подключиться к Cadence
kubectl logs -n cloudbridge <relay-pod-name> | grep -i cadence
# Ожидаемый результат: "Cadence Workflow Engine initialized successfully"
```

## Дополнительные файлы для изучения

- `internal/cadence/workflows/main.go` - основные workflows
- `internal/cadence/activities/` - activities для workflows
- `k8s/cloudbridge-relay-deployment-fixed.yaml` - развертывание relay сервера с Cadence
- `configs/config.json` - конфигурация relay сервера

## Приоритет: ВЫСОКИЙ

Cadence критически важен для работы P2P workflows и AI-оптимизированной маршрутизации.
