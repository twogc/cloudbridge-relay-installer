# Техническое задание: Интеграция Active Directory/LDAP с CloudBridge Relay

## 1. Общие сведения

### 1.1 Назначение документа
Данное техническое задание описывает требования к разработке интеграции Active Directory (AD) и LDAP с CloudBridge Relay через Keycloak для централизованного управления пользователями и группами с обеспечением изоляции между компаниями.

### 1.2 Область применения
Интеграция предназначена для:
- Автоматической синхронизации пользователей из AD/LDAP в Keycloak
- Управления группами и ролями на основе AD/LDAP структуры
- Обеспечения изоляции данных между разными компаниями
- Упрощения процесса добавления новых пользователей
- Централизованной аутентификации через корпоративные системы
- **Туннелирования LDAP/AD портов через CloudBridge Relay для безопасного доступа**

### 1.3 Архитектура решения

#### 1.3.1 Базовая архитектура (прямое подключение)
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   AD/LDAP   │    │   Keycloak  │    │   Relay     │    │   Client    │
│   Server    │    │             │    │             │    │             │
│             │    │             │    │             │    │             │
│ 1. Sync     │───▶│ 2. Import   │    │ 3. Validate │    │ 4. Auth     │
│ Users/Groups│    │ 3. Map      │    │ 4. Extract  │    │ 5. Use JWT  │
│             │    │ 4. Isolate  │───▶│ 5. Check    │◀───│             │
│             │    │             │    │ Permissions │    │             │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

#### 1.3.2 Архитектура с туннелированием LDAP/AD
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   AD/LDAP   │    │   Keycloak  │    │   Relay     │    │ Server      │
│   Server    │    │             │    │   Server    │    │ Client      │
│             │    │             │    │             │    │             │
│             │    │ 4. LDAP     │    │ 3. Tunnel   │    │ 1. Create   │
│             │◀───│ Query       │◀───│ Forward     │◀───│ Tunnel      │
│             │    │             │    │             │    │ 2. Connect  │
│             │    │             │    │             │    │ to Relay    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

## 2. Туннелирование LDAP/AD портов через CloudBridge Relay

### 2.1 Принцип работы туннелирования

#### 2.1.1 Архитектура туннелирования
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   AD/LDAP   │    │   Server    │    │   Relay     │    │   Keycloak  │
│   Server    │    │   Client    │    │   Server    │    │             │
│             │    │             │    │             │    │             │
│ 1. LDAP     │◀───│ 2. Tunnel   │◀───│ 3. Proxy    │◀───│ 4. LDAP     │
│ Requests    │    │ Port 389    │    │ LDAP        │    │ Queries     │
│             │    │             │    │ Traffic     │    │             │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                                              │
                                                              ▼
                                                    ┌─────────────┐
                                                    │  Desktop    │
                                                    │   Client    │
                                                    │             │
                                                    │ 5. Auth via │
                                                    │   Keycloak  │
                                                    └─────────────┘
