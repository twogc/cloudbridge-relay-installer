# 🎯 Cadence Best Practices для CloudBridge Relay

## 📋 Обзор

Данное руководство содержит лучшие практики для работы с Cadence Workflow в проекте CloudBridge Relay, основанные на официальной документации Cadence и опыте разработки distributed систем.

## 🔑 Ключевые принципы Cadence

### 1. Детерминированность Workflows

**Основной принцип**: Workflow код должен быть **детерминированным** и производить одинаковое поведение при повторном выполнении (replay).

#### ✅ Правильно:

```go
func MyWorkflow(ctx workflow.Context, input string) error {
    logger := workflow.GetLogger(ctx)
    
    // ✅ Используйте workflow.Now() вместо time.Now()
    startTime := workflow.Now(ctx)
    
    // ✅ Используйте workflow.Sleep() вместо time.Sleep()
    err := workflow.Sleep(ctx, time.Minute)
    if err != nil {
        return err
    }
    
    // ✅ Используйте workflow.Go() для конкурентности
    workflow.Go(ctx, func(ctx workflow.Context) {
        logger.Info("Concurrent execution")
    })
    
    // ✅ Все внешние действия через activities
    var result string
    err = workflow.ExecuteActivity(ctx, "ExternalActivity", input).Get(ctx, &result)
    return err
}
```

#### ❌ Неправильно:

```go
func BadWorkflow(ctx workflow.Context, input string) error {
    // ❌ НЕ используйте стандартный логгер
    log.Println("This will cause issues!")
    
    // ❌ НЕ используйте time.Now()
    now := time.Now()
    
    // ❌ НЕ используйте time.Sleep()
    time.Sleep(time.Minute)
    
    // ❌ НЕ используйте стандартные горутины
    go func() {
        // Недетерминированное поведение!
    }()
    
    // ❌ НЕ делайте прямые HTTP вызовы
    http.Get("https://api.example.com") // Должно быть в activity!
    
    // ❌ НЕ используйте range по maps
    data := map[string]int{"a": 1, "b": 2}
    for key, value := range data { // Порядок недетерминирован!
        // ...
    }
    
    return nil
}
```

### 2. Идемпотентность Activities

**Принцип**: Activities должны быть **идемпотентными** - повторное выполнение с теми же параметрами должно давать тот же результат.

#### ✅ Правильно:

```go
func CreateResourceActivity(ctx context.Context, resourceID string) error {
    // ✅ Проверяем, не был ли ресурс уже создан
    if resourceExists(resourceID) {
        return nil // Успех - ресурс уже существует
    }
    
    // ✅ Создаем только если не существует
    err := createResource(resourceID)
    if err != nil {
        // ✅ Различаем постоянные и временные ошибки
        if isPermanentError(err) {
            return activity.NewNonRetryableError(err)
        }
        return err // Временная ошибка - будет повтор
    }
    
    return nil
}
```

### 3. Использование Cadence API для времени и конкурентности

#### Время:
```go
// ✅ Правильно
currentTime := workflow.Now(ctx)
workflow.Sleep(ctx, time.Hour)

// ❌ Неправильно
currentTime := time.Now()
time.Sleep(time.Hour)
```

#### Конкурентность:
```go
// ✅ Правильно - используйте workflow.Go()
workflow.Go(ctx, func(ctx workflow.Context) {
    // Безопасная конкурентность в workflow
    workflow.ExecuteActivity(ctx, "ParallelActivity", data)
})

// ❌ Неправильно - стандартные горутины
go func() {
    // Это нарушает детерминированность!
}()
```

#### Каналы и Select:
```go
// ✅ Правильно - используйте workflow API
ch := workflow.NewChannel(ctx)
selector := workflow.NewSelector(ctx)
selector.AddReceive(ch, func(c workflow.Channel, more bool) {
    var data string
    c.Receive(ctx, &data)
})

// ❌ Неправильно - стандартные каналы
ch := make(chan string)
select {
case data := <-ch:
    // Недетерминированно!
}
```

## 🏗️ Архитектурные паттерны

### 1. Long-Running Workflows

Для долгоживущих workflows используйте комбинацию таймеров и signal handling:

```go
func LongRunningWorkflow(ctx workflow.Context) error {
    logger := workflow.GetLogger(ctx)
    
    // Настройка signal channels
    shutdownChannel := workflow.GetSignalChannel(ctx, "shutdown")
    configChannel := workflow.GetSignalChannel(ctx, "config-update")
    
    for {
        selector := workflow.NewSelector(ctx)
        
        // Периодическая работа
        timer := workflow.NewTimer(ctx, 24*time.Hour)
        selector.AddFuture(timer, func(f workflow.Future) {
            logger.Info("Daily maintenance")
            // Запускаем maintenance activities
        })
        
        // Обработка сигналов
        selector.AddReceive(shutdownChannel, func(c workflow.Channel, more bool) {
            var signal string
            c.Receive(ctx, &signal)
            if signal == "graceful" {
                logger.Info("Graceful shutdown initiated")
                return // Выход из workflow
            }
        })
        
        selector.AddReceive(configChannel, func(c workflow.Channel, more bool) {
            var newConfig map[string]interface{}
            c.Receive(ctx, &newConfig)
            logger.Info("Configuration updated")
            // Обработка новой конфигурации
        })
        
        selector.Select(ctx)
    }
}
```

