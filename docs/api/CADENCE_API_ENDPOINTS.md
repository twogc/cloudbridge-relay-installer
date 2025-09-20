# ⚡ **Cadence API Endpoints - Документация**

## 📋 **Обзор**

Документация по Cadence API endpoints для CloudBridge Relay. Cadence Workflow Engine обеспечивает оркестрацию P2P QUIC workflows и AI-оптимизированной маршрутизации.

---

## 🔧 **Cadence Web UI (порт 8088)**

### **Доступ к Web UI**

#### **Web Interface**
Cadence Web UI доступен через port forwarding:

```bash
kubectl port-forward service/cadence-web-service 8088:8088 -n cloudbridge
```

**URL:** `http://localhost:8088`

**Функции:**
- Просмотр workflows и их статуса
- Мониторинг выполнения activities
- Управление доменами и task lists
- Просмотр истории выполнения
- Отладка и диагностика

---

## 🔧 **Cadence Service API (порт 7933)**

### **Workflow Management**

#### **P2P QUIC Workflows**

##### **P2PQUICTunnelWorkflow**
Управление P2P QUIC туннелями.

**Входные параметры:**
```json
{
  "tenant_id": "string",
  "peer_id": "string",
  "operation": "create|update|delete|monitor",
  "tunnel_config": {
    "interface": "wg0",
    "private_key": "string",
    "public_key": "string",
    "allowed_ips": ["10.0.0.0/24"],
    "endpoint": "peer.example.com:51820",
    "persistent_keepalive": 25
  },
  "network_config": {
    "mtu": 1420,
    "compression": true,
    "encryption": true
  }
}
```

**Результат:**
```json
{
  "tenant_id": "string",
  "peer_id": "string",
  "operation": "create",
  "status": "completed",
  "tunnel_id": "tunnel-001",
  "result": {
    "tunnel_id": "tunnel-001",
    "status": "created",
    "created_at": "2025-09-20T13:30:42Z",
    "config": {}
  },
  "started_at": "2025-09-20T13:30:42Z",
  "completed_at": "2025-09-20T13:30:45Z",
  "message": "P2P QUIC туннель успешно обработан"
}
```

##### **P2PConnectionOrchestrationWorkflow**
Оркестрация P2P соединений.

**Входные параметры:**
```json
{
  "tenant_id": "string",
  "source_peer_id": "string",
  "target_peer_id": "string",
  "connection_type": "direct|relay|masque",
  "protocol": "quic|websocket|tcp",
  "configuration": {
    "timeout": 30,
    "retry_attempts": 3,
    "compression": true,
    "encryption": true
  }
}
```

**Результат:**
```json
{
  "tenant_id": "string",
  "source_peer_id": "string",
  "target_peer_id": "string",
  "connection_id": "conn-001",
  "status": "completed",
  "connection_type": "direct",
  "protocol": "quic",
  "latency": 45000000,
  "throughput": 2000,
  "result": {
    "connection_id": "conn-001",
    "status": "established",
    "latency": 45,
    "throughput": 2000,
    "established_at": "2025-09-20T13:30:42Z"
  },
  "started_at": "2025-09-20T13:30:42Z",
  "completed_at": "2025-09-20T13:30:45Z",
  "message": "P2P соединение успешно установлено"
}
```

##### **AIOptimizedRoutingWorkflow**
AI-оптимизированная маршрутизация.

**Входные параметры:**
```json
{
  "tenant_id": "string",
  "source_peer_id": "string",
  "target_peer_id": "string",
  "data_size": 1024,
  "priority": "low|medium|high|critical",
  "network_conditions": {
    "latency": 50,
    "bandwidth": 1000,
    "packet_loss": 0.01,
    "cpu_usage": 0.8,
    "memory_usage": 0.6
  }
}
```

