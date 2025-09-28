# Настройка внешнего балансировщика для CloudBridge

## Конфигурация

**Внешний IP:** `109.120.180.160`  
**Внутренний IP:** `10.0.0.242`  
**Порт:** `443` (HTTPS только)  

## Существующие DNS записи

```
b1.2gc.space        A       109.120.180.160
masque.2gc.space    CNAME   b1.2gc.space  
relay.2gc.space     CNAME   b1.2gc.space
```

## Рекомендуемые DNS записи для 2gc.ru

Создайте следующие A записи в панели управления DNS:

```dns
relay.2gc.ru     A    109.120.180.160
masque.2gc.ru    A    109.120.180.160
p2p.2gc.ru       A    109.120.180.160
admin.2gc.ru     A    109.120.180.160
monitor.2gc.ru   A    109.120.180.160
api.2gc.ru       A    109.120.180.160
```

## Архитектура маршрутизации

Все сервисы доступны через **порт 443** с маршрутизацией по доменам:

```
Internet (443) → LoadBalancer (109.120.180.160:443) → Kubernetes Ingress → Services
```

### Маршрутизация по доменам:

| Домен | Сервис | Внутренний порт | Описание |
|-------|--------|----------------|----------|
| `relay.2gc.ru` | `cloudbridge-relay-external` | 8443 | Основной relay сервис |
| `masque.2gc.ru` | `cloudbridge-relay-external` | 8443 | MASQUE протокол |
| `p2p.2gc.ru` | `p2p-api-loadbalancer` | 8082 | P2P API |
| `admin.2gc.ru` | `cloudbridge-relay-external` | 9091 | Админ панель |
| `monitor.2gc.ru` | `prometheus` | 9090 | Мониторинг |
| `api.2gc.ru` | `cloudbridge-relay-external` | 9091 | REST API |

### Специальные пути:

| Путь | Сервис | Протокол |
|------|--------|----------|
| `/api/v1/*` | REST API | HTTPS |
| `/grpc/*` | gRPC API | gRPC over HTTPS |
| `/ws/*` | WebSocket | WSS |
| `/.well-known/masque` | MASQUE | HTTPS |

## Настройка балансировщика

### 1. Применить Kubernetes конфигурацию

```bash
# Применить конфигурацию внешнего балансировщика
kubectl apply -f k8s/external-loadbalancer-config.yaml

# Применить HTTPS маршрутизацию
kubectl apply -f k8s/https-routing-config.yaml
```

### 2. Создать TLS сертификаты

```bash
# Создать самоподписанные сертификаты для тестирования
sudo mkdir -p /etc/ssl/cloudbridge-2gc
sudo openssl req -x509 -newkey rsa:2048 -keyout /etc/ssl/cloudbridge-2gc/2gc.ru.key \
    -out /etc/ssl/cloudbridge-2gc/2gc.ru.crt -days 365 -nodes \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=2GC/CN=*.2gc.ru"

# Создать Kubernetes Secret
kubectl create secret tls cloudbridge-2gc-ru-tls \
    --cert=/etc/ssl/cloudbridge-2gc/2gc.ru.crt \
    --key=/etc/ssl/cloudbridge-2gc/2gc.ru.key \
    -n cloudbridge
```

### 3. Настроить Let's Encrypt (рекомендуется для продакшена)

```bash
# Установить cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Создать ClusterIssuer для Let's Encrypt
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@2gc.ru
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

## Проверка работоспособности

### 1. Проверить DNS

```bash
# Проверить A записи
dig +short A relay.2gc.ru
dig +short A masque.2gc.ru
dig +short A api.2gc.ru

# Должны возвращать: 109.120.180.160
```

### 2. Проверить порт 443

```bash
# Проверить доступность порта
telnet 109.120.180.160 443

# Или с помощью curl
curl -k https://relay.2gc.ru/health
```

### 3. Проверить сервисы Kubernetes

```bash
# Проверить LoadBalancer сервисы
kubectl get svc -n cloudbridge -l component=external-loadbalancer

# Проверить Ingress
kubectl get ingress -n cloudbridge

# Проверить TLS сертификаты
kubectl get secrets -n cloudbridge | grep tls
```

## Конфигурация балансировщика (внешнего)

На сервере балансировщика (109.120.180.160) должна быть настроена переадресация:

### Nginx конфигурация:

```nginx
# /etc/nginx/sites-available/cloudbridge-2gc
upstream cloudbridge_backend {
    server 10.0.0.242:443;  # Внутренний IP Kubernetes
}

server {
    listen 443 ssl http2;
    server_name *.2gc.ru;
    
    ssl_certificate /etc/ssl/certs/2gc.ru.crt;
    ssl_certificate_key /etc/ssl/private/2gc.ru.key;
    
    location / {
        proxy_pass https://cloudbridge_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # gRPC support
        grpc_pass grpc://cloudbridge_backend;
    }
}
```

### HAProxy конфигурация:

```haproxy
# /etc/haproxy/haproxy.cfg
global
    daemon
    
defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    
frontend cloudbridge_frontend
    bind *:443 ssl crt /etc/ssl/certs/2gc.ru.pem
    default_backend cloudbridge_backend
    
backend cloudbridge_backend
    balance roundrobin
    server k8s1 10.0.0.242:443 check ssl verify none
```

## Мониторинг

### Health checks:

```bash
# Проверить здоровье всех сервисов
curl -k https://relay.2gc.ru/health
curl -k https://masque.2gc.ru/health  
curl -k https://p2p.2gc.ru/api/v1/health
curl -k https://admin.2gc.ru/health
curl -k https://api.2gc.ru/api/v1/health
```

### Метрики:

```bash
# Prometheus метрики
curl -k https://monitor.2gc.ru/metrics
```

## Безопасность

1. **Используйте валидные SSL сертификаты** от Let's Encrypt или коммерческого CA
2. **Настройте WAF** для защиты от атак
3. **Включите DDoS защиту** на уровне балансировщика
4. **Ограничьте доступ** по IP адресам для админ панели
5. **Настройте мониторинг** и алерты

## Troubleshooting

### Проблема: DNS не резолвится
```bash
# Проверить DNS сервера
nslookup relay.2gc.ru 8.8.8.8
```

### Проблема: Порт 443 недоступен
```bash
# Проверить firewall
sudo ufw status
sudo iptables -L

# Проверить nginx/haproxy
sudo systemctl status nginx
sudo nginx -t
```

### Проблема: SSL сертификат
```bash
# Проверить сертификат
openssl s_client -connect relay.2gc.ru:443 -servername relay.2gc.ru
```
