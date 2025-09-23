package cadence

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/cadence/worker"
	"gopkg.in/yaml.v2"
)

// CadenceConfig представляет конфигурацию Cadence Workflow
type CadenceConfig struct {
	// Базовые настройки
	Domain      string `yaml:"domain"`
	ServiceName string `yaml:"serviceName"`

	// Настройки клиента
	Client ClientConfig `yaml:"client"`

	// Настройки worker
	Worker WorkerConfig `yaml:"worker"`

	// Настройки workflows
	Workflows WorkflowsConfig `yaml:"workflows"`

	// Настройки retry политик
	Retry RetryConfig `yaml:"retry"`

	// Настройки логирования
	Logging LoggingConfig `yaml:"logging"`

	// Настройки метрик
	Metrics MetricsConfig `yaml:"metrics"`

	// Настройки архивирования
	Archival ArchivalConfig `yaml:"archival"`

	// Настройки безопасности
	Security SecurityConfig `yaml:"security"`

	// Настройки уведомлений
	Notifications NotificationsConfig `yaml:"notifications"`

	// Настройки мониторинга
	Monitoring MonitoringConfig `yaml:"monitoring"`

	// Настройки профилирования
	Profiling ProfilingConfig `yaml:"profiling"`

	// Настройки health endpoints
	Health HealthConfig `yaml:"health"`

	// Настройки debug endpoints
	Debug DebugConfig `yaml:"debug"`
}

// ClientConfig настройки Cadence клиента
type ClientConfig struct {
	HostPort      string `yaml:"hostPort"`
	Domain        string `yaml:"domain"`
	DataConverter string `yaml:"dataConverter"`
	EnableTLS     bool   `yaml:"enableTLS"`
	TLSCertPath   string `yaml:"tlsCertPath"`
	TLSKeyPath    string `yaml:"tlsKeyPath"`
	TLSCAPath     string `yaml:"tlsCAPath"`

	// Настройки service client
	ServiceClient ServiceClientConfig `yaml:"serviceClient"`
}

// ServiceClientConfig настройки для service client
type ServiceClientConfig struct {
	Timeout     time.Duration `yaml:"timeout"`
	ServiceName string        `yaml:"serviceName"`
	EnableTLS   bool          `yaml:"enableTLS"`
	TLSCertPath string        `yaml:"tlsCertPath"`
	TLSKeyPath  string        `yaml:"tlsKeyPath"`
	TLSCAPath   string        `yaml:"tlsCAPath"`
}

// WorkerConfig настройки Cadence worker
type WorkerConfig struct {
	MaxConcurrentActivityExecutionSize      int    `yaml:"maxConcurrentActivityExecutionSize"`
	MaxConcurrentWorkflowTaskExecutionSize  int    `yaml:"maxConcurrentWorkflowTaskExecutionSize"`
	MaxConcurrentLocalActivityExecutionSize int    `yaml:"maxConcurrentLocalActivityExecutionSize"`
	MaxTaskListActivitiesPerSecond          int    `yaml:"maxTaskListActivitiesPerSecond"`
	WorkerActivitiesPerSecond               int    `yaml:"workerActivitiesPerSecond"`
	TaskListName                            string `yaml:"taskListName"`
	Identity                                string `yaml:"identity"`
}

// WorkflowsConfig настройки workflows
type WorkflowsConfig struct {
	// Общие timeout настройки
	ExecutionTimeout            string `yaml:"executionTimeout"`
	DecisionTimeout             string `yaml:"decisionTimeout"`
	ActivityStartToCloseTimeout string `yaml:"activityStartToCloseTimeout"`
	WorkflowStartToCloseTimeout string `yaml:"workflowStartToCloseTimeout"`
	WorkflowTaskTimeout         string `yaml:"workflowTaskTimeout"`

	// Специфичные настройки для разных типов workflows
	Main      WorkflowTypeConfig `yaml:"main"`
	Health    WorkflowTypeConfig `yaml:"health"`
	WireGuard WorkflowTypeConfig `yaml:"wireguard"`
	Quantum   WorkflowTypeConfig `yaml:"quantum"`
	AI        WorkflowTypeConfig `yaml:"ai"`
	Scaling   WorkflowTypeConfig `yaml:"scaling"`
	Security  WorkflowTypeConfig `yaml:"security"`
}

// WorkflowTypeConfig настройки для конкретного типа workflow
type WorkflowTypeConfig struct {
	ExecutionTimeout            string `yaml:"executionTimeout"`
	WorkflowStartToCloseTimeout string `yaml:"workflowStartToCloseTimeout"`
	ActivityStartToCloseTimeout string `yaml:"activityStartToCloseTimeout"`
}

