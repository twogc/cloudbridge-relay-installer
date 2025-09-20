# 🤖 **AI API Endpoints - Документация**

## 📋 **Обзор**

Документация по AI API endpoints для CloudBridge Relay. Все endpoints доступны через AI Service (порт 8083) и AI Analytics (порт 8084).

---

## 🔧 **AI Service API (порт 8083)**

### **Health и Status**

#### **GET /health**
Проверка состояния AI Service.

**Ответ:**
```json
{
  "service": "ai",
  "status": "healthy"
}
```

#### **GET /metrics**
Prometheus метрики AI Service.

**Ответ:** Prometheus metrics format

#### **GET /api/v1/ai/status**
Получение статуса AI Service и компонентов.

**Ответ:**
```json
{
  "components": {
    "anomaly_detector": true,
    "data_pipeline": true,
    "load_predictor": true,
    "monitoring": true,
    "route_optimizer": true
  },
  "metrics": {
    "anomalies_detected": 50,
    "predictions_made": 200,
    "requests_processed": 1000
  },
  "started": "2025-09-20T13:36:08.136001161Z",
  "status": "running"
}
```

### **AI Models Management**

#### **GET /api/v1/ai/models**
Получение списка доступных AI моделей.

**Ответ:**
```json
{
  "models": [
    {
      "accuracy": 0.95,
      "id": "anomaly-detector",
      "last_trained": "2025-09-19T13:30:42.144868773Z",
      "name": "Network Anomaly Detector",
      "status": "active",
      "type": "isolation_forest"
    },
    {
      "accuracy": 0.85,
      "id": "load-predictor",
      "last_trained": "2025-09-20T01:30:42.144870485Z",
      "name": "Load Predictor",
      "status": "active",
      "type": "arima"
    },
    {
      "accuracy": 0.9,
      "id": "route-optimizer",
      "last_trained": "2025-09-20T07:30:42.144870983Z",
      "name": "Route Optimizer",
      "status": "active",
      "type": "multi_armed_bandit"
    }
  ]
}
```

#### **GET /api/v1/ai/models/{model_id}**
Получение детальной информации о конкретной AI модели.

**Параметры:**
- `model_id` (string) - ID модели

**Ответ:**
```json
{
  "id": "anomaly-detector",
  "name": "Network Anomaly Detector",
  "status": "active",
  "type": "isolation_forest",
  "accuracy": 0.95,
  "last_trained": "2025-09-20T13:30:42Z"
}
```

### **AI Analysis и Predictions**

#### **POST /api/v1/ai/analyze**
Анализ сетевых условий с использованием AI.

**Тело запроса:**
```json
{
  "tenant_id": "string",
  "source_peer": "string",
  "target_peer": "string",
  "data_size": 1024,
  "priority": "medium",
  "network_conditions": {
    "latency": 50,
    "bandwidth": 1000,
    "packet_loss": 0.01
  }
}
```

**Ответ:**
```json
{
  "analysis_id": "uuid",
  "status": "completed",
  "recommendations": [
    "route_optimization",
    "load_balancing"
  ],
  "confidence": 0.85,
  "processing_time_ms": 150
}
```

#### **POST /api/v1/ai/predict**
Создание AI предсказаний для оптимизации сети.

**Тело запроса:**
```json
{
  "tenant_id": "string",
  "prediction_type": "routing",
  "input_data": {
    "network_metrics": {
      "latency": 50,
      "bandwidth": 1000,
      "packet_loss": 0.01
    },
    "historical_data": {}
  }
}
```

**Ответ:**
```json
{
  "prediction_id": "uuid",
  "prediction_type": "routing",
  "result": {
    "optimal_route": "direct",
    "predicted_latency": 45,
    "confidence": 0.9
  },
  "processing_time_ms": 200
}
```

### **AI Optimization**

#### **POST /api/v1/ai/optimize-routing**
AI-оптимизированные рекомендации по маршрутизации.