**Результат:**
```json
{
  "tenant_id": "string",
  "source_peer_id": "string",
  "target_peer_id": "string",
  "status": "completed",
  "optimal_route": {
    "path": ["peer-001", "peer-002"],
    "latency": 45,
    "bandwidth": 1200,
    "reliability": 0.95
  },
  "predicted_latency": 45000000,
  "predicted_throughput": 1200,
  "confidence": 0.9,
  "ai_recommendations": [
    "use_direct_connection",
    "enable_compression"
  ],
  "result": {
    "analysis_id": "uuid",
    "status": "completed",
    "recommendations": ["route_optimization"],
    "confidence": 0.9,
    "processing_time_ms": 180
  },
  "started_at": "2025-09-20T13:30:42Z",
  "completed_at": "2025-09-20T13:30:45Z",
  "message": "AI-оптимизированная маршрутизация завершена"
}
```

### **Activities**

#### **P2P QUIC Activities**

##### **CreateP2PQUICTunnelActivity**
Создание P2P QUIC туннеля.

**Параметры:**
- `tenant_id` (string)
- `peer_id` (string)
- `config` (map[string]interface{})

**Результат:**
```json
{
  "tunnel_id": "tunnel-001",
  "tenant_id": "string",
  "peer_id": "string",
  "status": "created",
  "created_at": "2025-09-20T13:30:42Z",
  "config": {}
}
```

##### **MonitorP2PQUICTunnelActivity**
Мониторинг P2P QUIC туннеля.

**Параметры:**
- `tenant_id` (string)
- `peer_id` (string)

**Результат:**
```json
{
  "tenant_id": "string",
  "peer_id": "string",
  "status": "active",
  "bytes_in": 1024,
  "bytes_out": 2048,
  "connections": 5,
  "last_activity": "2025-09-20T13:30:42Z",
  "latency": 50,
  "throughput": 1000
}
```

##### **EstablishP2PConnectionActivity**
Установка P2P соединения.

**Параметры:**
- `tenant_id` (string)
- `source_peer_id` (string)
- `target_peer_id` (string)
- `connection_type` (string)
- `protocol` (string)

**Результат:**
```json
{
  "connection_id": "conn-001",
  "tenant_id": "string",
  "source_peer_id": "string",
  "target_peer_id": "string",
  "connection_type": "direct",
  "protocol": "quic",
  "status": "established",
  "latency": 45,
  "throughput": 2000,
  "established_at": "2025-09-20T13:30:42Z"
}
```

##### **AnalyzeNetworkConditionsActivity**
Анализ сетевых условий через AI.

**Параметры:**
- `tenant_id` (string)
- `network_conditions` (map[string]interface{})

**Результат:**
```json
{
  "analysis_id": "uuid",
  "status": "completed",
  "network_health": 0.85,
  "bottlenecks": ["bandwidth", "latency"],
  "recommendations": ["increase_bandwidth", "optimize_routing"],
  "confidence": 0.9,
  "processing_time_ms": 150
}
```

##### **PredictOptimalRouteActivity**
Предсказание оптимального маршрута.

**Параметры:**
- `tenant_id` (string)
- `source_peer_id` (string)
- `target_peer_id` (string)
- `data_size` (int64)
- `priority` (string)
- `network_analysis` (map[string]interface{})

**Результат:**
```json
{
  "prediction_id": "uuid",
  "optimal_route": {
    "path": ["peer-001", "peer-002"],
    "latency": 45,
    "bandwidth": 1200,
    "reliability": 0.95
  },
  "predicted_latency": 45,
  "predicted_throughput": 1200,
  "confidence": 0.9,
  "processing_time_ms": 200
}
```

##### **GenerateAIRecommendationsActivity**
Генерация AI рекомендаций.

**Параметры:**
- `tenant_id` (string)
- `analysis_data` (map[string]interface{})

**Результат:**
```json
{
  "recommendations_id": "uuid",
  "recommendations": [
    "use_direct_connection",
    "enable_compression",
    "optimize_bandwidth"
  ],
  "priority": "high",
  "confidence": 0.9,
  "processing_time_ms": 100
}
```

