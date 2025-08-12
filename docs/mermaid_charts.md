# 📊 Mermaid Charts для CloudBridge Relay

## 🏗️ Архитектура системы

### Основная архитектура

```mermaid
---
config:
  layout: dagre
  theme: redux
  look: neo
---
flowchart TD
    subgraph "Control Plane (Управляющий уровень)"
        Keycloak["Keycloak (OIDC, LDAP, MFA)"]
        DjangoAPI["Django API (UI, REST, биллинг)"]
        API["Open API Gateway"]
        AdminPanel["Admin Panel (React)"]
    end
    
    subgraph "Data Plane (Уровень трафика)"
        DesktopClient["Desktop Client (Tauri/Svelte)"]
        RelayServer["Relay Server (Go)"]
        ServerClient["Server Client (Go)"]
        InternalService["Internal Services (LDAP, RDP, DB, 1C)"]
    end
    
    subgraph "Observability (Мониторинг)"
        Prometheus["Prometheus (метрики)"]
        Grafana["Grafana (дашборды)"]
        AlertManager["AlertManager (алерты)"]
        Lynis["Lynis (аудит безопасности)"]
    end
    
    subgraph "Security Intelligence (SIEM)"
        SIEMCollector["SIEM Collector"]
        SIEMProcessor["SIEM Processor"]
        SIEMAlertManager["SIEM Alert Manager"]
        SIEMWebUI["SIEM Web UI"]
        Elasticsearch["Elasticsearch"]
        Kibana["Kibana"]
        Kafka["Kafka"]
        Redis["Redis"]
        ML["ML Engine"]
    end
    
    subgraph "Edge Infrastructure"
        EdgeNodes["Edge Nodes (распределённая сеть)"]
        HA["High Availability / Auto Scaling"]
    end

    %% Control Plane connections
    Keycloak --> DjangoAPI
    API --> DjangoAPI
    DjangoAPI --> AdminPanel
    
    %% Data Plane connections
    DesktopClient -- OIDC Token --> RelayServer
    ServerClient -- JWT Token --> RelayServer
    DjangoAPI -- Token Provisioning --> ServerClient
    RelayServer -- Tunnel Creation --> ServerClient
    DesktopClient -- Tunnel Access --> RelayServer
    RelayServer -- Forwarded TCP --> InternalService
    
    %% Observability connections
    RelayServer -- Metrics --> Prometheus
    DjangoAPI -- Metrics --> Prometheus
    Prometheus --> Grafana
    Prometheus --> AlertManager
    Lynis --> Prometheus
    
    %% SIEM connections
    RelayServer -- Logs / Access --> SIEMCollector
    DjangoAPI -- Audit Logs --> SIEMCollector
    Keycloak -- Auth Events --> SIEMCollector
    InternalService -- System Logs --> SIEMCollector
    
    %% SIEM internal flow
    SIEMCollector --> Kafka
    Kafka --> SIEMProcessor
    SIEMProcessor --> Elasticsearch
    SIEMProcessor --> Redis
    Elasticsearch --> Kibana
    Elasticsearch --> SIEMWebUI
    SIEMProcessor --> SIEMAlertManager
    SIEMAlertManager --> AlertManager
    
    %% ML integration
    SIEMProcessor --> ML
    ML -- Alerts / Risk Score --> SIEMAlertManager
    ML -- Anomaly Detection --> SIEMProcessor
    
    %% Edge and scaling
    RelayServer --> EdgeNodes
    EdgeNodes --> HA
    
    %% Styling
    style Keycloak fill:#00C853
    style DjangoAPI fill:#00C853
    style API fill:#C8E6C9
    style AdminPanel fill:#00C853
    style DesktopClient fill:#00C853
    style RelayServer fill:#00C853
    style SIEMCollector fill:#FF9800
    style SIEMProcessor fill:#FF9800
    style SIEMAlertManager fill:#FF9800
    style SIEMWebUI fill:#FF9800
    style Elasticsearch fill:#2196F3
    style Kibana fill:#2196F3
    style Kafka fill:#9C27B0
    style Redis fill:#F44336
    style ML fill:#E91E63
```

## 🌐 Глобальная сеть узлов

### Географическое распределение

