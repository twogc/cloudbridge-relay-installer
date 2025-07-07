# Техническое задание: Разработка серверного клиента для CloudBridge Relay

## 1. Общие сведения

### 1.1 Назначение документа
Данное техническое задание описывает требования к разработке серверного клиента для взаимодействия с CloudBridge Relay с учетом реализованных механизмов безопасности и мониторинга.

### 1.2 Область применения
Серверный клиент предназначен для:
- Установки защищенных соединений с relay-сервером
- Создания и управления туннелями
- Интеграции с системами аутентификации (JWT, Keycloak)
- Соблюдения требований безопасности и rate limiting

### 1.3 Реализованные критические функции
**Версия 2.0** включает следующие критические улучшения:
- **TCP Proxy**: Полнофункциональный TCP прокси для двунаправленной передачи данных
- **Tunnel Manager**: Централизованное управление активными туннелями с автоматической очисткой
- **Multi-tenancy**: Поддержка мультитенантности с лимитами ресурсов и IP-фильтрацией
- **Go 1.21**: Обновление до Go 1.21.13 с современными зависимостями
- **Enhanced Security**: Улучшенная интеграция с Keycloak и Django fallback
- **Comprehensive Testing**: Полный набор unit и integration тестов

## 2. Технические требования

### 2.1 Протокол связи

#### 2.1.1 Транспортный уровень
- **Обязательно:** TLS 1.3 с принудительным использованием
- **Поддерживаемые cipher suites:**
  - `TLS_AES_256_GCM_SHA384`
  - `TLS_CHACHA20_POLY1305_SHA256`
  - `TLS_AES_128_GCM_SHA256`
- **Минимальная версия TLS:** 1.3
- **Проверка сертификатов:** Обязательна

#### 2.1.2 Прикладной уровень
- **Формат сообщений:** JSON
- **Кодировка:** UTF-8
- **Сжатие:** Не используется

### 2.2 Последовательность взаимодействия

#### 2.2.1 Установка соединения
```json
// 1. Установка TLS соединения
// 2. Отправка hello сообщения
{
    "type": "hello",
    "version": "2.0",
    "features": ["tls", "heartbeat", "tunnel_info", "multi_tenant", "proxy"]
}

// 3. Получение hello ответа
{
    "type": "hello_response",
    "version": "2.0",
    "features": ["tls", "heartbeat", "tunnel_info", "multi_tenant", "proxy"]
}
```

#### 2.2.2 Аутентификация
```json
// 4. Отправка аутентификации
{
    "type": "auth",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "tenant_id": "tenant_001"
}

// 5. Получение ответа аутентификации
{
    "type": "auth_response",
    "status": "ok",
    "client_id": "user123",
    "tenant_id": "tenant_001",
    "limits": {
        "max_tunnels": 10,
        "max_connections": 100,
        "rate_limit": 1000
    }
}
```

#### 2.2.3 Создание туннеля
```json
// 6. Отправка информации о туннеле
{
    "type": "tunnel_info",
    "tunnel_id": "tunnel_001",
    "local_port": 3389,
    "remote_host": "192.168.1.100",
    "remote_port": 3389,
    "tenant_id": "tenant_001"
}

// 7. Получение подтверждения
{
    "type": "tunnel_response",
    "status": "ok",
    "tunnel_id": "tunnel_001",
    "proxy_port": 12345,
    "stats": {
        "bytes_sent": 0,
        "bytes_received": 0,
        "connections": 0
    }
}
```

#### 2.2.4 Heartbeat
```json
// 8. Периодическая отправка heartbeat
{
    "type": "heartbeat",
    "tunnel_stats": {
        "tunnel_001": {
            "bytes_sent": 1024,
            "bytes_received": 2048,
            "active_connections": 2
        }
    }
}

// 9. Получение heartbeat ответа
{
    "type": "heartbeat_response",
    "server_time": "2025-01-27T10:30:00Z"
}
```

### 2.3 Обработка ошибок

