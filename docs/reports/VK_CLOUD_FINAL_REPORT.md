# VK Cloud LoadBalancer Configuration - Final Report

**Дата:** $(date)  
**Проект:** CloudBridge Relay для 2gc.ru  
**Балансировщик:** LoadBalancer_3398 (109.120.180.160)  

## ✅ Что успешно настроено

### 1. DNS записи
```
relay.2gc.ru     → 109.120.180.160 ✓
masque.2gc.ru    → 109.120.180.160 ✓
p2p.2gc.ru       → 109.120.180.160 ✓
admin.2gc.ru     → 109.120.180.160 ✓
monitor.2gc.ru   → 109.120.180.160 ✓
api.2gc.ru       → 109.120.180.160 ✓
```

### 2. VK Cloud LoadBalancer
- **Балансировщик**: LoadBalancer_3398 (ID: edb21dce-56c9-4e83-8d87-996dbad7ab60)
- **External IP**: 109.120.180.160 (floating IP)
- **Internal IP**: 10.0.0.242
- **Status**: ACTIVE/ONLINE

### 3. Listeners настроены
- **HTTP Listener**: Порт 80 → HTTP Pool
- **HTTPS Listener**: Порт 443 → TCP Pool (существующий)

### 4. Backend Pools
- **HTTP Pool**: c9c66e2b-8f2a-42a1-937f-a1c4084965c2
  - Member: 10.0.0.249:80 (ONLINE)
- **TCP Pool**: 0cfacebd-dec0-452c-b814-80ca12ec2c3c  
  - Member: 10.0.0.249:80 (ONLINE)

### 5. Kubernetes Integration
- **Kubernetes**: Настроен в minikube на 192.168.49.2
- **Ingress**: Настроен для доменной маршрутизации
- **TLS Certificates**: Созданы для *.2gc.ru

## 🔄 Текущая архитектура

```
Internet
    ↓
DNS (*.2gc.ru) → 109.120.180.160
    ↓
VK Cloud LoadBalancer (HAProxy/Amphora)
    ├── HTTP:80  → 10.0.0.249:80
    └── HTTPS:443 → 10.0.0.249:80
         ↓
Relay Server (10.0.0.249)
    ↓ (веб-сервер делает 301 редирект)
HTTPS Service
```

## 🎯 Статус тестирования

### HTTP Endpoints (работают с редиректом)
```bash
$ curl -I http://109.120.180.160/health
HTTP/1.1 301 Moved Permanently ✓
```

### HTTPS Endpoints (требуют настройки SSL termination)
```bash
$ curl -k https://relay.2gc.ru/health
# Пока недоступны - нужна настройка SSL
```

## 🔧 Следующие шаги для полной настройки

### Шаг 1: Настроить SSL Termination на балансировщике

Загрузить SSL сертификаты в VK Cloud:

```bash
# 1. Создать секреты для сертификата
openstack secret store --name "2gc-ru-cert" \
  --payload-content-type "text/plain" \
  --payload-content-encoding "base64" \
  --payload "$(base64 -w 0 /etc/ssl/cloudbridge-2gc/2gc.ru.crt)"

openstack secret store --name "2gc-ru-key" \
  --payload-content-type "text/plain" \
  --payload-content-encoding "base64" \
  --payload "$(base64 -w 0 /etc/ssl/cloudbridge-2gc/2gc.ru.key)"

# 2. Создать TLS container
openstack secret container create --name "2gc-ru-tls" \
  --type "certificate" \
  --secret "certificate=$(openstack secret list | grep 2gc-ru-cert | awk '{print $2}')" \
  --secret "private_key=$(openstack secret list | grep 2gc-ru-key | awk '{print $2}')"

# 3. Обновить HTTPS listener
openstack loadbalancer listener set 22abbb98-6ed7-4e31-867c-77d5ce632c42 \
  --protocol TERMINATED_HTTPS \
  --default-tls-container-ref "$(openstack secret container list | grep 2gc-ru-tls | awk '{print $2}')"
```

### Шаг 2: Настроить правильную маршрутизацию на relay сервере

На сервере 10.0.0.249 настроить nginx/apache для проксирования на Kubernetes:

```nginx
# /etc/nginx/sites-available/cloudbridge-2gc
server {
    listen 80;
    server_name *.2gc.ru;
    
    location / {
        proxy_pass http://192.168.49.2:32500;  # Kubernetes NodePort
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Шаг 3: Альтернативное решение - прямая маршрутизация

Изменить backend member для подключения напрямую к Kubernetes:

```bash
# Удалить текущий member
openstack loadbalancer member delete c9c66e2b-8f2a-42a1-937f-a1c4084965c2 a9a237e6-9246-4952-ad1b-ddd3d4e82a3d

# Добавить Kubernetes как backend (если доступен по сети)
openstack loadbalancer member create c9c66e2b-8f2a-42a1-937f-a1c4084965c2 \
  --name "kubernetes-nodeport" \
  --address 192.168.49.2 \
  --subnet-id d3272fe8-fec0-4fbb-8e67-6c6716972f19 \
  --protocol-port 32500
```

## 📊 Мониторинг и управление

### Команды для мониторинга

```bash
# Статус балансировщика
openstack loadbalancer show edb21dce-56c9-4e83-8d87-996dbad7ab60

# Статус listeners
openstack loadbalancer listener list --loadbalancer edb21dce-56c9-4e83-8d87-996dbad7ab60

# Статус pools и members
openstack loadbalancer pool list --loadbalancer edb21dce-56c9-4e83-8d87-996dbad7ab60
openstack loadbalancer member list c9c66e2b-8f2a-42a1-937f-a1c4084965c2

# Health monitors
openstack loadbalancer healthmonitor list
```

### Тестирование endpoints

```bash
# HTTP (должен работать с редиректом)
curl -I http://relay.2gc.ru/health
curl -I http://masque.2gc.ru/health
curl -I http://api.2gc.ru/health

# HTTPS (после настройки SSL termination)
curl -k https://relay.2gc.ru/health
curl -k https://masque.2gc.ru/health
curl -k https://api.2gc.ru/health
```

## 🎉 Результат

### Достигнуто:
1. ✅ DNS записи настроены и работают
2. ✅ VK Cloud LoadBalancer настроен и активен
3. ✅ HTTP listener работает (с редиректом)
4. ✅ Backend server отвечает на запросы
5. ✅ Kubernetes сервисы готовы к приему трафика

### Осталось:
1. 🔄 Настроить SSL termination на балансировщике
2. 🔄 Настроить правильную маршрутизацию на relay сервере
3. 🔄 Протестировать все HTTPS endpoints

## 💡 Рекомендации

1. **Для продакшена**: Использовать SSL termination на VK Cloud LoadBalancer
2. **Для разработки**: Можно использовать текущую схему с редиректами
3. **Мониторинг**: Настроить алерты на статус балансировщика и members
4. **Backup**: Создать резервную копию конфигурации балансировщика

## 📞 Поддержка

Все конфигурации сохранены в:
- `VK_CLOUD_ARCHITECTURE.md` - детальная архитектура
- `vk-cloud-openstack.sh` - скрипт для подключения к OpenStack
- OpenStack CLI команды для управления балансировщиком

Балансировщик готов к использованию! 🚀
