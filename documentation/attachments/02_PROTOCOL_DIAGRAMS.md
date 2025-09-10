# Протокольные диаграммы CloudBridge Relay

## 1. QUIC/HTTP-3 Protocol Stack

```mermaid
graph TB
    subgraph "Application Layer"
        APP1[CloudBridge Protocol]
        APP2[HTTP/3]
        APP3[WebSocket over QUIC]
    end
    
    subgraph "QUIC Layer"
        QUIC1[QUIC Connection]
        QUIC2[QUIC Streams]
        QUIC3[QUIC Congestion Control]
        QUIC4[QUIC Error Recovery]
    end
    
    subgraph "Security Layer"
        SEC1[TLS 1.3 Handshake]
        SEC2[0-RTT Connection]
        SEC3[Perfect Forward Secrecy]
        SEC4[Post-Quantum Hybrid]
    end
    
    subgraph "Transport Layer"
        TRANS1[UDP]
        TRANS2[Connection Migration]
        TRANS3[Multiplexing]
    end
    
    subgraph "Network Layer"
        NET1[IPv4/IPv6]
        NET2[Anycast Routing]
    end
    
    APP1 --> QUIC1
    APP2 --> QUIC2
    APP3 --> QUIC3
    
    QUIC1 --> SEC1
    QUIC2 --> SEC2
    QUIC3 --> SEC3
    QUIC4 --> SEC4
    
    SEC1 --> TRANS1
    SEC2 --> TRANS2
    SEC3 --> TRANS3
    
    TRANS1 --> NET1
    TRANS2 --> NET2
```

## 2. QUIC Connection Establishment

```mermaid
sequenceDiagram
    participant C as Client
    participant S as Server
    participant K as Keycloak
    
    Note over C,K: QUIC Connection with 0-RTT
    
    C->>S: 1. QUIC Initial Packet (Client Hello)
    S->>C: 2. QUIC Initial Packet (Server Hello)
    S->>C: 3. QUIC Handshake Packet (Certificate)
    C->>S: 4. QUIC Handshake Packet (Finished)
    S->>C: 5. QUIC Handshake Packet (Finished)
    
    Note over C,S: Connection Established
    
    C->>S: 6. 0-RTT Application Data
    S->>C: 7. 0-RTT Response
    
    Note over C,K: Authentication Flow
    
    C->>K: 8. OIDC Authentication Request
    K->>C: 9. JWT Token
    C->>S: 10. Authenticated Request with JWT
    S->>C: 11. Authorized Response
```

## 3. WireGuard P2P Handshake

```mermaid
sequenceDiagram
    participant A as Peer A
    participant B as Peer B
    participant R as Relay Server
    
    Note over A,R: WireGuard P2P Connection Setup
    
    A->>R: 1. Register Public Key
    B->>R: 2. Register Public Key
    R->>A: 3. Peer B's Endpoint Info
    R->>B: 4. Peer A's Endpoint Info
    
    Note over A,B: Direct P2P Handshake
    
    A->>B: 5. Handshake Initiation (HI)
    Note right of A: Ephemeral Key + Timestamp
    B->>A: 6. Handshake Response (HR)
    Note right of B: Ephemeral Key + Timestamp
    A->>B: 7. Handshake Confirmation (HC)
    Note right of A: Encrypted with shared secret
    
    Note over A,B: Data Transfer
    
    A->>B: 8. Encrypted Data Packets
    B->>A: 9. Encrypted Data Packets
    
    Note over A,B: Keep-alive
    
    A->>B: 10. Keep-alive (every 25s)
    B->>A: 11. Keep-alive Response
```

## 4. Hybrid Post-Quantum Cryptography Handshake

```mermaid
sequenceDiagram
    participant C as Client
    participant S as Server
    participant PQ as Post-Quantum Module
    
    Note over C,PQ: Hybrid PQC Handshake
    
    C->>S: 1. Client Hello (Classical + PQ Key Share)
    Note right of C: ECDHE + CRYSTALS-Kyber
    
    S->>PQ: 2. Generate PQ Key Pair
    PQ->>S: 3. PQ Public Key
    
    S->>C: 4. Server Hello (Classical + PQ Key Share)
    Note right of S: ECDHE + CRYSTALS-Kyber
    
    C->>PQ: 5. Generate Shared Secret
    PQ->>C: 6. PQ Shared Secret
    
    S->>PQ: 7. Generate Shared Secret
    PQ->>S: 8. PQ Shared Secret
    
    Note over C,S: Hybrid Key Derivation
    
    C->>C: 9. Derive Hybrid Key (Classical + PQ)
    S->>S: 10. Derive Hybrid Key (Classical + PQ)
    
    C->>S: 11. Finished (Hybrid Signature)
    Note right of C: CRYSTALS-Dilithium
    
    S->>C: 12. Finished (Hybrid Signature)
    Note right of S: CRYSTALS-Dilithium
    
    Note over C,S: Secure Channel Established
```

## 5. Enhanced QUIC with Multi-Tenancy

