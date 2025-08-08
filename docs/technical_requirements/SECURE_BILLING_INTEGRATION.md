# 🔒 Безопасная интеграция с внешним биллингом через Cadence Workflow

## 📋 Проблема безопасности

**Внешний биллинг API требует передачи данных, но содержит конфиденциальную информацию:**
- Персональные данные пользователей
- Финансовая информация
- Детали подписок и тарифов
- История платежей
- Корпоративные данные

**Риски:**
- Утечка персональных данных (GDPR/152-ФЗ)
- Компрометация финансовой информации
- Нарушение коммерческой тайны
- Юридическая ответственность

## 🛡️ Решение: Безопасная архитектура интеграции

### **Принципы безопасности:**

```
┌─────────────────────────────────────────────────────────────────┐
│                    Secure Billing Integration                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   CloudBridge│    │   Data      │    │   External  │         │
│  │   Relay      │    │   Masking   │    │   Billing   │         │
│  │   (Internal) │    │   Layer     │    │   API       │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
│         │                   │                   │               │
│         │                   │                   │               │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   Cadence   │    │   Encryption│    │   Rate      │         │
│  │   Workflow  │    │   Gateway   │    │   Limiting  │         │
│  └─────────────┘    └─────────────┘    └─────────────┘         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 🔐 Архитектура безопасности

### **1. Слой маскирования данных (Data Masking Layer)**

```go
// internal/billing/masking/data_masking.go
package masking

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "fmt"
)

type DataMaskingService struct {
    encryptionKey []byte
    salt          []byte
}

type MaskedUser struct {
    ID           string `json:"id"`
    Email        string `json:"email"`        // Маскированный
    Phone        string `json:"phone"`        // Маскированный
    Name         string `json:"name"`         // Маскированный
    Organization string `json:"organization"` // Маскированный
    InternalID   string `json:"internal_id"`  // Внутренний ID
}

type MaskedSubscription struct {
    ID          string  `json:"id"`
    UserID      string  `json:"user_id"`      // Маскированный
    TariffID    string  `json:"tariff_id"`
    Amount      float64 `json:"amount"`
    Status      string  `json:"status"`
    InternalID  string  `json:"internal_id"`  // Внутренний ID
}

func (dms *DataMaskingService) MaskUser(user *User) *MaskedUser {
    return &MaskedUser{
        ID:           dms.generateExternalID(user.ID),
        Email:        dms.maskEmail(user.Email),
        Phone:        dms.maskPhone(user.Phone),
        Name:         dms.maskName(user.Name),
        Organization: dms.maskOrganization(user.Organization),
        InternalID:   user.ID,
    }
}

func (dms *DataMaskingService) maskEmail(email string) string {
    if email == "" {
        return ""
    }
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***@***"
    }
    username := parts[0]
    domain := parts[1]
    
    if len(username) <= 2 {
        return "***@" + domain
    }
    
    maskedUsername := username[:1] + "***" + username[len(username)-1:]
    return maskedUsername + "@" + domain
}

func (dms *DataMaskingService) maskPhone(phone string) string {
    if len(phone) < 4 {
        return "***"
    }
    return phone[:2] + "***" + phone[len(phone)-2:]
}

func (dms *DataMaskingService) maskName(name string) string {
    if len(name) <= 2 {
        return "***"
    }
    return name[:1] + "***" + name[len(name)-1:]
}

func (dms *DataMaskingService) generateExternalID(internalID string) string {
    // Генерируем уникальный внешний ID
    hash := sha256.Sum256([]byte(internalID + dms.salt))
    return base64.URLEncoding.EncodeToString(hash[:])[:16]
}
```

### **2. Шифрование данных (Encryption Gateway)**

```go
// internal/billing/encryption/encryption_gateway.go
package encryption

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "fmt"
)

type EncryptionGateway struct {
    key []byte
}

