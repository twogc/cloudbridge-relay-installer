# 🏗️ Архитектура CloudBridge Relay + SIEM

## Диаграмма архитектуры

```mermaid
---
config:
  layout: dagre
  theme: redux
  look: neo
---
flowchart TD
 subgraph Control_Plane["Control Plane (Управляющий уровень)"]
        Keycloak["Keycloak (OIDC, LDAP, MFA)"]
        DjangoAPI["Django API (UI, REST, биллинг, управление)"]
        API["Open API Gateway (REST + Webhooks)"]
  end
 subgraph Data_Plane["Data Plane (Уровень трафика)"]
        DesktopClient["Desktop Client (Tauri/Svelte)"]
        RelayServer["Relay Server (Go)"]
        ServerClient["Server Client (Go)"]
        InternalService["Internal Services (LDAP, RDP, DB, 1C)"]
  end
 subgraph Observability["Observability (Мониторинг и логирование)"]
        Prometheus["Prometheus (метрики)"]
        Grafana["Grafana (дашборды)"]
        AlertManager["AlertManager (алерты)"]
        Lynis["Lynis (аудит безопасности)"]
  end
 subgraph Security_Intelligence["Security Intelligence (Безопасность и SIEM)"]
        SIEMCollector["SIEM Collector (сбор данных)"]
        SIEMProcessor["SIEM Processor (обработка событий)"]
        SIEMAlertManager["SIEM Alert Manager (уведомления)"]
        SIEMWebUI["SIEM Web UI (веб-интерфейс)"]
        Elasticsearch["Elasticsearch (хранение данных)"]
        Kibana["Kibana (анализ и визуализация)"]
        Kafka["Kafka (потоковая обработка)"]
        Redis["Redis (кэширование)"]
        ML["ML Engine (аномалии, риск, auto-response)"]
  end
 subgraph Edge_Infra["Edge & Scaling"]
        EdgeNodes["Edge Nodes (распределённая сеть узлов)"]
        HA["High Availability / Auto Scaling"]
  end

    %% Control Plane connections
    Keycloak --> DjangoAPI
    API --> DjangoAPI
    
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
    
    %% External integrations
    API --> DjangoAPI
    API --> SIEMWebUI
    API --> SIEMAlertManager
    
    %% Edge and scaling
    RelayServer --> EdgeNodes
    EdgeNodes --> HA
    
    %% Styling
    style Keycloak fill:#00C853
    style DjangoAPI fill:#00C853
    style API fill:#C8E6C9
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

## 📊 Компоненты SIEM системы

### 🔍 **SIEM Collector**
- Сбор данных из CloudBridge Relay
- Сбор событий аутентификации из Keycloak
- Сбор системных логов
- Отправка данных в Kafka

### ⚙️ **SIEM Processor**
- Обработка событий безопасности
- Детекция угроз (brute force, подозрительные IP)
- Обогащение данных (геолокация, информация об IP)
- Сохранение в Elasticsearch

### 🚨 **SIEM Alert Manager**
- Генерация уведомлений
- Поддержка Telegram, Email, Slack
- Дедупликация и фильтрация
- Эскалация инцидентов

### 🌐 **SIEM Web UI**
- Веб-интерфейс для управления
- Дашборды и статистика
- API для интеграций
- Мониторинг здоровья системы

### 💾 **Хранилище данных**
- **Elasticsearch** - основное хранилище событий
- **Kibana** - анализ и визуализация
- **Kafka** - потоковая обработка
- **Redis** - кэширование и сессии

### 🤖 **ML Engine**
- Анализ аномалий
- Оценка рисков
- Автоматические ответы
- Поведенческий анализ

## 🔗 Интеграции

### **Внутренние интеграции**
- CloudBridge Relay - логи туннелей и аутентификации
- Keycloak - события аутентификации и авторизации
- Django API - аудит и метрики
- Prometheus - метрики производительности

### **Внешние интеграции**
- Telegram - мгновенные уведомления
- Email - детальные отчеты
- Slack - командные уведомления
- Webhook - интеграция с внешними системами

## 🛡️ Безопасность

### **Соответствие требованиям**
- ✅ 152-ФЗ "О персональных данных"
- ✅ 187-ФЗ "О безопасности критической информационной инфраструктуры"
- ✅ ГОСТ Р 57580.1-2017
- ✅ ГОСТ Р 57580.2-2017

### **Функции безопасности**
- Шифрование данных в покое и в движении
- Аудит всех действий
- Многофакторная аутентификация
- Ротация ключей и сертификатов

## 📈 Масштабируемость

### **Горизонтальное масштабирование**
- Множественные экземпляры коллекторов
- Кластеризация Elasticsearch
- Партиционирование Kafka
- Балансировка нагрузки

### **Высокая доступность**
- Репликация данных
- Автоматическое восстановление
- Мониторинг здоровья
- Резервное копирование

## 🚀 Развертывание

### **Docker Compose**
```bash
cd siem
./start_siem.sh
```

### **Kubernetes**
```bash
kubectl apply -f k8s/
```

### **Мониторинг**
- Prometheus метрики
- Grafana дашборды
- Логирование в Elasticsearch
- Алерты через AlertManager 