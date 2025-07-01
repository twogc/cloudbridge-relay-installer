# TCP Proxy Guide - CloudBridge Relay v2.0

## Обзор

TCP Proxy в CloudBridge Relay v2.0 представляет собой полнофункциональный компонент для двунаправленной передачи данных между клиентами и целевыми сервисами. Этот компонент обеспечивает безопасное проксирование TCP соединений с поддержкой мониторинга, метрик и управления ресурсами.

## Архитектура TCP Proxy

### Основные компоненты

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client App    │    │   TCP Proxy     │    │  Target Service │
│                 │    │                 │    │                 │
│  localhost:12345│◄──►│  Port 12345     │◄──►│  192.168.1.100  │
│                 │    │  (Proxy Port)   │    │      :3389      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Компоненты системы

1. **Proxy Manager** - управляет жизненным циклом прокси
2. **Connection Pool** - пул активных соединений
3. **Buffer Manager** - управление буферами данных
4. **Metrics Collector** - сбор метрик производительности
5. **Port Allocator** - выделение портов для туннелей

## Конфигурация TCP Proxy

### Основные настройки

```json
{
    "proxy": {
        "enabled": true,
        "buffer_size": 8192,
        "timeout": "60s",
        "max_connections": 1000,
        "port_range": {
            "start": 10000,
            "end": 20000
        },
        "metrics_enabled": true,
        "connection_pool_size": 100,
        "idle_timeout": "5m",
        "max_retries": 3,
        "retry_delay": "1s"
    }
}
```

### Параметры конфигурации

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|--------------|----------|
| `enabled` | boolean | true | Включение/выключение TCP proxy |
| `buffer_size` | integer | 8192 | Размер буфера для передачи данных (байт) |
| `timeout` | duration | 60s | Таймаут соединения |
| `max_connections` | integer | 1000 | Максимальное количество одновременных соединений |
| `port_range.start` | integer | 10000 | Начальный порт диапазона |
| `port_range.end` | integer | 20000 | Конечный порт диапазона |
| `metrics_enabled` | boolean | true | Включение сбора метрик |
| `connection_pool_size` | integer | 100 | Размер пула соединений |
| `idle_timeout` | duration | 5m | Таймаут неактивных соединений |
| `max_retries` | integer | 3 | Максимальное количество попыток переподключения |
| `retry_delay` | duration | 1s | Задержка между попытками |

## Использование TCP Proxy

### Создание туннеля

```go
// Создание туннеля через API
tunnelRequest := map[string]interface{}{
    "type": "tunnel_info",
    "tunnel_id": "tunnel_001",
    "local_port": 3389,
    "remote_host": "192.168.1.100",
    "remote_port": 3389,
    "tenant_id": "tenant_001"
}

// Отправка запроса
encoder := json.NewEncoder(conn)
if err := encoder.Encode(tunnelRequest); err != nil {
    log.Fatal("Failed to send tunnel request:", err)
}

// Получение ответа
var response map[string]interface{}
decoder := json.NewDecoder(conn)
if err := decoder.Decode(&response); err != nil {
    log.Fatal("Failed to read response:", err)
}

if response["status"] == "ok" {
    proxyPort := int(response["proxy_port"].(float64))
    fmt.Printf("Tunnel created on proxy port: %d\n", proxyPort)
    
    // Теперь можно подключаться к localhost:proxyPort
    // для доступа к 192.168.1.100:3389
}
```

### Подключение к прокси

```go
// Подключение к прокси порту
proxyConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", proxyPort))
if err != nil {
    log.Fatal("Failed to connect to proxy:", err)
}
defer proxyConn.Close()

// Отправка данных через прокси
data := []byte("Hello, target service!")
_, err = proxyConn.Write(data)
if err != nil {
    log.Printf("Failed to write to proxy: %v", err)
}

// Чтение ответа
buffer := make([]byte, 1024)
n, err := proxyConn.Read(buffer)
if err != nil {
    log.Printf("Failed to read from proxy: %v", err)
} else {
    fmt.Printf("Received: %s\n", string(buffer[:n]))
}
```

## Мониторинг и метрики

### Prometheus метрики

TCP Proxy экспортирует следующие метрики:

