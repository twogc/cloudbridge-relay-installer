# 🏢 Интеграция Zitadel + IPAM для изоляции компаний

## 🎯 **Архитектура: Компания = Тенант = Изолированная подсеть**

### **✅ Полное понимание архитектуры:**

```
┌─────────────────────────────────────────────────────────────┐
│                ZITADEL + CLOUDBRIDGE INTEGRATION           │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  🏢 Company A (Zitadel Organization)                       │
│     ├── Zitadel Org ID: org-company-a-12345                │
│     ├── Tenant ID: tenant-company-a-enterprise             │
│     ├── Domain: company-a.com                              │
│     ├── Users: user1@company-a.com, user2@company-a.com   │
│     ├── Подсеть: 10.66.19.0/24                            │
│     ├── Peers: peer-a-1, peer-a-2, peer-a-3               │
│     └── Изоляция: ✅ Полная изоляция от других компаний    │
│                                                             │
│  🏢 Company B (Zitadel Organization)                       │
│     ├── Zitadel Org ID: org-company-b-67890                │
│     ├── Tenant ID: tenant-company-b-enterprise             │
│     ├── Domain: company-b.com                              │
│     ├── Users: user1@company-b.com, user2@company-b.com   │
│     ├── Подсеть: 10.29.187.0/24                           │
│     ├── Peers: peer-b-1, peer-b-2, peer-b-3               │
│     └── Изоляция: ✅ Полная изоляция от других компаний    │
│                                                             │
│  🏢 Company C (Zitadel Organization)                       │
│     ├── Zitadel Org ID: org-company-c-11111                │
│     ├── Tenant ID: tenant-company-c-enterprise             │
│     ├── Domain: company-c.com                              │
│     ├── Users: user1@company-c.com, user2@company-c.com   │
│     ├── Подсеть: 10.58.33.0/24                            │
│     ├── Peers: peer-c-1, peer-c-2, peer-c-3               │
│     └── Изоляция: ✅ Полная изоляция от других компаний    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 **Процесс создания компании**

### **1. Создание организации в Zitadel:**

```go
// Создание новой компании через Zitadel
tenantReq := &zitadel.TenantCreateRequest{
    Name:        "Company A Enterprise",
    DisplayName: "Company A Enterprise",
    Domain:      "company-a.com",
    Quotas: &zitadel.TenantQuotas{
        MaxPeers:       50,
        MaxConnections: 1000,
        MaxTunnels:     200,
        BandwidthLimit: "1Gbps",
        StorageLimit:   "100GB",
    },
    Settings: &zitadel.TenantSettings{
        EnableP2P:          true,
        EnableMASQUE:       true,
        EnableICE:          true,
        AllowedProtocols:   []string{"quic", "tcp", "udp"},
        SecurityLevel:      "enterprise",
        LoggingLevel:       "info",
        SessionTimeout:     "8h",
        MaxSessionDuration: "24h",
    },
}

tenant, err := tenantManager.CreateTenant(ctx, tenantReq)
```

### **2. Автоматическое выделение подсети:**

```go
// После создания тенанта в Zitadel, автоматически выделяется подсеть
func (pm *PeerManager) CreateTenantWithSubnet(tenantID string) error {
    // 1. Получаем информацию о тенанте из Zitadel
    tenant, err := pm.zitadelClient.GetTenant(tenantID)
    if err != nil {
        return fmt.Errorf("failed to get tenant from Zitadel: %w", err)
    }
    
    // 2. Выделяем уникальную подсеть для компании
    subnet, err := pm.allocateTenantSubnet(tenantID)
    if err != nil {
        return fmt.Errorf("failed to allocate subnet for tenant: %w", err)
    }
    
    // 3. Сохраняем связь тенант -> подсеть
    pm.tenantSubnets[tenantID] = subnet
    
    // 4. Настраиваем сетевые политики для изоляции
    err = pm.setupTenantNetworkPolicies(tenantID, subnet)
    if err != nil {
        return fmt.Errorf("failed to setup network policies: %w", err)
    }
    
    pm.logger.Info("Tenant subnet allocated",
        zap.String("tenant_id", tenantID),
        zap.String("company_name", tenant.Name),
        zap.String("subnet", subnet))
    
    return nil
}
```

### **3. Детерминированное выделение подсетей:**

```go
func (pm *PeerManager) allocateTenantSubnet(tenantID string) (string, error) {
    // Используем хеш от tenantID для детерминированного выделения
    hash := sha256.Sum256([]byte(tenantID))
    subnetID := int(hash[0])<<8 + int(hash[1]) // 0-65535
    
    // Генерируем подсеть в диапазоне 10.0.0.0/8
    x := (subnetID / 256) % 256
    y := subnetID % 256
    
    subnet := fmt.Sprintf("10.%d.%d.0/24", x, y)
    
    // Проверяем, что подсеть не занята
    if pm.isSubnetAvailable(subnet) {
        return subnet, nil
    }
    
    // Если подсеть занята, ищем следующую доступную
    return pm.findNextAvailableSubnet(subnetID)
}
```

---

## 🔒 **Гарантии изоляции**

### **1. Сетевая изоляция:**

```yaml
# Calico Network Policy для изоляции тенантов
apiVersion: projectcalico.org/v3
kind: NetworkPolicy
metadata:
  name: tenant-isolation-company-a
  namespace: cloudbridge
