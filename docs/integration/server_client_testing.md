# Тестирование серверного клиента и работа Relay-сервера

## Тестирование серверного клиента

### 1. Подготовка тестового окружения

1. **Запуск relay-сервера:**
```bash
# Запуск в режиме отладки на порту 8080 (по умолчанию)
./relay-server --debug

# Запуск на другом порту
./relay-server --debug --port 9090

# Проверка логов
tail -f relay-server.log
```

2. **Запуск тест-клиента:**
```bash
# По умолчанию подключается к порту 8080
python3 test_client.py

# Указание другого порта
python3 test_client.py --port 9090
```

### 2. Тестовые сценарии

#### 2.1 Базовое подключение
```python
# test_basic_connection.py
import socket
import json
import time

def test_basic_connection():
    # Подключение к relay-серверу
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect(('localhost', 8080))
    
    # Отправка регистрационного сообщения
    register_msg = {
        "type": "register",
        "token": "your-jwt-token"
    }
    sock.send(json.dumps(register_msg).encode())
    
    # Получение ответа
    response = json.loads(sock.recv(1024).decode())
    assert response["status"] == "ok"
    
    sock.close()
```

#### 2.2 Тест создания туннеля
```python
# test_tunnel_creation.py
def test_tunnel_creation():
    sock = connect_to_relay()  # Используем функцию из предыдущего примера
    
    # Создание туннеля
    connect_msg = {
        "type": "connect",
        "local_port": 3389
    }
    sock.send(json.dumps(connect_msg).encode())
    
    # Проверка ответа
    response = json.loads(sock.recv(1024).decode())
    assert response["status"] == "ok"
    assert "tunnel_id" in response
```

#### 2.3 Тест обработки ошибок
```python
# test_error_handling.py
def test_invalid_token():
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect(('localhost', 8080))
    
    # Отправка неверного токена
    register_msg = {
        "type": "register",
        "token": "invalid-token"
    }
    sock.send(json.dumps(register_msg).encode())
    
    response = json.loads(sock.recv(1024).decode())
    assert response["status"] == "error"
    assert response["error"] == "invalid_token"
```

### 3. Автоматизированное тестирование

```python
# test_suite.py
import unittest
import time

class RelayClientTests(unittest.TestCase):
    def setUp(self):
        # Подготовка тестового окружения
        self.sock = None
        
    def tearDown(self):
        # Очистка после тестов
        if self.sock:
            self.sock.close()
            
    def test_connection_flow(self):
        # Тест полного цикла подключения
        self.sock = connect_to_relay()
        self.assertTrue(self.sock)
        
        # Тест создания туннеля
        tunnel_id = create_tunnel(self.sock)
        self.assertIsNotNone(tunnel_id)
        
        # Тест передачи данных
        test_data = b"Hello, Relay!"
        self.sock.send(test_data)
        response = self.sock.recv(1024)
        self.assertEqual(response, test_data)

if __name__ == '__main__':
    unittest.main()
```

## Работа Relay-сервера

### 1. Архитектура

Relay-сервер состоит из следующих компонентов:

1. **TCP-сервер:**
   - Принимает входящие соединения
   - Обрабатывает регистрацию клиентов
   - Управляет туннелями

2. **JWT-валидатор:**
   - Проверяет подпись токенов
   - Валидирует claims (exp, iat, nbf)
   - Проверяет права доступа

3. **Менеджер туннелей:**
   - Создает и удаляет туннели
   - Отслеживает состояние соединений
   - Обрабатывает таймауты

### 2. Протокол обмена сообщениями

#### 2.1 Последовательность взаимодействия
1. **Установка соединения:**
   - Клиент устанавливает TCP/TLS-соединение с relay-сервером
   - Сервер отправляет приветствие (hello)

2. **Аутентификация:**
   - Клиент отправляет JWT-токен
   - Сервер проверяет токен и отправляет auth_response

3. **Основной цикл:**
   - После успешной аутентификации клиент запускает event loop
   - Обрабатывает команды от сервера (register, heartbeat и др.)

#### 2.2 Форматы сообщений

##### Приветствие от сервера
```json
{
    "type": "hello",
    "version": "1.0",
    "features": ["tls", "heartbeat", "tunnel_info"]
}
```