```bash
# Основные метрики прокси
relay_proxy_connections_active{tenant_id="tenant_001",tunnel_id="tunnel_001"} 5
relay_proxy_connections_total{tenant_id="tenant_001",tunnel_id="tunnel_001"} 100
relay_proxy_bytes_sent_total{tenant_id="tenant_001",tunnel_id="tunnel_001"} 1048576
relay_proxy_bytes_received_total{tenant_id="tenant_001",tunnel_id="tunnel_001"} 2097152
relay_proxy_errors_total{tenant_id="tenant_001",tunnel_id="tunnel_001",error_type="connection_failed"} 2
relay_proxy_connection_duration_seconds{tenant_id="tenant_001",tunnel_id="tunnel_001"} 30.5

# Метрики производительности
relay_proxy_buffer_usage_bytes{tenant_id="tenant_001",tunnel_id="tunnel_001"} 4096
relay_proxy_throughput_bytes_per_second{tenant_id="tenant_001",tunnel_id="tunnel_001"} 1024
relay_proxy_latency_seconds{tenant_id="tenant_001",tunnel_id="tunnel_001"} 0.001
```

### Grafana дашборд

Пример запросов для Grafana:

```sql
-- Активные соединения по тенантам
sum(relay_proxy_connections_active) by (tenant_id)

-- Трафик по туннелям
rate(relay_proxy_bytes_sent_total[5m]) + rate(relay_proxy_bytes_received_total[5m])

-- Ошибки по типам
sum(rate(relay_proxy_errors_total[5m])) by (error_type)

-- Средняя задержка
avg(relay_proxy_latency_seconds) by (tenant_id)
```

## Управление ресурсами

### Ограничения по тенантам

```yaml
tenants:
  tenant_001:
    limits:
      max_tunnels: 10
      max_connections_per_tunnel: 100
      max_bandwidth_mbps: 100
      max_data_transfer_gb: 1000
```

### Мониторинг использования

```go
// Получение статистики туннеля
tunnelStats := proxy.GetTunnelStats("tunnel_001")
fmt.Printf("Active connections: %d\n", tunnelStats.ActiveConnections)
fmt.Printf("Bytes sent: %d\n", tunnelStats.BytesSent)
fmt.Printf("Bytes received: %d\n", tunnelStats.BytesReceived)
fmt.Printf("Errors: %d\n", tunnelStats.Errors)

// Проверка лимитов
if tunnelStats.ActiveConnections >= tenantLimits.MaxConnections {
    log.Warn("Connection limit reached")
}
```

## Безопасность

### IP-фильтрация

```yaml
tenants:
  tenant_001:
    allowed_ips:
      - "192.168.1.0/24"
      - "10.0.0.0/8"
    blocked_ips:
      - "192.168.1.100"
      - "10.0.0.50"
```

### TLS для прокси соединений

```go
// Настройка TLS для прокси
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS13,
    CipherSuites: []uint16{
        tls.TLS_AES_256_GCM_SHA384,
        tls.TLS_CHACHA20_POLY1305_SHA256,
    },
    InsecureSkipVerify: false,
}

// Применение TLS к прокси соединению
tlsConn := tls.Client(proxyConn, tlsConfig)
```

## Обработка ошибок

### Типичные ошибки

1. **Connection refused** - целевой сервис недоступен
2. **Connection timeout** - превышен таймаут соединения
3. **Port already in use** - порт уже занят
4. **Buffer overflow** - переполнение буфера
5. **Rate limit exceeded** - превышен лимит соединений

### Стратегии восстановления

```go
func handleProxyError(err error, tunnelID string) {
    switch {
    case strings.Contains(err.Error(), "connection refused"):
        // Попытка переподключения с экспоненциальной задержкой
        retryWithBackoff(tunnelID, 5*time.Second)
        
    case strings.Contains(err.Error(), "timeout"):
        // Увеличение таймаута и повторная попытка
        increaseTimeout(tunnelID)
        
    case strings.Contains(err.Error(), "port already in use"):
        // Выделение нового порта
        allocateNewPort(tunnelID)
        
    case strings.Contains(err.Error(), "rate limit"):
        // Ожидание и повторная попытка
        time.Sleep(time.Second)
        retryConnection(tunnelID)
    }
}
```

## Производительность

### Оптимизация буферов

