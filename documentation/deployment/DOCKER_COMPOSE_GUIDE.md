# 🐳 Docker Compose Guide для CloudBridge Relay

## 📋 Обзор

Этот документ описывает использование Docker Compose для развертывания CloudBridge Relay Server с полным стеком сервисов, включая Cadence Workflow Engine, Cassandra, Prometheus и Grafana.

## 🏗️ Архитектура сервисов

### Схема развертывания:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CloudBridge   │    │   AI Service    │    │  Cadence Service│
│     Relay       │    │                 │    │                 │
│   (Port 9090)   │    │   (Port 8081)   │    │   (Port 8082)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
         │  QUIC Service   │    │ WireGuard Service│   │  Cadence Engine │
         │   (Port 9090)   │    │  (Port 51820)   │    │   (Port 7933)   │
         └─────────────────┘    └─────────────────┘    └─────────────────┘
                                 │
         ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
         │   Cassandra     │    │   Prometheus    │    │    Grafana      │
         │   (Port 9042)   │    │   (Port 9092)   │    │   (Port 3000)   │
         └─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📁 Файлы конфигурации

### Основные файлы:

- `docker-compose.enhanced.yml` - полная конфигурация для production
- `docker-compose.yml` - базовая конфигурация
- `docker-compose.production.yml` - production-ready конфигурация

## 🚀 Быстрый старт

### 1. **Запуск всех сервисов**

```bash
# Запуск в фоновом режиме
docker-compose -f docker-compose.enhanced.yml up -d

# Просмотр статуса
docker-compose -f docker-compose.enhanced.yml ps

# Просмотр логов
docker-compose -f docker-compose.enhanced.yml logs -f
```

### 2. **Запуск отдельных сервисов**

```bash
# Запуск только relay сервиса
docker-compose -f docker-compose.enhanced.yml up -d cloudbridge-relay

# Запуск с зависимостями
docker-compose -f docker-compose.enhanced.yml up -d cloudbridge-relay cassandra cadence
```

### 3. **Остановка сервисов**

```bash
# Остановка всех сервисов
docker-compose -f docker-compose.enhanced.yml down

# Остановка с удалением volumes
docker-compose -f docker-compose.enhanced.yml down -v

# Остановка с удалением образов
docker-compose -f docker-compose.enhanced.yml down --rmi all
```

## 🔧 Конфигурация сервисов

### 1. **CloudBridge Relay**

```yaml
cloudbridge-relay:
  build:
    context: .
    dockerfile: Dockerfile.relay.enhanced
    args:
      VERSION: ${VERSION:-latest}
      BUILD_TIME: ${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}
      GIT_COMMIT: ${GIT_COMMIT:-$(git rev-parse HEAD)}
  image: cloudbridge-relay:${VERSION:-latest}
  container_name: cloudbridge-relay
  restart: unless-stopped
  ports:
    - "9090:9090"    # CloudBridge Relay API
    - "9090:9090/udp" # QUIC Protocol
    - "51820:51820/udp" # WireGuard P2P
    - "8084:8084"    # Metrics
  environment:
    - CONFIG_PATH=/app/configs/config.json
    - LOG_LEVEL=${LOG_LEVEL:-info}
    - WIREGUARD_ENABLED=true
    - GIN_MODE=release
    - TZ=UTC
  volumes:
    - ./configs:/app/configs:ro
    - ./logs:/app/logs
    - /etc/wireguard:/etc/wireguard:ro
  networks:
    - cloudbridge-network
  depends_on:
    - cadence
    - cassandra
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:9090/health"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 40s
  deploy:
    resources:
      limits:
        cpus: '1.0'
        memory: 1G
      reservations:
        cpus: '0.5'
        memory: 512M
```

**Ключевые особенности:**
- Multi-protocol поддержка (TCP/UDP)
- Health checks для мониторинга
- Resource limits для стабильности
- Read-only конфигурация
- Зависимости от Cadence и Cassandra

### 2. **AI Service**

```yaml
cloudbridge-ai:
  build:
    context: .
    dockerfile: Dockerfile.ai.enhanced
  ports:
    - "8081:8081"    # AI Service API
    - "8084:8084"    # Metrics
  environment:
    - CONFIG_PATH=/app/configs/ai.yaml
    - LOG_LEVEL=${LOG_LEVEL:-info}
    - GIN_MODE=release
    - TZ=UTC
  volumes:
    - ./configs:/app/configs:ro
    - ./logs:/app/logs
    - ./models:/app/models
  networks:
    - cloudbridge-network
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8081/health"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 30s
  deploy:
    resources:
      limits:
        cpus: '0.5'
        memory: 512M
      reservations:
        cpus: '0.25'
        memory: 256M
```

**Ключевые особенности:**
- Python ML библиотеки
- Отдельный volume для моделей
- Оптимизированные ресурсы для AI workloads

### 3. **Cadence Workflow Engine**

