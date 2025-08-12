# 👤 Пример: Cadence Workflow для работы с пользователями

## 📋 Описание

Реализация безопасного Cadence Workflow для синхронизации пользователей с внешним биллинг API с полной защитой конфиденциальных данных.

## 🔐 Архитектура безопасности

```
┌─────────────────────────────────────────────────────────────────┐
│                    User Sync Workflow                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐         │
│  │   Internal  │    │   Data      │    │   External  │         │
│  │   User DB   │    │   Masking   │    │   Billing   │         │
│  │             │    │   Service   │    │   API       │         │
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

## 💻 Реализация

### **1. Структуры данных**

```go
// internal/billing/models/user.go
package models

import (
    "time"
    "encoding/json"
)

// Внутренняя модель пользователя
type User struct {
    ID           string                 `json:"id" db:"id"`
    Email        string                 `json:"email" db:"email"`
    Phone        string                 `json:"phone" db:"phone"`
    Name         string                 `json:"name" db:"name"`
    Organization string                 `json:"organization" db:"organization"`
    Status       string                 `json:"status" db:"status"`
    CreatedAt    time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
    Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
}

// Маскированная модель для внешнего API
type MaskedUser struct {
    ID           string                 `json:"id"`
    Email        string                 `json:"email"`        // Маскированный
    Phone        string                 `json:"phone"`        // Маскированный
    Name         string                 `json:"name"`         // Маскированный
    Organization string                 `json:"organization"` // Маскированный
    Status       string                 `json:"status"`
    CreatedAt    time.Time              `json:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
    InternalID   string                 `json:"internal_id"`  // Внутренний ID
    ExternalID   string                 `json:"external_id,omitempty"`
}

// Ответ от внешнего API
type ExternalUserResponse struct {
    Success bool   `json:"success"`
    ID      string `json:"id"`
    Message string `json:"message,omitempty"`
    Error   string `json:"error,omitempty"`
}

// Статус синхронизации
type SyncStatus struct {
    UserID       string    `json:"user_id"`
    ExternalID   string    `json:"external_id"`
    Status       string    `json:"status"` // pending, synced, failed
    LastSyncAt   time.Time `json:"last_sync_at"`
    ErrorMessage string    `json:"error_message,omitempty"`
    RetryCount   int       `json:"retry_count"`
}
```

### **2. Сервис маскирования данных**

```go
// internal/billing/masking/user_masking.go
package masking

import (
    "crypto/sha256"
    "encoding/base64"
    "strings"
    "regexp"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/billing/models"
)

type UserMaskingService struct {
    salt []byte
}

func NewUserMaskingService(salt []byte) *UserMaskingService {
    return &UserMaskingService{
        salt: salt,
    }
}

// Маскирование пользователя
func (ums *UserMaskingService) MaskUser(user *models.User) *models.MaskedUser {
    return &models.MaskedUser{
        ID:           ums.generateExternalID(user.ID),
        Email:        ums.maskEmail(user.Email),
        Phone:        ums.maskPhone(user.Phone),
        Name:         ums.maskName(user.Name),
        Organization: ums.maskOrganization(user.Organization),
        Status:       user.Status,
        CreatedAt:    user.CreatedAt,
        UpdatedAt:    user.UpdatedAt,
        InternalID:   user.ID,
    }
}

// Маскирование email
func (ums *UserMaskingService) maskEmail(email string) string {
    if email == "" {
        return ""
    }
    
    // Проверяем формат email
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(email) {
        return "***@***"
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
    
    // Маскируем username, оставляя первый и последний символ
    maskedUsername := username[:1] + "***" + username[len(username)-1:]
    return maskedUsername + "@" + domain
}

// Маскирование телефона
func (ums *UserMaskingService) maskPhone(phone string) string {
    if phone == "" {
        return ""
    }
    
    // Убираем все нецифровые символы
    digits := regexp.MustCompile(`[^0-9]`).ReplaceAllString(phone, "")
    
    if len(digits) < 4 {
        return "***"
    }
    
    // Маскируем середину номера
    return digits[:2] + "***" + digits[len(digits)-2:]
}

// Маскирование имени
func (ums *UserMaskingService) maskName(name string) string {
    if name == "" {
        return ""
    }
    
    if len(name) <= 2 {
        return "***"
    }
    
    // Маскируем имя, оставляя первую и последнюю букву
    return name[:1] + "***" + name[len(name)-1:]
}

// Маскирование организации
func (ums *UserMaskingService) maskOrganization(org string) string {
    if org == "" {
        return ""
    }
    
    if len(org) <= 3 {
        return "***"
    }
    
    // Маскируем название организации
    return org[:2] + "***" + org[len(org)-2:]
}

// Генерация внешнего ID
func (ums *UserMaskingService) generateExternalID(internalID string) string {
    hash := sha256.Sum256([]byte(internalID + string(ums.salt)))
    return base64.URLEncoding.EncodeToString(hash[:])[:16]
}
```

### **3. Cadence Workflow для синхронизации пользователей**

```go
// internal/cadence/billing/user_sync_workflow.go
package billing

import (
    "time"
    "fmt"
    "context"
    
    "go.uber.org/cadence/workflow"
    "go.uber.org/cadence/activity"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/billing/models"
    "github.com/twogc/cloudbridge-relay-installer/internal/billing/masking"
    "github.com/twogc/cloudbridge-relay-installer/internal/billing/encryption"
    "github.com/twogc/cloudbridge-relay-installer/internal/billing/api"
)

type UserSyncWorkflow struct {
    maskingService    *masking.UserMaskingService
    encryptionGateway *encryption.EncryptionGateway
    apiClient         *api.ExternalBillingAPIClient
    userRepository    UserRepository
    syncRepository    SyncRepository
}

// Основной Workflow для синхронизации пользователей
func (usw *UserSyncWorkflow) SyncUsersWorkflow(ctx workflow.Context, batchSize int) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting user sync workflow", "batch_size", batchSize)
    
    // 1. Получаем пользователей для синхронизации
    var users []*models.User
    err := workflow.ExecuteActivity(ctx, usw.userRepository.GetUsersForSync, batchSize).Get(ctx, &users)
    if err != nil {
        return fmt.Errorf("failed to get users for sync: %w", err)
    }
    
    logger.Info("Found users for sync", "count", len(users))
    
    // 2. Обрабатываем каждого пользователя
    for i, user := range users {
        // Проверяем контекст на отмену
        if ctx.Err() != nil {
            return ctx.Err()
        }
        
        // Обрабатываем пользователя
        err := workflow.ExecuteActivity(ctx, usw.processUser, user).Get(ctx, nil)
        if err != nil {
            logger.Error("Failed to process user", 
                "user_id", user.ID, 
                "error", err)
            
            // Продолжаем с следующим пользователем
            continue
        }
        
        // Добавляем задержку для rate limiting
        if i < len(users)-1 {
            workflow.Sleep(ctx, time.Second*2)
        }
    }
    
    logger.Info("User sync workflow completed", "processed_count", len(users))
    return nil
}

// Workflow для создания нового пользователя
func (usw *UserSyncWorkflow) CreateUserWorkflow(ctx workflow.Context, user *models.User) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting create user workflow", "user_id", user.ID)
    
    // 1. Маскируем данные пользователя
    maskedUser := usw.maskingService.MaskUser(user)
    
    // 2. Шифруем данные
    encryptedData, err := usw.encryptionGateway.EncryptSensitiveData(maskedUser)
    if err != nil {
        return fmt.Errorf("failed to encrypt user data: %w", err)
    }
    
    // 3. Отправляем во внешний API
    var result *models.ExternalUserResponse
    err = workflow.ExecuteActivity(ctx, usw.apiClient.CreateUser, encryptedData).Get(ctx, &result)
    if err != nil {
        return fmt.Errorf("failed to create user in external billing: %w", err)
    }
    
    if !result.Success {
        return fmt.Errorf("external API returned error: %s", result.Error)
    }
    
    // 4. Сохраняем маппинг ID
    syncStatus := &models.SyncStatus{
        UserID:       user.ID,
        ExternalID:   result.ID,
        Status:       "synced",
        LastSyncAt:   time.Now(),
        RetryCount:   0,
    }
    
    err = workflow.ExecuteActivity(ctx, usw.syncRepository.SaveSyncStatus, syncStatus).Get(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to save sync status: %w", err)
    }
    
    logger.Info("User created successfully", 
        "user_id", user.ID, 
        "external_id", result.ID)
    
    return nil
}

// Workflow для обновления пользователя
func (usw *UserSyncWorkflow) UpdateUserWorkflow(ctx workflow.Context, user *models.User) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting update user workflow", "user_id", user.ID)
    
    // 1. Получаем внешний ID пользователя
    var syncStatus *models.SyncStatus
    err := workflow.ExecuteActivity(ctx, usw.syncRepository.GetSyncStatus, user.ID).Get(ctx, &syncStatus)
    if err != nil {
        return fmt.Errorf("failed to get sync status: %w", err)
    }
    
    if syncStatus == nil || syncStatus.ExternalID == "" {
        // Пользователь не синхронизирован, создаем его
        return usw.CreateUserWorkflow(ctx, user)
    }
    
    // 2. Маскируем данные пользователя
    maskedUser := usw.maskingService.MaskUser(user)
    maskedUser.ExternalID = syncStatus.ExternalID
    
    // 3. Шифруем данные
    encryptedData, err := usw.encryptionGateway.EncryptSensitiveData(maskedUser)
    if err != nil {
        return fmt.Errorf("failed to encrypt user data: %w", err)
    }
    
    // 4. Отправляем обновление во внешний API
    var result *models.ExternalUserResponse
    err = workflow.ExecuteActivity(ctx, usw.apiClient.UpdateUser, syncStatus.ExternalID, encryptedData).Get(ctx, &result)
    if err != nil {
        return fmt.Errorf("failed to update user in external billing: %w", err)
    }
    
    if !result.Success {
        return fmt.Errorf("external API returned error: %s", result.Error)
    }
    
    // 5. Обновляем статус синхронизации
    syncStatus.Status = "synced"
    syncStatus.LastSyncAt = time.Now()
    
    err = workflow.ExecuteActivity(ctx, usw.syncRepository.UpdateSyncStatus, syncStatus).Get(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to update sync status: %w", err)
    }
    
    logger.Info("User updated successfully", 
        "user_id", user.ID, 
        "external_id", syncStatus.ExternalID)
    
    return nil
}

// Workflow для удаления пользователя
func (usw *UserSyncWorkflow) DeleteUserWorkflow(ctx workflow.Context, userID string) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting delete user workflow", "user_id", userID)
    
    // 1. Получаем внешний ID пользователя
    var syncStatus *models.SyncStatus
    err := workflow.ExecuteActivity(ctx, usw.syncRepository.GetSyncStatus, userID).Get(ctx, &syncStatus)
    if err != nil {
        return fmt.Errorf("failed to get sync status: %w", err)
    }
    
    if syncStatus == nil || syncStatus.ExternalID == "" {
        logger.Info("User not synced, skipping deletion", "user_id", userID)
        return nil
    }
    
    // 2. Удаляем из внешнего API
    var result *models.ExternalUserResponse
    err = workflow.ExecuteActivity(ctx, usw.apiClient.DeleteUser, syncStatus.ExternalID).Get(ctx, &result)
    if err != nil {
        return fmt.Errorf("failed to delete user from external billing: %w", err)
    }
    
    if !result.Success {
        return fmt.Errorf("external API returned error: %s", result.Error)
    }
    
    // 3. Удаляем статус синхронизации
    err = workflow.ExecuteActivity(ctx, usw.syncRepository.DeleteSyncStatus, userID).Get(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to delete sync status: %w", err)
    }
    
    logger.Info("User deleted successfully", 
        "user_id", userID, 
        "external_id", syncStatus.ExternalID)
    
    return nil
}
```

### **4. Activity для обработки пользователя**

```go
// internal/cadence/billing/user_activities.go
package billing

import (
    "context"
    "time"
    
    "go.uber.org/cadence/activity"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/billing/models"
)

// Activity для обработки одного пользователя
func (usw *UserSyncWorkflow) processUser(ctx context.Context, user *models.User) error {
    logger := activity.GetLogger(ctx)
    logger.Info("Processing user", "user_id", user.ID)
    
    // 1. Проверяем статус синхронизации
    syncStatus, err := usw.syncRepository.GetSyncStatus(ctx, user.ID)
    if err != nil {
        return err
    }
    
    // 2. Определяем действие
    var action string
    if syncStatus == nil {
        action = "create"
    } else if user.Status == "deleted" {
        action = "delete"
    } else {
        action = "update"
    }
    
    // 3. Выполняем действие
    switch action {
    case "create":
        return usw.createUserActivity(ctx, user)
    case "update":
        return usw.updateUserActivity(ctx, user, syncStatus)
    case "delete":
        return usw.deleteUserActivity(ctx, user.ID, syncStatus)
    default:
        return fmt.Errorf("unknown action: %s", action)
    }
}

// Activity для создания пользователя
func (usw *UserSyncWorkflow) createUserActivity(ctx context.Context, user *models.User) error {
    logger := activity.GetLogger(ctx)
    
    // 1. Маскируем данные
    maskedUser := usw.maskingService.MaskUser(user)
    
    // 2. Шифруем
    encryptedData, err := usw.encryptionGateway.EncryptSensitiveData(maskedUser)
    if err != nil {
        return err
    }
    
    // 3. Отправляем в API
    result, err := usw.apiClient.CreateUser(ctx, encryptedData)
    if err != nil {
        // Сохраняем статус с ошибкой
        syncStatus := &models.SyncStatus{
            UserID:       user.ID,
            Status:       "failed",
            LastSyncAt:   time.Now(),
            ErrorMessage: err.Error(),
            RetryCount:   1,
        }
        usw.syncRepository.SaveSyncStatus(ctx, syncStatus)
        return err
    }
    
    // 4. Сохраняем успешный статус
    syncStatus := &models.SyncStatus{
        UserID:       user.ID,
        ExternalID:   result.ID,
        Status:       "synced",
        LastSyncAt:   time.Now(),
        RetryCount:   0,
    }
    
    return usw.syncRepository.SaveSyncStatus(ctx, syncStatus)
}

// Activity для обновления пользователя
func (usw *UserSyncWorkflow) updateUserActivity(ctx context.Context, user *models.User, syncStatus *models.SyncStatus) error {
    logger := activity.GetLogger(ctx)
    
    // 1. Маскируем данные
    maskedUser := usw.maskingService.MaskUser(user)
    maskedUser.ExternalID = syncStatus.ExternalID
    
    // 2. Шифруем
    encryptedData, err := usw.encryptionGateway.EncryptSensitiveData(maskedUser)
    if err != nil {
        return err
    }
    
    // 3. Отправляем в API
    result, err := usw.apiClient.UpdateUser(ctx, syncStatus.ExternalID, encryptedData)
    if err != nil {
        // Обновляем статус с ошибкой
        syncStatus.Status = "failed"
        syncStatus.LastSyncAt = time.Now()
        syncStatus.ErrorMessage = err.Error()
        syncStatus.RetryCount++
        usw.syncRepository.UpdateSyncStatus(ctx, syncStatus)
        return err
    }
    
    // 4. Обновляем успешный статус
    syncStatus.Status = "synced"
    syncStatus.LastSyncAt = time.Now()
    syncStatus.ErrorMessage = ""
    
    return usw.syncRepository.UpdateSyncStatus(ctx, syncStatus)
}

// Activity для удаления пользователя
func (usw *UserSyncWorkflow) deleteUserActivity(ctx context.Context, userID string, syncStatus *models.SyncStatus) error {
    logger := activity.GetLogger(ctx)
    
    // 1. Удаляем из API
    result, err := usw.apiClient.DeleteUser(ctx, syncStatus.ExternalID)
    if err != nil {
        return err
    }
    
    // 2. Удаляем статус синхронизации
    return usw.syncRepository.DeleteSyncStatus(ctx, userID)
}
```

## 🔧 Конфигурация

### **configs/user_sync_workflow.yaml**

```yaml
cadence:
  workflows:
    user_sync:
      # Основная синхронизация
      sync_users:
        enabled: true
        schedule: "0 */6 * * *"  # Каждые 6 часов
        batch_size: 100
        timeout: "30m"
        retry_policy:
          initial_interval: "1s"
          maximum_interval: "1m"
          maximum_attempts: 3
          backoff_coefficient: 2.0
          
      # Создание пользователя
      create_user:
        enabled: true
        timeout: "5m"
        retry_policy:
          initial_interval: "1s"
          maximum_interval: "30s"
          maximum_attempts: 3
          backoff_coefficient: 2.0
          
      # Обновление пользователя
      update_user:
        enabled: true
        timeout: "5m"
        retry_policy:
          initial_interval: "1s"
          maximum_interval: "30s"
          maximum_attempts: 3
          backoff_coefficient: 2.0
          
      # Удаление пользователя
      delete_user:
        enabled: true
        timeout: "3m"
        retry_policy:
          initial_interval: "1s"
          maximum_interval: "30s"
          maximum_attempts: 3
          backoff_coefficient: 2.0
```

## 📊 Метрики и мониторинг

```go
// internal/metrics/user_sync_metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Метрики синхронизации пользователей
    UserSyncOperationsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "user_sync_operations_total",
            Help: "Total number of user sync operations",
        },
        []string{"operation", "status"},
    )
    
    // Метрики маскирования данных
    UserDataMaskingTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "user_data_masking_total",
            Help: "Total number of user data masking operations",
        },
        []string{"data_type", "status"},
    )
    
    // Метрики времени синхронизации
    UserSyncDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "user_sync_duration_seconds",
            Help: "Duration of user sync operations",
            Buckets: prometheus.DefBuckets,
        },
        []string{"operation"},
    )
    
    // Метрики ошибок синхронизации
    UserSyncErrorsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "user_sync_errors_total",
            Help: "Total number of user sync errors",
        },
        []string{"error_type"},
    )
)
```

## 🚀 Преимущества реализации

### **1. Безопасность:**
- ✅ Полное маскирование персональных данных
- ✅ Шифрование всех чувствительных данных
- ✅ Rate limiting для защиты API
- ✅ Аудит всех операций

### **2. Надежность:**
- ✅ Retry логика с экспоненциальной задержкой
- ✅ Circuit breaker для защиты от сбоев
- ✅ Graceful degradation при недоступности API
- ✅ Сохранение статуса синхронизации

### **3. Масштабируемость:**
- ✅ Batch обработка пользователей
- ✅ Асинхронная синхронизация
- ✅ Настраиваемые расписания
- ✅ Параллельная обработка

### **4. Мониторинг:**
- ✅ Детальные метрики
- ✅ Логирование всех операций
- ✅ Алерты при ошибках
- ✅ Отслеживание производительности

## 🎯 Результат

**Безопасная и надежная синхронизация пользователей с внешним биллингом, которая:**
- Защищает конфиденциальные данные пользователей
- Обеспечивает надежную синхронизацию
- Масштабируется с ростом количества пользователей
- Предоставляет полный аудит всех операций
- Соответствует требованиям GDPR/152-ФЗ

**Cadence Workflow обеспечивает надежную оркестрацию всех процессов синхронизации! 🔒**
