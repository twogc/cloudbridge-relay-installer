package integration

import (
	"context"
	"fmt"
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

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// CadenceIntegrationTestSuite представляет набор интеграционных тестов для Cadence
type CadenceIntegrationTestSuite struct {
	suite.Suite
	logger         *zap.Logger
	config         *cadence.CadenceConfig
	client         *cadence.Client
	worker         *worker.CadenceWorker
	cadenceMetrics *metrics.CadenceMetrics
	ctx            context.Context
	cancel         context.CancelFunc
}

// SetupSuite выполняется один раз перед всеми тестами
func (suite *CadenceIntegrationTestSuite) SetupSuite() {
	// Настройка логгера для тестов
	suite.logger = zaptest.NewLogger(suite.T())

	// Создание контекста
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 30*time.Minute)

	// Загрузка конфигурации для тестов
	var err error
	suite.config, err = suite.loadTestConfig()
	require.NoError(suite.T(), err, "Не удалось загрузить тестовую конфигурацию")

	// Создание Cadence метрик
	suite.cadenceMetrics = metrics.NewCadenceMetrics(suite.logger)

	// Создание Cadence клиента
	cadenceConfig := cadence.Config{
		HostPort: suite.config.Client.HostPort,
		Domain:   suite.config.Client.Domain,
		TaskList: suite.config.Worker.TaskListName,
	}
	suite.client, err = cadence.NewClient(cadenceConfig, suite.logger)
	require.NoError(suite.T(), err, "Не удалось создать Cadence клиент")

	// Запуск Cadence клиента
	err = suite.client.Start()
	require.NoError(suite.T(), err, "Не удалось запустить Cadence клиент")

	// Ожидание готовности клиента
	suite.waitForCadenceReady()

	// Создание и запуск worker
	suite.worker, err = worker.NewCadenceWorker(suite.client, suite.config, suite.logger, suite.cadenceMetrics)
	require.NoError(suite.T(), err, "Не удалось создать Cadence worker")

	err = suite.worker.Start()
	require.NoError(suite.T(), err, "Не удалось запустить Cadence worker")

	// Ожидание готовности worker
	suite.waitForWorkerReady()

	suite.logger.Info("Интеграционная тестовая среда готова")
}

// TearDownSuite выполняется один раз после всех тестов
func (suite *CadenceIntegrationTestSuite) TearDownSuite() {
	suite.logger.Info("Завершение интеграционного тестирования")

	// Остановка worker
	if suite.worker != nil {
		err := suite.worker.Stop()
		if err != nil {
			suite.logger.Warn("Ошибка при остановке worker", zap.Error(err))
		}
	}

	// Остановка клиента
	if suite.client != nil {
		suite.client.Stop() // Stop() не возвращает ошибку
	}

	// Отмена контекста
	suite.cancel()

	suite.logger.Info("Интеграционное тестирование завершено")
}

// TestModuleInitializationWorkflow тестирует workflow инициализации модулей
func (suite *CadenceIntegrationTestSuite) TestModuleInitializationWorkflow() {
	suite.logger.Info("Тестирование ModuleInitializationWorkflow")

	// Подготовка входных данных
	input := workflows.ModuleInitializationInput{
		ModuleName: "test-module",
		Configuration: map[string]interface{}{
			"test_param": "test_value",
			"timeout":    30,
		},
		Dependencies: []string{"dependency1", "dependency2"},
	}

	// Запуск workflow
	execution, err := suite.client.StartWorkflow(
		suite.ctx,
		"ModuleInitializationWorkflow",
		input,
	)
	require.NoError(suite.T(), err, "Не удалось запустить ModuleInitializationWorkflow")
	assert.NotEmpty(suite.T(), execution.ID, "ID workflow не должен быть пустым")

	// Ожидание завершения workflow
	var result workflows.ModuleInitializationResult
	err = suite.client.GetWorkflowResult(suite.ctx, execution.ID, execution.RunID, &result)
	require.NoError(suite.T(), err, "Не удалось получить результат workflow")

	// Проверка результата
	assert.Equal(suite.T(), "test-module", result.ModuleName)
	assert.Equal(suite.T(), "completed", result.Status)
	assert.NotEmpty(suite.T(), result.Message)
	assert.False(suite.T(), result.StartedAt.IsZero())
	assert.False(suite.T(), result.CompletedAt.IsZero())
	assert.True(suite.T(), result.CompletedAt.After(result.StartedAt))

	suite.logger.Info("ModuleInitializationWorkflow протестирован успешно",
		zap.String("workflow_id", execution.ID),
		zap.String("status", result.Status),
	)
}

