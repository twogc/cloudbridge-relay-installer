# CloudBridge Relay P2P Mesh Setup

## 🎯 Обзор

Настроена P2P mesh сеть с двумя серверами (А и Б) в одном тенанте `mesh-network-test`, которые могут подключаться друг к другу через CloudBridge Relay.

## 📁 Конфигурационные файлы

### Сервер А
- **Файл**: `test/relay-client/config-server-a.yaml`
- **Server ID**: `server-a`
- **Token**: `eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...` (HS256)
- **IP**: `10.100.77.10`

### Сервер Б
- **Файл**: `test/relay-client/config-server-b.yaml`
- **Server ID**: `server-b`
- **Token**: `eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...` (HS256)
- **IP**: `10.100.77.11`

## 🚀 Запуск P2P Mesh

### Автоматический запуск (рекомендуется)
```bash
cd /home/ubuntu/cloudbridge-relay-installer
./test-p2p-mesh.sh
```

### Ручной запуск
```bash
# Терминал 1 - Сервер А
cd /home/ubuntu/cloudbridge-relay-installer
./test/relay-client/cloudbridge-client p2p --config test/relay-client/config-server-a.yaml

# Терминал 2 - Сервер Б
cd /home/ubuntu/cloudbridge-relay-installer
./test/relay-client/cloudbridge-client p2p --config test/relay-client/config-server-b.yaml
```

## 📊 Мониторинг

### Логи серверов
```bash
# Сервер А
tail -f server-a.log

# Сервер Б
tail -f server-b.log
```

### CloudBridge Relay метрики
```bash
# Health check
curl https://edge.2gc.ru:5552/health | jq .

# Метрики
curl https://edge.2gc.ru:5552/metrics | jq .
```

## 🔧 Конфигурация

### Основные параметры
- **Relay Host**: `edge.2gc.ru:5552`
- **Tenant ID**: `mesh-network-test`
- **Protocol**: QUIC + ICE/STUN/TURN
- **TLS**: Включен (Let's Encrypt)
- **JWT**: HS256 с валидацией

### Сетевые настройки
- **STUN**: `edge.2gc.ru:19302`
- **TURN**: `edge.2gc.ru:3478`
- **QUIC**: Включен с MASQUE поддержкой
- **WebSocket**: Fallback включен

## 🧪 Тестирование

### Проверка подключения
```bash
# Тест токена сервера А
curl -s https://edge.2gc.ru:5552/api/v1/tenants/mesh-network-test/peers/register \
  -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token-server-a>" \
  -d '{"public_key": "test-key", "allowed_ips": ["10.0.0.0/8"]}' | jq .

# Тест токена сервера Б
curl -s https://edge.2gc.ru:5552/api/v1/tenants/mesh-network-test/peers/register \
  -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token-server-b>" \
  -d '{"public_key": "test-key", "allowed_ips": ["10.0.0.0/8"]}' | jq .
```

### Проверка P2P соединения
1. Запустите оба сервера
2. Проверьте логи на наличие сообщений о подключении
3. Убедитесь, что серверы видят друг друга в mesh сети

## 🛠️ Устранение неполадок

### Проблемы с подключением
1. **Проверьте CloudBridge Relay**: `curl https://edge.2gc.ru:5552/health`
2. **Проверьте токены**: Убедитесь, что токены валидны
3. **Проверьте сеть**: Убедитесь, что порты 5552, 19302 доступны

### Проблемы с P2P
1. **Проверьте STUN**: `nc -u edge.2gc.ru 19302`
2. **Проверьте логи**: Ищите ошибки в логах серверов
3. **Проверьте NAT**: Убедитесь, что NAT traversal работает

## 📈 Мониторинг производительности

### Метрики CloudBridge Relay
- **Active Connections**: Количество активных соединений
- **Online Peers**: Количество онлайн пиров
- **Memory Usage**: Использование памяти
- **Uptime**: Время работы

### Логи серверов
- **P2P Events**: События P2P соединений
- **ICE Events**: События NAT traversal
- **QUIC Events**: События QUIC соединений
- **Error Events**: Ошибки и предупреждения

## 🔒 Безопасность

### JWT токены
- **Алгоритм**: HS256
- **Secret**: `85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=`
- **Валидация**: Включена на стороне CloudBridge Relay

### Сетевая безопасность
- **TLS**: Все соединения зашифрованы
- **JWT**: Аутентификация обязательна
- **Rate Limiting**: Готов к настройке
- **Firewall**: Рекомендуется настроить правила

## 🎯 Следующие шаги

1. **Масштабирование**: Добавить больше серверов в mesh
2. **Мониторинг**: Настроить Prometheus + Grafana
3. **Автоматизация**: Настроить автоматическое развертывание
4. **Безопасность**: Настроить более строгие правила безопасности
