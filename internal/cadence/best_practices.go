package cadence

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// BestPracticesGuide содержит примеры следования лучшим практикам Cadence
// Этот файл служит справочником для разработчиков

// ✅ ПРАВИЛЬНО: Детерминированный workflow
func CorrectWorkflowExample(ctx workflow.Context, input string) error {
	logger := workflow.GetLogger(ctx)
	
	// ✅ Используем workflow.Now() вместо time.Now()
	startTime := workflow.Now(ctx)
	logger.Info("Workflow started", zap.Time("start_time", startTime), zap.String("input", input))
	
	// ✅ Используем workflow.Sleep() вместо time.Sleep()
	err := workflow.Sleep(ctx, time.Minute)
	if err != nil {
		return err
	}
	
	// ✅ Используем workflow.Go() для конкурентности
	workflow.Go(ctx, func(ctx workflow.Context) {
		logger.Info("Concurrent execution in workflow.Go")
		// Здесь можно безопасно вызывать другие workflow API
	})
	
	// ✅ Все внешние действия через activities
	var result string
	err = workflow.ExecuteActivity(ctx, "ExternalServiceActivity", input).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed", zap.Error(err))
		return err
	}
	
	// ✅ Используем workflow API для каналов
	ch := workflow.NewChannel(ctx)
	
	// ✅ Используем workflow.Selector вместо select
	selector := workflow.NewSelector(ctx)
	selector.AddReceive(ch, func(c workflow.Channel, more bool) {
		var data string
		c.Receive(ctx, &data)
		logger.Info("Received data", zap.String("data", data))
	})
	
	logger.Info("Workflow completed successfully")
	return nil
}

// ❌ НЕПРАВИЛЬНО: Недетерминированный workflow
// Этот пример показывает, чего НЕ нужно делать
func IncorrectWorkflowExample(ctx workflow.Context, input string) error {
	// ❌ НЕ используйте стандартный логгер
	// log.Println("This is wrong!") 
	
	// ❌ НЕ используйте time.Now()
	// now := time.Now()
	
	// ❌ НЕ используйте time.Sleep()
	// time.Sleep(time.Minute)
	
	// ❌ НЕ используйте стандартные горутины
	// go func() {
	//     // Это приведет к недетерминированному поведению
	// }()
	
	// ❌ НЕ делайте внешние вызовы напрямую в workflow
	// http.Get("https://api.example.com") // Это должно быть в activity!
	
	// ❌ НЕ используйте range по maps (недетерминированный порядок)
	// data := map[string]int{"a": 1, "b": 2}
	// for key, value := range data {
	//     // Порядок итерации может различаться при replay
	// }
	
	// ❌ НЕ используйте стандартные каналы и select
	// ch := make(chan string)
	// select {
	// case data := <-ch:
	//     // Недетерминированно
	// }
	
	return nil
}

// ActivityBestPractices демонстрирует лучшие практики для activities
type ActivityBestPractices struct{}

// ✅ ПРАВИЛЬНО: Идемпотентная activity
func (a *ActivityBestPractices) IdempotentActivity(ctx context.Context, resourceID string) error {
	// ✅ Проверяем, не был ли ресурс уже создан
	if resourceExists(resourceID) {
		// ✅ Возвращаем успех, если ресурс уже существует
		return nil
	}
	
	// ✅ Создаем ресурс только если его нет
	err := createResource(resourceID)
	if err != nil {
		// ✅ Если ошибка постоянная, возвращаем ошибку как есть
		if isPermanentError(err) {
			return err // В Cadence Go нет NewNonRetryableError
		}
		// ✅ Временные ошибки автоматически повторятся
		return err
	}
	
	return nil
}

