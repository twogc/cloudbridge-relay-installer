# =========================================================
# CloudBridge Relay (Debian-based, без git.zx2c4 зависимостей)
# =========================================================

# ---------- Builder: Go binary ----------
  FROM golang:1.25.12-bookworm AS builder

  ENV CGO_ENABLED=1 GOOS=linux
  WORKDIR /app
  
  RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates git build-essential \
      && rm -rf /var/lib/apt/lists/*
  
  # Cache go modules
  COPY go.mod go.sum ./
  RUN go mod download
  
  # App sources
  COPY . .
  
  # Metadata
  ARG VERSION=dev
  ARG BUILD_TIME
  ARG GIT_COMMIT
  
  # Build
  RUN go build -ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
        -o cloudbridge-relay ./cmd/relay
  
  
  # ---------- (Optional) Builder: wireguard-go from GitHub mirror ----------
  # Используется только если WITH_WG_GO=true
  FROM golang:1.25.12-bookworm AS wggo-builder
  
  WORKDIR /tmp/wireguard-go
  RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates git build-essential make \
      && rm -rf /var/lib/apt/lists/*
  
  # Официальное зеркало WireGuard-go на GitHub (mirror-only)
  RUN git clone --depth=1 https://github.com/WireGuard/wireguard-go.git . \
   && make \
   && install -m 0755 wireguard-go /usr/local/bin/wireguard-go
  
  
  # ---------- Final image ----------
  FROM debian:bookworm-slim
  
  # Toggle: собирать и класть userspace wireguard-go внутрь контейнера
  # По умолчанию false (рассчитываем на ядровой модуль у хоста)
  ARG WITH_WG_GO=false
  
  # Базовые пакеты для рантайма
  RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates tzdata bash curl \
        iproute2 iptables wireguard-tools \
      && rm -rf /var/lib/apt/lists/*
  
  # Если включен userspace WG — добавить бинарник из wggo-builder
  # (Работать будет только при запуске контейнера с CAP_NET_ADMIN и /dev/net/tun)
  COPY --from=wggo-builder /usr/local/bin/wireguard-go /usr/local/bin/wireguard-go
  RUN if [ "${WITH_WG_GO}" != "true" ]; then rm -f /usr/local/bin/wireguard-go 2>/dev/null || true; fi
  
  # Пользователь и директории
  RUN groupadd -g 1000 cloudbridge && \
      useradd  -m -u 1000 -g 1000 -s /bin/bash cloudbridge && \
      mkdir -p /app/configs /app/logs /etc/wireguard /var/run/wireguard /var/log/cloudbridge-relay
  
  # Бинарник приложения
  COPY --from=builder /app/cloudbridge-relay /usr/local/bin/cloudbridge-relay
  
  # (Опционально) базовый конфиг внутрь образа — можно заменить своим томом
  COPY configs/config.json /app/configs/config.json
  
  # Права
  RUN chmod +x /usr/local/bin/cloudbridge-relay \
   && chown -R cloudbridge:cloudbridge /app /etc/wireguard /var/run/wireguard /var/log/cloudbridge-relay
  
  USER cloudbridge
  WORKDIR /app
  
  # --------- Порты ---------
  # 9090: TCP API/QUIC (TCP/UDP), 51820/udp: WireGuard, 8084: метрики
  EXPOSE 9090 \
         9090/udp \
         51820/udp \
         8084
  
  # Healthcheck (на API порт 9090)
  HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -fsS http://localhost:9090/health || exit 1
  
  # ENV
  ENV CONFIG_PATH=/app/configs/config.json \
      LOG_LEVEL=info \
      WIREGUARD_ENABLED=true
  
  ENTRYPOINT ["/usr/local/bin/cloudbridge-relay"]
  CMD ["--config", "/app/configs/config.json"]
  