# 🌐 Zitadel Tenant Network Integration

## 🎯 **Цель интеграции**

Перенести управление подсетями тенантов из CloudBridge Relay в Zitadel для обеспечения:
- ✅ **Единого источника истины** для всех данных о тенантах
- ✅ **Централизованного управления** сетевыми конфигурациями
- ✅ **Интеграции с RBAC** для безопасности
- ✅ **Масштабируемости** и консистентности

## 🏗️ **Архитектура**

### **Zitadel (Identity Provider + Network Management):**
```
┌─────────────────────────────────────┐
│              Zitadel                │
├─────────────────────────────────────┤
│  ✅ Tenant Management               │
│  ✅ User Authentication             │
│  ✅ JWT Tokens                      │
│  ✅ RBAC Permissions                │
│  🆕 Tenant Network Configs          │ ← НОВОЕ
│  🆕 IPAM Pool Management            │ ← НОВОЕ
│  🆕 Subnet Allocation               │ ← НОВОЕ
└─────────────────────────────────────┘
```

### **CloudBridge Relay (Stateless Application):**
```
┌─────────────────────────────────────┐
│         CloudBridge Relay           │
├─────────────────────────────────────┤
│  ✅ P2P Networking                  │
│  ✅ MASQUE Protocol                 │
│  ✅ TURN/STUN Servers               │
│  🆕 IP Allocation (from Zitadel)    │ ← ИЗМЕНЕНО
│  🆕 Network Config Cache            │ ← НОВОЕ
│  ❌ Local Subnet Storage            │ ← УДАЛЕНО
└─────────────────────────────────────┘
```

## 🔧 **API Endpoints**

### **1. Tenant Network Configuration:**

```bash
# Получение сетевой конфигурации тенанта
GET /management/v1/tenants/{tenant_id}/network-config

# Создание сетевой конфигурации
POST /management/v1/tenants/{tenant_id}/network-config

# Обновление сетевой конфигурации
PUT /management/v1/tenants/{tenant_id}/network-config

# Удаление сетевой конфигурации
DELETE /management/v1/tenants/{tenant_id}/network-config
```

### **2. IPAM Pool Management:**

```bash
# Получение списка IPAM пулов
GET /management/v1/ipam/pools

# Создание IPAM пула
POST /management/v1/ipam/pools

# Выделение подсети из пула
POST /management/v1/ipam/pools/{pool_id}/allocate-subnet

# Освобождение подсети
DELETE /management/v1/ipam/pools/{pool_id}/subnets/{subnet}
```

### **3. Validation & Monitoring:**

```bash
# Проверка изоляции тенантов
GET /management/v1/ipam/validate-isolation

# Синхронизация с CloudBridge Relay
POST /management/v1/sync/cloudbridge-relay
```

## 📊 **Data Models**

### **TenantNetworkConfig:**
```json
{
  "tenant_id": "tenant-216420165",
  "allocated_subnet": "10.100.1.0/24",
  "ipam_pool": "cloudbridge-pool-1",
  "network_policies": [
    "allow-intra-tenant",
    "deny-cross-tenant"
  ],
  "allocated_ips": {
    "peer-001": "10.100.1.10",
    "peer-002": "10.100.1.11"
  },
  "created_at": "2025-10-02T18:00:00Z",
  "updated_at": "2025-10-02T18:30:00Z"
}
```

### **IPAMPool:**
```json
{
  "id": "cloudbridge-pool-1",
  "name": "CloudBridge Pool 1",
  "cidr": "10.0.0.0/8",
  "subnet_size": 24,
  "allocated": [
    "10.100.1.0/24",
    "10.100.2.0/24"
  ],
  "available": [
    "10.100.3.0/24",
    "10.100.4.0/24"
  ],
  "created_at": "2025-10-02T18:00:00Z",
  "updated_at": "2025-10-02T18:30:00Z"
}
```

## 🔄 **Workflow**

### **1. Создание нового тенанта:**
```
1. Zitadel создает тенанта
2. Автоматически выделяется подсеть из IPAM пула
3. Создается сетевая конфигурация
4. Применяются Network Policies
5. CloudBridge Relay получает уведомление
```

