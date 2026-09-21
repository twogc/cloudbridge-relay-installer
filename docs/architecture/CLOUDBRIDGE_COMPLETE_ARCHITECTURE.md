# CloudBridge Complete Architecture

## ASCII Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ CloudBridge Platform Architecture │
├─────────────────────────────────────────────────────────────────────────────────┤
│ │
│ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│ │ Control Plane │ │ AI Service │ │ DNS Network │ │
│ │ │ │ │ │ │ │
│ │ • Zitadel IdP │ │ • ML Analytics │ │ • GeoDNS │ │
│ │ • Cadence │ │ • Route Opt │ │ • Anycast │ │
│ │ • PostgreSQL │ │ • Anomaly Det │ │ • DNSSEC │ │
│ │ • Redis │ │ • Prediction │ │ • Load Bal │ │
│ └─────────────────┘ └─────────────────┘ └─────────────────┘ │
│ │ │ │ │
│ └───────────────────────┼───────────────────────┘ │
│ │ │
│ ┌─────────────────────────────────┼─────────────────────────────────┐ │
│ │ Edge PoP Layer │ │
│ │ │ │
│ │ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │ │
│ │ │ Moscow PoP │ │Frankfurt PoP│ │Amsterdam PoP│ │ │
│ │ │ RU-MOW │ │ EU-FRA │ │ EU-AMS │ │ │
│ │ │ │ │ │ │ │ │ │
│ │ │ QUIC: 443 │ │ QUIC: 443 │ │ QUIC: 443 │ │ │
│ │ │ WG: 51820 │ │ WG: 51820 │ │ WG: 51820 │ │ │
│ │ │ TURN: 3478 │ │ TURN: 3478 │ │ TURN: 3478 │ │ │
│ │ └─────────────┘ └─────────────┘ └─────────────┘ │ │
│ │ │ │ │ │ │
│ │ └───────────────────┼───────────────────┘ │ │
│ │ │ │ │
│ │ ┌───────────────────────────┼───────────────────────────┐ │ │
│ │ │ Relay Layer │ │ │
│ │ │ │ │ │
│ │ │ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │ │
│ │ │ │CB Relay MOW │ │CB Relay FRA │ │CB Relay AMS │ │ │
│ │ │ │ │ │ │ │ │ │ │
│ │ │ │ Port: 9090 │ │ Port: 9090 │ │ Port: 9090 │ │ │
│ │ │ │ MASQUE/QUIC │ │ MASQUE/QUIC │ │ MASQUE/QUIC │ │ │
│ │ │ │ P2P Mesh │ │ P2P Mesh │ │ P2P Mesh │ │ │
│ │ │ └─────────────┘ └─────────────┘ └─────────────┘ │ │
│ │ │ │ │ │ │ │
│ │ │ └───────────────────┼───────────────────┘ │ │
│ │ │ │ │ │
│ │ │ ┌───────────────────┼───────────────────┐ │ │
│ │ │ │ Inter-PoP MASQUE/QUIC Tunnels │ │ │
│ │ │ │ mTLS Secured │ │ │
│ │ │ │ Smart Routing │ │ │
│ │ │ └───────────────────────────────────────┘ │ │
│ │ └──────────────────────────────────────────────────────────┘ │
│ └─────────────────────────────────────────────────────────────────────────┘
│ │
│ ┌─────────────────────────────────────────────────────────────────────────┐
│ │ Client Layer │
│ │ │
│ │ ┌─────────────────┐ ┌─────────────────┐ │
│ │ │ CloudBridge │ │ Yandex.Browser │ │
│ │ │ Client │ │ │ │
│ │ │ │ │ • Chromium Core │ │
│ │ │ • Windows │ │ • HTTP/3 Client │ │
│ │ │ • Linux │ │ • MASQUE Client │ │
│ │ │ • macOS │ │ • QUIC Support │ │
│ │ │ │ │ │ │
│ │ │ • P2P Connect │ │ • Web Services │ │
│ │ │ • MASQUE/QUIC │ │ • Dzen, Search │ │
│ │ │ • WireGuard │ │ • Mail, API │ │
│ │ └─────────────────┘ └─────────────────┘ │
│ └─────────────────────────────────────────────────────────────────────────┘
│ │
│ ┌─────────────────────────────────────────────────────────────────────────┐
│ │ Monitoring Layer │
│ │ │
│ │ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│ │ │ Prometheus │ │ Grafana │ │ AlertManager │ │
│ │ │ │ │ │ │ │ │
│ │ │ • Metrics Coll │ │ • Dashboards │ │ • Alert Routing │ │
│ │ │ • Time Series │ │ • Visualization │ │ • Notifications │ │
│ │ │ • Query Engine │ │ • Alerting │ │ • Escalation │ │
│ │ └─────────────────┘ └─────────────────┘ └─────────────────┘ │
│ └─────────────────────────────────────────────────────────────────────────┘
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Data Flow Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Data Flow Overview │
├─────────────────────────────────────────────────────────────────────────────────┤
│ │
│ Client Request Flow: │
│ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │
│ │ Client │───▶│ DNS │───▶│ GeoDNS │───▶│ PoP │───▶│ Relay │ │
│ │ │ │ Query │ │ Steering │ │ Selection│ │ Server │ │
│ └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │
│ │ │ │
│ │ ▼ │
│ │ ┌─────────┐ │
│ │ │Control │ │
│ │ │ Plane │ │
│ │ │ Auth │ │
│ │ └─────────┘ │
│ │ │ │
│ │ ▼ │
│ │ ┌─────────┐ │
│ │ │ JWT │ │
│ │ │ Token │ │
│ │ └─────────┘ │
│ │ │ │
│ │ ▼ │
│ │ ┌─────────┐ │
│ │ │P2P │ │
│ │ │Connection│ │
│ │ │Established│ │
│ └─────────────────────────────────────────────────────────└─────────┘ │
│ │
│ Inter-PoP Communication: │
│ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │
│ │ Moscow │◀──▶│Frankfurt│◀──▶│Amsterdam│◀──▶│ Moscow │◀──▶│ AI │ │
│ │ PoP │ │ PoP │ │ PoP │ │ PoP │ │ Service │ │
│ │ │ │ │ │ │ │ │ │ │ │
│ │MASQUE/ │ │MASQUE/ │ │MASQUE/ │ │MASQUE/ │ │Route │ │
│ │QUIC │ │QUIC │ │QUIC │ │QUIC │ │Optimize │ │
│ │mTLS │ │mTLS │ │mTLS │ │mTLS │ │ │ │
│ └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │
│ │ │ │ │ │ │
│ └───────────────┼───────────────┼───────────────┼───────────────┘ │
│ │ │ │ │
│ ▼ ▼ ▼ │
│ ┌─────────┐ ┌─────────┐ ┌─────────┐ │
│ │Metrics │ │Metrics │ │Metrics │ │
│ │RTT/Loss │ │RTT/Loss │ │RTT/Loss │ │
│ │Jitter │ │Jitter │ │Jitter │ │
│ └─────────┘ └─────────┘ └─────────┘ │
│ │
│ Monitoring Flow: │
│ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │
│ │ Relay │───▶│Prometheus│───▶│ Grafana │───▶│AlertMgr │───▶│ AI │ │
│ │ Metrics │ │ │ │ │ │ │ │ Service │ │
│ │ │ │Time │ │Dashboards│ │Alert │ │ │ │
│ │RTT/Loss │ │Series │ │Viz │ │Routing │ │Optimize │ │
│ │Jitter │ │DB │ │Alerting │ │Notify │ │Action │ │
│ │Throughput│ │ │ │ │ │ │ │ │ │
│ └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Network Topology

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Network Topology Overview │
├─────────────────────────────────────────────────────────────────────────────────┤
│ │
│ Internet │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ │ │
│ │ ┌─────────────────┐ ┌─────────────────┐ │ │
│ │ │ Clients │ │ Web Services │ │ │
│ │ │ │ │ │ │ │
│ │ │ • Windows │ │ • Dzen │ │ │
│ │ │ • Linux │ │ • Search │ │ │
│ │ │ • macOS │ │ • Mail │ │ │
│ │ │ • Yandex.Br │ │ • API │ │ │
│ │ └─────────────────┘ └─────────────────┘ │ │
│ │ │ │ │ │
│ │ │ │ │ │
│ │ ▼ ▼ │ │
│ │ ┌─────────────────────────────────────────────────────────────────┐ │ │
│ │ │ CloudBridge Platform │ │ │
│ │ │ │ │ │
│ │ │ ┌─────────────────────────────────────────────────────────┐ │ │ │
│ │ │ │ Control Plane │ │ │ │
│ │ │ │ │ │ │ │
│ │ │ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │ │ │
│ │ │ │ │Zitadel │ │Cadence │ │PostgreSQL│ │ Redis │ │ │ │ │
│ │ │ │ │IdP │ │Workflow │ │ │ │ │ │ │ │ │
│ │ │ │ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │ │ │ │
│ │ │ │ │ │ │ │ │ │ │ │
│ │ │ │ └─────────────┼─────────────┼─────────────┘ │ │ │ │
│ │ │ │ │ │ │ │ │ │
│ │ │ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │ │ │
│ │ │ │ │AI Service│ │DNS │ │Monitoring│ │Load │ │ │ │ │
│ │ │ │ │ │ │Network │ │ │ │Balancer │ │ │ │ │
│ │ │ │ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │ │ │ │
│ │ │ └─────────────────────────────────────────────────────────┘ │ │ │
│ │ │ │ │ │
│ │ │ ┌─────────────────────────────────────────────────────────┐ │ │ │
│ │ │ │ Edge PoP Layer │ │ │ │
│ │ │ │ │ │ │ │
│ │ │ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │ │ │
│ │ │ │ │ Moscow │ │Frankfurt│ │Amsterdam│ │ │ │ │
│ │ │ │ │ RU-MOW │ │ EU-FRA │ │ EU-AMS │ │ │ │ │
│ │ │ │ │ │ │ │ │ │ │ │ │ │
│ │ │ │ │ QUIC │ │ QUIC │ │ QUIC │ │ │ │ │
│ │ │ │ │ WG │ │ WG │ │ WG │ │ │ │ │
│ │ │ │ │ TURN │ │ TURN │ │ TURN │ │ │ │ │
│ │ │ │ └─────────┘ └─────────┘ └─────────┘ │ │ │ │
│ │ │ │ │ │ │ │ │ │ │
│ │ │ │ └───────────────┼───────────────┘ │ │ │ │
│ │ │ │ │ │ │ │ │
│ │ │ │ ┌─────────────────────┼─────────────────────┐ │ │ │ │
│ │ │ │ │ CloudBridge Relay │ │ │ │ │
│ │ │ │ │ │ │ │ │ │
│ │ │ │ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │ │ │ │
│ │ │ │ │ │Relay MOW│ │Relay FRA│ │Relay AMS│ │ │ │ │ │
│ │ │ │ │ │ │ │ │ │ │ │ │ │ │ │
│ │ │ │ │ │Port:9090│ │Port:9090│ │Port:9090│ │ │ │ │ │
│ │ │ │ │ │MASQUE │ │MASQUE │ │MASQUE │ │ │ │ │ │
│ │ │ │ │ │P2P Mesh │ │P2P Mesh │ │P2P Mesh │ │ │ │ │ │
│ │ │ │ │ └─────────┘ └─────────┘ └─────────┘ │ │ │ │ │
│ │ │ │ │ │ │ │ │ │ │ │ │
│ │ │ │ │ └───────────────┼───────────────┘ │ │ │ │ │
│ │ │ │ │ │ │ │ │ │ │
│ │ │ │ │ ┌─────────────┼─────────────┐ │ │ │ │ │
│ │ │ │ │ │ Inter-PoP MASQUE/QUIC │ │ │ │ │ │
│ │ │ │ │ │ mTLS Secured │ │ │ │ │ │
│ │ │ │ │ │ Smart Routing │ │ │ │ │ │
│ │ │ │ │ └───────────────────────────┘ │ │ │ │ │
│ │ │ │ └─────────────────────────────────────────────┘ │ │ │ │
│ │ │ └─────────────────────────────────────────────────────────┘ │ │ │
│ │ └─────────────────────────────────────────────────────────────────┘ │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Общая архитектура CloudBridge Platform

