# Техническое задание: Исправление метрик Zabbix для CloudBridge Relay

## Общая информация

**Дата:** 25 сентября 2025
**Приоритет:** Высокий
**Статус:** В работе
**Исполнитель:** Разработчик DevOps/Мониторинг

## Цель

Исправить проблемы с метриками Zabbix агента для CloudBridge Relay, чтобы все метрики из шаблона `cloudbridge-relay-template-fixed.xml` корректно собирались и передавались на Zabbix сервер.

## Выявленные проблемы

### 1. Zabbix агент не запускается
**Симптомы:**
- Ошибка: `cannot load user parameters: user parameter "ai.health" already exists`
- Сервис не может запуститься: `Failed with result 'protocol'`
- PID файл не создается

**Причина:** Дублирующиеся или конфликтующие UserParameters в конфигурации

### 2. Не все метрики работают
**Работающие метрики:**
- Системные: `cloudbridge.system.cpu.usage`, `cloudbridge.system.memory.usage`, `cloudbridge.system.disk.usage`
- Docker: `cloudbridge.docker.containers.running`, `cloudbridge.docker.containers.total`, `cloudbridge.docker.images.total`
- Kubernetes: `cloudbridge.k8s.pods.running`, `cloudbridge.k8s.pods.total`, `cloudbridge.k8s.nodes.ready`

**Неработающие метрики:**
- AI метрики: `ai.status`, `ai.health`, `ai.anomaly_detections`, и др.
- P2P метрики: `p2p.quic.connections.active`, `p2p.quic.peers.active`, и др.
- Relay JSON метрики: `cloudbridge.relay.json.active_connections`, и др.

### 3. Проблемы с правами доступа
- Скрипты P2P метрик недоступны для пользователя zabbix
- Некоторые скрипты возвращают `Permission denied`

## Анализ шаблона

**Файл:** `cloudbridge-relay-template-fixed.xml`
**Всего метрик:** 45+ метрик
**Группы метрик:**
- Relay Server Items (JSON API)
- AI Integration Items
- P2P QUIC Items
- System Items
- Docker Items
- Kubernetes Items

## Задачи для выполнения

### Задача 1: Исправить конфигурацию Zabbix агента
**Приоритет:** Критический

1. **Очистить конфигурацию от дублирующихся ключей:**
   ```bash
   # Проверить и удалить дублирующиеся UserParameters
   grep "UserParameter=" /etc/zabbix/zabbix_agentd.conf | sort | uniq -d
   ```

2. **Создать чистую конфигурацию с правильными UserParameters:**
   - Удалить все закомментированные и дублирующиеся ключи
   - Добавить только необходимые метрики
   - Проверить синтаксис конфигурации

3. **Проверить права доступа к скриптам:**
   ```bash
   # Убедиться, что пользователь zabbix может выполнять скрипты
   sudo -u zabbix /usr/local/bin/ai_metrics_collector.sh health
   sudo -u zabbix /usr/local/bin/p2p_metrics_collector.sh connections_active
   sudo -u zabbix /etc/zabbix/scripts/cloudbridge_relay_json_metrics.sh active_connections
   ```

### Задача 2: Настроить AI метрики
**Приоритет:** Высокий

1. **Проверить скрипт AI метрик:**
   - Файл: `/usr/local/bin/ai_metrics_collector.sh`
   - Убедиться, что все метрики из шаблона поддерживаются

2. **Добавить UserParameters для AI метрик:**
   ```bash
   UserParameter=ai.status,/usr/local/bin/ai_metrics_collector.sh status
   UserParameter=ai.health,/usr/local/bin/ai_metrics_collector.sh health
   UserParameter=ai.anomaly_detections,/usr/local/bin/ai_metrics_collector.sh anomaly_detections
   UserParameter=ai.load_predictions,/usr/local/bin/ai_metrics_collector.sh load_predictions
   UserParameter=ai.routes_optimized,/usr/local/bin/ai_metrics_collector.sh routes_optimized
   UserParameter=ai.processing_time,/usr/local/bin/ai_metrics_collector.sh processing_time
   UserParameter=ai.cache_hits,/usr/local/bin/ai_metrics_collector.sh cache_hits
   UserParameter=ai.cache_misses,/usr/local/bin/ai_metrics_collector.sh cache_misses
   UserParameter=ai.circuit_breaker_state,/usr/local/bin/ai_metrics_collector.sh circuit_breaker_state
   UserParameter=ai.anomaly_detection_accuracy,/usr/local/bin/ai_metrics_collector.sh anomaly_detection_accuracy
   UserParameter=ai.load_prediction_accuracy,/usr/local/bin/ai_metrics_collector.sh load_prediction_accuracy
   UserParameter=ai.route_optimization_accuracy,/usr/local/bin/ai_metrics_collector.sh route_optimization_accuracy
   UserParameter=ai.active_connections,/usr/local/bin/ai_metrics_collector.sh active_connections
   UserParameter=ai.requests_per_second,/usr/local/bin/ai_metrics_collector.sh requests_per_second
   UserParameter=ai.error_rate,/usr/local/bin/ai_metrics_collector.sh error_rate
   ```

