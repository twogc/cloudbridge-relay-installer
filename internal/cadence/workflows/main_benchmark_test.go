package workflows

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// BenchmarkModuleInitializationWorkflow тестирует производительность инициализации модулей
func BenchmarkModuleInitializationWorkflow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		suite := &testsuite.WorkflowTestSuite{}
		logger, _ := zap.NewDevelopment()
		suite.SetLogger(logger)

		env := suite.NewTestWorkflowEnvironment()

		// Мокаем activities
		env.OnActivity("CheckDependenciesActivity", mock.Anything).Return("all_dependencies_available", nil)
		env.OnActivity("InitializeModuleActivity", mock.Anything, mock.Anything).Return("module_test_initialized", nil)
		env.OnActivity("CheckModuleHealthActivity", mock.Anything).Return("healthy", nil)

		input := ModuleInitializationInput{
			ModuleName:    "benchmark-module",
			Configuration: map[string]interface{}{"test": "benchmark"},
			Dependencies:  []string{"dep1", "dep2"},
		}

		env.ExecuteWorkflow(ModuleInitializationWorkflow, input)

		if env.GetWorkflowError() != nil {
			b.Fatalf("Module initialization workflow failed: %v", env.GetWorkflowError())
		}
	}
}

// BenchmarkActivityExecution тестирует производительность выполнения activities
func BenchmarkActivityExecution(b *testing.B) {
	for i := 0; i < b.N; i++ {
		suite := &testsuite.WorkflowTestSuite{}
		logger, _ := zap.NewDevelopment()
		suite.SetLogger(logger)

		env := suite.NewTestWorkflowEnvironment()

		// Мокаем простые activities
		env.OnActivity("CheckDependenciesActivity", mock.Anything).Return("all_dependencies_available", nil)

		// Простой workflow для тестирования
		workflow := func(ctx workflow.Context, deps []string) (string, error) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				ScheduleToCloseTimeout: time.Minute,
			})

			var result string
			err := workflow.ExecuteActivity(ctx, "CheckDependenciesActivity", deps).Get(ctx, &result)
			return result, err
		}

		env.ExecuteWorkflow(workflow, []string{"dep1", "dep2"})

		if env.GetWorkflowError() != nil {
			b.Fatalf("Simple workflow failed: %v", env.GetWorkflowError())
		}
	}
}

// BenchmarkWorkflowWithConcurrency тестирует производительность с параллельными workflow
func BenchmarkWorkflowWithConcurrency(b *testing.B) {
	input := ModuleInitializationInput{
		ModuleName:    "concurrent-module",
		Configuration: map[string]interface{}{"concurrent": true},
		Dependencies:  []string{},
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			suite := &testsuite.WorkflowTestSuite{}
			logger, _ := zap.NewDevelopment()
			suite.SetLogger(logger)

			env := suite.NewTestWorkflowEnvironment()

			// Мокаем activities для каждого environment
			env.OnActivity("CheckDependenciesActivity", mock.Anything).Return("all_dependencies_available", nil)
			env.OnActivity("InitializeModuleActivity", mock.Anything, mock.Anything).Return("module_concurrent_initialized", nil)
			env.OnActivity("CheckModuleHealthActivity", mock.Anything).Return("healthy", nil)

			env.ExecuteWorkflow(ModuleInitializationWorkflow, input)

			if env.GetWorkflowError() != nil {
				b.Fatalf("Concurrent workflow failed: %v", env.GetWorkflowError())
			}
		}
	})
}

// BenchmarkMemoryUsage тестирует использование памяти
func BenchmarkMemoryUsage(b *testing.B) {
	// Создаем большой input для тестирования памяти
	largeConfig := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		largeConfig[string(rune('A'+i%26))+string(rune('a'+i%26))] = "large_value_" + string(rune(i))
	}

	input := ModuleInitializationInput{
		ModuleName:    "memory-test-module",
		Configuration: largeConfig,
		Dependencies:  make([]string, 50), // Большой список зависимостей
	}

	b.ReportAllocs() // Включаем отчет по аллокациям памяти

	for i := 0; i < b.N; i++ {
		suite := &testsuite.WorkflowTestSuite{}
		logger, _ := zap.NewDevelopment()
		suite.SetLogger(logger)

		env := suite.NewTestWorkflowEnvironment()

		env.OnActivity("CheckDependenciesActivity", mock.Anything).Return("all_dependencies_available", nil)
		env.OnActivity("InitializeModuleActivity", mock.Anything, mock.Anything).Return("module_memory_initialized", nil)
		env.OnActivity("CheckModuleHealthActivity", mock.Anything).Return("healthy", nil)

		env.ExecuteWorkflow(ModuleInitializationWorkflow, input)

		if env.GetWorkflowError() != nil {
			b.Fatalf("Memory test workflow failed: %v", env.GetWorkflowError())
		}
	}
}

// BenchmarkWorkflowScaling тестирует масштабируемость workflows
func BenchmarkWorkflowScaling(b *testing.B) {
	scales := []int{1, 10, 50}

	for _, scale := range scales {
		b.Run("scale_"+string(rune('0'+scale/10))+string(rune('0'+scale%10)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Симулируем запуск множества workflows
				for j := 0; j < scale; j++ {
					suite := &testsuite.WorkflowTestSuite{}
					logger, _ := zap.NewDevelopment()
					suite.SetLogger(logger)

					env := suite.NewTestWorkflowEnvironment()

					env.OnActivity("CheckDependenciesActivity", mock.Anything).Return("all_dependencies_available", nil)
					env.OnActivity("InitializeModuleActivity", mock.Anything, mock.Anything).Return("module_scale_initialized", nil)
					env.OnActivity("CheckModuleHealthActivity", mock.Anything).Return("healthy", nil)

					input := ModuleInitializationInput{
						ModuleName:    "scale-module-" + string(rune('0'+j%10)),
						Configuration: map[string]interface{}{"scale": scale, "instance": j},
						Dependencies:  []string{},
					}

					env.ExecuteWorkflow(ModuleInitializationWorkflow, input)

					if env.GetWorkflowError() != nil {
						b.Fatalf("Scale test workflow %d failed: %v", j, env.GetWorkflowError())
					}
				}
			}
		})
	}
}
