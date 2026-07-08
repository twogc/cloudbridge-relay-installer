package interfaces

import (
	"time"
)

// Status представляет статус операции
type Status string

const (
	StatusCreated   Status = "created"
	StatusActive    Status = "active"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusDeleted   Status = "deleted"
	StatusPending   Status = "pending"
	StatusCancelled Status = "cancelled"
)

// Severity представляет уровень серьезности
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Priority представляет приоритет операции
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// ActionType представляет тип действия
type ActionType string

const (
	ActionTypeCreate   ActionType = "create"
	ActionTypeUpdate   ActionType = "update"
	ActionTypeDelete   ActionType = "delete"
	ActionTypeScale    ActionType = "scale"
	ActionTypeRotate   ActionType = "rotate"
	ActionTypeValidate ActionType = "validate"
	ActionTypeMonitor  ActionType = "monitor"
	ActionTypeAnalyze  ActionType = "analyze"
	ActionTypePlan     ActionType = "plan"
	ActionTypeOptimize ActionType = "optimize"
)

// ConnectionType представляет тип соединения
type ConnectionType string

const (
	ConnectionTypeDirect    ConnectionType = "direct"
	ConnectionTypeRelay     ConnectionType = "relay"
	ConnectionTypeTunnel    ConnectionType = "tunnel"
	ConnectionTypeQUIC      ConnectionType = "quic"
	ConnectionTypeWebSocket ConnectionType = "websocket"
)

// ThreatLevel представляет уровень угрозы
type ThreatLevel string

const (
	ThreatLevelLow      ThreatLevel = "low"
	ThreatLevelMedium   ThreatLevel = "medium"
	ThreatLevelHigh     ThreatLevel = "high"
	ThreatLevelCritical ThreatLevel = "critical"
)

// ComplianceStandard представляет стандарт соответствия
type ComplianceStandard string

const (
	ComplianceStandardSOC2     ComplianceStandard = "SOC2"
	ComplianceStandardGDPR     ComplianceStandard = "GDPR"
	ComplianceStandardHIPAA    ComplianceStandard = "HIPAA"
	ComplianceStandardISO27001 ComplianceStandard = "ISO27001"
)

// DurationMillis представляет длительность в миллисекундах для сериализации
type DurationMillis int64

// ToDuration конвертирует DurationMillis в time.Duration
func (d DurationMillis) ToDuration() time.Duration {
	return time.Duration(d) * time.Millisecond
}

// FromDuration создает DurationMillis из time.Duration
func FromDuration(d time.Duration) DurationMillis {
	return DurationMillis(d.Milliseconds())
}

// TimeRange представляет временной диапазон
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// Duration возвращает длительность диапазона
func (tr TimeRange) Duration() time.Duration {
	return tr.To.Sub(tr.From)
}

// SchemaVersion представляет версию схемы для эволюции API
type SchemaVersion string

const (
	SchemaVersionV1 SchemaVersion = "v1"
	SchemaVersionV2 SchemaVersion = "v2"
)

// QualityScore представляет оценку качества (0.0 - 1.0)
type QualityScore float64

// HealthScore представляет оценку здоровья (0.0 - 1.0)
type HealthScore float64

// RiskScore представляет оценку риска (0.0 - 1.0)
type RiskScore float64

// Confidence представляет уровень уверенности (0.0 - 1.0)
type Confidence float64

// AccuracyScore представляет оценку точности (0.0 - 1.0)
type AccuracyScore float64

// ImprovementScore представляет оценку улучшения (0.0 - 1.0)
type ImprovementScore float64

// ComplianceScoreValue представляет оценку соответствия (0.0 - 1.0)
type ComplianceScoreValue float64

// StrengthScore представляет оценку силы (0.0 - 1.0)
type StrengthScore float64

// LatencyMs представляет задержку в миллисекундах
type LatencyMs int64

