package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// CadenceMetrics содержит все метрики для Cadence Workflow
type CadenceMetrics struct {
	// Workflow метрики
	WorkflowExecutions *prometheus.CounterVec
	WorkflowDuration   *prometheus.HistogramVec
	WorkflowErrors     *prometheus.CounterVec
	WorkflowStatus     *prometheus.GaugeVec
	ActiveWorkflows    *prometheus.GaugeVec

	// Activity метрики
	ActivityExecutions *prometheus.CounterVec
	ActivityDuration   *prometheus.HistogramVec
	ActivityErrors     *prometheus.CounterVec
	ActivityRetries    *prometheus.CounterVec

	// Worker метрики
	WorkerConnections *prometheus.GaugeVec
	WorkerPollers     *prometheus.GaugeVec
	WorkerLatency     *prometheus.HistogramVec
	WorkerQueueDepth  *prometheus.GaugeVec

	// Клиент метрики
	ClientRequests *prometheus.CounterVec
	ClientLatency  *prometheus.HistogramVec
	ClientErrors   *prometheus.CounterVec

	// Специфичные метрики для CloudBridge
	QUICWorkflows     *prometheus.CounterVec
	AIWorkflows       *prometheus.CounterVec
	ScalingWorkflows  *prometheus.CounterVec
	SecurityWorkflows *prometheus.CounterVec

	// Здоровье системы
	HealthChecks        *prometheus.CounterVec
	SystemUptime        *prometheus.GaugeVec
	ResourceUtilization *prometheus.GaugeVec

	logger *zap.Logger
}

// NewCadenceMetrics создает новые Cadence метрики
func NewCadenceMetrics(logger *zap.Logger) *CadenceMetrics {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &CadenceMetrics{
		// Workflow метрики
		WorkflowExecutions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "workflow_executions_total",
				Help:      "Общее количество выполнений workflow",
			},
			[]string{"workflow_type", "status", "domain"},
		),

		WorkflowDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "workflow_duration_seconds",
				Help:      "Длительность выполнения workflow в секундах",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 20), // От 1ms до ~17 минут
			},
			[]string{"workflow_type", "domain"},
		),

		WorkflowErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "workflow_errors_total",
				Help:      "Общее количество ошибок workflow",
			},
			[]string{"workflow_type", "error_type", "domain"},
		),

		WorkflowStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "workflow_status",
				Help:      "Текущий статус workflow (1 = активен, 0 = неактивен)",
			},
			[]string{"workflow_type", "workflow_id", "domain"},
		),

		ActiveWorkflows: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "active_workflows",
				Help:      "Количество активных workflow",
			},
			[]string{"workflow_type", "domain"},
		),

		// Activity метрики
		ActivityExecutions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "activity_executions_total",
				Help:      "Общее количество выполнений activity",
			},
			[]string{"activity_type", "status", "domain"},
		),

		ActivityDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "activity_duration_seconds",
				Help:      "Длительность выполнения activity в секундах",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15), // От 1ms до ~32 секунд
			},
			[]string{"activity_type", "domain"},
		),

		ActivityErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "activity_errors_total",
				Help:      "Общее количество ошибок activity",
			},
			[]string{"activity_type", "error_type", "domain"},
		),

		ActivityRetries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "activity_retries_total",
				Help:      "Общее количество повторных попыток activity",
			},
			[]string{"activity_type", "attempt", "domain"},
		),

		// Worker метрики
		WorkerConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "worker_connections",
				Help:      "Количество активных worker соединений",
			},
			[]string{"worker_identity", "domain"},
		),

		WorkerPollers: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "worker_pollers",
				Help:      "Количество активных pollers",
			},
			[]string{"poller_type", "worker_identity", "domain"},
		),

		WorkerLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "worker_latency_seconds",
				Help:      "Задержка worker в секундах",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 12), // От 1ms до ~4 секунд
			},
			[]string{"operation", "worker_identity", "domain"},
		),

		WorkerQueueDepth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "worker_queue_depth",
				Help:      "Глубина очереди worker",
			},
			[]string{"queue_type", "worker_identity", "domain"},
		),

		// Клиент метрики
		ClientRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "client_requests_total",
				Help:      "Общее количество запросов клиента",
			},
			[]string{"operation", "status", "domain"},
		),

		ClientLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "client_latency_seconds",
				Help:      "Задержка клиента в секундах",
				Buckets:   prometheus.ExponentialBuckets(0.001, 2, 12),
			},
			[]string{"operation", "domain"},
		),

		ClientErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "client_errors_total",
				Help:      "Общее количество ошибок клиента",
			},
			[]string{"operation", "error_type", "domain"},
		),

		// Специфичные метрики для CloudBridge
		QUICWorkflows: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "quic_workflows_total",
				Help:      "Общее количество QUIC workflow",
			},
			[]string{"workflow_subtype", "status", "peer_id"},
		),

		AIWorkflows: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "ai_workflows_total",
				Help:      "Общее количество AI workflow",
			},
			[]string{"workflow_subtype", "status", "model_type"},
		),

		ScalingWorkflows: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "scaling_workflows_total",
				Help:      "Общее количество Scaling workflow",
			},
			[]string{"workflow_subtype", "status", "scale_direction"},
		),

		SecurityWorkflows: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "security_workflows_total",
				Help:      "Общее количество Security workflow",
			},
			[]string{"workflow_subtype", "status", "threat_level"},
		),

		// Здоровье системы
		HealthChecks: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "health_checks_total",
				Help:      "Общее количество проверок здоровья",
			},
			[]string{"check_type", "status", "component"},
		),

		SystemUptime: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "system_uptime_seconds",
				Help:      "Время работы системы в секундах",
			},
			[]string{"component", "instance"},
		),

		ResourceUtilization: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "cloudbridge",
				Subsystem: "cadence",
				Name:      "resource_utilization_ratio",
				Help:      "Использование ресурсов (0-1)",
			},
			[]string{"resource_type", "component"},
		),

		logger: logger,
	}
}