// TestHealthCheckWorkflow тестирует workflow проверки здоровья
func (suite *CadenceIntegrationTestSuite) TestHealthCheckWorkflow() {
	suite.logger.Info("Тестирование HealthCheckWorkflow")

	// Создание контекста с коротким таймаутом для этого теста
	ctx, cancel := context.WithTimeout(suite.ctx, 30*time.Second)
	defer cancel()

	// Подготовка входных данных
	input := workflows.HealthCheckInput{
		Component: "test-component",
		Checks:    []string{"core", "network"},
		Config: map[string]interface{}{
			"check_timeout": "5s",
			"retries":       2,
			"test_mode":     true,
		},
	}

	// Запуск workflow
	execution, err := suite.client.StartWorkflow(ctx, "HealthCheckWorkflow", input)
	require.NoError(suite.T(), err, "Не удалось запустить HealthCheckWorkflow")

	// Даем время workflow запуститься и выполнить первую проверку
	time.Sleep(3 * time.Second)

	// Отправляем сигнал завершения workflow
	err = suite.client.SignalWorkflow(ctx, execution.ID, execution.RunID, "control-command", "shutdown")
	require.NoError(suite.T(), err, "Не удалось отправить сигнал завершения")

	// Ждем завершения workflow с таймаутом
	done := make(chan bool)
	var result workflows.HealthCheckResult
	go func() {
		suite.client.GetWorkflowResult(ctx, execution.ID, execution.RunID, &result)
		done <- true
	}()

	select {
	case <-done:
		suite.logger.Info("HealthCheckWorkflow завершился корректно")
	case <-time.After(10 * time.Second):
		suite.T().Fatal("HealthCheckWorkflow не завершился в ожидаемое время")
	}

	// Проверка результата
	assert.Equal(suite.T(), "test-component", result.Component)
	assert.Equal(suite.T(), "stopped", result.Status)

	suite.logger.Info("HealthCheckWorkflow протестирован успешно",
		zap.String("workflow_id", execution.ID),
		zap.String("component", result.Component),
		zap.String("status", result.Status),
	)
}

// TestPeerLifecycleWorkflow тестирует WireGuard workflow управления пирами
func (suite *CadenceIntegrationTestSuite) TestPeerLifecycleWorkflow() {
	suite.logger.Info("Тестирование PeerLifecycleWorkflow")

	// Подготовка входных данных для добавления пира
	input := workflows.PeerLifecycleWorkflowInput{
		PeerID:    "test-peer-12345",
		Operation: "add",
		PeerConfig: map[string]interface{}{
			"public_key":  "xTIBA5rboUvnH4htodjb6e697QjS4WSWgpfE5DBe4Fg=",
			"allowed_ips": []string{"10.0.0.0/24"},
			"endpoint":    "test.example.com:51820",
			"keepalive":   25,
		},
		NetworkConfig: map[string]interface{}{
			"network_id": "test-network",
			"subnet":     "10.0.0.0/16",
		},
	}

	// Запуск workflow для добавления пира
	execution, err := suite.client.StartWorkflow(suite.ctx, "PeerLifecycleWorkflow", input)
	require.NoError(suite.T(), err, "Не удалось запустить PeerLifecycleWorkflow")

	// Ожидание завершения workflow
	var result workflows.PeerLifecycleWorkflowResult
	err = suite.client.GetWorkflowResult(suite.ctx, execution.ID, execution.RunID, &result)
	require.NoError(suite.T(), err, "Не удалось получить результат workflow")

	// Проверка результата добавления
	assert.Equal(suite.T(), "test-peer-12345", result.PeerID)
	assert.Equal(suite.T(), "add", result.Operation)
	assert.Equal(suite.T(), "completed", result.Status)
	assert.NotEmpty(suite.T(), result.Result)

	suite.logger.Info("PeerLifecycleWorkflow (add) протестирован успешно",
		zap.String("peer_id", result.PeerID),
		zap.String("operation", result.Operation),
		zap.String("status", result.Status),
	)

	// Теперь тестируем удаление пира
	input.Operation = "remove"

	execution2, err := suite.client.StartWorkflow(suite.ctx, "PeerLifecycleWorkflow", input)
	require.NoError(suite.T(), err, "Не удалось запустить PeerLifecycleWorkflow для удаления")

	var result2 workflows.PeerLifecycleWorkflowResult
	err = suite.client.GetWorkflowResult(suite.ctx, execution2.ID, execution2.RunID, &result2)
	require.NoError(suite.T(), err, "Не удалось получить результат workflow удаления")

	// Проверка результата удаления
	assert.Equal(suite.T(), "test-peer-12345", result2.PeerID)
	assert.Equal(suite.T(), "remove", result2.Operation)
	assert.Equal(suite.T(), "completed", result2.Status)

	suite.logger.Info("PeerLifecycleWorkflow (remove) протестирован успешно",
		zap.String("peer_id", result2.PeerID),
		zap.String("operation", result2.Operation),
		zap.String("status", result2.Status),
	)
}

