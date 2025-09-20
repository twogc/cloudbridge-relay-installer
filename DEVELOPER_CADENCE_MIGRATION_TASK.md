# 🔄 **Задача для разработчика: Миграция Cadence с WireGuard на P2P QUIC**

## 📋 **Краткое описание задачи**

**Цель:** Перенести Cadence Workflow Engine с WireGuard на P2P QUIC для создания интеллектуальной оркестрации P2P mesh сетей с AI-оптимизацией.

**Приоритет:** ВЫСОКИЙ  
**Срок:** 3-5 дней  
**Статус:** Требует реализации

---

## 🎯 **Контекст и обоснование**

### **Текущее состояние:**
- ✅ **Cadence развернут** в Minikube и работает
- ❌ **Cadence отключен** в конфигурации (`"enabled": false`)
- ❌ **Cadence привязан к WireGuard** (устаревшая архитектура)
- ✅ **P2P QUIC работает** независимо от Cadence
- ❌ **Нет оркестрации** P2P соединений

### **Проблема:**
Cadence Workflow Engine предназначен для WireGuard, но проект перешел на современную P2P QUIC архитектуру. Нужна полная миграция для создания интеллектуальной системы управления P2P mesh сетями.

---

## 🏗️ **Целевая архитектура**

### **До миграции (текущее состояние):**
```
┌─────────────────┐    ┌─────────────────┐
│   Cadence       │    │   P2P QUIC      │
│   (отключен)    │    │   (активен)     │
│                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ WireGuard   │ │    │ │ Tunnel      │ │
│ │ Workflows   │ │    │ │ Manager     │ │
│ └─────────────┘ │    │ └─────────────┘ │
│                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ AI          │ │    │ │ QUIC        │ │
│ │ Workflows   │ │    │ │ Underlay    │ │
│ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘
        │                       │
        └───────────────────────┼───────────────────────┐
                                │                       │
                        ┌───────▼────────┐    ┌────────▼────────┐
                        │ CloudBridge    │    │ P2P API         │
                        │ Relay          │    │ (порт 8082)     │
                        │ (основной)     │    │                 │
                        └────────────────┘    └─────────────────┘
```

### **После миграции (целевое состояние):**
```
┌─────────────────────────────────────────────────────────────┐
│                    Cadence Workflow Engine                  │
│                        (активен)                            │
│                                                             │
│ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│ │ P2P QUIC        │ │ AI Optimization │ │ Security        │ │
│ │ Workflows       │ │ Workflows       │ │ Workflows       │ │
│ │                 │ │                 │ │                 │ │
│ │ • Tunnel        │ │ • Performance   │ │ • Key           │ │
│ │   Management    │ │   Optimization  │ │   Rotation      │ │
│ │ • Connection    │ │ • Load          │ │ • Threat        │ │
│ │   Orchestration │ │   Balancing     │ │   Detection     │ │
│ │ • NAT Traversal │ │ • Route         │ │ • Access        │ │
│ │ • Peer          │ │   Optimization  │ │   Control       │ │
│ │   Discovery     │ │ • ML Models     │ │ • Audit         │ │
│ └─────────────────┘ └─────────────────┘ └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                        ┌───────▼────────┐
                        │ CloudBridge    │
                        │ Relay          │
                        │ (оркестрируемый)│
                        │                 │
                        │ ┌─────────────┐ │
                        │ │ P2P QUIC    │ │
                        │ │ Manager     │ │
                        │ │ (управляемый)│ │
                        │ └─────────────┘ │
                        │                 │
                        │ ┌─────────────┐ │
                        │ │ AI Analytics│ │
                        │ │ (интегрирован)│ │
                        │ └─────────────┘ │
                        └─────────────────┘
```

---

## 📋 **Детальный план миграции**

### **Этап 1: Активация и конфигурация Cadence (1 день)**

#### **1.1 Включить Cadence в конфигурации:**
```json
// configs/config.json
"cadence": {
  "enabled": true,  // ← ИЗМЕНИТЬ с false на true
  "host_port": "cadence:7933",  // ← ИЗМЕНИТЬ с localhost на cadence
  "domain": "cloudbridge-p2p",  // ← ИЗМЕНИТЬ домен
  "task_list": "p2p-quic-tasks"  // ← ИЗМЕНИТЬ task list
}
```

