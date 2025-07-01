# Руководство по миграции на CloudBridge Relay v2.0

## Обзор

Это руководство описывает процесс миграции с CloudBridge Relay v1.x на v2.0. Версия 2.0 включает критические улучшения: TCP proxy, multi-tenancy, обновление до Go 1.21, и расширенные возможности безопасности.

## 🚀 Ключевые изменения v2.0

### Новые функции
- **TCP Proxy**: Полнофункциональный TCP прокси для двунаправленной передачи данных
- **Multi-tenancy**: Поддержка мультитенантности с изоляцией ресурсов
- **Tunnel Manager**: Централизованное управление активными туннелями
- **Enhanced Security**: Улучшенная интеграция с Keycloak и Django fallback
- **Go 1.21**: Обновление до Go 1.21.13 с современными зависимостями

### Критические изменения
- Обязательный TLS 1.3 для всех соединений
- Новый протокол с hello handshake
- Поддержка multi-tenancy в аутентификации
- Изменение механизма создания туннелей через TCP proxy
- Расширенные метрики и мониторинг

## 📋 Предварительные требования

### Системные требования
- **Go 1.21+** (обязательно)
- **Linux** (Ubuntu 20.04+, CentOS 8+, RHEL 8+)
- **4+ GB RAM** (рекомендуется 8+ GB для multi-tenancy)
- **20+ GB свободного места**

### Зависимости
```bash
# Обновление Go до 1.21+
wget https://go.dev/dl/go1.21.13.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.13.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version  # Должно показать go version go1.21.13 linux/amd64
```

## 🔄 Пошаговая миграция

### Шаг 1: Подготовка окружения

```bash
# Создание бэкапа текущей конфигурации
sudo cp -r /etc/relay /etc/relay.backup
sudo cp /usr/local/bin/relay /usr/local/bin/relay.backup

# Остановка текущего сервиса
sudo systemctl stop relay
```

### Шаг 2: Обновление кода

```bash
# Обновление репозитория
cd /path/to/cloudbridge-relay-installer
git pull origin main

# Обновление зависимостей
go mod tidy

# Сборка новой версии
go build -o relay cmd/relay/main.go
```

### Шаг 3: Обновление конфигурации

#### Обновление config.json

**Старая конфигурация (v1.x):**
```json
{
    "server": {
        "address": ":8080",
        "port": 8080,
        "timeout": "30s",
        "max_tunnels": 1000
    },
    "django": {
        "api_endpoint": "http://localhost:8000/api",
        "jwt_secret": "your-secret"
    },
    "keycloak": {
        "enabled": false
    },
    "security": {
        "tls_enabled": true,
        "cert_file": "/etc/relay/certs/fullchain.pem",
        "key_file": "/etc/relay/certs/privkey.pem"
    },
    "metrics": {
        "enabled": true,
        "port": ":8081"
    }
}
```

**Новая конфигурация (v2.0):**
```json
{
    "server": {
        "address": ":8080",
        "port": 8080,
        "timeout": "30s",
        "max_tunnels": 1000,
        "ping_interval": "30s",
        "cleanup_interval": "1m"
    },
    "django": {
        "api_endpoint": "http://localhost:8000/api",
        "jwt_secret": "your-secret",
        "jwt_expiry": "24h",
        "rest_timeout": "5s",
        "rest_cache_ttl": "5m",
        "rest_max_retries": 3,
        "rest_retry_delay": "100ms"
    },
    "keycloak": {
        "enabled": true,
        "server_url": "https://keycloak.example.com",
        "realm": "cloudbridge",
        "client_id": "relay-client",
        "client_secret": "your-client-secret",
        "public_key": "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
        "token_validation": {
            "issuer": "https://keycloak.example.com/realms/cloudbridge",
            "audience": "relay-client",
            "clock_skew": "30s"
        },
        "jwks_cache_ttl": "1h",
        "http_timeout": "30s",
        "fallback_to_django": true
    },
    "tenant": {
        "enabled": true,
        "default_limits": {
            "max_tunnels": 10,
            "max_connections_per_tunnel": 100,
            "max_bandwidth_mbps": 100,
            "max_data_transfer_gb": 1000,
            "rate_limit_per_second": 1000
        },
        "cleanup_interval": "5m",
        "inactivity_timeout": "1h"
    },
    "proxy": {
        "enabled": true,
        "buffer_size": 8192,
        "timeout": "60s",
        "max_connections": 1000,
        "port_range": {
            "start": 10000,
            "end": 20000
        },
        "metrics_enabled": true
    },
    "tunnel": {
        "manager_enabled": true,
        "max_tunnels_per_tenant": 50,
        "cleanup_interval": "2m",
        "heartbeat_timeout": "5m"
    },
    "rate_limiting": {
        "enabled": true,
        "requests_per_second": 100,
        "burst": 200,
        "per_tenant": true
    },
    "logging": {
        "level": "info",
        "format": "json",
        "output": "/var/log/relay/relay.log",
        "max_size": "100MB",
        "max_age": "30d",
        "max_backups": 10
    },
    "metrics": {
        "enabled": true,
        "address": ":8081",
        "path": "/metrics",
        "collect_interval": "15s"
    },
    "tls": {
        "enabled": true,
        "cert_file": "/etc/relay/certs/server.crt",
        "key_file": "/etc/relay/certs/server.key",
        "ca_file": "/etc/relay/certs/ca.crt",
        "min_version": "1.3",
        "cipher_suites": [
            "TLS_AES_256_GCM_SHA384",
            "TLS_CHACHA20_POLY1305_SHA256",
            "TLS_AES_128_GCM_SHA256"
        ]
    }
}
```

