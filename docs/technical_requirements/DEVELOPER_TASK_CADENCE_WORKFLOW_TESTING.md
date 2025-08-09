# 🧪 Задание разработчику: Полное тестирование Cadence Workflow

## 📋 Описание задачи

Необходимо устранить проблемы зависания тестов в Cadence Workflow реализации и обеспечить полное покрытие тестами всех компонентов системы, следуя Золотым правилам разработки CloudBridge Relay.

## 🎯 Основные цели

1. **Устранить зависание тестов** в существующей реализации
2. **Обеспечить стабильное выполнение** всех типов тестов (unit, integration, benchmark)
3. **Повысить надежность** долгоживущих workflows
4. **Соблюсти Золотые правила** CloudBridge Relay Development

## 🔧 Технические требования

### 1. Исправление долгоживущих Workflows

#### 1.1 AutoScalingWorkflow
**Проблема**: Workflow зависает в бесконечном цикле мониторинга метрик
**Решение**:
```go
// Добавить корректную обработку сигналов остановки
func AutoScalingWorkflow(ctx workflow.Context, input AutoScalingInput) error {
    // Настройка канала для получения сигналов
    controlSignal := workflow.GetSignalChannel(ctx, "control-command")
    
    for {
        selector := workflow.NewSelector(ctx)
        
        // Добавить обработку сигнала остановки
        selector.AddReceive(controlSignal, func(c workflow.Channel, more bool) {
            var signal string
            c.Receive(ctx, &signal)
            if signal == "shutdown" || signal == "stop" {
                workflow.GetLogger(ctx).Info("Получен сигнал остановки AutoScaling")
                return // Выход из функции
            }
        })
        
        // Добавить таймер с более коротким интервалом для тестов
        timer := workflow.NewTimer(ctx, getScalingInterval(ctx))
        selector.AddFuture(timer, func(f workflow.Future) {
            // Логика сбора метрик и масштабирования
        })
        
        selector.Select(ctx)
    }
}

func getScalingInterval(ctx workflow.Context) time.Duration {
    // В тестовом окружении использовать короткие интервалы
    if workflow.IsReplaying(ctx) || isTestEnvironment() {
        return time.Second * 1
    }
    return time.Minute * 2
}
```

#### 1.2 HealthCheckWorkflow
**Проблема**: Бесконечный цикл проверки здоровья
**Решение**:
```go
func HealthCheckWorkflow(ctx workflow.Context, input HealthCheckInput) error {
    controlSignal := workflow.GetSignalChannel(ctx, "control-command")
    
    for {
        selector := workflow.NewSelector(ctx)
        
        // Обработка команд управления
        selector.AddReceive(controlSignal, func(c workflow.Channel, more bool) {
            var command string
            c.Receive(ctx, &command)
            switch command {
            case "shutdown", "stop":
                workflow.GetLogger(ctx).Info("Остановка HealthCheck workflow")
                return
            case "pause":
                // Логика паузы
            }
        })
        
        // Периодическая проверка здоровья
        timer := workflow.NewTimer(ctx, getHealthCheckInterval(ctx))
        selector.AddFuture(timer, func(f workflow.Future) {
            performHealthCheck(ctx, input)
        })
        
        selector.Select(ctx)
    }
}
```

### 2. Улучшение тестовой инфраструктуры

#### 2.1 Unit тесты с корректными таймаутами
```go
func TestAutoScalingWorkflow(t *testing.T) {
    suite := &testsuite.WorkflowTestSuite{}
    env := suite.NewTestWorkflowEnvironment()
    
    // Установка короткого таймаута для тестов
    env.SetTestTimeout(time.Second * 5)
    
    // Мокирование activities
    env.OnActivity("CollectScalingMetricsActivity", mock.Anything).Return(
        map[string]interface{}{"cpu_usage": 0.3}, nil)
    
    // Запуск workflow в отдельной горутине
    go func() {
        env.ExecuteWorkflow(AutoScalingWorkflow, AutoScalingInput{
            ResourceGroup: "test-group",
            ScalingPolicy: "cpu-based",
        })
    }()
    
    // Ожидание запуска workflow
    time.Sleep(100 * time.Millisecond)
    
    // Отправка сигнала остановки
    env.SignalWorkflow("control-command", "shutdown")
    
    // Проверка корректного завершения
    assert.True(t, env.IsWorkflowCompleted())
    assert.NoError(t, env.GetWorkflowError())
}
```

