<<<<<<< HEAD
# 2GC CloudBridge tunnel - TCP Port Forwarding with Relay

2GC CloudBridge tunnel — это простая система проброса TCP-портов, позволяющая публиковать локальные сервисы через релейный сервер. Аналогична ngrok, но с более простым протоколом и реализацией.

## Features

- Простой проброс TCP через релейный сервер
- Поддержка ролей agent (backend) и consumer (frontend)
- Автоматическое переподключение
- Токенизированная идентификация туннеля
- Двунаправленная передача данных
- Кроссплатформенность (Linux, macOS, Windows)
- Поддержка российских дистрибутивов Linux (Astra Linux, ALT Linux, ROSA Linux, RED OS)
- systemd-сервисы для Linux
- Windows-сервисы
- launchd-сервисы для macOS

## Building

### Prerequisites

- Go 1.16 или новее
- Make (опционально, для Makefile)

### Using Makefile

Самый простой способ собрать для всех платформ — использовать Makefile:

```bash
make build
make build-relay
make build-client
make build-russian
make release-all
make rpm
make deb
sudo make install-linux
make install-windows
sudo make install-macos
```

Бинарники будут в директории `build`:
- `build/relay/cloudbridge-relay-{os}-{arch}` — бинарники релея
- `build/client/cloudbridge-client-{os}-{arch}` — бинарники клиента

Поддерживаемые платформы:
- Linux (amd64, arm64, 386)
- macOS (amd64, arm64)
- Windows (amd64, 386)

Поддерживаемые российские дистрибутивы:
- Astra Linux (amd64, arm64)
- ALT Linux (amd64, arm64)
- ROSA Linux (amd64, arm64)
- RED OS (amd64, arm64)

### Manual Build

```bash
# Build relay server
go build -o cloudbridge-relay ./cmd/relay
# Build client
go build -o cloudbridge-client ./cmd/client
```

Для сборки под конкретную платформу:

```bash
GOOS=windows GOARCH=amd64 go build -o cloudbridge-relay.exe ./cmd/relay
GOOS=windows GOARCH=amd64 go build -o cloudbridge-client.exe ./cmd/client
```

## Installation

### From Binary

Скачайте бинарник для вашей платформы с releases.

### From Package

#### RPM-based distributions (Astra Linux, ALT Linux, ROSA Linux, RED OS)

```bash
sudo rpm -i cloudbridge-*.rpm
```

#### DEB-based distributions

```bash
sudo dpkg -i cloudbridge_*.deb
```

### Service Installation

#### Linux (systemd)

```bash
sudo systemctl enable cloudbridge-relay
sudo systemctl start cloudbridge-relay
sudo systemctl enable cloudbridge-agent
sudo systemctl start cloudbridge-agent
sudo systemctl status cloudbridge-relay
sudo systemctl status cloudbridge-agent
```

#### Windows

1. Win+R → services.msc
2. Найдите "CloudBridge Relay Server" или "CloudBridge Agent Client"
3. Управляйте сервисами

Или через PowerShell:

```powershell
Start-Service CloudBridgeRelay
Start-Service CloudBridgeAgent
Stop-Service CloudBridgeRelay
Stop-Service CloudBridgeAgent
Get-Service CloudBridgeRelay
Get-Service CloudBridgeAgent
```

#### macOS (launchd)

```bash
launchctl list | grep cloudbridge
```

## Usage

### Relay Server

```bash
./cloudbridge-relay --addr :8080
cloudbridge-relay.exe --addr :8080
```

### Agent (Backend)

```bash
./cloudbridge-client --relay relay.localhost:8080 --local localhost:3000 --token my-tunnel --mode agent
cloudbridge-client.exe --relay relay.localhost:8080 --local localhost:3000 --token my-tunnel --mode agent
```

### Consumer (Frontend)

```bash
./cloudbridge-client --relay relay.localhost:8080 --local localhost:8080 --token my-tunnel --mode consumer
cloudbridge-client.exe --relay relay.localhost:8080 --local localhost:8080 --token my-tunnel --mode consumer
```

## Protocol

Протокол простой и основан на JSON:

```json
{
    "type": "agent|consumer",
    "token": "tunnel-token"
}
```

## Security Considerations

- TCP без шифрования
- Токены в открытом виде
- Для продакшена используйте TLS
- Реализуйте аутентификацию при необходимости

## Future Improvements

- [ ] TLS
- [ ] Аутентификация
- [ ] WebSocket
- [ ] UDP
- [ ] Rate limiting
- [ ] Connection pooling
- [ ] Метрики и мониторинг
- [ ] Поддержка новых российских дистрибутивов
- [ ] systemd для российских дистрибутивов

## License

MIT License

# CloudBridge Relay Server

Высокопроизводительный Relay-сервер для туннелирования TCP-соединений, оптимизированный для Ubuntu Server 22.04 LTS.

## Особенности

- Высокопроизводительная обработка TCP-соединений
- Поддержка TLS для безопасной передачи данных
- Система аутентификации на основе токенов
- Механизм keepalive для поддержания соединений
- Rate limiting и защита от DoS-атак
- Метрики Prometheus для мониторинга
- Поддержка systemd для управления сервисом
- Оптимизация для Ubuntu Server 22.04 LTS

