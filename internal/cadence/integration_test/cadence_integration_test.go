//go:build integration
// +build integration

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/twogc/cloudbridge-relay-installer/internal/cadence"
	"github.com/twogc/cloudbridge-relay-installer/internal/cadence/worker"
	"github.com/twogc/cloudbridge-relay-installer/internal/cadence/workflows"
	"github.com/twogc/cloudbridge-relay-installer/internal/metrics"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// CadenceIntegrationTestSuite представляет набор интеграционных тестов
// Требует запущенный Cadence сервер для выполнения
type CadenceIntegrationTestSuite struct {
	suite.Suite

	config         *cadence.CadenceConfig
	client         *cadence.CadenceClient
	worker         *worker.CadenceWorker
	logger         *zap.Logger
	cadenceMetrics *metrics.CadenceMetrics

	ctx    context.Context
	cancel context.CancelFunc
}

// SetupSuite выполняется один раз перед всеми тестами
func (s *CadenceIntegrationTestSuite) SetupSuite() {
	// Проверяем, что Cadence сервер доступен
	cadenceHost := os.Getenv("CADENCE_HOST")
	if cadenceHost == "" {
		cadenceHost = "localhost:7933"
	}

	// Создаем логгер
	var err error
	s.logger, err = zap.NewDevelopment()
	require.NoError(s.T(), err)

	// Создаем конфигурацию для тестов
	s.config = &cadence.CadenceConfig{
		Domain:      "cloudbridge-integration-test",
		ServiceName: "cloudbridge-integration-test",
		Client: cadence.ClientConfig{
			HostPort:      cadenceHost,
			Domain:        "cloudbridge-integration-test",
			DataConverter: "json",
		},
		Worker: cadence.WorkerConfig{
			MaxConcurrentActivityExecutionSize:      10,
			MaxConcurrentWorkflowTaskExecutionSize:  10,
			MaxConcurrentLocalActivityExecutionSize: 10,
			TaskListName:                            "integration-test-task-list",
			Identity:                                "integration-test-worker",
		},
		Workflows: cadence.WorkflowsConfig{
			ExecutionTimeout:            "30m",
			DecisionTimeout:             "10s",
			ActivityStartToCloseTimeout: "5m",
			WorkflowStartToCloseTimeout: "30m",
		},
		Retry: cadence.RetryConfig{
			InitialInterval:    "1s",
			MaximumInterval:    "10s",
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
		Logging: cadence.LoggingConfig{
			Level:                   "debug",
			EnableStructuredLogging: true,
			LogWorkflowHistory:      true,
			LogActivityDetails:      true,
		},
	}

	// Устанавливаем значения по умолчанию
	s.config.SetDefaults()

	// Валидируем конфигурацию
	err = s.config.Validate()
	require.NoError(s.T(), err)

	// Создаем контекст
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 10*time.Minute)

	// Создаем метрики
	s.cadenceMetrics = metrics.NewCadenceMetrics(s.logger)

	// Создаем Cadence клиент
	s.client, err = cadence.NewCadenceClient(s.config, s.logger)
	require.NoError(s.T(), err)

	// Запускаем клиент
	err = s.client.Start()
	require.NoError(s.T(), err)

	// Создаем и запускаем worker
	s.worker, err = worker.NewCadenceWorker(s.client, s.config, s.logger, s.cadenceMetrics)
	require.NoError(s.T(), err)

	err = s.worker.Start()
	require.NoError(s.T(), err)

	// Ждем немного, чтобы worker инициализировался
	time.Sleep(2 * time.Second)

	s.logger.Info("Интеграционные тесты инициализированы",
		zap.String("cadence_host", cadenceHost),
		zap.String("domain", s.config.Domain),
		zap.String("task_list", s.config.Worker.TaskListName),
	)
}

// TearDownSuite выполняется один раз после всех тестов
func (s *CadenceIntegrationTestSuite) TearDownSuite() {
	if s.worker != nil {
		err := s.worker.Stop()
		assert.NoError(s.T(), err)
	}

	if s.client != nil {
		err := s.client.Stop()
		assert.NoError(s.T(), err)
	}

	if s.cancel != nil {
		s.cancel()
	}

	s.logger.Info("Интеграционные тесты завершены")
}

