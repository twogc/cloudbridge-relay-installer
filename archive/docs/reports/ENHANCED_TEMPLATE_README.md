# CloudBridge Relay Enhanced Template

## ОБЗОР

Расширенный шаблон Zabbix для мониторинга CloudBridge Relay с полным покрытием современных сетевых технологий:
- **MASQUE Protocol** - HTTP/3 туннелирование
- **ICE/STUN/TURN** - NAT traversal мониторинг  
- **Enhanced QUIC** - детальная производительность
- **AI Analytics** - интеллектуальная аналитика соединений
- **Advanced Security** - расширенный мониторинг безопасности
- **Performance QoS** - контроль качества соединений

---

## СТАТИСТИКА УЛУЧШЕНИЙ

| Компонент | Оригинал | Enhanced | Прирост |
|-----------|----------|----------|---------|
| **Items** | 54 | 87 | +33 (+61%) |
| **Triggers** | 28 | 52 | +24 (+86%) |
| **Macros** | 22 | 45 | +23 (+105%) |
| **Lines** | 1,373 | 2,374 | +1,001 (+73%) |

---

## 🆕 НОВЫЕ СЕКЦИИ

### **MASQUE Protocol** (6 items + 3 triggers)
- `masque.tunnels.active` - Активные туннели
- `masque.connect.udp.requests.total` - CONNECT-UDP запросы
- `masque.connect.ip.requests.total` - CONNECT-IP запросы  
- `masque.tunnel.establishment.time.seconds` - Время установки туннеля
- `websocket.fallback.connections.active` - WebSocket fallback
- `tcp.fallback.connections.active` - TCP fallback

### 🔄 **NAT Traversal** (7 items + 5 triggers)
- `ice.candidate.gathering.time.seconds` - ICE gathering time
- `ice.connection.success.rate` - Успешность ICE соединений
- `stun.binding.success.rate` - Успешность STUN
- `stun.rtt.seconds` - STUN RTT
- `turn.allocations.active` - Активные TURN аллокации
- `turn.relayed.bytes.inbound` - TURN трафик
- `nat.traversal.success.rate` - Общая успешность NAT

### **Enhanced QUIC** (6 items + 3 triggers)
- `quic.rtt.p50.seconds` - RTT 50-й перцентиль
- `quic.rtt.p95.seconds` - RTT 95-й перцентиль
- `quic.congestion.window.bytes` - Congestion window
- `quic.bytes.in.flight` - Данные в полете
- `quic.stream.count.bidirectional` - Bidirectional streams
- `quic.stream.count.unidirectional` - Unidirectional streams

### 🤖 **AI Analytics** (4 items + 3 triggers)
- `ai.connection.anomaly.score` - Оценка аномальности
- `ai.route.prediction.accuracy` - Точность предсказания маршрутов
- `ai.bandwidth.prediction.accuracy` - Точность предсказания полосы
- `ai.routing.optimization.effect.percent` - Эффект оптимизации

### **Security** (5 items + 4 triggers)
- `tls.handshake.duration.seconds` - Время TLS handshake
- `jwt.validation.duration.seconds` - Время валидации JWT
- `security.policy.violations.total` - Нарушения политик
- `geo.blocked.connections.total` - Геоблокировки
- `suspicious.connection.patterns.total` - Подозрительные паттерны

### **Performance** (5 items + 7 triggers)
- `connection.quality.score` - Оценка качества соединения
- `bandwidth.utilization.inbound.percent` - Входящая утилизация
- `bandwidth.utilization.outbound.percent` - Исходящая утилизация
- `connection.jitter.seconds` - Джиттер соединений
- `latency.sla.compliance.percent` - SLA соответствие

---

## 🎛️ НОВЫЕ МАКРОСЫ (23 шт)

### MASQUE Thresholds
- `{$CB.MASQUE.TUNNELS.MIN}` = 1 - Минимум туннелей
- `{$CB.MASQUE.ESTABLISHMENT.WARN}` = 2000ms - Предупреждение
- `{$CB.MASQUE.ESTABLISHMENT.HIGH}` = 5000ms - Критично