#### **1.2 Обновить deployment.yaml:**
```yaml
# k8s/deployment.yaml
env:
  - name: CADENCE_ENABLED
    value: "true"
  - name: CADENCE_HOST_PORT
    value: "cadence:7933"
  - name: CADENCE_DOMAIN
    value: "cloudbridge-p2p"
  - name: CADENCE_TASK_LIST
    value: "p2p-quic-tasks"
```

#### **1.3 Проверить подключение:**
```bash
# Собрать новый образ с активированным Cadence
docker build -t cloudbridge-relay:cadence-p2p-v1 .

# Загрузить в Minikube
minikube image load cloudbridge-relay:cadence-p2p-v1

# Обновить deployment
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:cadence-p2p-v1 -n cloudbridge
```

### **Этап 2: Создание P2P QUIC Workflows (2 дня)**

#### **2.1 Создать новые workflow файлы:**

**`internal/cadence/workflows/p2p_quic.go`:**
```go
package workflows

import (
    "time"
    "go.uber.org/cadence/workflow"
    "go.uber.org/zap"
)

// P2PQUICTunnelWorkflowInput входные параметры для управления P2P QUIC туннелями
type P2PQUICTunnelWorkflowInput struct {
    TenantID        string                 `json:"tenant_id"`
    PeerID          string                 `json:"peer_id"`
    Operation       string                 `json:"operation"` // create, update, delete, monitor
    TunnelConfig    map[string]interface{} `json:"tunnel_config"`
    NetworkConfig   map[string]interface{} `json:"network_config"`
}

// P2PQUICTunnelWorkflowResult результат управления P2P QUIC туннелями
type P2PQUICTunnelWorkflowResult struct {
    TenantID      string                 `json:"tenant_id"`
    PeerID        string                 `json:"peer_id"`
    Operation     string                 `json:"operation"`
    Status        string                 `json:"status"`
    TunnelID      string                 `json:"tunnel_id"`
    Result        map[string]interface{} `json:"result"`
    StartedAt     time.Time              `json:"started_at"`
    CompletedAt   time.Time              `json:"completed_at"`
    Message       string                 `json:"message"`
}

// P2PQUICTunnelWorkflow workflow для управления P2P QUIC туннелями
func P2PQUICTunnelWorkflow(ctx workflow.Context, input P2PQUICTunnelWorkflowInput) (*P2PQUICTunnelWorkflowResult, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("Запуск P2PQUICTunnelWorkflow",
        zap.String("tenant_id", input.TenantID),
        zap.String("peer_id", input.PeerID),
        zap.String("operation", input.Operation),
    )

    result := &P2PQUICTunnelWorkflowResult{
        TenantID:    input.TenantID,
        PeerID:      input.PeerID,
        Operation:   input.Operation,
        Status:      "running",
        StartedAt:   workflow.Now(ctx),
        Result:      make(map[string]interface{}),
    }

    // Настройки для activities
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

    switch input.Operation {
    case "create":
        // Создание P2P QUIC туннеля
        var tunnelResult map[string]interface{}
        err := workflow.ExecuteActivity(
            workflow.WithActivityOptions(ctx, activityOptions),
            "CreateP2PQUICTunnelActivity",
            input.TenantID,
            input.PeerID,
            input.TunnelConfig,
        ).Get(ctx, &tunnelResult)

        if err != nil {
            result.Status = "failed"
            result.Message = "Ошибка создания туннеля: " + err.Error()
            return result, err
        }

        result.TunnelID = tunnelResult["tunnel_id"].(string)
        result.Result = tunnelResult

    case "monitor":
        // Мониторинг P2P QUIC туннеля
        var monitorResult map[string]interface{}
        err := workflow.ExecuteActivity(
            workflow.WithActivityOptions(ctx, activityOptions),
            "MonitorP2PQUICTunnelActivity",
            input.TenantID,
            input.PeerID,
        ).Get(ctx, &monitorResult)

        if err != nil {
            result.Status = "failed"
            result.Message = "Ошибка мониторинга туннеля: " + err.Error()
            return result, err
        }

        result.Result = monitorResult
    }

    result.Status = "completed"
    result.CompletedAt = workflow.Now(ctx)
    result.Message = "P2P QUIC туннель успешно обработан"

    return result, nil
}
```

