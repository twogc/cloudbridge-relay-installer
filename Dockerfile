# Multi-stage build для CloudBridge Relay WireGuard
FROM golang:1.23-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache \
    git \
    build-base \
    linux-headers \
    wireguard-tools

# Установка рабочей директории
WORKDIR /app

# Копирование go mod файлов
COPY go.mod go.sum ./

# Скачивание зависимостей
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o cloudbridge-relay ./cmd/relay

# Сборка WireGuard-go
RUN git clone https://git.zx2c4.com/wireguard-go.git /tmp/wireguard-go && \
    cd /tmp/wireguard-go && \
    make && \
    cp wireguard-go /app/

# Финальный образ
FROM alpine:3.18

# Установка необходимых пакетов
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    iproute2 \
    wireguard-tools \
    iptables \
    bash \
    curl \
    && rm -rf /var/cache/apk/*

# Создание пользователя
RUN addgroup -g 1000 cloudbridge && \
    adduser -D -s /bin/bash -u 1000 -G cloudbridge cloudbridge

# Создание необходимых директорий
RUN mkdir -p \
    /app/configs \
    /app/logs \
    /etc/wireguard \
    /var/run/wireguard \
    /var/log/cloudbridge-relay

# Копирование бинарных файлов
COPY --from=builder /app/cloudbridge-relay /usr/local/bin/
COPY --from=builder /app/wireguard-go /usr/local/bin/

# Установка прав
RUN chmod +x /usr/local/bin/cloudbridge-relay && \
    chmod +x /usr/local/bin/wireguard-go && \
    chown -R cloudbridge:cloudbridge /app /etc/wireguard /var/run/wireguard /var/log/cloudbridge-relay

# Копирование конфигурации по умолчанию
COPY configs/config.json /app/configs/config.json

# Переключение на пользователя
USER cloudbridge

# Рабочая директория
WORKDIR /app

# Открытие портов
EXPOSE 8080 8081 8082 80 443 9090 51820/udp

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Переменные окружения
ENV CONFIG_PATH=/app/configs/config.json
ENV LOG_LEVEL=info
ENV WIREGUARD_ENABLED=true

# Команда запуска
ENTRYPOINT ["/usr/local/bin/cloudbridge-relay"]
CMD ["--config", "/app/configs/config.json"] 