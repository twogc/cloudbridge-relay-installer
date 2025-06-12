# Руководство по интеграции с CloudBridge Relay

## Обзор
Этот документ описывает процесс интеграции вашего приложения с CloudBridge Relay для создания безопасных туннелей.

## Протокол взаимодействия

### 1. Установка соединения
```go
// Пример на Go
conn, err := net.Dial("tcp", "relay-server:8080")
if err != nil {
    log.Fatal("Failed to connect to relay:", err)
}
defer conn.Close()
```

### 2. Регистрация туннеля
После установки соединения необходимо отправить сообщение регистрации:

```go
type RegisterMessage struct {
    Type  string `json:"type"`
    Token string `json:"token"`
}

msg := RegisterMessage{
    Type:  "register",
    Token: "your-token-from-django",
}

// Отправка сообщения
encoder := json.NewEncoder(conn)
if err := encoder.Encode(msg); err != nil {
    log.Fatal("Failed to send register message:", err)
}

// Получение ответа
var response map[string]interface{}
decoder := json.NewDecoder(conn)
if err := decoder.Decode(&response); err != nil {
    log.Fatal("Failed to read response:", err)
}

if response["status"] != "ok" {
    log.Fatal("Registration failed:", response["error"])
}
```

### 3. Создание туннеля
После успешной регистрации можно создавать туннели:

```go
type ConnectMessage struct {
    Type     string `json:"type"`
    LocalPort int    `json:"local_port"`
}

msg := ConnectMessage{
    Type:      "connect",
    LocalPort: 3389, // для RDP
}

if err := encoder.Encode(msg); err != nil {
    log.Fatal("Failed to send connect message:", err)
}
```

## Примеры интеграции

### Python
```python
import socket
import json

def connect_to_relay(relay_host, relay_port, token):
    # Установка соединения
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect((relay_host, relay_port))
    
    # Регистрация
    register_msg = {
        "type": "register",
        "token": token
    }
    sock.send(json.dumps(register_msg).encode())
    
    # Получение ответа
    response = json.loads(sock.recv(1024).decode())
    if response.get("status") != "ok":
        raise Exception(f"Registration failed: {response.get('error')}")
    
    return sock

# Использование
try:
    relay = connect_to_relay("relay-server", 8080, "your-token")
    # Создание туннеля
    connect_msg = {
        "type": "connect",
        "local_port": 3389
    }
    relay.send(json.dumps(connect_msg).encode())
except Exception as e:
    print(f"Error: {e}")
```

### JavaScript/Node.js
```javascript
const net = require('net');
const JSONStream = require('json-stream');

async function connectToRelay(relayHost, relayPort, token) {
    return new Promise((resolve, reject) => {
        const client = new net.Socket();
        const jsonStream = new JSONStream();
        
        client.connect(relayPort, relayHost, () => {
            // Регистрация
            const registerMsg = {
                type: 'register',
                token: token
            };
            client.write(JSON.stringify(registerMsg));
        });
        
        client.pipe(jsonStream);
        
        jsonStream.on('data', (data) => {
            if (data.status === 'ok') {
                resolve(client);
            } else {
                reject(new Error(`Registration failed: ${data.error}`));
            }
        });
        
        client.on('error', (err) => {
            reject(err);
        });
    });
}

// Использование
async function main() {
    try {
        const relay = await connectToRelay('relay-server', 8080, 'your-token');
        
        // Создание туннеля
        const connectMsg = {
            type: 'connect',
            local_port: 3389
        };
        relay.write(JSON.stringify(connectMsg));
    } catch (err) {
        console.error('Error:', err);
    }
}
```

## Обработка ошибок

### Типичные ошибки
1. `invalid_token` - неверный токен
2. `rate_limit_exceeded` - превышен лимит запросов
3. `connection_limit_reached` - достигнут лимит соединений
4. `server_unavailable` - сервер недоступен

### Рекомендации по обработке
```go
func handleRelayError(err error) {
    switch err.Error() {
    case "invalid_token":
        // Запросить новый токен
    case "rate_limit_exceeded":
        // Подождать и повторить
        time.Sleep(time.Second)
    case "connection_limit_reached":
        // Очистить неиспользуемые соединения
    case "server_unavailable":
        // Попробовать другой сервер из кластера
    }
}
```

## Рекомендации по безопасности

1. **Хранение токенов:**
   - Не храните токены в открытом виде
   - Используйте безопасное хранилище (например, keychain)
   - Регулярно обновляйте токены

2. **Сетевая безопасность:**
   - Используйте TLS для соединения с relay-сервером
   - Проверяйте сертификаты сервера
   - Не передавайте чувствительные данные в открытом виде

3. **Обработка соединений:**
   - Закрывайте неиспользуемые соединения
   - Обрабатывайте разрывы соединения
   - Реализуйте механизм переподключения

## Мониторинг и отладка

### Метрики клиента
Рекомендуется собирать следующие метрики:
- Количество активных туннелей
- Время жизни туннелей
- Количество ошибок
- Задержка соединения

### Логирование
```go
type RelayLogger struct {
    logger *log.Logger
}

func (l *RelayLogger) LogConnection(conn net.Conn) {
    l.logger.Printf("New connection from %s", conn.RemoteAddr())
}

func (l *RelayLogger) LogError(err error) {
    l.logger.Printf("Error: %v", err)
}
```

## Примеры использования

### RDP через туннель
```go
func createRDPTunnel(relayHost string, relayPort int, token string) error {
    // Подключение к relay
    conn, err := connectToRelay(relayHost, relayPort, token)
    if err != nil {
        return err
    }
    
    // Создание туннеля для RDP
    msg := ConnectMessage{
        Type:      "connect",
        LocalPort: 3389,
    }
    
    if err := sendMessage(conn, msg); err != nil {
        return err
    }
    
    // Теперь можно подключаться к localhost:3389
    return nil
}
```

### SSH через туннель
```go
func createSSHTunnel(relayHost string, relayPort int, token string) error {
    conn, err := connectToRelay(relayHost, relayPort, token)
    if err != nil {
        return err
    }
    
    msg := ConnectMessage{
        Type:      "connect",
        LocalPort: 22,
    }
    
    if err := sendMessage(conn, msg); err != nil {
        return err
    }
    
    // Теперь можно подключаться к localhost:22
    return nil
}
```

## Поддержка
При возникновении проблем:
1. Проверьте логи клиента и сервера
2. Убедитесь в правильности токена
3. Проверьте доступность relay-сервера
4. Свяжитесь с поддержкой через t.me/ilanies 