### ICE/STUN/TURN Thresholds  
- `{$CB.ICE.SUCCESS.WARN}` = 80% - Предупреждение успешности ICE
- `{$CB.STUN.RTT.WARN}` = 200ms - Предупреждение STUN RTT
- `{$CB.NAT.SUCCESS.WARN}` = 75% - Предупреждение NAT traversal

### AI Analytics Thresholds
- `{$CB.AI.ANOMALY.SCORE.WARN}` = 0.7 - Предупреждение аномалий
- `{$CB.AI.PREDICTION.ACCURACY.MIN}` = 80% - Минимальная точность

### Performance Thresholds
- `{$CB.CONNECTION.QUALITY.MIN}` = 7 - Минимальное качество
- `{$CB.BANDWIDTH.UTIL.WARN}` = 85% - Предупреждение утилизации
- `{$CB.SLA.COMPLIANCE.MIN}` = 95% - Минимальное SLA

---

## 📥 УСТАНОВКА

1. **Импорт в Zabbix:**
   ```bash
   # Загрузить файл через Web UI:
   Configuration → Templates → Import
   # Выбрать: cloudbridge-relay-template-enhanced.xml
   ```

2. **Связать с хостом:**
   ```bash
   Configuration → Hosts → [Your CloudBridge Host] → Templates
   # Добавить: "CloudBridge Relay Template Enhanced"
   ```

3. **Создать коллекторы:**
   ```bash
   # Используйте существующие коллекторы:
   /usr/local/bin/p2p_metrics_collector.sh
   /usr/local/bin/ai_extended_collector.sh
   /usr/local/bin/ai_stats_collector.sh
   ```

---

## НОВЫЕ АЛЕРТЫ

### Критичные
- `MASQUE: No Active Tunnels` - Нет активных MASQUE туннелей
- `ICE: Critical Connection Success Rate` - Критическая успешность ICE
- `AI: Critical Anomaly Score Detected` - Критическая аномалия
- `Security: Critical Policy Violations Rate` - Критические нарушения

### Предупреждения
- `QUIC: High P95 RTT` - Высокая 95-я перцентиль RTT
- `Performance: Low Connection Quality` - Низкое качество соединений
- `NAT: Low Traversal Success Rate` - Низкая успешность NAT

---

## ИСПОЛЬЗОВАНИЕ

### Dashboard Configuration
1. **Executive Level**: Connection quality, security incidents
2. **Operations Level**: Protocol performance, AI analytics
3. **Engineering Level**: Detailed QUIC metrics, NAT traversal

### Monitoring Strategy
1. **Real-time**: Connection quality, anomaly detection
2. **Trend Analysis**: AI prediction accuracy, bandwidth utilization
3. **Capacity Planning**: QUIC streams, TURN allocations

---

## 🛠️ СОЗДАНИЕ КОЛЛЕКТОРОВ

Для новых метрик необходимо создать соответствующие коллекторы:

```bash
# Пример MASQUE коллектора:
sudo tee /usr/local/bin/masque_collector.sh > /dev/null << 'EOF'
#!/bin/bash
METRIC_TYPE="$1"
case "$METRIC_TYPE" in
    "tunnels_active")
        # Логика получения количества активных MASQUE туннелей
        echo "2"  # Заглушка
        ;;
    *)
        echo "0"
        ;;
esac
EOF

sudo chmod +x /usr/local/bin/masque_collector.sh
```

---

## 📞 ПОДДЕРЖКА

Для вопросов по расширенному шаблону:
- Техническая документация: `SYSTEM_OVERVIEW.md`
- Быстрый старт: `QUICK_START_GUIDE.md`
- Метрики валидация: `TASK_METRICS_VALIDATION.md`

---

**Готово! Теперь у вас есть enterprise-grade мониторинг CloudBridge Relay!** 🚀

