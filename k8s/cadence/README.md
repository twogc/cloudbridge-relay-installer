# Cadence Configuration

## Описание

Этот каталог содержит конфигурацию для Cadence workflow engine в Kubernetes.

## Быстрый старт

```bash
# 1. Создать базы данных
kubectl exec -n cloudbridge-system postgres-587dfcc9cd-cvwsk -- psql -U ipam_user -d cloudbridge_ipam -c "CREATE DATABASE cadence; CREATE DATABASE cadence_visibility;"

# 2. Применить конфигурацию
kubectl apply -f k8s/cadence/cadence.yaml

# 3. Проверить статус
kubectl get pods -n cloudbridge | grep cadence
```

## Конфигурация

### Переменные окружения

- `CADENCE_USE_ENV_VARS=true` - Использовать переменные окружения
- `DB=postgres` - Тип базы данных
- `DB_HOST=postgres.cloudbridge-system.svc` - Хост PostgreSQL
- `DB_NAME=cadence` - Имя основной базы данных
- `VISIBILITY_STORE=postgres` - Тип хранилища видимости
- `VISIBILITY_DB_NAME=cadence_visibility` - Имя базы данных видимости

### База данных

- **Основная БД**: `cadence`
- **БД видимости**: `cadence_visibility`
- **Пользователь**: `ipam_user`
- **Пароль**: `secure_password_123`

## Мониторинг

```bash
# Статус подов
kubectl get pods -n cloudbridge | grep cadence

# Логи
kubectl logs -n cloudbridge deployment/cadence

# Проверка подключения
kubectl port-forward -n cloudbridge svc/cadence 7933:7933
```

## Диагностика

### Полезные команды

```bash
# Проверка переменных окружения
kubectl exec -n cloudbridge deployment/cadence -- env | grep -E "(DB|VISIBILITY|CADENCE)"

# Проверка подключения к базе
kubectl exec -n cloudbridge deployment/cadence -- nc -zv postgres.cloudbridge-system.svc 5432

# Проверка конфигурации
kubectl exec -n cloudbridge deployment/cadence -- cat /etc/cadence/config/docker.yaml
```

## Частые проблемы

### 1. "password authentication failed"
**Решение**: Проверить пользователя и пароль в переменных окружения

### 2. "no config files found"
**Решение**: Убедиться, что `CADENCE_USE_ENV_VARS=true`

### 3. "not supported plugin sqlite3"
**Решение**: Использовать PostgreSQL вместо SQLite

## Дополнительные ресурсы

- [Полная документация](../CADENCE_SETUP_GUIDE.md)
- [Скрипт автоматической настройки](../setup-cadence.sh)


