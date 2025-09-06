# 🐳 CloudBridge Relay - Docker Containerization

## 📋 Обзор

Этот документ описывает Docker containerization для CloudBridge Relay Server с интеграцией Cadence Workflow Engine. Все сервисы containerized с использованием multi-stage builds для оптимизации размера и безопасности.

## 🏗️ Архитектура Docker

### Компоненты системы:

1. **CloudBridge Relay** - основной сервис (порт 9090)
2. **AI Service** - сервис искусственного интеллекта (порт 8081)
3. **Cadence Service** - workflow orchestration (порт 8082)
4. **QUIC Service** - высокопроизводительный протокол (порт 9090)
5. **WireGuard Service** - P2P mesh networking (порт 51820)
6. **Cadence Workflow Engine** - движок workflow (порт 7933)
7. **Cassandra Database** - база данных (порт 9042)
8. **Prometheus** - метрики (порт 9092)
9. **Grafana** - дашборды (порт 3000)

## 📁 Структура Docker файлов

```
cloudbridge-relay-installer/
├── Dockerfile.relay.enhanced      # Основной relay сервис
├── Dockerfile.ai.enhanced         # AI сервис
├── Dockerfile.cadence.enhanced    # Cadence сервис
├── Dockerfile.quic.enhanced       # QUIC сервис
├── Dockerfile.wireguard.enhanced  # WireGuard сервис
├── Dockerfile.relay.simple        # Упрощенная версия
├── docker-compose.enhanced.yml    # Полная конфигурация
├── build-docker.sh                # Скрипт сборки
└── .dockerignore                  # Исключения для сборки
```

## 🔧 Dockerfile'ы

### 1. **Dockerfile.relay.enhanced**

Основной сервис CloudBridge Relay с multi-stage build:

```dockerfile
# Stage 1: Go Builder
FROM golang:1.25-alpine AS builder
# ... build dependencies and compilation

# Stage 2: Production Image
FROM alpine:3.18
# ... runtime dependencies and configuration
```

**Особенности:**
- Multi-stage build для минимизации размера
- Non-root пользователь для безопасности
- Health checks для мониторинга
- WireGuard tools для P2P networking
- Оптимизированные labels и metadata

### 2. **Dockerfile.ai.enhanced**

AI сервис с Python ML библиотеками:

```dockerfile
# Stage 1: Go Builder
FROM golang:1.25-alpine AS builder
# ... Go compilation

# Stage 2: Production Image
FROM alpine:3.18
# ... Python ML libraries installation
```

**Особенности:**
- Python 3 и ML библиотеки (numpy, scikit-learn, pandas)
- Оптимизированный для AI workloads
- Отдельные volumes для моделей

### 3. **Dockerfile.cadence.enhanced**

Cadence workflow сервис:

```dockerfile
# Stage 1: Go Builder
FROM golang:1.25-alpine AS builder
# ... Go compilation

# Stage 2: Production Image
FROM alpine:3.18
# ... Cadence-specific configuration
```

**Особенности:**
- Специализирован для workflow orchestration
- Интеграция с Cadence Workflow Engine
- Отдельные volumes для workflows

### 4. **Dockerfile.quic.enhanced**

QUIC протокол сервис:

```dockerfile
# Stage 1: Go Builder
FROM golang:1.25-alpine AS builder
# ... Go compilation

# Stage 2: Production Image
FROM alpine:3.18
# ... QUIC-specific configuration
```

**Особенности:**
- Оптимизирован для QUIC протокола
- UDP и TCP порты
- Высокая производительность

### 5. **Dockerfile.wireguard.enhanced**

WireGuard P2P сервис:

```dockerfile
# Stage 1: Go Builder
FROM golang:1.25-alpine AS builder
# ... Go compilation

# Stage 2: WireGuard Go Builder
FROM golang:1.25-alpine AS wireguard-builder
# ... WireGuard Go compilation

# Stage 3: Production Image
FROM alpine:3.18
# ... WireGuard tools and configuration
```

**Особенности:**
- WireGuard Go userspace binary
- CAP_NET_ADMIN capabilities
- /dev/net/tun device
- P2P mesh networking

## 🚀 Docker Compose

### **docker-compose.enhanced.yml**

Полная конфигурация для всех сервисов:

```yaml
version: '3.8'

services:
  cloudbridge-relay:
    build:
      context: .
      dockerfile: Dockerfile.relay.enhanced
      args:
        VERSION: ${VERSION:-latest}
        BUILD_TIME: ${BUILD_TIME:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}
        GIT_COMMIT: ${GIT_COMMIT:-$(git rev-parse HEAD)}
    ports:
      - "9090:9090"    # CloudBridge Relay API
      - "9090:9090/udp" # QUIC Protocol
      - "51820:51820/udp" # WireGuard P2P
      - "8084:8084"    # Metrics
    environment:
      - CONFIG_PATH=/app/configs/config.json
      - LOG_LEVEL=${LOG_LEVEL:-info}
      - WIREGUARD_ENABLED=true
    volumes:
      - ./configs:/app/configs:ro
      - ./logs:/app/logs
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
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

**Особенности:**
- Resource limits и reservations
- Health checks для всех сервисов
- Proper networking и volumes
- Environment variables
- Dependencies между сервисами

## 🛠️ Build Scripts

### **build-docker.sh**

Автоматический скрипт сборки всех образов:

```bash
# Сборка всех сервисов
./build-docker.sh --service all

# Сборка конкретного сервиса
./build-docker.sh --service relay

# Сборка с push в registry
./build-docker.sh --service relay --push