spec:
  selector: tenant.cloudbridge.io/id == "tenant-company-a-enterprise"
  ingress:
  - action: Allow
    source:
      selector: tenant.cloudbridge.io/id == "tenant-company-a-enterprise"
  - action: Deny
    source: {}
  egress:
  - action: Allow
    destination:
      selector: tenant.cloudbridge.io/id == "tenant-company-a-enterprise"
  - action: Deny
    destination: {}
```

### **2. IPAM изоляция:**

```go
// Каждая компания получает свою подсеть
Company A: 10.66.19.0/24   -> Peers: 10.66.19.10, 10.66.19.11, 10.66.19.12
Company B: 10.29.187.0/24  -> Peers: 10.29.187.10, 10.29.187.11, 10.29.187.12
Company C: 10.58.33.0/24   -> Peers: 10.58.33.10, 10.58.33.11, 10.58.33.12
```

### **3. Аутентификация и авторизация:**

```go
// Проверка принадлежности пира к компании
func (pm *PeerManager) ValidatePeerTenant(peerID, tenantID string) error {
    // 1. Проверяем, что пир принадлежит тенанту
    if !pm.isPeerInTenant(peerID, tenantID) {
        return fmt.Errorf("peer %s does not belong to tenant %s", peerID, tenantID)
    }
    
    // 2. Проверяем квоты тенанта
    if err := pm.checkTenantQuotas(tenantID); err != nil {
        return fmt.Errorf("tenant quotas exceeded: %w", err)
    }
    
    // 3. Проверяем сетевые политики
    if err := pm.validateNetworkPolicies(peerID, tenantID); err != nil {
        return fmt.Errorf("network policy violation: %w", err)
    }
    
    return nil
}
```

---

## 🎯 **Примеры использования**

### **1. Создание новой компании:**

```bash
# 1. Создание организации в Zitadel
curl -X POST https://zitadel.company.com/management/v1/organizations \
  -H "Authorization: Bearer $ZITADEL_TOKEN" \
  -d '{
    "name": "Company A Enterprise",
    "displayName": "Company A Enterprise",
    "domain": "company-a.com"
  }'

# 2. Автоматическое выделение подсети
# CloudBridge автоматически выделяет подсеть 10.66.19.0/24

# 3. Регистрация пиров компании
curl -X POST https://cloudbridge.company.com/api/v1/tenants/tenant-company-a-enterprise/peers/register \
  -H "Authorization: Bearer $COMPANY_A_TOKEN" \
  -d '{
    "peer_id": "peer-a-1",
    "tenant_id": "tenant-company-a-enterprise",
    "public_key": "company-a-peer-1-key"
  }'
```

### **2. Проверка изоляции:**

```bash
# Компания A не может взаимодействовать с Компанией B
ping 10.29.187.10  # ❌ БЛОКИРОВАНО - разные подсети
ping 10.66.19.10   # ✅ РАЗРЕШЕНО - та же компания

# Проверка сетевых политик
kubectl get networkpolicies -n cloudbridge
```

---

## 🏆 **Заключение**

### **✅ Полная изоляция компаний обеспечена:**

1. **🏢 Zitadel Organization** = **Компания** = **Тенант**
2. **🔒 Уникальная подсеть** для каждой компании
3. **🛡️ Сетевые политики** для изоляции трафика
4. **🔐 Аутентификация** через Zitadel
5. **📊 Квоты и ограничения** на уровне компании
6. **🚀 Масштабируемость** до 65,536 компаний

**Каждая компания полностью изолирована от других компаний!** 🎯