---

## 🔧 **Cadence Client API**

### **Workflow Execution**

#### **Execute Workflow**
```go
// Пример использования Cadence Client
workflowOptions := cadence.StartWorkflowOptions{
    ID:        "tunnel-creation-tenant-001-peer-001",
    TaskList:  "p2p-quic-tasks",
    ExecutionTimeout: 10 * time.Minute,
}

input := workflows.P2PQUICTunnelWorkflowInput{
    TenantID:      "tenant-001",
    PeerID:        "peer-001",
    Operation:     "create",
    TunnelConfig:  config,
    NetworkConfig: networkConfig,
}

execution, err := cadenceClient.ExecuteWorkflow(ctx, workflowOptions, workflows.P2PQUICTunnelWorkflow, input)
```

#### **Get Workflow Status**
```go
// Получение статуса workflow
workflowRun, err := cadenceClient.GetWorkflow(ctx, workflowID, "")
if err != nil {
    return err
}

var result workflows.P2PQUICTunnelWorkflowResult
err = workflowRun.Get(ctx, &result)
```

#### **Cancel Workflow**
```go
// Отмена workflow
err := cadenceClient.CancelWorkflow(ctx, workflowID, "")
```

### **Activity Execution**

#### **Execute Activity**
```go
// Выполнение activity
activityOptions := workflow.ActivityOptions{
    StartToCloseTimeout: 5 * time.Minute,
    HeartbeatTimeout:    30 * time.Second,
    RetryPolicy: &workflow.RetryPolicy{
        InitialInterval:    time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    30 * time.Second,
        MaximumAttempts:    3,
    },
}

var result map[string]interface{}
err := workflow.ExecuteActivity(
    workflow.WithActivityOptions(ctx, activityOptions),
    "CreateP2PQUICTunnelActivity",
    tenantID,
    peerID,
    config,
).Get(ctx, &result)
```

---

## 🔧 **Интеграция с P2P Manager**

### **CadenceP2PIntegration**

#### **OrchestrateTunnelCreation**
```go
// Оркестрация создания туннеля через Cadence
func (c *CadenceP2PIntegration) OrchestrateTunnelCreation(ctx context.Context, tenantID, peerID string, config map[string]interface{}) error {
    workflowOptions := cadence.StartWorkflowOptions{
        ID:        fmt.Sprintf("tunnel-creation-%s-%s", tenantID, peerID),
        TaskList:  "p2p-quic-tasks",
        ExecutionTimeout: 10 * time.Minute,
    }

    input := workflows.P2PQUICTunnelWorkflowInput{
        TenantID:      tenantID,
        PeerID:        peerID,
        Operation:     "create",
        TunnelConfig:  config,
        NetworkConfig: make(map[string]interface{}),
    }

    _, err := c.cadenceClient.ExecuteWorkflow(ctx, workflowOptions, workflows.P2PQUICTunnelWorkflow, input)
    return err
}
```

#### **OrchestrateConnectionEstablishment**
```go
// Оркестрация установки соединения через Cadence
func (c *CadenceP2PIntegration) OrchestrateConnectionEstablishment(ctx context.Context, tenantID, sourcePeerID, targetPeerID string, config map[string]interface{}) error {
    workflowOptions := cadence.StartWorkflowOptions{
        ID:        fmt.Sprintf("connection-establishment-%s-%s-%s", tenantID, sourcePeerID, targetPeerID),
        TaskList:  "p2p-quic-tasks",
        ExecutionTimeout: 5 * time.Minute,
    }

    input := workflows.P2PConnectionOrchestrationWorkflowInput{
        TenantID:       tenantID,
        SourcePeerID:   sourcePeerID,
        TargetPeerID:   targetPeerID,
        ConnectionType: "direct",
        Protocol:       "quic",
        Configuration:  config,
    }

    _, err := c.cadenceClient.ExecuteWorkflow(ctx, workflowOptions, workflows.P2PConnectionOrchestrationWorkflow, input)
    return err
}
```