```yaml
cadence:
  image: ubercadence/server:latest
  container_name: cadence-server
  restart: unless-stopped
  ports:
    - "7933:7933"    # Cadence gRPC
    - "7934:7934"    # Cadence Web
    - "7935:7935"    # Cadence Admin
  environment:
    - CASSANDRA_SEEDS=cassandra
    - DB=cassandra
    - DB_PORT=9042
    - KEYSPACE=cadence
    - VISIBILITY_KEYSPACE=cadence_visibility
    - RINGPOP_SEEDS=cadence:7933
    - LOG_LEVEL=info
    - NUM_HISTORY_SHARDS=4
    - SERVICES=history,matching,frontend,worker
  networks:
    - cloudbridge-network
  depends_on:
    - cassandra
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:7934/health"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 60s
  deploy:
    resources:
      limits:
        cpus: '1.0'
        memory: 1G
      reservations:
        cpus: '0.5'
        memory: 512M
```

**Ключевые особенности:**
- Официальный Cadence образ
- Интеграция с Cassandra
- Множественные сервисы (history, matching, frontend, worker)
- Длительный start-period для инициализации

### 4. **Cassandra Database**

```yaml
cassandra:
  image: cassandra:3.11
  container_name: cassandra
  restart: unless-stopped
  ports:
    - "9042:9042"    # Cassandra CQL
  environment:
    - CASSANDRA_CLUSTER_NAME=cloudbridge-cluster
    - CASSANDRA_DC=datacenter1
    - CASSANDRA_RACK=rack1
    - CASSANDRA_ENDPOINT_SNITCH=GossipingPropertyFileSnitch
    - HEAP_NEWSIZE=128M
    - MAX_HEAP_SIZE=512M
  volumes:
    - cassandra_data:/var/lib/cassandra
    - ./scripts/cassandra-init.cql:/docker-entrypoint-initdb.d/init.cql:ro
  networks:
    - cloudbridge-network
  healthcheck:
    test: ["CMD", "cqlsh", "-e", "describe keyspaces"]
    interval: 30s
    timeout: 10s
    retries: 5
    start_period: 60s
  deploy:
    resources:
      limits:
        cpus: '1.0'
        memory: 1G
      reservations:
        cpus: '0.5'
        memory: 512M
```

**Ключевые особенности:**
- Persistent volume для данных
- Инициализация схемы через CQL скрипты
- Оптимизированные heap настройки
- Длительный start-period для инициализации

### 5. **Prometheus Metrics**

```yaml
prometheus:
  image: prom/prometheus:latest
  container_name: prometheus
  restart: unless-stopped
  ports:
    - "9092:9090"    # Prometheus (changed port to avoid conflict)
  volumes:
    - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    - prometheus_data:/prometheus
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.path=/prometheus'
    - '--web.console.libraries=/etc/prometheus/console_libraries'
    - '--web.console.templates=/etc/prometheus/consoles'
    - '--storage.tsdb.retention.time=200h'
    - '--web.enable-lifecycle'
  networks:
    - cloudbridge-network
  healthcheck:
    test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:9090/-/healthy"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 30s
  deploy:
    resources:
      limits:
        cpus: '0.5'
        memory: 512M
      reservations:
        cpus: '0.25'
        memory: 256M
```

**Ключевые особенности:**
- Persistent volume для метрик
- Retention policy (200h)
- Lifecycle API для перезагрузки конфигурации
- Измененный порт для избежания конфликтов

### 6. **Grafana Dashboard**

```yaml
grafana:
  image: grafana/grafana:latest
  container_name: grafana
  restart: unless-stopped
  ports:
    - "3000:3000"    # Grafana Dashboard
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}
    - GF_USERS_ALLOW_SIGN_UP=false
    - GF_INSTALL_PLUGINS=grafana-piechart-panel
  volumes:
    - grafana_data:/var/lib/grafana
    - ./grafana/provisioning:/etc/grafana/provisioning:ro
    - ./grafana/dashboards:/var/lib/grafana/dashboards:ro
  networks:
    - cloudbridge-network
  depends_on:
    - prometheus
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:3000/api/health"]
    interval: 30s
    timeout: 10s
    retries: 3
    start_period: 30s
  deploy:
    resources:
      limits:
        cpus: '0.25'
        memory: 256M
      reservations:
        cpus: '0.1'
        memory: 128M
```

**Ключевые особенности:**
- Provisioning дашбордов и datasources
- Безопасный пароль администратора
- Отключена регистрация пользователей
- Зависимость от Prometheus

## 🌐 Networking

### Конфигурация сети:

```yaml
networks:
  cloudbridge-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

**Особенности:**
- Изолированная сеть для всех сервисов
- Настроенная подсеть для предсказуемых IP адресов
- Bridge driver для локального развертывания

## 💾 Volumes

### Persistent volumes:

```yaml
volumes:
  cassandra_data:
    driver: local
  prometheus_data:
    driver: local
  grafana_data:
    driver: local
```

**Особенности:**
- Локальные volumes для данных
- Persistent storage для критических данных
- Автоматическое управление Docker

## 🔧 Environment Variables

### Основные переменные:

```bash
# Версия и сборка
VERSION=latest
BUILD_TIME=2025-09-06T10:00:00Z
GIT_COMMIT=abc123

