package interfaces

import (
	"context"
	"time"
)

// P2PServiceInterface интерфейс для P2P сервиса
type P2PServiceInterface interface {
	// Создание и управление туннелями
	CreateTunnel(ctx context.Context, peerID, endpoint string, config map[string]interface{}) (*TunnelResult, error)
	UpdateTunnel(ctx context.Context, tunnelID string, config map[string]interface{}) (*TunnelResult, error)
	DeleteTunnel(ctx context.Context, tunnelID string) error
	MonitorTunnel(ctx context.Context, tunnelID string) (*MonitorResult, error)

	// Управление пирами
	RegisterPeer(ctx context.Context, peerID string, endpoint string) error
	UnregisterPeer(ctx context.Context, peerID string) error
	GetPeerStatus(ctx context.Context, peerID string) (*PeerStatus, error)

	// Анализ топологии
	AnalyzeTopology(ctx context.Context) (*TopologyAnalysis, error)
	PlanTopology(ctx context.Context, requirements map[string]interface{}) (*TopologyPlan, error)
	ValidateConnection(ctx context.Context, peerID1, peerID2 string) (*ConnectionValidation, error)
}

// AIServiceInterface интерфейс для AI/ML сервиса
type AIServiceInterface interface {
	// Анализ угроз
	CollectThreatData(ctx context.Context, timeRange time.Duration) (*ThreatData, error)
	PreprocessThreatData(ctx context.Context, data *ThreatData) (*ProcessedThreatData, error)
	PerformThreatAnalysis(ctx context.Context, data *ProcessedThreatData) (*ThreatAnalysis, error)

	// Анализ поведения
	AnalyzeUserBehavior(ctx context.Context, userID string, timeRange time.Duration) (*BehaviorAnalysis, error)
	GenerateRecommendations(ctx context.Context, analysis *BehaviorAnalysis) (*AIRecommendations, error)

	// Обучение моделей
	TrainModel(ctx context.Context, modelType string, trainingData interface{}) (*ModelTraining, error)
	UpdateModel(ctx context.Context, modelID string, newData interface{}) (*ModelUpdate, error)
}

// SecurityServiceInterface интерфейс для сервиса безопасности
type SecurityServiceInterface interface {
	// Соответствие требованиям
	GetComplianceRequirements(ctx context.Context, standard string) (*ComplianceRequirements, error)
	AuditSecurityArea(ctx context.Context, area string) (*SecurityAudit, error)
	CalculateComplianceScore(ctx context.Context, audit *SecurityAudit) (*ComplianceScore, error)

	// Логирование аудита
	LogSecurityEvent(ctx context.Context, event *SecurityEvent) error
	GetAuditLogs(ctx context.Context, timeRange time.Duration) (*AuditLogs, error)
}

// ScalingServiceInterface интерфейс для автомасштабирования
type ScalingServiceInterface interface {
	// Сбор метрик
	CollectScalingMetrics(ctx context.Context) (*ScalingMetrics, error)
	AnalyzeScalingNeed(ctx context.Context, metrics *ScalingMetrics) (*ScalingNeed, error)

	// Выполнение масштабирования
	PerformScaling(ctx context.Context, action *ScalingAction) (*ScalingResult, error)
	ValidateScaling(ctx context.Context, result *ScalingResult) (*ScalingValidation, error)
}

// Результаты и структуры данных

type TunnelResult struct {
	TunnelID      string                 `json:"tunnel_id"`
	Status        string                 `json:"status"`
	PeerEndpoints []string               `json:"peer_endpoints"`
	Config        map[string]interface{} `json:"config"`
	CreatedAt     time.Time              `json:"created_at"`
}

type MonitorResult struct {
	TunnelID    string    `json:"tunnel_id"`
	HealthScore float64   `json:"health_score"`
	Latency     int64     `json:"latency_ms"`
	Throughput  int64     `json:"throughput_bps"`
	Errors      int       `json:"error_count"`
	LastCheck   time.Time `json:"last_check"`
}