#### 2.3.1 Коды ошибок
- `invalid_token` - Неверный или истекший JWT токен
- `rate_limit_exceeded` - Превышен лимит запросов для пользователя
- `connection_limit_reached` - Достигнут лимит соединений
- `server_unavailable` - Сервер недоступен
- `invalid_tunnel_info` - Неверная информация о туннеле
- `unknown_message_type` - Неизвестный тип сообщения
- `tenant_not_found` - Тенант не найден
- `tenant_limit_exceeded` - Превышен лимит ресурсов тенанта
- `proxy_error` - Ошибка TCP прокси
- `tunnel_not_found` - Туннель не найден

#### 2.3.2 Формат ошибок
```json
{
    "type": "error",
    "code": "rate_limit_exceeded",
    "message": "Rate limit exceeded for user",
    "retry_after": 30,
    "tenant_id": "tenant_001"
}
```

## 3. Требования безопасности

### 3.1 Аутентификация

#### 3.1.1 JWT токены
- **Алгоритм подписи:** HS256
- **Секрет:** Должен совпадать с настройками relay-сервера
- **Claims:** Обязательно поле `sub` (subject) для идентификации пользователя
- **Валидация:** Проверка подписи, срока действия, issuer

#### 3.1.2 Keycloak интеграция
- **Протокол:** OpenID Connect
- **Токены:** Access tokens и ID tokens
- **JWKS:** Автоматическое обновление ключей
- **Claims:** Проверка ролей и разрешений
- **Fallback:** Django аутентификация при недоступности Keycloak

### 3.2 Multi-tenancy

#### 3.2.1 Тенанты
- **Изоляция:** Полная изоляция ресурсов между тенантами
- **Лимиты:** Настраиваемые лимиты на туннели, соединения, трафик
- **IP фильтрация:** Whitelist/blacklist IP адресов
- **Активность:** Отслеживание активности и автоматическая очистка

#### 3.2.2 Ресурсные лимиты
```json
{
    "tenant_id": "tenant_001",
    "limits": {
        "max_tunnels": 10,
        "max_connections_per_tunnel": 100,
        "max_bandwidth_mbps": 100,
        "max_data_transfer_gb": 1000,
        "allowed_ips": ["192.168.1.0/24", "10.0.0.0/8"],
        "blocked_ips": ["192.168.1.100"]
    }
}
```

### 3.3 Rate Limiting

#### 3.3.1 Ограничения
- **Уровень:** Per-tenant и per-user
- **Лимит по умолчанию:** 100 запросов/сек на пользователя
- **Burst:** 200 запросов
- **Настраиваемость:** Через конфигурацию сервера

#### 3.3.2 Обработка превышения лимита
- **Ожидание:** Exponential backoff
- **Retry:** Максимум 3 попытки
- **Логирование:** Все попытки превышения лимита

### 3.4 TLS требования

#### 3.4.1 Обязательные настройки
- **Версия:** TLS 1.3
- **Cipher suites:** Только TLS 1.3
- **Certificate verification:** Строгая проверка сертификатов
- **SNI:** Поддержка Server Name Indication

#### 3.4.2 Рекомендуемые настройки
- **Certificate pinning:** Для дополнительной безопасности
- **HSTS:** HTTP Strict Transport Security
- **OCSP stapling:** Оптимизация проверки сертификатов

## 4. Требования к реализации

### 4.1 Языки программирования
- **Рекомендуемые:** Go, Python, Node.js, Java
- **Минимальная версия Go:** 1.21+
- **Минимальная версия Python:** 3.8+
- **Минимальная версия Node.js:** 16+

### 4.2 Зависимости

#### 4.2.1 Go (обновлено до 1.21)
```go
import (
    "crypto/tls"
    "encoding/json"
    "net"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/gorilla/mux"
    "github.com/prometheus/client_golang/prometheus"
    "go.uber.org/zap"
)
```

#### 4.2.2 Python
```python
import ssl
import json
import socket
import jwt
from typing import Dict, Any
```