type EncryptedPayload struct {
    Data      string `json:"data"`
    IV        string `json:"iv"`
    Signature string `json:"signature"`
}

func (eg *EncryptionGateway) EncryptSensitiveData(data interface{}) (*EncryptedPayload, error) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal data: %w", err)
    }
    
    // Генерируем IV
    iv := make([]byte, aes.BlockSize)
    if _, err := rand.Read(iv); err != nil {
        return nil, fmt.Errorf("failed to generate IV: %w", err)
    }
    
    // Шифруем данные
    block, err := aes.NewCipher(eg.key)
    if err != nil {
        return nil, fmt.Errorf("failed to create cipher: %w", err)
    }
    
    ciphertext := make([]byte, len(jsonData))
    stream := cipher.NewCFBEncrypter(block, iv)
    stream.XORKeyStream(ciphertext, jsonData)
    
    // Создаем подпись
    signature := eg.createSignature(ciphertext, iv)
    
    return &EncryptedPayload{
        Data:      base64.StdEncoding.EncodeToString(ciphertext),
        IV:        base64.StdEncoding.EncodeToString(iv),
        Signature: base64.StdEncoding.EncodeToString(signature),
    }, nil
}

func (eg *EncryptionGateway) createSignature(data []byte, iv []byte) []byte {
    // Создаем HMAC подпись для проверки целостности
    h := hmac.New(sha256.New, eg.key)
    h.Write(data)
    h.Write(iv)
    return h.Sum(nil)
}
```

### **3. Cadence Workflow для безопасной интеграции**

```go
// internal/cadence/billing/external_billing_workflow.go
package billing

import (
    "time"
    "go.uber.org/cadence/workflow"
    "go.uber.org/cadence/activity"
)

type ExternalBillingWorkflow struct {
    maskingService    *masking.DataMaskingService
    encryptionGateway *encryption.EncryptionGateway
    apiClient         *ExternalBillingAPIClient
}

// Workflow для создания пользователя во внешнем биллинге
func (ebw *ExternalBillingWorkflow) CreateUserWorkflow(ctx workflow.Context, user *User) error {
    // 1. Маскируем данные пользователя
    maskedUser := ebw.maskingService.MaskUser(user)
    
    // 2. Шифруем чувствительные данные
    encryptedData, err := ebw.encryptionGateway.EncryptSensitiveData(maskedUser)
    if err != nil {
        return fmt.Errorf("failed to encrypt user data: %w", err)
    }
    
    // 3. Отправляем во внешний API
    var result *APIResponse
    err = workflow.ExecuteActivity(ctx, ebw.apiClient.CreateUser, encryptedData).Get(ctx, &result)
    if err != nil {
        return fmt.Errorf("failed to create user in external billing: %w", err)
    }
    
    // 4. Сохраняем маппинг ID
    err = workflow.ExecuteActivity(ctx, ebw.saveIDMapping, user.ID, result.ExternalID).Get(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to save ID mapping: %w", err)
    }
    
    return nil
}

// Workflow для синхронизации подписок
func (ebw *ExternalBillingWorkflow) SyncSubscriptionsWorkflow(ctx workflow.Context) error {
    // 1. Получаем все активные подписки
    var subscriptions []*Subscription
    err := workflow.ExecuteActivity(ctx, ebw.getActiveSubscriptions).Get(ctx, &subscriptions)
    if err != nil {
        return fmt.Errorf("failed to get subscriptions: %w", err)
    }
    
    // 2. Обрабатываем каждую подписку
    for _, sub := range subscriptions {
        // Маскируем данные
        maskedSub := ebw.maskingService.MaskSubscription(sub)
        
        // Шифруем
        encryptedData, err := ebw.encryptionGateway.EncryptSensitiveData(maskedSub)
        if err != nil {
            workflow.GetLogger(ctx).Error("Failed to encrypt subscription data", 
                "subscription_id", sub.ID, "error", err)
            continue
        }
        
        // Отправляем во внешний API
        err = workflow.ExecuteActivity(ctx, ebw.apiClient.UpdateSubscription, encryptedData).Get(ctx, nil)
        if err != nil {
            workflow.GetLogger(ctx).Error("Failed to sync subscription", 
                "subscription_id", sub.ID, "error", err)
            continue
        }
        
        // Добавляем задержку для rate limiting
        workflow.Sleep(ctx, time.Second*2)
    }
    
    return nil
}

