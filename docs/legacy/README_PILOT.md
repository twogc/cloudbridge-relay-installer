# CloudBridge Relay - Инструкция для тестового пилота

## Готовые конфигурации

В корне проекта находятся два готовых конфига для пилотирования:

- **`server-a-pilot.yaml`** - Конфигурация для Server A
- **`server-b-pilot.yaml`** - Конфигурация для Server B

## Инструкции для клиента

### 1. **Подготовка**
```bash
# Убедиться, что cloudbridge-client исполняемый
chmod +x cloudbridge-client
```

### 2. **Запуск Server A**
```bash
./test/relay-client/cloudbridge-client p2p --config server-a-pilot.yaml --log-level debug
```

### 3. **Запуск Server B**
```bash
./test/relay-client/cloudbridge-client p2p --config server-b-pilot.yaml --log-level debug
```

## Внешние endpoints для тестирования

### **Health Check:**
```bash
curl -k https://edge.2gc.ru/health
```

### **API тестирование:**
```bash
curl -X POST -k https://edge.2gc.ru/api/v1/tenants/mesh-network-test/peers/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer [TOKEN]" \
  -d '{"peer_id": "test-peer", "public_key": "test-key", "allowed_ips": ["10.100.0.1/32"]}'
```

## Ожидаемые результаты

- Server A подключится к `https://relay.2gc.ru`
- Server B подключится к `https://relay.2gc.ru`
- Оба сервера получат IP в сети `10.100.0.0/16`
- Серверы увидят друг друга в P2P mesh
- Установится прямое P2P соединение

## 🔑 JWT токены

Токены уже встроены в конфигурации и действительны без ограничений по времени.

## Поддомены

- **relay.2gc.ru** - Основной P2P релей
- **p2p.2gc.ru** - P2P Discovery API
- **admin.2gc.ru** - Административная панель
- **monitor.2gc.ru** - Мониторинг и метрики
- **api.2gc.ru** - Общий API шлюз

## ⚠️ Важно

- Токены действительны без ограничений по времени
- Внешний доступ работает через поддомены
- Безопасность настроена (JWT + SSL)
- Мониторинг доступен через веб-интерфейс

## Готово к пилотированию!

Конфигурации готовы для передачи клиенту. Оба сервера смогут подключиться к CloudBridge Relay и образовать P2P mesh сеть.