### **2. Регистрация нового пира:**
```
1. Пир запрашивает IP у CloudBridge Relay
2. Relay запрашивает конфигурацию у Zitadel
3. Zitadel выделяет IP в рамках подсети тенанта
4. Relay возвращает IP пиру
5. Zitadel обновляет конфигурацию
```

### **3. Синхронизация:**
```
1. CloudBridge Relay периодически синхронизируется с Zitadel
2. Получает обновленные конфигурации
3. Обновляет локальный кэш
4. Применяет изменения в Network Policies
```

## 🚀 **Implementation**

### **1. Zitadel Extensions:**

```go
// Добавить в Zitadel Management API
type TenantNetworkService struct {
    ipamPools map[string]*IPAMPool
    tenantConfigs map[string]*TenantNetworkConfig
}

func (s *TenantNetworkService) CreateTenantNetworkConfig(tenantID string, config *TenantNetworkConfig) error {
    // Выделить подсеть из IPAM пула
    subnet, err := s.allocateSubnetFromPool(config.IPAMPool)
    if err != nil {
        return err
    }
    
    config.AllocatedSubnet = subnet
    s.tenantConfigs[tenantID] = config
    
    return nil
}
```

### **2. CloudBridge Relay Integration:**

```go
// Обновленный PeerManager
type PeerManagerZitadel struct {
    zitadelClient *zitadel.TenantNetworkClient
    cache         *NetworkConfigCache
    logger        *zap.Logger
}

func (pm *PeerManagerZitadel) AllocateIPForTenant(tenantID, peerID string) (string, error) {
    // Получить конфигурацию из Zitadel
    config, err := pm.zitadelClient.GetTenantNetworkConfig(tenantID)
    if err != nil {
        return "", err
    }
    
    // Выделить IP в рамках подсети
    ip, err := pm.allocateIPInSubnet(config.AllocatedSubnet, peerID)
    if err != nil {
        return "", err
    }
    
    // Обновить конфигурацию в Zitadel
    return ip, pm.zitadelClient.AllocateIPForTenant(tenantID, peerID, ip)
}
```

## 🔒 **Security**

### **1. RBAC Integration:**
- Сетевая конфигурация защищена теми же правами доступа
- Аудит и логирование в единой системе
- JWT токены для аутентификации API

### **2. Network Isolation:**
- Calico Network Policies применяются автоматически
- Изоляция на уровне Kubernetes
- Мониторинг межтенантного трафика

## 📈 **Benefits**

### **1. Архитектурные:**
- ✅ **Единый источник истины** - все в Zitadel
- ✅ **Stateless CloudBridge Relay** - легко масштабируется
- ✅ **Централизованное управление** - простое администрирование

### **2. Операционные:**
- ✅ **Консистентность данных** - нет рассинхронизации
- ✅ **Упрощенное резервное копирование** - все в Zitadel
- ✅ **Единый аудит** - все действия в одном месте

### **3. Безопасность:**
- ✅ **Интегрированная авторизация** - RBAC из коробки
- ✅ **Централизованное управление ключами** - JWT токены
- ✅ **Аудит и мониторинг** - единая система

## 🎯 **Migration Plan**

### **Phase 1: Zitadel Extensions (1-2 дня)**
1. Добавить Tenant Network API в Zitadel
2. Создать IPAM Pool Management
3. Реализовать базовые CRUD операции

### **Phase 2: CloudBridge Integration (2-3 дня)**
1. Создать Zitadel Network Client
2. Обновить PeerManager для работы с Zitadel
3. Добавить кэширование конфигураций

### **Phase 3: Testing & Migration (1-2 дня)**
1. Протестировать интеграцию
2. Мигрировать существующие данные
3. Удалить локальное хранение

## 🏆 **Conclusion**

**Интеграция с Zitadel для управления подсетями тенантов обеспечивает:**

- ✅ **Архитектурную чистоту** - каждый компонент отвечает за свою область
- ✅ **Централизованное управление** - все в Zitadel
- ✅ **Лучшую безопасность** - единая система авторизации
- ✅ **Упрощение CloudBridge Relay** - становится stateless
- ✅ **Масштабируемость** - легко добавлять новые relay серверы

**Это правильный архитектурный подход для enterprise-grade решения!** 🚀
