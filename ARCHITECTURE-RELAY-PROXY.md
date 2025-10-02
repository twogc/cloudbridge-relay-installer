# 🚀 CloudBridge Relay как передатчик - Архитектура

## 🎯 **Правильная архитектура (CloudBridge Relay как передатчик):**

```
[Cloudflared Clients] → [CloudBridge Relay:9090] → [Magic WAN туннель] → [Cloudflare] → [Origin Services]
```

## 📋 **Как это работает:**

### **1. Клиенты cloudflared подключаются к CloudBridge Relay:**
- Клиенты настраивают cloudflared для подключения к `edge.2gc.ru:9090`
- CloudBridge Relay принимает подключения на порту 9090
- nginx проксирует трафик к CloudBridge Relay в подах

### **2. CloudBridge Relay перенаправляет через Magic WAN:**
- CloudBridge Relay обрабатывает запросы от клиентов
- Перенаправляет трафик через Magic WAN туннель к Cloudflare
- Magic WAN обеспечивает безопасное подключение

### **3. Cloudflare получает трафик через Magic WAN:**
- Cloudflare получает трафик через Magic WAN туннель
- Перенаправляет к Origin Services
- Возвращает ответы обратно через Magic WAN

## 🔧 **Настройка:**

### **1. CloudBridge Relay (сервер):**
```nginx
# nginx на порту 9090 принимает cloudflared клиентов
server {
    listen 9090 ssl;
    server_name edge.2gc.ru;
    
    # Проксирование к CloudBridge Relay в подах
    location /api/ {
        proxy_pass http://192.168.58.2:32500/;
    }
    
    location /health {
        proxy_pass http://192.168.58.2:30080/health;
    }
}
```

### **2. Cloudflared клиенты:**
```yaml
# Клиенты подключаются к CloudBridge Relay
ingress:
  - hostname: app.yourdomain.com
    service: https://edge.2gc.ru:9090/api/
  - hostname: health.yourdomain.com
    service: https://edge.2gc.ru:9090/health
```

### **3. Magic WAN туннель:**
```bash
# Туннель между CloudBridge Relay и Cloudflare
sudo swanctl --list-sas
# magic-wan-primary: #1, ESTABLISHED, IKEv2
```

## 🎯 **Преимущества:**

1. **Централизованное управление** - все клиенты подключаются к одному серверу
2. **Безопасность** - Magic WAN туннель обеспечивает шифрование
3. **Масштабируемость** - CloudBridge Relay может обрабатывать множество клиентов
4. **Мониторинг** - централизованный мониторинг всех подключений
5. **Управление трафиком** - контроль и фильтрация трафика

## 📊 **Поток данных:**

```
1. [Cloudflared Client] → [edge.2gc.ru:9090]
2. [nginx] → [CloudBridge Relay в подах]
3. [CloudBridge Relay] → [Magic WAN туннель]
4. [Magic WAN туннель] → [Cloudflare]
5. [Cloudflare] → [Origin Services]
6. [Ответ] ← [Обратный путь]
```

## 🔍 **Проверка работы:**

```bash
# Проверка CloudBridge Relay
curl -k https://edge.2gc.ru:9090/health

# Проверка Magic WAN туннеля
sudo swanctl --list-sas

# Проверка cloudflared клиентов
curl -k https://app.yourdomain.com
```

## 🚀 **Готово к использованию:**

CloudBridge Relay готов принимать cloudflared клиентов на порту 9090 и перенаправлять их трафик через Magic WAN туннель к Cloudflare!