// Workflow методы для записи метрик

// RecordWorkflowStarted записывает метрику запуска workflow
func (cm *CadenceMetrics) RecordWorkflowStarted(workflowType, domain, workflowID string) {
	cm.WorkflowExecutions.WithLabelValues(workflowType, "started", domain).Inc()
	cm.WorkflowStatus.WithLabelValues(workflowType, workflowID, domain).Set(1)
	cm.ActiveWorkflows.WithLabelValues(workflowType, domain).Inc()

	cm.logger.Debug("Зафиксирован запуск workflow",
		zap.String("workflow_type", workflowType),
		zap.String("domain", domain),
		zap.String("workflow_id", workflowID),
	)
}

// RecordWorkflowCompleted записывает метрику завершения workflow
func (cm *CadenceMetrics) RecordWorkflowCompleted(workflowType, domain, workflowID string, duration time.Duration) {
	cm.WorkflowExecutions.WithLabelValues(workflowType, "completed", domain).Inc()
	cm.WorkflowDuration.WithLabelValues(workflowType, domain).Observe(duration.Seconds())
	cm.WorkflowStatus.WithLabelValues(workflowType, workflowID, domain).Set(0)
	cm.ActiveWorkflows.WithLabelValues(workflowType, domain).Dec()

	cm.logger.Debug("Зафиксировано завершение workflow",
		zap.String("workflow_type", workflowType),
		zap.String("domain", domain),
		zap.String("workflow_id", workflowID),
		zap.Duration("duration", duration),
	)
}

// RecordWorkflowFailed записывает метрику ошибки workflow
func (cm *CadenceMetrics) RecordWorkflowFailed(workflowType, domain, workflowID, errorType string, duration time.Duration) {
	cm.WorkflowExecutions.WithLabelValues(workflowType, "failed", domain).Inc()
	cm.WorkflowErrors.WithLabelValues(workflowType, errorType, domain).Inc()
	cm.WorkflowDuration.WithLabelValues(workflowType, domain).Observe(duration.Seconds())
	cm.WorkflowStatus.WithLabelValues(workflowType, workflowID, domain).Set(0)
	cm.ActiveWorkflows.WithLabelValues(workflowType, domain).Dec()

	cm.logger.Warn("Зафиксирована ошибка workflow",
		zap.String("workflow_type", workflowType),
		zap.String("domain", domain),
		zap.String("workflow_id", workflowID),
		zap.String("error_type", errorType),
		zap.Duration("duration", duration),
	)
}

// Activity методы

// RecordActivityStarted записывает метрику запуска activity
func (cm *CadenceMetrics) RecordActivityStarted(activityType, domain string) {
	cm.ActivityExecutions.WithLabelValues(activityType, "started", domain).Inc()

	cm.logger.Debug("Зафиксирован запуск activity",
		zap.String("activity_type", activityType),
		zap.String("domain", domain),
	)
}

// RecordActivityCompleted записывает метрику завершения activity
func (cm *CadenceMetrics) RecordActivityCompleted(activityType, domain string, duration time.Duration) {
	cm.ActivityExecutions.WithLabelValues(activityType, "completed", domain).Inc()
	cm.ActivityDuration.WithLabelValues(activityType, domain).Observe(duration.Seconds())

	cm.logger.Debug("Зафиксировано завершение activity",
		zap.String("activity_type", activityType),
		zap.String("domain", domain),
		zap.Duration("duration", duration),
	)
}

// RecordActivityFailed записывает метрику ошибки activity
func (cm *CadenceMetrics) RecordActivityFailed(activityType, domain, errorType string, duration time.Duration) {
	cm.ActivityExecutions.WithLabelValues(activityType, "failed", domain).Inc()
	cm.ActivityErrors.WithLabelValues(activityType, errorType, domain).Inc()
	cm.ActivityDuration.WithLabelValues(activityType, domain).Observe(duration.Seconds())

	cm.logger.Warn("Зафиксирована ошибка activity",
		zap.String("activity_type", activityType),
		zap.String("domain", domain),
		zap.String("error_type", errorType),
		zap.Duration("duration", duration),
	)
}