### Задача 3: Настроить P2P метрики
**Приоритет:** Высокий

1. **Проверить скрипт P2P метрик:**
   - Файл: `/usr/local/bin/p2p_metrics_collector.sh`
   - Убедиться, что все метрики из шаблона поддерживаются

2. **Добавить UserParameters для P2P метрик:**
   ```bash
   UserParameter=p2p.quic.connections.active,/usr/local/bin/p2p_metrics_collector.sh connections_active
   UserParameter=p2p.quic.connections.total,/usr/local/bin/p2p_metrics_collector.sh connections_total
   UserParameter=p2p.quic.peers.active,/usr/local/bin/p2p_metrics_collector.sh peers_active
   UserParameter=p2p.quic.peers.discovered,/usr/local/bin/p2p_metrics_collector.sh peers_discovered
   UserParameter=p2p.quic.peers.offline,/usr/local/bin/p2p_metrics_collector.sh peers_offline
   UserParameter=p2p.quic.throughput.inbound,/usr/local/bin/p2p_metrics_collector.sh throughput_inbound
   UserParameter=p2p.quic.throughput.outbound,/usr/local/bin/p2p_metrics_collector.sh throughput_outbound
   UserParameter=p2p.quic.latency.average,/usr/local/bin/p2p_metrics_collector.sh latency_average
   UserParameter=p2p.quic.mesh.routes.active,/usr/local/bin/p2p_metrics_collector.sh mesh_routes_active
   UserParameter=p2p.quic.security.ddos.blocked,/usr/local/bin/p2p_metrics_collector.sh ddos_blocked
   UserParameter=p2p.quic.handshake.duration,/usr/local/bin/p2p_metrics_collector.sh handshake_duration
   UserParameter=p2p.quic.packets.sent,/usr/local/bin/p2p_metrics_collector.sh packets_sent
   UserParameter=p2p.quic.packets.received,/usr/local/bin/p2p_metrics_collector.sh packets_received
   UserParameter=p2p.quic.packets.lost,/usr/local/bin/p2p_metrics_collector.sh packets_lost
   UserParameter=p2p.quic.streams.active,/usr/local/bin/p2p_metrics_collector.sh streams_active
   UserParameter=p2p.quic.enhanced.connection.migrations,/usr/local/bin/p2p_metrics_collector.sh connection_migrations
   UserParameter=p2p.quic.enhanced.multipath.connections,/usr/local/bin/p2p_metrics_collector.sh multipath_connections
   UserParameter=p2p.quic.enhanced.zero.rtt.connections,/usr/local/bin/p2p_metrics_collector.sh zero_rtt_connections
   UserParameter=p2p.quic.fallback.events,/usr/local/bin/p2p_metrics_collector.sh fallback_events
   UserParameter=p2p.quic.security.rate.limit.hits,/usr/local/bin/p2p_metrics_collector.sh rate_limit_hits
   UserParameter=p2p.quic.security.whitelist.violations,/usr/local/bin/p2p_metrics_collector.sh whitelist_violations
   ```

### Задача 4: Настроить Relay JSON метрики
**Приоритет:** Средний

1. **Проверить скрипт Relay JSON метрик:**
   - Файл: `/etc/zabbix/scripts/cloudbridge_relay_json_metrics.sh`
   - Убедиться, что все метрики из шаблона поддерживаются

2. **Добавить UserParameters для Relay JSON метрик:**
   ```bash
   UserParameter=cloudbridge.relay.json.active_connections,/etc/zabbix/scripts/cloudbridge_relay_json_metrics.sh active_connections
   UserParameter=cloudbridge.relay.json.connections_total,/etc/zabbix/scripts/cloudbridge_relay_json_metrics.sh connections_total
   UserParameter=cloudbridge.relay.json.uptime_seconds,/etc/zabbix/scripts/cloudbridge_relay_json_metrics.sh uptime_seconds
   UserParameter=cloudbridge.relay.json.memory_alloc_mb,/etc/zabbix/scripts/cloudbridge_relay_json_metrics.sh memory_alloc_mb
   UserParameter=cloudbridge.relay.json.goroutines,/etc/zabbix/scripts/cloudbridge_relay_json_metrics.sh goroutines
   ```