```mermaid
flowchart TB
    %% Клиентские серверы (производственные площадки)
    subgraph " Производственные площадки"
        direction TB
        A[" Сервер A<br/>Москва (RU-MOW)<br/>10.100.77.x<br/> Офис компании"]
        B[" Сервер B<br/>Хабаровск (RU-KHV)<br/>10.100.79.x<br/>Завод"]
    end

    %% Edge слой - транспорт и туннелирование
    subgraph "Edge / Transport Layer"
        direction TB
        subgraph "NAT Traversal"
            STUN[" STUN Server<br/>:19302/UDP<br/>NAT Discovery"]
            TURN[" TURN Server<br/>:3478/UDP<br/>Relay Fallback"]
        end

        subgraph "Edge PoP Moscow"
            MOW["Edge PoP MOW<br/> mow.pop.cb:443<br/>QUIC/MASQUE :443/UDP<br/>WireGuard :51820/UDP<br/>Metrics & Health"]
        end

        subgraph "MASQUE Protocol"
            MASQUE["MASQUE/QUIC Context<br/>CONNECT-UDP<br/>CONNECT-IP<br/> Context ID Management"]
        end
    end

    %% Control Plane - управление и безопасность
    subgraph " Control Plane"
        direction TB
        subgraph "Identity & Auth"
            ZIT["Zitadel IdP<br/>OIDC/JWT (JWKS)<br/>User Management<br/>RBAC Policies"]
        end

        subgraph "CloudBridge Core"
            RELAY["CloudBridge Relay Control<br/>gRPC :8444<br/>REST API :8443<br/>Policy Engine<br/> Connection Manager"]
        end

        subgraph "Monitoring & Discovery"
            MON["Monitoring Stack<br/>Prometheus<br/>Grafana<br/>AlertManager"]
            DNS[" Anycast/DNS<br/>PoP Discovery<br/>GeoDNS Steering<br/>Load Balancing"]
        end
    end

    %% Соединения - основные потоки данных
    %% Прямое P2P соединение (приоритет)
    A -.->|"Direct P2P<br/>(ICE Success)"| B

    %% Fallback #1: Через Edge PoP
    A -->|" Fallback #1<br/>QUIC/MASQUE"| MOW
    MOW -->|"MASQUE Context<br/>Context ID"| B

    %% Fallback #2: Через TURN (симметричный NAT)
    A -.->|" Fallback #2<br/>(Symmetric NAT)"| TURN
    TURN -.->|"TURN Relay"| B

    %% NAT Discovery
    A -->|"NAT Discovery"| STUN
    B -->|"NAT Discovery"| STUN

    %% MASQUE Protocol
    MOW <-->|"MASQUE Protocol"| MASQUE

    %% Аутентификация
    A -->|"OIDC/JWT"| ZIT
    B -->|"OIDC/JWT"| ZIT

    %% Регистрация и политики
    A -->|"Registration<br/>Policies"| RELAY
    B -->|"Registration<br/>Policies"| RELAY

    %% Мониторинг
    RELAY -->|"Metrics"| MON
    MOW -->|"Metrics"| MON

    %% Discovery
    DNS -->|"PoP Discovery"| MOW
    DNS -->|"Service Discovery"| RELAY

    %% Zitadel интеграция
    ZIT -->|"JWT Validation"| RELAY

    %% Стили для визуального разделения
    classDef site fill:#fff7ed,stroke:#ea580c,stroke-width:3px,color:#000
    classDef edge fill:#f0fdf4,stroke:#16a34a,stroke-width:2px,color:#000
    classDef ctrl fill:#eff6ff,stroke:#2563eb,stroke-width:2px,color:#000
    classDef nat fill:#fef3c7,stroke:#d97706,stroke-width:2px,color:#000
    classDef masque fill:#f3e8ff,stroke:#9333ea,stroke-width:2px,color:#000

    class A,B site
    class MOW edge
    class STUN,TURN nat
    class MASQUE masque
    class ZIT,RELAY,MON,DNS ctrl
```

