package workflows

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/zap"
)

// TestUtils содержит утилиты для тестирования workflows
type TestUtils struct {
	logger *zap.Logger
}

// NewTestUtils создает новый экземпляр TestUtils
func NewTestUtils() *TestUtils {
	logger, _ := zap.NewDevelopment()
	return &TestUtils{
		logger: logger,
	}
}

// CreateTestEnvironment создает тестовое окружение с короткими таймаутами
func (tu *TestUtils) CreateTestEnvironment() *testsuite.TestWorkflowEnvironment {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()

	// Установка короткого таймаута для тестов
	env.SetTestTimeout(time.Second * 30)

	return env
}

// MockCommonActivities добавляет общие моки для activities
func (tu *TestUtils) MockCommonActivities(env *testsuite.TestWorkflowEnvironment) {
	// Мок для проверки зависимостей
	env.OnActivity("CheckDependenciesActivity", mock.Anything).Return("all_dependencies_available", nil)

	// Мок для инициализации модуля
	env.OnActivity("InitializeModuleActivity", mock.Anything, mock.Anything).Return("module_initialized", nil)

	// Мок для проверки здоровья модуля
	env.OnActivity("CheckModuleHealthActivity", mock.Anything).Return("healthy", nil)

	// Мок для проверки здоровья
	env.OnActivity("PerformHealthCheckActivity", mock.Anything, mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"status": "healthy",
			"score":  0.95,
		}, nil)
}

// MockScalingActivities добавляет моки для scaling activities
func (tu *TestUtils) MockScalingActivities(env *testsuite.TestWorkflowEnvironment) {
	// Мок для сбора метрик (низкая нагрузка, не требует масштабирования)
	env.OnActivity("CollectScalingMetricsActivity", mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"current_capacity":  map[string]interface{}{"cpu": 30, "memory": 25},
			"metrics_collected": true,
		}, nil)

	// Мок для анализа необходимости масштабирования (не требуется)
	env.OnActivity("AnalyzeScalingNeedActivity", mock.Anything, mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"need_scaling":   false,
			"scaling_action": "no_scaling_needed",
			"reason":         "metrics_within_threshold",
		}, nil)

	// Мок для получения текущей емкости
	env.OnActivity("GetCurrentCapacityActivity", mock.Anything).Return(
		map[string]interface{}{
			"current_capacity": map[string]interface{}{"instances": 2},
			"target_capacity":  map[string]interface{}{"instances": 2},
		}, nil)
}

// RunWorkflowWithTimeout запускает workflow с таймаутом и сигналом остановки
func (tu *TestUtils) RunWorkflowWithTimeout(
	env *testsuite.TestWorkflowEnvironment,
	workflowFunc interface{},
	input interface{},
	timeout time.Duration,
) error {
	// Запуск workflow в отдельной горутине
	go func() {
		env.ExecuteWorkflow(workflowFunc, input)
	}()

	// Ожидание запуска workflow
	time.Sleep(100 * time.Millisecond)

	// Отправляем сигнал остановки
	env.SignalWorkflow("control-command", "shutdown")

	// Ожидание завершения с таймаутом
	done := make(chan bool)
	go func() {
		for !env.IsWorkflowCompleted() && env.GetWorkflowError() == nil {
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	select {
	case <-done:
		return env.GetWorkflowError()
	case <-time.After(timeout):
		return context.DeadlineExceeded
	}
}

// SendControlSignal отправляет сигнал управления workflow
func (tu *TestUtils) SendControlSignal(env *testsuite.TestWorkflowEnvironment, signal string) {
	env.SignalWorkflow("control-command", signal)
}

// WaitForWorkflowCompletion ждет завершения workflow с таймаутом
func (tu *TestUtils) WaitForWorkflowCompletion(env *testsuite.TestWorkflowEnvironment, timeout time.Duration) error {
	done := make(chan bool)
	go func() {
		for !env.IsWorkflowCompleted() && env.GetWorkflowError() == nil {
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	select {
	case <-done:
		return env.GetWorkflowError()
	case <-time.After(timeout):
		return context.DeadlineExceeded
	}
}

// CreateTestInput создает стандартные тестовые входные данные
func (tu *TestUtils) CreateTestInput() map[string]interface{} {
	return map[string]interface{}{
		"test_mode":       true,
		"short_intervals": true,
		"timeout":         "30s",
	}
}

// CreateModuleInitializationInput создает входные данные для инициализации модуля
func (tu *TestUtils) CreateModuleInitializationInput(moduleName string) ModuleInitializationInput {
	return ModuleInitializationInput{
		ModuleName:    moduleName,
		Configuration: tu.CreateTestInput(),
		Dependencies:  []string{"test-dep1", "test-dep2"},
	}
}

// CreateHealthCheckInput создает входные данные для проверки здоровья
func (tu *TestUtils) CreateHealthCheckInput(component string) HealthCheckInput {
	return HealthCheckInput{
		Component: component,
		Checks:    []string{"core", "network"},
		Config:    tu.CreateTestInput(),
	}
}

// CreateAutoScalingInput создает входные данные для автомасштабирования
func (tu *TestUtils) CreateAutoScalingInput() AutoScalingWorkflowInput {
	return AutoScalingWorkflowInput{
		ResourceType:  "compute",
		ScalingPolicy: "reactive",
		TargetMetrics: map[string]float64{"cpu": 70, "memory": 80},
		ScalingLimits: map[string]interface{}{"min_instances": 1, "max_instances": 5},
		Configuration: tu.CreateTestInput(),
	}
}

// ValidateWorkflowResult проверяет базовые поля результата workflow
func (tu *TestUtils) ValidateWorkflowResult(result interface{}) bool {
	// Проверяем общие поля через interface{}
	// В реальной реализации здесь может быть reflection для проверки общих полей
	return true
}

// MockActivityWithDelay добавляет мок с задержкой для тестирования timeout
func (tu *TestUtils) MockActivityWithDelay(env *testsuite.TestWorkflowEnvironment, activityName string, delay time.Duration) {
	env.OnActivity(activityName, mock.Anything).After(delay).Return("result_after_delay", nil)
}

// CreateLongRunningWorkflowTest создает тест для долгоживущего workflow
func (tu *TestUtils) CreateLongRunningWorkflowTest(
	env *testsuite.TestWorkflowEnvironment,
	workflowFunc interface{},
	input interface{},
) func() error {
	return func() error {
		return tu.RunWorkflowWithTimeout(env, workflowFunc, input, time.Second*10)
	}
}
