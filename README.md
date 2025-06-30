# CloudBridge Relay v2.0

CloudBridge Relay v2.0 — высокопроизводительный сервер туннелирования с поддержкой TCP proxy и multi-tenancy, обеспечивающий защищённый доступ к внутренним ресурсам с централизованной аутентификацией через Keycloak. Это альтернатива Cloudflare Argo Tunnel с полным контролем над инфраструктурой и данными.

## 🚀 Быстрый запуск

### Автоматическая установка (рекомендуется)

```bash
# Клонирование репозитория
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer

# Генерация production секретов
./scripts/generate_secrets.sh --env .env.production

# Запуск всех сервисов в production режиме
./scripts/start_production.sh --env .env.production

# Проверка статуса
docker-compose ps
```

### Установка как системный сервис (автозапуск)

```bash
# Установка systemd сервиса
sudo ./scripts/install_systemd.sh install

# Проверка статуса
sudo systemctl status cloudbridge-relay

# Управление сервисом
sudo systemctl start cloudbridge-relay
sudo systemctl stop cloudbridge-relay
sudo systemctl restart cloudbridge-relay
```

### Доступ к сервисам

После запуска будут доступны:
- **Keycloak**: http://localhost:8080 (admin/admin)
- **Relay**: http://localhost:8081/metrics
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **AlertManager**: http://localhost:9093

## 🖥️ Развертывание SIEM на отдельном сервере

**⚠️ ВАЖНО: SIEM систему необходимо развертывать на отдельном сервере** для обеспечения оптимальной производительности, безопасности и масштабируемости.

### Почему нужен отдельный сервер?

- **Ресурсные требования**: SIEM требует 16-64GB RAM, интенсивные I/O операции
- **Безопасность**: Изоляция security инфраструктуры от production систем
- **Производительность**: Отсутствие конкуренции ресурсов с основными сервисами
- **Масштабируемость**: Независимое масштабирование SIEM и Relay систем
- **Надежность**: Отказоустойчивость и независимое восстановление

### Быстрое развертывание SIEM

```bash
# Скачивание скрипта развертывания
wget https://raw.githubusercontent.com/maxlanies/cloudbridge-relay-installer/main/scripts/deploy_siem_server.sh

# Установка прав на выполнение
chmod +x deploy_siem_server.sh

# Запуск полного развертывания
sudo ./deploy_siem_server.sh --full

# Или минимальное развертывание
sudo ./deploy_siem_server.sh --minimal
```

### Требования к серверу SIEM

#### Минимальные требования
```yaml
hardware:
  cpu: "4 vCPU"
  ram: "16 GB DDR4"
  storage: "500 GB SSD"
  network: "1 Gbps"
  
software:
  os: "Ubuntu 20.04+ или 22.04+"
  docker: "20.10+"
  docker_compose: "2.0+"
```

#### Рекомендуемые требования
```yaml
hardware:
  cpu: "8 vCPU"
  ram: "32 GB DDR4"
  storage: "2 TB NVMe SSD"
  network: "10 Gbps"
```

#### Production требования
```yaml
hardware:
  cpu: "16 vCPU"
  ram: "64 GB DDR4 ECC"
  storage: "4 TB NVMe SSD (RAID 1)"
  network: "25 Gbps"
```

### Архитектура с отдельным SIEM сервером

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Relay Server  │    │  Keycloak Server│    │  Django Server  │
│                 │    │                 │    │                 │
│ - CloudBridge   │    │ - Authentication│    │ - Management    │
│ - TCP Proxy     │    │ - User Mgmt     │    │ - API           │
│ - Multi-tenant  │    │ - SSO           │    │ - Dashboard     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   SIEM Server   │
                    │                 │
                    │ - Elasticsearch │
                    │ - Kibana        │
                    │ - Kafka         │
                    │ - Collectors    │
                    │ - Processors    │
                    │ - Alert Manager │
                    │ - Web UI        │
                    └─────────────────┘
```

### Интеграция с CloudBridge Relay

```bash
# Настройка сбора данных с Relay сервера
ssh-keygen -t rsa -b 4096 -f /opt/siem/.ssh/id_rsa -N ""
ssh-copy-id -i /opt/siem/.ssh/id_rsa.pub ubuntu@your_relay_server_ip

# Конфигурация удаленного сбора
cat > /opt/siem/siem/config/remote_collectors.yaml << 'EOF'
remote_collectors:
  relay_server:
    host: "your_relay_server_ip"
    port: 22
    username: "ubuntu"
    key_file: "/opt/siem/.ssh/id_rsa"
    sources:
      - path: "/var/log/relay/relay.log"
        type: "relay_logs"
        interval: 60
      - path: "http://localhost:8081/metrics"
        type: "relay_metrics"
        interval: 30
EOF
```

### Доступ к SIEM сервисам

После развертывания будут доступны:
- **SIEM Web UI**: http://localhost:8081
- **Kibana**: http://localhost:5601
- **Grafana**: http://localhost:3000
- **Prometheus**: http://localhost:9090
- **Elasticsearch**: http://localhost:9200

### Управление SIEM

```bash
# Управление сервисом
sudo systemctl start siem
sudo systemctl stop siem
sudo systemctl restart siem
sudo systemctl status siem

