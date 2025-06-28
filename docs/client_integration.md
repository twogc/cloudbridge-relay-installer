# Руководство по интеграции с CloudBridge Relay v2.0

## Обзор
Этот документ описывает процесс интеграции вашего приложения с CloudBridge Relay v2.0 для создания безопасных туннелей с поддержкой TCP proxy и multi-tenancy.

## Новые возможности v2.0

### Ключевые улучшения
- **TCP Proxy**: Полнофункциональный TCP прокси для двунаправленной передачи данных
- **Multi-tenancy**: Поддержка мультитенантности с изоляцией ресурсов
- **Enhanced Security**: Улучшенная интеграция с Keycloak и Django fallback
- **Tunnel Manager**: Централизованное управление туннелями
- **Comprehensive Metrics**: Расширенные метрики производительности

## Протокол взаимодействия v2.0

### 1. Установка TLS соединения
```go
// Пример на Go с TLS 1.3
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS13,
    CipherSuites: []uint16{
        tls.TLS_AES_256_GCM_SHA384,
        tls.TLS_CHACHA20_POLY1305_SHA256,
        tls.TLS_AES_128_GCM_SHA256,
    },
    InsecureSkipVerify: false,
}

conn, err := tls.Dial("tcp", "relay-server:8080", tlsConfig)
if err != nil {
    log.Fatal("Failed to connect to relay:", err)
}
defer conn.Close()
```

### 2. Hello handshake
```go
// Отправка hello сообщения
helloMsg := map[string]interface{}{
    "type": "hello",
    "version": "2.0",
    "features": []string{"tls", "heartbeat", "tunnel_info", "multi_tenant", "proxy"},
}

encoder := json.NewEncoder(conn)
if err := encoder.Encode(helloMsg); err != nil {
    log.Fatal("Failed to send hello:", err)
}

// Получение hello ответа
var helloResponse map[string]interface{}
decoder := json.NewDecoder(conn)
if err := decoder.Decode(&helloResponse); err != nil {
    log.Fatal("Failed to read hello response:", err)
}

if helloResponse["version"] != "2.0" {
    log.Fatal("Unsupported relay version")
}
```

### 3. Аутентификация с multi-tenancy
```go
type AuthMessage struct {
    Type      string `json:"type"`
    Token     string `json:"token"`
    TenantID  string `json:"tenant_id"`
}

authMsg := AuthMessage{
    Type:     "auth",
    Token:    "your-jwt-token",
    TenantID: "tenant_001",
}

if err := encoder.Encode(authMsg); err != nil {
    log.Fatal("Failed to send auth:", err)
}

// Получение ответа аутентификации
var authResponse map[string]interface{}
if err := decoder.Decode(&authResponse); err != nil {
    log.Fatal("Failed to read auth response:", err)
}

if authResponse["status"] != "ok" {
    log.Fatal("Authentication failed:", authResponse["error"])
}

// Сохранение лимитов тенанта
tenantLimits := authResponse["limits"].(map[string]interface{})
maxTunnels := int(tenantLimits["max_tunnels"].(float64))
maxConnections := int(tenantLimits["max_connections"].(float64))
```

### 4. Создание туннеля с TCP proxy
```go
type TunnelMessage struct {
    Type       string `json:"type"`
    TunnelID   string `json:"tunnel_id"`
    LocalPort  int    `json:"local_port"`
    RemoteHost string `json:"remote_host"`
    RemotePort int    `json:"remote_port"`
    TenantID   string `json:"tenant_id"`
}

tunnelMsg := TunnelMessage{
    Type:       "tunnel_info",
    TunnelID:   "tunnel_001",
    LocalPort:  3389,
    RemoteHost: "192.168.1.100",
    RemotePort: 3389,
    TenantID:   "tenant_001",
}

if err := encoder.Encode(tunnelMsg); err != nil {
    log.Fatal("Failed to send tunnel info:", err)
}

// Получение ответа о создании туннеля
var tunnelResponse map[string]interface{}
if err := decoder.Decode(&tunnelResponse); err != nil {
    log.Fatal("Failed to read tunnel response:", err)
}

if tunnelResponse["status"] != "ok" {
    log.Fatal("Tunnel creation failed:", tunnelResponse["error"])
}

// Получение порта прокси
proxyPort := int(tunnelResponse["proxy_port"].(float64))
fmt.Printf("Tunnel created, proxy port: %d\n", proxyPort)
```

