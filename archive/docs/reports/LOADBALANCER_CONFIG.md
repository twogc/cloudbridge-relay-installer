# Конфигурация внешнего балансировщика (109.120.180.160)

## Статус проверки

**DNS записи настроены корректно**
**Порт 443 открыт на балансировщике**  
**Kubernetes сервисы работают**
**TLS сертификаты созданы**
**Балансировщик не настроен для переадресации**

## Текущая архитектура

```
Internet → 109.120.180.160:443 → ??? → 10.0.0.242 → Kubernetes (192.168.49.2)
```

## Требуемая настройка балансировщика

На сервере **109.120.180.160** нужно настроить переадресацию на Kubernetes.

### Вариант 1: Nginx (рекомендуется)

Создать файл `/etc/nginx/sites-available/cloudbridge-2gc`:

```nginx
# Upstream для Kubernetes
upstream kubernetes_backend {
    server 10.0.0.242:32500;  # NodePort сервиса cloudbridge-relay-external
}

# HTTPS сервер
server {
    listen 443 ssl http2;
    server_name *.2gc.ru;
    
    # SSL сертификаты
    ssl_certificate /etc/ssl/certs/2gc.ru.crt;
    ssl_certificate_key /etc/ssl/private/2gc.ru.key;
    
    # SSL настройки
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    
    # Основная локация
    location / {
        proxy_pass http://kubernetes_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket поддержка
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Таймауты
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
    
    # Health check
    location /health {
        proxy_pass http://kubernetes_backend/health;
        proxy_set_header Host $host;
    }
}

# HTTP редирект на HTTPS
server {
    listen 80;
    server_name *.2gc.ru;
    return 301 https://$server_name$request_uri;
}
```

Активировать конфигурацию:
```bash
sudo ln -s /etc/nginx/sites-available/cloudbridge-2gc /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Вариант 2: HAProxy

Создать файл `/etc/haproxy/haproxy.cfg`:

```haproxy
global
    daemon
    maxconn 4096
    
defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    
# Frontend для HTTPS
frontend cloudbridge_https
    bind *:443 ssl crt /etc/ssl/certs/2gc.ru.pem
    default_backend kubernetes_backend
    
# Frontend для HTTP (редирект)
frontend cloudbridge_http
    bind *:80
    redirect scheme https code 301
    
# Backend для Kubernetes
backend kubernetes_backend
    balance roundrobin
    option httpchk GET /health
    server k8s1 10.0.0.242:32500 check
```

### Вариант 3: Простая переадресация (iptables)

```bash
# Переадресация порта 443 на Kubernetes NodePort
sudo iptables -t nat -A PREROUTING -p tcp --dport 443 -j DNAT --to-destination 10.0.0.242:32500
sudo iptables -t nat -A POSTROUTING -j MASQUERADE
```

## Копирование TLS сертификатов

Скопировать сертификаты на балансировщик:

```bash
# С Kubernetes сервера на балансировщик
scp /etc/ssl/cloudbridge-2gc/2gc.ru.crt root@109.120.180.160:/etc/ssl/certs/
scp /etc/ssl/cloudbridge-2gc/2gc.ru.key root@109.120.180.160:/etc/ssl/private/

# Для HAProxy объединить в один файл
cat /etc/ssl/certs/2gc.ru.crt /etc/ssl/private/2gc.ru.key > /etc/ssl/certs/2gc.ru.pem
```

## Проверка после настройки

После настройки балансировщика проверить:

```bash
# Тест доступности
curl -k https://relay.2gc.ru/health
curl -k https://masque.2gc.ru/health
curl -k https://api.2gc.ru/health

# Тест всех доменов
for domain in relay masque p2p admin monitor api; do
    echo "Testing $domain.2gc.ru:"
    curl -k -s -w "HTTP: %{http_code}, Time: %{time_total}s\n" \
         -o /dev/null "https://$domain.2gc.ru/health"
done
```

## Мониторинг балансировщика

### Nginx статус
```bash
# Добавить в конфигурацию Nginx
location /nginx_status {
    stub_status on;
    access_log off;
    allow 127.0.0.1;
    deny all;
}
```

### HAProxy статус
```bash
# Добавить в конфигурацию HAProxy
listen stats
    bind *:8404
    stats enable
    stats uri /stats
    stats refresh 30s
```

## Следующие шаги

1. **Выбрать вариант балансировщика** (рекомендуется Nginx)
2. **Настроить конфигурацию** на сервере 109.120.180.160
3. **Скопировать TLS сертификаты**
4. **Перезапустить сервис** балансировщика
5. **Протестировать все endpoints**

## Текущие рабочие endpoints (для тестирования)

Пока балансировщик не настроен, можно тестировать напрямую:

```bash
# Прямой доступ к Kubernetes
MINIKUBE_IP=$(minikube ip)
curl "http://$MINIKUBE_IP:32500/health"

# Через port-forward
kubectl port-forward -n cloudbridge svc/cloudbridge-relay-external 8080:5552 &
curl "http://localhost:8080/health"
```

## Автоматизация

Создать скрипт для автоматической настройки:

```bash
#!/bin/bash
# setup-nginx-lb.sh

# Установка Nginx
sudo apt update
sudo apt install -y nginx

# Создание конфигурации
sudo tee /etc/nginx/sites-available/cloudbridge-2gc > /dev/null << 'EOF'
# [вставить конфигурацию Nginx]
EOF

# Активация
sudo ln -sf /etc/nginx/sites-available/cloudbridge-2gc /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx

echo "Балансировщик настроен!"
```