#### 2.2 Интеграционные тесты с реальным Cadence
```go
func TestCadenceIntegrationWithTimeout(t *testing.T) {
    // Настройка с коротким таймаутом
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Запуск workflow
    execution, err := cadenceClient.StartWorkflow(ctx, "AutoScalingWorkflow", input)
    require.NoError(t, err)
    
    // Ожидание запуска
    time.Sleep(2 * time.Second)
    
    // Отправка сигнала остановки
    err = cadenceClient.SignalWorkflow(ctx, execution.ID, execution.RunID, "control-command", "shutdown")
    require.NoError(t, err)
    
    // Проверка завершения в разумное время
    done := make(chan bool)
    go func() {
        cadenceClient.GetWorkflowResult(ctx, execution.ID, execution.RunID, nil)
        done <- true
    }()
    
    select {
    case <-done:
        t.Log("Workflow завершился корректно")
    case <-time.After(10 * time.Second):
        t.Fatal("Workflow не завершился в ожидаемое время")
    }
}
```

### 3. Benchmark тесты с оптимизацией

#### 3.1 Отдельные benchmark файлы
```go
// internal/cadence/workflows/scaling_benchmark_test.go
func BenchmarkAutoScalingWorkflowShortRun(b *testing.B) {
    for i := 0; i < b.N; i++ {
        suite := &testsuite.WorkflowTestSuite{}
        env := suite.NewTestWorkflowEnvironment()
        env.SetTestTimeout(time.Second * 2)
        
        // Мокирование для быстрого выполнения
        env.OnActivity("CollectScalingMetricsActivity", mock.Anything).Return(
            map[string]interface{}{"action": "no_scaling_needed"}, nil)
        
        go env.ExecuteWorkflow(AutoScalingWorkflow, input)
        time.Sleep(50 * time.Millisecond)
        env.SignalWorkflow("control-command", "shutdown")
        
        // Ожидание завершения
        for !env.IsWorkflowCompleted() && env.GetWorkflowError() == nil {
            time.Sleep(10 * time.Millisecond)
        }
    }
}
```

## 🏆 Золотые правила CloudBridge Relay Development

### Правило 1: Архитектура и структура проекта
- **Применение**: Все тесты должны следовать стандартной структуре `internal/cadence/*/`
- **Требование**: Каждый workflow должен иметь соответствующий тест в том же пакете

### Правило 2: Обработка ошибок и логирование
```go
// ✅ Правильно
func TestWorkflowWithErrorHandling(t *testing.T) {
    logger := zap.NewDevelopment()
    defer logger.Sync()
    
    // Тест с обработкой ошибок
    if err := env.GetWorkflowError(); err != nil {
        logger.Error("Workflow завершился с ошибкой", 
            zap.Error(err),
            zap.String("workflow", "AutoScalingWorkflow"))
        // Анализ типа ошибки
        if !isTimeoutError(err) {
            t.Fatalf("Неожиданная ошибка: %v", err)
        }
    }
}
```

### Правило 3: Управление конфигурацией
```go
// Использование тестовых конфигураций
func loadTestConfig() *cadence.CadenceConfig {
    config := &cadence.CadenceConfig{
        Domain: "test-domain",
        TaskList: "test-task-list",
        WorkflowTimeout: "30s", // Короткие таймауты для тестов
        ActivityTimeout: "5s",
    }
    return config
}
```

### Правило 4: Безопасность и валидация
- **Тестовые данные**: Использовать только тестовые credentials и endpoints
- **Изоляция**: Каждый тест должен быть изолированным
- **Cleanup**: Обязательная очистка ресурсов после тестов

### Правило 5: Производительность и оптимизация
```go
// Benchmark с измерением производительности
func BenchmarkWorkflowPerformance(b *testing.B) {
    b.ReportAllocs() // Отчет по аллокациям памяти
    
    for i := 0; i < b.N; i++ {
        startTime := time.Now()
        
        // Выполнение теста
        runWorkflowTest()
        
        duration := time.Since(startTime)
        if duration > time.Second*5 {
            b.Fatalf("Workflow выполняется слишком долго: %v", duration)
        }
    }
}
```