// Workflow для обработки платежей
func (ebw *ExternalBillingWorkflow) ProcessPaymentWorkflow(ctx workflow.Context, payment *Payment) error {
    // 1. Проверяем авторизацию
    var authToken string
    err := workflow.ExecuteActivity(ctx, ebw.apiClient.GetAuthToken).Get(ctx, &authToken)
    if err != nil {
        return fmt.Errorf("failed to get auth token: %w", err)
    }
    
    // 2. Маскируем данные платежа
    maskedPayment := ebw.maskingService.MaskPayment(payment)
    
    // 3. Шифруем
    encryptedData, err := ebw.encryptionGateway.EncryptSensitiveData(maskedPayment)
    if err != nil {
        return fmt.Errorf("failed to encrypt payment data: %w", err)
    }
    
    // 4. Отправляем платеж
    var result *PaymentResponse
    err = workflow.ExecuteActivity(ctx, ebw.apiClient.ProcessPayment, authToken, encryptedData).Get(ctx, &result)
    if err != nil {
        return fmt.Errorf("failed to process payment: %w", err)
    }
    
    // 5. Обновляем статус в локальной БД
    err = workflow.ExecuteActivity(ctx, ebw.updatePaymentStatus, payment.ID, result.Status).Get(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to update payment status: %w", err)
    }
    
    return nil
}
```

### **4. API клиент с rate limiting и retry логикой**

```go
// internal/billing/api/external_billing_client.go
package api

import (
    "context"
    "time"
    "net/http"
    "golang.org/x/time/rate"
)

type ExternalBillingAPIClient struct {
    baseURL     string
    httpClient  *http.Client
    rateLimiter *rate.Limiter
    retryConfig *RetryConfig
}

type RetryConfig struct {
    MaxRetries    int
    BaseDelay     time.Duration
    MaxDelay      time.Duration
    BackoffFactor float64
}

func (client *ExternalBillingAPIClient) CreateUser(ctx context.Context, encryptedData *EncryptedPayload) (*APIResponse, error) {
    // Rate limiting
    err := client.rateLimiter.Wait(ctx)
    if err != nil {
        return nil, fmt.Errorf("rate limit exceeded: %w", err)
    }
    
    // Retry logic
    var response *APIResponse
    err = client.retryWithBackoff(ctx, func() error {
        resp, err := client.makeRequest(ctx, "POST", "/users", encryptedData)
        if err != nil {
            return err
        }
        response = resp
        return nil
    })
    
    return response, err
}