## Детальная архитектура компонентов

### 1. CloudBridge Control Plane
```mermaid
graph TB
    subgraph "Control Plane Components"
        ZT[Zitadel IdP<br/>Authentication<br/>JWT/OIDC<br/>RBAC]
        CD[Cadence Workflow<br/>Orchestration<br/>Auto-scaling<br/>Deployment]
        PG[PostgreSQL<br/>State Storage<br/>Tenant Data<br/>Configuration]
        RD[Redis<br/>Cache<br/>Sessions<br/>Temporary Data]
    end

    ZT --> CD
    CD --> PG
    CD --> RD
    ZT --> PG
    ZT --> RD
```

### 2. CloudBridge AI Service
```mermaid
graph TB
    subgraph "AI Service Components"
        ML[ML Analytics<br/>Route Optimization<br/>Performance Analysis]
        AD[Anomaly Detection<br/>Real-time Monitoring<br/>Alert Generation]
        PE[Prediction Engine<br/>Load Forecasting<br/>Capacity Planning]
        RO[Route Optimizer<br/>Inter-PoP Routing<br/>Smart Routing]
    end

    ML --> AD
    AD --> PE
    PE --> RO
    ML --> RO
```

### 3. CloudBridge DNS Network
```mermaid
graph TB
    subgraph "DNS Network Components"
        GD[GeoDNS Steering<br/>Geographic Routing<br/>Client Selection]
        AR[Anycast Routing<br/>BGP Advertisement<br/>Traffic Distribution]
        DS[DNSSEC<br/>Security<br/>Authentication]
        LB[Load Balancing<br/>Health Checks<br/>Failover]
    end

    GD --> AR
    AR --> DS
    DS --> LB
    GD --> LB
```

