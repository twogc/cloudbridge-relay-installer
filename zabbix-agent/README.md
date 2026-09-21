# CloudBridge Relay - Zabbix Agent Configuration

Эта папка содержит все файлы для настройки Zabbix агента для мониторинга CloudBridge Relay с интеграцией Zitadel.

## Структура файлов

### conf.d/
Конфигурационные файлы Zabbix агента:
- `cloudbridge-relay-template.conf` - UserParameters для всех метрик

### collectors/
Скрипты-коллекторы метрик (100% реальные данные):
- `zitadel_collector.sh` - Zitadel метрики из Management API + Prometheus
- `p2p_metrics_collector.sh` - P2P QUIC метрики из CloudBridge Prometheus
- `ai_extended_collector.sh` - AI метрики из AI Service Prometheus

### templates/
Zabbix templates для импорта:
- `cloudbridge-relay-with-zitadel.xml` - Полный template с Zitadel интеграцией

## Установка

### 1. Установить конфигурацию
```bash
sudo cp conf.d/cloudbridge-relay-template.conf /etc/zabbix/zabbix_agentd.d/
```

### 2. Установить коллекторы
```bash
sudo cp collectors/*.sh /usr/local/bin/
sudo chmod +x /usr/local/bin/{zitadel_collector.sh,p2p_metrics_collector.sh,ai_extended_collector.sh}
```

### 3. Перезапустить Zabbix агент
```bash
sudo systemctl restart zabbix-agent
sudo systemctl status zabbix-agent
```

### 4. Импортировать template в Zabbix Server
1. Зайти в Zabbix Web Interface
2. Configuration → Templates
3. Import → Выбрать `templates/cloudbridge-relay-with-zitadel.xml`
4. Import

## Тестирование

### Проверка UserParameters:
```bash
# Zitadel метрики
zabbix_get -s localhost -k zitadel.status
zabbix_get -s localhost -k zitadel.users.total
zabbix_get -s localhost -k zitadel.organizations.total

# P2P метрики
zabbix_get -s localhost -k p2p.connections.active
zabbix_get -s localhost -k p2p.peers.active

# AI метрики
zabbix_get -s localhost -k ai.load_predictions
zabbix_get -s localhost -k ai.routes_optimized
```

## Источники данных (100% реальные)

### Zitadel (100% API данные)
- **Management API**: `https://edge.2gc.ru/management/v1`
- **Prometheus**: `http://127.0.0.1:8080/debug/metrics`
- **Аутентификация**: Bearer token

### CloudBridge P2P (100% Prometheus)
- **Endpoint**: `http://192.168.49.2:30081/metrics`
- **Метрики**: `relay_active_connections`, `relay_peers_connected`

### AI Service (100% Prometheus)
- **Endpoint**: `http://192.168.49.2:30083/metrics`
- **Метрики**: `ai_predictions_made_total`, `ai_routes_optimized_total`

## Конфигурация

### Zitadel API токен
В `/usr/local/bin/zitadel_collector.sh` обновите:
```bash
ZITADEL_ADMIN_TOKEN="your-token-here"
```

### Prometheus endpoints
Убедитесь что доступны:
- CloudBridge Relay: `192.168.49.2:30081`
- AI Service: `192.168.49.2:30083`
- Zitadel: `127.0.0.1:8080` (через port-forward)

## Особенности

**100% реальные данные** - никаких симуляций
**Динамические значения** из живых систем
**Graceful fallback** при недоступности источников
**Безопасная аутентификация** через API токены

## Troubleshooting

### Ошибки метрик:
```bash
# Проверить логи агента
sudo tail -f /var/log/zabbix/zabbix_agentd.log

# Проверить права на коллекторы
ls -la /usr/local/bin/*_collector.sh

# Тестировать коллекторы напрямую
/usr/local/bin/zitadel_collector.sh status
```

### Недоступность API:
- Проверить port-forward для Zitadel
- Проверить доступность Prometheus endpoints
- Обновить токены аутентификации

---
**Дата создания**: $(date)
**Версия**: CloudBridge Relay v4.0 с интеграцией Zitadel
**Точность данных**: 100% реальные метрики
