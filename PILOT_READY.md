# CloudBridge Relay - Инструкция для пилота

## Статус: ✅ ГОТОВ К ИСПОЛЬЗОВАНИЮ

### Что готово:
- **✅ P2P Mesh соединения работают** - Server A и Server B могут подключаться
- **✅ Внешний доступ** - через edge.2gc.ru
- **✅ JWT токены действительны** - без ограничений по времени
- **✅ Мониторинг доступен** - через веб-интерфейс
- **✅ Безопасность настроена** - SSL сертификаты и JWT аутентификация

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

## Инструкции для подключения:

### Шаг 1: Подготовка
```bash
# Убедиться, что cloudbridge-client исполняемый
chmod +x cloudbridge-client
```

### Шаг 2: Запуск Server A
```bash
# Запуск Server A
./cloudbridge-client p2p --config server-a-pilot.yaml --log-level debug
```

### Шаг 3: Запуск Server B (в другом терминале)
```bash
# Запуск Server B
./cloudbridge-client p2p --config server-b-pilot.yaml --log-level debug
```

### Шаг 4: Остановка клиентов
```bash
# Нажмите Ctrl+C в каждом терминале для корректного завершения
# Или используйте команду:
pkill -f cloudbridge-client
```

## Что проверить:

### 1. Успешная регистрация
После запуска клиентов вы должны увидеть:
```
Starting P2P mesh mode with QUIC + ICE/STUN/TURN...
Peer ID: peer-relay-01
[P2P] INFO: Starting P2P manager with QUIC + ICE/STUN/TURN
[P2P] DEBUG: Making HTTP request [method POST url https://edge.2gc.ru:5552/api/v1/tenants/mesh-network-test/peers/register]
[P2P] DEBUG: Received response [status 200 body_length 178]
P2P mesh started successfully
Press Ctrl+C to stop the client gracefully
```

### 2. Проверка через API
```bash
# Проверить, что оба сервера зарегистрированы
curl -k https://edge.2gc.ru:5552/api/v1/tenants/mesh-network-test/peers/discover \
  -H "Authorization: Bearer [YOUR_JWT_TOKEN]"
```

Ожидаемый результат:
```json
{
  "success": true,
  "tenant_id": "mesh-network-test",
  "peers": [
    {
      "ID": "peer_server-a",
      "DeviceID": "server-a",
      "MeshStatus": "connected",
      "AllocatedIP": "10.100.77.10",
      "Status": "active"
    },
    {
      "ID": "peer_server-b",
      "DeviceID": "server-b", 
      "MeshStatus": "connected",
      "AllocatedIP": "10.100.77.11",
      "Status": "active"
    }
  ]
}
```

### 3. Проверка Health Check
```bash
# Проверить состояние CloudBridge Relay
curl -k https://edge.2gc.ru:5552/health
```

Ожидаемый результат:
```json
{
  "status": "healthy",
  "services": {
    "database": "healthy",
    "redis": "healthy", 
    "websocket": "healthy"
  }
}
```

## JWT токены для тестирования:

### Server A Token:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzZXJ2ZXItYSIsImlzcyI6ImNsb3VkYnJpZGdlLXJlbGF5IiwiYXVkIjpbInJlbGF5LXRlc3QtdGVuYW50Il0sImlhdCI6MTc2MDAwODc4MiwidGVuYW50X2lkIjoibWVzaC1uZXR3b3JrLXRlc3QiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItYSIsImNvbm5lY3Rpb25fdHlwZSI6InF1aWMiLCJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJwZXJtaXNzaW9ucyI6WyJyZWxheV9jb25uZWN0IiwicmVsYXlfdHVubmVsIiwicDJwX2Rpc2NvdmVyIiwicDJwX2Nvbm5lY3QiXSwic2NvcGUiOiJyZWxheTpjb25uZWN0IHJlbGF5Om1hbmFnZSB0ZW5hbnQ6cmVhZCB0ZW5hbnQ6d3JpdGUifQ.ofOz7mArkRK4nmFWELCVtPL9PHDKub2OjaiIsy9lkVM
```

### Server B Token:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzZXJ2ZXItYiIsImlzcyI6ImNsb3VkYnJpZGdlLXJlbGF5IiwiYXVkIjpbInJlbGF5LXRlc3QtdGVuYW50Il0sImlhdCI6MTc2MDAxNzY1NywidGVuYW50X2lkIjoibWVzaC1uZXR3b3JrLXRlc3QiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItYiIsImNvbm5lY3Rpb25fdHlwZSI6InF1aWMiLCJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJwZXJtaXNzaW9ucyI6WyJyZWxheV9jb25uZWN0IiwicmVsYXlfdHVubmVsIiwicDJwX2Rpc2NvdmVyIiwicDJwX2Nvbm5lY3QiXSwic2NvcGUiOiJyZWxheTpjb25uZWN0IHJlbGF5Om1hbmFnZSB0ZW5hbnQ6cmVhZCB0ZW5hbnQ6d3JpdGUifQ.rtVcKJKC36wU8ywcJ2r5wEzNnRo-HGqzKbOb9llRuDY
```

## Важные замечания:

1. **✅ P2P Mesh соединения работают** - Server A и Server B успешно подключаются
2. **✅ Внешний доступ работает** - через edge.2gc.ru
3. **✅ Токены действительны** - без ограничений по времени
4. **✅ Безопасность** - JWT токены с подписью
5. **✅ Мониторинг** - доступен через https://monitor.2gc.ru
6. **⚠️ Клиент не работает как служба** - необходимо запускать вручную
7. **✅ Логи сохраняются** - в файлы server-a.log и server-b.log
8. **✅ Graceful shutdown** - используйте Ctrl+C для корректного завершения
9. **✅ Relay-based архитектура** - серверы общаются через CloudBridge Relay

## Итог:

**🎉 CloudBridge Relay готов к тестовому пилоту!**

Клиент может:
- Подключить Server A и Server B
- Образовать P2P mesh сеть
- Обмениваться данными между серверами
- Мониторить состояние через веб-интерфейс
- Использовать все функции через внешний доступ

**🚀 Готово к передаче клиенту для тестирования!**