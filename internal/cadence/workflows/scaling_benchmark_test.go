package workflows

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// BenchmarkAutoScalingWorkflowShortRun быстрый benchmark для AutoScalingWorkflow
func BenchmarkAutoScalingWorkflowShortRun(b *testing.B) {
	testUtils := NewTestUtils()
	
	for i := 0; i < b.N; i++ {
		env := testUtils.CreateTestEnvironment()
		env.SetTestTimeout(time.Second * 5)
		
		// Мокирование для быстрого выполнения (без масштабирования)
		testUtils.MockScalingActivities(env)
		
		input := testUtils.CreateAutoScalingInput()
		
		// Запуск и быстрая остановка workflow
		go func() {
			env.ExecuteWorkflow(AutoScalingWorkflow, input)
		}()
		
		time.Sleep(50 * time.Millisecond)
		env.SignalWorkflow("control-command", "shutdown")
		
		// Ожидание завершения
		err := testUtils.WaitForWorkflowCompletion(env, time.Second*3)
		if err != nil {
			b.Fatalf("Workflow не завершился: %v", err)
		}
	}
}

// BenchmarkAutoScalingWorkflowFullRun полный benchmark для AutoScalingWorkflow
func BenchmarkAutoScalingWorkflowFullRun(b *testing.B) {
	if testing.Short() {
		b.Skip("Пропуск полного benchmark в коротком режиме")
	}
	
	testUtils := NewTestUtils()
	
	for i := 0; i < b.N; i++ {
		env := testUtils.CreateTestEnvironment()
		env.SetTestTimeout(time.Second * 30)
		
		// Полные моки с масштабированием
		env.OnActivity("CollectScalingMetricsActivity", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"current_capacity": map[string]interface{}{"cpu": 85, "memory": 90},
				"metrics_collected": true,
			}, nil)
		env.OnActivity("AnalyzeScalingNeedActivity", mock.Anything, mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"need_scaling":    true,
				"scaling_action": "scale_up",
				"reason":         "high_resource_usage",
			}, nil)
		env.OnActivity("PerformScalingActivity", mock.Anything, mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"scaling_action": "scale_up_completed",
				"new_capacity":   map[string]interface{}{"instances": 4},
			}, nil)
		env.OnActivity("GetCurrentCapacityActivity", mock.Anything).Return(
			map[string]interface{}{
				"current_capacity": map[string]interface{}{"instances": 4},
				"target_capacity":  map[string]interface{}{"instances": 4},
			}, nil)
		
		input := testUtils.CreateAutoScalingInput()
		
		startTime := time.Now()
		
		// Запуск workflow с выполнением нескольких циклов
		go func() {
			env.ExecuteWorkflow(AutoScalingWorkflow, input)
		}()
		
		// Даем время на несколько циклов масштабирования
		time.Sleep(3 * time.Second)
		env.SignalWorkflow("control-command", "shutdown")
		
		err := testUtils.WaitForWorkflowCompletion(env, time.Second*10)
		duration := time.Since(startTime)
		
		if err != nil {
			b.Fatalf("Workflow не завершился: %v", err)
		}
		
		// Проверка производительности
		if duration > time.Second*15 {
			b.Fatalf("Workflow выполняется слишком долго: %v", duration)
		}
	}
}

// BenchmarkHealthCheckWorkflowShortRun быстрый benchmark для HealthCheckWorkflow
func BenchmarkHealthCheckWorkflowShortRun(b *testing.B) {
	testUtils := NewTestUtils()
	
	for i := 0; i < b.N; i++ {
		env := testUtils.CreateTestEnvironment()
		env.SetTestTimeout(time.Second * 5)
		
		// Быстрые моки
		testUtils.MockCommonActivities(env)
		
		input := testUtils.CreateHealthCheckInput("test-component")
		
		startTime := time.Now()
		
		go func() {
			env.ExecuteWorkflow(HealthCheckWorkflow, input)
		}()
		
		time.Sleep(50 * time.Millisecond)
		env.SignalWorkflow("control-command", "shutdown")
		
		err := testUtils.WaitForWorkflowCompletion(env, time.Second*2)
		duration := time.Since(startTime)
		
		if err != nil {
			b.Fatalf("Workflow не завершился: %v", err)
		}
		
		// Проверка быстродействия
		if duration > time.Second*3 {
			b.Fatalf("Быстрый benchmark выполняется слишком долго: %v", duration)
		}
	}
}

// BenchmarkWorkflowMemoryUsage benchmark использования памяти
func BenchmarkWorkflowMemoryUsage(b *testing.B) {
	b.ReportAllocs() // Отчет по аллокациям памяти
	
	testUtils := NewTestUtils()
	
	for i := 0; i < b.N; i++ {
		env := testUtils.CreateTestEnvironment()
		testUtils.MockCommonActivities(env)
		
		input := testUtils.CreateModuleInitializationInput("memory-test-module")
		
		env.ExecuteWorkflow(ModuleInitializationWorkflow, input)
		
		assert.True(b, env.IsWorkflowCompleted())
		assert.NoError(b, env.GetWorkflowError())
	}
}