#### 4.2.3 Node.js
```javascript
const tls = require('tls');
const jwt = require('jsonwebtoken');
const net = require('net');
```

### 4.3 Структура клиента

#### 4.3.1 Основные компоненты
1. **ConnectionManager** - Управление соединениями
2. **AuthenticationManager** - Аутентификация и валидация токенов
3. **TunnelManager** - Создание и управление туннелями
4. **ProxyManager** - Управление TCP прокси соединениями
5. **TenantManager** - Управление мультитенантностью
6. **HeartbeatManager** - Поддержание соединения
7. **ErrorHandler** - Обработка ошибок и retry логика

#### 4.3.2 Конфигурация
```json
{
    "relay": {
        "host": "relay.example.com",
        "port": 8080,
        "timeout": "30s",
        "tls": {
            "enabled": true,
            "min_version": "1.3",
            "verify_cert": true,
            "ca_cert": "/path/to/ca.pem"
        }
    },
    "auth": {
        "type": "jwt",
        "secret": "your-jwt-secret",
        "keycloak": {
            "enabled": true,
            "server_url": "https://keycloak.example.com",
            "realm": "cloudbridge",
            "client_id": "relay-client",
            "fallback_to_django": true
        }
    },
    "tenant": {
        "enabled": true,
        "default_limits": {
            "max_tunnels": 5,
            "max_connections": 50,
            "rate_limit": 500
        }
    },
    "proxy": {
        "enabled": true,
        "buffer_size": 8192,
        "timeout": "60s",
        "max_connections": 1000
    },
    "rate_limiting": {
        "enabled": true,
        "max_retries": 3,
        "backoff_multiplier": 2,
        "max_backoff": "30s"
    }
}
```

## 5. Примеры реализации

### 5.1 Go клиент (обновленный)

```go
package main

import (
    "crypto/tls"
    "encoding/json"
    "fmt"
    "net"
    "time"
    "context"
)

type RelayClient struct {
    host         string
    port         int
    tlsConfig    *tls.Config
    conn         net.Conn
    tenantID     string
    tunnelMgr    *TunnelManager
    proxyMgr     *ProxyManager
}

type TunnelManager struct {
    tunnels map[string]*Tunnel
    limits  *TenantLimits
}

type ProxyManager struct {
    proxies map[string]*TCPProxy
    stats   *ProxyStats
}

type TCPProxy struct {
    ID           string
    LocalPort    int
    RemoteHost   string
    RemotePort   int
    Connections  int
    BytesSent    int64
    BytesRecv    int64
}

func NewRelayClient(host string, port int, tenantID string) *RelayClient {
    tlsConfig := &tls.Config{
        MinVersion: tls.VersionTLS13,
        CipherSuites: []uint16{
            tls.TLS_AES_256_GCM_SHA384,
            tls.TLS_CHACHA20_POLY1305_SHA256,
            tls.TLS_AES_128_GCM_SHA256,
        },
        InsecureSkipVerify: false,
    }

    return &RelayClient{
        host:      host,
        port:      port,
        tlsConfig: tlsConfig,
        tenantID:  tenantID,
        tunnelMgr: &TunnelManager{
            tunnels: make(map[string]*Tunnel),
        },
        proxyMgr: &ProxyManager{
            proxies: make(map[string]*TCPProxy),
        },
    }
}

func (c *RelayClient) Connect() error {
    conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port), c.tlsConfig)
    if err != nil {
        return fmt.Errorf("failed to connect: %v", err)
    }
    c.conn = conn
    return nil
}

func (c *RelayClient) Authenticate(token string) error {
    msg := map[string]interface{}{
        "type":      "auth",
        "token":     token,
        "tenant_id": c.tenantID,
    }
    
    return c.sendMessage(msg)
}

func (c *RelayClient) CreateTunnel(tunnelID string, localPort int, remoteHost string, remotePort int) (*TCPProxy, error) {
    msg := map[string]interface{}{
        "type":       "tunnel_info",
        "tunnel_id":  tunnelID,
        "local_port": localPort,
        "remote_host": remoteHost,
        "remote_port": remotePort,
        "tenant_id":   c.tenantID,
    }
    
    if err := c.sendMessage(msg); err != nil {
        return nil, err
    }
    
    // Создаем локальный TCP прокси
    proxy := &TCPProxy{
        ID:         tunnelID,
        LocalPort:  localPort,
        RemoteHost: remoteHost,
        RemotePort: remotePort,
    }
    
    c.proxyMgr.proxies[tunnelID] = proxy
    return proxy, nil
}

func (c *RelayClient) sendMessage(msg map[string]interface{}) error {
    encoder := json.NewEncoder(c.conn)
    return encoder.Encode(msg)
}
```