## Требования

- Ubuntu Server 22.04 LTS
- Go 1.21 или выше
- systemd
- Prometheus (опционально, для метрик)

## Установка

### Из исходного кода

```bash
git clone https://github.com/your-org/cloudbridge-relay.git
cd cloudbridge-relay
make build
sudo make install
```

### Из deb-пакета

```bash
wget https://github.com/your-org/cloudbridge-relay/releases/latest/download/cloudbridge-relay.deb
sudo dpkg -i cloudbridge-relay.deb
```

## Конфигурация

Основной конфигурационный файл находится в `/etc/cloudbridge-relay/config.yaml`:

```yaml
server:
  listen: ":443"
  tls:
    enabled: true
    cert_file: "/etc/cloudbridge-relay/certs/server.crt"
    key_file: "/etc/cloudbridge-relay/certs/server.key"
  max_connections: 10000
  connection_timeout: 60s

security:
  token_validation:
    min_length: 8
    max_length: 32
    allowed_chars: "[a-zA-Z0-9-_]"
  rate_limit:
    enabled: true
    requests_per_second: 10
    burst: 20

metrics:
  enabled: true
  listen: ":9090"
  path: "/metrics"

logging:
  level: "info"
  format: "json"
  output: "/var/log/cloudbridge-relay/relay.log"
```

## Управление сервисом

```bash
# Запуск сервиса
sudo systemctl start cloudbridge-relay

# Проверка статуса
sudo systemctl status cloudbridge-relay

# Остановка сервиса
sudo systemctl stop cloudbridge-relay

# Включение автозапуска
sudo systemctl enable cloudbridge-relay
```

## Мониторинг

Сервер предоставляет метрики Prometheus по адресу `http://localhost:9090/metrics`:

- `cloudbridge_relay_active_connections` - количество активных соединений
- `cloudbridge_relay_total_connections` - общее количество соединений
- `cloudbridge_relay_connection_errors` - количество ошибок соединения
- `cloudbridge_relay_bytes_transferred` - количество переданных байт
- `cloudbridge_relay_registration_requests` - количество запросов на регистрацию

## Протокол

Сервер использует JSON-based протокол для обмена сообщениями:

### Типы сообщений

- `register` - регистрация туннеля
- `connect` - запрос на соединение
- `ping` - проверка соединения
- `pong` - ответ на ping
- `error` - сообщение об ошибке

### Формат сообщений

```json
{
  "type": "register",
  "token": "your-token",
  "version": "1.0.0",
  "local_port": 3000,
  "user_agent": "cloudbridge-agent/1.0.0",
  "platform": "linux/amd64",
  "build_version": "1.0.0"
}
```

## Безопасность

- Все соединения защищены TLS
- Токены должны соответствовать требованиям безопасности
- Rate limiting для защиты от DoS-атак
- Логирование всех действий для аудита

## Лицензия

MIT License

# CloudBridge Relay Server Installer

Установщик для CloudBridge Relay Server - высокопроизводительного сервера для туннелирования TCP-соединений.
=======
# CloudBridge Relay Server Installer

This repository contains the installation script and configuration files for the CloudBridge Relay Server.
>>>>>>> d9da572 (feat: add systemd service and config files)

## Contents

- `install.sh` - Main installation script
- `systemd/cloudbridge-relay.service` - Systemd service file
- `config/config.yaml` - Default configuration file
- `LICENSE` - MIT License

## Quick Install

To install the CloudBridge Relay Server, run:

```bash
curl -sSL https://raw.githubusercontent.com/twogc/cloudbridge-relay-installer/main/install.sh | sudo bash
```

## Manual Installation

1. Clone this repository:
```bash
git clone https://github.com/twogc/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer
```

2. Make the installer executable:
```bash
chmod +x install.sh
```

3. Run the installer:
```bash
sudo ./install.sh
```

## What Gets Installed

The installer will:

1. Create necessary directories:
   - `/etc/cloudbridge-relay/` - Configuration files
   - `/var/lib/cloudbridge-relay/` - Data directory
   - `/var/log/cloudbridge-relay/` - Log files

2. Create a system user `cloudbridge-relay`

3. Download and install the latest release of CloudBridge Relay Server

4. Install and configure the systemd service

5. Set up the default configuration

6. Start the service

## Post-Installation

After installation:

1. Check the service status:
```bash
systemctl status cloudbridge-relay
```

2. View the logs:
```bash
tail -f /var/log/cloudbridge-relay/relay.log
```

3. Edit the configuration:
```bash
sudo nano /etc/cloudbridge-relay/config.yaml
```

4. Restart the service after configuration changes:
```bash
sudo systemctl restart cloudbridge-relay
```

## Configuration

The default configuration is installed at `/etc/cloudbridge-relay/config.yaml`. You can modify it to:

- Change the listening port
- Enable/disable TLS
- Configure rate limiting
- Adjust logging settings
- Enable/disable metrics

## Security

The installer:
- Creates a dedicated system user
- Sets appropriate file permissions
- Configures systemd service with security hardening
- Uses TLS by default

## Support

For issues and feature requests, please visit:
https://github.com/twogc/cloudbridge-relay/issues

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 