### 5. Heartbeat с метриками туннелей
```go
type HeartbeatMessage struct {
    Type         string                 `json:"type"`
    TunnelStats  map[string]interface{} `json:"tunnel_stats"`
}

heartbeatMsg := HeartbeatMessage{
    Type: "heartbeat",
    TunnelStats: map[string]interface{}{
        "tunnel_001": map[string]interface{}{
            "bytes_sent":        1024,
            "bytes_received":    2048,
            "active_connections": 2,
        },
    },
}

// Отправка heartbeat каждые 30 секунд
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        if err := encoder.Encode(heartbeatMsg); err != nil {
            log.Printf("Failed to send heartbeat: %v", err)
        }
    }
}
```

## Примеры интеграции v2.0

### Python с поддержкой multi-tenancy
```python
import ssl
import socket
import json
import time
import threading
from typing import Dict, Any

class RelayClientV2:
    def __init__(self, host: str, port: int, tenant_id: str):
        self.host = host
        self.port = port
        self.tenant_id = tenant_id
        self.conn = None
        self.tunnels = {}
        
    def connect(self):
        # Создание SSL контекста
        context = ssl.create_default_context()
        context.minimum_version = ssl.TLSVersion.TLSv1_3
        context.verify_mode = ssl.CERT_REQUIRED
        
        # Установка соединения
        sock = socket.create_connection((self.host, self.port))
        self.conn = context.wrap_socket(sock, server_hostname=self.host)
        
        # Hello handshake
        self._send_hello()
        
    def _send_hello(self):
        hello_msg = {
            "type": "hello",
            "version": "2.0",
            "features": ["tls", "heartbeat", "tunnel_info", "multi_tenant", "proxy"]
        }
        self._send_message(hello_msg)
        
        response = self._receive_message()
        if response.get("version") != "2.0":
            raise Exception("Unsupported relay version")
            
    def authenticate(self, token: str) -> Dict[str, Any]:
        auth_msg = {
            "type": "auth",
            "token": token,
            "tenant_id": self.tenant_id
        }
        self._send_message(auth_msg)
        
        response = self._receive_message()
        if response.get("status") != "ok":
            raise Exception(f"Authentication failed: {response.get('error')}")
            
        return response.get("limits", {})
        
    def create_tunnel(self, tunnel_id: str, local_port: int, 
                     remote_host: str, remote_port: int) -> int:
        tunnel_msg = {
            "type": "tunnel_info",
            "tunnel_id": tunnel_id,
            "local_port": local_port,
            "remote_host": remote_host,
            "remote_port": remote_port,
            "tenant_id": self.tenant_id
        }
        self._send_message(tunnel_msg)
        
        response = self._receive_message()
        if response.get("status") != "ok":
            raise Exception(f"Tunnel creation failed: {response.get('error')}")
            
        proxy_port = response.get("proxy_port")
        self.tunnels[tunnel_id] = {
            "proxy_port": proxy_port,
            "stats": response.get("stats", {})
        }
        
        return proxy_port
        
    def start_heartbeat(self):
        def heartbeat_loop():
            while True:
                try:
                    heartbeat_msg = {
                        "type": "heartbeat",
                        "tunnel_stats": {
                            tunnel_id: {
                                "bytes_sent": 0,
                                "bytes_received": 0,
                                "active_connections": 0
                            }
                            for tunnel_id in self.tunnels
                        }
                    }
                    self._send_message(heartbeat_msg)
                    time.sleep(30)
                except Exception as e:
                    print(f"Heartbeat error: {e}")
                    break
                    
        thread = threading.Thread(target=heartbeat_loop, daemon=True)
        thread.start()
        
    def _send_message(self, msg: Dict[str, Any]):
        data = json.dumps(msg).encode('utf-8')
        self.conn.send(data + b'\n')
        
    def _receive_message(self) -> Dict[str, Any]:
        data = self.conn.recv(4096).decode('utf-8')
        return json.loads(data)

# Использование
try:
    client = RelayClientV2("relay-server", 8080, "tenant_001")
    client.connect()
    
    limits = client.authenticate("your-jwt-token")
    print(f"Tenant limits: {limits}")
    
    proxy_port = client.create_tunnel("tunnel_001", 3389, "192.168.1.100", 3389)
    print(f"Tunnel created on proxy port: {proxy_port}")
    
    client.start_heartbeat()
    
    # Поддержание соединения
    while True:
        time.sleep(1)
        
except Exception as e:
    print(f"Error: {e}")
```

