# 🚀 Инструкции для DevOps клиента - Тестирование P2P маршрутизации

## 📋 Обзор изменений

В этой версии реализованы следующие компоненты:

### ✅ **Реализованные компоненты:**
1. **P2P Relay Manager** (`internal/p2p/relay_manager.go`)
   - Управление P2P соединениями между клиентами
   - Изоляция по tenant_id
   - QUIC packet forwarding
   - IP allocation для tenant'ов

2. **Heartbeat Manager** (`internal/p2p/heartbeat_manager.go`)
   - Поддержание соединений с интервалом 30 секунд
   - Очистка неактивных пиров
   - Мониторинг статуса соединений

3. **QUIC Relay Handler** (`cmd/relay/quic_relay_handler.go`)
   - Обработка входящих QUIC соединений
   - Извлечение peer_id и tenant_id
   - Релеирование QUIC пакетов между пирами

4. **Heartbeat API** (`internal/api/heartbeat_handlers.go`)
   - REST API для heartbeat системы
   - Статистика tenant'ов
   - Мониторинг онлайн пиров

## 🔧 Инструкции по развертыванию

### 1. Сборка нового образа

```bash
# Перейти в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# Собрать образ с новым тегом
docker build -f Dockerfile.p2p -t cloudbridge-relay:p2p-routing-$(date +%Y%m%d-%H%M%S)-$(git rev-parse --short HEAD) .
```

### 2. Загрузка образа в Minikube

```bash
# Загрузить образ в Minikube
minikube image load cloudbridge-relay:p2p-routing-$(date +%Y%m%d-%H%M%S)-$(git rev-parse --short HEAD)

# Проверить, что образ загружен
minikube image ls | grep cloudbridge-relay
```

### 3. Обновление deployment

```bash
# Обновить deployment с новым тегом
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:p2p-routing-$(date +%Y%m%d-%H%M%S)-$(git rev-parse --short HEAD) -n cloudbridge-relay

# Принудительно перезапустить deployment
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge-relay
```

### 4. Проверка развертывания

```bash
# Проверить статус deployment
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge-relay

# Проверить поды
kubectl get pods -n cloudbridge-relay

# Проверить логи
kubectl logs -n cloudbridge-relay deployment/cloudbridge-relay --tail=50
```

## 🧪 Тестирование функциональности

### 1. Проверка запуска P2P компонентов

В логах должны появиться следующие сообщения:
```
"P2P Relay components initialized successfully"
"Heartbeat Manager started"
"QUIC Relay Handler ready"
```

### 2. Тестирование Heartbeat API

```bash
# Проверить health endpoint
curl -s http://localhost:8080/health | jq .

# Тестировать heartbeat endpoint
curl -X POST http://localhost:8080/api/v1/tenants/test-tenant/peers/test-peer/heartbeat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"status": "online", "last_seen": "2025-09-18T21:00:00Z"}'

# Получить статус пира
curl -s http://localhost:8080/api/v1/tenants/test-tenant/peers/test-peer/status | jq .

# Получить статистику tenant'а
curl -s http://localhost:8080/api/v1/tenants/test-tenant/stats | jq .

# Получить список онлайн пиров
curl -s http://localhost:8080/api/v1/tenants/test-tenant/peers/online | jq .
```

### 3. Тестирование P2P маршрутизации

```bash
# Проверить P2P API
curl -s http://localhost:8082/health | jq .

# Зарегистрировать пир
curl -X POST http://localhost:8082/api/v1/tenants/test-tenant/peers/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "peer_id": "test-peer-1",
    "public_key": "test-public-key",
    "endpoint": "192.168.1.100:9090",
    "capabilities": ["quic", "ice"]
  }'

# Обнаружить пиры
curl -s http://localhost:8082/api/v1/tenants/test-tenant/peers/discover | jq .
```

## 🔍 Отладка

### Проверка логов

```bash
# Посмотреть все логи
kubectl logs -n cloudbridge-relay deployment/cloudbridge-relay -f

# Фильтровать логи по P2P компонентам
kubectl logs -n cloudbridge-relay deployment/cloudbridge-relay | grep -E "(P2P|Heartbeat|QUIC|Relay)"

# Проверить ошибки
kubectl logs -n cloudbridge-relay deployment/cloudbridge-relay | grep -i error
```

### Проверка статуса подов

```bash
# Детальная информация о поде
kubectl describe pod -n cloudbridge-relay -l app=cloudbridge-relay

# Проверить события
kubectl get events -n cloudbridge-relay --sort-by='.lastTimestamp'
```

## 📊 Мониторинг

### Проверка метрик

```bash
# Проверить Prometheus метрики
curl -s http://localhost:9090/api/v1/query?query=relay_connections_total

# Проверить health endpoint
curl -s http://localhost:8080/health | jq .
```

### Проверка сетевых соединений

```bash
# Проверить открытые порты
kubectl exec -n cloudbridge-relay deployment/cloudbridge-relay -- netstat -tulpn

# Проверить QUIC соединения
kubectl exec -n cloudbridge-relay deployment/cloudbridge-relay -- ss -ulpn | grep 9090
```

## ⚠️ Важные моменты

1. **JWT токены**: Убедитесь, что JWT токены содержат `tenant_id` для изоляции
2. **Сетевые порты**: Проверьте, что порты 8080, 8082, 9090, 9092 открыты
3. **Логи**: Мониторьте логи на предмет ошибок инициализации P2P компонентов
4. **Ресурсы**: Убедитесь, что у Minikube достаточно ресурсов

## 🚨 Устранение неполадок

### Проблема: P2P компоненты не запускаются

**Решение:**
1. Проверить логи на ошибки компиляции
2. Убедиться, что образ содержит новые изменения
3. Проверить переменные окружения

### Проблема: Heartbeat API не отвечает

**Решение:**
1. Проверить, что порт 8080 открыт
2. Убедиться, что JWT токен валиден
3. Проверить логи на ошибки аутентификации

### Проблема: QUIC соединения не устанавливаются

**Решение:**
1. Проверить, что порт 9090 открыт
2. Убедиться, что STUN сервер работает
3. Проверить сетевую конфигурацию

## 📞 Поддержка

При возникновении проблем:
1. Соберите логи: `kubectl logs -n cloudbridge-relay deployment/cloudbridge-relay > logs.txt`
2. Проверьте статус подов: `kubectl get pods -n cloudbridge-relay -o wide`
3. Создайте issue с описанием проблемы и приложенными логами

---

**Удачного тестирования! 🚀**
