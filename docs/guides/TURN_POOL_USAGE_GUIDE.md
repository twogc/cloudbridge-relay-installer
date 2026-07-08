# **Руководство по использованию TURN Pool Manager**

## **Обзор**

TURN Pool Manager - это полнофункциональная система управления пулами TURN серверов в CloudBridge Relay. Она обеспечивает:

- **Балансировку нагрузки** между TURN серверами
- **Health checking** и автоматический failover
- **Метрики и мониторинг** через Prometheus
- **Гибкую конфигурацию** пулов и серверов
- **Интеграцию с ICE agent** для NAT traversal

##  **Конфигурация**

### **Основные параметры в `config.json`:**

```json
{
  "turn_pool": {
    "enabled": true,
    "default_pool_name": "primary",
    "allocation_timeout": "10m",
    "health_check_interval": "30s",
    "health_check_timeout": "5s",
    "max_allocations_per_client": 10,
    "load_balancing_strategy": "weighted",
    "weighted_threshold": 0.8,
    "metrics_enabled": true,
    "stats_interval": "60s",
    "latency_history_size": 100,
    "failover_enabled": true,
    "failover_timeout": "30s",
    "max_retries": 3,
    "backoff_multiplier": 2.0
  }
}
```

### **Конфигурация пулов и серверов:**

```json
{
  "turn_pool": {
    "pools": {
      "primary": {
        "name": "primary",
        "region": "ru-central1",
        "servers": [
          {
            "address": "turn1.cloudbridge.ru:3478",
            "username": "cloudbridge",
            "password": "${TURN_PASSWORD}",
            "realm": "cloudbridge.ru",
            "priority": 100,
            "weight": 80,
            "max_connections": 1000,
            "region": "ru-central1",
            "provider": "vkcloud",
            "capabilities": ["udp", "tcp", "tls"],
            "tags": {
              "environment": "production",
              "tier": "primary"
            }
          }
        ]
      }
    }
  }
}
```

## **Использование**

### **1. Инициализация**

```go
// Создание конфигурации
config := &nat.TURNPoolConfig{
    DefaultPoolName:         "primary",
    AllocationTimeout:       10 * time.Minute,
    HealthCheckInterval:     30 * time.Second,
    MaxAllocationsPerClient: 10,
    LoadBalancingStrategy:   "weighted",
    MetricsEnabled:          true,
}

// Создание TURN Pool Manager
manager := nat.NewTURNPoolManager(config, logger)
```

### **2. Создание пулов и добавление серверов**

```go
// Создание пула
err := manager.CreatePool("production")
if err != nil {
    log.Fatal(err)
}

// Добавление сервера в пул
server := &nat.TURNServerInfo{
    TURNServerPool: &nat.TURNServerPool{
        Address:        "turn1.example.com:3478",
        Username:       "user",
        Password:       "pass",
        Realm:          "example.com",
        Priority:       100,
        Weight:         80,
        MaxConnections: 1000,
    },
    Region:   "us-east1",
    Provider: "aws",
    Tags:     map[string]string{"tier": "primary"},
}

err = manager.AddServerToPool("production", server)
```

### **3. Создание TURN allocations**

```go
// Создание allocation с предпочтениями
preferences := &nat.AllocationPreferences{
    PreferredRegion:   "us-east1",
    PreferredProvider: "aws",
    MaxLatency:        50 * time.Millisecond,
    PoolName:          "production",
}

allocation, err := manager.AllocateTURN("client-123", preferences)
if err != nil {
    log.Printf("Failed to allocate TURN: %v", err)
    return
}

log.Printf("Allocated TURN: %s on server %s", 
    allocation.AllocationID, 
    allocation.Server.Address)
```

### **4. Освобождение allocations**

```go
// Освобождение allocation
err := manager.ReleaseTURN(allocation.AllocationID)
if err != nil {
    log.Printf("Failed to release TURN allocation: %v", err)
}
```

### **5. Интеграция с ICE Agent**

```go
// Создание ICE Agent с TURN Pool Manager
iceAgent := nat.NewICEAgentWithTURNPoolManager(
    logger, 
    stunServers, 
    turnPoolManager,
)

// Создание TURN relay через ICE Agent
allocation, err := iceAgent.AllocateTURNRelay("client-123", preferences)
if err != nil {
    log.Printf("Failed to allocate TURN relay: %v", err)
}

// Освобождение через ICE Agent
err = iceAgent.ReleaseTURNRelay(allocation.AllocationID)
```

