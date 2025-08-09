# 🛠️ Инструменты CloudBridge Relay

Утилиты и инструменты для установки, настройки и управления CloudBridge Relay.

## 📁 Структура

### `/setup/` - Установка и настройка
- `install.sh` - Универсальный установщик CloudBridge Relay
- `setup_cloudbridge.sh` - Настройка системы для CloudBridge

### `/utils/` - Утилиты
- `build.sh` - Сборка проекта
- `start_production.sh` - Производственный запуск

## 🚀 Быстрая установка

### Автоматическая установка:
```bash
curl -sSL https://raw.githubusercontent.com/twogc/cloudbridge-relay-installer/main/tools/setup/install.sh | sudo bash
```

### Локальная установка:
```bash
cd setup/
sudo ./install.sh
```

### Docker установка:
```bash
cd setup/
sudo ./install.sh --docker
```

## 📋 Требования

- **Linux** - Ubuntu 20.04+ / CentOS 8+ / Debian 11+
- **Docker** - для контейнерного развертывания
- **WireGuard** - для P2P mesh функциональности
- **Go 1.21+** - для сборки из исходников

## 🔧 Конфигурация

После установки отредактируйте конфигурационные файлы:
- `/etc/cloudbridge/config.yaml` - основная конфигурация
- `/etc/cloudbridge/wireguard.yaml` - настройки P2P mesh
- `/etc/cloudbridge/enhanced_quic.yaml` - настройки QUIC

## 📊 Мониторинг

Для производственного использования настройте мониторинг:
```bash
cd utils/
./start_production.sh --with-monitoring
```

Это запустит:
- Prometheus на порту 9090
- Grafana на порту 3000
- AlertManager на порту 9093