# Просмотр логов
sudo journalctl -u siem.service -f
tail -f /opt/siem/logs/*.log

# Мониторинг ресурсов
/opt/siem/scripts/monitor_siem.sh

# Резервное копирование
/opt/siem/scripts/backup_siem.sh
```

### Документация по SIEM

Подробная документация доступна в:
- [📊 Анализ SIEM системы](docs/siem_integration_analysis.md)
- [🖥️ Развертывание на отдельном сервере](docs/siem_separate_server_deployment.md)
- [⚡ Быстрое развертывание](docs/siem_quick_deployment.md)
- [🔗 Руководство по интеграции](docs/siem_integration_guide.md)
- [🔄 Руководство по миграции](docs/siem_migration_guide.md)
- [📋 Полное руководство](docs/SIEM_DEPLOYMENT_COMPLETE_GUIDE.md)

### Автоматизированные скрипты

- [🔧 Скрипт развертывания SIEM сервера](scripts/deploy_siem_server.sh)
- [🔗 Скрипт интеграции](scripts/integrate_siem.sh)

## Новые возможности v2.0

### Ключевые улучшения
- **TCP Proxy**: Полнофункциональный TCP прокси для двунаправленной передачи данных
- **Multi-tenancy**: Поддержка мультитенантности с изоляцией ресурсов и IP-фильтрацией
- **⚡ Go 1.21**: Обновление до Go 1.21.13 с современными зависимостями
- **Enhanced Security**: Улучшенная интеграция с Keycloak и Django fallback
- **Comprehensive Testing**: Полный набор unit и integration тестов
- **Tunnel Manager**: Централизованное управление активными туннелями
- **🛡️ Lynis Security Audit**: Автоматический аудит безопасности с интеграцией Prometheus/Grafana

### Технические улучшения
- **Производительность**: Улучшенная производительность TCP proxy с буферизацией
- **Мониторинг**: Расширенные метрики по тенантам и туннелям
- **Безопасность**: Улучшенная изоляция ресурсов между тенантами
- **Масштабируемость**: Поддержка тысяч одновременных соединений
- **Надежность**: Graceful shutdown и автоматическая очистка ресурсов
- **Security Auditing**: Автоматический аудит безопасности с соответствием ФЗ-152/187

## Основная концепция

CloudBridge Relay v2.0 позволяет безопасно туннелировать доступ к внутренним сервисам (AD/LDAP, RDP, базы данных) через публичный Relay Server с поддержкой multi-tenancy:

* Безопасный доступ к внутренним сервисам без VPN
* Централизованная аутентификация через Keycloak с Django fallback
* Полная изоляция между компаниями (мультитенантность)
* TCP proxy для эффективной передачи данных
* Контроль и аудит операций на стороне Keycloak и Django
* Поведенческий анализ и ML-детекция угроз
* Интеграция с SIEM и внешними системами безопасности

## Архитектура компонентов

### Ключевые компоненты

#### Keycloak (или другой IdP)
* Централизованное хранилище пользователей, групп и ролей
* Поддержка OIDC, OAuth2, LDAP, MFA
* JWKS (RS256) валидация токенов на стороне Relay
* Realm = изолированный клиент/арендатор
* Группы и атрибуты → маппинг на доступ к туннелям/сервисам
* **Fallback на Django аутентификацию при недоступности**

#### Django API
* Управление компаниями, серверами, тарифами, сервисами
* Выдача серверных токенов (Server Client)
* Web UI для администраторов и клиентов
* REST API для внешних систем (SIEM, DLP, партнеры)
* Аудит, биллинг, роли и разграничение доступа
* **Управление multi-tenancy и лимитами ресурсов**

#### Relay Server v2.0 (Go 1.21)
* **TCP Proxy** для прокладки туннелей между Server Client ↔ Desktop Client
* Валидация JWT токенов через Keycloak с Django fallback
* Rate limiting, heartbeat, TLS/mTLS
* **Multi-tenancy** с изоляцией туннелей по tenant
* **Tunnel Manager** для централизованного управления туннелями
* Интеграция с Prometheus (расширенные метрики)
* **Автоматическая очистка неактивных ресурсов**

#### Server Client (Go)
* Развёртывается на стороне защищённой инфраструктуры
* Запрашивает токен у Django → подключается к Relay
* Объявляет список локальных сервисов (LDAP, RDP)
* **Поддержка TCP proxy для эффективной передачи данных**
* **Multi-tenant аутентификация**

#### Desktop Client (Tauri или CLI)
* Выполняет OIDC login через Keycloak
* Получает JWT токен
* Подключается к Relay Server
* Автоматически подключается к нужному туннелю (RDP, SSH)
* Скачивает .rdp-файл или запускает RDP-клиент
* **Поддержка multi-tenancy и TCP proxy**

#### SIEM-модуль (Python + Elasticsearch + Kafka)
* **🖥️ Развертывается на отдельном сервере** для оптимальной производительности и безопасности
* Приём событий из Relay/Django (входы, туннели, ошибки)
* Обогащение: user-agent, IP, geolocation, threat-level
* Анализ в реальном времени + историческое хранилище
* Отправка alert-ов → Telegram/Email/API
* UI-дэшборды (Grafana или встроенные)
* **Мониторинг multi-tenant активности**
* **Автоматическое развертывание** через скрипт `deploy_siem_server.sh`
* **Интеграция с внешними системами** через REST API и webhooks

#### ML-модуль
* Feature extraction: паттерны времени, частоты, роли, IP
* Обучение моделей: классификация/аномалия/кластеризация
* Реакция: mark-as-suspicious, disable token, alert
* Отчёты по пользовательским профилям риска
* Режим обучения + режим обнаружения
* **Анализ поведения по тенантам**

#### Open API Gateway (REST + Webhooks)
* Открытое API для:
  * событий (входы, туннели, логи)
  * управления туннелями
  * просмотра статуса подключения
  * интеграции с DLP/SOAR
  * **управления multi-tenancy**
* Webhooks для событий: tunnel_opened, login_failed, alert_triggered

#### Production Infrastructure
* **High Availability (HA)** - кластеризация Relay Server с балансировкой нагрузки
* **Auto Scaling** - автоматическое масштабирование на основе метрик
* **Edge Networks** - распределенная сеть edge-узлов для глобального покрытия
* **Load Balancing** - L4/L7 балансировка с health checks
* **Circuit Breaker** - защита от каскадных отказов
* **Multi-tenant isolation** - полная изоляция ресурсов между тенантами
* **🖥️ Dedicated SIEM Server** - отдельный сервер для SIEM системы с высокой производительностью

#### Advanced Security
* **mTLS (Mutual TLS)** - двусторонняя аутентификация между Relay и клиентами
* **Misuse Protection** - защита от злоупотреблений и атак
* **Rate Limiting** - многоуровневое ограничение скорости
* **DDoS Protection** - защита от распределенных атак
* **IP Whitelisting** - белые списки IP-адресов по тенантам
* **Session Management** - управление сессиями и их жизненным циклом
* **Tenant isolation** - полная изоляция ресурсов между клиентами

#### Performance & Monitoring
* **Load Testing** - тестирование производительности под нагрузкой
* **Capacity Planning** - планирование ресурсов
* **Performance Metrics** - детальные метрики производительности
* **Resource Optimization** - оптимизация использования ресурсов
* **Multi-tenant metrics** - метрики по тенантам и туннелям

### Общий поток взаимодействия v2.0

#### Server Client:

1. Администратор создаёт сервер через Django UI.
2. Server Client запрашивает токен через Django API с указанием tenant_id.
3. Django делает запрос к Keycloak от имени Server Client.
4. Keycloak возвращает подписанный JWT токен с tenant_id.
5. Server Client подключается к Relay Server с этим токеном.
6. **Relay создает TCP proxy для эффективной передачи данных.**

#### Relay Server v2.0:

1. Получает токен от Server Client с tenant_id.
2. Валидирует его через Keycloak JWKS (RS256) или Django fallback.
3. Проверяет лимиты ресурсов для тенанта.
4. Извлекает информацию о туннелях.
5. **Создаёт TCP proxy туннели и начинает пересылку трафика.**
6. **Отслеживает активность и автоматически очищает неактивные ресурсы.**

#### Desktop Client:

1. Пользователь аутентифицируется через Keycloak (включая LDAP federation).
2. Получает JWT токен с tenant_id.
3. Подключается к Relay Server и получает доступ к нужным туннелям по правам.
4. **Использует TCP proxy для эффективного доступа к сервисам.**

## Потоки данных v2.0

### Аутентификация с Multi-tenancy
```
User → Keycloak (OIDC login) → Access Token (tenant_id) → Relay
Server Client → Django → Keycloak → Token (tenant_id) → Relay
Relay → Keycloak JWKS validation → Django fallback (if needed)
```

### TCP Proxy Туннелирование
```
Desktop Client ↔ Relay TCP Proxy ↔ Server Client ↔ Internal Service
                (Port 10000-20000)  (Bidirectional)
```

### Метрики и события
```
Relay / Django → Prometheus → Grafana (multi-tenant dashboards)
Relay / Django → Kafka → SIEM (отдельный сервер) → AlertManager
SIEM → Telegram, Email, API
```

### Поведенческий анализ
```
SIEM Logs (отдельный сервер) → ML-модель → Risk Score (per tenant)
ML-модель → Response Engine (отключение токена, предупреждение)
```

## TCP Proxy Архитектура

### Модель TCP Proxy

```
┌─────────────────────────────────────────────────────────────────┐
│                    TCP Proxy Architecture                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   Desktop   │    │   TCP       │    │   Server    │         │
│  │   Client    │    │   Proxy     │    │   Client    │         │
│  │             │    │             │    │             │         │
│  │ localhost:  │◄──►│ Port        │◄──►│ Internal    │         │
│  │ 12345       │    │ 12345       │    │ Service     │         │
│  │             │    │ Buffer:     │    │ 192.168.1.  │         │
│  │             │    │ 8KB         │    │ 100:3389    │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
│           │                 │                 │                │
│           ▼                 ▼                 ▼                │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                    Metrics & Monitoring                     │ │
│  │                                                             │ │
│  │ • Bytes transferred                                         │ │
│  │ • Active connections                                        │ │
│  │ • Connection duration                                       │ │
│  │ • Error rates                                               │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### TCP Proxy Features

#### 1. **Bidirectional Data Transfer**
```go
// Двунаправленная передача данных
func (p *TCPProxy) handleConnection(clientConn, serverConn net.Conn) {
    // Client → Server
    go p.transferData(clientConn, serverConn, "client_to_server")
    // Server → Client
    go p.transferData(serverConn, clientConn, "server_to_client")
}
```

#### 2. **Buffer Management**
```go
// Управление буферами
type BufferManager struct {
    BufferSize    int           `json:"buffer_size"`
    MaxBuffers    int           `json:"max_buffers"`
    BufferPool    chan []byte   `json:"buffer_pool"`
}
```

#### 3. **Connection Pooling**
```go
// Пул соединений
type ConnectionPool struct {
    MaxConnections int                    `json:"max_connections"`
    Connections    map[string]*Connection `json:"connections"`
    Mutex          sync.RWMutex           `json:"mutex"`
}
```

## Multi-tenancy Архитектура

### Модель Multi-tenancy

```
┌─────────────────────────────────────────────────────────────────┐
│                    Multi-tenancy Model                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Tenant A      │  │   Tenant B      │  │   Tenant C      │ │
│  │                 │  │                 │  │                 │ │
│  │ • Tunnels: 5/10 │  │ • Tunnels: 2/20 │  │ • Tunnels: 0/5  │ │
│  │ • Connections:  │  │ • Connections:  │  │ • Connections:  │ │
│  │   50/100        │  │   150/200       │  │   0/50          │ │
│  │ • Bandwidth:    │  │ • Bandwidth:    │  │ • Bandwidth:    │ │
│  │   50/100 Mbps   │  │   300/500 Mbps  │  │   0/100 Mbps    │ │
│  │ • IP Range:     │  │ • IP Range:     │  │ • IP Range:     │ │
│  │   192.168.1.0/24│  │   192.168.2.0/24│  │   192.168.3.0/24│ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                    Shared Infrastructure                       │
│  • CPU, Memory, Network                                        │
│  • TLS Certificates                                            │
│  • Monitoring & Metrics                                        │
│  • Rate Limiting                                               │
└─────────────────────────────────────────────────────────────────┘
```

### Multi-tenancy Features

#### 1. **Resource Isolation**
```yaml
tenants:
  tenant_001:
    limits:
      max_tunnels: 20
      max_connections_per_tunnel: 200
      max_bandwidth_mbps: 500
      max_data_transfer_gb: 5000
      rate_limit_per_second: 2000
```

#### 2. **IP Filtering**
```yaml
tenants:
  tenant_001:
    allowed_ips:
      - "192.168.1.0/24"
      - "10.0.0.0/8"
    blocked_ips:
      - "192.168.1.100"
      - "10.0.0.50"
```

#### 3. **Activity Tracking**
```go
type TenantActivity struct {
    TenantID      string    `json:"tenant_id"`
    LastActivity  time.Time `json:"last_activity"`
    ActiveTunnels int       `json:"active_tunnels"`
    TotalTraffic  int64     `json:"total_traffic"`
}
```

## mTLS Архитектура

### Модель взаимной аутентификации

```
┌─────────────────────────────────────────────────────────────────┐
│                    mTLS Authentication Model                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   Desktop   │    │   Relay     │    │   Server    │         │
│  │   Client    │    │   Server    │    │   Client    │         │
│  │             │    │             │    │             │         │
│  │ • Client    │◄──►│ • Server    │◄──►│ • Client    │         │
│  │   Cert      │    │   Cert      │    │   Cert      │         │
│  │ • Private   │    │ • Private   │    │ • Private   │         │
│  │   Key       │    │   Key       │    │   Key       │         │
│  │ • CA Chain  │    │ • CA Chain  │    │ • CA Chain  │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
│           │                 │                 │                │
│           ▼                 ▼                 ▼                │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                    Certificate Authority                   │ │
│  │                                                             │ │
│  │ • Root CA (self-signed)                                    │ │
│  │ • Intermediate CA (for clients)                            │ │
│  │ • Certificate Revocation List (CRL)                       │ │
│  │ • Online Certificate Status Protocol (OCSP)               │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Процесс mTLS handshake

#### 1. **Client Hello**
```json
{
  "client_certificate": "-----BEGIN CERTIFICATE-----...",
  "client_key_exchange": "RSA/ECDHE",
  "cipher_suites": ["TLS_AES_256_GCM_SHA384", "TLS_CHACHA20_POLY1305_SHA256"],
  "supported_groups": ["P-256", "P-384", "X25519"]
}
```

#### 2. **Server Hello**
```json
{
  "server_certificate": "-----BEGIN CERTIFICATE-----...",
  "selected_cipher": "TLS_AES_256_GCM_SHA384",
  "selected_group": "P-256",
  "session_id": "unique-session-id"
}
```

#### 3. **Certificate Verification**
```go
// Проверка сертификатов
func verifyCertificate(cert *x509.Certificate, roots *x509.CertPool) error {
    // Проверка подписи
    if err := cert.CheckSignatureFrom(roots); err != nil {
        return fmt.Errorf("certificate signature verification failed: %v", err)
    }
    
    // Проверка срока действия
    if time.Now().Before(cert.NotBefore) || time.Now().After(cert.NotAfter) {
        return fmt.Errorf("certificate expired or not yet valid")
    }
    
    // Проверка отзыва (CRL/OCSP)
    if err := checkRevocation(cert); err != nil {
        return fmt.Errorf("certificate revoked: %v", err)
    }
    
    return nil
}
```

### Конфигурация mTLS

```json
{
  "tls": {
    "enabled": true,
    "mode": "mutual",
    "cert_file": "/etc/relay/certs/server.crt",
    "key_file": "/etc/relay/certs/server.key",
    "ca_file": "/etc/relay/certs/ca.crt",
    "client_ca_file": "/etc/relay/certs/client-ca.crt",
    "cipher_suites": [
      "TLS_AES_256_GCM_SHA384",
      "TLS_CHACHA20_POLY1305_SHA256"
    ],
    "min_version": "TLS 1.3",
    "session_tickets": true,
    "session_cache_size": 32768
  }
}
```

## Edge Networks Architecture

### Распределенная сеть edge-узлов

```
┌─────────────────────────────────────────────────────────────────┐
│                    Global Edge Network                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   Europe    │    │   Americas  │    │   Asia      │         │
│  │   Region    │    │   Region    │    │   Region    │         │
│  │             │    │             │    │             │         │
│  │ ┌─────────┐ │    │ ┌─────────┐ │    │ ┌─────────┐ │         │
│  │ │ Edge-1  │ │    │ │ Edge-1  │ │    │ │ Edge-1  │ │         │
│  │ │ DE      │ │    │ │ US-East │ │    │ │ JP      │ │         │
│  │ └─────────┘ │    │ └─────────┘ │    │ └─────────┘ │         │
│  │             │    │             │    │             │         │
│  │ ┌─────────┐ │    │ ┌─────────┐ │    │ ┌─────────┐ │         │
│  │ │ Edge-2  │ │    │ │ Edge-2  │ │    │ │ Edge-2  │ │         │
│  │ │ UK      │ │    │ │ US-West │ │    │ │ SG      │ │         │
│  │ └─────────┘ │    │ └─────────┘ │    │ └─────────┘ │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
│           │                 │                 │                │
│           ▼                 ▼                 ▼                │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                    Global Load Balancer                    │ │
│  │                                                             │ │
│  │ • Geo-based routing                                        │ │
│  │ • Health checks                                            │ │
│  │ • Failover logic                                           │ │
│  │ • Traffic optimization                                     │ │
│  └─────────────────────────────────────────────────────────────┘ │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                    Core Infrastructure                     │ │
│  │                                                             │ │
│  │ • Relay Server Cluster                                     │ │
│  │ • Keycloak HA Setup                                        │ │
│  │ • Django API Cluster                                       │ │
│  │ • Database Replication                                     │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Edge Node Configuration

```yaml
# edge-node-config.yaml
edge_nodes:
  - name: "edge-eu-de-1"
    region: "europe"
    country: "DE"
    location: "Frankfurt"
    capacity: "10k_connections"
    upstream:
      - "relay-core-1:8080"
      - "relay-core-2:8080"
    health_check:
      interval: "30s"
      timeout: "5s"
      unhealthy_threshold: 3
      healthy_threshold: 2
    load_balancing:
      algorithm: "least_connections"
      sticky_sessions: true
      session_timeout: "1h"
    security:
      rate_limit: "1000_rps"
      ddos_protection: true
      ip_whitelist: ["10.0.0.0/8", "172.16.0.0/12"]
    monitoring:
      metrics_endpoint: ":8081"
      log_level: "INFO"
      alerting: true
```

### Geo-routing и failover

```go
// Geo-based routing logic
func selectEdgeNode(clientIP net.IP, userAgent string) (*EdgeNode, error) {
    // Определение региона клиента
    region := geoip.Lookup(clientIP)
    
    // Выбор edge-узла в регионе
    edgeNodes := getEdgeNodesInRegion(region)
    
    // Проверка доступности
    availableNodes := filterHealthyNodes(edgeNodes)
    
    if len(availableNodes) == 0 {
        // Failover на ближайший регион
        return selectFailoverNode(region)
    }
    
    // Выбор по алгоритму балансировки
    return selectByLoadBalancingAlgorithm(availableNodes)
}
```

## Production Load Handling

### High Availability (HA) Setup

#### Relay Server Cluster
```yaml
# relay-cluster.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: relay-server
spec:
  replicas: 3
  serviceName: relay-service
  selector:
    matchLabels:
      app: relay-server
  template:
    metadata:
      labels:
        app: relay-server
    spec:
      containers:
      - name: relay
        image: cloudbridge/relay:latest
        ports:
        - containerPort: 8080
        - containerPort: 8081  # metrics
        env:
        - name: RELAY_CLUSTER_MODE
          value: "true"
        - name: RELAY_CLUSTER_NODES
          value: "relay-0,relay-1,relay-2"
        - name: RELAY_ELECTION_TIMEOUT
          value: "5s"
        volumeMounts:
        - name: relay-config
          mountPath: /etc/relay
        - name: relay-certs
          mountPath: /etc/relay/certs
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### Auto Scaling Configuration
```yaml
# relay-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: relay-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: StatefulSet
    name: relay-server
  minReplicas: 3
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
  - type: Object
    object:
      metric:
        name: active_connections
      describedObject:
        apiVersion: v1
        kind: Service
        name: relay-service
      target:
        type: AverageValue
        averageValue: 1000
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
```

### Load Testing и Capacity Planning

#### Load Testing Scenarios
```python
# load_test_scenarios.py
import asyncio
import aiohttp
import time

class LoadTest:
    def __init__(self, relay_url, num_clients=1000):
        self.relay_url = relay_url
        self.num_clients = num_clients
        self.results = []
    
    async def test_connection_establishment(self):
        """Тест установки соединений"""
        start_time = time.time()
        tasks = []
        
        for i in range(self.num_clients):
            task = asyncio.create_task(self.establish_connection(i))
            tasks.append(task)
        
        await asyncio.gather(*tasks)
        duration = time.time() - start_time
        
        return {
            "test": "connection_establishment",
            "clients": self.num_clients,
            "duration": duration,
            "rate": self.num_clients / duration
        }
    
    async def test_tunnel_throughput(self):
        """Тест пропускной способности туннелей"""
        # Реализация теста пропускной способности
        pass
    
    async def test_concurrent_tunnels(self):
        """Тест одновременных туннелей"""
        # Реализация теста одновременных туннелей
        pass

# Сценарии тестирования
scenarios = [
    {"name": "baseline", "clients": 100, "duration": 300},
    {"name": "peak_load", "clients": 1000, "duration": 600},
    {"name": "stress_test", "clients": 5000, "duration": 900},
    {"name": "endurance", "clients": 500, "duration": 3600}
]
```

#### Performance Metrics
```promql
# Метрики производительности
# Пропускная способность туннелей
rate(relay_tunnel_bytes_transferred_total[5m])

# Задержка соединений
histogram_quantile(0.95, rate(relay_connection_duration_seconds_bucket[5m]))

# Использование ресурсов
relay_cpu_usage_percent
relay_memory_usage_bytes
relay_network_io_bytes_total

# Ошибки и отказы
rate(relay_connection_errors_total[5m])
rate(relay_authentication_failures_total[5m])

# Автоскейлинг метрики
relay_active_connections
relay_pending_connections
relay_tunnel_creation_duration_seconds
```

### Misuse Protection

#### Многоуровневая защита
```go
// Misuse Protection Implementation
type MisuseProtector struct {
    rateLimiters map[string]*rate.Limiter
    ipBlacklist  *sync.Map
    userScores   *sync.Map
    rules        []MisuseRule
}

type MisuseRule struct {
    Name        string
    Condition   func(*Connection) bool
    Action      MisuseAction
    Threshold   int
    Window      time.Duration
}

type MisuseAction struct {
    Type        string // "block", "throttle", "alert", "disable"
    Duration    time.Duration
    Message     string
}

// Правила защиты от злоупотреблений
var misuseRules = []MisuseRule{
    {
        Name: "rapid_connections",
        Condition: func(conn *Connection) bool {
            return conn.rate > 100 // > 100 connections per minute
        },
        Action: MisuseAction{
            Type:     "throttle",
            Duration: 5 * time.Minute,
            Message:  "Too many rapid connections",
        },
        Threshold: 5,
        Window:    1 * time.Minute,
    },
    {
        Name: "failed_auth_attempts",
        Condition: func(conn *Connection) bool {
            return conn.authFailures > 10
        },
        Action: MisuseAction{
            Type:     "block",
            Duration: 30 * time.Minute,
            Message:  "Too many failed authentication attempts",
        },
        Threshold: 3,
        Window:    10 * time.Minute,
    },
    {
        Name: "suspicious_patterns",
        Condition: func(conn *Connection) bool {
            return detectSuspiciousPattern(conn)
        },
        Action: MisuseAction{
            Type:     "alert",
            Duration: 0,
            Message:  "Suspicious connection pattern detected",
        },
        Threshold: 1,
        Window:    1 * time.Hour,
    },
}
```

#### DDoS Protection
```go
// DDoS Protection
type DDoSProtector struct {
    ipCounters   *sync.Map
    globalLimiter *rate.Limiter
    whitelist    map[string]bool
    blacklist    map[string]bool
}

func (d *DDoSProtector) CheckConnection(conn *Connection) error {
    // Проверка белого списка
    if d.whitelist[conn.IP] {
        return nil
    }
    
    // Проверка черного списка
    if d.blacklist[conn.IP] {
        return errors.New("IP is blacklisted")
    }
    
    // Глобальное ограничение
    if !d.globalLimiter.Allow() {
        return errors.New("global rate limit exceeded")
    }
    
    // Проверка по IP
    counter := d.getIPCounter(conn.IP)
    if counter.Count() > 1000 { // 1000 connections per minute per IP
        d.blacklist[conn.IP] = true
        return errors.New("IP rate limit exceeded")
    }
    
    return nil
}
```

## Слои архитектуры

### A. Control Plane (управляющий уровень)
* Keycloak
* Django API
* Биллинг
* Админ-панель
* Панель клиента
* API Gateway

### B. Data Plane (трафик)
* Relay Server
* Server Client
* Desktop Client
* TCP-туннели

### C. Observability
* Prometheus
* Grafana
* AlertManager
* Логирование, tracing

### D. Security Intelligence
* SIEM
* ML Engine
* Реакция на угрозы (rules + AI)
* Интеграция с внешними DLP/SOC

## Преимущества архитектуры

* Единый источник доверия — все токены управляются Keycloak
* Безопасность на основе открытых стандартов (OAuth2, OIDC)
* Масштабируемость — можно подключить любые внешние IdP
* Полная изоляция данных между клиентами
* Аудит — все действия логируются в Keycloak и Django
* Поведенческий анализ и ML-детекция угроз
* Интеграция с SIEM и внешними системами безопасности

## Компоненты

### Relay Server

* Принимает соединения от Server Client и Desktop Client
* Валидирует токены через Keycloak (JWKS)
* Создаёт и управляет туннелями
* Проксирует TCP трафик
* Поддержка TLS, rate limiting, мониторинга

### Server Client

* Запрашивает токен через Django API
* Подключается к Relay Server
* Передаёт информацию о туннелях (remote_host, remote_port)
* Проксирует трафик от Relay Server во внутреннюю сеть

### Desktop Client

* Выполняет вход через Keycloak (OIDC flow)
* Получает access token
* Подключается к Relay Server и использует туннели по правам доступа

### Django API

* Предоставляет REST endpoint для генерации токена Server Client
* Делегирует генерацию токена Keycloak
* Хранит информацию о серверах, конфигурации, мониторинге

### Keycloak

* Централизованная система авторизации
* Поддерживает локальных пользователей, LDAP, внешние IdP
* Подписывает JWT токены
* Выдаёт access и refresh токены
* JWKS используется Relay Server для проверки подлинности токенов

## Потоки

### Аутентификация Server Client

```
Админ создает сервер (Django UI)
→ Server Client запрашивает токен (POST /api/token)
→ Django → Keycloak (client credentials grant)
→ Получение токена
→ Server Client → Relay Server (с токеном)
→ Relay Server → JWKS → Валидация
→ Туннель создан
```

### Аутентификация Desktop Client

```
Пользователь логинится через Keycloak (OIDC)
→ Получает access token
→ Подключается к Relay Server
→ Relay Server проверяет подпись токена
→ Доступ разрешён
```

## Пример токена

```json
{
  "sub": "user@example.com",
  "aud": "relay-client",
  "exp": 1710000000,
  "iat": 1709990000,
  "client_id": "server-123",
  "tunnels": [
    {
      "id": "ldap-main",
      "remote_host": "192.168.0.10",
      "remote_port": 389
    }
  ]
}
```

## Протокол соединения с Relay Server

### 1. Установка соединения и приветствие

```json
{
  "type": "hello",
  "version": "1.0"
}
```

Ответ:

```json
{
  "type": "hello_response",
  "version": "1.0"
}
```

### 2. Аутентификация токеном

```json
{
  "type": "auth",
  "token": "<JWT от Keycloak>",
  "role": "server|desktop"
}
```

Ответ:

```json
{
  "type": "auth_response",
  "status": "ok",
  "client_id": "server-123"
}
```

### 3. Информация о туннелях (только для Server Client)

```json
{
  "type": "tunnel_info",
  "tunnels": [
    {
      "id": "ldap-main",
      "remote_host": "192.168.0.10",
      "remote_port": 389
    }
  ]
}
```

Ответ:

```json
{
  "type": "tunnel_response",
  "status": "ok",
  "external_port": 3890
}
```

### 4. Heartbeat

```json
{ "type": "heartbeat" }
```

Ответ:

```json
{ "type": "heartbeat_response" }
```

## Конфигурация валидации токенов (Relay Server)

```json
{
  "keycloak": {
    "enabled": true,
    "jwks_url": "https://keycloak.example.com/realms/cloudbridge/protocol/openid-connect/certs",
    "issuer": "https://keycloak.example.com/realms/cloudbridge",
    "audience": "relay-client"
  }
}
```

## Минимальные требования токена

* Подпись: RS256
* Claims:
  * `sub` — идентификатор пользователя
  * `client_id` — ID сервера или пользователя
  * `exp`, `iat` — срок действия
  * `aud` — audience (`relay-client`)
  * `tunnels` — список разрешённых туннелей (для Server Client)

## Пример запроса токена (Server Client → Django → Keycloak)

```bash
curl -X POST https://api.example.com/api/relay/token \
  -H "Authorization: Bearer <admin_token>" \
  -d '{"server_id": "123"}'
```

Django получает токен от Keycloak через client_credentials flow и возвращает его Server Client.

## Пример входа (Desktop Client → Keycloak)

```bash
curl -X POST https://keycloak.example.com/realms/cloudbridge/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=relay-client" \
  -d "username=user@example.com" \
  -d "password=password"
```

## Защита и контроль

* Все соединения шифруются (TLS 1.3)
* Все токены валидируются через Keycloak JWKS
* Включён rate limiting и защита от DDoS
* Поддержка мультидоменных реалмов и LDAP-федерации
* Изоляция туннелей по `client_id` из токена

## Мониторинг

* Prometheus метрики по туннелям, токенам, соединениям
* Grafana дашборды
* Алерты через AlertManager
* Аудит доступа — по логам Django и Keycloak

## Основные возможности

- ✅ **Туннелирование TCP-соединений** с поддержкой TLS/mTLS
- ✅ **JWT аутентификация** через Keycloak (OAuth 2.0 / OpenID Connect)
- ✅ **AD/LDAP интеграция** через Keycloak User Federation
- ✅ **Мультитенантность** с изоляцией данных
- ✅ **Балансировка нагрузки** (round-robin)
- ✅ **Ограничение скорости запросов** (rate limiting)
- ✅ **Расширенный мониторинг** через Prometheus + Grafana
- ✅ **Алерты и уведомления** через AlertManager
- ✅ **Структурированное логирование** (JSON) с ротацией
- ✅ **Автоматическое переподключение**
- ✅ **Graceful shutdown**
- ✅ **Health checks**
- ✅ **Полное тестовое покрытие** (unit, integration, E2E)
- ✅ **SIEM интеграция** с поведенческим анализом
- ✅ **ML-детекция угроз** и аномалий
- ✅ **Open API Gateway** для внешних интеграций
- ✅ **Webhooks** для событий безопасности
- ✅ **mTLS (Mutual TLS)** - двусторонняя аутентификация
- ✅ **Edge Networks** - глобальная распределенная сеть
- ✅ **High Availability и Auto Scaling**
- ✅ **Продвинутая защита от злоупотреблений**
- ✅ **Load Testing и Capacity Planning**
- ✅ **UDP/QUIC/WireGuard поддержка** — туннелирование UDP-трафика, WireGuard VPN, QUIC/HTTP3, DNS и VoIP, поддержка современных протоколов и низкой задержки
- ✅ **VPN-аналоги** — создание защищённых сервер-сервер туннелей, split tunneling, корпоративный VPN с мониторингом и SIEM

## 🚀 UDP/QUIC/WireGuard и VPN-функционал

CloudBridge Relay поддерживает не только TCP, но и UDP-трафик, что позволяет реализовать:
- **WireGuard VPN** (UDP, современный стандарт корпоративного VPN)
- **QUIC/HTTP3** (UDP, для современных веб-приложений)
- **DNS/VoIP туннелирование** (UDP)
- **Split tunneling** и маршрутизацию через relay
- **Грейсфул рестарты UDP** (без потери соединений)
- **Мониторинг VPN-туннелей** (Prometheus, Grafana)
- **SIEM и алерты по VPN**

### Пример создания VPN-туннеля (WireGuard):
```bash
./scripts/server_to_server_tunnel.sh create \
  -i company-vpn \
  -s1 office-server:51820 \
  -s2 cloud-server:51820 \
  --protocol udp \
  --flow-dissector WIREGUARD \
  --vpn-mode true
```

Подробнее: [docs/vpn_quick_start.md](docs/vpn_quick_start.md), [docs/vpn_analog_implementation.md](docs/vpn_analog_implementation.md)

## 🏢 Коробочное решение и коммерческие возможности

CloudBridge Relay доступен как **коробочное решение** (appliance) для корпоративных клиентов, MSP и интеграторов:
- **CloudBridge Relay Appliance** — сервер с предустановленным ПО, web-интерфейсом, мониторингом, SIEM, поддержкой
- **Virtual Appliance** — OVA/VMware/KVM-образ для быстрой установки
- **MSP Edition** — multi-tenant, white-label, API для биллинга и управления клиентами
- **Лицензирование** — Starter, Professional, Enterprise, Cloud Edition
- **Готовые решения** — SMB, Enterprise, MSP bundles
- **Сертификация и аудит** — ISO 27001, SOC2, GDPR, HIPAA, PCI DSS, FIPS 140-2
- **Техническая поддержка** — SLA, обучение, кастомизация, white-label

**Подробнее:** [docs/commercial_edition.md](docs/commercial_edition.md)

## Безопасность

### Комплексная защита

CloudBridge Relay v2.0 включает многоуровневую систему безопасности:

#### 🔥 Firewall Protection
- **Автоматическая настройка**: `sudo ./scripts/setup_firewall.sh`
- **Блокировка всех портов**: кроме SSH (22) и Relay (8082)
- **Локальный доступ**: порты мониторинга доступны только с localhost
- **DDoS защита**: rate limiting и автоматическая блокировка
- **fail2ban**: защита от брутфорс атак

#### 🔐 Аутентификация и авторизация
- **Keycloak Integration**: OAuth 2.0 / OpenID Connect
- **JWT токены**: безопасная передача информации о пользователе
- **mTLS (Mutual TLS)**: двусторонняя аутентификация
- **Multi-factor authentication**: поддержка MFA через Keycloak

#### 🛡️ Шифрование и сертификаты
- **Let's Encrypt**: автоматическое обновление сертификатов
- **Strong cipher suites**: современные алгоритмы шифрования
- **Российские криптоалгоритмы**: ГОСТ Р 34.10-2012, ГОСТ Р 34.11-2012
- **Perfect Forward Secrecy**: защита от компрометации ключей

#### 📊 Мониторинг безопасности
- **Security Analytics Dashboard**: анализ угроз и атак
- **Real-time alerts**: уведомления о подозрительной активности
- **ML threat detection**: поведенческий анализ
- **Audit logging**: детальное логирование всех событий

#### 🛡️ Lynis Security Audit
- **Автоматический аудит**: ежедневные проверки безопасности системы
- **Интеграция с Prometheus**: метрики безопасности в реальном времени
- **Grafana дашборды**: визуализация результатов аудита
- **Соответствие стандартам**: ФЗ-152, ФЗ-187, ГОСТ Р 57580.1-2017
- **Автоматическое исправление**: устранение выявленных проблем

```bash
# Установка Lynis Security Audit
sudo bash scripts/setup_lynis.sh

# Запуск аудита безопасности
sudo relay-security-audit full

# Управление аудитом
sudo relay-security-manager status
sudo relay-security-manager report
sudo relay-security-manager fix all
```

Подробная документация: [📖 Lynis Integration Guide](docs/lynis_integration.md)

### Обязательные действия после установки

1. **Настройте firewall**:
   ```bash
   # Автоматическая настройка (рекомендуется)
   sudo ./scripts/setup_firewall.sh
   
   # Проверка статуса
   sudo ./scripts/setup_firewall.sh --check
   
   # Управление firewall
   relay-firewall status
   relay-firewall allow-ip 192.168.1.100
   ```

2. **Измените пароли по умолчанию**:
   ```bash
   # Используйте скрипт генерации секретов
   ./scripts/generate_secrets.sh --env .env
   
   # Или измените вручную в файле .env:
   KEYCLOAK_ADMIN_PASSWORD=your-secure-password
   GRAFANA_ADMIN_PASSWORD=your-secure-password
   ```

3. **Настройте HTTPS**:
   ```bash
   # Получение Let's Encrypt сертификатов
   sudo certbot certonly --dns-cloudflare --dns-cloudflare-credentials ~/.secrets/cloudflare.ini -d edge.2gc.ru
   
   # Автоматическое обновление
   sudo crontab -e
   # Добавить: 0 12 * * * /usr/bin/certbot renew --quiet
   ```

4. **Настройте fail2ban**:
   ```bash
   # Проверка статуса
   sudo fail2ban-client status
   
   # Просмотр заблокированных IP
   sudo fail2ban-client status sshd
   sudo fail2ban-client status relay
   ```

5. **Используйте менеджер секретов в продакшене**:
   - HashiCorp Vault
   - AWS Secrets Manager
   - Azure Key Vault
   - Google Secret Manager

### Генерация секретов

```bash
# Генерация всех секретов
./scripts/generate_secrets.sh --all

# Пароль указанной длины
./scripts/generate_secrets.sh --password 20

# .env файл с секретами
./scripts/generate_secrets.sh --env .env.production
```

### Управление безопасностью

#### Firewall Management
```bash
# Показать статус
relay-firewall status

# Разрешить IP адрес
relay-firewall allow-ip 192.168.1.100

# Заблокировать IP адрес
relay-firewall deny-ip 10.0.0.50

# Перезагрузить firewall
relay-firewall reload
```

#### fail2ban Management
```bash
# Проверка статуса
sudo fail2ban-client status

# Разблокировка IP
sudo fail2ban-client set sshd unbanip 192.168.1.100

# Просмотр логов
sudo tail -f /var/log/fail2ban.log
```

#### Certificate Management
```bash
# Проверка сертификатов
sudo certbot certificates

# Тестирование обновления
sudo certbot renew --dry-run

# Post-renewal hook для перезапуска relay
sudo nano /etc/letsencrypt/renewal-hooks/post/restart-relay.sh
```

### Соответствие российским требованиям

- ✅ **152-ФЗ**: Защита персональных данных
- ✅ **187-ФЗ**: Безопасность критической инфраструктуры
- ✅ **ГОСТ Р 34.10-2012**: Алгоритмы подписи
- ✅ **ГОСТ Р 34.11-2012**: Алгоритмы хеширования
- ✅ **ГОСТ Р 57580.1-2017**: Требования к средствам защиты

### Рекомендации по безопасности

#### Регулярные проверки
1. **Ежедневно**: Проверка логов безопасности
2. **Еженедельно**: Обновление системы и пакетов
3. **Ежемесячно**: Аудит настроек безопасности
4. **Ежеквартально**: Тестирование на проникновение

#### Мониторинг
```bash
# Проверка статуса сервисов
docker-compose ps

# Просмотр логов безопасности
docker-compose logs -f relay | grep -i "security\|auth\|error"

# Проверка метрик безопасности
curl http://localhost:8081/metrics | grep -i "security\|auth"
```

#### Резервное копирование
```bash
# Автоматическое резервное копирование
./scripts/backup_keycloak.sh

# Проверка целостности резервных копий
./scripts/test_integration.sh
```

### Документация по безопасности

Подробная документация доступна в:
- [📖 Руководство по безопасности](docs/security.md)
- [🛡️ Lynis Security Audit](docs/lynis_integration.md)
- [🔧 Настройка firewall](scripts/setup_firewall.sh)
- [📊 Дашборды безопасности](docs/dashboard_guide.md)
- [🔐 Интеграция с Keycloak](docs/keycloak_integration.md)
- [🖥️ SIEM на отдельном сервере](docs/SIEM_DEPLOYMENT_COMPLETE_GUIDE.md)
- [📊 Анализ SIEM системы](docs/siem_integration_analysis.md)
- [🔗 Интеграция SIEM](docs/siem_integration_guide.md)

## Тестирование дашбордов и генерация метрик

Для проверки работы Grafana/Prometheus дашбордов и генерации тестовых метрик используйте следующие скрипты из папки `scripts/`:

- `create_real_connections.sh` — имитирует реальные TCP соединения и туннели к relay серверу.
- `generate_business_metrics.sh` — генерирует бизнес-метрики (активные пользователи, доход, мультитенантность).
- `test_vpn_client.sh` — тестирует работу VPN клиента и создает реальные VPN-соединения.
- `test_tls_connections.sh` — создает TLS-соединения к relay серверу (актуально, если сервер работает только по TLS).
- `run_all_tests.sh` — запускает все тесты и демонстрации для комплексной проверки дашбордов.

### Как использовать

1. Убедитесь, что сервисы relay, Prometheus и Grafana запущены (`docker-compose up -d`).
2. Сделайте скрипты исполняемыми (если нужно):
   ```bash
   chmod +x scripts/*.sh
   ```
3. Запустите комплексный тест:
   ```bash
   ./scripts/run_all_tests.sh
   ```
   Скрипт автоматически проверит доступность сервисов, создаст тестовые соединения, сгенерирует бизнес-метрики и покажет ссылки для доступа к дашбордам.

4. Для индивидуального теста используйте нужный скрипт, например:
   ```bash
   ./scripts/create_real_connections.sh
   ./scripts/generate_business_metrics.sh
   ./scripts/test_tls_connections.sh
   ```

### Примечания
- Для появления бизнес-метрик убедитесь, что backend поддерживает соответствующие API (например, `/metrics/business`).
- Для проверки ошибок и отклонённых соединений используйте скрипты, создающие некорректные или plain TCP соединения (`test_tls_connections.sh`).
- Все метрики появятся в Prometheus и будут отображаться на дашбордах Grafana через 10-60 секунд после запуска тестов.

---

**Вопросы и доработки:**
Если дашборды не показывают данные — проверьте логи relay сервера и убедитесь, что тестовые скрипты выполняются без ошибок. Для расширения метрик или интеграции с Keycloak используйте REST API Keycloak или обратитесь к документации в папке `docs/`.
