# CloudBridge Relay как Cloudflare Origin - Полная настройка

## Цель
Настроить CloudBridge Relay (`edge.2gc.ru`) как Origin endpoint для Cloudflare через Magic WAN туннель, чтобы клиенты `cloudflared` подключались через ваш сервер.

## Архитектура

```
[Cloudflared Clients] → [Cloudflare] → [Magic WAN туннель] → [edge.2gc.ru:443] → [nginx] → [CloudBridge Relay в подах]
```

## Пошаговая настройка

### 1. **Автоматическая настройка (рекомендуется)**

```bash
# Запуск автоматического скрипта настройки
sudo ./setup-cloudflare-origin.sh
```

### 2. **Ручная настройка**

#### 2.1 Проверка Magic WAN туннеля
```bash
# Проверка статуса туннеля
sudo swanctl --list-sas

# Должно показать:
# magic-wan-primary: #1, ESTABLISHED, IKEv2
```

#### 2.2 Проверка CloudBridge Relay
```bash
# Проверка подов
kubectl get pods -n cloudbridge -l app=cloudbridge-relay

# Проверка health endpoint
curl -k https://edge.2gc.ru/health
```

#### 2.3 Обновление nginx конфигурации
```bash
# Создание резервной копии
sudo cp /etc/nginx/conf.d/edge-2gc-ru.conf /etc/nginx/conf.d/edge-2gc-ru.conf.backup

# Установка новой конфигурации
sudo cp nginx/cloudflare-origin-proxy.conf /etc/nginx/conf.d/edge-2gc-ru.conf

# Проверка конфигурации
sudo nginx -t

# Перезагрузка nginx
sudo systemctl reload nginx
```

### 3. **Настройка Cloudflare Dashboard**

#### 3.1 Создание Tunnel
1. Перейдите в **Zero Trust** → **Access** → **Tunnels**
2. Создайте новый tunnel: `cloudbridge-relay-tunnel`
3. Скопируйте **Tunnel ID** и **Token**

#### 3.2 Настройка Ingress Rules
В разделе **Public Hostnames** добавьте:

```
Hostname: app.client.company.ru
Service: https://edge.2gc.ru:443
```

Или для множественных доменов:
```
Hostname: *.client.company.ru
Service: https://edge.2gc.ru:443
```

#### 3.3 Настройка Magic WAN
1. Перейдите в **Magic WAN** → **Tunnels**
2. Убедитесь, что туннель к `95.163.250.190` активен
3. Проверьте, что `edge.2gc.ru` доступен через Magic WAN

### 4. **Настройка cloudflared клиентов**

#### 4.1 Установка cloudflared
```bash
# Ubuntu/Debian
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared-linux-amd64.deb

# CentOS/RHEL
wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.rpm
sudo rpm -i cloudflared-linux-amd64.rpm
```

#### 4.2 Конфигурация cloudflared
Создайте файл `/etc/cloudflared/config.yml`:

```yaml
tunnel: <YOUR_TUNNEL_UUID>
credentials-file: /etc/cloudflared/credentials.json

ingress:
  # Основной CloudBridge Relay API
  - hostname: app.yourdomain.com
    service: https://edge.2gc.ru/api/
    originRequest:
      httpHostHeader: app.yourdomain.com
      noTLSVerify: false
      tlsTimeout: 10s
      connectTimeout: 10s

  # CloudBridge Relay Health Check
  - hostname: health.yourdomain.com
    service: https://edge.2gc.ru/health
    originRequest:
      httpHostHeader: health.yourdomain.com
      noTLSVerify: false

  # CloudBridge Relay P2P API
  - hostname: p2p.yourdomain.com
    service: https://edge.2gc.ru/p2p/
    originRequest:
      httpHostHeader: p2p.yourdomain.com
      noTLSVerify: false

  # Catch-all правило
  - service: http_status:404
```

#### 4.3 Запуск cloudflared
```bash
# Создание директории
sudo mkdir -p /etc/cloudflared

# Копирование конфигурации
sudo cp config.yml /etc/cloudflared/

# Установка прав
sudo chmod 600 /etc/cloudflared/config.yml

# Запуск
sudo cloudflared tunnel run <TUNNEL_NAME>
```

#### 4.4 Автозапуск (systemd)
```bash
# Создание сервиса
sudo tee /etc/systemd/system/cloudflared.service << 'EOF'
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
EOF

# Активация
sudo systemctl daemon-reload
sudo systemctl enable cloudflared
sudo systemctl start cloudflared
```