**`internal/cadence/workflows/p2p_connection.go`:**
```go
package workflows

// P2PConnectionOrchestrationWorkflowInput входные параметры для оркестрации P2P соединений
type P2PConnectionOrchestrationWorkflowInput struct {
    TenantID          string                 `json:"tenant_id"`
    SourcePeerID      string                 `json:"source_peer_id"`
    TargetPeerID      string                 `json:"target_peer_id"`
    ConnectionType    string                 `json:"connection_type"` // direct, relay, masque
    Protocol          string                 `json:"protocol"`        // quic, websocket, tcp
    Configuration     map[string]interface{} `json:"configuration"`
}

// P2PConnectionOrchestrationWorkflowResult результат оркестрации P2P соединений
type P2PConnectionOrchestrationWorkflowResult struct {
    TenantID          string                 `json:"tenant_id"`
    SourcePeerID      string                 `json:"source_peer_id"`
    TargetPeerID      string                 `json:"target_peer_id"`
    ConnectionID      string                 `json:"connection_id"`
    Status            string                 `json:"status"`
    ConnectionType    string                 `json:"connection_type"`
    Protocol          string                 `json:"protocol"`
    Latency           time.Duration          `json:"latency"`
    Throughput        int64                  `json:"throughput"`
    Result            map[string]interface{} `json:"result"`
    StartedAt         time.Time              `json:"started_at"`
    CompletedAt       time.Time              `json:"completed_at"`
    Message           string                 `json:"message"`
}

// P2PConnectionOrchestrationWorkflow workflow для оркестрации P2P соединений
func P2PConnectionOrchestrationWorkflow(ctx workflow.Context, input P2PConnectionOrchestrationWorkflowInput) (*P2PConnectionOrchestrationWorkflowResult, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("Запуск P2PConnectionOrchestrationWorkflow",
        zap.String("tenant_id", input.TenantID),
        zap.String("source_peer", input.SourcePeerID),
        zap.String("target_peer", input.TargetPeerID),
        zap.String("connection_type", input.ConnectionType),
        zap.String("protocol", input.Protocol),
    )

    result := &P2PConnectionOrchestrationWorkflowResult{
        TenantID:       input.TenantID,
        SourcePeerID:   input.SourcePeerID,
        TargetPeerID:   input.TargetPeerID,
        ConnectionType: input.ConnectionType,
        Protocol:       input.Protocol,
        Status:         "running",
        StartedAt:      workflow.Now(ctx),
        Result:         make(map[string]interface{}),
    }

    // 1. Проверка доступности пиров
    var peerCheckResult map[string]interface{}
    err := workflow.ExecuteActivity(
        workflow.WithActivityOptions(ctx, activityOptions),
        "CheckPeersAvailabilityActivity",
        input.TenantID,
        []string{input.SourcePeerID, input.TargetPeerID},
    ).Get(ctx, &peerCheckResult)

    if err != nil {
        result.Status = "failed"
        result.Message = "Ошибка проверки доступности пиров: " + err.Error()
        return result, err
    }

    // 2. Выбор оптимального типа соединения
    var connectionTypeResult map[string]interface{}
    err = workflow.ExecuteActivity(
        workflow.WithActivityOptions(ctx, activityOptions),
        "SelectOptimalConnectionTypeActivity",
        input.TenantID,
        input.SourcePeerID,
        input.TargetPeerID,
        input.Configuration,
    ).Get(ctx, &connectionTypeResult)

    if err != nil {
        result.Status = "failed"
        result.Message = "Ошибка выбора типа соединения: " + err.Error()
        return result, err
    }

    // 3. Установка P2P соединения
    var connectionResult map[string]interface{}
    err = workflow.ExecuteActivity(
        workflow.WithActivityOptions(ctx, activityOptions),
        "EstablishP2PConnectionActivity",
        input.TenantID,
        input.SourcePeerID,
        input.TargetPeerID,
        connectionTypeResult["connection_type"].(string),
        input.Protocol,
    ).Get(ctx, &connectionResult)

    if err != nil {
        result.Status = "failed"
        result.Message = "Ошибка установки соединения: " + err.Error()
        return result, err
    }

    result.ConnectionID = connectionResult["connection_id"].(string)
    result.Latency = time.Duration(connectionResult["latency"].(int64)) * time.Millisecond
    result.Throughput = connectionResult["throughput"].(int64)
    result.Result = connectionResult

    result.Status = "completed"
    result.CompletedAt = workflow.Now(ctx)
    result.Message = "P2P соединение успешно установлено"

    return result, nil
}
```

#### **2.2 Создать Activities для P2P QUIC:**