// RetryConfig настройки retry политик
type RetryConfig struct {
	InitialInterval    string          `yaml:"initialInterval"`
	MaximumInterval    string          `yaml:"maximumInterval"`
	BackoffCoefficient float64         `yaml:"backoffCoefficient"`
	MaximumAttempts    int             `yaml:"maximumAttempts"`
	Critical           RetryTypeConfig `yaml:"critical"`
	External           RetryTypeConfig `yaml:"external"`
}

// RetryTypeConfig настройки retry для конкретного типа операций
type RetryTypeConfig struct {
	InitialInterval    string  `yaml:"initialInterval"`
	MaximumInterval    string  `yaml:"maximumInterval"`
	BackoffCoefficient float64 `yaml:"backoffCoefficient"`
	MaximumAttempts    int     `yaml:"maximumAttempts"`
}

// LoggingConfig настройки логирования
type LoggingConfig struct {
	Level                   string `yaml:"level"`
	EnableStructuredLogging bool   `yaml:"enableStructuredLogging"`
	LogWorkflowHistory      bool   `yaml:"logWorkflowHistory"`
	LogActivityDetails      bool   `yaml:"logActivityDetails"`
}

// MetricsConfig настройки метрик
type MetricsConfig struct {
	Enabled           bool   `yaml:"enabled"`
	Scope             string `yaml:"scope"`
	ReportingInterval string `yaml:"reportingInterval"`
	EnablePrometheus  bool   `yaml:"enablePrometheus"`
	PrometheusPort    int    `yaml:"prometheusPort"`
}

// ArchivalConfig настройки архивирования
type ArchivalConfig struct {
	Enabled          bool `yaml:"enabled"`
	RetentionDays    int  `yaml:"retentionDays"`
	ArchiveAfterDays int  `yaml:"archiveAfterDays"`
}

// SecurityConfig настройки безопасности
type SecurityConfig struct {
	EnableAuthorization   bool   `yaml:"enableAuthorization"`
	AuthorizationProvider string `yaml:"authorizationProvider"`
	EnableAuditLogging    bool   `yaml:"enableAuditLogging"`
	EnableEncryption      bool   `yaml:"enableEncryption"`
}

// NotificationsConfig настройки уведомлений
type NotificationsConfig struct {
	Enabled  bool                        `yaml:"enabled"`
	Channels []NotificationChannelConfig `yaml:"channels"`
}

// NotificationChannelConfig настройки канала уведомлений
type NotificationChannelConfig struct {
	Type    string      `yaml:"type"`
	Webhook string      `yaml:"webhook,omitempty"`
	SMTP    *SMTPConfig `yaml:"smtp,omitempty"`
	Enabled bool        `yaml:"enabled"`
}

// SMTPConfig настройки SMTP для email уведомлений
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// MonitoringConfig настройки мониторинга
type MonitoringConfig struct {
	HealthCheckInterval string                `yaml:"healthCheckInterval"`
	AlertThresholds     AlertThresholdsConfig `yaml:"alertThresholds"`
}

// AlertThresholdsConfig пороги для алертов
type AlertThresholdsConfig struct {
	WorkflowFailureRate float64 `yaml:"workflowFailureRate"`
	ActivityLatency     string  `yaml:"activityLatency"`
	QueueDepth          int     `yaml:"queueDepth"`
}

// ProfilingConfig настройки профилирования
type ProfilingConfig struct {
	Enabled                  bool   `yaml:"enabled"`
	Port                     string `yaml:"port"`
	EnableMemoryProfiling    bool   `yaml:"enableMemoryProfiling"`
	EnableCPUProfiling       bool   `yaml:"enableCPUProfiling"`
	EnableGoroutineProfiling bool   `yaml:"enableGoroutineProfiling"`
}

// HealthConfig настройки health endpoints
type HealthConfig struct {
	Enabled              bool   `yaml:"enabled"`
	Port                 string `yaml:"port"`
	ReadinessTimeout     string `yaml:"readinessTimeout"`
	LivenessTimeout      string `yaml:"livenessTimeout"`
	EnableDetailedHealth bool   `yaml:"enableDetailedHealth"`
}