```

#### 2.1.2 Последовательность операций
1. **Пользователь** в личном кабинете настраивает интеграцию с AD/LDAP
2. **Server Client** подключается к Relay Server с JWT токеном
3. **Server Client** создает туннель к локальному AD/LDAP серверу (порт 389/636)
4. **Keycloak** настраивается на подключение к AD/LDAP через Relay Server
5. **Desktop Client** авторизуется через Keycloak
6. **Keycloak** проверяет учетные данные в AD/LDAP через туннель
7. **Relay Server** проксирует LDAP запросы через туннель к AD/LDAP серверу

### 2.2 Конфигурация туннелирования

#### 2.2.1 Конфигурация Server Client
```json
{
  "relay": {
    "host": "relay.example.com",
    "port": 8080,
    "tls": {
      "enabled": true,
      "cert_file": "/etc/relay/certs/client.crt",
      "key_file": "/etc/relay/certs/client.key"
    }
  },
  "tunnels": [
    {
      "id": "ldap-company-a",
      "local_port": 389,
      "remote_host": "ldap.company-a.local",
      "remote_port": 389,
      "protocol": "tcp",
      "company": "company-a",
      "description": "LDAP tunnel for Company A"
    },
    {
      "id": "ldaps-company-a",
      "local_port": 636,
      "remote_host": "ldap.company-a.local",
      "remote_port": 636,
      "protocol": "tcp",
      "company": "company-a",
      "description": "LDAPS tunnel for Company A"
    }
  ]
}
```

#### 2.2.2 Конфигурация Keycloak для туннелированного LDAP
```json
{
  "ldap": {
    "enabled": true,
    "server_url": "ldap://relay.example.com:389",
    "bind_dn": "CN=SyncUser,OU=ServiceAccounts,DC=example,DC=com",
    "bind_password": "REPLACE_WITH_SECURE_PASSWORD",
    "use_ssl": false,
    "companies": [
      {
        "name": "company-a",
        "realm": "company-a",
        "ldap_url": "ldap://relay.example.com:389",
        "ldaps_url": "ldap://relay.example.com:636",
        "base_dn": "OU=CompanyA,DC=example,DC=com",
        "user_dn": "OU=Users,OU=CompanyA,DC=example,DC=com",
        "group_dn": "OU=Groups,OU=CompanyA,DC=example,DC=com"
      },
      {
        "name": "company-b",
        "realm": "company-b",
        "ldap_url": "ldap://relay.example.com:3890",
        "base_dn": "OU=CompanyB,DC=example,DC=com",
        "user_dn": "OU=Users,OU=CompanyB,DC=example,DC=com",
        "group_dn": "OU=Groups,OU=CompanyB,DC=example,DC=com"
      }
    ]
  }
}
```

### 2.3 Интеграция с личным кабинетом

#### 2.3.1 API для настройки LDAP интеграции
```go
// API endpoints для личного кабинета
type LDAPIntegrationAPI struct {
    // Настройка LDAP интеграции
    POST /api/v1/companies/{company_id}/ldap/integration
    {
        "enabled": true,
        "ldap_server": "ldap.company-a.local",
        "ldap_port": 389,
        "ldaps_port": 636,
        "bind_dn": "CN=SyncUser,OU=ServiceAccounts,DC=company-a,DC=local",
        "bind_password": "secure_password",
        "base_dn": "OU=CompanyA,DC=company-a,DC=local",
        "user_dn": "OU=Users,OU=CompanyA,DC=company-a,DC=local",
        "group_dn": "OU=Groups,OU=CompanyA,DC=company-a,DC=local",
        "sync_users": true,
        "sync_groups": true,
        "sync_interval": "1h"
    }
    
    // Получение статуса LDAP интеграции
    GET /api/v1/companies/{company_id}/ldap/integration
    
    // Тестирование LDAP подключения
    POST /api/v1/companies/{company_id}/ldap/test-connection
    
    // Синхронизация пользователей
    POST /api/v1/companies/{company_id}/ldap/sync-users
    
    // Синхронизация групп
    POST /api/v1/companies/{company_id}/ldap/sync-groups
    
    // Отключение LDAP интеграции
    DELETE /api/v1/companies/{company_id}/ldap/integration
}
```

#### 2.3.2 Интерфейс личного кабинета
```javascript
// Пример интерфейса для настройки LDAP интеграции
class LDAPIntegrationManager {
    async setupLDAPIntegration(companyId, config) {
        const response = await fetch(`/api/v1/companies/${companyId}/ldap/integration`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${this.token}`
            },
            body: JSON.stringify(config)
        });
        
        if (!response.ok) {
            throw new Error('Failed to setup LDAP integration');
        }
        
        return await response.json();
    }
    
    async testLDAPConnection(companyId, config) {
        const response = await fetch(`/api/v1/companies/${companyId}/ldap/test-connection`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${this.token}`
            },
            body: JSON.stringify(config)
        });
        
        return await response.json();
    }
    
    async syncUsers(companyId) {
        const response = await fetch(`/api/v1/companies/${companyId}/ldap/sync-users`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.token}`
            }
        });
        
        return await response.json();
    }
    
    async getIntegrationStatus(companyId) {
        const response = await fetch(`/api/v1/companies/${companyId}/ldap/integration`, {
            headers: {
                'Authorization': `Bearer ${this.token}`
            }
        });
        
        return await response.json();
    }
}

// Пример использования в UI
class LDAPIntegrationUI {
    constructor(companyId) {
        this.companyId = companyId;
        this.manager = new LDAPIntegrationManager();
    }
    
    async setupIntegration() {
        const config = {
            enabled: true,
            ldap_server: document.getElementById('ldap-server').value,
            ldap_port: parseInt(document.getElementById('ldap-port').value),
            bind_dn: document.getElementById('bind-dn').value,
            bind_password: document.getElementById('bind-password').value,
            base_dn: document.getElementById('base-dn').value,
            user_dn: document.getElementById('user-dn').value,
            group_dn: document.getElementById('group-dn').value,
            sync_users: document.getElementById('sync-users').checked,
            sync_groups: document.getElementById('sync-groups').checked
        };
        
        try {
            // Сначала тестируем подключение
            const testResult = await this.manager.testLDAPConnection(this.companyId, config);
            if (!testResult.success) {
                throw new Error(`LDAP connection test failed: ${testResult.error}`);
            }
            
            // Настраиваем интеграцию
            const result = await this.manager.setupLDAPIntegration(this.companyId, config);
            this.showSuccess('LDAP integration setup successfully');
            
            // Запускаем синхронизацию
            if (config.sync_users) {
                await this.manager.syncUsers(this.companyId);
            }
            
        } catch (error) {
            this.showError(`Failed to setup LDAP integration: ${error.message}`);
        }
    }
}
```

### 2.4 Автоматическая настройка туннелей

#### 2.4.1 Создание туннеля при настройке LDAP интеграции
```go
// internal/ldap/integration.go
type LDAPIntegration struct {
    CompanyID     string
    Enabled       bool
    LDAPServer    string
    LDAPPort      int
    LDAPSPort     int
    BindDN        string
    BindPassword  string
    BaseDN        string
    UserDN        string
    GroupDN       string
    SyncUsers     bool
    SyncGroups    bool
    SyncInterval  time.Duration
}

func (i *LDAPIntegration) Setup() error {
    // 1. Создание туннеля к LDAP серверу
    tunnelConfig := &TunnelConfig{
        ID:          fmt.Sprintf("ldap-%s", i.CompanyID),
        LocalPort:   i.LDAPPort,
        RemoteHost:  i.LDAPServer,
        RemotePort:  i.LDAPPort,
        Protocol:    "tcp",
        Company:     i.CompanyID,
        Description: fmt.Sprintf("LDAP tunnel for %s", i.CompanyID),
    }
    
    err := i.createTunnel(tunnelConfig)
    if err != nil {
        return fmt.Errorf("failed to create LDAP tunnel: %v", err)
    }
    
    // 2. Настройка Keycloak LDAP User Federation
    err = i.configureKeycloakLDAP()
    if err != nil {
        return fmt.Errorf("failed to configure Keycloak LDAP: %v", err)
    }
    
    // 3. Запуск синхронизации
    if i.SyncUsers {
        err = i.syncUsers()
        if err != nil {
            return fmt.Errorf("failed to sync users: %v", err)
        }
    }
    
    if i.SyncGroups {
        err = i.syncGroups()
        if err != nil {
            return fmt.Errorf("failed to sync groups: %v", err)
        }
    }
    
    return nil
}

func (i *LDAPIntegration) createTunnel(config *TunnelConfig) error {
    // Создание туннеля через Server Client API
    tunnelRequest := map[string]interface{}{
        "type": "tunnel_info",
        "tunnel_id": config.ID,
        "local_port": config.LocalPort,
        "remote_host": config.RemoteHost,
        "remote_port": config.RemotePort,
        "protocol": config.Protocol,
        "company": config.Company,
        "description": config.Description,
    }
    
    // Отправка запроса на создание туннеля
    return i.sendTunnelRequest(tunnelRequest)
}

func (i *LDAPIntegration) configureKeycloakLDAP() error {
    // Настройка LDAP User Federation в Keycloak
    ldapConfig := map[string]interface{}{
        "enabled": true,
        "connectionUrl": fmt.Sprintf("ldap://relay.example.com:%d", i.LDAPPort),
        "bindDn": i.BindDN,
        "bindCredential": i.BindPassword,
        "searchBase": i.BaseDN,
        "userDn": i.UserDN,
        "groupDn": i.GroupDN,
        "syncRegistrations": true,
        "importUsers": true,
        "periodicFullSync": true,
        "periodicChangedUsersSync": true,
        "fullSyncPeriod": int(i.SyncInterval.Minutes()),
        "changedSyncPeriod": int(i.SyncInterval.Minutes()),
    }
    
    return i.updateKeycloakLDAPConfig(ldapConfig)
}
```

### 2.5 Процесс авторизации Desktop Client

#### 2.5.1 Последовательность авторизации
```go
// internal/auth/ldap_auth.go
type LDAPAuthFlow struct {
    keycloakClient *keycloak.Client
    logger         *logger.Logger
}

func (a *LDAPAuthFlow) AuthenticateUser(username, password string) (*UserInfo, error) {
    // 1. Desktop Client отправляет учетные данные в Keycloak
    // 2. Keycloak проверяет учетные данные в AD/LDAP через туннель
    // 3. Если аутентификация успешна, Keycloak возвращает JWT токен
    // 4. Desktop Client использует JWT токен для доступа к Relay Server
    
    // Создание токена через Keycloak
    tokenResponse, err := a.keycloakClient.GetToken(username, password)
    if err != nil {
        return nil, fmt.Errorf("authentication failed: %v", err)
    }
    
    // Валидация токена
    userInfo, err := a.keycloakClient.ValidateToken(tokenResponse.AccessToken)
    if err != nil {
        return nil, fmt.Errorf("token validation failed: %v", err)
    }
    
    return userInfo, nil
}
```

#### 2.5.2 Интеграция с Desktop Client
```javascript
// desktop-client/auth.js
class LDAPAuthenticator {
    constructor(keycloakUrl, realm, clientId) {
        this.keycloakUrl = keycloakUrl;
        this.realm = realm;
        this.clientId = clientId;
    }
    
    async authenticate(username, password) {
        try {
            // Получение токена от Keycloak
            const tokenResponse = await fetch(`${this.keycloakUrl}/realms/${this.realm}/protocol/openid-connect/token`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded'
                },
                body: new URLSearchParams({
                    grant_type: 'password',
                    client_id: this.clientId,
                    username: username,
                    password: password
                })
            });
            
            if (!tokenResponse.ok) {
                throw new Error('Authentication failed');
            }
            
            const tokenData = await tokenResponse.json();
            
            // Использование токена для подключения к Relay Server
            await this.connectToRelay(tokenData.access_token);
            
            return tokenData;
            
        } catch (error) {
            console.error('LDAP authentication failed:', error);
            throw error;
        }
    }
    
    async connectToRelay(accessToken) {
        // Подключение к Relay Server с JWT токеном от Keycloak
        const relayClient = new RelayClient('relay.example.com', 8080);
        await relayClient.connect();
        await relayClient.authenticate(accessToken);
        
        return relayClient;
    }
}
```

### 2.6 Мониторинг и логирование

#### 2.6.1 Метрики LDAP интеграции
```go
type LDAPMetrics struct {
    IntegrationCount    prometheus.Gauge
    SyncDuration        prometheus.Histogram
    SyncErrors          prometheus.Counter
    AuthenticationTime  prometheus.Histogram
    TunnelConnections   prometheus.Gauge
    LDAPQueries         prometheus.Counter
}

func (m *LDAPMetrics) RecordAuthentication(duration time.Duration, success bool) {
    m.AuthenticationTime.Observe(duration.Seconds())
    if !success {
        m.SyncErrors.Inc()
    }
}

func (m *LDAPMetrics) RecordLDAPQuery(queryType string) {
    m.LDAPQueries.WithLabelValues(queryType).Inc()
}
```

#### 2.6.2 Логирование операций
```go
func (i *LDAPIntegration) logOperation(operation string, details map[string]interface{}) {
    i.logger.Info("LDAP integration operation",
        "operation", operation,
        "company_id", i.CompanyID,
        "details", details)
}

// Примеры использования
i.logOperation("tunnel_created", map[string]interface{}{
    "tunnel_id": tunnelConfig.ID,
    "local_port": tunnelConfig.LocalPort,
    "remote_host": tunnelConfig.RemoteHost,
})

i.logOperation("user_sync_completed", map[string]interface{}{
    "users_synced": userCount,
    "duration": syncDuration,
    "errors": errorCount,
})
```

## 3. Технические требования

### 3.1 Поддерживаемые протоколы

#### 3.1.1 Active Directory
- **Протокол:** LDAP over SSL/TLS (LDAPS)
- **Порты:** 636 (LDAPS), 389 (LDAP)
- **Аутентификация:** NTLM, Kerberos, Simple Bind
- **Поиск:** Global Catalog (GC) на порту 3268/3269

#### 3.1.2 LDAP
- **Протокол:** LDAP v3
- **Порты:** 636 (LDAPS), 389 (LDAP)
- **Аутентификация:** Simple Bind, SASL
- **Поиск:** Стандартные LDAP фильтры

### 3.2 Структура данных AD/LDAP

#### 3.2.1 Пользователи
```ldap
# Пример структуры пользователя в AD
dn: CN=John Doe,OU=Users,OU=CompanyA,DC=example,DC=com
objectClass: user
sAMAccountName: jdoe
userPrincipalName: jdoe@example.com
mail: john.doe@example.com
displayName: John Doe
givenName: John
sn: Doe
memberOf: CN=RelayUsers,OU=Groups,OU=CompanyA,DC=example,DC=com
memberOf: CN=RelayAdmins,OU=Groups,OU=CompanyA,DC=example,DC=com
company: CompanyA
department: IT
title: System Administrator
```

#### 3.2.2 Группы
```ldap
# Пример структуры группы в AD
dn: CN=RelayUsers,OU=Groups,OU=CompanyA,DC=example,DC=com
objectClass: group
sAMAccountName: RelayUsers
displayName: Relay Users
description: Users with basic relay access
member: CN=John Doe,OU=Users,OU=CompanyA,DC=example,DC=com
member: CN=Jane Smith,OU=Users,OU=CompanyA,DC=example,DC=com
```

### 3.3 Маппинг данных

#### 3.3.1 Пользователи
| AD/LDAP атрибут | Keycloak атрибут | Описание |
|-----------------|------------------|----------|
| sAMAccountName | username | Логин пользователя |
| userPrincipalName | email | Email адрес |
| displayName | firstName + lastName | Полное имя |
| givenName | firstName | Имя |
| sn | lastName | Фамилия |
| company | company | Компания |
| department | department | Отдел |
| title | title | Должность |
| memberOf | groups | Членство в группах |

#### 3.3.2 Группы
| AD/LDAP группа | Keycloak роль | Описание |
|----------------|---------------|----------|
| RelayUsers | relay-user | Базовый доступ к relay |
| RelayAdmins | relay-admin | Административный доступ |
| RelayReadOnly | relay-readonly | Только чтение |
| RelayMonitors | relay-monitor | Доступ к мониторингу |

## 4. Требования к изоляции

### 4.1 Изоляция по компаниям

#### 4.1.1 Структура realms
```
Keycloak
├── realm: company-a
│   ├── users: company-a-users
│   ├── groups: company-a-groups
│   └── clients: company-a-clients
├── realm: company-b
│   ├── users: company-b-users
│   ├── groups: company-b-groups
│   └── clients: company-b-clients
└── realm: company-c
    ├── users: company-c-users
    ├── groups: company-c-groups
    └── clients: company-c-clients
```

#### 4.1.2 Правила изоляции
- **Полная изоляция данных:** Пользователи одной компании не могут видеть данные другой
- **Изоляция групп:** Группы создаются в рамках своего realm
- **Изоляция ролей:** Роли назначаются только в рамках своего realm
- **Изоляция клиентов:** Каждая компания имеет свои клиенты

### 4.2 Конфигурация изоляции

#### 4.2.1 Маппинг компаний
```yaml
# config/company_mapping.yaml
companies:
  company-a:
    realm: company-a
    ldap_base_dn: "OU=CompanyA,DC=example,DC=com"
    ldap_user_dn: "OU=Users,OU=CompanyA,DC=example,DC=com"
    ldap_group_dn: "OU=Groups,OU=CompanyA,DC=example,DC=com"
    allowed_groups:
      - "RelayUsers"
      - "RelayAdmins"
      - "RelayReadOnly"
    role_mapping:
      RelayUsers: relay-user
      RelayAdmins: relay-admin
      RelayReadOnly: relay-readonly
      
  company-b:
    realm: company-b
    ldap_base_dn: "OU=CompanyB,DC=example,DC=com"
    ldap_user_dn: "OU=Users,OU=CompanyB,DC=example,DC=com"
    ldap_group_dn: "OU=Groups,OU=CompanyB,DC=example,DC=com"
    allowed_groups:
      - "RelayUsers"
      - "RelayAdmins"
    role_mapping:
      RelayUsers: relay-user
      RelayAdmins: relay-admin
```

## 5. Требования к реализации

### 5.1 Компоненты системы

#### 5.1.1 LDAP Sync Service
```go
// internal/ldap/sync.go
type LDAPSyncService struct {
    config     *LDAPConfig
    keycloak   *keycloak.Client
    logger     *logger.Logger
    companies  map[string]*CompanyConfig
}

type LDAPConfig struct {
    ServerURL      string
    BindDN         string
    BindPassword   string
    UseSSL         bool
    InsecureSkipVerify bool
    Timeout        time.Duration
    SyncInterval   time.Duration
}

type CompanyConfig struct {
    Name           string
    Realm          string
    BaseDN         string
    UserDN         string
    GroupDN        string
    AllowedGroups  []string
    RoleMapping    map[string]string
}
```

#### 5.1.2 LDAP Client
```go
// internal/ldap/client.go
type LDAPClient struct {
    conn     *ldap.Conn
    config   *LDAPConfig
    logger   *logger.Logger
}

func (c *LDAPClient) SearchUsers(baseDN string, filter string) ([]*User, error)
func (c *LDAPClient) SearchGroups(baseDN string, filter string) ([]*Group, error)
func (c *LDAPClient) GetUserGroups(userDN string) ([]string, error)
```

#### 5.1.3 Keycloak Integration
```go
// internal/keycloak/ldap_integration.go
type LDAPIntegration struct {
    client    *Client
    sync      *LDAPSyncService
    logger    *logger.Logger
}

func (i *LDAPIntegration) SyncCompany(company *CompanyConfig) error
func (i *LDAPIntegration) CreateUser(realm string, user *User) error
func (i *LDAPIntegration) UpdateUser(realm string, user *User) error
func (i *LDAPIntegration) DeleteUser(realm string, username string) error
func (i *LDAPIntegration) SyncGroups(realm string, groups []*Group) error
```

### 5.2 Конфигурация

#### 5.2.1 Основная конфигурация
```json
{
  "ldap": {
    "enabled": true,
    "server_url": "ldaps://ldap.example.com:636",
    "bind_dn": "CN=SyncUser,OU=ServiceAccounts,DC=example,DC=com",
    "bind_password": "REPLACE_WITH_SECURE_PASSWORD",
    "use_ssl": true,
    "insecure_skip_verify": false,
    "timeout": "30s",
    "sync_interval": "15m",
    "companies": [
      {
        "name": "company-a",
        "realm": "company-a",
        "ldap_base_dn": "OU=CompanyA,DC=example,DC=com",
        "ldap_user_dn": "OU=Users,OU=CompanyA,DC=example,DC=com",
        "ldap_group_dn": "OU=Groups,OU=CompanyA,DC=example,DC=com",
        "allowed_groups": ["RelayUsers", "RelayAdmins", "RelayReadOnly"],
        "role_mapping": {
          "RelayUsers": "relay-user",
          "RelayAdmins": "relay-admin",
          "RelayReadOnly": "relay-readonly"
        }
      }
    ]
  }
}
```

#### 5.2.2 Переменные окружения
```bash
# LDAP Configuration
export LDAP_SERVER_URL="ldaps://ldap.example.com:636"
export LDAP_BIND_DN="CN=SyncUser,OU=ServiceAccounts,DC=example,DC=com"
export LDAP_BIND_PASSWORD="REPLACE_WITH_SECURE_PASSWORD"
export LDAP_USE_SSL="true"
export LDAP_INSECURE_SKIP_VERIFY="false"
export LDAP_TIMEOUT="30s"
export LDAP_SYNC_INTERVAL="15m"

# Company Configuration
export COMPANY_A_REALM="company-a"
export COMPANY_A_LDAP_BASE_DN="OU=CompanyA,DC=example,DC=com"
export COMPANY_A_ALLOWED_GROUPS="RelayUsers,RelayAdmins,RelayReadOnly"

export COMPANY_B_REALM="company-b"
export COMPANY_B_LDAP_BASE_DN="OU=CompanyB,DC=example,DC=com"
export COMPANY_B_ALLOWED_GROUPS="RelayUsers,RelayAdmins"
```

### 5.3 Процесс синхронизации

#### 5.3.1 Алгоритм синхронизации
1. **Подключение к LDAP/AD**
2. **Для каждой компании:**
   - Поиск пользователей в соответствующем OU
   - Поиск групп в соответствующем OU
   - Сопоставление пользователей с группами
   - Синхронизация с Keycloak realm
3. **Обработка изменений:**
   - Создание новых пользователей
   - Обновление существующих пользователей
   - Удаление неактивных пользователей
   - Синхронизация членства в группах

#### 5.3.2 Обработка ошибок
```go
type SyncError struct {
    Company     string
    Operation   string
    User        string
    Error       error
    Timestamp   time.Time
}

func (s *LDAPSyncService) handleSyncError(err *SyncError) {
    s.logger.Error("Sync error", 
        "company", err.Company,
        "operation", err.Operation,
        "user", err.User,
        "error", err.Error)
    
    // Отправка уведомления администратору
    s.sendNotification(err)
    
    // Продолжение синхронизации других компаний
}
```

## 6. Безопасность

### 6.1 Аутентификация LDAP

#### 6.1.1 Безопасное подключение
- **Обязательно:** LDAPS (LDAP over SSL/TLS)
- **Сертификаты:** Проверка сертификатов сервера
- **Аутентификация:** Использование service account
- **Пароли:** Хранение в менеджере секретов

#### 6.1.2 Service Account
```ldap
# Создание service account для синхронизации
dn: CN=RelaySync,OU=ServiceAccounts,DC=example,DC=com
objectClass: user
sAMAccountName: relaysync
userPrincipalName: relaysync@example.com
description: Service account for CloudBridge Relay LDAP sync
userAccountControl: 512
```

### 6.2 Изоляция данных

#### 6.2.1 Разделение realms
- Каждая компания имеет отдельный realm в Keycloak
- Пользователи не могут пересекать границы realms
- Группы и роли изолированы по компаниям

#### 6.2.2 Контроль доступа
```go
func (c *Client) CheckCompanyAccess(tokenInfo *TokenInfo, company string) bool {
    // Проверка, что пользователь принадлежит к указанной компании
    userCompany := tokenInfo.Claims["company"].(string)
    return userCompany == company
}
```

### 6.3 Аудит и логирование

#### 6.3.1 Логирование синхронизации
```json
{
  "timestamp": "2024-03-14T12:00:00Z",
  "level": "INFO",
  "message": "LDAP sync completed",
  "fields": {
    "company": "company-a",
    "users_created": 5,
    "users_updated": 12,
    "users_deleted": 2,
    "groups_synced": 3,
    "duration": "2.5s"
  }
}
```

#### 6.3.2 Метрики синхронизации
```prometheus
# Метрики LDAP синхронизации
relay_ldap_sync_total{company="company-a",result="success"}
relay_ldap_sync_total{company="company-a",result="failure"}
relay_ldap_users_synced_total{company="company-a",operation="created"}
relay_ldap_users_synced_total{company="company-a",operation="updated"}
relay_ldap_users_synced_total{company="company-a",operation="deleted"}
relay_ldap_sync_duration_seconds{company="company-a"}
```

## 7. Мониторинг и алерты

### 7.1 Алерты синхронизации

```yaml
# prometheus/alerts/ldap_sync_alerts.yml
groups:
  - name: ldap_sync
    rules:
      - alert: LDAPSyncFailure
        expr: relay_ldap_sync_total{result="failure"} > 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "LDAP синхронизация завершилась с ошибкой"
          description: "Синхронизация LDAP для компании {{ $labels.company }} завершилась с ошибкой"
          
      - alert: LDAPSyncDelay
        expr: time() - relay_ldap_sync_last_success_timestamp > 900
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "LDAP синхронизация не выполнялась более 15 минут"
          
      - alert: HighUserDeletionRate
        expr: rate(relay_ldap_users_synced_total{operation="deleted"}[5m]) > 10
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "Высокий темп удаления пользователей из LDAP"
```

### 7.2 Дашборды Grafana

#### 7.2.1 LDAP Sync Dashboard
- Количество синхронизированных пользователей по компаниям
- Время выполнения синхронизации
- Количество ошибок синхронизации
- Статус подключения к LDAP серверам

#### 7.2.2 User Management Dashboard
- Активные пользователи по компаниям
- Распределение по группам и ролям
- Статистика аутентификации
- Активность пользователей

## 8. Тестирование

### 8.1 Unit тесты

```go
// tests/ldap/sync_test.go
func TestLDAPSyncService_SyncCompany(t *testing.T) {
    // Тестирование синхронизации одной компании
}

func TestLDAPSyncService_UserMapping(t *testing.T) {
    // Тестирование маппинга пользователей
}

func TestLDAPSyncService_GroupMapping(t *testing.T) {
    // Тестирование маппинга групп
}

func TestLDAPSyncService_Isolation(t *testing.T) {
    // Тестирование изоляции между компаниями
}
```

### 8.2 Integration тесты

```go
// tests/integration/ldap_integration_test.go
func TestLDAPKeycloakIntegration(t *testing.T) {
    // Тестирование полной интеграции LDAP → Keycloak → Relay
}

func TestMultiCompanyIsolation(t *testing.T) {
    // Тестирование изоляции между несколькими компаниями
}
```

### 8.3 Security тесты

```go
// tests/security/ldap_security_test.go
func TestLDAPConnectionSecurity(t *testing.T) {
    // Тестирование безопасности подключения к LDAP
}

func TestDataIsolation(t *testing.T) {
    // Тестирование изоляции данных между компаниями
}
```

## 9. Документация

### 9.1 Обязательная документация
- README с инструкциями по настройке
- Конфигурационные файлы с примерами
- Troubleshooting guide
- Security considerations

### 9.2 Рекомендуемая документация
- Architecture overview
- Performance tuning guide
- Migration guide
- Best practices

## 10. Критерии приемки

### 10.1 Функциональные требования
- [ ] Успешная синхронизация пользователей из AD/LDAP
- [ ] Успешная синхронизация групп из AD/LDAP
- [ ] Корректный маппинг атрибутов
- [ ] Изоляция данных между компаниями
- [ ] Автоматическая синхронизация по расписанию
- [ ] Обработка ошибок и восстановление

### 10.2 Нефункциональные требования
- [ ] Время синхронизации < 5 минут для 1000 пользователей
- [ ] Доступность системы 99.9%
- [ ] Безопасное хранение учетных данных
- [ ] Полное логирование операций
- [ ] Мониторинг и алерты

### 10.3 Требования безопасности
- [ ] Использование LDAPS
- [ ] Проверка сертификатов
- [ ] Изоляция данных между компаниями
- [ ] Аудит всех операций
- [ ] Безопасное хранение паролей

## 11. Сроки и этапы

### 11.1 Этап 1 (2-3 недели)
- Базовая реализация LDAP клиента
- Интеграция с Keycloak
- Синхронизация пользователей одной компании

### 11.2 Этап 2 (1-2 недели)
- Поддержка множественных компаний
- Изоляция данных
- Синхронизация групп

### 11.3 Этап 3 (1 неделя)
- Мониторинг и алерты
- Тестирование
- Документация

### 11.4 Этап 4 (1 неделя)
- Code review
- Security audit
- Развертывание в продакшен

## 12. Риски и митигация

### 12.1 Технические риски
- **Риск:** Проблемы с подключением к LDAP
- **Митигация:** Retry логика, fallback механизмы

- **Риск:** Высокая нагрузка на LDAP сервер
- **Митигация:** Оптимизация запросов, кэширование

### 12.2 Безопасность риски
- **Риск:** Компрометация учетных данных
- **Митигация:** Использование менеджеров секретов

- **Риск:** Нарушение изоляции данных
- **Митигация:** Строгие проверки доступа, тестирование

## 13. Контакты

### 13.1 Команда разработки
- **Lead Developer:** [Имя] - [Email]
- **Security Specialist:** [Имя] - [Email]
- **DevOps Engineer:** [Имя] - [Email]

### 13.2 Зависимости
- **Keycloak Administrator:** [Имя] - [Email]
- **LDAP Administrator:** [Имя] - [Email]
- **Network Administrator:** [Имя] - [Email]

---

**Дата создания:** 2024-03-14  
**Версия:** 1.0  
**Статус:** Черновик  
**Утверждено:** [Имя] - [Дата] 