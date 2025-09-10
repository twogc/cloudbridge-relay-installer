# Архитектурные схемы CloudBridge Relay

## 1. Общая архитектурная схема (Edge → Relay → Backend → Management)

```mermaid
graph TB
    subgraph "Edge Layer - Публичные точки входа"
        E1[Edge Server EU]
        E2[Edge Server US]
        E3[Edge Server AP]
        E4[Edge Server RU]
    end
    
    subgraph "Relay Layer - Защищённые узлы туннелирования"
        R1[Relay Node EU]
        R2[Relay Node US]
        R3[Relay Node AP]
        R4[Relay Node RU]
    end
    
    subgraph "Backend Layer - Корпоративные сервисы"
        B1[Corporate Servers]
        B2[Database Cluster]
        B3[Application Servers]
        B4[File Storage]
    end
    
    subgraph "Management Layer - Управление и мониторинг"
        M1[Cadence Workflow Engine]
        M2[AI Service]
        M3[Keycloak IAM]
        M4[Monitoring Stack]
    end
    
    subgraph "Trust Boundaries"
        TB1[Internet - Untrusted]
        TB2[Edge Network - Semi-trusted]
        TB3[Relay Network - Trusted]
        TB4[Backend Network - Highly Trusted]
        TB5[Management Network - Critical Trusted]
    end
    
    %% Edge to Relay connections
    E1 -.->|QUIC/TLS 1.3| R1
    E2 -.->|QUIC/TLS 1.3| R2
    E3 -.->|QUIC/TLS 1.3| R3
    E4 -.->|QUIC/TLS 1.3| R4
    
    %% Relay to Backend connections
    R1 -.->|WireGuard P2P| B1
    R2 -.->|WireGuard P2P| B2
    R3 -.->|WireGuard P2P| B3
    R4 -.->|WireGuard P2P| B4
    
    %% Management connections
    M1 -.->|mTLS| R1
    M1 -.->|mTLS| R2
    M1 -.->|mTLS| R3
    M1 -.->|mTLS| R4
    
    M2 -.->|mTLS| M1
    M3 -.->|OIDC| M1
    M4 -.->|Prometheus| M1
    
    %% Trust boundary indicators
    TB1 -.->|Crosses| E1
    TB2 -.->|Crosses| R1
    TB3 -.->|Crosses| B1
    TB4 -.->|Crosses| M1
```

## 2. Trust Boundaries и Security Zones

```mermaid
graph LR
    subgraph "Zone 0: Internet (Untrusted)"
        I[Internet Users]
        A[Attackers]
    end
    
    subgraph "Zone 1: Edge Network (Semi-trusted)"
        E[Edge Servers]
        LB[Load Balancers]
        FW1[Firewall Layer 1]
    end
    
    subgraph "Zone 2: Relay Network (Trusted)"
        R[Relay Servers]
        WG[WireGuard Mesh]
        FW2[Firewall Layer 2]
    end
    
    subgraph "Zone 3: Backend Network (Highly Trusted)"
        B[Backend Services]
        DB[Databases]
        FW3[Firewall Layer 3]
    end
    
    subgraph "Zone 4: Management Network (Critical Trusted)"
        M[Management Services]
        AI[AI Services]
        CAD[Cadence Engine]
        FW4[Firewall Layer 4]
    end
    
    I -->|Authenticated Access| E
    A -.->|Blocked| FW1
    E -->|Validated Traffic| R
    R -->|Encrypted P2P| B
    M -->|mTLS Control| R
    AI -->|ML Analysis| M
    CAD -->|Workflow Control| M
```

## 3. Data Flow Architecture

```mermaid
sequenceDiagram
    participant C as Client
    participant E as Edge Server
    participant R as Relay Server
    participant B as Backend Service
    participant M as Management Layer
    
    Note over C,M: Client Connection Flow
    
    C->>E: 1. QUIC Connection Request
    E->>E: 2. DDoS Protection Check
    E->>M: 3. Authentication Request
    M->>E: 4. JWT Validation
    E->>R: 5. Route to Optimal Relay
    R->>R: 6. WireGuard Handshake
    R->>B: 7. Secure Tunnel Established
    B->>R: 8. Service Response
    R->>E: 9. Encrypted Response
    E->>C: 10. Final Response
    
    Note over C,M: Management & Monitoring
    
    M->>R: 11. Health Check
    R->>M: 12. Status Report
    M->>M: 13. AI Analysis
    M->>R: 14. Optimization Commands
```

## 4. Multi-Tenant Architecture