// ✅ ПРАВИЛЬНО: Activity с heartbeat для долгих операций
func (a *ActivityBestPractices) LongRunningActivity(ctx context.Context, items []string) error {
	for i, item := range items {
		// ✅ Отправляем heartbeat для индикации прогресса
		activity.RecordHeartbeat(ctx, i)
		
		// ✅ Проверяем отмену через context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Продолжаем работу
		}
		
		// Выполняем работу
		err := processItem(item)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// ✅ ПРАВИЛЬНО: Использование retry policies
func WorkflowWithRetryExample(ctx workflow.Context) error {
	// ✅ Настраиваем retry policy для activity
	retryPolicy := &workflow.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    5,
		// NonRetryableErrorTypes не поддерживается в этой версии Cadence
	}
	
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    time.Minute,
		RetryPolicy:         retryPolicy,
	}
	
	ctx = workflow.WithActivityOptions(ctx, ao)
	
	// ✅ Activity будет автоматически повторена при сбоях
	return workflow.ExecuteActivity(ctx, "ReliableActivity", "data").Get(ctx, nil)
}

// ✅ ПРАВИЛЬНО: Использование child workflows
func ParentWorkflowExample(ctx workflow.Context, tasks []string) error {
	logger := workflow.GetLogger(ctx)
	
	// ✅ Настройки для child workflows
	childOptions := workflow.ChildWorkflowOptions{
		ExecutionStartToCloseTimeout: time.Hour,
		TaskStartToCloseTimeout:      time.Minute,
	}
	
	// ✅ Запускаем child workflows параллельно
	futures := make([]workflow.Future, len(tasks))
	for i, task := range tasks {
		childCtx := workflow.WithChildOptions(ctx, childOptions)
		futures[i] = workflow.ExecuteChildWorkflow(childCtx, ProcessTaskWorkflow, task)
	}
	
	// ✅ Ждем завершения всех child workflows
	for i, future := range futures {
		var result string
		err := future.Get(ctx, &result)
		if err != nil {
			logger.Error("Child workflow failed", zap.Int("task_index", i), zap.Error(err))
			// Решаем, продолжать ли выполнение остальных
			return err
		}
		logger.Info("Child workflow completed", zap.Int("task_index", i), zap.String("result", result))
	}
	
	return nil
}

// ✅ ПРАВИЛЬНО: Signal handling в workflow
func SignalHandlingWorkflowExample(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	
	// ✅ Создаем каналы для сигналов
	pauseSignal := workflow.GetSignalChannel(ctx, "pause")
	resumeSignal := workflow.GetSignalChannel(ctx, "resume")
	stopSignal := workflow.GetSignalChannel(ctx, "stop")
	
	isPaused := false
	
	for {
		selector := workflow.NewSelector(ctx)
		
		// ✅ Обрабатываем сигналы
		selector.AddReceive(pauseSignal, func(c workflow.Channel, more bool) {
			var signal interface{}
			c.Receive(ctx, &signal)
			isPaused = true
			logger.Info("Workflow paused")
		})
		
		selector.AddReceive(resumeSignal, func(c workflow.Channel, more bool) {
			var signal interface{}
			c.Receive(ctx, &signal)
			isPaused = false
			logger.Info("Workflow resumed")
		})
		
		selector.AddReceive(stopSignal, func(c workflow.Channel, more bool) {
			var signal interface{}
			c.Receive(ctx, &signal)
			logger.Info("Workflow stop requested")
			// Выходим из функции
		})
		
		// ✅ Добавляем таймер для периодической работы
		if !isPaused {
			timer := workflow.NewTimer(ctx, 30*time.Second)
			selector.AddFuture(timer, func(f workflow.Future) {
				logger.Info("Performing periodic work")
				// Здесь можно вызвать activity или child workflow
			})
		}
		
		selector.Select(ctx)
	}
}

// Вспомогательные функции (заглушки для примера)
func resourceExists(id string) bool { return false }
func createResource(id string) error { return nil }
func isPermanentError(err error) bool { return false }
func processItem(item string) error { return nil }

// ProcessTaskWorkflow - пример child workflow
func ProcessTaskWorkflow(ctx workflow.Context, task string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Processing task", zap.String("task", task))
	
	// Выполняем обработку через activity
	var result string
	err := workflow.ExecuteActivity(ctx, "ProcessTaskActivity", task).Get(ctx, &result)
	return result, err
}
