package interfaces

import (
	"time"
)

// TunnelConfig представляет конфигурацию туннеля
type TunnelConfig struct {
	Protocol       Protocol           `json:"protocol"`
	MTU            MTU                `json:"mtu,omitempty"`
	BandwidthLimit BandwidthLimitBps  `json:"bandwidth_limit_bps,omitempty"`
	KeepAliveMs    KeepAliveMs        `json:"keep_alive_ms,omitempty"`
	TimeoutMs      DurationMillis     `json:"timeout_ms,omitempty"`
	RetryCount     int                `json:"retry_count,omitempty"`
	RetryDelayMs   DurationMillis     `json:"retry_delay_ms,omitempty"`
	Encryption     EncryptionConfig   `json:"encryption,omitempty"`
	Compression    CompressionConfig  `json:"compression,omitempty"`
	Quality        QualityConfig      `json:"quality,omitempty"`
	Metadata       Metadata           `json:"metadata,omitempty"`
}

// EncryptionConfig представляет конфигурацию шифрования
type EncryptionConfig struct {
	Algorithm    Algorithm `json:"algorithm"`
	KeyLength    int       `json:"key_length,omitempty"`
	KeyRotation  bool      `json:"key_rotation,omitempty"`
	KeyRotationMs DurationMillis `json:"key_rotation_ms,omitempty"`
}

// CompressionConfig представляет конфигурацию сжатия
type CompressionConfig struct {
	Enabled   bool    `json:"enabled"`
	Algorithm string  `json:"algorithm,omitempty"`
	Level     int     `json:"level,omitempty"`
}

// QualityConfig представляет конфигурацию качества
type QualityConfig struct {
	MinHealthScore HealthScore `json:"min_health_score,omitempty"`
	MaxLatencyMs   LatencyMs   `json:"max_latency_ms,omitempty"`
	MinThroughput  ThroughputBps `json:"min_throughput_bps,omitempty"`
}

// ValidationConfig представляет конфигурацию валидации
type ValidationConfig struct {
	Enabled       bool            `json:"enabled"`
	IntervalMs    DurationMillis  `json:"interval_ms,omitempty"`
	TimeoutMs     DurationMillis  `json:"timeout_ms,omitempty"`
	RetryCount    int             `json:"retry_count,omitempty"`
	StrictMode    bool            `json:"strict_mode,omitempty"`
}

// ThreatAnalysisConfig представляет конфигурацию анализа угроз
type ThreatAnalysisConfig struct {
	TimeRange     TimeRange       `json:"time_range"`
	DataSources   []string        `json:"data_sources"`
	AnalysisType  string          `json:"analysis_type"`
	Sensitivity   SensitivityLevel `json:"sensitivity"`
	Filters       map[string]string `json:"filters,omitempty"`
	Metadata      Metadata        `json:"metadata,omitempty"`
}

// SensitivityLevel представляет уровень чувствительности
type SensitivityLevel string

const (
	SensitivityLevelLow    SensitivityLevel = "low"
	SensitivityLevelMedium SensitivityLevel = "medium"
	SensitivityLevelHigh   SensitivityLevel = "high"
	SensitivityLevelCritical SensitivityLevel = "critical"
)

// BehaviorAnalysisConfig представляет конфигурацию анализа поведения
type BehaviorAnalysisConfig struct {
	UserID        string          `json:"user_id"`
	TimeRange     TimeRange       `json:"time_range"`
	AnalysisType  string          `json:"analysis_type"`
	BaselineDays  int             `json:"baseline_days,omitempty"`
	Thresholds    ThresholdConfig `json:"thresholds,omitempty"`
	Metadata      Metadata        `json:"metadata,omitempty"`
}

// ThresholdConfig представляет конфигурацию порогов
type ThresholdConfig struct {
	RiskScore     RiskScore `json:"risk_score,omitempty"`
	AnomalyScore  float64   `json:"anomaly_score,omitempty"`
	Confidence    Confidence `json:"confidence,omitempty"`
}

// ScalingConfig представляет конфигурацию масштабирования
type ScalingConfig struct {
	TargetService    string              `json:"target_service"`
	MinInstances     int                 `json:"min_instances"`
	MaxInstances     int                 `json:"max_instances"`
	ScalingPolicy    ScalingPolicy       `json:"scaling_policy"`
	Metrics          ScalingMetricsConfig `json:"metrics,omitempty"`
	CooldownMs       DurationMillis      `json:"cooldown_ms,omitempty"`
	Metadata         Metadata            `json:"metadata,omitempty"`
}

// ScalingPolicy представляет политику масштабирования
type ScalingPolicy struct {
	Type        string            `json:"type"`
	Thresholds  map[string]float64 `json:"thresholds,omitempty"`
	Actions     []ScalingAction   `json:"actions,omitempty"`
	Schedule    *ScheduleConfig   `json:"schedule,omitempty"`
}

// ScalingMetricsConfig представляет конфигурацию метрик масштабирования
type ScalingMetricsConfig struct {
	CPUThreshold    float64 `json:"cpu_threshold,omitempty"`
	MemoryThreshold float64 `json:"memory_threshold,omitempty"`
	NetworkThreshold float64 `json:"network_threshold,omitempty"`
	ResponseTimeThreshold ResponseTimeMs `json:"response_time_threshold,omitempty"`
}

// ScheduleConfig представляет конфигурацию расписания
type ScheduleConfig struct {
	Timezone   string    `json:"timezone"`
	StartTime  time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	DaysOfWeek []int     `json:"days_of_week,omitempty"`
	DaysOfMonth []int    `json:"days_of_month,omitempty"`
}