### 4. CloudBridge Edge PoP
```mermaid
graph TB
    subgraph "Edge PoP Components"
        QS[QUIC Server<br/>Port: 443/UDP<br/>MASQUE Support<br/>BBRv2]
        WS[WireGuard Server<br/>Port: 51820/UDP<br/>P2P Tunneling<br/>Encryption]
        TS[TURN Server<br/>Port: 3478/UDP<br/>NAT Traversal<br/>ICE Support]
        IP[IPAM Service<br/>Tenant IP Management<br/>Deterministic Allocation]
    end

    QS --> WS
    WS --> TS
    TS --> IP
    QS --> IP
```

### 5. CloudBridge Relay
```mermaid
graph TB
    subgraph "Relay Components"
        MS[MASQUE/QUIC<br/>Inter-PoP Tunneling<br/>mTLS Security<br/>Smart Routing]
        PM[P2P Mesh<br/>Client Connections<br/>Tenant Isolation<br/>VRF]
        TP[Tenant Policy<br/>Calico Network Policies<br/>Isolation Rules<br/>Security]
        AI[AI Integration<br/>Metrics Collection<br/>Route Optimization<br/>SLO Feedback]
    end

    MS --> PM
    PM --> TP
    TP --> AI
    MS --> AI
```

### 6. CloudBridge Monitoring
```mermaid
graph TB
    subgraph "Monitoring Components"
        PM[Prometheus<br/>Metrics Collection<br/>Time Series DB<br/>Query Engine]
        GF[Grafana<br/>Dashboards<br/>Visualization<br/>Alerting]
        AM[AlertManager<br/>Alert Routing<br/>Notification<br/>Escalation]
        MC[Metrics Collection<br/>RTT, Loss, Jitter<br/>Throughput<br/>SLO Metrics]
    end

    PM --> GF
    GF --> AM
    AM --> MC
    PM --> MC
```