**`internal/cadence/activities/p2p_quic.go`:**
```go
package activities

import (
    "context"
    "fmt"
    "time"
    "go.uber.org/cadence/activity"
    "go.uber.org/zap"
)

// CreateP2PQUICTunnelActivity создает P2P QUIC туннель
func CreateP2PQUICTunnelActivity(ctx context.Context, tenantID, peerID string, config map[string]interface{}) (map[string]interface{}, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Создание P2P QUIC туннеля",
        zap.String("tenant_id", tenantID),
        zap.String("peer_id", peerID),
    )

    // Здесь должна быть реальная логика создания P2P QUIC туннеля
    // Интеграция с P2P Manager из internal/p2p/
    
    tunnelID := fmt.Sprintf("tunnel-%s-%s-%d", tenantID, peerID, time.Now().UnixNano())
    
    result := map[string]interface{}{
        "tunnel_id":    tunnelID,
        "tenant_id":    tenantID,
        "peer_id":      peerID,
        "status":       "created",
        "created_at":   time.Now(),
        "config":       config,
    }

    logger.Info("P2P QUIC туннель создан",
        zap.String("tunnel_id", tunnelID),
    )

    return result, nil
}

// MonitorP2PQUICTunnelActivity мониторит P2P QUIC туннель
func MonitorP2PQUICTunnelActivity(ctx context.Context, tenantID, peerID string) (map[string]interface{}, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Мониторинг P2P QUIC туннеля",
        zap.String("tenant_id", tenantID),
        zap.String("peer_id", peerID),
    )

    // Здесь должна быть реальная логика мониторинга
    // Интеграция с P2P Manager для получения статистики
    
    result := map[string]interface{}{
        "tenant_id":        tenantID,
        "peer_id":          peerID,
        "status":           "active",
        "bytes_in":         1024,
        "bytes_out":        2048,
        "connections":      5,
        "last_activity":    time.Now(),
        "latency":          50, // ms
        "throughput":       1000, // bytes/sec
    }

    return result, nil
}

// EstablishP2PConnectionActivity устанавливает P2P соединение
func EstablishP2PConnectionActivity(ctx context.Context, tenantID, sourcePeerID, targetPeerID, connectionType, protocol string) (map[string]interface{}, error) {
    logger := activity.GetLogger(ctx)
    logger.Info("Установка P2P соединения",
        zap.String("tenant_id", tenantID),
        zap.String("source_peer", sourcePeerID),
        zap.String("target_peer", targetPeerID),
        zap.String("connection_type", connectionType),
        zap.String("protocol", protocol),
    )

    // Здесь должна быть реальная логика установки P2P соединения
    // Интеграция с P2P Connection Manager
    
    connectionID := fmt.Sprintf("conn-%s-%s-%s-%d", tenantID, sourcePeerID, targetPeerID, time.Now().UnixNano())
    
    result := map[string]interface{}{
        "connection_id":    connectionID,
        "tenant_id":        tenantID,
        "source_peer_id":   sourcePeerID,
        "target_peer_id":   targetPeerID,
        "connection_type":  connectionType,
        "protocol":         protocol,
        "status":           "established",
        "latency":          45, // ms
        "throughput":       2000, // bytes/sec
        "established_at":   time.Now(),
    }

    logger.Info("P2P соединение установлено",
        zap.String("connection_id", connectionID),
    )

    return result, nil
}
```

### **Этап 3: Интеграция с P2P Manager (1 день)**

#### **3.1 Создать Cadence интеграцию в P2P Manager:**