func (client *ExternalBillingAPIClient) retryWithBackoff(ctx context.Context, operation func() error) error {
    var lastErr error
    
    for attempt := 0; attempt <= client.retryConfig.MaxRetries; attempt++ {
        if err := operation(); err == nil {
            return nil
        } else {
            lastErr = err
        }
        
        if attempt < client.retryConfig.MaxRetries {
            delay := time.Duration(float64(client.retryConfig.BaseDelay) * 
                math.Pow(client.retryConfig.BackoffFactor, float64(attempt)))
            
            if delay > client.retryConfig.MaxDelay {
                delay = client.retryConfig.MaxDelay
            }
            
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(delay):
                continue
            }
        }
    }
    
    return fmt.Errorf("operation failed after %d retries: %w", client.retryConfig.MaxRetries, lastErr)
}
```

## 🔧 Конфигурация безопасности

### **configs/secure_billing.yaml**

```yaml
billing:
  external:
    enabled: true
    base_url: "https://external-billing-api.com"
    
    # Безопасность
    security:
      encryption_key: "${ENCRYPTION_KEY}"  # Из переменной окружения
      salt: "${ENCRYPTION_SALT}"           # Из переменной окружения
      data_retention_days: 30              # Хранение зашифрованных данных
      
    # Rate limiting
    rate_limiting:
      requests_per_second: 10
      burst_size: 20
      
    # Retry configuration
    retry:
      max_retries: 3
      base_delay: "1s"
      max_delay: "30s"
      backoff_factor: 2.0
      
    # Data masking
    masking:
      email_mask_pattern: "***@***"
      phone_mask_pattern: "***"
      name_mask_pattern: "***"
      
    # API endpoints mapping
    endpoints:
      users: "/users"
      subscriptions: "/subscriptions"
      payments: "/payments"
      agreements: "/agreements"
      tariffs: "/tariffs"
      
    # Audit logging
    audit:
      enabled: true
      log_level: "info"
      sensitive_fields:
        - "email"
        - "phone"
        - "name"
        - "organization"
        - "payment_details"
```

## 📊 Мониторинг и аудит

### **Метрики безопасности**

```go
// internal/metrics/secure_billing.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Метрики шифрования
    EncryptionOperationsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "billing_encryption_operations_total",
            Help: "Total number of encryption operations",
        },
        []string{"operation", "status"},
    )
    
    // Метрики маскирования
    DataMaskingOperationsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "billing_data_masking_operations_total",
            Help: "Total number of data masking operations",
        },
        []string{"data_type", "status"},
    )
    
    // Метрики API вызовов
    ExternalAPICallsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "billing_external_api_calls_total",
            Help: "Total number of external API calls",
        },
        []string{"endpoint", "method", "status"},
    )
    
    // Метрики rate limiting
    RateLimitHitsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "billing_rate_limit_hits_total",
            Help: "Total number of rate limit hits",
        },
        []string{"endpoint"},
    )
)
```

## 🚀 Преимущества архитектуры

### **1. Безопасность данных:**
- ✅ Маскирование персональных данных
- ✅ Шифрование чувствительной информации
- ✅ Rate limiting для защиты от DDoS
- ✅ Аудит всех операций

### **2. Надежность:**
- ✅ Retry логика с экспоненциальной задержкой
- ✅ Circuit breaker для защиты от сбоев
- ✅ Graceful degradation при недоступности API

### **3. Масштабируемость:**
- ✅ Cadence Workflow для оркестрации
- ✅ Асинхронная обработка
- ✅ Batch операции для оптимизации

### **4. Соответствие требованиям:**
- ✅ GDPR/152-ФЗ compliance
- ✅ Аудит всех операций
- ✅ Контроль доступа к данным

## 📋 План внедрения

### **Этап 1: Базовая безопасность (1-2 недели)**
1. Реализация Data Masking Layer
2. Настройка Encryption Gateway
3. Базовые Cadence Workflows

### **Этап 2: API интеграция (2-3 недели)**
1. External Billing API Client
2. Rate limiting и retry логика
3. Обработка ошибок

### **Этап 3: Мониторинг и аудит (1 неделя)**
1. Метрики безопасности
2. Audit logging
3. Алерты и уведомления

### **Этап 4: Тестирование и оптимизация (1-2 недели)**
1. Penetration testing
2. Load testing
3. Оптимизация производительности

## 🎯 Результат

**Безопасная интеграция с внешним биллингом, которая:**
- Защищает конфиденциальные данные пользователей
- Обеспечивает надежную обработку биллинговых операций
- Соответствует требованиям регуляторов
- Масштабируется с ростом нагрузки
- Предоставляет полный аудит всех операций

**Cadence Workflow обеспечивает надежную оркестрацию всех процессов! 🔒**