#### Создание конфигурации multi-tenancy

```bash
# Создание файла конфигурации тенантов
sudo tee /etc/relay/tenants.yaml << EOF
tenants:
  tenant_001:
    name: "Default Tenant"
    enabled: true
    limits:
      max_tunnels: 10
      max_connections_per_tunnel: 100
      max_bandwidth_mbps: 100
      max_data_transfer_gb: 1000
      rate_limit_per_second: 1000
    allowed_ips:
      - "0.0.0.0/0"  # Разрешить все IP для миграции
    metadata:
      contact: "admin@company.com"
      environment: "production"
EOF
```

### Шаг 4: Обновление сертификатов

```bash
# Обновление TLS сертификатов для v2.0
sudo mkdir -p /etc/relay/certs
sudo cp /etc/relay/certs/fullchain.pem /etc/relay/certs/server.crt
sudo cp /etc/relay/certs/privkey.pem /etc/relay/certs/server.key

# Создание CA сертификата (если используется mTLS)
sudo openssl genrsa -out /etc/relay/certs/ca.key 4096
sudo openssl req -new -x509 -days 365 -key /etc/relay/certs/ca.key -out /etc/relay/certs/ca.crt
```

### Шаг 5: Обновление firewall

```bash
# Открытие портов для TCP proxy
sudo ufw allow 10000:20000/tcp

# Или для iptables
sudo iptables -A INPUT -p tcp --dport 10000:20000 -j ACCEPT
```

### Шаг 6: Установка новой версии

```bash
# Установка нового бинарного файла
sudo cp relay /usr/local/bin/relay
sudo chmod +x /usr/local/bin/relay

# Обновление systemd сервиса
sudo systemctl daemon-reload
```

### Шаг 7: Тестирование

```bash
# Проверка конфигурации
sudo /usr/local/bin/relay --config /etc/relay/config.json --validate

# Запуск сервиса
sudo systemctl start relay
sudo systemctl status relay

# Проверка логов
sudo journalctl -u relay -f

# Проверка метрик
curl http://localhost:8081/metrics

# Проверка API
curl http://localhost:8081/health
```

## 🔧 Обновление клиентов

### Server Client

#### Обновление Go клиента

```go
// Старый код (v1.x)
type AuthMessage struct {
    Type  string `json:"type"`
    Token string `json:"token"`
}

// Новый код (v2.0)
type AuthMessage struct {
    Type      string `json:"type"`
    Token     string `json:"token"`
    TenantID  string `json:"tenant_id"`
}

// Добавление hello handshake
func (c *RelayClient) Connect() error {
    // Установка TLS соединения
    conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", c.host, c.port), c.tlsConfig)
    if err != nil {
        return err
    }
    
    // Hello handshake
    helloMsg := map[string]interface{}{
        "type": "hello",
        "version": "2.0",
        "features": []string{"tls", "heartbeat", "tunnel_info", "multi_tenant", "proxy"},
    }
    
    encoder := json.NewEncoder(conn)
    if err := encoder.Encode(helloMsg); err != nil {
        return err
    }
    
    var response map[string]interface{}
    decoder := json.NewDecoder(conn)
    if err := decoder.Decode(&response); err != nil {
        return err
    }
    
    if response["version"] != "2.0" {
        return fmt.Errorf("unsupported relay version")
    }
    
    return nil
}

// Обновление создания туннелей
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
    
    // Создание локального TCP прокси
    proxy := &TCPProxy{
        ID:         tunnelID,
        LocalPort:  localPort,
        RemoteHost: remoteHost,
        RemotePort: remotePort,
    }
    
    return proxy, nil
}
```

#### Обновление Python клиента

```python
# Старый код (v1.x)
def authenticate(self, token: str):
    msg = {
        "type": "auth",
        "token": token
    }
    self._send_message(msg)

# Новый код (v2.0)
def authenticate(self, token: str) -> Dict[str, Any]:
    msg = {
        "type": "auth",
        "token": token,
        "tenant_id": self.tenant_id
    }
    self._send_message(msg)
    
    response = self._receive_message()
    if response.get("status") != "ok":
        raise Exception(f"Authentication failed: {response.get('error')}")
        
    return response.get("limits", {})

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
    
    response = self._receive_message()
    if response.get("status") != "ok":
        raise Exception(f"Tunnel creation failed: {response.get('error')}")
        
    proxy_port = response.get("proxy_port")
    proxy = TCPProxy(tunnel_id, local_port, remote_host, remote_port)
    self.proxies[tunnel_id] = proxy
    return proxy
```

### Desktop Client

