# 🐳 Docker Best Practices для CloudBridge Relay

## 📋 Обзор

Этот документ описывает лучшие практики использования Docker для CloudBridge Relay Server, включая оптимизацию, безопасность, мониторинг и troubleshooting.

## 🏗️ Архитектурные принципы

### 1. **Multi-stage Builds**

```dockerfile
# ✅ ХОРОШО: Multi-stage build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o cloudbridge-relay ./cmd/relay

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/cloudbridge-relay /usr/local/bin/
CMD ["cloudbridge-relay"]
```

```dockerfile
# ❌ ПЛОХО: Single-stage build
FROM golang:1.25-alpine
WORKDIR /app
COPY . .
RUN go build -o cloudbridge-relay ./cmd/relay
CMD ["cloudbridge-relay"]
```

**Преимущества multi-stage:**
- Размер образа: 57MB vs 500MB+
- Безопасность: нет build tools в production
- Производительность: быстрый запуск

### 2. **Layer Optimization**

```dockerfile
# ✅ ХОРОШО: Оптимизированные слои
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Копируем go.mod первым для кэширования
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .
RUN go build -o cloudbridge-relay ./cmd/relay
```

```dockerfile
# ❌ ПЛОХО: Неэффективные слои
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download && go build -o cloudbridge-relay ./cmd/relay
```

**Преимущества:**
- Кэширование зависимостей
- Быстрая пересборка при изменениях кода
- Эффективное использование Docker cache

## 🔒 Безопасность

### 1. **Non-root пользователи**

```dockerfile
# ✅ ХОРОШО: Non-root пользователь
FROM alpine:3.18
RUN addgroup -g 1000 cloudbridge && \
    adduser -D -s /bin/bash -u 1000 -G cloudbridge cloudbridge
COPY --from=builder /app/cloudbridge-relay /usr/local/bin/
RUN chown cloudbridge:cloudbridge /usr/local/bin/cloudbridge-relay
USER cloudbridge
CMD ["cloudbridge-relay"]
```

```dockerfile
# ❌ ПЛОХО: Root пользователь
FROM alpine:3.18
COPY --from=builder /app/cloudbridge-relay /usr/local/bin/
CMD ["cloudbridge-relay"]
```

### 2. **Минимальные привилегии**

```yaml
# ✅ ХОРОШО: Минимальные привилегии
services:
  cloudbridge-relay:
    user: "1000:1000"
    read_only: true
    tmpfs:
      - /tmp
      - /var/run
```

```yaml
# ❌ ПЛОХО: Избыточные привилегии
services:
  cloudbridge-relay:
    privileged: true
    user: "root"
```

### 3. **Secrets Management**

```yaml
# ✅ ХОРОШО: Docker secrets
services:
  cloudbridge-relay:
    secrets:
      - db_password
      - api_key
    environment:
      - DB_PASSWORD_FILE=/run/secrets/db_password

secrets:
  db_password:
    file: ./secrets/db_password.txt
  api_key:
    external: true
```

```yaml
# ❌ ПЛОХО: Пароли в environment
services:
  cloudbridge-relay:
    environment:
      - DB_PASSWORD=secret123
      - API_KEY=abc123
```

## 📊 Мониторинг и Health Checks

### 1. **Health Checks**

```dockerfile
# ✅ ХОРОШО: Comprehensive health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:9090/health || exit 1
```

```dockerfile
# ❌ ПЛОХО: Отсутствие health check
# No health check defined
```

### 2. **Метрики и логирование**

```dockerfile
# ✅ ХОРОШО: Structured logging
ENV LOG_LEVEL=info
ENV LOG_FORMAT=json
CMD ["cloudbridge-relay", "--log-format=json"]
```

```yaml
# ✅ ХОРОШО: Logging configuration
services:
  cloudbridge-relay:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## 🚀 Производительность

### 1. **Resource Limits**

```yaml
# ✅ ХОРОШО: Resource limits
services:
  cloudbridge-relay:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 512M
```

```yaml
# ❌ ПЛОХО: Без resource limits
services:
  cloudbridge-relay:
    # No resource limits
```

### 2. **Volume Optimization**

```yaml
# ✅ ХОРОШО: Оптимизированные volumes
services:
  cloudbridge-relay:
    volumes:
      - ./configs:/app/configs:ro  # Read-only config
      - relay-logs:/app/logs       # Named volume for logs
      - /tmp/cloudbridge           # tmpfs for temporary files

volumes:
  relay-logs:
    driver: local
```

```yaml
# ❌ ПЛОХО: Неэффективные volumes
services:
  cloudbridge-relay:
    volumes:
      - .:/app                     # Bind mount всего проекта
      - /var/log:/app/logs         # System directory