# Логирование
LOG_LEVEL=info

# Grafana
GRAFANA_PASSWORD=admin123

# WireGuard
WITH_WIREGUARD_GO=false
```

### Создание .env файла:

```bash
# Создание .env файла
cat > .env << EOF
VERSION=latest
LOG_LEVEL=info
GRAFANA_PASSWORD=admin123
WITH_WIREGUARD_GO=false
EOF
```

## 🚀 Развертывание

### 1. **Development Environment**

```bash
# Клонирование репозитория
git clone <repository-url>
cd cloudbridge-relay-installer

# Создание .env файла
cp env.example .env

# Запуск сервисов
docker-compose -f docker-compose.enhanced.yml up -d

# Проверка статуса
docker-compose -f docker-compose.enhanced.yml ps
```

### 2. **Production Environment**

```bash
# Сборка образов
./build-docker.sh --service all

# Запуск в production режиме
docker-compose -f docker-compose.enhanced.yml up -d

# Проверка health status
docker-compose -f docker-compose.enhanced.yml ps
```

### 3. **Scaling**

```bash
# Масштабирование relay сервиса
docker-compose -f docker-compose.enhanced.yml up -d --scale cloudbridge-relay=3

# Масштабирование AI сервиса
docker-compose -f docker-compose.enhanced.yml up -d --scale cloudbridge-ai=2
```

## 📊 Мониторинг

### 1. **Health Checks**

```bash
# Проверка health status всех сервисов
docker-compose -f docker-compose.enhanced.yml ps

# Детальная информация о health checks
docker inspect cloudbridge-relay | jq '.[0].State.Health'
```

### 2. **Логи**

```bash
# Просмотр логов всех сервисов
docker-compose -f docker-compose.enhanced.yml logs

# Просмотр логов конкретного сервиса
docker-compose -f docker-compose.enhanced.yml logs cloudbridge-relay

# Следить за логами в реальном времени
docker-compose -f docker-compose.enhanced.yml logs -f cloudbridge-relay
```

### 3. **Метрики**

```bash
# Проверка метрик relay сервиса
curl http://localhost:8084/metrics

# Проверка Prometheus targets
curl http://localhost:9092/api/v1/targets

# Доступ к Grafana
open http://localhost:3000
```

## 🔧 Troubleshooting

### 1. **Common Issues**

#### Проблема: Сервис не запускается

```bash
# Проверка логов
docker-compose -f docker-compose.enhanced.yml logs cloudbridge-relay

# Проверка конфигурации
docker-compose -f docker-compose.enhanced.yml config

# Перезапуск сервиса
docker-compose -f docker-compose.enhanced.yml restart cloudbridge-relay
```

#### Проблема: Порт уже используется

```bash
# Проверка занятых портов
netstat -tuln | grep :9090

# Изменение портов в docker-compose.yml
ports:
  - "9091:9090"  # Изменить внешний порт
```

#### Проблема: Недостаточно памяти

```bash
# Проверка использования ресурсов
docker stats

# Уменьшение resource limits
deploy:
  resources:
    limits:
      memory: 512M
```

### 2. **Debugging Commands**

```bash
# Вход в контейнер
docker-compose -f docker-compose.enhanced.yml exec cloudbridge-relay bash

# Проверка сетевых соединений
docker-compose -f docker-compose.enhanced.yml exec cloudbridge-relay netstat -tuln

# Проверка процессов
docker-compose -f docker-compose.enhanced.yml exec cloudbridge-relay ps aux
```

## 📚 Полезные команды

### Управление сервисами

```bash
# Запуск сервисов
docker-compose -f docker-compose.enhanced.yml up -d

# Остановка сервисов
docker-compose -f docker-compose.enhanced.yml down

# Перезапуск сервисов
docker-compose -f docker-compose.enhanced.yml restart

# Пересборка и запуск
docker-compose -f docker-compose.enhanced.yml up -d --build
```

### Мониторинг

```bash
# Статус сервисов
docker-compose -f docker-compose.enhanced.yml ps

# Использование ресурсов
docker-compose -f docker-compose.enhanced.yml top

# Логи сервисов
docker-compose -f docker-compose.enhanced.yml logs -f
```

### Управление данными

```bash
# Создание backup
docker-compose -f docker-compose.enhanced.yml exec cassandra nodetool snapshot

# Очистка volumes
docker-compose -f docker-compose.enhanced.yml down -v

# Просмотр volumes
docker volume ls
```

## 🎯 Заключение

Docker Compose обеспечивает:

- ✅ **Простое развертывание** - одна команда для запуска всех сервисов
- ✅ **Управление зависимостями** - автоматический порядок запуска
- ✅ **Масштабирование** - легкое масштабирование сервисов
- ✅ **Мониторинг** - health checks и логирование
- ✅ **Изоляция** - отдельная сеть для сервисов
- ✅ **Персистентность** - сохранение данных между перезапусками

---

*Документ создан: 2025-09-06*  
*Статус: ✅ Готов к использованию* 🚀