### 2. Child Workflows для модульности

```go
func ParentWorkflow(ctx workflow.Context, modules []string) error {
    // Настройки для child workflows
    childOptions := workflow.ChildWorkflowOptions{
        ExecutionStartToCloseTimeout: time.Hour,
        TaskStartToCloseTimeout:      time.Minute,
    }
    
    // Параллельный запуск child workflows
    futures := make([]workflow.Future, len(modules))
    for i, module := range modules {
        childCtx := workflow.WithChildWorkflowOptions(ctx, childOptions)
        futures[i] = workflow.ExecuteChildWorkflow(childCtx, ModuleWorkflow, module)
    }
    
    // Сбор результатов
    for i, future := range futures {
        var result string
        err := future.Get(ctx, &result)
        if err != nil {
            logger.Error("Module failed", "module", modules[i], "error", err)
            // Решение о том, продолжать ли выполнение
        }
    }
    
    return nil
}
```

### 3. Retry Policies

Настройка автоматических повторов для activities:

```go
func WorkflowWithRetries(ctx workflow.Context) error {
    // Конфигурация retry policy
    retryPolicy := &workflow.RetryPolicy{
        InitialInterval:    time.Second,
        BackoffCoefficient: 2.0,
        MaximumInterval:    time.Minute,
        MaximumAttempts:    5,
        NonRetryableErrorTypes: []string{
            "ValidationError",
            "AuthenticationError",
        },
    }
    
    ao := workflow.ActivityOptions{
        StartToCloseTimeout: 10 * time.Minute,
        HeartbeatTimeout:    time.Minute,
        RetryPolicy:         retryPolicy,
    }
    
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // Activity будет автоматически повторена при сбоях
    return workflow.ExecuteActivity(ctx, "UnreliableActivity", data).Get(ctx, nil)
}
```

## 🧪 Тестирование Workflows

### Unit Testing с TestWorkflowEnvironment

```go
func TestMyWorkflow(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()
    
    // Мокирование activities
    env.OnActivity("ExternalActivity", mock.Anything).Return("success", nil)
    
    // Выполнение workflow
    env.ExecuteWorkflow(MyWorkflow, "test-input")
    
    // Проверка результатов
    assert.True(t, env.IsWorkflowCompleted())
    assert.NoError(t, env.GetWorkflowError())
    
    var result string
    err := env.GetWorkflowResult(&result)
    assert.NoError(t, err)
    assert.Equal(t, "expected-result", result)
}
```

### Тестирование с таймерами

```go
func TestWorkflowWithTimers(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()
    
    // Настройка автоматического fast-forward таймеров
    env.SetTestTimeout(time.Minute * 5)
    
    env.ExecuteWorkflow(TimerBasedWorkflow)
    
    // Таймеры выполнятся мгновенно в тестах
    assert.True(t, env.IsWorkflowCompleted())
}
```

### Тестирование signals

```go
func TestWorkflowWithSignals(t *testing.T) {
    testSuite := &testsuite.WorkflowTestSuite{}
    env := testSuite.NewTestWorkflowEnvironment()
    
    env.ExecuteWorkflow(SignalHandlingWorkflow)
    
    // Отправка signal во время выполнения
    env.SignalWorkflow("pause", nil)
    env.SignalWorkflow("resume", nil)
    env.SignalWorkflow("shutdown", nil)
    
    assert.True(t, env.IsWorkflowCompleted())
}
```

## 📊 Monitoring и Observability

### Structured Logging

```go
func MonitoredWorkflow(ctx workflow.Context, orderID string) error {
    logger := workflow.GetLogger(ctx)
    
    // ✅ Структурированное логирование с контекстом
    logger.Info("Processing order",
        "order_id", orderID,
        "workflow_start", workflow.Now(ctx),
    )
    
    // Логирование прогресса
    logger.Info("Validating order", "order_id", orderID)
    
    var validationResult bool
    err := workflow.ExecuteActivity(ctx, "ValidateOrder", orderID).Get(ctx, &validationResult)
    if err != nil {
        logger.Error("Order validation failed",
            "order_id", orderID,
            "error", err,
        )
        return err
    }
    
    logger.Info("Order processed successfully",
        "order_id", orderID,
        "validation_result", validationResult,
    )
    
    return nil
}
```