## Поток данных и взаимодействие

### 1. Client Connection Flow
```mermaid
sequenceDiagram
    participant C as Client
    participant D as DNS Network
    participant R as Relay
    participant CP as Control Plane
    participant AI as AI Service

    C->>D: DNS Query (client.cb.ru)
    D->>C: GeoDNS Response (nearest PoP)
    C->>R: QUIC Connection
    R->>CP: Authentication Request
    CP->>R: JWT Token
    R->>C: Connection Established
    R->>AI: Metrics Collection
    AI->>R: Route Optimization
```

### 2. Inter-PoP Communication
```mermaid
sequenceDiagram
    participant M as Moscow PoP
    participant F as Frankfurt PoP
    participant A as Amsterdam PoP
    participant AI as AI Service

    M->>F: MASQUE/QUIC Tunnel
    F->>A: MASQUE/QUIC Tunnel
    A->>M: MASQUE/QUIC Tunnel

    M->>AI: RTT Metrics
    F->>AI: Loss Metrics
    A->>AI: Jitter Metrics

    AI->>M: Route Optimization
    AI->>F: Route Optimization
    AI->>A: Route Optimization
```

### 3. Monitoring and Alerting
```mermaid
sequenceDiagram
    participant R as Relay
    participant P as Prometheus
    participant G as Grafana
    participant A as AlertManager
    participant AI as AI Service

    R->>P: Metrics Export
    P->>G: Data Query
    G->>A: Alert Trigger
    A->>AI: Alert Notification
    AI->>R: Optimization Action
```

