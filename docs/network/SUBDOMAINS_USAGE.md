# CloudBridge Relay - Использование поддоменов

## Назначение поддоменов

### 1. **relay.2gc.ru** - Основной P2P релей
- **Порт**: 5552 (P2P API)
- **Функции**: 
  - Регистрация клиентов
  - P2P соединения
  - Health check
- **API endpoints**:
  - `POST /api/v1/tenants/{tenant_id}/peers/register`
  - `GET /health`
  - `GET /api/v1/tenants/{tenant_id}/peers`

### 2. **masque.2gc.ru** - HTTP/3 MASQUE прокси
- **Порт**: 8443 (MASQUE)
- **Функции**:
  - HTTP/3 туннелирование
  - Обход блокировок
  - Анонимный доступ
- **API endpoints**:
  - `POST /masque/connect`
  - `GET /masque/status`

### 3. **p2p.2gc.ru** - P2P Discovery API
- **Порт**: 5552 (P2P API)
- **Функции**:
  - P2P discovery
  - Peer management
  - Connection management
- **API endpoints**:
  - `GET /p2p/discover`
  - `POST /p2p/connect`
  - `GET /p2p/peers`

### 4. **admin.2gc.ru** - Административная панель
- **Порт**: 5551 (Admin API)
- **Функции**:
  - Управление пирами
  - Статистика соединений
  - Конфигурация
- **API endpoints**:
  - `GET /admin/stats`
  - `POST /admin/config`
  - `GET /admin/peers`

### 5. **monitor.2gc.ru** - Мониторинг и метрики
- **Порт**: 9090 (Metrics)
- **Функции**:
  - Prometheus метрики
  - Grafana дашборды
  - Алерты
- **API endpoints**:
  - `GET /metrics/`
  - `GET /prometheus/`
  - `GET /grafana/`

### 6. **api.2gc.ru** - Общий API шлюз
- **Порт**: 5552 (General API)
- **Функции**:
  - Все API endpoints
  - Документация API
  - Тестирование
- **API endpoints**:
  - `GET /api/docs`
  - `GET /api/health`
  - `POST /api/v1/...`

##  Конфигурация клиентов

### Server A (config-test-server-a-subdomains.yaml):
```yaml
p2p:
  server_id: "server-a"
  tenant_id: "mesh-network-test"
  relay_url: "https://relay.2gc.ru"
  p2p_url: "https://p2p.2gc.ru"
  admin_url: "https://admin.2gc.ru"
  monitor_url: "https://monitor.2gc.ru"
  api_url: "https://api.2gc.ru"

relay:
  enabled: true
  url: "https://relay.2gc.ru"
  masque_url: "https://masque.2gc.ru"
  p2p_url: "https://p2p.2gc.ru"
```

### Server B (config-test-server-b-subdomains.yaml):
```yaml
p2p:
  server_id: "server-b"
  tenant_id: "mesh-network-test"
  relay_url: "https://relay.2gc.ru"
  p2p_url: "https://p2p.2gc.ru"
  admin_url: "https://admin.2gc.ru"
  monitor_url: "https://monitor.2gc.ru"
  api_url: "https://api.2gc.ru"

relay:
  enabled: true
  url: "https://relay.2gc.ru"
  masque_url: "https://masque.2gc.ru"
  p2p_url: "https://p2p.2gc.ru"
```

## Тестирование

### 1. Проверка доступности API:
```bash
# Health check
curl -I -k https://relay.2gc.ru/health

# API registration
curl -X POST -k https://relay.2gc.ru/api/v1/tenants/test/peers/register \
  -H "Content-Type: application/json" \
  -d '{"peer_id": "test-peer", "public_key": "test-key", "allowed_ips": ["10.100.0.1/32"]}'
```

### 2. Запуск клиентов:
```bash
# Server A
cd /home/ubuntu/cloudbridge-relay-installer/test/relay-client
./cloudbridge-client p2p --config config-test-server-a-subdomains.yaml --log-level debug

# Server B
cd /home/ubuntu/cloudbridge-relay-installer/test/relay-client
./cloudbridge-client p2p --config config-test-server-b-subdomains.yaml --log-level debug
```

## Преимущества поддоменов

### 1. **Разделение функций**
- Каждый поддомен для своей задачи
- Четкое разделение ответственности
- Легче отладка и мониторинг

### 2. **Безопасность**
- Разные порты для разных функций
- Возможность настройки разных правил доступа
- Изоляция сервисов

### 3. **Масштабируемость**
- Можно балансировать нагрузку по поддоменам
- Независимое масштабирование сервисов
- Горизонтальное масштабирование

### 4. **Мониторинг**
- Отдельные метрики для каждого сервиса
- Специализированные дашборды
- Точная диагностика проблем

### 5. **Управление**
- Легче настраивать права доступа
- Простое управление конфигурацией
- Гибкость в развертывании

## Отладка

### Проверка Ingress:
```bash
kubectl get ingress -n cloudbridge
kubectl describe ingress cloudbridge-api-ingress -n cloudbridge
```

### Проверка сервисов:
```bash
kubectl get services -n cloudbridge
kubectl get endpoints -n cloudbridge cloudbridge-relay-external
```

### Проверка логов:
```bash
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=20
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller --tail=20
```

### Тестирование портов:
```bash
# P2P API (порт 5552)
curl -I http://192.168.58.2:32500/health

# Admin API (порт 5551)
curl -I http://192.168.58.2:30081/health

# Metrics (порт 9090)
curl -I http://192.168.58.2:30090/metrics
```

## Следующие шаги

1. **Протестировать клиенты** с новыми конфигурациями
2. **Проверить P2P соединение** между Server A и Server B
3. **Настроить мониторинг** через monitor.2gc.ru
4. **Добавить SSL сертификаты** для поддоменов
5. **Настроить балансировку нагрузки** для высоконагруженных сервисов