// DebugConfig настройки debug endpoints
type DebugConfig struct {
	Enabled             bool `yaml:"enabled"`
	EnableWorkflowDebug bool `yaml:"enableWorkflowDebug"`
	EnableActivityDebug bool `yaml:"enableActivityDebug"`
	EnableMetricsDebug  bool `yaml:"enableMetricsDebug"`
	EnableConfigDebug   bool `yaml:"enableConfigDebug"`
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(configPath string) (*CadenceConfig, error) {
	// Читаем файл конфигурации
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл конфигурации %s: %w", configPath, err)
	}

	// Заменяем переменные окружения
	expandedData := os.ExpandEnv(string(data))

	var fullConfig struct {
		Cadence CadenceConfig `yaml:"cadence"`
	}

	// Парсим YAML
	if err := yaml.Unmarshal([]byte(expandedData), &fullConfig); err != nil {
		return nil, fmt.Errorf("не удалось распарсить конфигурацию: %w", err)
	}

	config := &fullConfig.Cadence

	// Валидируем конфигурацию
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("невалидная конфигурация: %w", err)
	}

	// Устанавливаем значения по умолчанию
	config.SetDefaults()

	return config, nil
}

// Validate проверяет корректность конфигурации
func (c *CadenceConfig) Validate() error {
	if c.Domain == "" {
		return fmt.Errorf("domain не может быть пустым")
	}

	if c.ServiceName == "" {
		return fmt.Errorf("serviceName не может быть пустым")
	}

	if c.Client.HostPort == "" {
		return fmt.Errorf("client.hostPort не может быть пустым")
	}

	if c.Client.Domain == "" {
		return fmt.Errorf("client.domain не может быть пустым")
	}

	if c.Worker.TaskListName == "" {
		return fmt.Errorf("worker.taskListName не может быть пустым")
	}

	// Проверяем timeout значения
	timeouts := map[string]string{
		"workflows.executionTimeout":            c.Workflows.ExecutionTimeout,
		"workflows.decisionTimeout":             c.Workflows.DecisionTimeout,
		"workflows.activityStartToCloseTimeout": c.Workflows.ActivityStartToCloseTimeout,
		"workflows.workflowStartToCloseTimeout": c.Workflows.WorkflowStartToCloseTimeout,
	}

	for field, timeout := range timeouts {
		if timeout != "" {
			if _, err := time.ParseDuration(timeout); err != nil {
				return fmt.Errorf("невалидный timeout в %s: %w", field, err)
			}
		}
	}

	return nil
}

// SetDefaults устанавливает значения по умолчанию
func (c *CadenceConfig) SetDefaults() {
	// Базовые значения по умолчанию
	if c.Client.DataConverter == "" {
		c.Client.DataConverter = "json"
	}

	if c.Worker.TaskListName == "" {
		c.Worker.TaskListName = "cloudbridge-task-list"
	}

	if c.Worker.Identity == "" {
		c.Worker.Identity = "cloudbridge-worker"
	}

	if c.Worker.MaxConcurrentActivityExecutionSize == 0 {
		c.Worker.MaxConcurrentActivityExecutionSize = 1000
	}

	if c.Worker.MaxConcurrentWorkflowTaskExecutionSize == 0 {
		c.Worker.MaxConcurrentWorkflowTaskExecutionSize = 1000
	}

	// Timeout значения по умолчанию
	if c.Workflows.ExecutionTimeout == "" {
		c.Workflows.ExecutionTimeout = "1h"
	}

	if c.Workflows.DecisionTimeout == "" {
		c.Workflows.DecisionTimeout = "10s"
	}

	if c.Workflows.ActivityStartToCloseTimeout == "" {
		c.Workflows.ActivityStartToCloseTimeout = "5m"
	}

	// Retry значения по умолчанию
	if c.Retry.InitialInterval == "" {
		c.Retry.InitialInterval = "1s"
	}

	if c.Retry.MaximumInterval == "" {
		c.Retry.MaximumInterval = "10s"
	}

	if c.Retry.BackoffCoefficient == 0 {
		c.Retry.BackoffCoefficient = 2.0
	}

	if c.Retry.MaximumAttempts == 0 {
		c.Retry.MaximumAttempts = 3
	}

	// Logging значения по умолчанию
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}

	// Metrics значения по умолчанию
	if c.Metrics.Scope == "" {
		c.Metrics.Scope = "cloudbridge"
	}

	if c.Metrics.ReportingInterval == "" {
		c.Metrics.ReportingInterval = "30s"
	}

	if c.Metrics.PrometheusPort == 0 {
		c.Metrics.PrometheusPort = 9091
	}

	// Archival значения по умолчанию
	if c.Archival.RetentionDays == 0 {
		c.Archival.RetentionDays = 30
	}

	if c.Archival.ArchiveAfterDays == 0 {
		c.Archival.ArchiveAfterDays = 7
	}

	// Monitoring значения по умолчанию
	if c.Monitoring.HealthCheckInterval == "" {
		c.Monitoring.HealthCheckInterval = "30s"
	}

	if c.Monitoring.AlertThresholds.WorkflowFailureRate == 0 {
		c.Monitoring.AlertThresholds.WorkflowFailureRate = 0.05
	}

	if c.Monitoring.AlertThresholds.ActivityLatency == "" {
		c.Monitoring.AlertThresholds.ActivityLatency = "30s"
	}

	if c.Monitoring.AlertThresholds.QueueDepth == 0 {
		c.Monitoring.AlertThresholds.QueueDepth = 1000
	}

	// Profiling значения по умолчанию
	if c.Profiling.Port == "" {
		c.Profiling.Port = ":6060"
	}

	// Health значения по умолчанию
	if c.Health.Port == "" {
		c.Health.Port = ":8080"
	}

	if c.Health.ReadinessTimeout == "" {
		c.Health.ReadinessTimeout = "10s"
	}

	if c.Health.LivenessTimeout == "" {
		c.Health.LivenessTimeout = "5s"
	}
}

