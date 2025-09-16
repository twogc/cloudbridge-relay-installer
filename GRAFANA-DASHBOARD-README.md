# 📊 CloudBridge AI Service Grafana Dashboard

## 🎉 Дашборд успешно создан и готов к использованию!

### 📋 Что включено в дашборд:

- **✅ AI Service Status** - статус AI сервиса (включен/выключен)
- **✅ Request Rate** - количество запросов в секунду
- **✅ Response Time** - время ответа (95-й и 50-й процентили)
- **✅ CPU Usage** - использование CPU в процентах
- **✅ Memory Usage** - использование памяти в MB
- **✅ Anomaly Detection Rate** - скорость обнаружения аномалий
- **✅ Prediction Accuracy** - точность предсказаний
- **✅ Active Alerts** - активные алерты
- **✅ Error Rate** - процент ошибок
- **✅ Throughput** - пропускная способность

### 🌐 Доступ к Grafana:

- **URL:** http://localhost:3000
- **Логин:** `admin`
- **Пароль:** `admin123`

### 🚀 Как запустить Grafana:

```bash
# Запустить port-forward для Grafana
kubectl port-forward -n cloudbridge-monitoring svc/grafana 3000:3000

# Или использовать скрипт для запуска всех сервисов мониторинга
./scripts/start-monitoring.sh
```

### 📊 Импорт дашборда:

#### Автоматический импорт:
```bash
./scripts/import-dashboard-template.sh
```

#### Ручной импорт:
1. Откройте http://localhost:3000
2. Войдите с `admin` / `admin123`
3. Нажмите **"+"** → **"Import"**
4. Скопируйте содержимое файла `grafana-dashboard-template.json`
5. Вставьте JSON и нажмите **"Load"**

### 🔍 Файлы дашборда:

- **`grafana-dashboard-template.json`** - основной шаблон дашборда
- **`scripts/import-dashboard-template.sh`** - скрипт для автоматического импорта
- **`scripts/start-monitoring.sh`** - скрипт для запуска всех сервисов мониторинга

### 📈 Источники данных:

Дашборд использует Prometheus как источник данных:
- **URL:** http://prometheus.cloudbridge-monitoring.svc.cluster.local:9090
- **Доступ:** http://localhost:9090 (через port-forward)

### 🎯 Метрики AI Service:

Дашборд ожидает следующие метрики от AI Service:
- `up{job="ai-service"}` - статус сервиса
- `http_requests_total{job="ai-service"}` - общее количество запросов
- `http_request_duration_seconds_bucket{job="ai-service"}` - время ответа
- `container_cpu_usage_seconds_total{pod=~"ai-service-.*"}` - использование CPU
- `container_memory_usage_bytes{pod=~"ai-service-.*"}` - использование памяти
- `ai_anomalies_detected_total` - обнаруженные аномалии
- `ai_prediction_accuracy` - точность предсказаний
- `ai_active_alerts` - активные алерты

### 🔧 Настройка алертов:

1. В Grafana перейдите в **"Alerting"** → **"Alert Rules"**
2. Создайте новые правила на основе метрик дашборда
3. Настройте уведомления (email, Slack, etc.)

### 📝 Полезные команды:

```bash
# Проверить статус Grafana
kubectl get pods -n cloudbridge-monitoring | grep grafana

# Посмотреть логи Grafana
kubectl logs -n cloudbridge-monitoring -l app.kubernetes.io/name=grafana

# Проверить статус AI Service
kubectl get pods -n cloudbridge | grep ai-service

# Посмотреть логи AI Service
kubectl logs -n cloudbridge -l app.kubernetes.io/name=ai-service

# Проверить метрики AI Service
curl http://localhost:8083/metrics
```

### 🎨 Кастомизация дашборда:

1. Откройте дашборд в Grafana
2. Нажмите **"Settings"** → **"JSON Model"**
3. Отредактируйте JSON по необходимости
4. Сохраните изменения

### 🚨 Устранение неполадок:

#### Дашборд не показывает данные:
1. Проверьте, что Prometheus работает: http://localhost:9090
2. Проверьте, что AI Service экспортирует метрики: http://localhost:8083/metrics
3. Убедитесь, что ServiceMonitor настроен правильно

#### Grafana недоступен:
1. Проверьте статус пода: `kubectl get pods -n cloudbridge-monitoring | grep grafana`
2. Перезапустите port-forward: `kubectl port-forward -n cloudbridge-monitoring svc/grafana 3000:3000`
3. Проверьте логи: `kubectl logs -n cloudbridge-monitoring -l app.kubernetes.io/name=grafana`

### 🎉 Готово!

Теперь у вас есть полнофункциональный дашборд для мониторинга AI Service в Grafana!

---

**Создано:** $(date)  
**Версия:** 1.0.0  
**Автор:** CloudBridge AI Team