```mermaid
---
config:
  layout: dagre
  theme: redux
  look: neo
---
flowchart TD
    subgraph "Global Edge Network"
        subgraph "Europe"
            Moscow["Moscow Edge Node<br/>100 Gbps"]
            SPb["St. Petersburg Edge Node<br/>50 Gbps"]
            London["London Edge Node<br/>100 Gbps"]
            Frankfurt["Frankfurt Edge Node<br/>50 Gbps"]
        end
        
        subgraph "Americas"
            NewYork["New York Edge Node<br/>100 Gbps"]
            Chicago["Chicago Edge Node<br/>50 Gbps"]
            SaoPaulo["São Paulo Edge Node<br/>50 Gbps"]
        end
        
        subgraph "Asia Pacific"
            Singapore["Singapore Edge Node<br/>100 Gbps"]
            Tokyo["Tokyo Edge Node<br/>50 Gbps"]
            Sydney["Sydney Edge Node<br/>50 Gbps"]
        end
    end
    
    subgraph "Regional Relay Networks"
        subgraph "Russia"
            Office1["Office 1 Relay<br/>10 Gbps"]
            Office2["Office 2 Relay<br/>10 Gbps"]
            Branch1["Branch 1 Relay<br/>5 Gbps"]
            Branch2["Branch 2 Relay<br/>5 Gbps"]
        end
        
        subgraph "Europe"
            EUOffice1["EU Office 1 Relay<br/>10 Gbps"]
            EUOffice2["EU Office 2 Relay<br/>10 Gbps"]
        end
        
        subgraph "USA"
            USOffice1["US Office 1 Relay<br/>10 Gbps"]
            USOffice2["US Office 2 Relay<br/>10 Gbps"]
        end
    end
    
    subgraph "Client Devices"
        Desktop["Desktop Client"]
        Mobile["Mobile Client"]
        Server["Server Client"]
        IoT["IoT Device"]
    end
    
    %% Edge to Relay connections
    Moscow --> Office1
    Moscow --> Office2
    SPb --> Branch1
    SPb --> Branch2
    
    London --> EUOffice1
    Frankfurt --> EUOffice2
    
    NewYork --> USOffice1
    Chicago --> USOffice2
    
    %% Relay to Client connections
    Office1 --> Desktop
    Office2 --> Server
    Branch1 --> Mobile
    Branch2 --> IoT
    
    %% Styling
    style Moscow fill:#4CAF50
    style SPb fill:#4CAF50
    style London fill:#4CAF50
    style Frankfurt fill:#4CAF50
    style NewYork fill:#4CAF50
    style Chicago fill:#4CAF50
    style SaoPaulo fill:#4CAF50
    style Singapore fill:#4CAF50
    style Tokyo fill:#4CAF50
    style Sydney fill:#4CAF50
    
    style Office1 fill:#2196F3
    style Office2 fill:#2196F3
    style Branch1 fill:#2196F3
    style Branch2 fill:#2196F3
    style EUOffice1 fill:#2196F3
    style EUOffice2 fill:#2196F3
    style USOffice1 fill:#2196F3
    style USOffice2 fill:#2196F3
    
    style Desktop fill:#FF9800
    style Mobile fill:#FF9800
    style Server fill:#FF9800
    style IoT fill:#FF9800
```

## 🔄 Сценарии использования

### 1. Подключение клиента к корпоративным ресурсам

```mermaid
sequenceDiagram
    participant Client as Desktop Client
    participant Relay as Relay Server
    participant Keycloak as Keycloak
    participant Django as Django API
    participant Internal as Internal Service
    
    Client->>Keycloak: 1. Аутентификация (OIDC)
    Keycloak-->>Client: 2. Access Token
    
    Client->>Relay: 3. Подключение с токеном
    Relay->>Keycloak: 4. Валидация токена
    Keycloak-->>Relay: 5. Подтверждение
    
    Relay->>Django: 6. Проверка прав доступа
    Django-->>Relay: 7. Разрешение доступа
    
    Relay-->>Client: 8. Подтверждение подключения
    
    Client->>Relay: 9. TCP туннель
    Relay->>Internal: 10. Проксирование трафика
    Internal-->>Relay: 11. Ответ
    Relay-->>Client: 12. Данные
```

### 2. Автоматическое масштабирование

```mermaid
flowchart TD
    Start([Мониторинг нагрузки]) --> Check{Нагрузка > 80%?}
    Check -->|Да| ScaleUp[Создание нового Relay узла]
    Check -->|Нет| Continue[Продолжение мониторинга]
    
    ScaleUp --> Deploy[Развертывание контейнера]
    Deploy --> Register[Регистрация в DNS]
    Register --> LoadBalancer[Обновление балансировщика]
    LoadBalancer --> HealthCheck{Проверка здоровья}
    
    HealthCheck -->|Успех| Traffic[Направление трафика]
    HealthCheck -->|Ошибка| Retry[Повторная попытка]
    Retry --> Deploy
    
    Traffic --> Monitor[Мониторинг нового узла]
    Monitor --> Continue
    
    Continue --> Start
```

