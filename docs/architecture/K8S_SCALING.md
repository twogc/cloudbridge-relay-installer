# 🚀 Kubernetes масштабирование CloudBridge Relay

## 📊 Архитектура масштабирования в K8s

```mermaid
graph TB
    subgraph "External Traffic"
        Internet[🌐 Internet]
        LB[Load Balancer]
    end
    
    subgraph "Kubernetes Cluster"
        subgraph "Node 1"
            AUTH1[Auth Pod 1]
            QUIC1[QUIC Pod 1]
            WG1[WireGuard Pod]
        end
        
        subgraph "Node 2"
            AUTH2[Auth Pod 2]
            METRICS1[Metrics Pod 1]
            WG2[WireGuard Pod]
        end
        
        subgraph "Node 3"
            QUIC2[QUIC Pod 2]
            METRICS2[Metrics Pod 2]
            WG3[WireGuard Pod]
        end
    end
    
    subgraph "Control Plane"
        HPA[HPA Controller]
        PROMETHEUS[Prometheus]
        GRAFANA[Grafana]
    end
    
    Internet --> LB
    LB --> QUIC1
    LB --> QUIC2
    
    HPA --> AUTH1
    HPA --> AUTH2
    HPA --> QUIC1
    HPA --> QUIC2
    HPA --> METRICS1
    HPA --> METRICS2
    
    PROMETHEUS --> AUTH1
    PROMETHEUS --> AUTH2
    PROMETHEUS --> QUIC1
    PROMETHEUS --> QUIC2
    PROMETHEUS --> METRICS1
    PROMETHEUS --> METRICS2
```

## 🔄 HPA (Horizontal Pod Autoscaler)

```mermaid
graph LR
    subgraph "Triggers"
        CPU[CPU > 70%]
        MEM[Memory > 80%]
        CUSTOM[Custom Metrics]
    end
    
    subgraph "HPA Controller"
        HPA[HPA Controller]
    end
    
    subgraph "Actions"
        SCALE_UP[Scale Up<br/>+1 Replica]
        SCALE_DOWN[Scale Down<br/>-1 Replica]
    end
    
    subgraph "Services"
        AUTH[Auth Service<br/>2-10 replicas]
        QUIC[QUIC Service<br/>3-15 replicas]
        METRICS[Metrics Service<br/>2-5 replicas]
    end
    
    CPU --> HPA
    MEM --> HPA
    CUSTOM --> HPA
    
    HPA --> SCALE_UP
    HPA --> SCALE_DOWN
    
    SCALE_UP --> AUTH
    SCALE_UP --> QUIC
    SCALE_UP --> METRICS
    
    SCALE_DOWN --> AUTH
    SCALE_DOWN --> QUIC
    SCALE_DOWN --> METRICS
```

## 📈 Масштабирование по нагрузке

```mermaid
graph LR
    subgraph "Low Load"
        LL[2-3 replicas<br/>100m CPU each]
    end
    
    subgraph "Medium Load"
        ML[5-8 replicas<br/>200m CPU each]
    end
    
    subgraph "High Load"
        HL[10-15 replicas<br/>400m CPU each]
    end
    
    subgraph "Peak Load"
        PL[20+ replicas<br/>600m CPU each]
    end
    
    LL --> ML
    ML --> HL
    HL --> PL
```

## 🎯 Ключевые компоненты

### 1. **Deployments** (для HTTP сервисов)
- Auth Service: 2-10 реплик
- Metrics Service: 2-5 реплик
- QUIC Service: 3-15 реплик

### 2. **DaemonSet** (для сетевых сервисов)
- WireGuard: 1 под на каждый узел

### 3. **HPA** (автоматическое масштабирование)
- CPU > 70% → Scale Up
- Memory > 80% → Scale Up
- Custom metrics → Scale Up/Down

### 4. **Services** (сетевой доступ)
- ClusterIP для внутренних сервисов
- LoadBalancer для внешних сервисов
- NodePort для WireGuard

## 📊 Ресурсы и лимиты

| Сервис | Requests | Limits | Replicas |
|--------|----------|--------|----------|
| Auth | 100m CPU, 256Mi RAM | 500m CPU, 1Gi RAM | 2-10 |
| Metrics | 100m CPU, 128Mi RAM | 300m CPU, 512Mi RAM | 2-5 |
| QUIC | 200m CPU, 512Mi RAM | 1000m CPU, 2Gi RAM | 3-15 |
| WireGuard | 50m CPU, 64Mi RAM | 200m CPU, 256Mi RAM | 1/node |

## ✅ Преимущества Kubernetes

- **Автоматическое масштабирование** через HPA
- **Высокая доступность** - поды по узлам
- **Self-healing** - автоматическое восстановление
- **Rolling updates** - обновления без простоя
- **Resource management** - точные лимиты
- **Monitoring** - встроенный Prometheus Stack

---

**Статус**: ✅ Готов к масштабированию  
**Платформа**: Kubernetes  
**Автоматизация**: Полная