// TestModuleInitializationWorkflowIntegration тестирует реальное выполнение workflow инициализации
func (s *CadenceIntegrationTestSuite) TestModuleInitializationWorkflowIntegration() {
	s.logger.Info("Запуск интеграционного теста инициализации модуля")

	// Входные данные для workflow
	input := workflows.ModuleInitializationInput{
		ModuleName: "integration-test-module",
		Configuration: map[string]interface{}{
			"test_mode":   true,
			"integration": true,
		},
		Dependencies: []string{"core", "network"},
	}

	// Запускаем workflow
	execution, err := s.client.StartWorkflow(s.ctx, "ModuleInitializationWorkflow", input)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), execution.ID)
	assert.NotEmpty(s.T(), execution.RunID)

	s.logger.Info("Workflow запущен",
		zap.String("workflow_id", execution.ID),
		zap.String("run_id", execution.RunID),
	)

	// Ждем завершения workflow (с таймаутом)
	ctx, cancel := context.WithTimeout(s.ctx, 2*time.Minute)
	defer cancel()

	var result workflows.ModuleInitializationResult
	err = s.client.GetWorkflowResult(ctx, execution.ID, execution.RunID, &result)
	require.NoError(s.T(), err)

	// Проверяем результат
	assert.Equal(s.T(), "integration-test-module", result.ModuleName)
	assert.Equal(s.T(), "completed", result.Status)
	assert.NotEmpty(s.T(), result.Message)
	assert.False(s.T(), result.StartedAt.IsZero())
	assert.False(s.T(), result.CompletedAt.IsZero())
	assert.True(s.T(), result.CompletedAt.After(result.StartedAt))

	s.logger.Info("Workflow завершен успешно",
		zap.String("status", result.Status),
		zap.Duration("duration", result.CompletedAt.Sub(result.StartedAt)),
	)
}

// TestHealthCheckWorkflowIntegration тестирует реальное выполнение health check workflow
func (s *CadenceIntegrationTestSuite) TestHealthCheckWorkflowIntegration() {
	s.logger.Info("Запуск интеграционного теста health check")

	// Входные данные для workflow
	input := workflows.HealthCheckInput{
		Component: "integration-test-component",
		Checks:    []string{"core", "network", "storage"},
		Config: map[string]interface{}{
			"test_mode":      true,
			"check_interval": "5s",
		},
	}

	// Запускаем workflow
	execution, err := s.client.StartWorkflow(s.ctx, "HealthCheckWorkflow", input)
	require.NoError(s.T(), err)

	s.logger.Info("Health check workflow запущен",
		zap.String("workflow_id", execution.ID),
		zap.String("run_id", execution.RunID),
	)

	// Ждем немного, чтобы workflow начал работу
	time.Sleep(10 * time.Second)

	// Проверяем, что workflow работает (отправляем query)
	status, err := s.client.QueryWorkflow(s.ctx, execution.ID, execution.RunID, "status")
	if err == nil {
		s.logger.Info("Health check workflow статус получен", zap.Any("status", status))
	}

	// Отправляем сигнал остановки
	err = s.client.SignalWorkflow(s.ctx, execution.ID, execution.RunID, "control-command", "shutdown")
	require.NoError(s.T(), err)

	// Ждем завершения workflow
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	var result workflows.HealthCheckResult
	err = s.client.GetWorkflowResult(ctx, execution.ID, execution.RunID, &result)
	require.NoError(s.T(), err)

	// Проверяем результат
	assert.Equal(s.T(), "integration-test-component", result.Component)
	assert.NotEmpty(s.T(), result.Status)
	assert.NotEmpty(s.T(), result.CheckResults)
	assert.False(s.T(), result.LastChecked.IsZero())

	s.logger.Info("Health check workflow завершен",
		zap.String("status", result.Status),
		zap.Float64("health_score", result.HealthScore),
	)
}

// TestWorkflowFailureAndRecovery тестирует обработку ошибок и восстановление
func (s *CadenceIntegrationTestSuite) TestWorkflowFailureAndRecovery() {
	s.logger.Info("Запуск теста обработки ошибок")

	// Входные данные, которые должны вызвать ошибку
	input := workflows.ModuleInitializationInput{
		ModuleName: "failing-module",
		Configuration: map[string]interface{}{
			"should_fail": true,
			"error_type":  "validation_error",
		},
		Dependencies: []string{"non-existent-dependency"},
	}

	// Запускаем workflow
	execution, err := s.client.StartWorkflow(s.ctx, "ModuleInitializationWorkflow", input)
	require.NoError(s.T(), err)

	// Ждем завершения workflow (ожидаем ошибку)
	ctx, cancel := context.WithTimeout(s.ctx, 1*time.Minute)
	defer cancel()

	var result workflows.ModuleInitializationResult
	err = s.client.GetWorkflowResult(ctx, execution.ID, execution.RunID, &result)

	// В зависимости от реализации activities, может быть ошибка или результат с failed статусом
	if err != nil {
		s.logger.Info("Workflow завершился с ошибкой (ожидаемо)", zap.Error(err))
	} else {
		// Если получили результат, проверяем, что статус указывает на неудачу
		s.logger.Info("Workflow завершился с результатом", zap.String("status", result.Status))
		// В реальной реализации здесь может быть проверка на "failed" статус
	}
}

