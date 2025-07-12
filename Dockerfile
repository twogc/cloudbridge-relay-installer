# CloudBridge Relay v2.0 Dockerfile
# Enterprise-grade высокопроизводительный relay сервер

# Multi-stage build для оптимизации размера образа
FROM golang:1.23-alpine AS builder

# Установка необходимых пакетов для сборки
RUN apk add --no-cache git ca-certificates tzdata

# Установка рабочей директории
WORKDIR /app

# Копирование файлов зависимостей
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка relay сервера с оптимизациями
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags "-w -s -X main.Version=2.0.0 -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
    -o relay ./cmd/relay

# Сборка API сервера
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags "-w -s -X main.Version=2.0.0 -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S')" \
    -o api ./cmd/api

# Финальный образ
FROM alpine:latest

# Установка необходимых пакетов
RUN apk --no-cache add ca-certificates tzdata curl

# Создание пользователя для безопасности
RUN addgroup -g 1001 -S cloudbridge && \
    adduser -u 1001 -S cloudbridge -G cloudbridge

# Установка рабочей директории
WORKDIR /app

# Копирование бинарных файлов из builder stage
COPY --from=builder /app/relay .
COPY --from=builder /app/api .
RUN chmod +x /app/relay /app/api

# Копирование конфигурационных файлов
COPY --from=builder /app/config ./config

# Создание директорий для логов и данных
RUN mkdir -p /app/logs /app/data && \
    chown -R cloudbridge:cloudbridge /app

# Установка supervisor для Alpine
RUN apk add --no-cache supervisor

# Копирование конфигурации supervisor
COPY supervisord.conf /etc/supervisord.conf

# Переключение на непривилегированного пользователя
USER cloudbridge

# Проверка здоровья
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

# Экспорт портов
EXPOSE 8080 8081 8082

# Метаданные
LABEL maintainer="CloudBridge Team <dev@cloudbridge.com>"
LABEL version="2.0.0"
LABEL description="Enterprise-grade высокопроизводительный relay сервер"
LABEL org.opencontainers.image.title="CloudBridge Relay"
LABEL org.opencontainers.image.description="Enterprise-grade высокопроизводительный relay сервер с интеграцией лучших практик Cloudflare"
LABEL org.opencontainers.image.version="2.0.0"
LABEL org.opencontainers.image.vendor="CloudBridge"
LABEL org.opencontainers.image.source="https://github.com/cloudbridge/relay"

# Переменные окружения
ENV RELAY_CONFIG_PATH=/app/config/config.json
ENV RELAY_LOG_LEVEL=info
ENV RELAY_METRICS_PORT=8081
ENV RELAY_SERVER_PORT=8080
ENV API_PORT=8082

# Запуск через supervisor
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"] 