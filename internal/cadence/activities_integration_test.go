package cadence

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// logContains проверяет, содержит ли какой-либо лог указанную подстроку
func logContains(logs *observer.ObservedLogs, sub string) bool {
	for _, e := range logs.All() {
		if strings.Contains(e.Message, sub) {
			return true
		}
	}
	return false
}

func TestValidateConfigActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(ValidateConfigActivity)

	// Вход
	config := map[string]interface{}{"host": "localhost", "port": 8080}

	// Запуск по функции
	fut, err := env.ExecuteActivity(ValidateConfigActivity, config)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result ConfigValidationResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Valid)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Validating relay configuration"), "expected log 'Validating relay configuration'")
}

func TestSetupWireGuardActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(SetupWireGuardActivity)

	// Вход
	config := map[string]interface{}{
		"interface": "wg0",
		"port":      51820,
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(SetupWireGuardActivity, config)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result WireGuardSetupResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Equal(t, "wg0", result.Config)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Setting up WireGuard"), "expected log 'Setting up WireGuard'")
}

func TestSetupQUICServerActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(SetupQUICServerActivity)

	// Вход
	config := map[string]interface{}{
		"port": 9090,
		"cert": "/path/to/cert.pem",
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(SetupQUICServerActivity, config)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result QUICSetupResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Equal(t, 9090, result.Port)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Setting up QUIC server"), "expected log 'Setting up QUIC server'")
}

func TestStartServicesActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(StartServicesActivity)

	// Вход
	services := []string{"relay", "wireguard", "quic"}

	// Запуск по функции
	fut, err := env.ExecuteActivity(StartServicesActivity, services)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result ServiceStartResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Equal(t, services, result.Services)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Starting services"), "expected log 'Starting services'")
}

func TestHealthCheckActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(HealthCheckActivity)

	// Вход
	input := HealthCheckInput{
		Services: []string{"relay", "wireguard", "quic"},
		Timeout:  30 * time.Second,
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(HealthCheckActivity, input)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result HealthCheckResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Healthy)
	assert.Len(t, result.Services, 3)
	assert.True(t, result.Services["relay"])
	assert.True(t, result.Services["wireguard"])
	assert.True(t, result.Services["quic"])
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Performing health check"), "expected log 'Performing health check'")
}

func TestGenerateCertificatesActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(GenerateCertificatesActivity)

	// Вход
	config := map[string]interface{}{
		"domain": "example.com",
		"days":   365,
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(GenerateCertificatesActivity, config)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result CertificateGenerationResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Len(t, result.Certificates, 2)
	assert.Contains(t, result.Certificates, "/etc/ssl/certs/relay.crt")
	assert.Contains(t, result.Certificates, "/etc/ssl/private/relay.key")
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Generating certificates"), "expected log 'Generating certificates'")
}

func TestValidateCertificatesActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(ValidateCertificatesActivity)

	// Вход
	certificates := []string{"/etc/ssl/certs/relay.crt", "/etc/ssl/private/relay.key"}

	// Запуск по функции
	fut, err := env.ExecuteActivity(ValidateCertificatesActivity, certificates)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result CertificateValidationResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Valid)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Validating certificates"), "expected log 'Validating certificates'")
}

func TestUpdateCertificateConfigActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(UpdateCertificateConfigActivity)

	// Вход
	input := UpdateCertificateConfigInput{
		Certificates: []string{"/etc/ssl/certs/relay.crt"},
		Services:     []string{"relay", "quic"},
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(UpdateCertificateConfigActivity, input)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result ConfigUpdateResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Updating certificate configuration"), "expected log 'Updating certificate configuration'")
}

func TestRestartServicesActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(RestartServicesActivity)

	// Вход
	services := []string{"relay", "wireguard"}

	// Запуск по функции
	fut, err := env.ExecuteActivity(RestartServicesActivity, services)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result ServiceRestartResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Equal(t, services, result.Services)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Restarting services"), "expected log 'Restarting services'")
}

func TestAnalyzeNetworkActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(AnalyzeNetworkActivity)

	// Вход
	config := map[string]interface{}{
		"interfaces": []string{"eth0", "wg0"},
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(AnalyzeNetworkActivity, config)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result NetworkAnalysisResult
	require.NoError(t, fut.Get(&result))
	assert.Empty(t, result.Issues)
	assert.Contains(t, result.Metrics, "latency")
	assert.Contains(t, result.Metrics, "bandwidth")
	assert.Equal(t, json.Number("10"), result.Metrics["latency"])
	assert.Equal(t, json.Number("1000"), result.Metrics["bandwidth"])
	assert.Contains(t, result.Recommendations, "Network is healthy")

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Analyzing network state"), "expected log 'Analyzing network state'")
}

func TestPlanMaintenanceActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(PlanMaintenanceActivity)

	// Вход
	analysis := NetworkAnalysisResult{
		Issues:          []string{},
		Metrics:         map[string]interface{}{"latency": 10},
		Recommendations: []string{"Network is healthy"},
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(PlanMaintenanceActivity, analysis)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result MaintenancePlanResult
	require.NoError(t, fut.Get(&result))
	assert.Empty(t, result.Tasks)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Planning maintenance tasks"), "expected log 'Planning maintenance tasks'")
}

func TestExecuteMaintenanceTaskActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(ExecuteMaintenanceTaskActivity)

	// Вход
	task := MaintenanceTask{
		Name:     "test-task",
		Type:     "network",
		Config:   map[string]interface{}{"timeout": 30},
		Priority: 1,
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(ExecuteMaintenanceTaskActivity, task)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result MaintenanceTaskResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Equal(t, "Task test-task completed", result.Result)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Executing maintenance task"), "expected log 'Executing maintenance task'")
}

func TestScanVulnerabilitiesActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(ScanVulnerabilitiesActivity)

	// Вход
	config := map[string]interface{}{
		"scan_type": "full",
		"timeout":   300,
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(ScanVulnerabilitiesActivity, config)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result VulnerabilityScanResult
	require.NoError(t, fut.Get(&result))
	assert.Empty(t, result.Vulnerabilities)
	assert.NotZero(t, result.ScanTime)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Scanning for vulnerabilities"), "expected log 'Scanning for vulnerabilities'")
}

func TestAnalyzeSecurityConfigActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(AnalyzeSecurityConfigActivity)

	// Вход
	config := map[string]interface{}{
		"firewall": "enabled",
		"ssl":      "enabled",
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(AnalyzeSecurityConfigActivity, config)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result SecurityConfigAnalysisResult
	require.NoError(t, fut.Get(&result))
	assert.Empty(t, result.Issues)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Analyzing security configuration"), "expected log 'Analyzing security configuration'")
}

func TestCheckComplianceActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(CheckComplianceActivity)

	// Вход
	config := map[string]interface{}{
		"standards": []string{"ISO27001", "SOC2"},
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(CheckComplianceActivity, config)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result ComplianceCheckResult
	require.NoError(t, fut.Get(&result))
	assert.Empty(t, result.Issues)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Checking compliance"), "expected log 'Checking compliance'")
}

func TestGenerateSecurityReportActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(GenerateSecurityReportActivity)

	// Вход
	input := SecurityReportInput{
		Vulnerabilities:  []Vulnerability{},
		ConfigIssues:     []SecurityIssue{},
		ComplianceIssues: []ComplianceIssue{},
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(GenerateSecurityReportActivity, input)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result SecurityReportResult
	require.NoError(t, fut.Get(&result))
	assert.Equal(t, "/var/log/security-report.json", result.ReportPath)
	assert.NotZero(t, result.GeneratedAt)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Generating security report"), "expected log 'Generating security report'")
}

func TestPrepareBackupActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(PrepareBackupActivity)

	// Вход
	config := map[string]interface{}{
		"backup_type": "full",
		"compression": "gzip",
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(PrepareBackupActivity, config)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result BackupPreparationResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Preparing backup"), "expected log 'Preparing backup'")
}

func TestBackupConfigActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(BackupConfigActivity)

	// Вход
	paths := []string{"/etc/relay", "/etc/wireguard"}

	// Запуск по функции
	fut, err := env.ExecuteActivity(BackupConfigActivity, paths)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result ConfigBackupResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Equal(t, "/backup/config-backup.tar.gz", result.BackupPath)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Backing up configuration"), "expected log 'Backing up configuration'")
}

func TestBackupDataActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(BackupDataActivity)

	// Вход
	paths := []string{"/var/relay", "/var/log"}

	// Запуск по функции
	fut, err := env.ExecuteActivity(BackupDataActivity, paths)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result DataBackupResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Equal(t, "/backup/data-backup.tar.gz", result.BackupPath)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Backing up data"), "expected log 'Backing up data'")
}

func TestValidateBackupActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(ValidateBackupActivity)

	// Вход
	input := BackupValidationInput{
		ConfigBackup: "/backup/config-backup.tar.gz",
		DataBackup:   "/backup/data-backup.tar.gz",
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(ValidateBackupActivity, input)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result BackupValidationResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Valid)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Validating backups"), "expected log 'Validating backups'")
}

func TestCleanupOldBackupsActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(CleanupOldBackupsActivity)

	// Вход
	config := map[string]interface{}{
		"retention_days": 30,
		"backup_path":    "/backup",
	}

	// Запуск по функции
	_, err := env.ExecuteActivity(CleanupOldBackupsActivity, config)
	require.NoError(t, err)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Cleaning up old backups"), "expected log 'Cleaning up old backups'")
}

func TestAssessDamageActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(AssessDamageActivity)

	// Вход
	config := map[string]interface{}{
		"assessment_type": "full",
		"components":      []string{"network", "storage"},
	}

	// Запуск по функции
	fut, err := env.ExecuteActivity(AssessDamageActivity, config)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result DamageAssessmentResult
	require.NoError(t, fut.Get(&result))
	assert.Equal(t, "low", result.Severity)
	assert.Empty(t, result.Affected)
	assert.Empty(t, result.Issues)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Assessing damage"), "expected log 'Assessing damage'")
}

func TestRecoverConfigActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(RecoverConfigActivity)

	// Вход
	backupPath := "/backup/config-backup.tar.gz"

	// Запуск по функции
	fut, err := env.ExecuteActivity(RecoverConfigActivity, backupPath)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result ConfigRecoveryResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Recovering configuration"), "expected log 'Recovering configuration'")
}

func TestRecoverDataActivity_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	// Регистрируем activity
	env.RegisterActivity(RecoverDataActivity)

	// Вход
	backupPath := "/backup/data-backup.tar.gz"

	// Запуск по функции
	fut, err := env.ExecuteActivity(RecoverDataActivity, backupPath)
	require.NoError(t, err)

	// Результат — Get ОДИН раз
	var result DataRecoveryResult
	require.NoError(t, fut.Get(&result))
	assert.True(t, result.Success)
	assert.Empty(t, result.Error)

	// Проверка логов — ищем сообщение в любом entry
	assert.True(t, logContains(logs, "Recovering data"), "expected log 'Recovering data'")
}

// Тесты для edge cases
func TestActivities_EdgeCases_WithTestSuite(t *testing.T) {
	t.Parallel()

	var suite testsuite.WorkflowTestSuite

	// наблюдаемый zap
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	suite.SetLogger(logger)

	env := suite.NewTestActivityEnvironment()

	t.Run("ValidateConfigActivity with nil config", func(t *testing.T) {
		env.RegisterActivity(ValidateConfigActivity)
		fut, err := env.ExecuteActivity(ValidateConfigActivity, nil)
		require.NoError(t, err)

		var result ConfigValidationResult
		require.NoError(t, fut.Get(&result))
		assert.True(t, result.Valid)
	})

	t.Run("StartServicesActivity with empty services", func(t *testing.T) {
		env.RegisterActivity(StartServicesActivity)
		fut, err := env.ExecuteActivity(StartServicesActivity, []string{})
		require.NoError(t, err)

		var result ServiceStartResult
		require.NoError(t, fut.Get(&result))
		assert.True(t, result.Success)
		assert.Empty(t, result.Services)
	})

	t.Run("HealthCheckActivity with empty services", func(t *testing.T) {
		env.RegisterActivity(HealthCheckActivity)
		input := HealthCheckInput{
			Services: []string{},
			Timeout:  10 * time.Second,
		}
		fut, err := env.ExecuteActivity(HealthCheckActivity, input)
		require.NoError(t, err)

		var result HealthCheckResult
		require.NoError(t, fut.Get(&result))
		assert.True(t, result.Healthy)
		assert.Empty(t, result.Services)
	})

	t.Run("ValidateCertificatesActivity with empty certificates", func(t *testing.T) {
		env.RegisterActivity(ValidateCertificatesActivity)
		fut, err := env.ExecuteActivity(ValidateCertificatesActivity, []string{})
		require.NoError(t, err)

		var result CertificateValidationResult
		require.NoError(t, fut.Get(&result))
		assert.True(t, result.Valid)
	})

	t.Run("BackupConfigActivity with empty paths", func(t *testing.T) {
		env.RegisterActivity(BackupConfigActivity)
		fut, err := env.ExecuteActivity(BackupConfigActivity, []string{})
		require.NoError(t, err)

		var result ConfigBackupResult
		require.NoError(t, fut.Get(&result))
		assert.True(t, result.Success)
	})

	t.Run("BackupDataActivity with empty paths", func(t *testing.T) {
		env.RegisterActivity(BackupDataActivity)
		fut, err := env.ExecuteActivity(BackupDataActivity, []string{})
		require.NoError(t, err)

		var result DataBackupResult
		require.NoError(t, fut.Get(&result))
		assert.True(t, result.Success)
	})
}