// RecordActivityRetry записывает метрику повтора activity
func (cm *CadenceMetrics) RecordActivityRetry(activityType, domain string, attempt int) {
	cm.ActivityRetries.WithLabelValues(activityType, string(rune(attempt)), domain).Inc()

	cm.logger.Debug("Зафиксирован повтор activity",
		zap.String("activity_type", activityType),
		zap.String("domain", domain),
		zap.Int("attempt", attempt),
	)
}

// Worker методы

// UpdateWorkerConnections обновляет количество worker соединений
func (cm *CadenceMetrics) UpdateWorkerConnections(workerIdentity, domain string, count float64) {
	cm.WorkerConnections.WithLabelValues(workerIdentity, domain).Set(count)
}

// UpdateWorkerPollers обновляет количество pollers
func (cm *CadenceMetrics) UpdateWorkerPollers(pollerType, workerIdentity, domain string, count float64) {
	cm.WorkerPollers.WithLabelValues(pollerType, workerIdentity, domain).Set(count)
}

// RecordWorkerLatency записывает задержку worker
func (cm *CadenceMetrics) RecordWorkerLatency(operation, workerIdentity, domain string, duration time.Duration) {
	cm.WorkerLatency.WithLabelValues(operation, workerIdentity, domain).Observe(duration.Seconds())
}

// UpdateWorkerQueueDepth обновляет глубину очереди worker
func (cm *CadenceMetrics) UpdateWorkerQueueDepth(queueType, workerIdentity, domain string, depth float64) {
	cm.WorkerQueueDepth.WithLabelValues(queueType, workerIdentity, domain).Set(depth)
}

// Client методы

// RecordClientRequest записывает клиентский запрос
func (cm *CadenceMetrics) RecordClientRequest(operation, status, domain string, duration time.Duration) {
	cm.ClientRequests.WithLabelValues(operation, status, domain).Inc()
	cm.ClientLatency.WithLabelValues(operation, domain).Observe(duration.Seconds())
}

// RecordClientError записывает ошибку клиента
func (cm *CadenceMetrics) RecordClientError(operation, errorType, domain string) {
	cm.ClientErrors.WithLabelValues(operation, errorType, domain).Inc()
}

// CloudBridge специфичные методы

// RecordQUICWorkflow записывает QUIC workflow
func (cm *CadenceMetrics) RecordQUICWorkflow(subtype, status, peerID string) {
	cm.QUICWorkflows.WithLabelValues(subtype, status, peerID).Inc()
}

// RecordAIWorkflow записывает AI workflow
func (cm *CadenceMetrics) RecordAIWorkflow(subtype, status, modelType string) {
	cm.AIWorkflows.WithLabelValues(subtype, status, modelType).Inc()
}

// RecordScalingWorkflow записывает Scaling workflow
func (cm *CadenceMetrics) RecordScalingWorkflow(subtype, status, scaleDirection string) {
	cm.ScalingWorkflows.WithLabelValues(subtype, status, scaleDirection).Inc()
}

// RecordSecurityWorkflow записывает Security workflow
func (cm *CadenceMetrics) RecordSecurityWorkflow(subtype, status, threatLevel string) {
	cm.SecurityWorkflows.WithLabelValues(subtype, status, threatLevel).Inc()
}

// Система здоровья

// RecordHealthCheck записывает проверку здоровья
func (cm *CadenceMetrics) RecordHealthCheck(checkType, status, component string) {
	cm.HealthChecks.WithLabelValues(checkType, status, component).Inc()
}

// UpdateSystemUptime обновляет время работы системы
func (cm *CadenceMetrics) UpdateSystemUptime(component, instance string, uptime time.Duration) {
	cm.SystemUptime.WithLabelValues(component, instance).Set(uptime.Seconds())
}

// UpdateResourceUtilization обновляет использование ресурсов
func (cm *CadenceMetrics) UpdateResourceUtilization(resourceType, component string, utilization float64) {
	cm.ResourceUtilization.WithLabelValues(resourceType, component).Set(utilization)
}

// StartMetricsCollection запускает периодический сбор метрик
func (cm *CadenceMetrics) StartMetricsCollection(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	cm.logger.Info("Запуск сбора Cadence метрик", zap.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			cm.logger.Info("Остановка сбора Cadence метрик")
			return
		case <-ticker.C:
			cm.collectSystemMetrics()
		}
	}
}

// collectSystemMetrics собирает системные метрики
func (cm *CadenceMetrics) collectSystemMetrics() {
	// Здесь можно добавить сбор дополнительных системных метрик
	// Например, использование памяти, CPU, количество горутин и т.д.

	cm.logger.Debug("Сбор системных метрик")
}
