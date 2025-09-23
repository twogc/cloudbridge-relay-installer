# Чек-лист тестирования AI сервиса

## 🚀 Быстрый старт

```bash
cd /home/ubuntu/cloudbridge-relay-installer
./scripts/test_real_ai_data.sh
```

## ✅ Проверочный список

### 1. Проверка сервисов
- [ ] CloudBridge Relay запущен
- [ ] AI Service запущен  
- [ ] Prometheus запущен
- [ ] Cadence запущен

### 2. Проверка портов
- [ ] CloudBridge Relay: 5551 (метрики)
- [ ] AI Service: 8080 (API)
- [ ] Prometheus: 9090 (API)
- [ ] Cadence: 8080 (API)

### 3. Проверка метрик
```bash
# CloudBridge Relay метрики
kubectl port-forward -n cloudbridge svc/cloudbridge-relay 5551:5551 &
curl -s "http://localhost:5551/metrics" | grep relay_active_connections

# AI Service health
kubectl port-forward -n cloudbridge svc/ai-service 8080:8080 &
curl -s "http://localhost:8080/health"

# AI Service статус
curl -s "http://localhost:8080/api/v1/ai/status" | jq .
```

### 4. Тестирование AI функций

#### Обнаружение аномалий
```bash
curl -X POST "http://localhost:8080/api/v1/ai/anomaly" \
  -H "Content-Type: application/json" \
  -d '{
    "connections": 100,
    "bandwidth": 1000,
    "latency": 50,
    "cpu_usage": 75,
    "memory_usage": 60
  }'
```

#### Прогнозирование нагрузки
```bash
curl -X POST "http://localhost:8080/api/v1/ai/predict" \
  -H "Content-Type: application/json" \
  -d '{"horizon": 24}'
```

#### Оптимизация маршрутов
```bash
curl -X POST "http://localhost:8080/api/v1/ai/optimize" \
  -H "Content-Type: application/json" \
  -d '{
    "source": "node1",
    "destination": "node2",
    "constraints": {
      "max_latency": 100,
      "min_bandwidth": 1000
    }
  }'
```

### 5. Проверка логов
```bash
# AI Service логи
kubectl logs -n cloudbridge deployment/ai-service -c ai-service --tail=20

# CloudBridge Relay логи
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=20
```

### 6. Проверка конфигурации
```bash
# AI конфигурация
kubectl get configmap ai-config -n cloudbridge -o yaml

# CloudBridge конфигурация
kubectl get configmap cloudbridge-config -n cloudbridge -o yaml
```

## 🔧 Устранение проблем

### Проблема: AI Service не отвечает
```bash
# Перезапуск AI сервиса
kubectl rollout restart deployment/ai-service -n cloudbridge
kubectl rollout status deployment/ai-service -n cloudbridge
```

### Проблема: Неправильные порты
```bash
# Проверка сервисов
kubectl get svc -n cloudbridge

# Проверка подов
kubectl get pods -n cloudbridge
```

### Проблема: Ошибки подключения
```bash
# Проверка сетевой связности
kubectl exec -n cloudbridge deployment/ai-service -- nslookup cloudbridge-relay.cloudbridge.svc.cluster.local
kubectl exec -n cloudbridge deployment/ai-service -- nslookup prometheus.cloudbridge-monitoring.svc.cluster.local
```

## 📊 Ожидаемые результаты

### Успешный ответ AI Service
```json
{
  "anomaly_detector": true,
  "load_predictor": true,
  "route_optimizer": true,
  "monitoring": true,
  "data_pipeline": true
}
```

### Успешное обнаружение аномалий
```json
{
  "is_anomaly": false,
  "score": 0.3,
  "confidence": 0.85,
  "anomaly_type": "normal",
  "timestamp": "2025-09-22T22:00:00Z"
}
```

### Успешное прогнозирование
```json
{
  "timestamp": "2025-09-22T22:00:00Z",
  "horizon": 24,
  "confidence": 0.82,
  "model_id": "real_data_ensemble"
}
```

## 🎯 Критерии успеха

- [ ] Все сервисы отвечают на health check
- [ ] AI API возвращает корректные ответы
- [ ] Метрики собираются из реальных источников
- [ ] Нет ошибок в логах
- [ ] Конфигурация применена корректно
- [ ] Мониторинг работает

## 📝 Отчет о тестировании

После выполнения всех проверок создайте отчет:

```markdown
# Отчет о тестировании AI сервиса

**Дата:** [дата]
**Версия:** [версия]
**Тестировщик:** [имя]

## Результаты тестирования

### Сервисы
- CloudBridge Relay: ✅/❌
- AI Service: ✅/❌  
- Prometheus: ✅/❌
- Cadence: ✅/❌

### API функции
- Обнаружение аномалий: ✅/❌
- Прогнозирование нагрузки: ✅/❌
- Оптимизация маршрутов: ✅/❌

### Проблемы
- [описание проблем]

### Рекомендации
- [рекомендации по улучшению]
```