##### Аутентификация
```json
// Запрос
{
    "type": "auth",
    "token": "jwt-token"
}

// Ответ
{
    "type": "auth_response",
    "status": "ok",
    "client_id": "unique-client-id"
}
```

##### Heartbeat
```json
// Запрос от сервера
{
    "type": "heartbeat"
}

// Ответ клиента
{
    "type": "heartbeat_response",
    "status": "ok"
}
```

##### Tunnel Info
```json
// Запрос от сервера
{
    "type": "tunnel_info",
    "tunnel_id": "unique-tunnel-id",
    "protocol": "rdp",  // или "ssh", "http", "https"
    "local_port": 3389,
    "remote_host": "internal-server",
    "remote_port": 3389,
    "options": {
        "username": "user",
        "password": "pass",
        "domain": "domain"
    }
}
```

### 3. Примеры реализации

#### 3.1 Базовый клиент с обработкой протокола
```python
import socket
import ssl
import json
import asyncio
import logging

class RelayClient:
    def __init__(self, host, port, token, use_tls=True):
        self.host = host
        self.port = port
        self.token = token
        self.use_tls = use_tls
        self.client_id = None
        self.logger = logging.getLogger(__name__)

    async def connect(self):
        # Установка TCP соединения
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        if self.use_tls:
            context = ssl.create_default_context()
            self.sock = context.wrap_socket(self.sock)
        
        self.sock.connect((self.host, self.port))
        
        # Ожидание hello
        hello = await self.receive_message()
        if hello["type"] != "hello":
            raise Exception("Invalid server response")
        
        # Отправка auth
        await self.send_message({
            "type": "auth",
            "token": self.token
        })
        
        # Ожидание auth_response
        auth_response = await self.receive_message()
        if auth_response["status"] != "ok":
            raise Exception("Authentication failed")
        
        self.client_id = auth_response["client_id"]
        self.logger.info(f"Connected with client_id: {self.client_id}")

    async def event_loop(self):
        while True:
            try:
                message = await self.receive_message()
                await self.handle_message(message)
            except Exception as e:
                self.logger.error(f"Error in event loop: {e}")
                break

    async def handle_message(self, message):
        handlers = {
            "heartbeat": self.handle_heartbeat,
            "tunnel_info": self.handle_tunnel_info,
            "register": self.handle_register
        }
        
        handler = handlers.get(message["type"])
        if handler:
            await handler(message)
        else:
            self.logger.warning(f"Unknown message type: {message['type']}")

    async def handle_heartbeat(self, message):
        await self.send_message({
            "type": "heartbeat_response",
            "status": "ok"
        })

    async def handle_tunnel_info(self, message):
        # Запуск локального прокси на основе tunnel_info
        await self.start_local_proxy(message)

    async def handle_register(self, message):
        # Обработка команды регистрации
        pass

    async def start_local_proxy(self, tunnel_info):
        protocol = tunnel_info["protocol"]
        if protocol == "rdp":
            await self.start_rdp_proxy(tunnel_info)
        elif protocol == "ssh":
            await self.start_ssh_proxy(tunnel_info)
        elif protocol in ["http", "https"]:
            await self.start_http_proxy(tunnel_info)

    async def start_rdp_proxy(self, tunnel_info):
        # Пример запуска RDP прокси
        import subprocess
        
        cmd = [
            "xfreerdp",
            f"/v:{tunnel_info['remote_host']}",
            f"/port:{tunnel_info['remote_port']}",
            f"/u:{tunnel_info['options']['username']}",
            f"/p:{tunnel_info['options']['password']}",
            f"/d:{tunnel_info['options']['domain']}"
        ]
        
        try:
            process = await asyncio.create_subprocess_exec(
                *cmd,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )
            self.logger.info(f"Started RDP proxy for tunnel {tunnel_info['tunnel_id']}")
        except Exception as e:
            self.logger.error(f"Failed to start RDP proxy: {e}")

    async def send_message(self, message):
        data = json.dumps(message).encode() + b"\n"
        self.sock.send(data)

    async def receive_message(self):
        data = await asyncio.get_event_loop().sock_recv(self.sock, 1024)
        return json.loads(data.decode())

# Использование
async def main():
    client = RelayClient(
        host="relay-server",
        port=8080,
        token="your-jwt-token",
        use_tls=True
    )
    
    await client.connect()
    await client.event_loop()

if __name__ == "__main__":
    asyncio.run(main())
```