```mermaid
graph TB
    subgraph "Tenant A"
        TA1[QUIC Connection A]
        TA2[Stream Pool A]
        TA3[Congestion Control A]
        TA4[Rate Limiter A]
    end
    
    subgraph "Tenant B"
        TB1[QUIC Connection B]
        TB2[Stream Pool B]
        TB3[Congestion Control B]
        TB4[Rate Limiter B]
    end
    
    subgraph "Shared QUIC Engine"
        SE1[Connection Manager]
        SE2[Stream Multiplexer]
        SE3[Global Congestion Control]
        SE4[Resource Monitor]
    end
    
    subgraph "Security Layer"
        SL1[Tenant Isolation]
        SL2[Encryption Keys]
        SL3[Access Control]
        SL4[Audit Logging]
    end
    
    TA1 --> SE1
    TB1 --> SE1
    TA2 --> SE2
    TB2 --> SE2
    TA3 --> SE3
    TB3 --> SE3
    TA4 --> SE4
    TB4 --> SE4
    
    SL1 --> TA1
    SL1 --> TB1
    SL2 --> TA2
    SL2 --> TB2
    SL3 --> TA3
    SL3 --> TB3
    SL4 --> TA4
    SL4 --> TB4
```

## 6. Protocol Performance Comparison

```mermaid
graph LR
    subgraph "Latency Comparison"
        L1[TCP/TLS: 3-5 RTT]
        L2[QUIC: 1-2 RTT]
        L3[QUIC 0-RTT: 0 RTT]
        L4[WireGuard: 1 RTT]
    end
    
    subgraph "Throughput Comparison"
        T1[TCP: 1 Gbps]
        T2[QUIC: 2-3 Gbps]
        T3[WireGuard: 1.5 Gbps]
        T4[Hybrid PQC: 1.2 Gbps]
    end
    
    subgraph "Security Features"
        S1[TLS 1.3: Perfect Forward Secrecy]
        S2[QUIC: Built-in Encryption]
        S3[WireGuard: Modern Crypto]
        S4[PQC: Quantum-Resistant]
    end
    
    subgraph "Connection Overhead"
        O1[TCP: High]
        O2[QUIC: Medium]
        O3[WireGuard: Low]
        O4[PQC: Medium-High]
    end
```

## 7. Multi-Protocol Architecture

```mermaid
graph TB
    subgraph "Client Applications"
        CA1[Web Browser]
        CA2[Mobile App]
        CA3[Desktop Client]
        CA4[IoT Device]
    end
    
    subgraph "Protocol Selection Layer"
        PS1[Protocol Detector]
        PS2[Capability Negotiation]
        PS3[Fallback Manager]
        PS4[Performance Monitor]
    end
    
    subgraph "Protocol Implementations"
        PI1[QUIC/HTTP-3]
        PI2[WireGuard P2P]
        PI3[TCP/TLS 1.3]
        PI4[UDP/DTLS]
    end
    
    subgraph "Backend Services"
        BS1[Web Services]
        BS2[File Transfer]
        BS3[Real-time Apps]
        BS4[IoT Backend]
    end
    
    CA1 --> PS1
    CA2 --> PS2
    CA3 --> PS3
    CA4 --> PS4
    
    PS1 --> PI1
    PS2 --> PI2
    PS3 --> PI3
    PS4 --> PI4
    
    PI1 --> BS1
    PI2 --> BS2
    PI3 --> BS3
    PI4 --> BS4
```

## 8. Protocol Security Matrix

```mermaid
graph TB
    subgraph "Encryption Strength"
        ES1[AES-256-GCM: 256-bit]
        ES2[ChaCha20-Poly1305: 256-bit]
        ES3[CRYSTALS-Kyber: 128-bit PQ]
        ES4[Hybrid: 256-bit + 128-bit PQ]
    end
    
    subgraph "Authentication"
        AUTH1[HMAC-SHA256]
        AUTH2[Poly1305]
        AUTH3[CRYSTALS-Dilithium]
        AUTH4[Hybrid Signatures]
    end
    
    subgraph "Key Exchange"
        KE1[ECDHE-P256]
        KE2[ECDHE-P384]
        KE3[CRYSTALS-Kyber-768]
        KE4[Hybrid ECDHE + Kyber]
    end
    
    subgraph "Perfect Forward Secrecy"
        PFS1[TLS 1.3: Yes]
        PFS2[QUIC: Yes]
        PFS3[WireGuard: Yes]
        PFS4[PQC: Yes]
    end
```

## 9. Network Protocol Stack

```mermaid
graph TB
    subgraph "Application Protocols"
        AP1[CloudBridge Protocol]
        AP2[HTTP/3]
        AP3[WebSocket]
        AP4[Custom Protocols]
    end
    
    subgraph "Transport Protocols"
        TP1[QUIC]
        TP2[TCP]
        TP3[UDP]
        TP4[SCTP]
    end
    
    subgraph "Security Protocols"
        SP1[TLS 1.3]
        SP2[WireGuard]
        SP3[DTLS 1.3]
        SP4[Post-Quantum Crypto]
    end
    
    subgraph "Network Protocols"
        NP1[IPv4]
        NP2[IPv6]
        NP3[Anycast]
        NP4[Multicast]
    end
    
    subgraph "Data Link Protocols"
        DLP1[Ethernet]
        DLP2[WiFi]
        DLP3[4G/5G]
        DLP4[Satellite]
    end
    
    AP1 --> TP1
    AP2 --> TP1
    AP3 --> TP2
    AP4 --> TP3
    
    TP1 --> SP1
    TP2 --> SP1
    TP3 --> SP2
    TP4 --> SP3
    
    SP1 --> NP1
    SP2 --> NP2
    SP3 --> NP3
    SP4 --> NP4
    
    NP1 --> DLP1
    NP2 --> DLP2
    NP3 --> DLP3
    NP4 --> DLP4
```

---

*Эти протокольные диаграммы демонстрируют современные сетевые протоколы, используемые в CloudBridge Relay, включая QUIC/HTTP-3, WireGuard P2P, и гибридную пост-квантовую криптографию для обеспечения максимальной производительности и безопасности.*