**`internal/p2p/cadence_integration.go`:**
```go
package p2p

import (
    "context"
    "fmt"
    "go.uber.org/zap"
    "github.com/twogc/cloudbridge-relay-installer/internal/cadence"
)

// CadenceP2PIntegration интеграция P2P Manager с Cadence
type CadenceP2PIntegration struct {
    cadenceClient *cadence.Client
    logger        *zap.Logger
}

// NewCadenceP2PIntegration создает новую интеграцию
func NewCadenceP2PIntegration(cadenceClient *cadence.Client, logger *zap.Logger) *CadenceP2PIntegration {
    return &CadenceP2PIntegration{
        cadenceClient: cadenceClient,
        logger:        logger,
    }
}

// OrchestrateTunnelCreation оркестрирует создание туннеля через Cadence
func (c *CadenceP2PIntegration) OrchestrateTunnelCreation(ctx context.Context, tenantID, peerID string, config map[string]interface{}) error {
    c.logger.Info("Оркестрация создания туннеля через Cadence",
        zap.String("tenant_id", tenantID),
        zap.String("peer_id", peerID),
    )

    // Запуск workflow через Cadence
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
    if err != nil {
        c.logger.Error("Ошибка запуска workflow создания туннеля", zap.Error(err))
        return err
    }

    c.logger.Info("Workflow создания туннеля запущен успешно")
    return nil
}

// OrchestrateConnectionEstablishment оркестрирует установку соединения через Cadence
func (c *CadenceP2PIntegration) OrchestrateConnectionEstablishment(ctx context.Context, tenantID, sourcePeerID, targetPeerID string, config map[string]interface{}) error {
    c.logger.Info("Оркестрация установки соединения через Cadence",
        zap.String("tenant_id", tenantID),
        zap.String("source_peer", sourcePeerID),
        zap.String("target_peer", targetPeerID),
    )

    // Запуск workflow через Cadence
    workflowOptions := cadence.StartWorkflowOptions{
        ID:        fmt.Sprintf("connection-establishment-%s-%s-%s", tenantID, sourcePeerID, targetPeerID),
        TaskList:  "p2p-quic-tasks",
        ExecutionTimeout: 5 * time.Minute,
    }

    input := workflows.P2PConnectionOrchestrationWorkflowInput{
        TenantID:       tenantID,
        SourcePeerID:   sourcePeerID,
        TargetPeerID:   targetPeerID,
        ConnectionType: "direct", // или "relay", "masque"
        Protocol:       "quic",
        Configuration:  config,
    }

    _, err := c.cadenceClient.ExecuteWorkflow(ctx, workflowOptions, workflows.P2PConnectionOrchestrationWorkflow, input)
    if err != nil {
        c.logger.Error("Ошибка запуска workflow установки соединения", zap.Error(err))
        return err
    }

    c.logger.Info("Workflow установки соединения запущен успешно")
    return nil
}
```

#### **3.2 Обновить P2P Manager для использования Cadence:**

**`internal/p2p/manager.go` (добавить интеграцию):**
```go
// P2PManager управляет P2P соединениями с интеграцией Cadence
type P2PManager struct {
    // ... существующие поля ...
    cadenceIntegration *CadenceP2PIntegration
}

// NewP2PManager создает новый P2P Manager с Cadence интеграцией
func NewP2PManager(cadenceClient *cadence.Client, logger *zap.Logger) *P2PManager {
    return &P2PManager{
        // ... инициализация существующих полей ...
        cadenceIntegration: NewCadenceP2PIntegration(cadenceClient, logger),
    }
}

// CreateTunnelWithOrchestration создает туннель с оркестрацией через Cadence
func (pm *P2PManager) CreateTunnelWithOrchestration(ctx context.Context, tenantID, peerID string, config map[string]interface{}) error {
    pm.logger.Info("Создание туннеля с оркестрацией",
        zap.String("tenant_id", tenantID),
        zap.String("peer_id", peerID),
    )

    // Запуск оркестрации через Cadence
    err := pm.cadenceIntegration.OrchestrateTunnelCreation(ctx, tenantID, peerID, config)
    if err != nil {
        pm.logger.Error("Ошибка оркестрации создания туннеля", zap.Error(err))
        return err
    }

    // Дополнительная логика создания туннеля
    // ...

    return nil
}
```

### **Этап 4: AI интеграция и оптимизация (1 день)**

#### **4.1 Создать AI-оптимизированные workflows:**