## Проверка работы

### 1. **Проверка Magic WAN туннеля**
```bash
# Статус туннеля
sudo swanctl --list-sas

# Трафик через туннель
sudo tcpdump -i any esp
```

### 2. **Проверка CloudBridge Relay**
```bash
# Статус подов
kubectl get pods -n cloudbridge -l app=cloudbridge-relay

# Логи
kubectl logs -n cloudbridge -l app=cloudbridge-relay -f
```

### 3. **Проверка nginx**
```bash
# Статус nginx
sudo systemctl status nginx

# Логи nginx
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### 4. **Проверка endpoints**
```bash
# Health check
curl -k https://edge.2gc.ru/health

# API endpoint
curl -k https://edge.2gc.ru/api/

# P2P endpoint
curl -k https://edge.2gc.ru/p2p/
```

### 5. **Проверка cloudflared клиента**
```bash
# Статус туннеля
cloudflared tunnel list
cloudflared tunnel info <TUNNEL_NAME>

# Тест подключения
curl -k https://app.yourdomain.com/health
```

## Мониторинг

### 1. **Автоматический мониторинг**
```bash
# Запуск скрипта мониторинга
/usr/local/bin/cloudbridge-monitor.sh

# Просмотр логов мониторинга
tail -f /var/log/cloudbridge-monitor.log
```

### 2. **Ручной мониторинг**
```bash
# Magic WAN туннель
sudo swanctl --list-sas

# CloudBridge Relay поды
kubectl get pods -n cloudbridge -l app=cloudbridge-relay

# nginx статус
sudo systemctl status nginx

# Health check
curl -k https://edge.2gc.ru/health
```

## Устранение неполадок

### Проблема: "Connection refused"
```bash
# Проверьте Magic WAN туннель
sudo swanctl --list-sas

# Проверьте nginx
sudo systemctl status nginx

# Проверьте CloudBridge Relay
kubectl get pods -n cloudbridge -l app=cloudbridge-relay
```

### Проблема: "Tunnel not found"
```bash
# Проверьте TUNNEL_ID в конфигурации
cat /etc/cloudflared/config.yml

# Проверьте credentials-file
ls -la /etc/cloudflared/credentials.json

# Выполните повторную авторизацию
cloudflared tunnel login
```

### Проблема: "DNS resolution failed"
```bash
# Проверьте DNS записи в Cloudflare Dashboard
# Проверьте, что домен указывает на ваш туннель
```

### Проблема: "SSL certificate error"
```bash
# Проверьте SSL сертификаты
sudo certbot certificates

# Проверьте nginx SSL конфигурацию
sudo nginx -t
```

## Готовые конфигурации

### 1. **Простая конфигурация (только основной API)**
```yaml
tunnel: <TUNNEL_ID>
credentials-file: /etc/cloudflared/credentials.json

ingress:
  - hostname: app.yourdomain.com
    service: https://edge.2gc.ru/api/
  - service: http_status:404
```

### 2. **Полная конфигурация (все endpoints)**
Используйте файл: `cloudflared-client-configs/cloudflare-origin-client.yml`

### 3. **Конфигурация для разработки**
```yaml
tunnel: <TUNNEL_ID>
credentials-file: /etc/cloudflared/credentials.json

ingress:
  - hostname: dev.yourdomain.com
    service: https://edge.2gc.ru/api/
    originRequest:
      noTLSVerify: true
  - service: http_status:404
```

## Результат

После настройки клиенты смогут подключаться к CloudBridge Relay через:

- **Основной API**: `https://app.yourdomain.com` → `https://edge.2gc.ru/api/`
- **Health Check**: `https://health.yourdomain.com` → `https://edge.2gc.ru/health`
- **P2P API**: `https://p2p.yourdomain.com` → `https://edge.2gc.ru/p2p/`
- **QUIC Transport**: `https://quic.yourdomain.com` → `https://edge.2gc.ru/quic/`

Все подключения проходят через Magic WAN туннель с шифрованием AES-256!

## 📞 Поддержка

При возникновении проблем:

1. Проверьте логи: `/var/log/cloudbridge-monitor.log`
2. Запустите мониторинг: `/usr/local/bin/cloudbridge-monitor.sh`
3. Проверьте статус всех компонентов
4. Обратитесь к документации Cloudflare

**CloudBridge Relay готов к работе как Cloudflare Origin!**