// SecurityAuditConfig представляет конфигурацию аудита безопасности
type SecurityAuditConfig struct {
	Area           string                 `json:"area"`
	Standard       ComplianceStandard    `json:"standard"`
	Scope          []string              `json:"scope,omitempty"`
	Depth          AuditDepth            `json:"depth,omitempty"`
	IncludeDetails bool                  `json:"include_details,omitempty"`
	Metadata       Metadata              `json:"metadata,omitempty"`
}

// AuditDepth представляет глубину аудита
type AuditDepth string

const (
	AuditDepthBasic    AuditDepth = "basic"
	AuditDepthStandard AuditDepth = "standard"
	AuditDepthDeep     AuditDepth = "deep"
	AuditDepthComprehensive AuditDepth = "comprehensive"
)

// ComplianceConfig представляет конфигурацию соответствия
type ComplianceConfig struct {
	Standard       ComplianceStandard `json:"standard"`
	Version        string              `json:"version,omitempty"`
	Requirements   []string            `json:"requirements,omitempty"`
	Deadline       time.Time           `json:"deadline,omitempty"`
	AutoRemediation bool               `json:"auto_remediation,omitempty"`
	Metadata       Metadata            `json:"metadata,omitempty"`
}

// ModelTrainingConfig представляет конфигурацию обучения модели
type ModelTrainingConfig struct {
	ModelType      ModelType           `json:"model_type"`
	TrainingData   TrainingDataConfig  `json:"training_data"`
	Hyperparameters map[string]interface{} `json:"hyperparameters,omitempty"`
	Validation     ValidationConfig    `json:"validation,omitempty"`
	Metadata       Metadata            `json:"metadata,omitempty"`
}

// TrainingDataConfig представляет конфигурацию данных обучения
type TrainingDataConfig struct {
	Source        string          `json:"source"`
	Size          int64           `json:"size,omitempty"`
	Format        string          `json:"format,omitempty"`
	Preprocessing []string        `json:"preprocessing,omitempty"`
	Validation    ValidationConfig `json:"validation,omitempty"`
}

// NetworkAnalysisConfig представляет конфигурацию анализа сети
type NetworkAnalysisConfig struct {
	TimeRange     TimeRange       `json:"time_range"`
	IncludeTopology bool          `json:"include_topology,omitempty"`
	IncludeMetrics bool           `json:"include_metrics,omitempty"`
	IncludeBottlenecks bool       `json:"include_bottlenecks,omitempty"`
	Filters       map[string]string `json:"filters,omitempty"`
	Metadata      Metadata        `json:"metadata,omitempty"`
}

// TopologyPlanConfig представляет конфигурацию планирования топологии
type TopologyPlanConfig struct {
	Requirements  map[string]interface{} `json:"requirements"`
	Constraints   []string                `json:"constraints,omitempty"`
	Optimization  OptimizationConfig     `json:"optimization,omitempty"`
	Metadata      Metadata                `json:"metadata,omitempty"`
}

// OptimizationConfig представляет конфигурацию оптимизации
type OptimizationConfig struct {
	Goals         []string        `json:"goals,omitempty"`
	Weights       map[string]float64 `json:"weights,omitempty"`
	MaxIterations int             `json:"max_iterations,omitempty"`
	TimeoutMs     DurationMillis  `json:"timeout_ms,omitempty"`
}

// ConnectionValidationConfig представляет конфигурацию валидации соединения
type ConnectionValidationConfig struct {
	PeerID1       string          `json:"peer_id1"`
	PeerID2       string          `json:"peer_id2"`
	TestType      string          `json:"test_type,omitempty"`
	TimeoutMs     DurationMillis  `json:"timeout_ms,omitempty"`
	RetryCount    int             `json:"retry_count,omitempty"`
	Metadata      Metadata        `json:"metadata,omitempty"`
}

// AuditQuery представляет запрос аудита
type AuditQuery struct {
	TimeRange   TimeRange       `json:"time_range"`
	Pagination  PaginationQuery `json:"pagination,omitempty"`
	Filters     FilterQuery     `json:"filters,omitempty"`
	SortBy      string          `json:"sort_by,omitempty"`
	SortOrder   string          `json:"sort_order,omitempty"`
}

// SecurityEventConfig представляет конфигурацию события безопасности
type SecurityEventConfig struct {
	EventType   EventType       `json:"event_type"`
	Severity    Severity        `json:"severity"`
	Source      string          `json:"source"`
	Description string          `json:"description"`
	Tags        []string        `json:"tags,omitempty"`
	Metadata    Metadata        `json:"metadata,omitempty"`
}

// PerformanceMonitoringConfig представляет конфигурацию мониторинга производительности
type PerformanceMonitoringConfig struct {
	Duration        DurationMillis `json:"duration"`
	IntervalMs      DurationMillis `json:"interval_ms,omitempty"`
	Metrics         []string       `json:"metrics,omitempty"`
	Thresholds      map[string]float64 `json:"thresholds,omitempty"`
	Alerting        AlertingConfig `json:"alerting,omitempty"`
	Metadata        Metadata       `json:"metadata,omitempty"`
}

// AlertingConfig представляет конфигурацию оповещений
type AlertingConfig struct {
	Enabled     bool     `json:"enabled"`
	Channels    []string `json:"channels,omitempty"`
	Recipients  []string `json:"recipients,omitempty"`
	Thresholds  map[string]float64 `json:"thresholds,omitempty"`
}

// CapacityPlanningConfig представляет конфигурацию планирования мощности
type CapacityPlanningConfig struct {
	PlanningHorizon DurationMillis `json:"planning_horizon"`
	GrowthRate      float64        `json:"growth_rate,omitempty"`
	Constraints     []string        `json:"constraints,omitempty"`
	Optimization    OptimizationConfig `json:"optimization,omitempty"`
	Metadata        Metadata        `json:"metadata,omitempty"`
}