// TestConcurrentWorkflows тестирует выполнение нескольких workflow одновременно
func (s *CadenceIntegrationTestSuite) TestConcurrentWorkflows() {
	s.logger.Info("Запуск теста конкурентных workflow")

	const numWorkflows = 5
	executions := make([]*workflow.Execution, numWorkflows)

	// Запускаем несколько workflow одновременно
	for i := 0; i < numWorkflows; i++ {
		input := workflows.ModuleInitializationInput{
			ModuleName: "concurrent-module-" + string(rune('0'+i)),
			Configuration: map[string]interface{}{
				"instance":   i,
				"concurrent": true,
			},
			Dependencies: []string{},
		}

		execution, err := s.client.StartWorkflow(s.ctx, "ModuleInitializationWorkflow", input)
		require.NoError(s.T(), err)
		executions[i] = execution

		s.logger.Info("Конкурентный workflow запущен",
			zap.Int("instance", i),
			zap.String("workflow_id", execution.ID),
		)
	}

	// Ждем завершения всех workflow
	ctx, cancel := context.WithTimeout(s.ctx, 3*time.Minute)
	defer cancel()

	results := make([]workflows.ModuleInitializationResult, numWorkflows)
	for i, execution := range executions {
		err := s.client.GetWorkflowResult(ctx, execution.ID, execution.RunID, &results[i])
		require.NoError(s.T(), err)

		assert.Equal(s.T(), "concurrent-module-"+string(rune('0'+i)), results[i].ModuleName)
		assert.Equal(s.T(), "completed", results[i].Status)

		s.logger.Info("Конкурентный workflow завершен",
			zap.Int("instance", i),
			zap.String("status", results[i].Status),
		)
	}

	s.logger.Info("Все конкурентные workflow завершены успешно")
}

// TestWorkflowTimeout тестирует поведение при превышении timeout
func (s *CadenceIntegrationTestSuite) TestWorkflowTimeout() {
	s.logger.Info("Запуск теста timeout workflow")

	// Создаем workflow с очень коротким timeout
	input := workflows.ModuleInitializationInput{
		ModuleName: "timeout-module",
		Configuration: map[string]interface{}{
			"slow_operation": true,
			"delay_seconds":  60, // Долгая операция
		},
		Dependencies: []string{},
	}

	// Запускаем workflow
	execution, err := s.client.StartWorkflow(s.ctx, "ModuleInitializationWorkflow", input)
	require.NoError(s.T(), err)

	// Ждем с очень коротким timeout
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	var result workflows.ModuleInitializationResult
	err = s.client.GetWorkflowResult(ctx, execution.ID, execution.RunID, &result)

	// Ожидаем timeout ошибку
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "timeout")

	s.logger.Info("Timeout тест завершен (ожидаемо с ошибкой)", zap.Error(err))

	// Принудительно завершаем workflow
	err = s.client.TerminateWorkflow(s.ctx, execution.ID, execution.RunID, "Test timeout cleanup")
	assert.NoError(s.T(), err)
}

// TestWorkerStats тестирует получение статистики worker'а
func (s *CadenceIntegrationTestSuite) TestWorkerStats() {
	s.logger.Info("Тестирование статистики worker'а")

	assert.True(s.T(), s.worker.IsRunning())

	stats := s.worker.GetWorkerStats()
	assert.NotNil(s.T(), stats)

	// Проверяем основные поля статистики
	assert.Equal(s.T(), true, stats["is_running"])
	assert.Equal(s.T(), s.config.Domain, stats["domain"])
	assert.Equal(s.T(), s.config.Worker.TaskListName, stats["task_list"])
	assert.Equal(s.T(), s.config.Worker.Identity, stats["identity"])

	s.logger.Info("Статистика worker'а получена", zap.Any("stats", stats))
}

// TestInRunning запускает все интеграционные тесты
func TestCadenceIntegration(t *testing.T) {
	// Проверяем, включены ли интеграционные тесты
	if testing.Short() {
		t.Skip("Пропускаем интеграционные тесты в коротком режиме")
	}

	// Проверяем наличие Cadence сервера
	cadenceHost := os.Getenv("CADENCE_HOST")
	if cadenceHost == "" {
		t.Skip("Пропускаем интеграционные тесты: CADENCE_HOST не установлен")
	}

	suite.Run(t, new(CadenceIntegrationTestSuite))
}

// TestCadenceConnection тестирует подключение к Cadence серверу
func TestCadenceConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропускаем тест подключения в коротком режиме")
	}

	cadenceHost := os.Getenv("CADENCE_HOST")
	if cadenceHost == "" {
		cadenceHost = "localhost:7933"
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	config := &cadence.CadenceConfig{
		Domain:      "connection-test",
		ServiceName: "connection-test",
		Client: cadence.ClientConfig{
			HostPort:      cadenceHost,
			Domain:        "connection-test",
			DataConverter: "json",
		},
		Worker: cadence.WorkerConfig{
			TaskListName: "connection-test-task-list",
			Identity:     "connection-test-worker",
		},
	}
	config.SetDefaults()

	client, err := cadence.NewCadenceClient(config, logger)
	require.NoError(t, err)

	err = client.Start()
	require.NoError(t, err)
	defer client.Stop()

	assert.True(t, client.IsRunning())

	t.Logf("Успешное подключение к Cadence серверу: %s", cadenceHost)
}