# Сборка с очисткой
./build-docker.sh --service all --clean
```

**Опции:**
- `--service SERVICE` - сборка конкретного сервиса
- `--push` - push в registry после сборки
- `--clean` - очистка старых образов
- `--registry REGISTRY` - кастомный registry
- `--version VERSION` - кастомная версия
- `--with-wireguard-go` - включить WireGuard Go

### **.dockerignore**

Оптимизация контекста сборки:

```
# Git
.git
.gitignore

# Documentation
*.md
documentation/

# Build artifacts
build/
bin/
*.tar.gz

# Logs
*.log
logs/

# IDE files
.vscode/
.idea/

# Keep only necessary files
!cmd/
!internal/
!pkg/
!configs/
!go.mod
!go.sum
```

## 📊 Оптимизации

### 1. **Multi-stage Builds**

```dockerfile
# Stage 1: Builder
FROM golang:1.25-alpine AS builder
# ... build dependencies and compilation

# Stage 2: Production
FROM alpine:3.18
# ... only runtime dependencies
```

**Преимущества:**
- Минимальный размер образа
- Только необходимые runtime зависимости
- Безопасность (нет build tools в production)

### 2. **Layer Caching**

```dockerfile
# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .
```

**Преимущества:**
- Быстрая пересборка при изменениях кода
- Кэширование зависимостей
- Оптимизация времени сборки

### 3. **Security**

```dockerfile
# Create non-root user
RUN addgroup -g 1000 cloudbridge && \
    adduser -D -s /bin/bash -u 1000 -G cloudbridge cloudbridge

# Switch to non-root user
USER cloudbridge
```

**Преимущества:**
- Запуск от non-root пользователя
- Минимальные привилегии
- Безопасность контейнера

## 🔍 Health Checks

### Конфигурация Health Checks

```dockerfile
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:9090/health || exit 1
```

**Параметры:**
- `interval` - интервал проверки (30s)
- `timeout` - таймаут запроса (10s)
- `start-period` - период запуска (5s)
- `retries` - количество попыток (3)

### Health Check Endpoints

- **Relay Service**: `http://localhost:9090/health`
- **AI Service**: `http://localhost:8081/health`
- **Cadence Service**: `http://localhost:8082/health`
- **QUIC Service**: `http://localhost:9090/health`
- **WireGuard Service**: `http://localhost:8084/health`

## 📈 Мониторинг

### Метрики

Все сервисы экспортируют метрики на порту 8084:

```yaml
ports:
  - "8084:8084"    # Metrics
```

### Prometheus Integration

```yaml
prometheus:
  image: prom/prometheus:latest
  ports:
    - "9092:9090"
  volumes:
    - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
```

### Grafana Dashboards

```yaml
grafana:
  image: grafana/grafana:latest
  ports:
    - "3000:3000"
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-admin}
```

## 🚀 Развертывание

### 1. **Локальная разработка**

```bash
# Сборка образов
./build-docker.sh --service all

# Запуск сервисов
docker-compose -f docker-compose.enhanced.yml up -d

# Проверка статуса
docker-compose ps
```

### 2. **Production развертывание**

```bash
# Сборка с push в registry
./build-docker.sh --service all --push

# Развертывание в production
docker-compose -f docker-compose.enhanced.yml up -d
```

### 3. **Масштабирование**

```bash
# Масштабирование сервиса
docker-compose -f docker-compose.enhanced.yml up -d --scale cloudbridge-relay=3
```

## 🔧 Troubleshooting

### Общие проблемы

#### 1. **Build Failures**

```bash
# Очистка Docker cache
docker system prune -a

# Пересборка без cache
docker build --no-cache -f Dockerfile.relay.enhanced .
```

#### 2. **Permission Issues**

```bash
# Проверка прав пользователя
docker run --rm -v $(pwd):/workspace alpine ls -la /workspace

# Исправление прав
sudo chown -R $USER:$USER .
```

#### 3. **Port Conflicts**

```bash
# Проверка занятых портов
netstat -tuln | grep :9090

# Изменение портов в docker-compose.yml
ports:
  - "9091:9090"  # Изменить внешний порт
```

### Логи и отладка

```bash
# Просмотр логов
docker-compose logs cloudbridge-relay

# Следить за логами в реальном времени
docker-compose logs -f cloudbridge-relay

# Вход в контейнер
docker exec -it cloudbridge-relay bash
```

## 📚 Дополнительные ресурсы

### Полезные команды

```bash
# Список образов
docker images | grep cloudbridge

# Удаление неиспользуемых образов
docker image prune -a

# Информация об образе
docker inspect cloudbridge-relay:latest

# Размер образа
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"
```

### Best Practices

1. **Используйте multi-stage builds** для минимизации размера
2. **Настройте health checks** для мониторинга
3. **Используйте non-root пользователей** для безопасности
4. **Оптимизируйте layer caching** для быстрой сборки
5. **Настройте resource limits** для стабильности
6. **Используйте .dockerignore** для оптимизации контекста

## 🎯 Заключение

Docker containerization CloudBridge Relay Server обеспечивает:

- ✅ **Масштабируемость** - легкое масштабирование сервисов
- ✅ **Портабельность** - работа на любой платформе с Docker
- ✅ **Безопасность** - изоляция и non-root пользователи
- ✅ **Мониторинг** - health checks и метрики
- ✅ **Оптимизация** - минимальный размер образов
- ✅ **Автоматизация** - скрипты сборки и развертывания

---

*Документ создан: 2025-09-06*  
*Статус: ✅ Готов к использованию* 🚀