## **Мониторинг и метрики**

### **Prometheus метрики:**

- `turn_pool_total_servers` - Общее количество TURN серверов
- `turn_pool_healthy_servers` - Количество здоровых серверов
- `turn_pool_total_allocations` - Общее количество allocations
- `turn_pool_active_allocations` - Количество активных allocations
- `turn_pool_allocation_requests_total` - Счетчик запросов на allocation
- `turn_pool_allocation_failures_total` - Счетчик неудачных allocations
- `turn_pool_health_checks_total` - Счетчик health checks
- `turn_pool_allocation_duration_seconds` - Время создания allocation
- `turn_server_connections` - Активные соединения по серверам
- `turn_server_latency_seconds` - Задержка по серверам
- `turn_server_status` - Статус серверов (1=здоровый, 0=нездоровый)

### **Получение статистики:**

```go
// Статистика всех пулов
stats := manager.GetPoolStats()
fmt.Printf("Total pools: %d\n", stats["total_pools"])
fmt.Printf("Total servers: %d\n", stats["total_servers"])
fmt.Printf("Healthy servers: %d\n", stats["healthy_servers"])

// Список активных allocations
allocations := manager.ListAllocations()
for _, alloc := range allocations {
    fmt.Printf("Allocation %s: client %s on %s\n", 
        alloc.AllocationID, 
        alloc.ClientID, 
        alloc.Server.Address)
}
```

## ⚙️ **Стратегии балансировки нагрузки**

### **1. Round Robin (`"round_robin"`):**
- Равномерное распределение между серверами
- Простая и предсказуемая

### **2. Weighted (`"weighted"`):**
- Распределение по весам серверов
- Учитывает приоритет и вес
- **Рекомендуется для production**

### **3. Least Connections (`"least_connections"`):**
- Выбор сервера с наименьшим количеством соединений
- Оптимальная балансировка нагрузки

## 🔥 **Failover и отказоустойчивость**

### **Health Checking:**
- Автоматическая проверка доступности серверов
- Настраиваемые интервалы и таймауты
- Исключение нездоровых серверов из балансировки

### **Automatic Failover:**
- Переключение на резервные серверы при сбоях
- Настраиваемые политики retry
- Exponential backoff для повторных попыток

### **Graceful Degradation:**
- Продолжение работы при частичных сбоях
- Использование резервных пулов
- Логирование и алертинг

## **Безопасность**

### **Аутентификация:**
- Поддержка username/password для TURN серверов
- Безопасное хранение credentials
- Поддержка переменных окружения

### **Ограничения:**
- Лимиты на количество allocations на клиента
- Таймауты для allocations
- Rate limiting (планируется)

## **Лучшие практики**

### **1. Конфигурация:**
- Используйте переменные окружения для паролей
- Настройте health checks с разумными интервалами
- Включите метрики для мониторинга

### **2. Производительность:**
- Используйте weighted балансировку
- Настройте appropriate connection limits
- Мониторьте задержки серверов

### **3. Отказоустойчивость:**
- Настройте несколько пулов в разных регионах
- Используйте серверы разных провайдеров
- Настройте алерты на health check failures

### **4. Мониторинг:**
- Отслеживайте метрики allocation success rate
- Мониторьте latency и connection counts
- Настройте алерты на критические события

##  **Устранение неполадок**

### **Частые проблемы:**

1. **"No healthy TURN servers available"**
   - Проверьте connectivity к TURN серверам
   - Проверьте credentials
   - Увеличьте health check timeout

2. **"Allocation limit exceeded"**
   - Увеличьте `max_allocations_per_client`
   - Проверьте корректное освобождение allocations
   - Мониторьте allocation leaks

3. **Высокая задержка allocations**
   - Проверьте network connectivity
   - Увеличьте connection limits на серверах
   - Рассмотрите добавление серверов

### **Логирование:**
```go
// Включение debug логирования
logger.SetLevel(logger.DEBUG)

// Просмотр статистики
stats := manager.GetPoolStats()
logger.Info("TURN Pool Stats", zap.Any("stats", stats))
```

---

**TURN Pool Manager обеспечивает надежную и масштабируемую систему управления TURN серверами для CloudBridge Relay!**