**`internal/cadence/workflows/ai_optimization.go`:**
```go
package workflows

// AIOptimizedRoutingWorkflowInput входные параметры для AI-оптимизированной маршрутизации
type AIOptimizedRoutingWorkflowInput struct {
    TenantID          string                 `json:"tenant_id"`
    SourcePeerID      string                 `json:"source_peer_id"`
    TargetPeerID      string                 `json:"target_peer_id"`
    DataSize          int64                  `json:"data_size"`
    Priority          string                 `json:"priority"` // low, medium, high, critical
    NetworkConditions map[string]interface{} `json:"network_conditions"`
}

// AIOptimizedRoutingWorkflowResult результат AI-оптимизированной маршрутизации
type AIOptimizedRoutingWorkflowResult struct {
    TenantID          string                 `json:"tenant_id"`
    SourcePeerID      string                 `json:"source_peer_id"`
    TargetPeerID      string                 `json:"target_peer_id"`
    OptimalRoute      map[string]interface{} `json:"optimal_route"`
    PredictedLatency  time.Duration          `json:"predicted_latency"`
    PredictedThroughput int64                `json:"predicted_throughput"`
    Confidence        float64                `json:"confidence"`
    AIRecommendations []string               `json:"ai_recommendations"`
    Result            map[string]interface{} `json:"result"`
    StartedAt         time.Time              `json:"started_at"`
    CompletedAt       time.Time              `json:"completed_at"`
    Message           string                 `json:"message"`
}

// AIOptimizedRoutingWorkflow workflow для AI-оптимизированной маршрутизации
func AIOptimizedRoutingWorkflow(ctx workflow.Context, input AIOptimizedRoutingWorkflowInput) (*AIOptimizedRoutingWorkflowResult, error) {
    logger := workflow.GetLogger(ctx)
    logger.Info("Запуск AIOptimizedRoutingWorkflow",
        zap.String("tenant_id", input.TenantID),
        zap.String("source_peer", input.SourcePeerID),
        zap.String("target_peer", input.TargetPeerID),
        zap.String("priority", input.Priority),
    )

    result := &AIOptimizedRoutingWorkflowResult{
        TenantID:       input.TenantID,
        SourcePeerID:   input.SourcePeerID,
        TargetPeerID:   input.TargetPeerID,
        Status:         "running",
        StartedAt:      workflow.Now(ctx),
        Result:         make(map[string]interface{}),
    }

    // 1. Анализ сетевых условий через AI
    var networkAnalysisResult map[string]interface{}
    err := workflow.ExecuteActivity(
        workflow.WithActivityOptions(ctx, activityOptions),
        "AnalyzeNetworkConditionsActivity",
        input.TenantID,
        input.NetworkConditions,
    ).Get(ctx, &networkAnalysisResult)

    if err != nil {
        result.Status = "failed"
        result.Message = "Ошибка анализа сетевых условий: " + err.Error()
        return result, nil
    }

    // 2. ML-предсказание оптимального маршрута
    var routingPredictionResult map[string]interface{}
    err = workflow.ExecuteActivity(
        workflow.WithActivityOptions(ctx, activityOptions),
        "PredictOptimalRouteActivity",
        input.TenantID,
        input.SourcePeerID,
        input.TargetPeerID,
        input.DataSize,
        input.Priority,
        networkAnalysisResult,
    ).Get(ctx, &routingPredictionResult)

    if err != nil {
        result.Status = "failed"
        result.Message = "Ошибка предсказания маршрута: " + err.Error()
        return result, nil
    }

    result.OptimalRoute = routingPredictionResult["route"].(map[string]interface{})
    result.PredictedLatency = time.Duration(routingPredictionResult["latency"].(int64)) * time.Millisecond
    result.PredictedThroughput = routingPredictionResult["throughput"].(int64)
    result.Confidence = routingPredictionResult["confidence"].(float64)
    result.AIRecommendations = routingPredictionResult["recommendations"].([]string)

    result.Status = "completed"
    result.CompletedAt = workflow.Now(ctx)
    result.Message = "AI-оптимизированная маршрутизация завершена"

    return result, nil
}
```

### **Этап 5: Тестирование и валидация (1 день)**

#### **5.1 Создать тесты для Cadence интеграции:**

**`internal/cadence/workflows/p2p_quic_test.go`:**
```go
package workflows

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "go.uber.org/cadence/testsuite"
)

func TestP2PQUICTunnelWorkflow(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()

    // Регистрация activities
    env.RegisterActivity(CreateP2PQUICTunnelActivity)
    env.RegisterActivity(MonitorP2PQUICTunnelActivity)

    // Тестовые данные
    input := P2PQUICTunnelWorkflowInput{
        TenantID:      "test-tenant",
        PeerID:        "test-peer",
        Operation:     "create",
        TunnelConfig:  map[string]interface{}{
            "max_streams": 100,
            "keep_alive":  true,
        },
        NetworkConfig: map[string]interface{}{},
    }

    // Выполнение workflow
    env.ExecuteWorkflow(P2PQUICTunnelWorkflow, input)

    // Проверка результатов
    assert.True(t, env.IsWorkflowCompleted())
    assert.NoError(t, env.GetWorkflowError())

    var result P2PQUICTunnelWorkflowResult
    err := env.GetWorkflowResult(&result)
    assert.NoError(t, err)
    assert.Equal(t, "completed", result.Status)
    assert.Equal(t, "test-tenant", result.TenantID)
    assert.Equal(t, "test-peer", result.PeerID)
}
```

#### **5.2 Интеграционные тесты:**