### 5.2 Python клиент (обновленный)

```python
import ssl
import json
import socket
import time
import threading
from typing import Dict, Any, Optional

class RelayClient:
    def __init__(self, host: str, port: int, tenant_id: str):
        self.host = host
        self.port = port
        self.tenant_id = tenant_id
        self.conn = None
        self.tunnels = {}
        self.proxies = {}
        
    def connect(self):
        context = ssl.create_default_context()
        context.minimum_version = ssl.TLSVersion.TLSv1_3
        context.verify_mode = ssl.CERT_REQUIRED
        
        sock = socket.create_connection((self.host, self.port))
        self.conn = context.wrap_socket(sock, server_hostname=self.host)
        
    def authenticate(self, token: str):
        msg = {
            "type": "auth",
            "token": token,
            "tenant_id": self.tenant_id
        }
        self._send_message(msg)
        
    def create_tunnel(self, tunnel_id: str, local_port: int, 
                     remote_host: str, remote_port: int) -> 'TCPProxy':
        msg = {
            "type": "tunnel_info",
            "tunnel_id": tunnel_id,
            "local_port": local_port,
            "remote_host": remote_host,
            "remote_port": remote_port,
            "tenant_id": self.tenant_id
        }
        self._send_message(msg)
        
        # Создаем локальный TCP прокси
        proxy = TCPProxy(tunnel_id, local_port, remote_host, remote_port)
        self.proxies[tunnel_id] = proxy
        return proxy
        
    def _send_message(self, msg: Dict[str, Any]):
        data = json.dumps(msg).encode('utf-8')
        self.conn.send(data + b'\n')

class TCPProxy:
    def __init__(self, tunnel_id: str, local_port: int, 
                 remote_host: str, remote_port: int):
        self.tunnel_id = tunnel_id
        self.local_port = local_port
        self.remote_host = remote_host
        self.remote_port = remote_port
        self.connections = 0
        self.bytes_sent = 0
        self.bytes_recv = 0
        self.running = False
        
    def start(self):
        self.running = True
        thread = threading.Thread(target=self._proxy_loop)
        thread.daemon = True
        thread.start()
        
    def _proxy_loop(self):
        # Реализация TCP прокси
        pass
```

### 5.3 Node.js клиент (обновленный)