#### 3.2 Тестирование клиента
```python
# test_relay_client.py
import unittest
import asyncio
from relay_client import RelayClient

class TestRelayClient(unittest.TestCase):
    async def asyncSetUp(self):
        self.client = RelayClient(
            host="localhost",
            port=8080,
            token="test-token"
        )
        await self.client.connect()

    async def asyncTearDown(self):
        self.client.sock.close()

    async def test_connection(self):
        self.assertIsNotNone(self.client.client_id)

    async def test_heartbeat(self):
        # Симуляция heartbeat от сервера
        await self.client.handle_message({
            "type": "heartbeat"
        })
        # Проверка ответа...

    async def test_tunnel_info(self):
        # Симуляция tunnel_info от сервера
        tunnel_info = {
            "type": "tunnel_info",
            "tunnel_id": "test-tunnel",
            "protocol": "rdp",
            "local_port": 3389,
            "remote_host": "test-server",
            "remote_port": 3389,
            "options": {
                "username": "test",
                "password": "test",
                "domain": "test"
            }
        }
        await self.client.handle_message(tunnel_info)
        # Проверка запуска прокси...

def run_async_test(coro):
    return asyncio.get_event_loop().run_until_complete(coro)

if __name__ == '__main__':
    unittest.main()
```

### 4. Рекомендации по реализации

1. **Обработка ошибок:**
   - Все сетевые операции должны быть в try-except
   - Реализовать механизм переподключения
   - Логировать все ошибки

2. **Безопасность:**
   - Использовать TLS для всех соединений
   - Не хранить чувствительные данные в открытом виде
   - Регулярно обновлять JWT токены

3. **Производительность:**
   - Использовать асинхронное программирование
   - Оптимизировать обработку сообщений
   - Реализовать пул соединений

4. **Мониторинг:**
   - Добавить метрики (время отклика, количество ошибок)
   - Реализовать health checks
   - Настроить алерты

### 5. Технические детали для разработчиков

#### 5.1 Форматы сообщений

#### 5.2 Таймауты и лимиты

1. **Соединение:**
   - Таймаут установки соединения: 10 секунд
   - Таймаут чтения/записи: 30 секунд
   - Максимальный размер сообщения: 1MB

2. **Туннели:**
   - Максимум туннелей на клиент: 10
   - Таймаут неактивности туннеля: 5 минут
   - Максимальная скорость: 10MB/s

3. **Heartbeat:**
   - Интервал: 30 секунд
   - Таймаут ответа: 5 секунд
   - Максимум пропущенных heartbeat: 3

### 6. Рекомендации по реализации

#### 6.1 Обработка соединений
```python
class ConnectionManager:
    def __init__(self):
        self.reconnect_delay = 1  # начальная задержка в секундах
        self.max_reconnect_delay = 30  # максимальная задержка
        self.max_retries = 5  # максимальное количество попыток

    async def connect_with_retry(self, client):
        retries = 0
        while retries < self.max_retries:
            try:
                await client.connect()
                return True
            except Exception as e:
                retries += 1
                delay = min(self.reconnect_delay * (2 ** retries), self.max_reconnect_delay)
                await asyncio.sleep(delay)
        return False
```

#### 6.2 Обработка ошибок
```python
class ErrorHandler:
    def __init__(self):
        self.error_counts = {}
        self.max_errors = 3
        self.error_window = 60  # секунды

    def handle_error(self, error_code):
        now = time.time()
        if error_code not in self.error_counts:
            self.error_counts[error_code] = []
        
        # Очистка старых ошибок
        self.error_counts[error_code] = [
            t for t in self.error_counts[error_code]
            if now - t < self.error_window
        ]
        
        self.error_counts[error_code].append(now)
        
        if len(self.error_counts[error_code]) >= self.max_errors:
            raise Exception(f"Too many {error_code} errors")
```