#### **OrchestrateAIOptimizedRouting**
```go
// Оркестрация AI-оптимизированной маршрутизации через Cadence
func (c *CadenceP2PIntegration) OrchestrateAIOptimizedRouting(ctx context.Context, tenantID, sourcePeerID, targetPeerID string, dataSize int64, priority string, networkConditions map[string]interface{}) error {
    workflowOptions := cadence.StartWorkflowOptions{
        ID:        fmt.Sprintf("ai-optimized-routing-%s-%s-%s", tenantID, sourcePeerID, targetPeerID),
        TaskList:  "p2p-quic-tasks",
        ExecutionTimeout: 5 * time.Minute,
    }

    input := workflows.AIOptimizedRoutingWorkflowInput{
        TenantID:          tenantID,
        SourcePeerID:      sourcePeerID,
        TargetPeerID:      targetPeerID,
        DataSize:          dataSize,
        Priority:          priority,
        NetworkConditions: networkConditions,
    }

    _, err := c.cadenceClient.ExecuteWorkflow(ctx, workflowOptions, workflows.AIOptimizedRoutingWorkflow, input)
    return err
}
```

---

## 🔧 **Использование**

### **Примеры выполнения workflows:**

#### **Создание P2P QUIC туннеля:**
```bash
# Через P2P Manager
curl -X POST "http://localhost:8082/api/v1/tenants/tenant-001/tunnels/create" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tunnel_name": "tunnel-001",
    "source_peer_id": "peer-001",
    "target_peer_id": "peer-002",
    "tunnel_type": "quic"
  }'
```

#### **AI-оптимизированная маршрутизация:**
```bash
# Через AI Service
curl -X POST "http://localhost:8083/api/v1/ai/optimize-routing" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant-001",
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

### **Мониторинг через Web UI:**
```bash
# Открыть Cadence Web UI
kubectl port-forward service/cadence-web-service 8088:8088 -n cloudbridge
# Перейти на http://localhost:8088
```

---

## 📊 **Метрики**

### **Cadence метрики:**
- `cadence_workflow_started_total` - Количество запущенных workflows
- `cadence_workflow_completed_total` - Количество завершенных workflows
- `cadence_workflow_failed_total` - Количество неудачных workflows
- `cadence_activity_started_total` - Количество запущенных activities
- `cadence_activity_completed_total` - Количество завершенных activities
- `cadence_activity_failed_total` - Количество неудачных activities
- `cadence_workflow_duration_seconds` - Длительность выполнения workflows
- `cadence_activity_duration_seconds` - Длительность выполнения activities

---

## 🔍 **Отладка**

### **Проверка статуса:**
```bash
# Cadence Service
kubectl get pods -n cloudbridge | grep cadence

# Cadence Web UI
kubectl port-forward service/cadence-web-service 8088:8088 -n cloudbridge
```

### **Логи:**
```bash
# Cadence Service
kubectl logs deployment/cadence -n cloudbridge

# Cadence Web UI
kubectl logs deployment/cadence-web -n cloudbridge
```

### **Workflow статус:**
```bash
# Через Web UI
# Открыть http://localhost:8088
# Перейти в раздел Workflows
# Найти workflow по ID
```

---

## 🔐 **Конфигурация**

### **Cadence конфигурация:**
```json
{
  "cadence": {
    "enabled": true,
    "host_port": "cadence:7933",
    "domain": "cloudbridge-p2p",
    "task_list": "p2p-quic-tasks"
  }
}
```

### **Переменные окружения:**
```bash
CADENCE_ENABLED=true
CADENCE_HOST_PORT=cadence:7933
CADENCE_DOMAIN=cloudbridge-p2p
CADENCE_TASK_LIST=p2p-quic-tasks
```

---

**🎯 Cadence API Endpoints готовы к использованию для оркестрации P2P QUIC workflows и AI-оптимизации!**
