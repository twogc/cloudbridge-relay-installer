# Тестирование серверного клиента и работа Relay-сервера

## Тестирование серверного клиента

### 1. Подготовка тестового окружения

1. **Запуск relay-сервера:**
```bash
# Запуск в режиме отладки
./relay-server --debug

# Проверка логов
tail -f relay-server.log
```

2. **Запуск тест-клиента:**
```bash
python3 test_client.py
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

#### 2.1 Регистрация клиента
```json
// Запрос
{
    "type": "register",
    "token": "jwt-token"
}

// Ответ
{
    "status": "ok"
}
```

#### 2.2 Создание туннеля
```json
// Запрос
{
    "type": "connect",
    "local_port": 3389
}

// Ответ
{
    "status": "ok",
    "tunnel_id": "unique-tunnel-id"
}
```

#### 2.3 Обработка ошибок
```json
// Ответ с ошибкой
{
    "status": "error",
    "error": "error_code",
    "message": "Описание ошибки"
}
```

### 3. Ограничения и лимиты

1. **Лимиты соединений:**
   - Максимум 100 одновременных соединений
   - Таймаут неактивности: 5 минут
   - Максимальный размер сообщения: 1MB

2. **Ограничения туннелей:**
   - Максимум 10 туннелей на клиент
   - Таймаут туннеля: 30 минут
   - Максимальная скорость: 10MB/s

### 4. Мониторинг и логирование

1. **Метрики:**
   - Количество активных соединений
   - Количество созданных туннелей
   - Время отклика
   - Использование ресурсов

2. **Логи:**
   - Уровень: DEBUG
   - Формат: JSON
   - Ротация: ежедневная
   - Хранение: 30 дней

### 5. Отладка

1. **Включение отладочного режима:**
```bash
./relay-server --debug --log-level=debug
```

2. **Проверка состояния:**
```bash
# Проверка активных соединений
netstat -an | grep 8080

# Просмотр логов
tail -f relay-server.log
```

3. **Тестирование производительности:**
```bash
# Установка нагрузки
ab -n 1000 -c 100 http://localhost:8080/

# Мониторинг ресурсов
top -p $(pgrep relay-server)
```

## Рекомендации по тестированию

1. **Подготовка тестового окружения:**
   - Используйте изолированное окружение
   - Настройте тестовые сертификаты
   - Подготовьте тестовые JWT токены

2. **Покрытие тестами:**
   - Базовое подключение
   - Создание туннелей
   - Обработка ошибок
   - Нагрузочное тестирование
   - Тестирование отказоустойчивости

3. **Мониторинг тестов:**
   - Логирование всех операций
   - Сбор метрик
   - Анализ производительности

4. **Автоматизация:**
   - CI/CD интеграция
   - Автоматический запуск тестов
   - Генерация отчетов 