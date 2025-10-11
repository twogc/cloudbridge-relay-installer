# Cloudflare Origin Setup для CloudBridge Relay

## 🎯 Цель
Настроить CloudBridge Relay (`edge.2gc.ru`) как Origin endpoint в Cloudflare через Magic WAN туннель.

## 🔧 Шаги настройки

### 1. В Cloudflare Dashboard

#### 1.1 Создание Tunnel
1. Перейдите в **Zero Trust** → **Access** → **Tunnels**
2. Создайте новый tunnel: `cloudbridge-relay-tunnel`
3. Скопируйте **Tunnel ID** и **Token**

#### 1.2 Настройка Ingress Rules
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

#### 1.3 Настройка Magic WAN
1. Перейдите в **Magic WAN** → **Tunnels**
2. Убедитесь, что туннель к `95.163.250.190` активен
3. Проверьте, что `edge.2gc.ru` доступен через Magic WAN

### 2. На сервере (95.163.250.190)

#### 2.1 Проверка Magic WAN туннеля
```bash
# Проверка статуса туннеля
sudo swanctl --list-sas

# Должно показать:
# magic-wan-primary: #1, ESTABLISHED, IKEv2
```

#### 2.2 Проверка доступности CloudBridge Relay
```bash
# Проверка health endpoint
curl -k https://edge.2gc.ru/health

# Проверка API endpoint
curl -k https://edge.2gc.ru/api/v1/
```

### 3. Настройка nginx для Cloudflare

#### 3.1 Обновление nginx конфигурации
Нужно добавить поддержку Cloudflare IP ranges и правильные заголовки.

#### 3.2 Health Check для Cloudflare
Настроить `/health` endpoint для мониторинга Cloudflare.

## 🔍 Проверка работы

### 1. Тест подключения
```bash
# С клиентской машины
curl -k https://app.client.company.ru/health
```

### 2. Мониторинг логов
```bash
# nginx логи
sudo tail -f /var/log/nginx/access.log

# CloudBridge Relay логи
kubectl logs -n cloudbridge -l app=cloudbridge-relay -f
```

### 3. Проверка Magic WAN
```bash
# Статус туннеля
sudo swanctl --list-sas

# Трафик через туннель
sudo tcpdump -i any esp
```

## 🚨 Troubleshooting

### Проблема: "Connection refused"
- Проверьте Magic WAN туннель: `sudo swanctl --list-sas`
- Проверьте nginx: `sudo systemctl status nginx`
- Проверьте CloudBridge Relay: `kubectl get pods -n cloudbridge`

### Проблема: "DNS resolution failed"
- Проверьте DNS записи в Cloudflare
- Проверьте, что домен указывает на tunnel

### Проблема: "SSL certificate error"
- Проверьте SSL сертификаты: `sudo certbot certificates`
- Проверьте nginx SSL конфигурацию