### Правило 6: Мониторинг и наблюдаемость
- **Метрики**: Все тесты должны собирать метрики времени выполнения
- **Логирование**: Структурированное логирование с контекстом
- **Трейсинг**: Отслеживание выполнения в тестах

## 📝 Задачи для выполнения

### Высокий приоритет ⚡
1. **Исправить AutoScalingWorkflow**
   - Добавить обработку сигналов остановки
   - Реализовать корректное завершение цикла
   - Добавить тестовый режим с короткими интервалами

2. **Исправить HealthCheckWorkflow**
   - Добавить команды управления (stop, pause, resume)
   - Реализовать graceful shutdown
   - Оптимизировать для тестового окружения

3. **Обновить все unit тесты**
   - Добавить таймауты (максимум 30 секунд)
   - Реализовать корректную отправку сигналов
   - Добавить проверки завершения workflow

### Средний приоритет 🔧
4. **Создать тестовые утилиты**
   - Хелперы для создания тестовых environment
   - Общие моки для activities
   - Утилиты для отправки сигналов

5. **Улучшить benchmark тесты**
   - Разделить на быстрые и полные тесты
   - Добавить метрики производительности
   - Оптимизировать время выполнения

6. **Добавить интеграционные тесты**
   - Тесты с реальным Cadence сервером
   - Проверка взаимодействия компонентов
   - Тестирование отказоустойчивости

### Низкий приоритет 📋
7. **Расширить документацию**
   - Руководство по написанию тестов
   - Best practices для workflow тестирования
   - Примеры использования сигналов

8. **Добавить CI/CD интеграцию**
   - Автоматический запуск тестов
   - Проверка производительности
   - Отчеты о покрытии

## 🔍 Критерии приемки

### Функциональные требования ✅
- [ ] Все unit тесты выполняются менее чем за 2 минуты
- [ ] Интеграционные тесты завершаются менее чем за 5 минут
- [ ] Benchmark тесты предоставляют осмысленные метрики
- [ ] Нет зависающих тестов
- [ ] Покрытие кода тестами > 80%

### Технические требования 🔧
- [ ] Все долгоживущие workflows поддерживают graceful shutdown
- [ ] Реализована обработка сигналов управления
- [ ] Добавлены тестовые конфигурации с короткими таймаутами
- [ ] Логирование соответствует стандартам проекта
- [ ] Метрики интегрированы с Prometheus

### Качество кода 📊
- [ ] Соблюдены все Золотые правила CloudBridge Relay
- [ ] Код проходит линтер без предупреждений
- [ ] Документация обновлена
- [ ] Примеры использования добавлены

## 📚 Дополнительные ресурсы

### Документация
- `docs/cadence/WORKFLOW_GUIDE.md` - Руководство по workflows
- `docs/cadence/ACTIVITIES_GUIDE.md` - Руководство по activities
- `docs/cadence/CADENCE_BEST_PRACTICES.md` - Best practices

### Примеры кода
- `internal/cadence/workflows/main_test.go` - Основные unit тесты
- `internal/cadence/workflows/main_benchmark_test.go` - Benchmark тесты
- `internal/cadence/integration_test/` - Интеграционные тесты

### Конфигурация
- `configs/cadence.yaml` - Основная конфигурация
- Создать `configs/cadence-test.yaml` - Тестовая конфигурация

## ⏰ Временные рамки

**Общее время выполнения**: 3-5 дней

- **День 1**: Анализ проблем и исправление AutoScalingWorkflow
- **День 2**: Исправление HealthCheckWorkflow и обновление unit тестов
- **День 3**: Улучшение benchmark и интеграционных тестов
- **День 4**: Создание тестовых утилит и документации
- **День 5**: Финальное тестирование и валидация

## 🚀 Результат

После выполнения всех задач:
- ✅ Стабильная тестовая инфраструктура без зависаний
- ✅ Полное покрытие всех Cadence компонентов
- ✅ Соответствие Золотым правилам CloudBridge Relay
- ✅ Готовность к продакшн развертыванию
- ✅ Документированные best practices для команды
