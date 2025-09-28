# Настройка Relay Server (10.0.0.249) для проксирования на Kubernetes

## Подключение к relay серверу

```bash
# SSH подключение к relay серверу
ssh root@10.0.0.249
# или
ssh ubuntu@10.0.0.249
```

## Шаг 1: Установка и настройка Nginx

```bash
# Обновляем систему
sudo apt update

# Устанавливаем nginx если не установлен
sudo apt install -y nginx

# Проверяем статус
sudo systemctl status nginx
```

## Шаг 2: Создание конфигурации nginx

Создайте файл `/etc/nginx/sites-available/cloudbridge-2gc`:

```bash
sudo nano /etc/nginx/sites-available/cloudbridge-2gc
```

Вставьте следующую конфигурацию:

```nginx
# CloudBridge Relay Nginx Configuration for 2gc.ru

# Upstream для Kubernetes
upstream kubernetes_backend {
    server 192.168.49.2:32500;  # NodePort сервиса cloudbridge-relay-external
    keepalive 32;
}

# Основной сервер для всех поддоменов 2gc.ru
server {
    listen 80;
    server_name *.2gc.ru 2gc.ru;
    
    # Логирование
    access_log /var/log/nginx/cloudbridge-2gc-access.log;
    error_log /var/log/nginx/cloudbridge-2gc-error.log;
    
    # Основная локация - проксирование на Kubernetes
    location / {
        proxy_pass http://kubernetes_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Original-URI $request_uri;
        
        # WebSocket поддержка
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Таймауты
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Буферизация
        proxy_buffering on;
        proxy_buffer_size 128k;
        proxy_buffers 4 256k;
        proxy_busy_buffers_size 256k;
    }
    
    # Health check endpoint
    location /health {
        proxy_pass http://kubernetes_backend/health;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Быстрые таймауты для health check
        proxy_connect_timeout 5s;
        proxy_send_timeout 5s;
        proxy_read_timeout 5s;
    }
    
    # API endpoints
    location /api/ {
        proxy_pass http://kubernetes_backend/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Content-Type application/json;
        
        # CORS headers для API
        add_header 'Access-Control-Allow-Origin' '*' always;
        add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS' always;
        add_header 'Access-Control-Allow-Headers' 'Origin, Content-Type, Accept, Authorization, X-Requested-With' always;
        
        if ($request_method = 'OPTIONS') {
            add_header 'Access-Control-Allow-Origin' '*';
            add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS';
            add_header 'Access-Control-Allow-Headers' 'Origin, Content-Type, Accept, Authorization, X-Requested-With';
            add_header 'Access-Control-Max-Age' 1728000;
            add_header 'Content-Type' 'text/plain; charset=utf-8';
            add_header 'Content-Length' 0;
            return 204;
        }
    }
    
    # WebSocket endpoints
    location /ws/ {
        proxy_pass http://kubernetes_backend/ws/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket specific headers
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_cache_bypass $http_upgrade;
        
        # Длинные таймауты для WebSocket
        proxy_connect_timeout 7d;
        proxy_send_timeout 7d;
        proxy_read_timeout 7d;
    }
    
    # Статус nginx
    location /nginx_status {
        stub_status on;
        access_log off;
        allow 127.0.0.1;
        allow 10.0.0.0/8;
        deny all;
    }
}
```

## Шаг 3: Активация конфигурации

```bash
# Создаем резервную копию default конфигурации
sudo cp /etc/nginx/sites-enabled/default /etc/nginx/sites-enabled/default.backup 2>/dev/null || true

# Удаляем default конфигурацию
sudo rm -f /etc/nginx/sites-enabled/default

# Активируем новую конфигурацию
sudo ln -sf /etc/nginx/sites-available/cloudbridge-2gc /etc/nginx/sites-enabled/

# Проверяем конфигурацию
sudo nginx -t

# Перезагружаем nginx
sudo systemctl reload nginx

# Проверяем статус
sudo systemctl status nginx
```

## Шаг 4: Проверка работы

```bash
# Проверяем локально на relay сервере
curl -I http://localhost/health
curl -I http://localhost/nginx_status

# Проверяем доступность Kubernetes backend
curl -I http://192.168.49.2:32500/health
```

## Шаг 5: Мониторинг и логи

```bash
# Просмотр логов
sudo tail -f /var/log/nginx/cloudbridge-2gc-access.log
sudo tail -f /var/log/nginx/cloudbridge-2gc-error.log

# Проверка статуса nginx
sudo systemctl status nginx

# Проверка процессов nginx
ps aux | grep nginx

# Проверка портов
sudo netstat -tlnp | grep :80
```

## Шаг 6: Настройка мониторинга

Создайте скрипт мониторинга:

```bash
sudo nano /usr/local/bin/cloudbridge-monitor.sh
```

Содержимое скрипта:

```bash
#!/bin/bash
# CloudBridge Monitoring Script

# Проверяем nginx
if ! systemctl is-active --quiet nginx; then
    echo "[ERROR] Nginx не запущен"
    exit 1
fi

# Проверяем backend
if ! curl -s --max-time 5 http://192.168.49.2:32500/health > /dev/null; then
    echo "[ERROR] Backend недоступен"
    exit 1
fi

echo "[OK] Все сервисы работают"
```

Сделайте скрипт исполняемым и добавьте в cron:

```bash
sudo chmod +x /usr/local/bin/cloudbridge-monitor.sh

# Тестируем скрипт
sudo /usr/local/bin/cloudbridge-monitor.sh

# Добавляем в cron (каждые 5 минут)
echo '*/5 * * * * /usr/local/bin/cloudbridge-monitor.sh >> /var/log/cloudbridge-monitor.log 2>&1' | sudo crontab -

# Проверяем cron
sudo crontab -l
```

## Troubleshooting

### Проблема: nginx не запускается

```bash
# Проверяем конфигурацию
sudo nginx -t

# Проверяем логи
sudo journalctl -u nginx -f

# Проверяем, не занят ли порт 80
sudo netstat -tlnp | grep :80
```

### Проблема: Backend недоступен

```bash
# Проверяем доступность Kubernetes
ping 192.168.49.2

# Проверяем порт 32500
telnet 192.168.49.2 32500

# Проверяем из Kubernetes
kubectl get svc -n cloudbridge cloudbridge-relay-external
```

### Проблема: 502 Bad Gateway

```bash
# Проверяем логи nginx
sudo tail -f /var/log/nginx/cloudbridge-2gc-error.log

# Проверяем upstream
curl -I http://192.168.49.2:32500/health

# Проверяем DNS
nslookup 192.168.49.2
```

## Результат

После выполнения всех шагов:

1. ✅ Nginx настроен для проксирования на Kubernetes
2. ✅ Все поддомены 2gc.ru маршрутизируются корректно
3. ✅ Логирование и мониторинг настроены
4. ✅ Health checks работают

Теперь трафик будет идти по схеме:
```
Internet → VK Cloud LB (SSL termination) → Relay Server (nginx proxy) → Kubernetes
```