// GetExecutionTimeout возвращает timeout для конкретного типа workflow
func (c *CadenceConfig) GetExecutionTimeout(workflowType string) time.Duration {
	var timeout string

	switch workflowType {
	case "main":
		timeout = c.Workflows.Main.ExecutionTimeout
	case "health":
		timeout = c.Workflows.Health.ExecutionTimeout
	case "wireguard":
		timeout = c.Workflows.WireGuard.ExecutionTimeout
	case "quantum":
		timeout = c.Workflows.Quantum.ExecutionTimeout
	case "ai":
		timeout = c.Workflows.AI.ExecutionTimeout
	case "scaling":
		timeout = c.Workflows.Scaling.ExecutionTimeout
	case "security":
		timeout = c.Workflows.Security.ExecutionTimeout
	default:
		timeout = c.Workflows.ExecutionTimeout
	}

	if timeout == "" {
		timeout = c.Workflows.ExecutionTimeout
	}

	duration, err := time.ParseDuration(timeout)
	if err != nil {
		// Fallback на значение по умолчанию
		duration, _ = time.ParseDuration("1h")
	}

	return duration
}

// GetActivityTimeout возвращает timeout для activities конкретного типа workflow
func (c *CadenceConfig) GetActivityTimeout(workflowType string) time.Duration {
	var timeout string

	switch workflowType {
	case "main":
		timeout = c.Workflows.Main.ActivityStartToCloseTimeout
	case "health":
		timeout = c.Workflows.Health.ActivityStartToCloseTimeout
	case "wireguard":
		timeout = c.Workflows.WireGuard.ActivityStartToCloseTimeout
	case "quantum":
		timeout = c.Workflows.Quantum.ActivityStartToCloseTimeout
	case "ai":
		timeout = c.Workflows.AI.ActivityStartToCloseTimeout
	case "scaling":
		timeout = c.Workflows.Scaling.ActivityStartToCloseTimeout
	case "security":
		timeout = c.Workflows.Security.ActivityStartToCloseTimeout
	default:
		timeout = c.Workflows.ActivityStartToCloseTimeout
	}

	if timeout == "" {
		timeout = c.Workflows.ActivityStartToCloseTimeout
	}

	duration, err := time.ParseDuration(timeout)
	if err != nil {
		// Fallback на значение по умолчанию
		duration, _ = time.ParseDuration("5m")
	}

	return duration
}

// ToServiceClientOptions преобразует ClientConfig в ServiceClientOptions
func (c *ClientConfig) ToServiceClientOptions() ServiceClientOptions {
	return ServiceClientOptions{
		Timeout:     c.ServiceClient.Timeout,
		ServiceName: c.ServiceClient.ServiceName,
		EnableTLS:   c.ServiceClient.EnableTLS,
		TLSCertPath: c.ServiceClient.TLSCertPath,
		TLSKeyPath:  c.ServiceClient.TLSKeyPath,
		TLSCAPath:   c.ServiceClient.TLSCAPath,
	}
}

// ToClientConfig преобразует CadenceConfig в Config для client
func (c *CadenceConfig) ToClientConfig() Config {
	// Создаем worker.Options из WorkerConfig
	workerOptions := worker.Options{
		MaxConcurrentActivityExecutionSize:      c.Worker.MaxConcurrentActivityExecutionSize,
		MaxConcurrentLocalActivityExecutionSize: c.Worker.MaxConcurrentLocalActivityExecutionSize,
		WorkerActivitiesPerSecond:               float64(c.Worker.WorkerActivitiesPerSecond),
		Identity:                                c.Worker.Identity,
	}

	return Config{
		HostPort:       c.Client.HostPort,
		Domain:         c.Client.Domain,
		TaskList:       c.Worker.TaskListName,
		WorkerOptions:  workerOptions,
		ServiceOptions: c.Client.ToServiceClientOptions(),
	}
}
