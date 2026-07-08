# 🚀 Cadence Setup Guide

## 📋 Обзор

Cadence - это workflow engine для Kubernetes. Данный гайд описывает правильную настройку Cadence с PostgreSQL в Kubernetes.

## 🔧 Правильная конфигурация

### 1. Переменные окружения (рекомендуемый подход)

```yaml
env:
- name: CADENCE_USE_ENV_VARS
  value: "true"
- name: DB
  value: "postgres"
- name: DB_HOST
  value: "postgres.cloudbridge-system.svc"
- name: DB_PORT
  value: "5432"
- name: DB_NAME
  value: "cadence"
- name: DB_USER
  value: "ipam_user"
- name: DB_PASSWORD
  value: "secure_password_123"
- name: VISIBILITY_STORE
  value: "postgres"
- name: VISIBILITY_DB_HOST
  value: "postgres.cloudbridge-system.svc"
- name: VISIBILITY_DB_PORT
  value: "5432"
- name: VISIBILITY_DB_NAME
  value: "cadence_visibility"
- name: VISIBILITY_DB_USER
  value: "ipam_user"
- name: VISIBILITY_DB_PASSWORD
  value: "secure_password_123"
- name: CLUSTER_METADATA_ENABLE_GLOBAL_DOMAIN
  value: "false"
- name: CLUSTER_METADATA_FAILOVER_VERSION_INCREMENT
  value: "10"
- name: CLUSTER_METADATA_MASTER_CLUSTER_NAME
  value: "active"
- name: CLUSTER_METADATA_CURRENT_CLUSTER_NAME
  value: "active"
```

### 2. Команда запуска

```yaml
command: ["/bin/bash", "-lc"]
args:
  - |
    echo "Starting Cadence (environment variables mode)"
    exec /usr/local/bin/cadence-server start --services=frontend,history,matching,worker
```

## 🗄️ Подготовка базы данных

### 1. Создание баз данных

```bash
# Подключение к PostgreSQL
kubectl exec -n cloudbridge-system postgres-587dfcc9cd-cvwsk -- psql -U ipam_user -d cloudbridge_ipam

# Создание баз данных
CREATE DATABASE cadence;
CREATE DATABASE cadence_visibility;
```

### 2. Проверка подключения

```bash
# Тест подключения
kubectl exec -n cloudbridge-system postgres-587dfcc9cd-cvwsk -- psql -U ipam_user -d cadence -c "SELECT 1;"
```

## 🚨 Частые проблемы и решения

### Проблема 1: "no config files found"
**Причина:** Cadence ищет `base.yaml` в режиме директории
**Решение:** Использовать переменные окружения (`CADENCE_USE_ENV_VARS=true`)

### Проблема 2: "password authentication failed"
**Причина:** Неправильные учетные данные или DNS
**Решение:** 
- Проверить пользователя и пароль
- Использовать IP адрес вместо DNS имени
- Убедиться, что базы данных созданы

### Проблема 3: "not supported plugin sqlite3"
**Причина:** SQLite не поддерживается в новых версиях Cadence
**Решение:** Использовать PostgreSQL

### Проблема 4: "PublicClient.HostPort: zero value"
**Причина:** Отсутствует конфигурация publicClient
**Решение:** Добавить в конфигурацию:
```yaml
publicClient:
  hostPort: "cadence:7933"
```

## 📊 Мониторинг

### Проверка статуса

```bash
# Статус подов
kubectl get pods -n cloudbridge | grep cadence

# Логи
kubectl logs -n cloudbridge deployment/cadence

# Проверка подключения
kubectl port-forward -n cloudbridge svc/cadence 7933:7933
```

### Проверка базы данных

```bash
# Подключение к базе
kubectl exec -n cloudbridge-system postgres-587dfcc9cd-cvwsk -- psql -U ipam_user -d cadence -c "\dt"

# Проверка таблиц
kubectl exec -n cloudbridge-system postgres-587dfcc9cd-cvwsk -- psql -U ipam_user -d cadence_visibility -c "\dt"
```

## 🔧 Настройка для production

### 1. Использование секретов

```yaml
env:
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: cadence-db
      key: password
- name: VISIBILITY_DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: cadence-db
      key: password
```

### 2. Настройка ресурсов

```yaml
resources:
  requests:
    cpu: "250m"
    memory: "512Mi"
  limits:
    cpu: "500m"
    memory: "1Gi"
```

### 3. Настройка проб

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 7934
    scheme: HTTP
  initialDelaySeconds: 120
  periodSeconds: 30
  timeoutSeconds: 10
  failureThreshold: 5
readinessProbe:
  httpGet:
    path: /health
    port: 7934
    scheme: HTTP
  initialDelaySeconds: 90
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 5
```

## 📝 Чек-лист

- [ ] PostgreSQL запущен и доступен
- [ ] Базы данных `cadence` и `cadence_visibility` созданы
- [ ] Пользователь `ipam_user` имеет доступ к базам
- [ ] Переменные окружения настроены правильно
- [ ] Cadence запускается без ошибок
- [ ] Health checks проходят успешно
- [ ] Подключение к базе данных работает

## 🚀 Быстрый старт

```bash
# 1. Создать базы данных
kubectl exec -n cloudbridge-system postgres-587dfcc9cd-cvwsk -- psql -U ipam_user -d cloudbridge_ipam -c "CREATE DATABASE cadence; CREATE DATABASE cadence_visibility;"

# 2. Применить конфигурацию
kubectl apply -f k8s/cadence/cadence.yaml

# 3. Проверить статус
kubectl get pods -n cloudbridge | grep cadence

# 4. Проверить логи
kubectl logs -n cloudbridge deployment/cadence
```

## 🔍 Диагностика

### Полезные команды

```bash
# Проверка переменных окружения
kubectl exec -n cloudbridge deployment/cadence -- env | grep -E "(DB|VISIBILITY|CADENCE)"

# Проверка подключения к базе
kubectl exec -n cloudbridge deployment/cadence -- nc -zv postgres.cloudbridge-system.svc 5432

# Проверка конфигурации
kubectl exec -n cloudbridge deployment/cadence -- cat /etc/cadence/config/docker.yaml
```

## 📚 Дополнительные ресурсы

- [Cadence Documentation](https://cadenceworkflow.io/docs/)
- [Cadence Configuration](https://cadenceworkflow.io/docs/operation-guide/setup/)
- [PostgreSQL Setup](https://www.postgresql.org/docs/)