// TestQuantumKeyRotationWorkflow тестирует workflow ротации квантовых ключей
func (suite *CadenceIntegrationTestSuite) TestQuantumKeyRotationWorkflow() {
	suite.logger.Info("Тестирование QuantumKeyRotationWorkflow")

	// Подготовка входных данных
	input := workflows.QuantumKeyRotationWorkflowInput{
		KeyID:          "test-quantum-key-001",
		Algorithm:      "kyber",
		RotationPolicy: "time_based",
		Configuration: map[string]interface{}{
			"key_size":       192,
			"security_level": 3,
			"backup_enabled": true,
		},
		ForceRotation: true, // Принудительная ротация для теста
	}

	// Запуск workflow
	execution, err := suite.client.StartWorkflow(suite.ctx, "QuantumKeyRotationWorkflow", input)
	require.NoError(suite.T(), err, "Не удалось запустить QuantumKeyRotationWorkflow")

	// Ожидание завершения workflow
	var result workflows.QuantumKeyRotationWorkflowResult
	err = suite.client.GetWorkflowResult(suite.ctx, execution.ID, execution.RunID, &result)
	require.NoError(suite.T(), err, "Не удалось получить результат workflow")

	// Проверка результата
	assert.Equal(suite.T(), "test-quantum-key-001", result.OldKeyID)
	assert.NotEmpty(suite.T(), result.NewKeyID)
	assert.NotEqual(suite.T(), result.OldKeyID, result.NewKeyID)
	assert.Equal(suite.T(), "kyber", result.Algorithm)
	assert.Equal(suite.T(), "completed", result.Status)
	assert.NotEmpty(suite.T(), result.RotationInfo)

	suite.logger.Info("QuantumKeyRotationWorkflow протестирован успешно",
		zap.String("old_key_id", result.OldKeyID),
		zap.String("new_key_id", result.NewKeyID),
		zap.String("algorithm", result.Algorithm),
		zap.String("status", result.Status),
	)
}

// TestThreatAnalysisWorkflow тестирует AI workflow анализа угроз
func (suite *CadenceIntegrationTestSuite) TestThreatAnalysisWorkflow() {
	suite.logger.Info("Тестирование ThreatAnalysisWorkflow")

	// Подготовка входных данных
	input := workflows.ThreatAnalysisWorkflowInput{
		AnalysisType: "batch",
		DataSources:  []string{"logs", "network", "behavior"},
		TimeWindow:   "1h",
		ThreatModels: []string{"ddos", "intrusion", "anomaly"},
		Configuration: map[string]interface{}{
			"ai_model_version":     "v1.0",
			"confidence_threshold": 0.8,
			"alert_enabled":        true,
		},
	}

	// Запуск workflow
	execution, err := suite.client.StartWorkflow(suite.ctx, "ThreatAnalysisWorkflow", input)
	require.NoError(suite.T(), err, "Не удалось запустить ThreatAnalysisWorkflow")

	// Ожидание завершения workflow
	var result workflows.ThreatAnalysisWorkflowResult
	err = suite.client.GetWorkflowResult(suite.ctx, execution.ID, execution.RunID, &result)
	require.NoError(suite.T(), err, "Не удалось получить результат workflow")

	// Проверка результата
	assert.Equal(suite.T(), "batch", result.AnalysisType)
	assert.Contains(suite.T(), []string{"clean", "threats_detected"}, result.Status)
	assert.Contains(suite.T(), []string{"low", "medium", "high", "critical", "unknown"}, result.RiskLevel)
	assert.NotNil(suite.T(), result.DetectedThreats)
	assert.NotNil(suite.T(), result.ThreatScores)
	assert.NotNil(suite.T(), result.AnalysisResults)

	suite.logger.Info("ThreatAnalysisWorkflow протестирован успешно",
		zap.String("analysis_type", result.AnalysisType),
		zap.String("status", result.Status),
		zap.String("risk_level", result.RiskLevel),
		zap.Int("detected_threats", len(result.DetectedThreats)),
	)
}

