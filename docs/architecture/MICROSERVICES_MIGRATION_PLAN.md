# 🏗️ CloudBridge Relay - Детальный план миграции к микросервисам

## 📋 Анализ текущего состояния проекта

### **Текущая архитектура (Монолит)**
```yaml
# Текущий docker-compose.yml - ВСЕ В ОДНОМ КОНТЕЙНЕРЕ
services:
  cloudbridge-relay:
    image: twogc/cloudbridge-relay:latest  # ❌ Монолит
    privileged: true                       # ❌ Слишком широкие права
    ports:
      - "8081:8081"    # Relay Metrics
      - "8082:8082"    # Relay API  
      - "9090:9090"    # Prometheus
      - "51820:51820/udp"  # WireGuard
    deploy:
      resources:
        limits:
          cpus: '2.0'      # ❌ Ограничение для всех модулей
          memory: 2G       # ❌ Конкуренция за память
```

### **Выявленные модули в проекте:**
1. **WireGuard Service** (`internal/wireguard/`) - P2P Mesh Network
2. **QUIC Service** (`internal/quic/`) - Enhanced QUIC Protocol  
3. **P2P Service** (`internal/p2p/`) - Peer Discovery & Management
4. **Quantum Service** (`internal/quantum/`) - Post-Quantum Cryptography
5. **AI/ML Service** (`internal/ai/`) - AI/ML Monitoring & Anomaly Detection
6. **Cadence Service** (`internal/cadence/`) - Workflow Orchestration
7. **Auth Service** (`internal/auth/`) - Keycloak Integration
8. **Metrics Service** (`internal/metrics/`) - Prometheus Metrics
9. **API Service** (`internal/api/`) - REST API Gateway
10. **Health Service** (`internal/health/`) - Health Checks

## 🎯 Целевая микросервисная архитектура

### **Архитектурная диаграмма:**

```mermaid
graph TB
    subgraph "🌍 Global Edge Network"
        subgraph "🇺🇸 Region US-East"
            LB1[🔄 Load Balancer<br/>HAProxy/NGINX]
            subgraph "🚪 Gateway Layer"
                GW1[🌐 API Gateway 1<br/>Kong/Istio]
                GW2[🌐 API Gateway 2<br/>Kong/Istio]
            end
            subgraph "🛡️ Core Services"
                AUTH1[🔐 Auth Service<br/>Keycloak Integration]
                CONFIG1[⚙️ Config Service<br/>Consul/etcd]
                METRICS1[📊 Metrics Service<br/>Prometheus]
            end
            subgraph "🔌 Protocol Services"
                WG1[🌐 WireGuard Service<br/>P2P Tunneling]
                QUIC1[⚡ Enhanced QUIC<br/>High Performance]
                P2P1[🕸️ P2P Mesh Service<br/>Peer Discovery]
            end
            subgraph "🔒 Security Services"
                QUANTUM1[🔮 Quantum Crypto<br/>Post-Quantum Algorithms]
                AI1[🤖 AI/ML Service<br/>Threat Detection]
                SIEM1[🛡️ SIEM Service<br/>Security Monitoring]
            end
            subgraph "🎭 Orchestration"
                CADENCE1[🔄 Cadence Cluster<br/>Workflow Engine]
                WORKER1[👷 Worker Pool<br/>Activity Execution]
            end
        end
        
        subgraph "🇪🇺 Region EU-West"
            LB2[🔄 Load Balancer]
            GW3[🌐 API Gateway 3]
            AUTH2[🔐 Auth Service]
            WG2[🌐 WireGuard Service]
            P2P2[🕸️ P2P Mesh Service]
        end
        
        subgraph "🇯🇵 Region Asia-Pacific"
            LB3[🔄 Load Balancer]
            GW4[🌐 API Gateway 4]
            AUTH3[🔐 Auth Service]
            WG3[🌐 WireGuard Service]
            P2P3[🕸️ P2P Mesh Service]
        end
    end
    
    subgraph "💾 Shared Infrastructure"
        subgraph "🗄️ Data Layer"
            COCKROACH[(🦗 CockroachDB<br/>Global Distribution)]
            REDIS[(⚡ Redis Cluster<br/>Caching)]
            KAFKA[(📨 Kafka Streams<br/>Event Bus)]
        end
        subgraph "📈 Monitoring"
            PROM[📊 Prometheus Federation]
            GRAF[📈 Grafana Global]
            JAEGER[🔍 Jaeger Tracing]
        end
        subgraph "🕸️ Service Mesh"
            ISTIO[🕸️ Istio/Consul Connect]
            ENVOY[🔗 Envoy Proxy]
        end
    end
    
    %% Cross-region mesh connections
    P2P1 -.->|"🔗 Global P2P Mesh"| P2P2
    P2P2 -.->|"🔗 Global P2P Mesh"| P2P3
    P2P3 -.->|"🔗 Global P2P Mesh"| P2P1
    
    %% Service connections
    GW1 --> AUTH1
    GW1 --> WG1
    WG1 --> P2P1
    QUANTUM1 --> WG1
    AI1 --> SIEM1
    
    %% Data connections
    AUTH1 --> COCKROACH
    CONFIG1 --> COCKROACH
    METRICS1 --> PROM
    
    %% Orchestration
    CADENCE1 --> WORKER1
    WORKER1 --> WG1
    WORKER1 --> QUANTUM1
```