#### 6.3 Мониторинг
```python
class MetricsCollector:
    def __init__(self):
        self.metrics = {
            "connection_time": [],
            "message_latency": [],
            "error_count": 0,
            "tunnel_count": 0
        }

    def record_metric(self, metric_name, value):
        if metric_name in self.metrics:
            if isinstance(self.metrics[metric_name], list):
                self.metrics[metric_name].append(value)
            else:
                self.metrics[metric_name] = value

    def get_statistics(self):
        return {
            "avg_connection_time": statistics.mean(self.metrics["connection_time"]),
            "avg_message_latency": statistics.mean(self.metrics["message_latency"]),
            "error_rate": self.metrics["error_count"] / len(self.metrics["connection_time"]),
            "active_tunnels": self.metrics["tunnel_count"]
        }
```

### 7. Рекомендации по использованию

1. **Настройка таймаутов:**
   - Соединение: 10 секунд
   - Чтение/запись: 30 секунд
   - Heartbeat: 30 секунд интервал, 5 секунд таймаут

2. **Обработка ошибок:**
   - Все сетевые операции в try-except
   - Логирование всех ошибок
   - Автоматический reconnect при потере соединения

3. **Мониторинг:**
   - Отслеживание количества ошибок
   - Мониторинг пропущенных heartbeat
   - Сбор метрик производительности

4. **Безопасность:**
   - Использование TLS
   - Валидация всех входящих сообщений
   - Проверка размера сообщений

### 8. Детали реализации

#### 8.1 Обработка сообщений
```python
import json
import asyncio
from typing import Optional, Dict, Any

class MessageHandler:
    def __init__(self, reader: asyncio.StreamReader, writer: asyncio.StreamWriter):
        self.reader = reader
        self.writer = writer
        self.max_message_size = 1024 * 1024  # 1MB
        self.buffer = bytearray()

    async def read_message(self) -> Optional[Dict[str, Any]]:
        try:
            # Чтение до разделителя \n
            while True:
                chunk = await self.reader.read(1024)
                if not chunk:
                    return None
                
                self.buffer.extend(chunk)
                
                # Поиск разделителя
                if b'\n' in self.buffer:
                    message, self.buffer = self.buffer.split(b'\n', 1)
                    
                    # Проверка размера
                    if len(message) > self.max_message_size:
                        raise ValueError("Message too large")
                    
                    return json.loads(message.decode())
        except Exception as e:
            logger.error(f"Error reading message: {e}")
            return None

    async def write_message(self, message: Dict[str, Any]) -> bool:
        try:
            data = json.dumps(message).encode() + b'\n'
            if len(data) > self.max_message_size:
                raise ValueError("Message too large")
            
            self.writer.write(data)
            await self.writer.drain()
            return True
        except Exception as e:
            logger.error(f"Error writing message: {e}")
            return False
```

#### 8.2 Heartbeat и таймауты
```python
class HeartbeatManager:
    def __init__(self, message_handler: MessageHandler):
        self.message_handler = message_handler
        self.heartbeat_interval = 30  # секунды
        self.heartbeat_timeout = 5    # секунды
        self.max_missed_heartbeats = 3
        self.missed_heartbeats = 0
        self.last_heartbeat_time = 0
        self.heartbeat_task = None

    async def start_heartbeat(self):
        self.heartbeat_task = asyncio.create_task(self._heartbeat_loop())

    async def stop_heartbeat(self):
        if self.heartbeat_task:
            self.heartbeat_task.cancel()
            try:
                await self.heartbeat_task
            except asyncio.CancelledError:
                pass

    async def _heartbeat_loop(self):
        while True:
            try:
                # Отправка heartbeat
                success = await self.message_handler.write_message({
                    "type": "heartbeat"
                })
                
                if not success:
                    self.missed_heartbeats += 1
                else:
                    # Ожидание ответа с таймаутом
                    try:
                        async with asyncio.timeout(self.heartbeat_timeout):
                            response = await self.message_handler.read_message()
                            if response and response.get("type") == "heartbeat_response":
                                self.missed_heartbeats = 0
                                self.last_heartbeat_time = time.time()
                            else:
                                self.missed_heartbeats += 1
                    except asyncio.TimeoutError:
                        self.missed_heartbeats += 1
                
                # Проверка количества пропущенных heartbeat
                if self.missed_heartbeats >= self.max_missed_heartbeats:
                    logger.error("Too many missed heartbeats")
                    raise ConnectionError("Heartbeat timeout")
                
                await asyncio.sleep(self.heartbeat_interval)
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in heartbeat loop: {e}")
                break
```