#### Обновление JavaScript клиента

```javascript
// Старый код (v1.x)
async authenticate(token) {
    const msg = {
        type: 'auth',
        token: token
    };
    await this.sendMessage(msg);
}

// Новый код (v2.0)
async authenticate(token) {
    const msg = {
        type: 'auth',
        token: token,
        tenant_id: this.tenantId
    };
    
    await this.sendMessage(msg);
    const response = await this.receiveMessage();
    
    if (response.status !== 'ok') {
        throw new Error(`Authentication failed: ${response.error}`);
    }
    
    return response.limits;
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
    const response = await this.receiveMessage();
    
    if (response.status !== 'ok') {
        throw new Error(`Tunnel creation failed: ${response.error}`);
    }
    
    const proxyPort = response.proxy_port;
    const proxy = new TCPProxy(tunnelId, localPort, remoteHost, remotePort);
    this.proxies.set(tunnelId, proxy);
    return proxy;
}
```

## 📊 Мониторинг миграции

### Проверка метрик

```bash
# Проверка новых метрик v2.0
curl http://localhost:8081/metrics | grep -E "(relay_proxy|relay_tenant|relay_tunnel)"

# Ожидаемые метрики:
# relay_proxy_connections_active
# relay_tenant_tunnels_active
# relay_tunnel_manager_tunnels_total
```

### Проверка логов

```bash
# Проверка логов на ошибки
sudo journalctl -u relay | grep -i error

# Проверка успешной инициализации
sudo journalctl -u relay | grep -i "tcp proxy\|multi.tenant\|tunnel manager"
```

## 🔄 Rollback план

### В случае проблем

```bash
# Остановка новой версии
sudo systemctl stop relay

# Восстановление старой версии
sudo cp /usr/local/bin/relay.backup /usr/local/bin/relay
sudo cp -r /etc/relay.backup/* /etc/relay/

# Запуск старой версии
sudo systemctl start relay
sudo systemctl status relay
```

### Проверка совместимости

```bash
# Проверка конфигурации старой версии
sudo /usr/local/bin/relay.backup --config /etc/relay.backup/config.json --validate

# Тестирование соединений
curl http://localhost:8081/health
```

## 🧪 Тестирование

### Unit тесты

```bash
# Запуск всех тестов
go test ./...

# Запуск тестов с покрытием
go test -cover ./...

# Запуск интеграционных тестов
go test ./tests/integration/...
```

### Функциональные тесты

```bash
# Тестирование TCP proxy
go test ./internal/tunnel/...

# Тестирование multi-tenancy
go test ./internal/tenant/...

# Тестирование tunnel manager
go test ./internal/tunnel/manager_test.go
```

## 📈 Производительность

### Бенчмарки

```bash
# Тестирование производительности TCP proxy
go test -bench=. ./internal/tunnel/

# Тестирование multi-tenancy
go test -bench=. ./internal/tenant/
```

### Мониторинг производительности

```bash
# Проверка использования ресурсов
top -p $(pgrep relay)

# Проверка сетевых соединений
netstat -tlnp | grep relay

# Проверка метрик производительности
curl http://localhost:8081/metrics | grep -E "(relay_proxy_throughput|relay_tenant_performance)"
```

## 🔒 Безопасность

### Проверка безопасности

```bash
# Проверка TLS конфигурации
openssl s_client -connect localhost:8080 -tls1_3

# Проверка сертификатов
openssl x509 -in /etc/relay/certs/server.crt -text -noout

# Проверка прав доступа
ls -la /etc/relay/certs/
```

### Аудит безопасности

```bash
# Проверка логов безопасности
sudo journalctl -u relay | grep -i "security\|auth\|error"

# Проверка метрик безопасности
curl http://localhost:8081/metrics | grep -E "(relay_security|relay_auth)"
```

## 📞 Поддержка

### Полезные команды

```bash
# Полная диагностика
sudo relay --diagnostic

# Проверка конфигурации
sudo relay --config /etc/relay/config.json --validate

# Проверка статуса сервиса
sudo systemctl status relay

# Просмотр логов в реальном времени
sudo journalctl -u relay -f
```

### Контакты для поддержки

- **GitHub Issues**: https://github.com/maxlanies/cloudbridge-relay-installer/issues
- **Документация**: https://github.com/maxlanies/cloudbridge-relay-installer/docs
- **Telegram**: t.me/ilanies

## ✅ Чек-лист миграции

- [ ] Обновлен Go до версии 1.21+
- [ ] Создан бэкап текущей конфигурации
- [ ] Обновлен config.json для v2.0
- [ ] Создана конфигурация multi-tenancy
- [ ] Обновлены TLS сертификаты
- [ ] Открыты порты для TCP proxy (10000-20000)
- [ ] Установлена новая версия relay
- [ ] Протестирована конфигурация
- [ ] Запущен сервис и проверен статус
- [ ] Обновлены клиенты для поддержки v2.0
- [ ] Протестированы все функции
- [ ] Проверены метрики и логи
- [ ] Настроен мониторинг
- [ ] Проведено тестирование производительности
- [ ] Проверена безопасность
- [ ] Документированы изменения 