// NetworkConditions представляет сетевые условия
type NetworkConditions struct {
	Latency    int64                  `json:"latency"`
	Throughput int64                  `json:"throughput"`
	Quality    float64                `json:"quality"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RoutePrediction представляет предсказание маршрута
type RoutePrediction struct {
	Route      string                 `json:"route"`
	Latency    int64                  `json:"latency"`
	Throughput int64                  `json:"throughput"`
	Quality    float64                `json:"quality"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// BackupResult представляет результат создания резервной копии
type BackupResult struct {
	BackupID   string    `json:"backup_id"`
	BackupType string    `json:"backup_type"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// BackupValidation представляет результат валидации резервной копии
type BackupValidation struct {
	BackupID    string    `json:"backup_id"`
	Valid       bool      `json:"valid"`
	Issues      []string  `json:"issues,omitempty"`
	ValidatedAt time.Time `json:"validated_at"`
}

// ThroughputBps представляет пропускную способность в байтах в секунду
type ThroughputBps int64

// NetworkIO представляет сетевой ввод-вывод в байтах
type NetworkIO int64

// ResponseTimeMs представляет время отклика в миллисекундах
type ResponseTimeMs int64

// BandwidthLimitBps представляет лимит пропускной способности в байтах в секунду
type BandwidthLimitBps int64

// KeepAliveMs представляет интервал keep-alive в миллисекундах
type KeepAliveMs int64

// MTU представляет максимальный размер передаваемого блока
type MTU int

// Protocol представляет сетевой протокол
type Protocol string

const (
	ProtocolQUIC      Protocol = "quic"
	ProtocolWebSocket Protocol = "websocket"
	ProtocolTCP       Protocol = "tcp"
	ProtocolUDP       Protocol = "udp"
	ProtocolHTTP      Protocol = "http"
	ProtocolHTTPS     Protocol = "https"
)

// Algorithm представляет алгоритм шифрования
type Algorithm string

const (
	AlgorithmAES256      Algorithm = "aes256"
	AlgorithmChaCha20    Algorithm = "chacha20"
)

// ModelType представляет тип модели машинного обучения
type ModelType string

const (
	ModelTypeThreatDetection  ModelType = "threat_detection"
	ModelTypeBehaviorAnalysis ModelType = "behavior_analysis"
	ModelTypeAnomalyDetection ModelType = "anomaly_detection"
	ModelTypeRecommendation   ModelType = "recommendation"
)

// EventType представляет тип события
type EventType string

const (
	EventTypeSecurity      EventType = "security"
	EventTypeAudit         EventType = "audit"
	EventTypePerformance   EventType = "performance"
	EventTypeScaling       EventType = "scaling"
	EventTypeConnection    EventType = "connection"
	EventTypeDisconnection EventType = "disconnection"
	EventTypeError         EventType = "error"
	EventTypeWarning       EventType = "warning"
	EventTypeInfo          EventType = "info"
)

// ResourceType представляет тип ресурса
type ResourceType string

const (
	ResourceTypeCPU       ResourceType = "cpu"
	ResourceTypeMemory    ResourceType = "memory"
	ResourceTypeNetwork   ResourceType = "network"
	ResourceTypeStorage   ResourceType = "storage"
	ResourceTypeBandwidth ResourceType = "bandwidth"
)

// ScalingDirection представляет направление масштабирования
type ScalingDirection string

const (
	ScalingDirectionUp   ScalingDirection = "up"
	ScalingDirectionDown ScalingDirection = "down"
	ScalingDirectionNone ScalingDirection = "none"
)

// ValidationResult представляет результат валидации
type ValidationResult struct {
	Valid       bool      `json:"valid"`
	Issues      []string  `json:"issues,omitempty"`
	Warnings    []string  `json:"warnings,omitempty"`
	ValidatedAt time.Time `json:"validated_at"`
}

// PaginationQuery представляет параметры пагинации
type PaginationQuery struct {
	PageSize  int    `json:"page_size"`
	PageToken string `json:"page_token,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"` // "asc" or "desc"
}

// PaginationResult представляет результат пагинации
type PaginationResult struct {
	NextPageToken string `json:"next_page_token,omitempty"`
	HasMore       bool   `json:"has_more"`
	TotalCount    int    `json:"total_count,omitempty"`
}

// FilterQuery представляет параметры фильтрации
type FilterQuery struct {
	TimeRange TimeRange         `json:"time_range"`
	Filters   map[string]string `json:"filters,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
}

// Metrics представляет базовые метрики
type Metrics struct {
	CPUUsage          float64        `json:"cpu_usage"`
	MemoryUsage       float64        `json:"memory_usage"`
	NetworkIO         NetworkIO      `json:"network_io"`
	ResponseTime      ResponseTimeMs `json:"response_time_ms"`
	ActiveConnections int            `json:"active_connections"`
	CollectedAt       time.Time      `json:"collected_at"`
}

// HealthStatus представляет статус здоровья
type HealthStatus struct {
	Status      Status      `json:"status"`
	HealthScore HealthScore `json:"health_score"`
	LastCheck   time.Time   `json:"last_check"`
	Issues      []string    `json:"issues,omitempty"`
}

// ErrorInfo представляет информацию об ошибке
type ErrorInfo struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Retryable bool      `json:"retryable"`
	Category  string    `json:"category,omitempty"`
}

// Metadata представляет метаданные
type Metadata struct {
	SchemaVersion SchemaVersion     `json:"schema_version,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Tags          []string          `json:"tags,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}