#### 8.3 Reconnect и retry
```python
class ReconnectManager:
    def __init__(self, client: 'RelayClient'):
        self.client = client
        self.initial_delay = 1  # секунды
        self.max_delay = 30     # секунды
        self.max_retries = 5
        self.current_delay = self.initial_delay
        self.retry_count = 0
        self.reconnect_task = None

    async def start_reconnect(self):
        self.reconnect_task = asyncio.create_task(self._reconnect_loop())

    async def stop_reconnect(self):
        if self.reconnect_task:
            self.reconnect_task.cancel()
            try:
                await self.reconnect_task
            except asyncio.CancelledError:
                pass

    async def _reconnect_loop(self):
        while self.retry_count < self.max_retries:
            try:
                logger.info(f"Attempting to reconnect (attempt {self.retry_count + 1})")
                if await self.client.connect():
                    logger.info("Reconnected successfully")
                    self.retry_count = 0
                    self.current_delay = self.initial_delay
                    return True
            except Exception as e:
                logger.error(f"Reconnect attempt failed: {e}")
            
            self.retry_count += 1
            self.current_delay = min(self.current_delay * 2, self.max_delay)
            await asyncio.sleep(self.current_delay)
        
        logger.error("Max retry attempts reached")
        return False
```

#### 8.4 Обработка ошибок
```python
class ErrorHandler:
    ERROR_CODES = {
        'invalid_token': 'Неверный JWT токен',
        'token_expired': 'Токен истек',
        'unauthorized': 'Нет прав доступа',
        'invalid_message': 'Неверный формат сообщения',
        'connection_limit': 'Превышен лимит соединений',
        'tunnel_limit': 'Превышен лимит туннелей',
        'internal_error': 'Внутренняя ошибка сервера'
    }

    def __init__(self):
        self.error_counts = {code: 0 for code in self.ERROR_CODES}
        self.error_window = 60  # секунды
        self.error_timestamps = {code: [] for code in self.ERROR_CODES}

    def handle_error(self, error_code: str, message: str = None):
        if error_code not in self.ERROR_CODES:
            logger.warning(f"Unknown error code: {error_code}")
            return
        
        now = time.time()
        
        # Очистка старых ошибок
        self.error_timestamps[error_code] = [
            t for t in self.error_timestamps[error_code]
            if now - t < self.error_window
        ]
        
        # Добавление новой ошибки
        self.error_timestamps[error_code].append(now)
        self.error_counts[error_code] = len(self.error_timestamps[error_code])
        
        # Логирование
        error_message = message or self.ERROR_CODES[error_code]
        logger.error(f"Error {error_code}: {error_message}")
        
        # Проверка частоты ошибок
        if self.error_counts[error_code] >= 3:
            logger.error(f"Too many {error_code} errors in {self.error_window} seconds")
            raise Exception(f"Error threshold exceeded for {error_code}")

    def get_error_stats(self) -> Dict[str, int]:
        return {
            code: len(timestamps)
            for code, timestamps in self.error_timestamps.items()
        }
```

#### 8.5 Интеграция компонентов

#### 14.11 Мониторинг

1. **Метрики Prometheus:**
   ```
   # Счетчики
   relay_connections_total - общее количество подключений
   relay_rejected_connections_total - отклоненные подключения
   relay_errors_total{code="error_code"} - ошибки по кодам
   relay_heartbeats_total - количество heartbeat
   relay_heartbeat_errors_total - ошибки heartbeat

   # Гейджи
   relay_active_connections - активные подключения
   relay_active_tunnels - активные туннели

   # Гистограммы
   relay_connection_duration_seconds - время подключения
   relay_heartbeat_latency_seconds - задержка heartbeat
   ```

2. **Эндпоинты:**
   ```bash
   # Метрики Prometheus
   GET /metrics
   Content-Type: text/plain
   
   # Health check
   GET /health
   Content-Type: application/json
   {
       "status": "ok",
       "version": "1.0.0",
       "uptime": "24h",
       "metrics": {
           "active_connections": 10,
           "active_tunnels": 5
       }
   }
   ```

3. **Интеграция:**
   - Метрики доступны через `/metrics` эндпоинт
   - Формат метрик соответствует Prometheus
   - Health check доступен через `/health` эндпоинт
   - Все метрики готовы к сбору вашей системой мониторинга