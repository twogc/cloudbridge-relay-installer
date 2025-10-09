## Скачивание и установка клиента:

### Шаг 1: Скачивание клиента
1. Перейдите на страницу релизов: https://github.com/2gc-dev/relay-client/releases
2. Скачайте архив для вашей платформы (Linux/Windows/macOS)
3. Распакуйте архив
4. Скопируйте конфигурационные файлы в папку с клиентом

### Шаг 2: Конфигурационные файлы
Скопируйте эти файлы в папку с клиентом:
- **server-a-pilot.yaml** - Конфигурация для Server A
- **server-b-pilot.yaml** - Конфигурация для Server B

## Инструкции для клиента:

### Шаг 1: Подготовка
```bash
# Убедиться, что cloudbridge-client исполняемый
chmod +x cloudbridge-client
```

### Шаг 2: Запуск Server A (в фоновом режиме)
```bash
# Запуск в фоновом режиме
nohup ./cloudbridge-client p2p --config server-a-pilot.yaml --log-level debug > server-a.log 2>&1 &

# Проверить, что процесс запущен
ps aux | grep cloudbridge-client
```

### Шаг 3: Запуск Server B (в фоновом режиме)
```bash
# Запуск в фоновом режиме
nohup ./cloudbridge-client p2p --config server-b-pilot.yaml --log-level debug > server-b.log 2>&1 &

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
- **Server A и Server B видят друг друга** - через CloudBridge Relay API
- **P2P Mesh соединения установлены** - статус "connected" для обоих пиров
- **Обмен данными через mesh сеть** - через CloudBridge Relay (relay-based архитектура)
- **Прямые WireGuard туннели** - настроены для связи с CloudBridge Relay

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

**⚠️ ВАЖНО:** Прямой ping может не работать, так как в relay-based архитектуре серверы общаются через CloudBridge Relay, а не напрямую друг с другом. Это нормальное поведение!

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

## JWT токены (готовы к использованию):

### Server A Token:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzZXJ2ZXItYSIsInRlbmFudF9pZCI6Im1lc2gtbmV0d29yay10ZXN0Iiwic2VydmVyX2lkIjoic2VydmVyLWEiLCJjb25uZWN0aW9uX3R5cGUiOiJxdWljIiwicHJvdG9jb2xfdHlwZSI6InAycC1tZXNoIiwicGVybWlzc2lvbnMiOlsicmVsYXlfY29ubmVjdCIsInJlbGF5X3R1bm5lbCIsInAycF9kaXNjb3ZlciIsInAycF9jb25uZWN0Il0sImlzcyI6ImNsb3VkYnJpZGdlLXJlbGF5IiwiYXVkIjpbInJlbGF5LXRlc3QtdGVuYW50Il0sIm5iZiI6MTc1OTIyMTM0OCwiaWF0IjoxNzU5MjIxMzQ4fQ.1s16Kx4yS0Ya2-o7U9Ce_Ft61gKpOJqAXU8Ns6m-I9s
```

### Server B Token:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzZXJ2ZXItYiIsInRlbmFudF9pZCI6Im1lc2gtbmV0d29yay10ZXN0Iiwic2VydmVyX2lkIjoic2VydmVyLWIiLCJjb25uZWN0aW9uX3R5cGUiOiJxdWljIiwicHJvdG9jb2xfdHlwZSI6InAycC1tZXNoIiwicGVybWlzc2lvbnMiOlsicmVsYXlfY29ubmVjdCIsInJlbGF5X3R1bm5lbCIsInAycF9kaXNjb3ZlciIsInAycF9jb25uZWN0Il0sImlzcyI6ImNsb3VkYnJpZGdlLXJlbGF5IiwiYXVkIjpbInJlbGF5LXRlc3QtdGVuYW50Il0sIm5iZiI6MTc1OTIyMTM0OCwiaWF0IjoxNzU5MjIxMzQ4fQ.1s16Kx4yS0Ya2-o7U9Ce_Ft61gKpOJqAXU8Ns6m-I9s
```

## Поддомены для мониторинга:

- **relay.2gc.ru** - Основной P2P релей
- **p2p.2gc.ru** - P2P Discovery API
- **admin.2gc.ru** - Административная панель
- **monitor.2gc.ru** - Мониторинг и метрики
- **api.2gc.ru** - Общий API шлюз
- **masque.2gc.ru** - HTTP/3 MASQUE прокси

## Важные замечания:

1. **P2P Mesh соединения работают** - Server A и Server B успешно подключаются
2. **CloudBridge Relay стабилен** - все критические ошибки исправлены
3. **Токены действительны** - без ограничений по времени
4. **Внешний доступ работает** - через edge.2gc.ru
5. **Безопасность** - JWT токены с подписью
6. **Мониторинг** - доступен через https://monitor.2gc.ru
7. **Администрирование** - доступно через https://admin.2gc.ru
8. **⚠️ Клиент не работает как служба** - необходимо запускать в фоновом режиме
9. **Логи сохраняются** - в файлы server-a.log и server-b.log
10. **⚠️ Автоматический перезапуск** - не предусмотрен, нужно контролировать вручную
11. **Relay-based архитектура** - серверы общаются через CloudBridge Relay, не напрямую
