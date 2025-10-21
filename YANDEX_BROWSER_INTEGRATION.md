# CloudBridge Relay - Интеграция с Яндекс.Браузер

## Схема интеграции CloudBridge с Яндекс.Браузер

```mermaid
graph LR
    subgraph YB["Яндекс.Браузер (Chromium core)"]
        UA[User Agent<br/>Network Stack]
        H3[HTTP/3 Client<br/>QUIC Protocol]
        MASQ[MASQUE Client<br/>CONNECT-UDP / CONNECT-IP]
    end

    subgraph YS["Сервисы Яндекса"]
        DZEN[Dzen / Видео CDN<br/>Content Delivery]
        SRCH[Search / ya.ru<br/>Search Engine]
        MAIL[Mail / API<br/>Email Service]
    end

    subgraph POP["CloudBridge Relay PoP Cluster (Anycast)"]
        LB[Protocol Selector<br/>QUIC-first, fallback]
        RLY[Relay Core<br/>P2P/MASQUE Router]
        IPAM[Deterministic IPAM<br/>per-tenant]
        POL[Tenant Policy Controller<br/>Calico]
    end

    subgraph NT["NAT Traversal"]
        ICE[ICE Agent<br/>NAT Detection]
        STUN[STUN Server<br/>NAT Mapping]
        TURN[TURN Pool<br/>Relay Fallback]
    end

    subgraph AI["AI / SLO Loop"]
        METRICS[Telemetry Collector<br/>RTT/Jitter/Loss]
        PRED[Prediction & Anomaly<br/>AI Engine]
    end

    subgraph OBS["Наблюдаемость"]
        PM[Prometheus<br/>Metrics Collection]
        GF[Grafana<br/>Dashboards]
    end

    SEC[TLS 1.3 / JWT<br/>Zitadel Auth]

    %% Основные потоки данных
    UA -->|Initial request| H3
    H3 -->|alt-svc: h3| DZEN
    H3 -->|alt-svc: h3| SRCH
    H3 -->|alt-svc: h3| MAIL

    %% A/B эксперимент
    H3 -.->|A/B flag experiment| MASQ

    %% MASQUE/QUIC путь через Relay
    MASQ -->|HTTP/3 CONNECT-UDP| LB
    LB --> RLY
    RLY -->|Per-tenant IP assign| IPAM
    RLY --> POL

    %% NAT traversal
    RLY --> ICE
    ICE --> STUN
    ICE --> TURN

    %% Безопасность
    UA -.->|OIDC login| SEC
    SEC -.->|JWT validated| RLY

    %% Ускоренный путь к сервисам Яндекса
    RLY -->|Optimized QUIC path| DZEN
    RLY -->|Optimized QUIC path| SRCH
    RLY -->|Optimized QUIC path| MAIL

    %% Телеметрия и AI
    RLY --> METRICS
    METRICS --> PM
    PM --> GF
    METRICS --> PRED
    PRED -->|SLO targets, route hints| LB

    %% Fallback пути
    H3 -.->|if UDP/QUIC blocked| SRCH
    MASQ -.->|if MASQUE blocked| DZEN

    %% Стили
    classDef browser fill:#e3f2fd,stroke:#1e88e5,stroke-width:2px,color:#0d47a1
    classDef yandex fill:#fff3e0,stroke:#fb8c00,stroke-width:2px,color:#e65100
    classDef relay fill:#f3e5f5,stroke:#8e24aa,stroke-width:2px,color:#4a148c
    classDef nat fill:#ede7f6,stroke:#5e35b1,stroke-width:2px,color:#311b92
    classDef ai fill:#fff8e1,stroke:#fbc02d,stroke-width:2px,color:#795548
    classDef obs fill:#e8f5e9,stroke:#43a047,stroke-width:2px,color:#1b5e20
    classDef sec fill:#ffebee,stroke:#e53935,stroke-width:2px,color:#b71c1c

    class UA,H3,MASQ browser
    class DZEN,SRCH,MAIL yandex
    class LB,RLY,IPAM,POL relay
    class ICE,STUN,TURN nat
    class METRICS,PRED ai
    class PM,GF obs
    class SEC sec
```

## Описание интеграции

### Яндекс.Браузер (Chromium core)
- **User Agent**: Сетевой стек браузера
- **HTTP/3 Client**: QUIC протокол для быстрого соединения
- **MASQUE Client**: Поддержка CONNECT-UDP/CONNECT-IP для туннелирования

### Сервисы Яндекса
- **Dzen/Видео CDN**: Контент-доставка и видео
- **Search/ya.ru**: Поисковая система
- **Mail/API**: Почтовый сервис и API

### CloudBridge Relay PoP Cluster
- **Protocol Selector**: Выбор протокола (QUIC-first, fallback)
- **Relay Core**: P2P/MASQUE маршрутизатор
- **Deterministic IPAM**: Управление IP-адресами по арендаторам
- **Tenant Policy Controller**: Calico политики изоляции

### NAT Traversal
- **ICE Agent**: Детекция NAT
- **STUN Server**: Маппинг NAT
- **TURN Pool**: Relay fallback

### AI / SLO Loop
- **Telemetry Collector**: Сбор метрик RTT/Jitter/Loss
- **Prediction & Anomaly**: AI движок для предсказаний

### Наблюдаемость
- **Prometheus**: Сбор метрик
- **Grafana**: Дашборды и визуализация

## Поток данных

### 1. Инициализация
- Пользователь открывает Яндекс.Браузер
- Браузер отправляет HTTP/3 запросы к сервисам Яндекса
- Обнаруживается поддержка alt-svc: h3

### 2. A/B эксперимент
- Браузер получает флаг для участия в эксперименте
- Активируется MASQUE клиент
- Устанавливается соединение с CloudBridge Relay

### 3. Ускорение трафика
- MASQUE туннелирует трафик через CloudBridge Relay
- Relay применяет оптимизации маршрутизации
- Трафик доставляется к сервисам Яндекса по оптимизированному пути

### 4. NAT Traversal
- ICE агент определяет тип NAT
- STUN сервер помогает с маппингом
- TURN сервер используется как fallback

### 5. AI оптимизация
- Собираются метрики производительности
- AI движок анализирует паттерны
- Применяются оптимизации маршрутизации

## Преимущества интеграции

1. **Ускорение трафика** - Оптимизированные маршруты к сервисам Яндекса
2. **NAT Traversal** - Обход ограничений NAT
3. **AI оптимизация** - Умная маршрутизация на основе метрик
4. **Полная наблюдаемость** - Мониторинг производительности
5. **Безопасность** - TLS 1.3 и JWT аутентификация
6. **Multi-tenancy** - Изоляция по арендаторам
7. **Fallback механизмы** - Надежность соединения

CloudBridge Relay обеспечивает интеллектуальное ускорение трафика для Яндекс.Браузера с полной поддержкой современных протоколов и AI-оптимизацией.
