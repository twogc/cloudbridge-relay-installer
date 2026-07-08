# Улучшения Retry системы

## Обзор

Реализованы критические улучшения retry системы для обеспечения надежности, потокобезопасности и "вежливости" к внешним API.

## Ключевые улучшения

### 1. Потокобезопасный джиттер

**Проблема**: `math/rand` глобальный и небезопасен для гонок.

**Решение**: Локальный `rand.Rand` с мьютексом для каждого экземпляра конфигурации.

```go
type RetryConfigImproved struct {
    // ... другие поля
    rnd   *rand.Rand
    rndMu sync.Mutex
}
```

### 2. Валидация и нормализация конфигурации

**Проблема**: Некорректные значения могут привести к неожиданному поведению.

**Решение**: Метод `Normalize()` проверяет и исправляет все параметры:

```go
func (rc *RetryConfigImproved) Normalize() {
    if rc.MaxAttempts <= 0 {
        rc.MaxAttempts = 1
    }
    if rc.InitialDelay <= 0 {
        rc.InitialDelay = time.Second
    }
    // ... другие проверки
}
```

### 3. Поддержка Retry-After заголовков

**Проблема**: Игнорирование серверных указаний о времени ожидания.

**Решение**: Метод `CalculateDelayWithRetryAfter()` учитывает заголовок:

```go
func (rc *RetryConfigImproved) CalculateDelayWithRetryAfter(attempt int, retryAfterHeader string) time.Duration {
    if retryAfterHeader != "" {
        // Парсим Retry-After и используем его значение
        if secs, err := time.ParseDuration(retryAfterHeader + "s"); err == nil && secs > 0 {
            return clamp(secs, rc.MinDelay, rc.MaxDelay)
        }
    }
    return rc.CalculateDelay(attempt)
}
```

### 4. Расширенная классификация ошибок

**Проблема**: Неправильная обработка сетевых ошибок и контекстных ошибок.

**Решение**: Отдельные методы для разных типов ошибок:

```go
// HTTP статус коды
func (rc *RetryConfigImproved) IsRetryableError(statusCode int) bool

// Сетевые и контекстные ошибки
func (rc *RetryConfigImproved) IsRetryableNetError(err error) bool
```

### 5. Настраиваемые границы задержек

**Новые параметры**:
- `MinDelay` - минимальная задержка
- `JitterRatio` - доля джиттера (например, 0.1 = ±10%)

### 6. Безопасная нормализация attempt

**Проблема**: `attempt < 1` может сломать формулу.

**Решение**: Гарантированная нормализация:

```go
func (rc *RetryConfigImproved) CalculateDelay(attempt int) time.Duration {
    rc.Normalize()
    if attempt < 1 {
        attempt = 1
    }
    // ... остальная логика
}
```

## Использование

### Базовое использование

```go
// Создание с дефолтными настройками
config := network.DefaultRetryConfigImproved()

// Нормализация (обязательно перед использованием)
config.Normalize()

// Проверка retryable ошибок
if config.IsRetryableError(500) {
    // HTTP 500 - можно повторить
}

if config.IsRetryableNetError(err) {
    // Сетевая ошибка - можно повторить
}
```

### Кастомная конфигурация

```go
config := &network.RetryConfigImproved{
    MaxAttempts:     5,
    InitialDelay:    500 * time.Millisecond,
    MaxDelay:        60 * time.Second,
    MinDelay:        100 * time.Millisecond,
    BackoffFactor:   1.5,
    Jitter:          true,
    JitterRatio:     0.15, // ±15% джиттер
    RetryableErrors: []int{500, 502, 503, 504, 429, 408},
}

// Обязательно нормализуем
config.Normalize()
```

### Обработка Retry-After

```go
// В HTTP клиенте
retryAfter := resp.Header.Get("Retry-After")
delay := config.CalculateDelayWithRetryAfter(attempt, retryAfter)

// Ждем указанное сервером время
time.Sleep(delay)
```

## Миграция

### Обратная совместимость

Старый код продолжит работать, но рекомендуется миграция:

```go
// Старый способ
config := network.DefaultRetryConfig()
delay := config.CalculateDelay(attempt)

// Новый способ
config := network.DefaultRetryConfigImproved()
config.Normalize() // Обязательно!
delay := config.CalculateDelay(attempt)
```

### Постепенная миграция

1. Замените `RetryConfig` на `RetryConfigImproved`
2. Добавьте вызов `Normalize()` перед использованием
3. Обновите проверки ошибок на `IsRetryableNetError()`
4. Добавьте поддержку `Retry-After` заголовков

## Производительность

### Потокобезопасность

- Каждый экземпляр конфигурации имеет свой генератор случайности
- Мьютекс защищает только генерацию джиттера
- Минимальные блокировки

### Память

- Один дополнительный `sync.Mutex` на конфигурацию
- Один `rand.Rand` на конфигурацию
- Незначительное увеличение потребления памяти

## Тестирование

### Unit тесты

```go
func TestRetryConfigNormalize(t *testing.T) {
    config := &RetryConfigImproved{
        MaxAttempts: -1, // Некорректное значение
        InitialDelay: 0, // Некорректное значение
    }
    
    config.Normalize()
    
    assert.Equal(t, 1, config.MaxAttempts)
    assert.Equal(t, time.Second, config.InitialDelay)
}
```

### Интеграционные тесты

```go
func TestRetryWithRetryAfter(t *testing.T) {
    config := DefaultRetryConfigImproved()
    
    // Тест с Retry-After заголовком
    delay := config.CalculateDelayWithRetryAfter(1, "30")
    assert.Equal(t, 30*time.Second, delay)
}
```

## Рекомендации

### Для продакшена

1. **Всегда вызывайте `Normalize()`** перед использованием конфигурации
2. **Используйте `Retry-After`** заголовки для уважения серверных указаний
3. **Настройте `MinDelay`** для предотвращения слишком быстрых retry
4. **Мониторьте circuit breaker** метрики

### Для разработки

1. **Тестируйте с различными конфигурациями**
2. **Проверяйте поведение при сетевых ошибках**
3. **Валидируйте джиттер** для предотвращения thundering herd

## Заключение

Улучшенная retry система обеспечивает:

- **Потокобезопасность** - нет гонок данных
- **Предсказуемость** - валидация всех параметров
- **Вежливость** - уважение к серверным указаниям
- **Гибкость** - настраиваемые параметры
- **Надежность** - правильная обработка ошибок

Система готова к продакшену и обеспечивает стабильную работу с внешними API.