## Ключевые особенности архитектуры

### **CloudBridge Relay - Центральный компонент**
- **Расположение**: В каждом Edge PoP узле
- **Функции**: P2P mesh networking, MASQUE/QUIC tunneling, tenant isolation
- **Порты**: 9090/UDP (QUIC), 8444/TCP (gRPC), 51820/UDP (WireGuard)
- **Интеграция**: С Control Plane, AI Service, Monitoring

### **Inter-PoP Mesh Network**
- **Протокол**: MASQUE/QUIC с mTLS
- **Маршрутизация**: AI-оптимизированная на основе телеметрии
- **Изоляция**: Per-tenant VRF и Calico Network Policies
- **Масштабирование**: Горизонтальное и вертикальное

### **AI/ML Intelligence**
- **Route Optimization**: Динамическая оптимизация маршрутов
- **Anomaly Detection**: Обнаружение аномалий в реальном времени
- **Load Forecasting**: Предсказание нагрузки и планирование ресурсов
- **SLO Feedback**: Замкнутый контур оптимизации

### **Security & Compliance**
- **mTLS**: Все меж-PoP соединения
- **Tenant Isolation**: Полная изоляция на сетевом уровне
- **DNSSEC**: Безопасность DNS
- **Audit Logging**: Полное логирование операций