type PeerStatus struct {
	PeerID     string    `json:"peer_id"`
	Status     string    `json:"status"`
	Endpoint   string    `json:"endpoint"`
	LastSeen   time.Time `json:"last_seen"`
	Connection string    `json:"connection_type"`
}

type TopologyAnalysis struct {
	TotalPeers      int              `json:"total_peers"`
	ActivePeers     int              `json:"active_peers"`
	Connections     []ConnectionInfo `json:"connections"`
	Bottlenecks     []BottleneckInfo `json:"bottlenecks"`
	Recommendations []string         `json:"recommendations"`
}

type TopologyPlan struct {
	PlanID        string                 `json:"plan_id"`
	Actions       []TopologyAction       `json:"actions"`
	ExpectedGains map[string]interface{} `json:"expected_gains"`
	RiskLevel     string                 `json:"risk_level"`
}

type ConnectionInfo struct {
	FromPeer   string  `json:"from_peer"`
	ToPeer     string  `json:"to_peer"`
	Latency    int64   `json:"latency_ms"`
	Throughput int64   `json:"throughput_bps"`
	Quality    float64 `json:"quality_score"`
}

type BottleneckInfo struct {
	Location string  `json:"location"`
	Severity string  `json:"severity"`
	Impact   float64 `json:"impact_score"`
	Solution string  `json:"solution"`
}

type TopologyAction struct {
	ActionType string                 `json:"action_type"`
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
}

type ConnectionValidation struct {
	Valid           bool     `json:"valid"`
	Latency         int64    `json:"latency_ms"`
	Quality         float64  `json:"quality_score"`
	Issues          []string `json:"issues"`
	Recommendations []string `json:"recommendations"`
}

type ThreatData struct {
	DataID     string                 `json:"data_id"`
	Source     string                 `json:"source"`
	TimeRange  time.Duration          `json:"time_range"`
	DataPoints []ThreatDataPoint      `json:"data_points"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type ThreatDataPoint struct {
	Timestamp  time.Time              `json:"timestamp"`
	ThreatType string                 `json:"threat_type"`
	Severity   string                 `json:"severity"`
	Source     string                 `json:"source"`
	Details    map[string]interface{} `json:"details"`
}

type ProcessedThreatData struct {
	DataID      string                 `json:"data_id"`
	ProcessedAt time.Time              `json:"processed_at"`
	Features    []string               `json:"features"`
	Normalized  []float64              `json:"normalized_data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type ThreatAnalysis struct {
	AnalysisID      string       `json:"analysis_id"`
	ThreatLevel     string       `json:"threat_level"`
	Confidence      float64      `json:"confidence"`
	Threats         []ThreatInfo `json:"threats"`
	Recommendations []string     `json:"recommendations"`
	ProcessedAt     time.Time    `json:"processed_at"`
}

type ThreatInfo struct {
	ThreatID    string  `json:"threat_id"`
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Description string  `json:"description"`
	Mitigation  string  `json:"mitigation"`
	Confidence  float64 `json:"confidence"`
}

type BehaviorAnalysis struct {
	UserID     string            `json:"user_id"`
	AnalysisID string            `json:"analysis_id"`
	Patterns   []BehaviorPattern `json:"patterns"`
	Anomalies  []AnomalyInfo     `json:"anomalies"`
	RiskScore  float64           `json:"risk_score"`
	AnalyzedAt time.Time         `json:"analyzed_at"`
}

type BehaviorPattern struct {
	PatternID   string  `json:"pattern_id"`
	Type        string  `json:"type"`
	Frequency   float64 `json:"frequency"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

type AnomalyInfo struct {
	AnomalyID   string    `json:"anomaly_id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	DetectedAt  time.Time `json:"detected_at"`
}

