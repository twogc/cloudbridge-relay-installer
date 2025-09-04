# CloudBridge Relay Server - Архитектурные диаграммы

## 1. Общая архитектура системы

```mermaid
graph TB
    subgraph "CloudBridge Relay Server"
        A[TCP Server :9090] --> B[Connection Handler]
        C[UDP Server :51820] --> D[WireGuard Handler]
        B --> E[Auth Adapter]
        E --> F[Keycloak Client]
        E --> G[Django Client]
        B --> H[Tunnel Manager]
        H --> I[Proxy Tunnels]
        D --> J[P2P Mesh]
        K[Enhanced QUIC] --> L[Tenant Manager]
        M[Zero Trust] --> N[Circuit Breaker]
        O[Metrics] --> P[Prometheus]
    end
    
    subgraph "External Services"
        Q[Keycloak Server]
        R[Django API]
        S[Prometheus]
        T[Grafana]
    end
    
    F --> Q
    G --> R
    P --> S
    S --> T
```

## 2. Сетевой стек протоколов

```mermaid
graph TB
    subgraph "Application Layer"
        A[CloudBridge Protocol]
        B[WireGuard Protocol]
    end
    
    subgraph "Transport Layer"
        C[TCP :9090]
        D[UDP :51820]
        E[QUIC :9090]
    end
    
    subgraph "Network Layer"
        F[IPv4/IPv6]
    end
    
    subgraph "Data Link Layer"
        G[Ethernet]
    end
    
    A --> C
    A --> E
    B --> D
    C --> F
    D --> F
    E --> F
    F --> G
```

## 3. Процесс аутентификации

```mermaid
sequenceDiagram
    participant C as Client
    participant R as Relay Server
    participant D as Django API
    participant K as Keycloak
    
    C->>R: hello message
    R->>C: hello response
    C->>R: auth with JWT token
    R->>D: validate token
    D->>R: validation result
    R->>C: auth success/failure
    C->>R: tunnel_info
    R->>C: tunnel created
```

## 4. WireGuard P2P Handshake

```mermaid
sequenceDiagram
    participant C as Client
    participant S as Server
    
    C->>S: Initiation (type=1)
    Note over C,S: [01][00][sender_index][00]
    S->>C: Response (type=2)
    Note over S,C: [02][00][sender_index][receiver_index][00][00]
    C->>S: Cookie Reply (type=3)
    S->>C: Transport Data (type=4)
```

## 5. Создание и управление туннелями

```mermaid
sequenceDiagram
    participant C as Client
    participant T as TCP Server
    participant A as Auth Adapter
    participant TM as Tunnel Manager
    participant P as Proxy
    
    C->>T: hello message
    T->>A: validate token
    A->>T: auth success
    T->>C: auth response
    C->>T: tunnel_info
    T->>TM: create tunnel
    TM->>P: start proxy
    P->>C: tunnel ready
    C->>P: data transfer
    P->>C: data response
```

## 6. Multi-tenancy архитектура

```mermaid
graph TB
    subgraph "Tenant A"
        A1[Client A1]
        A2[Client A2]
        A3[Client A3]
    end
    
    subgraph "Tenant B"
        B1[Client B1]
        B2[Client B2]
    end
    
    subgraph "Relay Server"
        C[Tenant Manager]
        D[Resource Limits]
        E[Isolation Layer]
    end
    
    A1 --> C
    A2 --> C
    A3 --> C
    B1 --> C
    B2 --> C
    
    C --> D
    C --> E
```

## 7. Мониторинг и метрики

```mermaid
graph TB
    subgraph "CloudBridge Relay"
        A[Metrics Collector]
        B[Health Checker]
        C[Log Aggregator]
    end
    
    subgraph "Monitoring Stack"
        D[Prometheus]
        E[Grafana]
        F[AlertManager]
    end
    
    A --> D
    B --> D
    C --> D
    D --> E
    D --> F
```

## 8. Безопасность и защита

```mermaid
graph TB
    subgraph "Security Layers"
        A[TLS 1.3]
        B[JWT Validation]
        C[Rate Limiting]
        D[DDoS Protection]
        E[Firewall Rules]
    end
    
    subgraph "External Threats"
        F[Malicious Clients]
        G[DDoS Attacks]
        H[Unauthorized Access]
    end
    
    F --> A
    G --> D
    H --> B
    
    A --> C
    B --> C
    C --> E
    D --> E
```

## 9. Развертывание в Kubernetes

```mermaid
graph TB
    subgraph "Kubernetes Cluster"
        subgraph "Namespace: cloudbridge"
            A[Deployment]
            B[Service]
            C[ConfigMap]
            D[Secret]
        end
        
        subgraph "Monitoring"
            E[Prometheus]
            F[Grafana]
        end
    end
    
    subgraph "External"
        G[Load Balancer]
        H[Ingress]
    end
    
    G --> H
    H --> B
    B --> A
    A --> C
    A --> D
    A --> E
    E --> F
```

## 10. Поток данных в системе

```mermaid
flowchart TD
    A[Client Request] --> B{Protocol Type}
    B -->|TCP| C[TCP Handler]
    B -->|UDP| D[UDP Handler]
    
    C --> E[Auth Check]
    D --> F[WireGuard Handshake]
    
    E --> G{Authenticated?}
    G -->|Yes| H[Tunnel Creation]
    G -->|No| I[Reject Connection]
    
    H --> J[Proxy Setup]
    J --> K[Data Transfer]
    
    F --> L[P2P Connection]
    L --> M[Mesh Network]
    
    K --> N[Monitoring]
    M --> N
    N --> O[Metrics Collection]
```

## 11. Обработка ошибок и восстановление

```mermaid
graph TB
    subgraph "Error Handling"
        A[Error Detection]
        B[Error Classification]
        C[Recovery Strategy]
        D[Fallback Mechanism]
    end
    
    subgraph "Error Types"
        E[Network Errors]
        F[Auth Errors]
        G[Resource Errors]
        H[System Errors]
    end
    
    A --> B
    B --> E
    B --> F
    B --> G
    B --> H
    
    E --> C
    F --> C
    G --> C
    H --> C
    
    C --> D
```

## 12. Масштабирование системы

```mermaid
graph TB
    subgraph "Load Balancer"
        A[HAProxy/Nginx]
    end
    
    subgraph "Relay Instances"
        B[Instance 1]
        C[Instance 2]
        D[Instance 3]
        E[Instance N]
    end
    
    subgraph "Shared Services"
        F[Redis Cache]
        G[Database]
        H[Message Queue]
    end
    
    A --> B
    A --> C
    A --> D
    A --> E
    
    B --> F
    C --> F
    D --> F
    E --> F
    
    B --> G
    C --> G
    D --> G
    E --> G
    
    B --> H
    C --> H
    D --> H
    E --> H
```

---

*Диаграммы созданы с помощью Mermaid.js*
*Для просмотра используйте Mermaid Live Editor или поддерживаемые плагины*