### JavaScript/Node.js с поддержкой TLS
```javascript
const tls = require('tls');
const net = require('net');

class RelayClientV2 {
    constructor(host, port, tenantId) {
        this.host = host;
        this.port = port;
        this.tenantId = tenantId;
        this.conn = null;
        this.tunnels = new Map();
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
                this._sendHello().then(resolve).catch(reject);
            });
            
            this.conn.on('error', reject);
        });
    }
    
    async _sendHello() {
        const helloMsg = {
            type: 'hello',
            version: '2.0',
            features: ['tls', 'heartbeat', 'tunnel_info', 'multi_tenant', 'proxy']
        };
        
        await this.sendMessage(helloMsg);
        const response = await this.receiveMessage();
        
        if (response.version !== '2.0') {
            throw new Error('Unsupported relay version');
        }
    }
    
    async authenticate(token) {
        const authMsg = {
            type: 'auth',
            token: token,
            tenant_id: this.tenantId
        };
        
        await this.sendMessage(authMsg);
        const response = await this.receiveMessage();
        
        if (response.status !== 'ok') {
            throw new Error(`Authentication failed: ${response.error}`);
        }
        
        return response.limits;
    }
    
    async createTunnel(tunnelId, localPort, remoteHost, remotePort) {
        const tunnelMsg = {
            type: 'tunnel_info',
            tunnel_id: tunnelId,
            local_port: localPort,
            remote_host: remoteHost,
            remote_port: remotePort,
            tenant_id: this.tenantId
        };
        
        await this.sendMessage(tunnelMsg);
        const response = await this.receiveMessage();
        
        if (response.status !== 'ok') {
            throw new Error(`Tunnel creation failed: ${response.error}`);
        }
        
        const proxyPort = response.proxy_port;
        this.tunnels.set(tunnelId, {
            proxyPort: proxyPort,
            stats: response.stats || {}
        });
        
        return proxyPort;
    }
    
    startHeartbeat() {
        setInterval(async () => {
            try {
                const heartbeatMsg = {
                    type: 'heartbeat',
                    tunnel_stats: Object.fromEntries(
                        Array.from(this.tunnels.keys()).map(tunnelId => [
                            tunnelId,
                            {
                                bytes_sent: 0,
                                bytes_received: 0,
                                active_connections: 0
                            }
                        ])
                    )
                };
                
                await this.sendMessage(heartbeatMsg);
            } catch (error) {
                console.error('Heartbeat error:', error);
            }
        }, 30000);
    }
    
    async sendMessage(msg) {
        return new Promise((resolve, reject) => {
            this.conn.write(JSON.stringify(msg) + '\n', (err) => {
                if (err) reject(err);
                else resolve();
            });
        });
    }
    
    async receiveMessage() {
        return new Promise((resolve, reject) => {
            let data = '';
            
            const onData = (chunk) => {
                data += chunk.toString();
                if (data.includes('\n')) {
                    this.conn.removeListener('data', onData);
                    try {
                        const message = JSON.parse(data.trim());
                        resolve(message);
                    } catch (error) {
                        reject(error);
                    }
                }
            };
            
            this.conn.on('data', onData);
        });
    }
}

// Использование
async function main() {
    try {
        const client = new RelayClientV2('relay-server', 8080, 'tenant_001');
        await client.connect();
        
        const limits = await client.authenticate('your-jwt-token');
        console.log('Tenant limits:', limits);
        
        const proxyPort = await client.createTunnel('tunnel_001', 3389, '192.168.1.100', 3389);
        console.log('Tunnel created on proxy port:', proxyPort);
        
        client.startHeartbeat();
        
        // Поддержание соединения
        process.on('SIGINT', () => {
            console.log('Shutting down...');
            process.exit(0);
        });
        
    } catch (error) {
        console.error('Error:', error);
    }
}

main();
```

