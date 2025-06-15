# Отладка CloudBridge Relay

## Логирование

### Уровни логирования

Сервер поддерживает следующие уровни логирования:
- `DEBUG` - подробная отладочная информация
- `INFO` - информационные сообщения
- `WARNING` - предупреждения
- `ERROR` - ошибки

### Настройка логирования

```bash
# Установка уровня логирования
./relay --log-level DEBUG

# Запись логов в файл
./relay --log-file /var/log/relay.log
```

### Формат логов

Логи записываются в формате JSON:
```json
{
  "timestamp": "2024-03-14T12:00:00Z",
  "level": "INFO",
  "message": "Server started",
  "fields": {
    "port": 8080,
    "tls": true,
    "workers": 8
  }
}
```

## Метрики

### Prometheus метрики

Сервер предоставляет следующие метрики:

- `relay_active_connections` - количество активных соединений
- `relay_rejected_connections` - количество отклоненных соединений
- `relay_connection_duration_seconds` - длительность соединений
- `relay_bytes_transferred` - количество переданных байт
- `relay_errors_total` - количество ошибок
- `relay_latency_seconds` - задержка обработки запросов

### Просмотр метрик

```bash
# Через curl
curl http://localhost:8080/metrics

# Через Prometheus
prometheus --config.file=prometheus/prometheus.yml
```

## Отладка соединений

### Проверка доступности

```bash
# Проверка порта
nc -zv localhost 8080

# Проверка TLS
openssl s_client -connect localhost:8080 -tls1_2
```

### Мониторинг активных соединений

```bash
# Просмотр активных соединений
netstat -an | grep 8080

# Мониторинг в реальном времени
watch -n 1 'netstat -an | grep 8080'
```

## Отладка протокола

### Тестирование протокола

```python
import socket
import struct
import json

def send_message(sock, message):
    data = json.dumps(message).encode()
    length = struct.pack('>I', len(data))
    sock.sendall(length + data)

def read_message(sock):
    length_data = sock.recv(4)
    if len(length_data) != 4:
        raise Exception("Invalid message length")
    length = struct.unpack('>I', length_data)[0]
    data = sock.recv(length)
    return json.loads(data)

# Тестирование
sock = socket.socket()
sock.connect(('localhost', 8080))

# Отправка регистрации
send_message(sock, {
    "type": "register",
    "token": "test-token",
    "version": "1.0.0"
})

# Чтение ответа
response = read_message(sock)
print(response)
```

## Распространенные проблемы

### Таймауты соединений

1. Проверьте настройки таймаута:
   ```bash
   ./relay --timeout 60s
   ```

2. Проверьте сетевые задержки:
   ```bash
   ping localhost
   ```

3. Проверьте логи на наличие ошибок:
   ```bash
   tail -f /var/log/relay.log | grep "timeout"
   ```

### Ошибки валидации токенов

1. Проверьте настройки Django API:
   ```json
   {
     "django": {
       "api_endpoint": "http://localhost:8000/api",
       "jwt_secret": "your-secret-key"
     }
   }
   ```

2. Проверьте токен:
   ```bash
   curl -H "Authorization: Bearer your-token" http://localhost:8000/api/validate
   ```

### Ошибки протокола

1. Проверьте формат сообщений:
   ```python
   # Правильный формат
   {
     "type": "register",
     "token": "your-token",
     "version": "1.0.0"
   }
   ```

2. Проверьте префикс длины:
   ```python
   # 4 байта BigEndian
   length = struct.pack('>I', len(data))
   ```

### Проблемы производительности

1. Проверьте метрики:
   ```bash
   curl http://localhost:8080/metrics | grep relay
   ```

2. Настройте ограничения:
   ```bash
   ./relay --rate-limit 100 --rate-burst 200 --max-tunnels 1000
   ```

3. Мониторинг ресурсов:
   ```bash
   top -p $(pgrep relay)
   ```

## Полезные команды

### Системный мониторинг

```bash
# Мониторинг CPU и памяти
top -p $(pgrep relay)

# Мониторинг сети
iftop -i eth0

# Мониторинг диска
iostat -x 1
```

### Отладка сети

```bash
# Просмотр открытых портов
netstat -tulpn | grep relay

# Мониторинг TCP соединений
tcpdump -i eth0 port 8080

# Проверка маршрутизации
traceroute localhost
```

### Просмотр логов

```bash
# Просмотр всех логов
tail -f /var/log/relay.log

# Фильтрация по уровню
tail -f /var/log/relay.log | grep "ERROR"

# Поиск по ключевому слову
grep "connection" /var/log/relay.log
```

## Автоматическое переподключение

Сервер поддерживает автоматическое переподключение с экспоненциальной задержкой:

```go
config := reconnect.DefaultConfig()
config.MaxRetries = 5
config.RetryDelay = time.Second
config.MaxRetryDelay = time.Minute
config.BackoffFactor = 2.0
```

## Отладка в Docker

```bash
# Просмотр логов контейнера
docker logs relay

# Вход в контейнер
docker exec -it relay /bin/sh

# Проверка сетевых настроек
docker inspect relay | grep IPAddress
```

## Отправка отчетов об ошибках

При возникновении проблем, пожалуйста, предоставьте:

1. Версию сервера
2. Конфигурацию
3. Логи с уровнем DEBUG
4. Метрики Prometheus
5. Описание проблемы
6. Шаги для воспроизведения 