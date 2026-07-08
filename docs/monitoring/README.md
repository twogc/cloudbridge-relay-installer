# **Мониторинг CloudBridge Relay**

## **Обзор**

Этот раздел содержит документы по мониторингу CloudBridge Relay, включая метрики, логирование и настройку UFW.

## **Документы**

### **Анализ и отчеты:**
- [**Port Analysis Report**](PORT_ANALYSIS_REPORT.md) - Анализ портов и их использования
- [**HTTP3 Testing Report**](HTTP3_TESTING_REPORT.md) - Результаты тестирования HTTP3
- [**UFW Firewall Guide**](../UFW_FIREWALL_GUIDE.md) - Руководство по настройке UFW

## **Система мониторинга**

### **Prometheus:**
- **Сбор метрик** - Автоматический сбор метрик
- **Хранение данных** - Временные ряды
- **Запросы** - PromQL для анализа
- **Алерты** - Уведомления о проблемах

### **Grafana:**
- **Дашборды** - Визуализация метрик
- **Графики** - Временные ряды
- **Алерты** - Уведомления в реальном времени
- **Экспорт** - PDF отчеты

### **UFW Firewall:**
- **Логирование** - Логи сетевого трафика
- **Мониторинг** - Отслеживание подключений
- **Безопасность** - Защита от атак

##  **Метрики CloudBridge Relay**

### **Основные метрики:**
- **relay_connections_total** - Общее количество соединений
- **relay_connections_active** - Активные соединения
- **relay_bandwidth_bytes** - Пропускная способность
- **relay_latency_seconds** - Задержка
- **relay_errors_total** - Количество ошибок
- **relay_requests_total** - Количество запросов

### **Системные метрики:**
- **go_gc_duration_seconds** - Время сборки мусора
- **go_memstats_alloc_bytes** - Использование памяти
- **go_goroutines** - Количество горутин
- **process_cpu_seconds_total** - Использование CPU

## **Порты мониторинга**

### **CloudBridge Relay порты:**
| Порт | Протокол | Назначение | Описание |
|------|----------|------------|----------|
| 5550 | TCP | Main Server | Основной TCP сервер |
| 5551 | TCP | Metrics | Prometheus метрики |
| 5552 | TCP | P2P API | P2P API для управления |
| 5553 | TCP | HTTP API | HTTP API для клиентов |
| 5554 | TCP | HTTPS API | HTTPS API с TLS |
| 5555 | UDP | QUIC Transport | QUIC транспорт |

### **Дополнительные порты:**
| Порт | Протокол | Назначение | Описание |
|------|----------|------------|----------|
| 9090 | UDP | QUIC Legacy | Legacy QUIC |
| 9092 | UDP | Enhanced QUIC | Расширенный QUIC |
| 19302 | UDP | STUN Server | STUN сервер |
| 8443 | UDP | MASQUE | MASQUE протокол |

## 🛠️ **Настройка мониторинга**

### **Prometheus конфигурация:**
```yaml
scrape_configs:
  - job_name: 'cloudbridge-relay'
    static_configs:
      - targets: ['cloudbridge-relay.cloudbridge.svc.cluster.local:5551']
    metrics_path: '/metrics'
    scrape_interval: 10s
```

### **UFW настройка:**
```bash
# Добавить правила для мониторинга
sudo ufw allow 5551/tcp comment "CloudBridge Relay - Metrics"
sudo ufw allow 9090/tcp comment "Prometheus"
sudo ufw allow 3000/tcp comment "Grafana"
```

##  **Дашборды Grafana**

### **CloudBridge Relay Dashboard:**
- **Connections** - График соединений
- **Bandwidth** - Пропускная способность
- **Latency** - Задержка
- **Errors** - Количество ошибок
- **System Resources** - Системные ресурсы

### **Kubernetes Dashboard:**
- **Pod Status** - Статус подов
- **Resource Usage** - Использование ресурсов
- **Network Traffic** - Сетевой трафик
- **Logs** - Логи приложений

## **Диагностика**

### **Проверка метрик:**
```bash
# Проверка Prometheus
curl http://localhost:9090/api/v1/targets

# Проверка метрик CloudBridge Relay
curl http://localhost:5551/metrics

# Проверка Grafana
curl http://localhost:3000/api/health
```

### **Проверка портов:**
```bash
# Проверка TCP портов
netstat -tlnp | grep -E "555[0-5]"

# Проверка UDP портов
netstat -ulnp | grep -E "555[0-5]|909[0-2]|19302|8443"
```

## **Чек-лист мониторинга**

### **Prometheus:**
- [ ] Prometheus запущен
- [ ] Targets доступны
- [ ] Метрики собираются
- [ ] Алерты настроены

### **Grafana:**
- [ ] Grafana доступен
- [ ] Дашборды созданы
- [ ] Источники данных настроены
- [ ] Алерты работают

### **UFW:**
- [ ] UFW включен
- [ ] Правила настроены
- [ ] Логирование включено
- [ ] Мониторинг работает

### **CloudBridge Relay:**
- [ ] Метрики экспортируются
- [ ] Health checks работают
- [ ] Логи корректны
- [ ] Производительность в норме

---

**Мониторинг CloudBridge Relay настроен и работает!**