## 📋 Детальный план миграции

### **🎯 Фаза 1: Подготовка инфраструктуры (2-3 недели)**

#### **Неделя 1: Service Mesh и мониторинг**
```bash
# Установка Istio service mesh
curl -L https://istio.io/downloadIstio | sh -
kubectl apply -f istio-system/

# Настройка Prometheus Federation
kubectl apply -f monitoring/prometheus-federation.yaml

# Настройка Jaeger для distributed tracing
kubectl apply -f tracing/jaeger-operator.yaml
```

#### **Неделя 2-3: CI/CD и безопасность**
```bash
# Настройка GitOps с ArgoCD
kubectl apply -f gitops/argocd-install.yaml

# Установка cert-manager для автоматических сертификатов
kubectl apply -f security/cert-manager.yaml

# Настройка Vault для управления секретами
kubectl apply -f security/vault-operator.yaml
```

### **🔧 Фаза 2: Извлечение микросервисов (4-6 недель)**

#### **Приоритет извлечения согласно Золотым правилам:**

**1. Auth Service (Неделя 1)**
```dockerfile
# auth-service/Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY internal/auth/ ./internal/auth/
COPY cmd/auth-service/ ./cmd/auth-service/
RUN CGO_ENABLED=0 go build -o auth-service ./cmd/auth-service/

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/auth-service .
EXPOSE 8080
CMD ["./auth-service"]
```

**2. Metrics Service (Неделя 1-2)**
```yaml
# metrics-service deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: metrics-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: metrics-service
  template:
    metadata:
      labels:
        app: metrics-service
    spec:
      containers:
      - name: metrics-service
        image: cloudbridge/metrics-service:v2.0
        ports:
        - containerPort: 8080
        - containerPort: 9090  # Prometheus metrics
        env:
        - name: PROMETHEUS_URL
          value: "http://prometheus:9090"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

**3. WireGuard Service (Неделя 2-3)**
```yaml
# wireguard-service с привилегированными правами
apiVersion: apps/v1
kind: DaemonSet  # DaemonSet для работы с сетевыми интерфейсами
metadata:
  name: wireguard-service
spec:
  selector:
    matchLabels:
      app: wireguard-service
  template:
    metadata:
      labels:
        app: wireguard-service
    spec:
      hostNetwork: true  # Доступ к host сети
      containers:
      - name: wireguard-service
        image: cloudbridge/wireguard-service:v2.0
        securityContext:
          privileged: true  # Только для WireGuard сервиса
          capabilities:
            add:
            - NET_ADMIN
            - SYS_MODULE
        volumeMounts:
        - name: dev-net-tun
          mountPath: /dev/net/tun
        - name: lib-modules
          mountPath: /lib/modules
          readOnly: true
        ports:
        - containerPort: 51820
          protocol: UDP
          hostPort: 51820
        env:
        - name: WG_CONFIG_PATH
          value: "/etc/wireguard"
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: dev-net-tun
        hostPath:
          path: /dev/net/tun
      - name: lib-modules
        hostPath:
          path: /lib/modules