type AIRecommendations struct {
	RecommendationID string    `json:"recommendation_id"`
	Type             string    `json:"type"`
	Priority         string    `json:"priority"`
	Description      string    `json:"description"`
	Actions          []string  `json:"actions"`
	ExpectedOutcome  string    `json:"expected_outcome"`
	Confidence       float64   `json:"confidence"`
	GeneratedAt      time.Time `json:"generated_at"`
}

type ModelTraining struct {
	ModelID      string                 `json:"model_id"`
	Type         string                 `json:"type"`
	Status       string                 `json:"status"`
	Accuracy     float64                `json:"accuracy"`
	TrainingTime time.Duration          `json:"training_time"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type ModelUpdate struct {
	ModelID     string    `json:"model_id"`
	UpdateID    string    `json:"update_id"`
	Status      string    `json:"status"`
	Improvement float64   `json:"improvement"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ComplianceRequirements struct {
	Standard     string            `json:"standard"`
	Version      string            `json:"version"`
	Requirements []RequirementInfo `json:"requirements"`
	Deadline     time.Time         `json:"deadline"`
}

type RequirementInfo struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
}

type SecurityAudit struct {
	AuditID   string            `json:"audit_id"`
	Area      string            `json:"area"`
	Status    string            `json:"status"`
	Findings  []SecurityFinding `json:"findings"`
	Score     float64           `json:"score"`
	AuditedAt time.Time         `json:"audited_at"`
}

type SecurityFinding struct {
	FindingID   string `json:"finding_id"`
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Remediation string `json:"remediation"`
	Status      string `json:"status"`
}

type ComplianceScore struct {
	ScoreID         string             `json:"score_id"`
	OverallScore    float64            `json:"overall_score"`
	CategoryScores  map[string]float64 `json:"category_scores"`
	Recommendations []string           `json:"recommendations"`
	CalculatedAt    time.Time          `json:"calculated_at"`
}

type SecurityEvent struct {
	EventID     string                 `json:"event_id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type AuditLogs struct {
	Logs        []AuditLogEntry `json:"logs"`
	TotalCount  int             `json:"total_count"`
	TimeRange   time.Duration   `json:"time_range"`
	RetrievedAt time.Time       `json:"retrieved_at"`
}

type AuditLogEntry struct {
	EntryID   string                 `json:"entry_id"`
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	Timestamp time.Time              `json:"timestamp"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type ScalingMetrics struct {
	MetricsID         string    `json:"metrics_id"`
	CPUUsage          float64   `json:"cpu_usage"`
	MemoryUsage       float64   `json:"memory_usage"`
	NetworkIO         int64     `json:"network_io"`
	ActiveConnections int       `json:"active_connections"`
	ResponseTime      int64     `json:"response_time_ms"`
	CollectedAt       time.Time `json:"collected_at"`
}

type ScalingNeed struct {
	NeedID       string    `json:"need_id"`
	Action       string    `json:"action"` // scale_up, scale_down, no_action
	Reason       string    `json:"reason"`
	Priority     string    `json:"priority"`
	TargetCount  int       `json:"target_count"`
	CurrentCount int       `json:"current_count"`
	AnalyzedAt   time.Time `json:"analyzed_at"`
}

type ScalingAction struct {
	ActionID   string                 `json:"action_id"`
	Type       string                 `json:"type"`
	Target     string                 `json:"target"`
	Count      int                    `json:"count"`
	Parameters map[string]interface{} `json:"parameters"`
}

type ScalingResult struct {
	ResultID      string        `json:"result_id"`
	Status        string        `json:"status"`
	PreviousCount int           `json:"previous_count"`
	NewCount      int           `json:"new_count"`
	Duration      time.Duration `json:"duration"`
	CompletedAt   time.Time     `json:"completed_at"`
}

type ScalingValidation struct {
	ValidationID    string    `json:"validation_id"`
	Valid           bool      `json:"valid"`
	Issues          []string  `json:"issues"`
	Recommendations []string  `json:"recommendations"`
	ValidatedAt     time.Time `json:"validated_at"`
}
