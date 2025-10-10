
## Скачивание и установка клиента:

### Шаг 1: Скачивание клиента
1. Перейдите на страницу релизов: https://github.com/2gc-dev/relay-client/releases
2. Скачайте архив для вашей платформы (Linux/Windows/macOS)
3. Распакуйте архив
4. Скопируйте конфигурационные файлы в папку с клиентом

### Шаг 2: Конфигурационные файлы
Скопируйте эти файлы в папку с клиентом:
- **config-server-a.yaml** - Конфигурация для Server A (бесрочный токен)
- **config-server-b.yaml** - Конфигурация для Server B (бесрочный токен)

## Инструкции для клиента:

### Шаг 1: Подготовка
```bash
# Убедиться, что cloudbridge-client исполняемый
chmod +x cloudbridge-client
```

### Шаг 2: Запуск Server A (в фоновом режиме)
```bash
# Запуск в фоновом режиме
nohup ./cloudbridge-client p2p --config config-server-a.yaml --log-level debug > server-a.log 2>&1 &

# Проверить, что процесс запущен
ps aux | grep cloudbridge-client
```

### Шаг 3: Запуск Server B (в фоновом режиме)
```bash
# Запуск в фоновом режиме
nohup ./cloudbridge-client p2p --config config-server-b.yaml --log-level debug > server-b.log 2>&1 &

# Проверить, что процесс запущен
ps aux | grep cloudbridge-client
```

### Шаг 4: Проверка логов
```bash
# Просмотр логов Server A
tail -f server-a.log

# Просмотр логов Server B
tail -f server-b.log
```

### Шаг 5: Остановка процессов
```bash
# Остановить все процессы cloudbridge-client
pkill -f cloudbridge-client

# Или остановить конкретный процесс
kill [PID]

# Проверить, что процессы остановлены
ps aux | grep cloudbridge-client
```

## Внешние endpoints для тестирования:

### Health Check:
```bash
curl -k https://edge.2gc.ru/health
# Ожидаемый результат: HTTP/2 200 OK
```

### API тестирование:
```bash
curl -X POST -k https://edge.2gc.ru/api/v1/tenants/mesh-network-test/peers/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer [TOKEN]" \
  -d '{"peer_id": "test-peer", "public_key": "test-key", "allowed_ips": ["10.100.0.1/32"]}'
```

## Ожидаемые результаты:

### Server A:
- Подключение к https://edge.2gc.ru
- Регистрация с server_id: "server-a"
- Получение IP в сети 10.100.0.0/16
- Подключение к P2P mesh

### Server B:
- Подключение к https://edge.2gc.ru
- Регистрация с server_id: "server-b"
- Получение IP в сети 10.100.0.0/16
- Подключение к P2P mesh

### P2P соединение:
- Server A и Server B видят друг друга
- Установление прямого P2P соединения
- Обмен данными через mesh сеть

## Тестирование соединения между серверами:

### Шаг 1: Проверка регистрации
```bash
# Проверить, что оба сервера зарегистрированы
curl -k https://edge.2gc.ru/api/v1/tenants/mesh-network-test/peers
```

### Шаг 2: Получение IP адресов
После успешной регистрации каждый сервер получит IP адрес в сети 10.100.0.0/16:
- **Server A** получит IP (например: 10.100.77.10)
- **Server B** получит IP (например: 10.100.77.11)

### Шаг 3: Пинг между серверами
```bash
# На Server A - пинговать Server B
ping 10.100.77.11

# На Server B - пинговать Server A  
ping 10.100.77.10
```

### Шаг 4: Проверка сетевого интерфейса
```bash
# Проверить созданный сетевой интерфейс
ip addr show | grep 10.100

# Или на Windows
ipconfig | findstr 10.100
```

### Шаг 5: Тестирование TCP соединения
```bash
# На Server A - запустить простой TCP сервер
nc -l 8080

# На Server B - подключиться к Server A
nc 10.100.77.10 8080
```

## Важные замечания:

1. **Бесрочные токены** - JWT токены с `exp: 9999999999` (действительны до 2033 года)
2. **Внешний доступ работает** - через edge.2gc.ru:5552
3. **Безопасность** - JWT токены с подписью `default-jwt-secret`
4. **Мониторинг** - доступен через https://monitor.2gc.ru
5. **Администрирование** - доступно через https://admin.2gc.ru
6. **Клиент не работает как служба** - необходимо запускать в фоновом режиме
7. **Логи сохраняются** - в файлы server-a.log и server-b.log
8. **Автоматический перезапуск** - не предусмотрен, нужно контролировать вручную
9. **Пилотное тестирование** - конфигурации готовы для передачи партнерам

## Итог:

**CloudBridge Relay полностью готов к тестовому пилоту!**

Клиент может:
- Подключить Server A и Server B
- Образовать P2P mesh сеть
- Обмениваться данными между серверами