### 3. Детекция и блокировка угроз

```mermaid
sequenceDiagram
    participant Client as Подозрительный клиент
    participant Relay as Relay Server
    participant SIEM as SIEM Processor
    participant ML as ML Engine
    participant Alert as Alert Manager
    participant Admin as Admin Panel
    
    Client->>Relay: 1. Попытка подключения
    Relay->>SIEM: 2. Логирование события
    SIEM->>ML: 3. Анализ поведения
    ML-->>SIEM: 4. Оценка риска: ВЫСОКИЙ
    
    SIEM->>Alert: 5. Генерация алерта
    Alert->>Admin: 6. Уведомление администратора
    Alert->>Relay: 7. Команда блокировки
    
    Relay-->>Client: 8. Отклонение подключения
    Relay->>SIEM: 9. Логирование блокировки
```

### 4. Географическая маршрутизация

```mermaid
flowchart TD
    Client[Клиент в Москве] --> GeoCheck{Определение геолокации}
    GeoCheck --> MoscowNode[Московский Edge Node]
    GeoCheck --> SPbNode[СПб Edge Node]
    
    MoscowNode --> LoadCheck1{Нагрузка < 70%?}
    SPbNode --> LoadCheck2{Нагрузка < 70%?}
    
    LoadCheck1 -->|Да| Route1[Маршрутизация через Москву]
    LoadCheck1 -->|Нет| SPbNode
    
    LoadCheck2 -->|Да| Route2[Маршрутизация через СПб]
    LoadCheck2 -->|Нет| Fallback[Резервный узел]
    
    Route1 --> Target[Целевой сервис]
    Route2 --> Target
    Fallback --> Target
```

## 📊 Мониторинг и метрики

### Система метрик

```mermaid
graph LR
    subgraph "Data Sources"
        Relay[Relay Server]
        Keycloak[Keycloak]
        Django[Django API]
        SIEM[SIEM System]
    end
    
    subgraph "Collection"
        Prometheus[Prometheus]
        Collectors[Custom Collectors]
    end
    
    subgraph "Processing"
        AlertManager[Alert Manager]
        Rules[Alert Rules]
    end
    
    subgraph "Visualization"
        Grafana[Grafana]
        Dashboards[Dashboards]
    end
    
    subgraph "Storage"
        TSDB[Time Series DB]
        Logs[Log Storage]
    end
    
    Relay --> Prometheus
    Keycloak --> Prometheus
    Django --> Prometheus
    SIEM --> Collectors
    
    Collectors --> Prometheus
    Prometheus --> TSDB
    Prometheus --> AlertManager
    
    AlertManager --> Rules
    AlertManager --> Grafana
    
    TSDB --> Grafana
    Grafana --> Dashboards
```

### Дашборд метрик

```mermaid
graph TD
    subgraph "Real-time Metrics"
        Connections[Активные подключения]
        Throughput[Пропускная способность]
        Latency[Задержка]
        Errors[Ошибки]
    end
    
    subgraph "Business Metrics"
        Users[Активные пользователи]
        Tunnels[Активные туннели]
        Regions[Активность по регионам]
        Revenue[Доход]
    end
    
    subgraph "Security Metrics"
        Threats[Обнаруженные угрозы]
        Blocks[Блокировки]
        AuthFailures[Ошибки аутентификации]
        Compliance[Соответствие требованиям]
    end
    
    subgraph "Infrastructure Metrics"
        CPU[Использование CPU]
        Memory[Использование памяти]
        Disk[Использование диска]
        Network[Сетевая активность]
    end
    
    Connections --> Dashboard[Главный дашборд]
    Throughput --> Dashboard
    Latency --> Dashboard
    Errors --> Dashboard
    
    Users --> Dashboard
    Tunnels --> Dashboard
    Regions --> Dashboard
    Revenue --> Dashboard
    
    Threats --> Dashboard
    Blocks --> Dashboard
    AuthFailures --> Dashboard
    Compliance --> Dashboard
    
    CPU --> Dashboard
    Memory --> Dashboard
    Disk --> Dashboard
    Network --> Dashboard
```

## 🤖 AI/ML интеграция

### Машинное обучение в сети