```javascript
const tls = require('tls');
const jwt = require('jsonwebtoken');
const net = require('net');

class RelayClient {
    constructor(host, port, tenantId) {
        this.host = host;
        this.port = port;
        this.tenantId = tenantId;
        this.conn = null;
        this.tunnels = new Map();
        this.proxies = new Map();
    }
    
    async connect() {
        const options = {
            host: this.host,
            port: this.port,
            minVersion: 'TLSv1.3',
            maxVersion: 'TLSv1.3',
            ciphers: 'TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_128_GCM_SHA256',
            rejectUnauthorized: true
        };
        
        return new Promise((resolve, reject) => {
            this.conn = tls.connect(options, () => {
                resolve();
            });
            
            this.conn.on('error', reject);
        });
    }
    
    async authenticate(token) {
        const msg = {
            type: 'auth',
            token: token,
            tenant_id: this.tenantId
        };
        
        await this.sendMessage(msg);
    }
    
    async createTunnel(tunnelId, localPort, remoteHost, remotePort) {
        const msg = {
            type: 'tunnel_info',
            tunnel_id: tunnelId,
            local_port: localPort,
            remote_host: remoteHost,
            remote_port: remotePort,
            tenant_id: this.tenantId
        };
        
        await this.sendMessage(msg);
        
        // Создаем локальный TCP прокси
        const proxy = new TCPProxy(tunnelId, localPort, remoteHost, remotePort);
        this.proxies.set(tunnelId, proxy);
        return proxy;
    }
    
    async sendMessage(msg) {
        return new Promise((resolve, reject) => {
            this.conn.write(JSON.stringify(msg) + '\n', (err) => {
                if (err) reject(err);
                else resolve();
            });
        });
    }
}

class TCPProxy {
    constructor(tunnelId, localPort, remoteHost, remotePort) {
        this.tunnelId = tunnelId;
        this.localPort = localPort;
        this.remoteHost = remoteHost;
        this.remotePort = remotePort;
        this.connections = 0;
        this.bytesSent = 0;
        this.bytesRecv = 0;
        this.running = false;
    }
    
    start() {
        this.running = true;
        // Реализация TCP прокси
    }
}
```

## 6. Требования к тестированию

### 6.1 Unit тесты
- Тестирование аутентификации
- Тестирование создания туннелей
- Тестирование TCP прокси
- Тестирование мультитенантности
- Тестирование обработки ошибок
- Тестирование rate limiting

### 6.2 Integration тесты
- Тестирование полного цикла соединения
- Тестирование TLS 1.3 соединений
- Тестирование с реальным relay-сервером
- Тестирование TCP прокси с реальными сервисами

### 6.3 Security тесты
- Тестирование валидации сертификатов
- Тестирование JWT валидации
- Тестирование rate limiting
- Тестирование изоляции тенантов
- Penetration testing

## 7. Документация

### 7.1 Обязательная документация
- README с примерами использования
- API документация
- Конфигурационный файл
- Troubleshooting guide
- Multi-tenancy guide
- TCP Proxy guide

### 7.2 Рекомендуемая документация
- Architecture overview
- Security considerations
- Performance tuning guide
- Deployment guide
- Monitoring guide

## 8. Критерии приемки

### 8.1 Функциональные требования
- [x] Успешное подключение по TLS 1.3
- [x] Корректная аутентификация с JWT
- [x] Создание и управление туннелями
- [x] TCP прокси функциональность
- [x] Мультитенантность
- [x] Обработка всех типов ошибок
- [x] Поддержка heartbeat
- [x] Интеграция с Keycloak и Django fallback

### 8.2 Нефункциональные требования
- [x] Время подключения < 5 секунд
- [x] Поддержка rate limiting
- [x] Логирование всех операций
- [x] Graceful shutdown
- [x] Обработка сетевых ошибок
- [x] Мониторинг и метрики

### 8.3 Требования безопасности
- [x] Принудительное использование TLS 1.3
- [x] Валидация сертификатов
- [x] Безопасное хранение токенов
- [x] Аудит всех операций
- [x] Защита от timing attacks
- [x] Изоляция тенантов

## 9. Сроки и этапы

### 9.1 Этап 1 (1-2 недели) ✅ ЗАВЕРШЕН
- Базовая реализация клиента
- TLS 1.3 поддержка
- JWT аутентификация

### 9.2 Этап 2 (1 неделя) ✅ ЗАВЕРШЕН
- Rate limiting поддержка
- Обработка ошибок
- Тестирование

### 9.3 Этап 3 (1 неделя) ✅ ЗАВЕРШЕН
- TCP Proxy реализация
- Multi-tenancy
- Keycloak интеграция

### 9.4 Этап 4 (1 неделя) ✅ ЗАВЕРШЕН
- Документация
- Code review
- Security audit
- Go 1.21 обновление

## 10. Контакты

- **Технический контакт:** [Указать контакт]
- **Security контакт:** [Указать контакт]
- **Репозиторий:** [Указать ссылку]
- **Issue tracker:** [Указать ссылку] 