### Задача 5: Тестирование и валидация
**Приоритет:** Высокий

1. **Проверить синтаксис конфигурации:**
   ```bash
   sudo /usr/sbin/zabbix_agentd -t -c /etc/zabbix/zabbix_agentd.conf
   ```

2. **Запустить Zabbix агент:**
   ```bash
   sudo systemctl start zabbix-agent
   sudo systemctl status zabbix-agent
   ```

3. **Протестировать все метрики:**
   ```bash
   # Системные метрики
   zabbix_get -s localhost -k cloudbridge.system.cpu.usage
   zabbix_get -s localhost -k cloudbridge.system.memory.usage
   zabbix_get -s localhost -k cloudbridge.system.disk.usage

   # Docker метрики
   zabbix_get -s localhost -k cloudbridge.docker.containers.running
   zabbix_get -s localhost -k cloudbridge.docker.containers.total
   zabbix_get -s localhost -k cloudbridge.docker.images.total

   # Kubernetes метрики
   zabbix_get -s localhost -k cloudbridge.k8s.pods.running
   zabbix_get -s localhost -k cloudbridge.k8s.pods.total
   zabbix_get -s localhost -k cloudbridge.k8s.nodes.ready

   # AI метрики
   zabbix_get -s localhost -k ai.status
   zabbix_get -s localhost -k ai.health
   zabbix_get -s localhost -k ai.anomaly_detections

   # P2P метрики
   zabbix_get -s localhost -k p2p.quic.connections.active
   zabbix_get -s localhost -k p2p.quic.peers.active
   zabbix_get -s localhost -k p2p.quic.latency.average

   # Relay JSON метрики
   zabbix_get -s localhost -k cloudbridge.relay.json.active_connections
   zabbix_get -s localhost -k cloudbridge.relay.json.uptime_seconds
   zabbix_get -s localhost -k cloudbridge.relay.json.memory_alloc_mb
   ```

4. **Проверить подключение к Zabbix серверу:**
   ```bash
   sudo tail -f /var/log/zabbix/zabbix_agentd.log
   ```

## Критерии приемки

### Обязательные требования:
1. **Zabbix агент запускается без ошибок**
2. **Все метрики из шаблона работают корректно**
3. **Нет дублирующихся UserParameters в конфигурации**
4. **Все скрипты доступны для пользователя zabbix**
5. **Подключение к Zabbix серверу 10.0.0.241 работает**

### Дополнительные требования:
1. **Логи Zabbix агента не содержат ошибок**
2. **Все метрики возвращают корректные значения**
3. **Конфигурация оптимизирована и очищена**
4. **Документация обновлена**

## Технические детали

### Файлы для работы:
- `/etc/zabbix/zabbix_agentd.conf` - основная конфигурация
- `/usr/local/bin/ai_metrics_collector.sh` - скрипт AI метрик
- `/usr/local/bin/p2p_metrics_collector.sh` - скрипт P2P метрик
- `/etc/zabbix/scripts/cloudbridge_relay_json_metrics.sh` - скрипт Relay JSON метрик
- `cloudbridge-relay-template-fixed.xml` - шаблон Zabbix

### Команды для диагностики:
```bash
# Проверка статуса агента
sudo systemctl status zabbix-agent

# Проверка логов
sudo tail -f /var/log/zabbix/zabbix_agentd.log

# Тестирование метрик
zabbix_get -s localhost -k <metric_key>

# Проверка синтаксиса конфигурации
sudo /usr/sbin/zabbix_agentd -t -c /etc/zabbix/zabbix_agentd.conf
```

## Ожидаемый результат

После выполнения всех задач:

1. **Zabbix агент работает стабильно**
2. **Все 45+ метрик из шаблона собираются корректно**
3. **Данные передаются на Zabbix сервер 10.0.0.241**
4. **Мониторинг CloudBridge Relay полностью функционален**

## Критические моменты

1. **Не удалять работающие метрики** (системные, Docker, Kubernetes)
2. **Создать резервную копию конфигурации** перед изменениями
3. **Тестировать каждое изменение** перед применением
4. **Проверить права доступа** для всех скриптов

## Контакты

При возникновении вопросов или проблем обращаться к команде DevOps.

---

**Статус:** Ожидает выполнения
**Срок выполнения:** 1-2 дня
**Приоритет:** Высокий