```mermaid
graph TB
    subgraph "Tenant A - Enterprise Corp"
        TA1[Edge A1]
        TA2[Relay A1]
        TA3[Backend A1]
        TA4[Management A1]
    end
    
    subgraph "Tenant B - Startup Inc"
        TB1[Edge B1]
        TB2[Relay B1]
        TB3[Backend B1]
        TB4[Management B1]
    end
    
    subgraph "Shared Infrastructure"
        SI1[Shared Edge Pool]
        SI2[Shared Relay Pool]
        SI3[Monitoring Stack]
        SI4[AI Services]
    end
    
    subgraph "Isolation Mechanisms"
        IM1[Network Policies]
        IM2[Resource Quotas]
        IM3[Encryption Keys]
        IM4[Access Controls]
    end
    
    TA1 -->|Isolated| TA2
    TB1 -->|Isolated| TB2
    
    TA2 -.->|Shared| SI2
    TB2 -.->|Shared| SI2
    
    IM1 -->|Enforces| TA1
    IM1 -->|Enforces| TB1
    IM2 -->|Limits| TA2
    IM2 -->|Limits| TB2
    IM3 -->|Secures| TA3
    IM3 -->|Secures| TB3
    IM4 -->|Controls| TA4
    IM4 -->|Controls| TB4
```

## 5. Geographic Distribution

```mermaid
graph TB
    subgraph "Europe Region"
        EU1[Edge: London]
        EU2[Edge: Frankfurt]
        EU3[Relay: Amsterdam]
        EU4[Relay: Paris]
    end
    
    subgraph "North America Region"
        NA1[Edge: New York]
        NA2[Edge: Los Angeles]
        NA3[Relay: Chicago]
        NA4[Relay: Seattle]
    end
    
    subgraph "Asia Pacific Region"
        AP1[Edge: Tokyo]
        AP2[Edge: Singapore]
        AP3[Relay: Hong Kong]
        AP4[Relay: Sydney]
    end
    
    subgraph "Russia Region"
        RU1[Edge: Moscow]
        RU2[Edge: St. Petersburg]
        RU3[Relay: Yekaterinburg]
        RU4[Relay: Novosibirsk]
    end
    
    subgraph "Global Management"
        GM1[Cadence Cluster]
        GM2[AI Services]
        GM3[Monitoring]
        GM4[Keycloak]
    end
    
    %% Inter-region connections
    EU3 -.->|Cross-region| NA3
    NA3 -.->|Cross-region| AP3
    AP3 -.->|Cross-region| RU3
    RU3 -.->|Cross-region| EU3
    
    %% Management connections
    GM1 -.->|Global Control| EU3
    GM1 -.->|Global Control| NA3
    GM1 -.->|Global Control| AP3
    GM1 -.->|Global Control| RU3
    
    %% Regional connections
    EU1 --> EU3
    EU2 --> EU4
    NA1 --> NA3
    NA2 --> NA4
    AP1 --> AP3
    AP2 --> AP4
    RU1 --> RU3
    RU2 --> RU4
```

## 6. Security Architecture Layers

```mermaid
graph TB
    subgraph "Layer 7: Application Security"
        L7A[API Gateway]
        L7B[Rate Limiting]
        L7C[Input Validation]
        L7D[Authentication]
    end
    
    subgraph "Layer 6: Presentation Security"
        L6A[TLS 1.3]
        L6B[Certificate Management]
        L6C[Perfect Forward Secrecy]
        L6D[Post-Quantum Ready]
    end
    
    subgraph "Layer 5: Session Security"
        L5A[QUIC Sessions]
        L5B[Session Management]
        L5C[Token Validation]
        L5D[Multi-Factor Auth]
    end
    
    subgraph "Layer 4: Transport Security"
        L4A[WireGuard VPN]
        L4B[UDP/TCP Security]
        L4C[Port Security]
        L4D[Protocol Validation]
    end
    
    subgraph "Layer 3: Network Security"
        L3A[Network Policies]
        L3B[Firewall Rules]
        L3C[IP Filtering]
        L3D[Geographic Blocking]
    end
    
    subgraph "Layer 2: Data Link Security"
        L2A[VLAN Isolation]
        L2B[MAC Filtering]
        L2C[Switch Security]
        L2D[Link Encryption]
    end
    
    subgraph "Layer 1: Physical Security"
        L1A[Data Center Security]
        L1B[Hardware Security]
        L1C[Environmental Controls]
        L1D[Access Controls]
    end
    
    L7A --> L6A
    L6A --> L5A
    L5A --> L4A
    L4A --> L3A
    L3A --> L2A
    L2A --> L1A
```

---

*Эти архитектурные схемы демонстрируют многоуровневую архитектуру CloudBridge Relay с четкими trust boundaries, обеспечивающими безопасность и изоляцию на каждом уровне системы.*