```go
// Настройка размера буфера в зависимости от нагрузки
func optimizeBufferSize(throughput int64) int {
    switch {
    case throughput < 1024*1024: // < 1MB/s
        return 4096
    case throughput < 10*1024*1024: // < 10MB/s
        return 8192
    default:
        return 16384
    }
}
```

### Connection pooling

```go
// Пул соединений для переиспользования
type ConnectionPool struct {
    connections map[string]*ProxyConnection
    maxSize     int
    mutex       sync.RWMutex
}

func (p *ConnectionPool) GetConnection(tunnelID string) (*ProxyConnection, error) {
    p.mutex.RLock()
    if conn, exists := p.connections[tunnelID]; exists && conn.IsHealthy() {
        p.mutex.RUnlock()
        return conn, nil
    }
    p.mutex.RUnlock()
    
    // Создание нового соединения
    return p.createConnection(tunnelID)
}
```

## Примеры использования

### RDP через прокси

```go
func createRDPTunnel(relayHost string, token string, tenantID string) error {
    // Создание туннеля для RDP
    tunnelID := "rdp_tunnel_001"
    proxyPort, err := createTunnel(relayHost, token, tenantID, tunnelID, 3389, "192.168.1.100", 3389)
    if err != nil {
        return err
    }
    
    fmt.Printf("RDP tunnel created on port %d\n", proxyPort)
    fmt.Printf("Connect to localhost:%d to access RDP service\n", proxyPort)
    
    return nil
}
```

### SSH через прокси

```go
func createSSHTunnel(relayHost string, token string, tenantID string) error {
    // Создание туннеля для SSH
    tunnelID := "ssh_tunnel_001"
    proxyPort, err := createTunnel(relayHost, token, tenantID, tunnelID, 22, "192.168.1.101", 22)
    if err != nil {
        return err
    }
    
    fmt.Printf("SSH tunnel created on port %d\n", proxyPort)
    fmt.Printf("Use: ssh -p %d user@localhost\n", proxyPort)
    
    return nil
}
```

### HTTP/HTTPS через прокси

```go
func createHTTPTunnel(relayHost string, token string, tenantID string) error {
    // Создание туннеля для HTTP
    tunnelID := "http_tunnel_001"
    proxyPort, err := createTunnel(relayHost, token, tenantID, tunnelID, 80, "192.168.1.102", 80)
    if err != nil {
        return err
    }
    
    fmt.Printf("HTTP tunnel created on port %d\n", proxyPort)
    fmt.Printf("Access: http://localhost:%d\n", proxyPort)
    
    return nil
}
```

## Troubleshooting

### Диагностика проблем

```bash
# Проверка активных туннелей
curl http://localhost:8081/api/v1/tunnels \
  -H "Authorization: Bearer YOUR_TOKEN"

# Проверка метрик прокси
curl http://localhost:8081/metrics | grep relay_proxy

# Проверка логов прокси
tail -f /var/log/relay/proxy.log

# Проверка сетевых соединений
netstat -tlnp | grep 10000
ss -tlnp | grep 10000
```

### Частые проблемы

1. **Порт недоступен**
   - Проверьте диапазон портов в конфигурации
   - Убедитесь, что порты не заняты другими сервисами

2. **Высокая задержка**
   - Проверьте сетевую связность
   - Увеличьте размер буфера
   - Оптимизируйте настройки TCP

3. **Превышение лимитов**
   - Проверьте лимиты тенанта
   - Очистите неактивные соединения
   - Увеличьте лимиты при необходимости

## API Reference

### Endpoints

- `GET /api/v1/tunnels` - список активных туннелей
- `POST /api/v1/tunnels` - создание нового туннеля
- `DELETE /api/v1/tunnels/{id}` - удаление туннеля
- `GET /api/v1/tunnels/{id}/stats` - статистика туннеля
- `POST /api/v1/tunnels/cleanup` - очистка неактивных туннелей

### Примеры запросов

```bash
# Создание туннеля
curl -X POST http://localhost:8081/api/v1/tunnels \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tunnel_id": "tunnel_001",
    "local_port": 3389,
    "remote_host": "192.168.1.100",
    "remote_port": 3389,
    "tenant_id": "tenant_001"
  }'

# Получение статистики
curl http://localhost:8081/api/v1/tunnels/tunnel_001/stats \
  -H "Authorization: Bearer YOUR_TOKEN"
``` 