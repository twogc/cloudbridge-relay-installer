# CloudBridge Relay - Подключение через cloudflared

## 🎯 Доступные адреса для подключения

### Основной сервер CloudBridge Relay:
- **Публичный IP**: `95.163.250.190`
- **LoadBalancer IP**: `10.106.92.52` (внутренний)
- **NodePort**: `95.163.250.190:30080-30090` (внешний доступ)

## 🔧 Настройка cloudflared на клиентской машине

### 1. Установка cloudflared

**Ubuntu/Debian:**
```bash
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared-linux-amd64.deb
```

**CentOS/RHEL:**
```bash
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.rpm
sudo rpm -i cloudflared-linux-amd64.rpm
```

**Docker:**
```bash
docker run --rm -it cloudflare/cloudflared:latest --version
```

### 2. Конфигурация cloudflared

Создайте файл конфигурации `/etc/cloudflared/config.yml`:

```yaml
tunnel: <TUNNEL_ID>
credentials-file: /etc/cloudflared/credentials.json

# Настройки для подключения к CloudBridge Relay
ingress:
  # Основной CloudBridge Relay API
  - hostname: cloudbridge.yourdomain.com
    service: http://95.163.250.190:30080
    originRequest:
      httpHostHeader: cloudbridge.yourdomain.com
      noTLSVerify: true

  # CloudBridge Relay Health Check
  - hostname: cloudbridge-health.yourdomain.com
    service: http://95.163.250.190:30081
    originRequest:
      httpHostHeader: cloudbridge-health.yourdomain.com
      noTLSVerify: true

  # CloudBridge Relay P2P API
  - hostname: cloudbridge-p2p.yourdomain.com
    service: http://95.163.250.190:32500
    originRequest:
      httpHostHeader: cloudbridge-p2p.yourdomain.com
      noTLSVerify: true

  # CloudBridge Relay QUIC
  - hostname: cloudbridge-quic.yourdomain.com
    service: http://95.163.250.190:30087
    originRequest:
      httpHostHeader: cloudbridge-quic.yourdomain.com
      noTLSVerify: true

  # CloudBridge Relay STUN
  - hostname: cloudbridge-stun.yourdomain.com
    service: http://95.163.250.190:30302
    originRequest:
      httpHostHeader: cloudbridge-stun.yourdomain.com
      noTLSVerify: true

  # CloudBridge Relay Enhanced QUIC
  - hostname: cloudbridge-enhanced-quic.yourdomain.com
    service: http://95.163.250.190:30090
    originRequest:
      httpHostHeader: cloudbridge-enhanced-quic.yourdomain.com
      noTLSVerify: true

  # CloudBridge Relay MASQUE HTTP/3
  - hostname: cloudbridge-masque.yourdomain.com
    service: http://95.163.250.190:30443
    originRequest:
      httpHostHeader: cloudbridge-masque.yourdomain.com
      noTLSVerify: true

  # CloudBridge Relay WebRTC P2P
  - hostname: cloudbridge-webrtc.yourdomain.com
    service: http://95.163.250.190:30092
    originRequest:
      httpHostHeader: cloudbridge-webrtc.yourdomain.com
      noTLSVerify: true

  # Catch-all правило
  - service: http_status:404
```

### 3. Альтернативная простая конфигурация

Если нужен только базовый доступ, используйте упрощенную конфигурацию:

```yaml
tunnel: <TUNNEL_ID>
credentials-file: /etc/cloudflared/credentials.json

ingress:
  # Все запросы идут на CloudBridge Relay
  - service: http://95.163.250.190:30080
    originRequest:
      noTLSVerify: true
```

### 4. Настройка DNS записей

В Cloudflare Dashboard добавьте следующие DNS записи:

```
Type    Name                           Content                    Proxy
CNAME   cloudbridge                   your-tunnel.trycloudflare.com    ✅
CNAME   cloudbridge-health            your-tunnel.trycloudflare.com    ✅
CNAME   cloudbridge-p2p               your-tunnel.trycloudflare.com    ✅
CNAME   cloudbridge-quic              your-tunnel.trycloudflare.com    ✅
CNAME   cloudbridge-stun              your-tunnel.trycloudflare.com    ✅
CNAME   cloudbridge-enhanced-quic     your-tunnel.trycloudflare.com    ✅
CNAME   cloudbridge-masque            your-tunnel.trycloudflare.com    ✅
CNAME   cloudbridge-webrtc            your-tunnel.trycloudflare.com    ✅
```

### 5. Запуск cloudflared

```bash
# Создание директории для конфигурации
sudo mkdir -p /etc/cloudflared

# Копирование конфигурации
sudo cp config.yml /etc/cloudflared/

# Установка прав доступа
sudo chmod 600 /etc/cloudflared/config.yml

# Запуск cloudflared
sudo cloudflared tunnel run <TUNNEL_NAME>
```

### 6. Автозапуск (systemd)

Создайте файл `/etc/systemd/system/cloudflared.service`:

```ini
[Unit]
Description=Cloudflared Tunnel
After=network.target

[Service]
Type=simple
User=cloudflared
ExecStart=/usr/local/bin/cloudflared tunnel run <TUNNEL_NAME>
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Активация сервиса:
```bash
sudo systemctl daemon-reload
sudo systemctl enable cloudflared
sudo systemctl start cloudflared
```

## 🔍 Проверка подключения

### Тестирование доступности CloudBridge Relay:

```bash
# Проверка основного API
curl -k https://cloudbridge.yourdomain.com/health

# Проверка P2P API
curl -k https://cloudbridge-p2p.yourdomain.com/health

# Проверка QUIC
curl -k https://cloudbridge-quic.yourdomain.com/health
```

### Мониторинг логов:

```bash
# Логи cloudflared
sudo journalctl -u cloudflared -f

# Логи CloudBridge Relay (на сервере)
kubectl logs -n cloudbridge -l app=cloudbridge-relay -f
```

## 🚨 Устранение неполадок

### Проблема: "Connection refused"
- Проверьте, что CloudBridge Relay работает: `kubectl get pods -n cloudbridge`
- Проверьте файрвол: `sudo ufw status`
- Проверьте доступность портов: `telnet 95.163.250.190 30080`

### Проблема: "Tunnel not found"
- Проверьте TUNNEL_ID в конфигурации
- Проверьте credentials-file
- Выполните: `cloudflared tunnel login`

### Проблема: "DNS resolution failed"
- Проверьте DNS записи в Cloudflare Dashboard
- Проверьте, что домен указывает на ваш туннель

## 📊 Мониторинг

### Статус туннеля:
```bash
cloudflared tunnel list
cloudflared tunnel info <TUNNEL_NAME>
```

### Статус CloudBridge Relay:
```bash
# На сервере
kubectl get pods -n cloudbridge -l app=cloudbridge-relay
kubectl logs -n cloudbridge -l app=cloudbridge-relay --tail=50
```

## 🎯 Готово!

После настройки клиенты смогут подключаться к CloudBridge Relay через:
- `https://cloudbridge.yourdomain.com` - основной API
- `https://cloudbridge-p2p.yourdomain.com` - P2P API
- `https://cloudbridge-quic.yourdomain.com` - QUIC транспорт
- И другие сервисы...

Все подключения будут проходить через Magic WAN туннель с шифрованием AES-256!