```

**4. AI/ML Service (Неделя 3-4)**
```yaml
# ai-ml-service с GPU поддержкой
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-ml-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: ai-ml-service
  template:
    metadata:
      labels:
        app: ai-ml-service
    spec:
      containers:
      - name: ai-ml-service
        image: cloudbridge/ai-ml-service:v2.0
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
            nvidia.com/gpu: 1  # GPU для ML моделей
          limits:
            memory: "4Gi"
            cpu: "2000m"
            nvidia.com/gpu: 1
        env:
        - name: MODEL_PATH
          value: "/models"
        - name: KAFKA_BROKERS
          value: "kafka:9092"
        volumeMounts:
        - name: model-storage
          mountPath: /models
      volumes:
      - name: model-storage
        persistentVolumeClaim:
          claimName: ai-models-pvc
```

**5. Quantum Crypto Service (Неделя 4-5)**
```yaml
# quantum-crypto-service
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quantum-crypto-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: quantum-crypto-service
  template:
    metadata:
      labels:
        app: quantum-crypto-service
    spec:
      containers:
      - name: quantum-crypto-service
        image: cloudbridge/quantum-crypto-service:v2.0
        ports:
        - containerPort: 8080
        env:
        - name: CRYSTALS_KYBER_ENABLED
          value: "true"
        - name: CRYSTALS_DILITHIUM_ENABLED
          value: "true"
        - name: KEY_ROTATION_INTERVAL
          value: "24h"
        resources:
          requests:
            memory: "512Mi"
            cpu: "300m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          readOnlyRootFilesystem: true
```

### **🎭 Фаза 3: Оркестрация и интеграция (2-3 недели)**

#### **Cadence Workflows для микросервисов**
```go
// Workflow для координации между сервисами
func GlobalP2PMeshWorkflow(ctx workflow.Context, input GlobalMeshInput) error {
    logger := workflow.GetLogger(ctx)
    
    // Параллельное выполнение в разных регионах
    futures := []workflow.Future{}
    
    for _, region := range input.Regions {
        future := workflow.ExecuteActivity(ctx, 
            activity.Options{
                TaskList: fmt.Sprintf("wireguard-%s", region),
                ScheduleToCloseTimeout: time.Minute * 5,
                HeartbeatTimeout: time.Second * 30,
            },
            "EstablishRegionalMesh", region)
        futures = append(futures, future)
    }
    
    // Ожидание завершения во всех регионах
    for i, future := range futures {
        var result RegionalMeshResult
        if err := future.Get(ctx, &result); err != nil {
            logger.Error("Regional mesh failed", 
                zap.String("region", input.Regions[i]),
                zap.Error(err))
            return err
        }
        logger.Info("Regional mesh established", 
            zap.String("region", input.Regions[i]),
            zap.Int("peers", result.PeerCount))
    }
    
    return nil
}
```

#### **Event-driven архитектура с Kafka**
```yaml
# kafka-topics для межсервисного взаимодействия
apiVersion: v1
kind: ConfigMap
metadata:
  name: kafka-topics
data:
  topics.yaml: |
    topics:
      - name: "peer-lifecycle-events"
        partitions: 12
        replication: 3
        consumers: ["wireguard-service", "p2p-service", "metrics-service"]
      
      - name: "security-alerts"
        partitions: 6
        replication: 3
        consumers: ["siem-service", "ai-service", "notification-service"]
      
      - name: "quantum-key-rotation"
        partitions: 3
        replication: 3
        consumers: ["quantum-service", "wireguard-service"]
      
      - name: "scaling-decisions"
        partitions: 6
        replication: 3
        consumers: ["all-services"]
```

### **🌍 Фаза 4: Глобальное развертывание (3-4 недели)**

#### **Multi-region Kubernetes clusters**
```yaml
# kustomization.yaml для разных регионов
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

bases:
  - ../base

patchesStrategicMerge:
  - region-specific-config.yaml

configMapGenerator:
  - name: region-config
    literals:
      - REGION=us-east-1
      - ZONE=us-east-1a
      - REPLICAS=5
      - EDGE_LOCATIONS=["new-york", "boston", "washington"]

images:
  - name: cloudbridge/wireguard-service
    newTag: v2.0-us-east
