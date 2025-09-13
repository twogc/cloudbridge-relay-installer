# Локальный запуск CloudBridge Relay в Docker

## Быстрый старт

### 1. Автоматический запуск (РЕКОМЕНДУЕТСЯ)
```bash
./run-local.sh
```

### 2. Ручной запуск через Makefile
```bash
# Сборка Alpine образа
make docker-build-alpine

# Запуск контейнера
make docker-run-alpine
```

### 3. Прямые Docker команды
```bash
# Сборка
docker build -f Dockerfile.relay.alpine -t cloudbridge-relay:alpine .

# Запуск
docker run -d \
    --name cloudbridge-relay-alpine \
    --cap-add=NET_ADMIN \
    --device=/dev/net/tun \
    -p 80:80 \
    -p 8080:8080 \
    -p 8081:8081 \
    -p 8082:8082 \
    -p 9090:9090 \
    -p 51820:51820/udp \
    -v $(pwd)/configs:/app/configs:ro \
    -v $(pwd)/logs:/app/logs \
    cloudbridge-relay:alpine
```

## Управление контейнером

### Просмотр логов
```bash
docker logs cloudbridge-relay-alpine
# или
make docker-logs
```

### Статус контейнера
```bash
docker ps | grep cloudbridge
# или
make docker-status
```

### Остановка
```bash
docker stop cloudbridge-relay-alpine
docker rm cloudbridge-relay-alpine
# или
make docker-stop-local
```

### Перезапуск
```bash
./run-local.sh
```

## Доступные порты

- **80** - HTTP (Health checks)
- **8080** - Main API
- **8081** - Metrics (Prometheus)
- **8082** - Secondary API
- **9090** - QUIC Protocol
- **51820** - WireGuard P2P Mesh (UDP)

## Проверка работы

### Health check
```bash
curl http://localhost:80/health
```

### API
```bash
curl http://localhost:8080/api/v1/status
```

### Metrics
```bash
curl http://localhost:8081/metrics
```

## Конфигурация

Конфигурация загружается из `configs/config.json` и монтируется в контейнер как read-only.

## Логи

Логи приложения сохраняются в директории `logs/` на хосте.

## Troubleshooting

### Проблема с TUN устройством
```bash
# Проверить доступность TUN
ls -la /dev/net/tun

# Создать если отсутствует
sudo mkdir -p /dev/net
sudo mknod /dev/net/tun c 10 200
sudo chmod 600 /dev/net/tun
```

### Проблемы с портами
```bash
# Проверить занятые порты
netstat -tulpn | grep :80
netstat -tulpn | grep :8080
```

### Проблемы с правами
```bash
# Проверить права на configs и logs
ls -la configs/
ls -la logs/
```

## Преимущества Alpine версии

1. **Нет DNS проблем** - использует Alpine пакеты
2. **Быстрая сборка** - не компилирует wireguard-go
3. **Меньший размер** - оптимизированный образ
4. **Надежность** - проверенные пакеты дистрибутива