```

## 🔧 Troubleshooting

### 1. **Debugging контейнеров**

```bash
# ✅ ХОРОШО: Debugging команды
# Просмотр логов
docker logs cloudbridge-relay --tail 100 -f

# Вход в контейнер
docker exec -it cloudbridge-relay bash

# Проверка процессов
docker exec cloudbridge-relay ps aux

# Проверка сетевых соединений
docker exec cloudbridge-relay netstat -tuln
```

### 2. **Performance анализ**

```bash
# ✅ ХОРОШО: Performance команды
# Использование ресурсов
docker stats cloudbridge-relay

# Информация об образе
docker inspect cloudbridge-relay:latest

# Размер слоев
docker history cloudbridge-relay:latest
```

### 3. **Common Issues**

#### Проблема: Контейнер не запускается

```bash
# Диагностика
docker logs cloudbridge-relay
docker inspect cloudbridge-relay

# Проверка конфигурации
docker run --rm cloudbridge-relay:latest --help
```

#### Проблема: Высокое использование памяти

```bash
# Анализ памяти
docker exec cloudbridge-relay cat /proc/meminfo
docker exec cloudbridge-relay ps aux --sort=-%mem

# Профилирование
docker exec cloudbridge-relay curl http://localhost:6060/debug/pprof/heap
```

#### Проблема: Медленная сборка

```bash
# Очистка cache
docker builder prune -a

# Анализ времени сборки
docker build --progress=plain -f Dockerfile.relay.enhanced .
```

## 📈 Оптимизация

### 1. **Build Optimization**

```dockerfile
# ✅ ХОРОШО: Оптимизированная сборка
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Кэширование зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Оптимизированная сборка
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o cloudbridge-relay \
    ./cmd/relay
```

### 2. **Runtime Optimization**

```dockerfile
# ✅ ХОРОШО: Runtime оптимизации
FROM alpine:3.18

# Минимальные зависимости
RUN apk add --no-cache ca-certificates tzdata

# Оптимизированные environment variables
ENV GIN_MODE=release
ENV GOGC=100
ENV GOMAXPROCS=1

# Read-only filesystem
RUN mkdir -p /tmp/cloudbridge
VOLUME ["/tmp/cloudbridge"]
```

### 3. **Network Optimization**

```yaml
# ✅ ХОРОШО: Network оптимизации
services:
  cloudbridge-relay:
    networks:
      - cloudbridge-network
    ports:
      - "9090:9090"
      - "9090:9090/udp"  # QUIC protocol

networks:
  cloudbridge-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

## 🧪 Тестирование

### 1. **Unit Testing в контейнерах**

```dockerfile
# ✅ ХОРОШО: Testing stage
FROM golang:1.25-alpine AS tester
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test -v ./...
```

### 2. **Integration Testing**

```yaml
# ✅ ХОРОШО: Test environment
version: '3.8'
services:
  cloudbridge-relay:
    build:
      context: .
      dockerfile: Dockerfile.relay.enhanced
    environment:
      - LOG_LEVEL=debug
      - TEST_MODE=true
    depends_on:
      - test-cassandra

  test-cassandra:
    image: cassandra:3.11
    environment:
      - CASSANDRA_CLUSTER_NAME=test-cluster
```

## 📚 Полезные команды

### Build Commands

```bash
# Сборка с оптимизациями
docker build --no-cache --pull -f Dockerfile.relay.enhanced .

# Сборка для разных архитектур
docker buildx build --platform linux/amd64,linux/arm64 -f Dockerfile.relay.enhanced .

# Анализ размера образа
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"
```

### Runtime Commands

```bash
# Запуск с resource limits
docker run --memory=1g --cpus=1.0 cloudbridge-relay:latest

# Запуск с volumes
docker run -v $(pwd)/configs:/app/configs:ro cloudbridge-relay:latest

# Запуск в сети
docker run --network cloudbridge-network cloudbridge-relay:latest
```

### Monitoring Commands

```bash
# Мониторинг ресурсов
docker stats --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"

# Проверка health status
docker inspect --format='{{.State.Health.Status}}' cloudbridge-relay

# Анализ логов
docker logs cloudbridge-relay 2>&1 | grep ERROR
```

## 🎯 Заключение

Следование этим best practices обеспечивает:

- ✅ **Безопасность** - минимальные привилегии и non-root пользователи
- ✅ **Производительность** - оптимизированные образы и ресурсы
- ✅ **Надежность** - health checks и мониторинг
- ✅ **Масштабируемость** - resource limits и оптимизации
- ✅ **Поддерживаемость** - структурированные логи и debugging

---

*Документ создан: 2025-09-06*  
*Статус: ✅ Готов к использованию* 🚀