// BenchmarkConcurrentWorkflows benchmark параллельного выполнения workflows
func BenchmarkConcurrentWorkflows(b *testing.B) {
	if testing.Short() {
		b.Skip("Пропуск benchmark параллельных workflows в коротком режиме")
	}
	
	testUtils := NewTestUtils()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			env := testUtils.CreateTestEnvironment()
			testUtils.MockCommonActivities(env)
			
			input := testUtils.CreateModuleInitializationInput("concurrent-test-module")
			
			startTime := time.Now()
			env.ExecuteWorkflow(ModuleInitializationWorkflow, input)
			duration := time.Since(startTime)
			
			if !env.IsWorkflowCompleted() {
				b.Fatal("Workflow не завершился")
			}
			
			if env.GetWorkflowError() != nil {
				b.Fatalf("Workflow завершился с ошибкой: %v", env.GetWorkflowError())
			}
			
			// Проверка производительности
			if duration > time.Second*5 {
				b.Fatalf("Concurrent workflow выполняется слишком долго: %v", duration)
			}
		}
	})
}

// BenchmarkWorkflowThroughput benchmark пропускной способности
func BenchmarkWorkflowThroughput(b *testing.B) {
	testUtils := NewTestUtils()
	
	// Подготовка множества входных данных
	inputs := make([]ModuleInitializationInput, b.N)
	for i := 0; i < b.N; i++ {
		inputs[i] = testUtils.CreateModuleInitializationInput("throughput-test-module")
	}
	
	b.ResetTimer()
	b.StartTimer()
	
	for i := 0; i < b.N; i++ {
		env := testUtils.CreateTestEnvironment()
		testUtils.MockCommonActivities(env)
		
		env.ExecuteWorkflow(ModuleInitializationWorkflow, inputs[i])
		
		if !env.IsWorkflowCompleted() {
			b.Fatal("Workflow не завершился")
		}
	}
	
	b.StopTimer()
}

// BenchmarkSignalProcessing benchmark обработки сигналов
func BenchmarkSignalProcessing(b *testing.B) {
	testUtils := NewTestUtils()
	
	for i := 0; i < b.N; i++ {
		env := testUtils.CreateTestEnvironment()
		testUtils.MockCommonActivities(env)
		
		input := testUtils.CreateHealthCheckInput("signal-test")
		
		startTime := time.Now()
		
		go func() {
			env.ExecuteWorkflow(HealthCheckWorkflow, input)
		}()
		
		time.Sleep(10 * time.Millisecond) // Минимальная задержка
		
		// Измеряем время обработки сигнала
		signalStart := time.Now()
		env.SignalWorkflow("control-command", "shutdown")
		
		err := testUtils.WaitForWorkflowCompletion(env, time.Second*2)
		signalDuration := time.Since(signalStart)
		totalDuration := time.Since(startTime)
		
		if err != nil {
			b.Fatalf("Workflow не завершился: %v", err)
		}
		
		// Проверка скорости обработки сигнала
		if signalDuration > time.Millisecond*500 {
			b.Fatalf("Сигнал обрабатывается слишком долго: %v", signalDuration)
		}
		
		if totalDuration > time.Second*3 {
			b.Fatalf("Общее время выполнения слишком долгое: %v", totalDuration)
		}
	}
}

// BenchmarkWorkflowRetry benchmark retry логики
func BenchmarkWorkflowRetry(b *testing.B) {
	testUtils := NewTestUtils()
	
	for i := 0; i < b.N; i++ {
		env := testUtils.CreateTestEnvironment()
		
		// Мок с первоначальной ошибкой, затем успех
		env.OnActivity("CheckDependenciesActivity", mock.Anything).Return("", assert.AnError).Once()
		env.OnActivity("CheckDependenciesActivity", mock.Anything).Return("all_dependencies_available", nil).Once()
		env.OnActivity("InitializeModuleActivity", mock.Anything, mock.Anything).Return("module_initialized", nil)
		env.OnActivity("CheckModuleHealthActivity", mock.Anything).Return("healthy", nil)
		
		input := testUtils.CreateModuleInitializationInput("retry-test-module")
		
		startTime := time.Now()
		env.ExecuteWorkflow(ModuleInitializationWorkflow, input)
		duration := time.Since(startTime)
		
		if !env.IsWorkflowCompleted() {
			b.Fatal("Workflow не завершился")
		}
		
		if env.GetWorkflowError() != nil {
			b.Fatalf("Workflow завершился с ошибкой: %v", env.GetWorkflowError())
		}
		
		// Проверка времени выполнения с retry
		if duration > time.Second*10 {
			b.Fatalf("Workflow с retry выполняется слишком долго: %v", duration)
		}
	}
}