```

#### **Global Load Balancing с DNS**
```yaml
# global-dns-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: global-dns-config
data:
  dns-config.yaml: |
    zones:
      cloudbridge.network:
        records:
          - name: "api"
            type: "A"
            geo_policy: "latency"
            locations:
              us-east-1: "34.203.XX.XX"
              eu-west-1: "34.240.XX.XX"
              ap-southeast-1: "54.255.XX.XX"
          
          - name: "wireguard"
            type: "SRV"
            geo_policy: "weighted"
            locations:
              us-east-1: 
                weight: 50
                target: "wg-us.cloudbridge.network"
                port: 51820
              eu-west-1:
                weight: 30
                target: "wg-eu.cloudbridge.network"
                port: 51820
              ap-southeast-1:
                weight: 20
                target: "wg-ap.cloudbridge.network"
                port: 51820
```

## 🔧 Docker Compose для локальной разработки

### **Микросервисная версия docker-compose.yml**
```yaml
version: '3.8'

services:
  # ===== GATEWAY LAYER =====
  api-gateway:
    image: kong:latest
    container_name: cloudbridge-gateway
    ports:
      - "8080:8080"   # Admin API
      - "8000:8000"   # Proxy
    environment:
      KONG_DATABASE: "off"
      KONG_DECLARATIVE_CONFIG: "/kong/kong.yml"
      KONG_PROXY_ACCESS_LOG: "/dev/stdout"
      KONG_ADMIN_ACCESS_LOG: "/dev/stdout"
      KONG_PROXY_ERROR_LOG: "/dev/stderr"
      KONG_ADMIN_ERROR_LOG: "/dev/stderr"
      KONG_ADMIN_LISTEN: "0.0.0.0:8080"
    volumes:
      - ./configs/kong.yml:/kong/kong.yml:ro
    networks:
      - cloudbridge-network
    healthcheck:
      test: ["CMD", "kong", "health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # ===== CORE SERVICES =====
  auth-service:
    image: cloudbridge/auth-service:v2.0
    container_name: cloudbridge-auth
    ports:
      - "8081:8080"
      - "9091:9090"   # Prometheus metrics
    environment:
      - DATABASE_URL=postgresql://auth:password@postgres:5432/auth
      - REDIS_URL=redis://redis:6379/1
      - KEYCLOAK_URL=http://keycloak:8080/auth
      - LOG_LEVEL=info
    networks:
      - cloudbridge-network
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.1'
          memory: 128M

  config-service:
    image: cloudbridge/config-service:v2.0
    container_name: cloudbridge-config
    ports:
      - "8082:8080"
      - "9092:9090"   # Prometheus metrics
    environment:
      - CONSUL_ADDR=consul:8500
      - LOG_LEVEL=info
    networks:
      - cloudbridge-network
    depends_on:
      - consul
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  metrics-service:
    image: cloudbridge/metrics-service:v2.0
    container_name: cloudbridge-metrics
    ports:
      - "8083:8080"
      - "9093:9090"   # Prometheus metrics
    environment:
      - PROMETHEUS_URL=http://prometheus:9090
      - KAFKA_BROKERS=kafka:9092
      - LOG_LEVEL=info
    networks:
      - cloudbridge-network
    depends_on:
      - prometheus
      - kafka
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # ===== PROTOCOL SERVICES =====
  wireguard-service:
    image: cloudbridge/wireguard-service:v2.0
    container_name: cloudbridge-wireguard
    privileged: true  # Только для WireGuard!
    network_mode: "host"
    ports:
      - "51820:51820/udp"
      - "8084:8080"
      - "9094:9090"   # Prometheus metrics
    environment:
      - WG_CONFIG_PATH=/etc/wireguard
      - P2P_SERVICE_URL=http://p2p-service:8080
      - AUTH_SERVICE_URL=http://auth-service:8080
      - LOG_LEVEL=info
    volumes:
      - /dev/net/tun:/dev/net/tun
      - wireguard-configs:/etc/wireguard
      - /lib/modules:/lib/modules:ro
    networks:
      - cloudbridge-network
    depends_on:
      - auth-service
      - p2p-service
    healthcheck:
      test: ["CMD", "wg", "show"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.2'
          memory: 256M

  quic-service:
    image: cloudbridge/quic-service:v2.0
    container_name: cloudbridge-quic
    ports:
      - "4433:4433/udp"  # QUIC port
      - "8085:8080"
      - "9095:9090"      # Prometheus metrics
    environment:
      - QUIC_CERT_PATH=/certs/server.crt
      - QUIC_KEY_PATH=/certs/server.key
      - QUANTUM_SERVICE_URL=http://quantum-service:8080
      - LOG_LEVEL=info
    volumes:
      - quic-certs:/certs
    networks:
      - cloudbridge-network
    depends_on:
      - quantum-service
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  p2p-service:
    image: cloudbridge/p2p-service:v2.0
    container_name: cloudbridge-p2p
    ports:
      - "8086:8080"
      - "9096:9090"   # Prometheus metrics
    environment:
      - DISCOVERY_MODE=mdns,dht
      - BOOTSTRAP_PEERS=/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ
      - WIREGUARD_SERVICE_URL=http://wireguard-service:8080
      - LOG_LEVEL=info
    networks:
      - cloudbridge-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # ===== SECURITY SERVICES =====
  quantum-service:
    image: cloudbridge/quantum-service:v2.0
    container_name: cloudbridge-quantum
    ports:
      - "8087:8080"
      - "9097:9090"   # Prometheus metrics
    environment:
      - CRYSTALS_KYBER_ENABLED=true
      - CRYSTALS_DILITHIUM_ENABLED=true
      - KEY_ROTATION_INTERVAL=24h
      - VAULT_ADDR=http://vault:8200
      - LOG_LEVEL=info
    networks:
      - cloudbridge-network
    depends_on:
      - vault
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
        reservations:
          cpus: '0.3'
          memory: 512M

  ai-service:
    image: cloudbridge/ai-service:v2.0
    container_name: cloudbridge-ai
    ports:
      - "8088:8080"
      - "9098:9090"   # Prometheus metrics
    environment:
      - MODEL_PATH=/models
      - KAFKA_BROKERS=kafka:9092
      - SIEM_SERVICE_URL=http://siem-service:8080
      - GPU_ENABLED=false  # true для GPU серверов
      - LOG_LEVEL=info
    volumes:
      - ai-models:/models
    networks:
      - cloudbridge-network
    depends_on:
      - kafka
      - siem-service
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 60s
      timeout: 30s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '0.5'
          memory: 1G

  siem-service:
    image: cloudbridge/siem-service:v2.0
    container_name: cloudbridge-siem
    ports:
      - "8089:8080"
      - "9099:9090"   # Prometheus metrics
    environment:
      - ELASTICSEARCH_URL=http://elasticsearch:9200
      - KAFKA_BROKERS=kafka:9092
      - ALERT_WEBHOOK_URL=http://alertmanager:9093/api/v1/alerts
      - LOG_LEVEL=info
    networks:
      - cloudbridge-network
    depends_on:
      - elasticsearch
      - kafka
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # ===== ORCHESTRATION LAYER =====
  cadence-frontend:
    image: ubercadence/frontend:latest
    container_name: cloudbridge-cadence-frontend
    ports:
      - "7933:7933"
      - "7934:7934"
    environment:
      - CADENCE_CLI_ADDRESS=cadence-frontend:7933
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=cadence
      - POSTGRES_PWD=cadence
      - POSTGRES_SEEDS=postgres
      - DYNAMIC_CONFIG_FILE_PATH=config/dynamicconfig/development.yaml
    networks:
      - cloudbridge-network
    depends_on:
      postgres:
        condition: service_healthy

  cadence-history:
    image: ubercadence/history:latest
    container_name: cloudbridge-cadence-history
    environment:
      - CADENCE_CLI_ADDRESS=cadence-frontend:7933
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=cadence
      - POSTGRES_PWD=cadence
      - POSTGRES_SEEDS=postgres
    networks:
      - cloudbridge-network
    depends_on:
      postgres:
        condition: service_healthy

  cadence-matching:
    image: ubercadence/matching:latest
    container_name: cloudbridge-cadence-matching
    environment:
      - CADENCE_CLI_ADDRESS=cadence-frontend:7933
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=cadence
      - POSTGRES_PWD=cadence
      - POSTGRES_SEEDS=postgres
    networks:
      - cloudbridge-network
    depends_on:
      postgres:
        condition: service_healthy

  cadence-worker:
    image: ubercadence/worker:latest
    container_name: cloudbridge-cadence-worker
    environment:
      - CADENCE_CLI_ADDRESS=cadence-frontend:7933
      - DB=postgresql
      - DB_PORT=5432
      - POSTGRES_USER=cadence
      - POSTGRES_PWD=cadence
      - POSTGRES_SEEDS=postgres
    networks:
      - cloudbridge-network
    depends_on:
      cadence-frontend:
        condition: service_started

  cloudbridge-workers:
    image: cloudbridge/cadence-workers:v2.0
    container_name: cloudbridge-workers
    environment:
      - CADENCE_SERVICE=cadence-frontend:7933
      - TASK_LIST=cloudbridge-workers
      - AUTH_SERVICE_URL=http://auth-service:8080
      - WIREGUARD_SERVICE_URL=http://wireguard-service:8080
      - QUANTUM_SERVICE_URL=http://quantum-service:8080
      - AI_SERVICE_URL=http://ai-service:8080
      - LOG_LEVEL=info
    networks:
      - cloudbridge-network
    depends_on:
      - cadence-frontend
      - auth-service
      - wireguard-service
    deploy:
      replicas: 3

  # ===== INFRASTRUCTURE SERVICES =====
  postgres:
    image: postgres:16
    container_name: cloudbridge-postgres
    environment:
      POSTGRES_USER: cloudbridge
      POSTGRES_PASSWORD: cloudbridge
      POSTGRES_DB: cloudbridge
      POSTGRES_INITDB_ARGS: "--encoding=UTF-8"
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./configs/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql:ro
    ports:
      - "5432:5432"
    networks:
      - cloudbridge-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U cloudbridge -d cloudbridge"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    container_name: cloudbridge-redis
    command: redis-server --appendonly yes --maxmemory 1gb --maxmemory-policy allkeys-lru
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - cloudbridge-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  consul:
    image: consul:latest
    container_name: cloudbridge-consul
    command: agent -server -bootstrap -ui -client=0.0.0.0 -data-dir=/consul/data
    ports:
      - "8500:8500"
    volumes:
      - consul-data:/consul/data
    networks:
      - cloudbridge-network
    healthcheck:
      test: ["CMD", "consul", "members"]
      interval: 30s
      timeout: 10s
      retries: 3

  vault:
    image: vault:latest
    container_name: cloudbridge-vault
    cap_add:
      - IPC_LOCK
    ports:
      - "8200:8200"
    environment:
      VAULT_DEV_ROOT_TOKEN_ID: "cloudbridge-dev-token"
      VAULT_DEV_LISTEN_ADDRESS: "0.0.0.0:8200"
    networks:
      - cloudbridge-network
    healthcheck:
      test: ["CMD", "vault", "status"]
      interval: 30s
      timeout: 10s
      retries: 3

  # ===== MESSAGE BROKERS =====
  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    container_name: cloudbridge-zookeeper
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"
    networks:
      - cloudbridge-network

  kafka:
    image: confluentinc/cp-kafka:latest
    container_name: cloudbridge-kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: 'true'
    networks:
      - cloudbridge-network
    depends_on:
      - zookeeper

  # ===== MONITORING STACK =====
  prometheus:
    image: prom/prometheus:latest
    container_name: cloudbridge-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
      - '--web.enable-admin-api'
    networks:
      - cloudbridge-network

  grafana:
    image: grafana/grafana:latest
    container_name: cloudbridge-grafana
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
      GF_USERS_ALLOW_SIGN_UP: false
      GF_INSTALL_PLUGINS: grafana-piechart-panel,grafana-worldmap-panel
    volumes:
      - grafana-data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources:ro
    networks:
      - cloudbridge-network
    depends_on:
      - prometheus

  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: cloudbridge-jaeger
    ports:
      - "14268:14268"   # jaeger.thrift
      - "16686:16686"   # Web UI
    environment:
      COLLECTOR_ZIPKIN_HTTP_PORT: 9411
    networks:
      - cloudbridge-network

  alertmanager:
    image: prom/alertmanager:latest
    container_name: cloudbridge-alertmanager
    ports:
      - "9093:9093"
    volumes:
      - ./monitoring/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
      - '--web.external-url=http://localhost:9093'
    networks:
      - cloudbridge-network

  # ===== SEARCH AND ANALYTICS =====
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    container_name: cloudbridge-elasticsearch
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms1g -Xmx1g"
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch-data:/usr/share/elasticsearch/data
    networks:
      - cloudbridge-network
    healthcheck:
      test: ["CMD-SHELL", "curl -f http://localhost:9200/_cluster/health || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3

  kibana:
    image: docker.elastic.co/kibana/kibana:8.11.0
    container_name: cloudbridge-kibana
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    ports:
      - "5601:5601"
    networks:
      - cloudbridge-network
    depends_on:
      elasticsearch:
        condition: service_healthy

networks:
  cloudbridge-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16

volumes:
  postgres-data:
  redis-data:
  consul-data:
  prometheus-data:
  grafana-data:
  elasticsearch-data:
  ai-models:
  wireguard-configs:
  quic-certs:
```

## 🎭 Применение Золотых правил в микросервисах

### **🥇 Правило #1: Отказоустойчивость**
```yaml
# Каждый сервис имеет health checks
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s

# Circuit breaker в каждом сервисе
environment:
  - CIRCUIT_BREAKER_THRESHOLD=5
  - CIRCUIT_BREAKER_TIMEOUT=30s
  - GRACEFUL_SHUTDOWN_TIMEOUT=30s
```

### **🥈 Правило #2: Безопасность**
```yaml
# Mutual TLS между сервисами
volumes:
  - ./certs/service.crt:/certs/service.crt:ro
  - ./certs/service.key:/certs/service.key:ro
  - ./certs/ca.crt:/certs/ca.crt:ro

# Принцип least privilege
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

### **🥉 Правило #3: Наблюдаемость**
```yaml
# Каждый сервис экспортирует метрики
ports:
  - "9090:9090"   # Prometheus metrics endpoint

# Structured logging
environment:
  - LOG_FORMAT=json
  - LOG_LEVEL=info
  - TRACE_ENABLED=true
  - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

### **🏆 Правило #4: Автоматизация**
```bash
# Автоматическое развертывание через GitOps
git commit -m "feat: update wireguard-service to v2.1"
git push origin main

# ArgoCD автоматически применяет изменения
# Canary deployment с автоматическим rollback при ошибках
```

### **💎 Правило #5: Производительность**
```yaml
# Horizontal Pod Autoscaler для каждого сервиса
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: wireguard-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: wireguard-service
  minReplicas: 2
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## 🚀 Команды для запуска

### **Локальная разработка:**
```bash
# Запуск всех микросервисов
docker-compose up -d

# Проверка здоровья всех сервисов
docker-compose ps

# Просмотр логов конкретного сервиса
docker-compose logs -f wireguard-service

# Масштабирование сервиса
docker-compose up --scale ai-service=3 -d
```

### **Production развертывание:**
```bash
# Развертывание в Kubernetes
kubectl apply -f k8s/

# Проверка статуса всех сервисов
kubectl get pods -l app.kubernetes.io/name=cloudbridge

# Просмотр метрик
kubectl port-forward svc/prometheus 9090:9090
open http://localhost:9090

# Distributed tracing
kubectl port-forward svc/jaeger 16686:16686
open http://localhost:16686
```

## 📊 Мониторинг микросервисной архитектуры

### **Ключевые метрики:**
```promql
# Availability всех сервисов
avg(up{job=~"cloudbridge-.*"}) * 100

# Latency P99 для API Gateway
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job="api-gateway"}[5m]))

# Error rate для всех сервисов
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100

# Throughput для WireGuard
rate(wireguard_peers_total[5m])
```

## 🎯 Выводы и рекомендации

### **Преимущества микросервисной архитектуры:**
1. ✅ **Независимое масштабирование** каждого компонента
2. ✅ **Изоляция сбоев** - падение одного сервиса не влияет на другие
3. ✅ **Технологическое разнообразие** - каждый сервис может использовать оптимальные инструменты
4. ✅ **Быстрое развертывание** - независимые релизы
5. ✅ **Команды разработки** могут работать параллельно

### **Следование Золотым правилам:**
- **Отказоустойчивость** встроена в каждый микросервис
- **Безопасность** реализована на каждом уровне
- **Наблюдаемость** обеспечена для всех компонентов
- **Автоматизация** покрывает весь жизненный цикл
- **Производительность** оптимизирована для каждого сервиса

### **Следующие шаги:**
1. 🎯 Начать с извлечения Auth Service
2. 🔧 Настроить CI/CD для микросервисов
3. 📊 Внедрить комплексный мониторинг
4. 🌍 Запланировать multi-region развертывание

Эта архитектура обеспечит CloudBridge Relay способность масштабироваться до глобальной сети с миллионами P2P и P2C клиентов! 🚀