```mermaid
flowchart TD
    subgraph "Data Collection"
        Logs[Логи системы]
        Metrics[Метрики производительности]
        Events[События безопасности]
        UserBehavior[Поведение пользователей]
    end
    
    subgraph "Data Processing"
        Preprocessing[Предобработка данных]
        FeatureExtraction[Извлечение признаков]
        Normalization[Нормализация]
    end
    
    subgraph "ML Models"
        AnomalyDetection[Детекция аномалий]
        TrafficPrediction[Прогнозирование трафика]
        ThreatDetection[Детекция угроз]
        RouteOptimization[Оптимизация маршрутов]
    end
    
    subgraph "Actions"
        AutoScaling[Автоматическое масштабирование]
        ThreatBlocking[Блокировка угроз]
        RouteAdjustment[Корректировка маршрутов]
        ResourceOptimization[Оптимизация ресурсов]
    end
    
    Logs --> Preprocessing
    Metrics --> Preprocessing
    Events --> Preprocessing
    UserBehavior --> Preprocessing
    
    Preprocessing --> FeatureExtraction
    FeatureExtraction --> Normalization
    
    Normalization --> AnomalyDetection
    Normalization --> TrafficPrediction
    Normalization --> ThreatDetection
    Normalization --> RouteOptimization
    
    AnomalyDetection --> AutoScaling
    TrafficPrediction --> AutoScaling
    ThreatDetection --> ThreatBlocking
    RouteOptimization --> RouteAdjustment
    
    AutoScaling --> ResourceOptimization
    ThreatBlocking --> ResourceOptimization
    RouteAdjustment --> ResourceOptimization
```

## 🔐 Безопасность

### Zero Trust Architecture

```mermaid
flowchart TD
    subgraph "Identity Verification"
        User[Пользователь]
        Device[Устройство]
        Location[Местоположение]
        Time[Время доступа]
    end
    
    subgraph "Continuous Monitoring"
        Behavior[Поведение]
        Network[Сетевая активность]
        Application[Активность приложений]
        Data[Доступ к данным]
    end
    
    subgraph "Risk Assessment"
        RiskScore[Оценка риска]
        ThreatIntel[Угрозы]
        Compliance[Соответствие]
        Policy[Политики]
    end
    
    subgraph "Access Control"
        Allow[Разрешить доступ]
        Deny[Отклонить доступ]
        Challenge[Дополнительная проверка]
        Limit[Ограничить доступ]
    end
    
    User --> Behavior
    Device --> Network
    Location --> Application
    Time --> Data
    
    Behavior --> RiskScore
    Network --> ThreatIntel
    Application --> Compliance
    Data --> Policy
    
    RiskScore --> Allow
    ThreatIntel --> Deny
    Compliance --> Challenge
    Policy --> Limit
```

## 🌍 Глобальная экосистема

### Партнерская сеть

```mermaid
graph TD
    subgraph "CloudBridge Core"
        Core[Центральная платформа]
        API[API Gateway]
        Admin[Admin Panel]
    end
    
    subgraph "Partners"
        Resellers[Реселлеры]
        Integrators[Системные интеграторы]
        Consultants[Консультанты]
        Developers[Разработчики]
    end
    
    subgraph "Services"
        WhiteLabel[White-label решения]
        CustomDev[Кастомная разработка]
        Support[Техническая поддержка]
        Training[Обучение]
    end
    
    subgraph "Revenue"
        RevenueSharing[Разделение доходов]
        Licensing[Лицензирование]
        Subscriptions[Подписки]
        Services[Услуги]
    end
    
    Core --> API
    API --> Resellers
    API --> Integrators
    API --> Consultants
    API --> Developers
    
    Resellers --> WhiteLabel
    Integrators --> CustomDev
    Consultants --> Support
    Developers --> Training
    
    WhiteLabel --> RevenueSharing
    CustomDev --> Licensing
    Support --> Subscriptions
    Training --> Services
```

---

## 📝 Использование диаграмм

Эти Mermaid диаграммы можно использовать для:

1. **Документации** - визуализация архитектуры и процессов
2. **Презентаций** - демонстрация возможностей системы
3. **Планирования** - обсуждение будущих функций
4. **Обучения** - объяснение работы системы новым сотрудникам
5. **Маркетинга** - показ преимуществ и возможностей

### Интеграция с документацией

Диаграммы можно встраивать в:
- README.md
- Техническую документацию
- Пользовательские руководства
- Маркетинговые материалы
- Презентации для инвесторов

### Автоматическое обновление

Рекомендуется настроить автоматическое обновление диаграмм при изменении кода или конфигурации системы. 