## Component Interaction Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Component Interaction Flow │
├─────────────────────────────────────────────────────────────────────────────────┤
│ │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ Client Connection Sequence │ │
│ │ │ │
│ │ 1. Client Request │ │
│ │ ┌─────────┐ │ │
│ │ │ Client │───▶ DNS Query (client.cb.ru) │ │
│ │ └─────────┘ │ │
│ │ │ │
│ │ 2. DNS Resolution │ │
│ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │
│ │ │ DNS │───▶│ GeoDNS │───▶│ Nearest │ │ │
│ │ │ Server │ │ Steering │ │ PoP │ │ │
│ │ └─────────┘ └─────────┘ └─────────┘ │ │
│ │ │ │
│ │ 3. PoP Selection │ │
│ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │
│ │ │ Moscow │ │Frankfurt│ │Amsterdam│ │ │
│ │ │ RU-MOW │ │ EU-FRA │ │ EU-AMS │ │ │
│ │ └─────────┘ └─────────┘ └─────────┘ │ │
│ │ ▲ │ │
│ │ │ Selected based on: │ │
│ │ │ • Geographic proximity │ │
│ │ │ • Load balancing │ │
│ │ │ • Health status │ │
│ │ │ │
│ │ 4. Relay Connection │ │
│ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │
│ │ │ Client │───▶│ Relay │───▶│Control │ │ │
│ │ │ │ │ Server │ │ Plane │ │ │
│ │ └─────────┘ └─────────┘ └─────────┘ │ │
│ │ │ │
│ │ 5. Authentication │ │
│ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │
│ │ │ Relay │───▶│Zitadel │───▶│ JWT │ │ │
│ │ │ Server │ │ IdP │ │ Token │ │ │
│ │ └─────────┘ └─────────┘ └─────────┘ │ │
│ │ │ │
│ │ 6. P2P Connection Established │ │
│ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │
│ │ │ Client │◀──▶│ Relay │◀──▶│ Other │ │ │
│ │ │ │ │ Server │ │ Clients │ │ │
│ │ └─────────┘ └─────────┘ └─────────┘ │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│ │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ Inter-PoP Communication Flow │ │
│ │ │ │
│ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │
│ │ │ Moscow │◀──▶│Frankfurt│◀──▶│Amsterdam│◀──▶│ Moscow │ │ │
│ │ │ PoP │ │ PoP │ │ PoP │ │ PoP │ │ │
│ │ │ │ │ │ │ │ │ │ │ │
│ │ │MASQUE/ │ │MASQUE/ │ │MASQUE/ │ │MASQUE/ │ │ │
│ │ │QUIC │ │QUIC │ │QUIC │ │QUIC │ │ │
│ │ │mTLS │ │mTLS │ │mTLS │ │mTLS │ │ │
│ │ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │ │
│ │ │ │ │ │ │ │
│ │ └───────────────┼───────────────┼───────────────┘ │ │
│ │ │ │ │ │
│ │ ▼ ▼ │ │
│ │ ┌─────────┐ ┌─────────┐ │ │
│ │ │Metrics │ │Metrics │ │ │
│ │ │RTT/Loss │ │RTT/Loss │ │ │
│ │ │Jitter │ │Jitter │ │ │
│ │ └─────────┘ └─────────┘ │ │
│ │ │ │
│ │ ┌─────────────────────────────────────────────────────────────────┐ │ │
│ │ │ AI Service Integration │ │ │
│ │ │ │ │ │
│ │ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │ │
│ │ │ │Metrics │───▶│AI │───▶│Route │───▶│Optimize │ │ │ │
│ │ │ │Collection│ │Analytics│ │Analysis │ │Action │ │ │ │
│ │ │ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │ │ │
│ │ │ │ │ │ │ │ │ │
│ │ │ ▼ ▼ ▼ ▼ │ │ │
│ │ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │ │
│ │ │ │RTT/Loss │ │ML │ │Smart │ │Route │ │ │ │
│ │ │ │Jitter │ │Models │ │Routing │ │Update │ │ │ │
│ │ │ │Throughput│ │Training │ │Decision │ │Applied │ │ │ │
│ │ │ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │ │ │
│ │ └─────────────────────────────────────────────────────────────────┘ │ │
│ └─────────────────────────────────────────────────────────────────┘ │
│ │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ Monitoring and Alerting Flow │ │
│ │ │ │
│ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │
│ │ │ Relay │───▶│Prometheus│───▶│ Grafana │───▶│AlertMgr │ │ │
│ │ │ Metrics │ │ │ │ │ │ │ │ │
│ │ │ │ │Time │ │Dashboards│ │Alert │ │ │
│ │ │RTT/Loss │ │Series │ │Viz │ │Routing │ │ │
│ │ │Jitter │ │DB │ │Alerting │ │Notify │ │ │
│ │ │Throughput│ │ │ │ │ │ │ │ │
│ │ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │ │
│ │ │ │ │ │ │ │
│ │ └───────────────┼───────────────┼───────────────┘ │ │
│ │ │ │ │ │
│ │ ▼ ▼ │ │
│ │ ┌─────────┐ ┌─────────┐ │ │
│ │ │SLO │ │Alert │ │ │
│ │ │Metrics │ │Actions │ │ │
│ │ │Tracking │ │Executed │ │ │
│ │ └─────────┘ └─────────┘ │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Security Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Security Architecture Overview │
├─────────────────────────────────────────────────────────────────────────────────┤
│ │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ Security Layers │ │
│ │ │ │
│ │ ┌─────────────────────────────────────────────────────────────────┐ │ │
│ │ │ Authentication Layer │ │ │
│ │ │ │ │ │
│ │ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │ │
│ │ │ │Zitadel │ │ JWT │ │ OIDC │ │ RBAC │ │ │ │
│ │ │ │IdP │ │ Tokens │ │ Protocol│ │ Access │ │ │ │
│ │ │ │ │ │ │ │ │ │ Control │ │ │ │
│ │ │ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │ │ │
│ │ └─────────────────────────────────────────────────────────────────┘ │ │
│ │ │ │
│ │ ┌─────────────────────────────────────────────────────────────────┐ │ │
│ │ │ Encryption Layer │ │ │
│ │ │ │ │ │
│ │ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │ │
│ │ │ │ TLS 1.3 │ │ mTLS │ │WireGuard│ │ MASQUE │ │ │ │
│ │ │ │ │ │ │ │ │ │ │ │ │ │
│ │ │ │Client │ │Inter-PoP│ │P2P │ │Tunneling│ │ │ │
│ │ │ │Server │ │Comm │ │Tunnels │ │ │ │ │ │
│ │ │ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │ │ │
│ │ └─────────────────────────────────────────────────────────────────┘ │ │
│ │ │ │
│ │ ┌─────────────────────────────────────────────────────────────────┐ │ │
│ │ │ Network Security Layer │ │ │
│ │ │ │ │ │
│ │ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │ │
│ │ │ │Calico │ │Network │ │Firewall │ │DNSSEC │ │ │ │
│ │ │ │Policies │ │Isolation│ │Rules │ │Security │ │ │ │
│ │ │ │ │ │ │ │ │ │ │ │ │ │
│ │ │ │Tenant │ │VRF │ │Ingress │ │DNS │ │ │ │
│ │ │ │Isolation│ │Routing │ │Egress │ │Auth │ │ │ │
│ │ │ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │ │ │
│ │ └─────────────────────────────────────────────────────────────────┘ │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│ │
│ ┌─────────────────────────────────────────────────────────────────────────┐ │
│ │ Security Monitoring │ │
│ │ │ │
│ │ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ │ │
│ │ │Security │ │Audit │ │Threat │ │Incident │ │ │
│ │ │Events │ │Logging │ │Detection│ │Response │ │ │
│ │ │ │ │ │ │ │ │ │ │ │
│ │ │Auth │ │All │ │Anomaly │ │Alert │ │ │
│ │ │Failures │ │Actions │ │Detection│ │Escalation│ │ │
│ │ │Access │ │Tracked │ │ML-based │ │Auto │ │ │
│ │ │Attempts │ │Logged │ │Analysis │ │Response │ │ │
│ │ └─────────┘ └─────────┘ └─────────┘ └─────────┘ │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Заключение

**CloudBridge Relay** является центральным компонентом архитектуры, обеспечивающим:

1. **P2P mesh networking** между клиентами
2. **Inter-PoP MASQUE tunneling** для глобальной связности
3. **Tenant isolation** и безопасность
4. **AI-оптимизированную маршрутизацию**
5. **Интеграцию** со всеми компонентами платформы

Архитектура обеспечивает масштабируемость, отказоустойчивость и высокую производительность для enterprise-grade P2P mesh networking решения.