**Тело запроса:**
```json
{
  "tenant_id": "string",
  "source_peer_id": "string",
  "target_peer_id": "string",
  "data_size": 1024,
  "priority": "medium",
  "network_conditions": {
    "latency": 50,
    "bandwidth": 1000,
    "packet_loss": 0.01
  }
}
```

**Ответ:**
```json
{
  "optimization_id": "uuid",
  "status": "completed",
  "optimal_route": {
    "path": ["peer1", "peer2", "peer3"],
    "latency": 45,
    "bandwidth": 1200,
    "reliability": 0.95
  },
  "ai_recommendations": [
    "use_direct_connection",
    "enable_compression"
  ],
  "confidence": 0.9,
  "processing_time_ms": 180
}
```

### **Anomaly Detection**

#### **POST /api/v1/ai/anomaly-detection**
Обнаружение сетевых аномалий с использованием AI.

**Тело запроса:**
```json
{
  "tenant_id": "string",
  "network_metrics": {
    "latency": 50,
    "bandwidth": 1000,
    "packet_loss": 0.01,
    "cpu_usage": 0.8,
    "memory_usage": 0.6
  },
  "time_window": "5m"
}
```

**Ответ:**
```json
{
  "detection_id": "uuid",
  "status": "completed",
  "anomalies_detected": 2,
  "anomaly_details": [
    {
      "type": "latency_spike",
      "severity": "medium",
      "confidence": 0.85,
      "timestamp": "2025-09-20T13:30:42Z"
    }
  ],
  "recommendations": [
    "check_network_connectivity",
    "monitor_peer_status"
  ]
}
```

### **AI Analytics Integration**

#### **GET /api/v1/ai/analytics**
Получение AI аналитических данных.

**Ответ:**
```json
{
  "analytics": {
    "total_predictions": 1000,
    "accuracy_rate": 0.85,
    "anomalies_detected": 50,
    "optimization_success_rate": 0.9
  },
  "time_range": "24h",
  "last_updated": "2025-09-20T13:30:42Z"
}
```

### **Cadence Integration**

#### **GET /api/v1/ai/cadence/workflows**
Получение AI-связанных Cadence workflows.

**Ответ:**
```json
{
  "workflows": [
    {
      "id": "ai-optimized-routing",
      "name": "AI Optimized Routing Workflow",
      "status": "active",
      "last_execution": "2025-09-20T13:30:42Z"
    }
  ]
}
```

#### **POST /api/v1/ai/cadence/workflows/{workflow_id}/execute**
Выполнение AI Cadence workflow.

**Параметры:**
- `workflow_id` (string) - ID workflow

**Тело запроса:**
```json
{
  "tenant_id": "string",
  "input": {
    "source_peer": "peer1",
    "target_peer": "peer2",
    "optimization_type": "routing"
  }
}
```

**Ответ:**
```json
{
  "workflow_id": "uuid",
  "execution_id": "uuid",
  "status": "started",
  "estimated_completion": "2025-09-20T13:35:42Z"
}
```

---

## 📊 **AI Analytics API (порт 8084)**

### **Health и Status**

#### **GET /health**
Проверка состояния AI Analytics.

**Ответ:**
```json
{
  "service": "ai-analytics",
  "status": "healthy"
}
```

### **Analytics Data**

#### **GET /api/v1/analytics/metrics**
Получение метрик аналитики.

**Ответ:**
```json
{
  "metrics": {
    "total_requests": 1000,
    "success_rate": 0.95,
    "avg_response_time": 150,
    "error_rate": 0.05
  },
  "time_range": "24h",
  "last_updated": "2025-09-20T13:30:42Z"
}
```

#### **GET /api/v1/analytics/reports**
Получение отчетов аналитики.

**Ответ:**
```json
{
  "reports": [
    {
      "id": "daily-report",
      "name": "Daily Analytics Report",
      "date": "2025-09-20",
      "summary": {
        "total_events": 1000,
        "anomalies": 5,
        "optimizations": 50
      }
    }
  ]
}
```

#### **GET /api/v1/analytics/dashboard**
Получение данных дашборда.