**`tests/integration/cadence_p2p_integration_test.go`:**
```go
package integration

import (
    "context"
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/twogc/cloudbridge-relay-installer/internal/cadence"
    "github.com/twogc/cloudbridge-relay-installer/internal/p2p"
)

func TestCadenceP2PIntegration(t *testing.T) {
    // Инициализация Cadence клиента
    cadenceClient, err := cadence.NewClient(cadence.Config{
        HostPort: "localhost:7933",
        Domain:   "cloudbridge-p2p",
        TaskList: "p2p-quic-tasks",
    }, logger)
    assert.NoError(t, err)

    // Инициализация P2P Manager с Cadence интеграцией
    p2pManager := p2p.NewP2PManager(cadenceClient, logger)

    // Тест создания туннеля с оркестрацией
    ctx := context.Background()
    err = p2pManager.CreateTunnelWithOrchestration(ctx, "test-tenant", "test-peer", map[string]interface{}{
        "max_streams": 100,
        "keep_alive":  true,
    })
    assert.NoError(t, err)

    // Ожидание завершения workflow
    time.Sleep(5 * time.Second)

    // Проверка результатов
    // ...
}
```

---

## 🚀 **Инструкции по развертыванию**

### **1. Обязательное чтение документов:**

#### **📖 MINIKUBE_BUILD_GUIDE.md:**
- Полная инструкция по сборке и развертыванию в Minikube
- Команды для сборки Docker образов с новыми тегами
- Настройка Kubernetes конфигураций
- Тестирование и отладка

#### **📖 DEVELOPER_BRIEFING.md:**
- Понимание цели проекта (P2P mesh сети)
- Текущее состояние и архитектура
- План разработки и приоритеты
- Критерии успеха

### **2. Сборка с новым тегом:**

```bash
# 1. Перейти в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# 2. Собрать образ с новым тегом (обязательно новый тег!)
docker build -t cloudbridge-relay:cadence-p2p-v1 .

# 3. Загрузить в Minikube
minikube image load cloudbridge-relay:cadence-p2p-v1

# 4. Обновить deployment
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:cadence-p2p-v1 -n cloudbridge

# 5. Проверить статус
kubectl get pods -n cloudbridge
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge
```

### **3. Удаление старых образов:**

```bash
# Удалить старые образы из Minikube
minikube image rm cloudbridge-relay:security-fixed-20250920-093827-b65ac2ca
minikube image rm cloudbridge-relay:quic-proper-20250920-012537

# Удалить старые образы из Docker
docker rmi cloudbridge-relay:security-fixed-20250920-093827-b65ac2ca
docker rmi cloudbridge-relay:quic-proper-20250920-012537

# Очистить неиспользуемые образы
docker system prune -f
```

---

## 🎯 **Ожидаемый результат**

### **Архитектура после миграции:**

```
┌─────────────────────────────────────────────────────────────┐
│                    Cadence Workflow Engine                  │
│                        (активен)                            │
│                                                             │
│ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│ │ P2P QUIC        │ │ AI Optimization │ │ Security        │ │
│ │ Workflows       │ │ Workflows       │ │ Workflows       │ │
│ │                 │ │                 │ │                 │ │
│ │ • Tunnel        │ │ • Performance   │ │ • Key           │ │
│ │   Management    │ │   Optimization  │ │   Rotation      │ │
│ │ • Connection    │ │ • Load          │ │ • Threat        │ │
│ │   Orchestration │ │   Balancing     │ │   Detection     │ │
│ │ • NAT Traversal │ │ • Route         │ │ • Access        │ │
│ │ • Peer          │ │   Optimization  │ │   Control       │ │
│ │   Discovery     │ │ • ML Models     │ │ • Audit         │ │
│ └─────────────────┘ └─────────────────┘ └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                                │
                        ┌───────▼────────┐
                        │ CloudBridge    │
                        │ Relay          │
                        │ (оркестрируемый)│
                        │                 │
                        │ ┌─────────────┐ │
                        │ │ P2P QUIC    │ │
                        │ │ Manager     │ │
                        │ │ (управляемый)│ │
                        │ └─────────────┘ │
                        │                 │
                        │ ┌─────────────┐ │
                        │ │ AI Analytics│ │
                        │ │ (интегрирован)│ │
                        │ └─────────────┘ │
                        └─────────────────┘
```

### **Ключевые возможности:**

1. **Модульная система** с поддержкой современных протоколов (QUIC, ICE/STUN/TURN, MASQUE)
2. **AI интеграция** с ML моделями, Cadence workflows и real-time аналитикой
3. **Безопасность** с многоуровневой аутентификацией и мультитенантной изоляцией
4. **Масштабируемость** до 1000 туннелей и 10,000 одновременных соединений
5. **Production-ready** с полной контейнеризацией и Kubernetes развертыванием