## Обработка ошибок v2.0

### Новые типы ошибок
1. `invalid_token` - неверный токен
2. `rate_limit_exceeded` - превышен лимит запросов
3. `connection_limit_reached` - достигнут лимит соединений
4. `server_unavailable` - сервер недоступен
5. `tenant_not_found` - тенант не найден
6. `tenant_limit_exceeded` - превышен лимит ресурсов тенанта
7. `proxy_error` - ошибка TCP прокси
8. `tunnel_not_found` - туннель не найден

### Улучшенная обработка ошибок
```go
func handleRelayErrorV2(err error, response map[string]interface{}) {
    switch response["code"] {
    case "invalid_token":
        // Запросить новый токен
        refreshToken()
    case "rate_limit_exceeded":
        // Подождать указанное время
        retryAfter := time.Duration(response["retry_after"].(float64)) * time.Second
        time.Sleep(retryAfter)
    case "tenant_limit_exceeded":
        // Очистить неиспользуемые ресурсы тенанта
        cleanupTenantResources(response["tenant_id"].(string))
    case "proxy_error":
        // Пересоздать туннель
        recreateTunnel(response["tunnel_id"].(string))
    case "tunnel_not_found":
        // Проверить статус туннеля и пересоздать при необходимости
        checkAndRecreateTunnel(response["tunnel_id"].(string))
    }
}
```

## Рекомендации по безопасности v2.0

### 1. Хранение токенов и конфигурации
- Используйте безопасное хранилище для JWT токенов
- Шифруйте конфигурационные файлы
- Регулярно ротируйте секреты
- Используйте переменные окружения для чувствительных данных

### 2. Сетевая безопасность
- Принудительно используйте TLS 1.3
- Проверяйте сертификаты сервера
- Используйте certificate pinning для дополнительной безопасности
- Ограничивайте доступ по IP-адресам

### 3. Multi-tenancy безопасность
- Изолируйте ресурсы между тенантами
- Используйте IP-фильтрацию
- Мониторьте использование ресурсов
- Регулярно аудируйте активность тенантов

### 4. Обработка соединений
- Реализуйте graceful shutdown
- Обрабатывайте разрывы соединения
- Используйте connection pooling
- Мониторьте состояние туннелей

### 5. Мониторинг и аудит
- Логируйте все операции
- Собирайте метрики производительности
- Настройте алерты на аномальную активность
- Регулярно проверяйте логи безопасности

## Миграция с v1.x

### Основные изменения
1. **Обязательный TLS 1.3** - все соединения должны использовать TLS 1.3
2. **Hello handshake** - добавлен обязательный hello handshake
3. **Multi-tenancy** - добавлена поддержка тенантов
4. **TCP Proxy** - изменен механизм создания туннелей
5. **Enhanced heartbeat** - добавлены метрики туннелей в heartbeat

### Пошаговая миграция
1. Обновите клиент для поддержки TLS 1.3
2. Добавьте hello handshake
3. Обновите аутентификацию для поддержки тенантов
4. Измените создание туннелей для работы с TCP proxy
5. Обновите heartbeat для отправки метрик
6. Протестируйте все функции
7. Обновите обработку ошибок 