# VK Cloud Architecture Analysis

## Текущая архитектура

### Балансировщик LoadBalancer_3398
- **ID**: `edb21dce-56c9-4e83-8d87-996dbad7ab60`
- **VIP Address**: `10.0.0.242` (внутренний)
- **External IP**: `109.120.180.160` (floating IP)
- **Status**: ACTIVE/ONLINE
- **Provider**: Amphora (HAProxy)

### Listener Configuration
- **Protocol**: TCP
- **Port**: 443
- **Status**: ACTIVE/ONLINE

### Pool Configuration
- **Algorithm**: ROUND_ROBIN
- **Protocol**: TCP
- **Members**: 1 member

### Backend Member
- **Address**: `10.0.0.249` (relay server)
- **Port**: 443
- **Status**: ONLINE
- **Weight**: 10

### Floating IPs
```
109.120.180.160 → 10.0.0.242 (LoadBalancer)
95.163.250.190  → 10.0.0.249 (Relay Server)
```

## Проблема

Текущая конфигурация:
```
Internet → 109.120.180.160:443 → LoadBalancer(10.0.0.242):443 → RelayServer(10.0.0.249):443
```

**Проблема**: Relay сервер (10.0.0.249) слушает на порту 443, но у нас нет HTTPS сервера на этом порту.

## Решение

Нужно изменить конфигурацию балансировщика, чтобы он переадресовывал на правильный порт Kubernetes.

### Вариант 1: Изменить backend port на балансировщике

Изменить member port с 443 на 32500 (NodePort Kubernetes):

```bash
# Обновить member для использования NodePort
openstack loadbalancer member set \
  0cfacebd-dec0-452c-b814-80ca12ec2c3c \
  4fc4bc8f-3b2f-4b30-af3d-025f322777e1 \
  --protocol-port 32500
```

### Вариант 2: Создать новый listener для разных сервисов

Создать дополнительные listeners для разных портов:

```bash
# HTTP listener (порт 80) → NodePort 32500
openstack loadbalancer listener create edb21dce-56c9-4e83-8d87-996dbad7ab60 \
  --name "http-listener" \
  --protocol HTTP \
  --protocol-port 80 \
  --default-pool 0cfacebd-dec0-452c-b814-80ca12ec2c3c

# HTTPS listener (порт 443) → NodePort 32500  
# (уже существует, нужно только изменить backend port)
```

### Вариант 3: Настроить SSL termination на балансировщике

Загрузить SSL сертификат и настроить HTTPS termination:

```bash
# Создать секрет с сертификатом
openstack secret store --name "2gc-ru-cert" \
  --payload-content-type "text/plain" \
  --payload-content-encoding "base64" \
  --payload "$(base64 -w 0 /etc/ssl/cloudbridge-2gc/2gc.ru.crt)"

# Создать секрет с ключом  
openstack secret store --name "2gc-ru-key" \
  --payload-content-type "text/plain" \
  --payload-content-encoding "base64" \
  --payload "$(base64 -w 0 /etc/ssl/cloudbridge-2gc/2gc.ru.key)"

# Создать TLS container
openstack secret container create --name "2gc-ru-tls" \
  --type "certificate" \
  --secret "certificate=$(openstack secret list | grep 2gc-ru-cert | awk '{print $2}')" \
  --secret "private_key=$(openstack secret list | grep 2gc-ru-key | awk '{print $2}')"

# Обновить listener для использования TLS
openstack loadbalancer listener set 22abbb98-6ed7-4e31-867c-77d5ce632c42 \
  --protocol TERMINATED_HTTPS \
  --default-tls-container-ref "$(openstack secret container list | grep 2gc-ru-tls | awk '{print $2}')"
```

## Рекомендуемая архитектура

```
Internet
    ↓
109.120.180.160:443 (External IP)
    ↓
LoadBalancer (10.0.0.242)
    ↓ SSL Termination
RelayServer (10.0.0.249):32500 (HTTP NodePort)
    ↓
Kubernetes Ingress
    ↓ Domain routing
CloudBridge Services
```

## Команды для настройки

### 1. Простое решение - изменить backend port

```bash
# Изменить port member'а на NodePort Kubernetes
openstack loadbalancer member set \
  0cfacebd-dec0-452c-b814-80ca12ec2c3c \
  4fc4bc8f-3b2f-4b30-af3d-025f322777e1 \
  --protocol-port 32500
```

### 2. Проверить NodePort Kubernetes

```bash
# Найти правильный NodePort
kubectl get svc -n cloudbridge cloudbridge-relay-external -o yaml | grep nodePort
```

### 3. Тестирование

```bash
# После изменения конфигурации
curl -k https://relay.2gc.ru/health
curl -k https://masque.2gc.ru/health
curl -k https://api.2gc.ru/health
```

## Health Check

Текущий health monitor:

```bash
openstack loadbalancer healthmonitor show 4cbae8bd-ae34-4de0-b36b-a52edb679f85
```

Возможно, нужно будет обновить health check для работы с HTTP вместо TCP.

## Мониторинг

После настройки можно мониторить:

```bash
# Статус балансировщика
openstack loadbalancer show edb21dce-56c9-4e83-8d87-996dbad7ab60

# Статус members
openstack loadbalancer member list 0cfacebd-dec0-452c-b814-80ca12ec2c3c

# Статус health monitor
openstack loadbalancer healthmonitor show 4cbae8bd-ae34-4de0-b36b-a52edb679f85
```