### **Функциональность:**

- ✅ **Интеллектуальная оркестрация** P2P соединений через Cadence
- ✅ **AI-оптимизированная маршрутизация** с ML предсказаниями
- ✅ **Автоматическое управление** туннелями и соединениями
- ✅ **Real-time мониторинг** и аналитика
- ✅ **Автоматическое восстановление** после сбоев
- ✅ **Масштабирование** под нагрузку

---

## 📋 **Чек-лист выполнения**

### **Этап 1: Активация Cadence**
- [ ] Изменить `"enabled": true` в `configs/config.json`
- [ ] Обновить `host_port` на `"cadence:7933"`
- [ ] Изменить `domain` на `"cloudbridge-p2p"`
- [ ] Изменить `task_list` на `"p2p-quic-tasks"`
- [ ] Обновить `k8s/deployment.yaml` с новыми переменными окружения
- [ ] Собрать образ с тегом `cloudbridge-relay:cadence-p2p-v1`
- [ ] Загрузить образ в Minikube
- [ ] Обновить deployment
- [ ] Проверить подключение к Cadence

### **Этап 2: P2P QUIC Workflows**
- [ ] Создать `internal/cadence/workflows/p2p_quic.go`
- [ ] Создать `internal/cadence/workflows/p2p_connection.go`
- [ ] Создать `internal/cadence/workflows/ai_optimization.go`
- [ ] Создать `internal/cadence/activities/p2p_quic.go`
- [ ] Реализовать `P2PQUICTunnelWorkflow`
- [ ] Реализовать `P2PConnectionOrchestrationWorkflow`
- [ ] Реализовать `AIOptimizedRoutingWorkflow`
- [ ] Создать все необходимые Activities

### **Этап 3: Интеграция с P2P Manager**
- [ ] Создать `internal/p2p/cadence_integration.go`
- [ ] Обновить `internal/p2p/manager.go` для интеграции с Cadence
- [ ] Реализовать `OrchestrateTunnelCreation`
- [ ] Реализовать `OrchestrateConnectionEstablishment`
- [ ] Обновить `NewP2PManager` для принятия Cadence клиента
- [ ] Добавить методы `CreateTunnelWithOrchestration`

### **Этап 4: AI интеграция**
- [ ] Создать AI-оптимизированные workflows
- [ ] Интегрировать с существующей AI инфраструктурой
- [ ] Реализовать ML предсказания для маршрутизации
- [ ] Добавить real-time аналитику
- [ ] Создать AI рекомендации для оптимизации

### **Этап 5: Тестирование**
- [ ] Создать unit тесты для workflows
- [ ] Создать интеграционные тесты
- [ ] Протестировать создание туннелей через Cadence
- [ ] Протестировать оркестрацию соединений
- [ ] Протестировать AI оптимизацию
- [ ] Проверить производительность и стабильность

### **Финальная проверка**
- [ ] Все тесты проходят
- [ ] Cadence workflows работают корректно
- [ ] P2P QUIC интегрирован с Cadence
- [ ] AI оптимизация функционирует
- [ ] Производительность соответствует требованиям
- [ ] Документация обновлена

---

## 📞 **Поддержка и ресурсы**

### **Документация:**
- **MINIKUBE_BUILD_GUIDE.md** - Инструкция по сборке и развертыванию
- **DEVELOPER_BRIEFING.md** - Понимание проекта и архитектуры
- **CADENCE_P2P_QUIC_ANALYSIS.md** - Анализ текущего состояния

### **Конфигурации:**
- **configs/config.json** - Основная конфигурация
- **k8s/deployment.yaml** - Kubernetes deployment
- **k8s/cloudbridge-relay-external-service.yaml** - Service конфигурация

### **Тестирование:**
- **MINIKUBE_BUILD_GUIDE.md** - Раздел "Тестирование P2P соединений"
- **DEVELOPER_BRIEFING.md** - Раздел "Критерии успеха"

---

**🎯 Цель: Создать интеллектуальную P2P mesh сеть с AI-оптимизацией, управляемую через Cadence Workflow Engine, с поддержкой современных протоколов и production-ready архитектурой!**

**📋 Задача: Перенести Cadence с WireGuard на P2P QUIC, создать AI-оптимизированные workflows и интегрировать с существующей P2P инфраструктурой для создания масштабируемой и интеллектуальной системы управления P2P соединениями.**