// TestWorkflowMetrics тестирует сбор метрик workflows
func (suite *CadenceIntegrationTestSuite) TestWorkflowMetrics() {
	suite.logger.Info("Тестирование метрик Cadence")

	// Запускаем простой workflow для генерации метрик
	input := workflows.ModuleInitializationInput{
		ModuleName:    "metrics-test-module",
		Configuration: map[string]interface{}{},
		Dependencies:  []string{},
	}

	execution, err := suite.client.StartWorkflow(suite.ctx, "ModuleInitializationWorkflow", input)
	require.NoError(suite.T(), err)

	var result workflows.ModuleInitializationResult
	err = suite.client.GetWorkflowResult(suite.ctx, execution.ID, execution.RunID, &result)
	require.NoError(suite.T(), err)

	// Проверяем, что метрики записываются
	// В реальной реализации здесь должна быть проверка Prometheus метрик
	// Пока проверяем, что workflow завершился успешно
	assert.Equal(suite.T(), "completed", result.Status)

	suite.logger.Info("Метрики Cadence протестированы",
		zap.String("workflow_id", execution.ID),
	)
}

// TestAutoScalingWorkflowWithTimeout тестирует AutoScaling workflow с таймаутом
func (suite *CadenceIntegrationTestSuite) TestAutoScalingWorkflowWithTimeout() {
	suite.logger.Info("Тестирование AutoScalingWorkflow с таймаутом")

	// Создание контекста с коротким таймаутом
	ctx, cancel := context.WithTimeout(suite.ctx, 30*time.Second)
	defer cancel()

	// Подготовка входных данных
	input := workflows.AutoScalingWorkflowInput{
		ResourceType:  "compute",
		ScalingPolicy: "reactive",
		TargetMetrics: map[string]float64{"cpu": 70, "memory": 80},
		ScalingLimits: map[string]interface{}{"min_instances": 1, "max_instances": 5},
		Configuration: map[string]interface{}{
			"test_mode":       true,
			"short_intervals": true,
		},
	}

	// Запуск workflow
	execution, err := suite.client.StartWorkflow(ctx, "AutoScalingWorkflow", input)
	require.NoError(suite.T(), err, "Не удалось запустить AutoScalingWorkflow")

	// Ожидание запуска
	time.Sleep(2 * time.Second)

	// Отправка сигнала остановки
	err = suite.client.SignalWorkflow(ctx, execution.ID, execution.RunID, "control-command", "shutdown")
	require.NoError(suite.T(), err, "Не удалось отправить сигнал остановки")

	// Проверка завершения в разумное время
	done := make(chan bool)
	var result workflows.AutoScalingWorkflowResult
	go func() {
		suite.client.GetWorkflowResult(ctx, execution.ID, execution.RunID, &result)
		done <- true
	}()

	select {
	case <-done:
		suite.logger.Info("AutoScaling workflow завершился корректно")
		assert.Equal(suite.T(), "compute", result.ResourceType)
		assert.Equal(suite.T(), "reactive", result.ScalingPolicy)
		assert.Equal(suite.T(), "stopped", result.Status)
	case <-time.After(10 * time.Second):
		suite.T().Fatal("AutoScaling workflow не завершился в ожидаемое время")
	}

	suite.logger.Info("AutoScalingWorkflow протестирован успешно",
		zap.String("workflow_id", execution.ID),
		zap.String("resource_type", result.ResourceType),
		zap.String("status", result.Status),
	)
}

// TestWorkflowError тестирует обработку ошибок в workflows
func (suite *CadenceIntegrationTestSuite) TestWorkflowError() {
	suite.logger.Info("Тестирование обработки ошибок в workflows")

	// Создаем входные данные, которые вызовут ошибку
	input := workflows.PeerLifecycleWorkflowInput{
		PeerID:    "", // Пустой PeerID должен вызвать ошибку
		Operation: "add",
		PeerConfig: map[string]interface{}{
			"invalid_config": true,
		},
	}

	// Запуск workflow
	execution, err := suite.client.StartWorkflow(suite.ctx, "PeerLifecycleWorkflow", input)
	require.NoError(suite.T(), err)

	// Ожидаем, что workflow завершится с ошибкой
	var result workflows.PeerLifecycleWorkflowResult
	err = suite.client.GetWorkflowResult(suite.ctx, execution.ID, execution.RunID, &result)

	// Проверяем, что ошибка обработана корректно
	if err != nil {
		suite.logger.Info("Ошибка обработана корректно", zap.Error(err))
	} else {
		// Если workflow завершился без ошибки, проверяем статус результата
		assert.Contains(suite.T(), []string{"failed", "error"}, result.Status)
		suite.logger.Info("Workflow завершился с ошибкой в результате",
			zap.String("status", result.Status),
			zap.String("message", result.Message),
		)
	}
}

