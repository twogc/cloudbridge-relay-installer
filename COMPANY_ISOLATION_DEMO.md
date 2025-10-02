# 🏢 Демонстрация изоляции компаний

## 🎯 **Сценарий: 3 компании используют CloudBridge**

### **Компания A: "TechCorp Solutions"**
- **Zitadel Org**: `org-techcorp-12345`
- **Tenant ID**: `tenant-techcorp-enterprise`
- **Домен**: `techcorp.com`
- **Пользователи**: `admin@techcorp.com`, `engineer@techcorp.com`
- **Подсеть**: `10.66.19.0/24`
- **Пиры**: `techcorp-peer-1`, `techcorp-peer-2`

### **Компания B: "FinanceBank Ltd"**
- **Zitadel Org**: `org-financebank-67890`
- **Tenant ID**: `tenant-financebank-enterprise`
- **Домен**: `financebank.com`
- **Пользователи**: `admin@financebank.com`, `trader@financebank.com`
- **Подсеть**: `10.29.187.0/24`
- **Пиры**: `financebank-peer-1`, `financebank-peer-2`

### **Компания C: "HealthCare Inc"**
- **Zitadel Org**: `org-healthcare-11111`
- **Tenant ID**: `tenant-healthcare-enterprise`
- **Домен**: `healthcare.com`
- **Пользователи**: `admin@healthcare.com`, `doctor@healthcare.com`
- **Подсеть**: `10.58.33.0/24`
- **Пиры**: `healthcare-peer-1`, `healthcare-peer-2`

---

## 🔒 **Гарантии изоляции**

### **✅ Каждая компания изолирована:**

```
┌─────────────────────────────────────────────────────────────┐
│                    COMPANY ISOLATION                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  🏢 TechCorp Solutions (10.66.19.0/24)                    │
│     ├── techcorp-peer-1: 10.66.19.10                       │
│     ├── techcorp-peer-2: 10.66.19.11                       │
│     └── ❌ НЕ МОЖЕТ взаимодействовать с другими компаниями │
│                                                             │
│  🏢 FinanceBank Ltd (10.29.187.0/24)                        │
│     ├── financebank-peer-1: 10.29.187.10                   │
│     ├── financebank-peer-2: 10.29.187.11                   │
│     └── ❌ НЕ МОЖЕТ взаимодействовать с другими компаниями │
│                                                             │
│  🏢 HealthCare Inc (10.58.33.0/24)                         │
│     ├── healthcare-peer-1: 10.58.33.10                     │
│     ├── healthcare-peer-2: 10.58.33.11                      │
│     └── ❌ НЕ МОЖЕТ взаимодействовать с другими компаниями │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🧪 **Тестирование изоляции**

### **1. Создание компаний:**

```bash
# Создание TechCorp в Zitadel
curl -X POST https://zitadel.company.com/management/v1/organizations \
  -H "Authorization: Bearer $ZITADEL_TOKEN" \
  -d '{
    "name": "TechCorp Solutions",
    "displayName": "TechCorp Solutions",
    "domain": "techcorp.com"
  }'

# Результат: Автоматически выделена подсеть 10.66.19.0/24
```

### **2. Регистрация пиров:**

```bash
# TechCorp регистрирует своих пиров
curl -X POST https://cloudbridge.company.com/api/v1/tenants/tenant-techcorp-enterprise/peers/register \
  -H "Authorization: Bearer $TECHCORP_TOKEN" \
  -d '{
    "peer_id": "techcorp-peer-1",
    "tenant_id": "tenant-techcorp-enterprise",
    "public_key": "techcorp-peer-1-key"
  }'

# Результат: Пир получил IP 10.66.19.10
```

### **3. Проверка изоляции:**

```bash
# TechCorp пытается подключиться к FinanceBank
ping 10.29.187.10  # ❌ БЛОКИРОВАНО - разные компании

# TechCorp подключается к своему пиру
ping 10.66.19.11   # ✅ РАЗРЕШЕНО - та же компания
```

---

## 🔧 **Техническая реализация**

### **1. Zitadel Integration:**

```go
// Создание тенанта в Zitadel
func (tm *TenantManager) CreateCompanyTenant(companyName, domain string) (*Tenant, error) {
    tenantReq := &TenantCreateRequest{
        Name:        companyName,
        DisplayName: companyName,
        Domain:      domain,
        Quotas: &TenantQuotas{
            MaxPeers:       100,
            MaxConnections: 1000,
            MaxTunnels:     200,
            BandwidthLimit: "1Gbps",
            StorageLimit:   "100GB",
        },
        Settings: &TenantSettings{
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
    
    return tm.CreateTenant(ctx, tenantReq)
}
```

### **2. IPAM Integration:**

```go
// Автоматическое выделение подсети для компании
func (pm *PeerManager) AllocateCompanySubnet(tenantID string) (string, error) {
    // Получаем информацию о компании из Zitadel
    tenant, err := pm.zitadelClient.GetTenant(tenantID)
    if err != nil {
        return "", fmt.Errorf("failed to get tenant from Zitadel: %w", err)
    }
    
    // Выделяем уникальную подсеть
    subnet, err := pm.allocateTenantSubnet(tenantID)
    if err != nil {
        return "", fmt.Errorf("failed to allocate subnet: %w", err)
    }
    
    // Настраиваем сетевые политики
    err = pm.setupCompanyNetworkPolicies(tenantID, subnet)
    if err != nil {
        return "", fmt.Errorf("failed to setup network policies: %w", err)
    }
    
    pm.logger.Info("Company subnet allocated",
        zap.String("company", tenant.Name),
        zap.String("tenant_id", tenantID),
        zap.String("subnet", subnet))
    
    return subnet, nil
}
```

### **3. Network Policies:**

```yaml
# Calico Network Policy для TechCorp
apiVersion: projectcalico.org/v3
kind: NetworkPolicy
metadata:
  name: techcorp-isolation
  namespace: cloudbridge
spec:
  selector: tenant.cloudbridge.io/id == "tenant-techcorp-enterprise"
  ingress:
  - action: Allow
    source:
      selector: tenant.cloudbridge.io/id == "tenant-techcorp-enterprise"
  - action: Deny
    source: {}
  egress:
  - action: Allow
    destination:
      selector: tenant.cloudbridge.io/id == "tenant-techcorp-enterprise"
  - action: Deny
    destination: {}
```

---

## 🏆 **Результат**

### **✅ Полная изоляция компаний обеспечена:**

1. **🏢 Каждая компания** = **Zitadel Organization** = **Уникальный тенант**
2. **🔒 Уникальная подсеть** для каждой компании
3. **🛡️ Сетевые политики** блокируют межкомпанийное взаимодействие
4. **🔐 Аутентификация** через Zitadel для каждой компании
5. **📊 Квоты и ограничения** на уровне компании
6. **🚀 Масштабируемость** до 65,536 компаний

**Каждая компания полностью изолирована от других компаний!** 🎯

### **🔒 Гарантии безопасности:**
- ✅ **Сетевая изоляция** - разные подсети
- ✅ **Политики безопасности** - блокировка межкомпанийного трафика
- ✅ **Аутентификация** - только свои пользователи
- ✅ **Авторизация** - доступ только к своим ресурсам
- ✅ **Мониторинг** - логирование всех попыток доступа
