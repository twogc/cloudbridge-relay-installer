# Отладка и диагностика CloudBridge Relay

## Содержание

1. [Логирование](#логирование)
2. [Метрики](#метрики)
3. [Диагностика проблем](#диагностика-проблем)
4. [Профилирование](#профилирование)
5. [Устранение неполадок](#устранение-неполадок)

## Логирование

### Структурированное логирование

CloudBridge Relay использует структурированное логирование на основе [zap](https://github.com/uber-go/zap) с поддержкой JSON формата и автоматической ротации.

#### Конфигурация логгера

```json
{
  "logging": {
    "level": "INFO",
    "format": "json",
    "output": "/var/log/relay/relay.log",
    "max_size": 100,
    "max_age": 30,
    "max_backups": 10,
    "compress": true
  }
}
```

#### Уровни логирования

- **DEBUG** - детальная отладочная информация
- **INFO** - общая информация о работе системы
- **WARN** - предупреждения
- **ERROR** - ошибки

#### Форматы логов

**JSON формат (по умолчанию):**
```json
{
  "level": "INFO",
  "timestamp": "2024-03-14T12:00:00Z",
  "caller": "main.go:123",
  "msg": "Server started",
  "port": 8080,
  "tls": true,
  "workers": 4,
  "user_id": "user123",
  "request_id": "req456",
  "duration": "15ms"
}
```

**Console формат:**
```
2024-03-14T12:00:00Z INFO Server started port=8080 tls=true workers=4
```

#### Специализированные методы логирования

```go
// Логирование соединений
logger.LogConnection(connID, userID, remoteAddr, success)

// Логирование аутентификации
logger.LogAuthentication(userID, method, success, duration)

// Логирование валидации токенов
logger.LogTokenValidation(tokenID, method, success, duration, claims)

// Логирование проверки разрешений
logger.LogPermissionCheck(userID, resource, action, granted, duration)

// Логирование производительности
logger.LogPerformance(operation, duration, metadata)

// Логирование безопасности
logger.LogSecurity(event, userID, remoteAddr, details)

// Логирование бизнес-событий
logger.LogBusinessEvent(event, userID, data)
```

#### Ротация логов

Логи автоматически ротируются с помощью [lumberjack](https://github.com/natefinch/lumberjack):

- **Максимальный размер файла:** 100MB
- **Количество файлов:** 10
- **Срок хранения:** 30 дней
- **Сжатие:** включено

### Просмотр логов

```bash
# Просмотр логов в реальном времени
tail -f /var/log/relay/relay.log

# Просмотр логов с фильтрацией по уровню
tail -f /var/log/relay/relay.log | jq 'select(.level == "ERROR")'

# Поиск ошибок
grep -i error /var/log/relay/relay.log

# Поиск по конкретному пользователю
grep "user123" /var/log/relay/relay.log

# Анализ логов за последний час
journalctl -u relay --since "1 hour ago" | jq .

# Статистика логов
cat /var/log/relay/relay.log | jq -r '.level' | sort | uniq -c
```

## Метрики

### Prometheus метрики

Сервер экспортирует метрики на порту 8081:

```bash
# Получение всех метрик
curl http://localhost:8081/metrics

# Поиск конкретных метрик
curl -s http://localhost:8081/metrics | grep relay_

# Проверка метрик через Prometheus
curl -s "http://localhost:9090/api/v1/query?query=relay_active_connections"
```

### Основные метрики для мониторинга

```promql
# Активные соединения
relay_active_connections

# Общее количество соединений
relay_connections_total

# Количество ошибок
relay_errors_total

# Время соединения (95-й перцентиль)
histogram_quantile(0.95, rate(relay_connection_duration_seconds_bucket[5m]))

# Процент успешных валидаций токенов
rate(relay_keycloak_token_validations_total{result="success"}[5m]) / 
rate(relay_keycloak_token_validations_total[5m]) * 100
```

### Мониторинг производительности

```bash
# Проверка использования ресурсов
top -p $(pgrep relay)

# Проверка сетевых соединений
netstat -tlnp | grep relay

# Проверка открытых файлов
lsof -p $(pgrep relay)

# Проверка памяти
cat /proc/$(pgrep relay)/status | grep -E "(VmRSS|VmSize)"
```

## Диагностика проблем

### Проверка статуса сервиса

```bash
# Статус systemd сервиса
sudo systemctl status relay

# Логи systemd
sudo journalctl -u relay -f

# Проверка конфигурации
./relay --config config/config.json --help

# Валидация конфигурации
./relay --config config/config.json --validate
```

### Проверка сетевых соединений

```bash
# Проверка портов
netstat -tlnp | grep :8080
netstat -tlnp | grep :8081

# Тестирование подключения
telnet localhost 8080
telnet localhost 8081

# Проверка TLS сертификатов
openssl s_client -connect localhost:8080 -servername your-domain.com
```

### Проверка файловой системы

```bash
# Проверка прав доступа
ls -la /etc/relay/
ls -la /var/log/relay/

# Проверка места на диске
df -h /var/log/relay/

# Проверка inode
df -i /var/log/relay/
```

## Профилирование

### CPU профилирование

```bash
# Включение CPU профилирования
curl -X POST http://localhost:8081/debug/pprof/profile?seconds=30 -o cpu.prof

# Анализ профиля
go tool pprof cpu.prof
```

### Memory профилирование

```bash
# Включение memory профилирования
curl -X POST http://localhost:8081/debug/pprof/heap -o memory.prof

# Анализ профиля
go tool pprof memory.prof
```

### Goroutine профилирование

```bash
# Получение информации о goroutines
curl http://localhost:8081/debug/pprof/goroutine?debug=1
```

## Устранение неполадок

### Частые проблемы

#### 1. Ошибка парсинга конфигурации

**Симптомы:**
```
Failed to load configuration: failed to parse config file: json: cannot unmarshal string into Go struct field .server.timeout of type time.Duration
```

**Решение:**
Убедитесь, что все поля `time.Duration` указаны в виде строк с суффиксами времени:
```json
{
  "server": {
    "timeout": "30s",
    "ping_interval": "30s",
    "cleanup_interval": "1m"
  }
}
```

#### 2. Порт уже занят

**Симптомы:**
```
Failed to start server: failed to create listener: listen tcp :8080: bind: address already in use
```

**Решение:**
```bash
# Поиск процесса, использующего порт
sudo lsof -i :8080

# Остановка процесса
sudo kill -9 <PID>

# Или изменение порта в конфигурации
```

#### 3. Ошибка JWT секрета

**Симптомы:**
```
JWT secret must be set via RELAY_JWT_SECRET environment variable
```

**Решение:**
```bash
# Установка переменной окружения
export RELAY_JWT_SECRET=your-secure-secret

# Или замена placeholder в конфигурации
```

#### 4. Ошибка Docker контейнера

**Симптомы:**
```
docker: Error response from daemon: failed to create task for container: failed to create shim task: OCI runtime create failed: runc create failed: unable to start container process: error during container init: exec: "--config": executable file not found in $PATH
```

**Решение:**
Укажите правильную команду запуска:
```bash
docker run cloudbridge-relay ./relay --config config/config.json
```

#### 5. Высокое использование памяти

**Симптомы:**
- Медленная работа сервера
- Ошибки OOM
- Высокие значения `relay_technical_memory_usage_bytes`

**Решение:**
```bash
# Проверка утечек памяти
curl http://localhost:8081/debug/pprof/heap

# Анализ goroutines
curl http://localhost:8081/debug/pprof/goroutine

# Перезапуск сервера
sudo systemctl restart relay
```

#### 6. Проблемы с логированием

**Симптомы:**
- Логи не записываются
- Ошибки доступа к файлам логов
- Переполнение диска

**Решение:**
```bash
# Проверка прав доступа
sudo chown -R relay:relay /var/log/relay/

# Проверка места на диске
df -h /var/log/relay/

# Очистка старых логов
find /var/log/relay -name "*.log.*" -mtime +30 -delete
```

### Диагностические команды

```bash
# Комплексная проверка здоровья
#!/bin/bash
echo "=== CloudBridge Relay Health Check ==="

# Проверка процесса
echo "1. Process check:"
if pgrep relay > /dev/null; then
    echo "✓ Relay process is running (PID: $(pgrep relay))"
else
    echo "✗ Relay process is not running"
fi

# Проверка портов
echo "2. Port check:"
if netstat -tlnp | grep :8080 > /dev/null; then
    echo "✓ Port 8080 is listening"
else
    echo "✗ Port 8080 is not listening"
fi

if netstat -tlnp | grep :8081 > /dev/null; then
    echo "✓ Port 8081 is listening"
else
    echo "✗ Port 8081 is not listening"
fi

# Проверка метрик
echo "3. Metrics check:"
if curl -s http://localhost:8081/metrics > /dev/null; then
    echo "✓ Metrics endpoint is accessible"
else
    echo "✗ Metrics endpoint is not accessible"
fi

# Проверка логов
echo "4. Logs check:"
if [ -f /var/log/relay/relay.log ]; then
    echo "✓ Log file exists"
    echo "  Last log entry: $(tail -1 /var/log/relay/relay.log | jq -r '.msg // "unknown"')"
else
    echo "✗ Log file does not exist"
fi

# Проверка конфигурации
echo "5. Configuration check:"
if [ -f /etc/relay/config.json ]; then
    echo "✓ Configuration file exists"
    if jq empty /etc/relay/config.json 2>/dev/null; then
        echo "✓ Configuration file is valid JSON"
    else
        echo "✗ Configuration file is not valid JSON"
    fi
else
    echo "✗ Configuration file does not exist"
fi

echo "=== Health Check Complete ==="
```

### Получение поддержки

При возникновении проблем соберите следующую информацию:

1. **Логи:**
   ```bash
   sudo journalctl -u relay > relay-logs.txt
   tail -100 /var/log/relay/relay.log > recent-logs.txt
   ```

2. **Метрики:**
   ```bash
   curl http://localhost:8081/metrics > metrics.txt
   ```

3. **Конфигурация:**
   ```bash
   cat /etc/relay/config.json > config.txt
   ```

4. **Системная информация:**
   ```bash
   uname -a > system-info.txt
   free -h >> system-info.txt
   df -h >> system-info.txt
   ```

5. **Профили (если доступны):**
   ```bash
   curl http://localhost:8081/debug/pprof/goroutine > goroutines.txt
   curl http://localhost:8081/debug/pprof/heap > heap.prof
   ```

Создайте issue в репозитории с приложенными файлами и описанием проблемы.

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