// TestConcurrentWorkflows тестирует параллельное выполнение workflows
func (suite *CadenceIntegrationTestSuite) TestConcurrentWorkflows() {
	suite.logger.Info("Тестирование параллельного выполнения workflows")

	const numWorkflows = 5
	executions := make([]struct{ ID, RunID string }, numWorkflows)

	// Запускаем несколько workflows параллельно
	for i := 0; i < numWorkflows; i++ {
		input := workflows.ModuleInitializationInput{
			ModuleName:    fmt.Sprintf("concurrent-test-module-%d", i),
			Configuration: map[string]interface{}{"index": i},
			Dependencies:  []string{},
		}

		execution, err := suite.client.StartWorkflow(suite.ctx, "ModuleInitializationWorkflow", input)
		require.NoError(suite.T(), err, "Не удалось запустить workflow %d", i)

		executions[i] = struct{ ID, RunID string }{execution.ID, execution.RunID}
	}

	// Ждем завершения всех workflows
	for i, execution := range executions {
		var result workflows.ModuleInitializationResult
		err := suite.client.GetWorkflowResult(suite.ctx, execution.ID, execution.RunID, &result)
		require.NoError(suite.T(), err, "Не удалось получить результат workflow %d", i)
		assert.Equal(suite.T(), "completed", result.Status)

		expectedModuleName := fmt.Sprintf("concurrent-test-module-%d", i)
		assert.Equal(suite.T(), expectedModuleName, result.ModuleName)
	}

	suite.logger.Info("Параллельное выполнение workflows протестировано успешно",
		zap.Int("num_workflows", numWorkflows),
	)
}

// Helper methods

// loadTestConfig загружает конфигурацию для тестов
func (suite *CadenceIntegrationTestSuite) loadTestConfig() (*cadence.CadenceConfig, error) {
	// Попытка загрузить конфигурацию из файла
	configPath := os.Getenv("CADENCE_TEST_CONFIG")
	if configPath == "" {
		configPath = "../../configs/cadence.yaml"
	}

	if _, err := os.Stat(configPath); err == nil {
		return cadence.LoadConfig(configPath)
	}

	// Fallback на тестовую конфигурацию по умолчанию
	suite.logger.Warn("Используется тестовая конфигурация по умолчанию")

	config := &cadence.CadenceConfig{}
	config.Domain = "cloudbridge-relay-test"
	config.ServiceName = "cloudbridge-cadence-test"
	config.Client.HostPort = "localhost:7933"
	config.Client.Domain = "cloudbridge-relay-test"
	config.Client.DataConverter = "json"
	config.Worker.TaskListName = "cloudbridge-test-task-list"
	config.Worker.Identity = "cloudbridge-test-worker"
	config.SetDefaults()

	return config, nil
}

// waitForCadenceReady ждет готовности Cadence сервера
func (suite *CadenceIntegrationTestSuite) waitForCadenceReady() {
	suite.logger.Info("Ожидание готовности Cadence сервера")

	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		if suite.client.IsRunning() {
			suite.logger.Info("Cadence сервер готов")
			return
		}

		suite.logger.Debug("Ожидание Cadence сервера", zap.Int("attempt", i+1))
		time.Sleep(2 * time.Second)
	}

	suite.T().Fatal("Cadence сервер не готов после ожидания")
}

// waitForWorkerReady ждет готовности Cadence worker
func (suite *CadenceIntegrationTestSuite) waitForWorkerReady() {
	suite.logger.Info("Ожидание готовности Cadence worker")

	maxAttempts := 15
	for i := 0; i < maxAttempts; i++ {
		if suite.worker.IsRunning() {
			suite.logger.Info("Cadence worker готов")
			time.Sleep(2 * time.Second) // Дополнительное время для регистрации
			return
		}

		suite.logger.Debug("Ожидание Cadence worker", zap.Int("attempt", i+1))
		time.Sleep(2 * time.Second)
	}

	suite.T().Fatal("Cadence worker не готов после ожидания")
}

// TestCadenceIntegration запускает все интеграционные тесты
func TestCadenceIntegration(t *testing.T) {
	// Проверяем, что интеграционные тесты включены
	if testing.Short() {
		t.Skip("Пропуск интеграционных тестов в коротком режиме")
	}

	// Проверяем доступность Cadence сервера
	if os.Getenv("SKIP_CADENCE_TESTS") == "true" {
		t.Skip("Cadence тесты отключены через переменную окружения")
	}

	suite.Run(t, new(CadenceIntegrationTestSuite))
}
