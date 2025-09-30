# CloudBridge Relay - Анализ поддоменов

## Текущие поддомены и их назначение:

### 1. **relay.2gc.ru** - Основной релей сервер
- **Назначение**: Главный endpoint для P2P релея
- **Функции**: 
  - Регистрация клиентов
  - P2P соединения
  - Health check
- **API**: `/api/v1/tenants/{tenant_id}/peers/register`

### 2. **masque.2gc.ru** - MASQUE протокол
- **Назначение**: HTTP/3 MASQUE прокси
- **Функции**:
  - HTTP/3 туннелирование
  - Обход блокировок
  - Анонимный доступ
- **API**: `/masque/` endpoints

### 3. **p2p.2gc.ru** - P2P API
- **Назначение**: Специализированный P2P API
- **Функции**:
  - P2P discovery
  - Peer management
  - Connection management
- **API**: `/p2p/` endpoints

### 4. **admin.2gc.ru** - Административная панель
- **Назначение**: Управление и мониторинг
- **Функции**:
  - Статистика соединений
  - Управление пирами
  - Конфигурация
- **API**: `/admin/` endpoints

### 5. **monitor.2gc.ru** - Мониторинг
- **Назначение**: Метрики и мониторинг
- **Функции**:
  - Prometheus метрики
  - Grafana дашборды
  - Алерты
- **API**: `/metrics/` endpoints

### 6. **api.2gc.ru** - Общий API
- **Назначение**: Универсальный API endpoint
- **Функции**:
  - Все API endpoints
  - Документация API
  - Тестирование
- **API**: `/api/` endpoints

## Проблема с текущей конфигурацией:

### ❌ Все поддомены ведут на один порт (5552)
```yaml
# Текущая конфигурация - НЕПРАВИЛЬНО
- host: relay.2gc.ru
  backend:
    port: 5552  # Только P2P API
- host: masque.2gc.ru  
  backend:
    port: 5552  # Должен быть 8443 (MASQUE)
- host: p2p.2gc.ru
  backend:
    port: 5552  # Должен быть 5552 (правильно)
```

## Правильная конфигурация поддоменов:

### 1. **relay.2gc.ru** → порт 5552 (P2P API)
```yaml
- host: relay.2gc.ru
  http:
    paths:
    - path: /api/
      backend:
        service:
          name: cloudbridge-relay-external
          port:
            number: 5552
    - path: /health
      backend:
        service:
          name: cloudbridge-relay-external
          port:
            number: 5551
```

### 2. **masque.2gc.ru** → порт 8443 (MASQUE)
```yaml
- host: masque.2gc.ru
  http:
    paths:
    - path: /
      backend:
        service:
          name: cloudbridge-relay-external
          port:
            number: 8443
```

### 3. **p2p.2gc.ru** → порт 5552 (P2P API)
```yaml
- host: p2p.2gc.ru
  http:
    paths:
    - path: /p2p/
      backend:
        service:
          name: cloudbridge-relay-external
          port:
            number: 5552
```

### 4. **admin.2gc.ru** → порт 5551 (Admin API)
```yaml
- host: admin.2gc.ru
  http:
    paths:
    - path: /admin/
      backend:
        service:
          name: cloudbridge-relay-external
          port:
            number: 5551
```

### 5. **monitor.2gc.ru** → порт 9090 (Metrics)
```yaml
- host: monitor.2gc.ru
  http:
    paths:
    - path: /metrics/
      backend:
        service:
          name: cloudbridge-relay-external
          port:
            number: 9090
```

### 6. **api.2gc.ru** → порт 5552 (General API)
```yaml
- host: api.2gc.ru
  http:
    paths:
    - path: /api/
      backend:
        service:
          name: cloudbridge-relay-external
          port:
            number: 5552
```

## Рекомендуемая архитектура:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   relay.2gc.ru  │───▶│   P2P API       │───▶│   Port 5552     │
│   (Main Relay)  │    │   /api/         │    │   (TCP)         │
└─────────────────┘    └─────────────────┘    └─────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  masque.2gc.ru  │───▶│   MASQUE Proxy  │───▶│   Port 8443     │
│   (HTTP/3)      │    │   /masque/      │    │   (TCP)         │
└─────────────────┘    └─────────────────┘    └─────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│    p2p.2gc.ru   │───▶│   P2P Discovery │───▶│   Port 5552     │
│   (P2P API)     │    │   /p2p/         │    │   (TCP)         │
└─────────────────┘    └─────────────────┘    └─────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  admin.2gc.ru   │───▶│   Admin Panel   │───▶│   Port 5551      │
│   (Management)   │    │   /admin/       │    │   (TCP)          │
└─────────────────┘    └─────────────────┘    └─────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ monitor.2gc.ru  │───▶│   Metrics       │───▶│   Port 9090     │
│   (Monitoring)  │    │   /metrics/     │    │   (UDP)         │
└─────────────────┘    └─────────────────┘    └─────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   api.2gc.ru    │───▶│   General API   │───▶│   Port 5552     │
│   (All APIs)    │    │   /api/         │    │   (TCP)         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Преимущества такой архитектуры:

1. **Разделение функций** - каждый поддомен для своей задачи
2. **Безопасность** - разные порты для разных функций
3. **Масштабируемость** - можно балансировать нагрузку
4. **Мониторинг** - отдельные метрики для каждого сервиса
5. **Управление** - легче настраивать права доступа

## Следующие шаги:

1. **Обновить Ingress** с правильными портами
2. **Добавить API пути** для каждого поддомена
3. **Настроить Nginx** для проксирования API
4. **Протестировать** доступность каждого endpoint