**Ответ:**
```json
{
  "dashboard": {
    "widgets": [
      {
        "type": "chart",
        "title": "Network Performance",
        "data": {
          "labels": ["00:00", "06:00", "12:00", "18:00"],
          "datasets": [
            {
              "label": "Latency",
              "data": [50, 45, 55, 48]
            }
          ]
        }
      }
    ]
  }
}
```

### **Real-time Analytics**

#### **GET /api/v1/analytics/real-time**
Получение real-time аналитических данных.

**Ответ:**
```json
{
  "real_time": {
    "active_connections": 100,
    "current_latency": 45,
    "bandwidth_usage": 0.8,
    "anomalies_detected": 2,
    "last_updated": "2025-09-20T13:30:42Z"
  }
}
```

#### **GET /api/v1/analytics/historical**
Получение исторических аналитических данных.

**Ответ:**
```json
{
  "historical": {
    "time_range": "7d",
    "data_points": [
      {
        "timestamp": "2025-09-20T13:30:42Z",
        "latency": 45,
        "bandwidth": 1000,
        "connections": 100
      }
    ]
  }
}
```

#### **GET /api/v1/analytics/performance**
Получение метрик производительности.

**Ответ:**
```json
{
  "performance": {
    "cpu_usage": 0.6,
    "memory_usage": 0.8,
    "disk_usage": 0.4,
    "network_throughput": 1000,
    "response_time": 150
  }
}
```

---

## 🔧 **Использование**

### **Примеры запросов:**

#### **Анализ сетевых условий:**
```bash
curl -X POST "http://localhost:8083/api/v1/ai/analyze" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test-tenant",
    "source_peer": "peer-001",
    "target_peer": "peer-002",
    "data_size": 1024,
    "priority": "medium",
    "network_conditions": {
      "latency": 50,
      "bandwidth": 1000,
      "packet_loss": 0.01
    }
  }'
```

#### **AI-оптимизированная маршрутизация:**
```bash
curl -X POST "http://localhost:8083/api/v1/ai/optimize-routing" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test-tenant",
    "source_peer_id": "peer-001",
    "target_peer_id": "peer-002",
    "data_size": 1024,
    "priority": "medium",
    "network_conditions": {
      "latency": 50,
      "bandwidth": 1000,
      "packet_loss": 0.01
    }
  }'
```

#### **Обнаружение аномалий:**
```bash
curl -X POST "http://localhost:8083/api/v1/ai/anomaly-detection" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "test-tenant",
    "network_metrics": {
      "latency": 50,
      "bandwidth": 1000,
      "packet_loss": 0.01,
      "cpu_usage": 0.8,
      "memory_usage": 0.6
    },
    "time_window": "5m"
  }'
```

---

## 📊 **Метрики**

### **AI Service метрики:**
- `ai_active_alerts` - Количество активных алертов
- `ai_anomalies_detected_total` - Общее количество обнаруженных аномалий
- `ai_cpu_usage_percent` - Использование CPU AI сервисом
- `ai_memory_usage_bytes` - Использование памяти AI сервисом
- `ai_prediction_accuracy` - Точность предсказаний AI
- `ai_predictions_made_total` - Общее количество сделанных предсказаний
- `ai_requests_processed_total` - Общее количество обработанных запросов
- `ai_routes_optimized_total` - Общее количество оптимизированных маршрутов
- `ai_service_uptime_seconds` - Время работы AI сервиса

---

## 🔍 **Отладка**

### **Проверка статуса:**
```bash
# AI Service
curl http://localhost:8083/health
curl http://localhost:8083/api/v1/ai/status

# AI Analytics
curl http://localhost:8084/health
curl http://localhost:8084/api/v1/analytics/metrics
```

### **Логи:**
```bash
# AI Service
kubectl logs deployment/ai-service -n cloudbridge

# AI Analytics
kubectl logs deployment/ai-analytics -n cloudbridge
```

---

**🎯 AI API Endpoints готовы к использованию для интеллектуальной оптимизации P2P сетей!**