### Custom Metrics в Activities

```go
func MetricsAwareActivity(ctx context.Context, operation string) error {
    startTime := time.Now()
    
    // Выполнение операции
    err := performOperation(operation)
    
    // Запись метрик
    duration := time.Since(startTime)
    if err != nil {
        metrics.ActivityFailureCounter.WithLabelValues(operation).Inc()
    } else {
        metrics.ActivitySuccessCounter.WithLabelValues(operation).Inc()
    }
    metrics.ActivityDuration.WithLabelValues(operation).Observe(duration.Seconds())
    
    return err
}
```

## 🚨 Обработка ошибок

### Классификация ошибок

```go
func ErrorHandlingActivity(ctx context.Context, data string) error {
    err := externalServiceCall(data)
    if err != nil {
        // Проверяем тип ошибки
        switch {
        case isValidationError(err):
            // Не повторяем ошибки валидации
            return activity.NewNonRetryableError(fmt.Errorf("validation failed: %w", err))
            
        case isAuthError(err):
            // Не повторяем ошибки аутентификации
            return activity.NewNonRetryableError(fmt.Errorf("authentication failed: %w", err))
            
        case isRateLimitError(err):
            // Временная ошибка - будет повтор с backoff
            return fmt.Errorf("rate limited: %w", err)
            
        case isNetworkError(err):
            // Временная ошибка
            return fmt.Errorf("network error: %w", err)
            
        default:
            // По умолчанию считаем временной ошибкой
            return err
        }
    }
    
    return nil
}
```

### Circuit Breaker паттерн

```go
func CircuitBreakerActivity(ctx context.Context, service string) error {
    circuitBreaker := getCircuitBreaker(service)
    
    if circuitBreaker.IsOpen() {
        return activity.NewNonRetryableError(
            fmt.Errorf("circuit breaker open for service %s", service))
    }
    
    err := callService(service)
    if err != nil {
        circuitBreaker.RecordFailure()
        return err
    }
    
    circuitBreaker.RecordSuccess()
    return nil
}
```

## 🔐 Security Best Practices

### 1. Secure Configuration

```go
// Никогда не логируйте sensitive данные
func SecureWorkflow(ctx workflow.Context, apiKey string) error {
    logger := workflow.GetLogger(ctx)
    
    // ❌ Неправильно
    // logger.Info("Processing with API key", "api_key", apiKey)
    
    // ✅ Правильно
    logger.Info("Processing with API key", "api_key_length", len(apiKey))
    
    // Передаем sensitive данные только в activities
    return workflow.ExecuteActivity(ctx, "SecureAPICall", apiKey).Get(ctx, nil)
}
```

### 2. Input Validation

```go
func ValidatedWorkflow(ctx workflow.Context, input UserInput) error {
    // ✅ Валидация входных данных
    if err := validateInput(input); err != nil {
        return workflow.NewApplicationError("invalid_input", err.Error())
    }
    
    // Продолжение выполнения с валидными данными
    return processValidInput(ctx, input)
}
```

## 🔄 Migration и Versioning

### Workflow Versioning

```go
func VersionedWorkflow(ctx workflow.Context, input string) error {
    // Используйте GetVersion для безопасных изменений
    version := workflow.GetVersion(ctx, "feature-x", workflow.DefaultVersion, 2)
    
    if version == workflow.DefaultVersion {
        // Старая логика
        return oldLogic(ctx, input)
    } else {
        // Новая логика
        return newLogic(ctx, input)
    }
}
```

## 📋 Чек-лист для разработчиков

### Перед созданием Workflow:
- [ ] Workflow код детерминированный?
- [ ] Используете ли workflow.* API для времени и конкурентности?
- [ ] Все внешние вызовы вынесены в activities?
- [ ] Логирование через workflow.GetLogger()?
- [ ] Нет range по maps?

### Перед созданием Activity:
- [ ] Activity идемпотентная?
- [ ] Правильно классифицируете ошибки?
- [ ] Используете heartbeat для долгих операций?
- [ ] Есть proper timeout и retry настройки?

### Перед деплоем:
- [ ] Все workflows и activities зарегистрированы?
- [ ] Конфигурация timeout'ов адекватная?
- [ ] Есть мониторинг и алерты?
- [ ] Написаны тесты?

## 📚 Дополнительные ресурсы

- [Cadence Go Client Documentation](https://cadenceworkflow.io/docs/go-client/)
- [Workflow Implementation Guide](https://cadenceworkflow.io/docs/concepts/workflows/)
- [Activity Best Practices](https://cadenceworkflow.io/docs/concepts/activities/)
- [Testing Workflows](https://cadenceworkflow.io/docs/go-client/testing/)

---

**💡 Помните**: Соблюдение этих практик критично для стабильной работы Cadence workflows. При сомнениях всегда выбирайте более безопасный и детерминированный подход.
