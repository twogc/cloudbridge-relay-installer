package analytics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AIAnalyticsEngine представляет основной движок AI-аналитики
type AIAnalyticsEngine struct {
	mu                 sync.RWMutex
	dataCollectors     map[string]DataCollector
	llmClient          LLMClient
	reportGenerator    interface{} // Будет определен позже
	notificationSystem interface{} // Будет определен позже
	scheduler          interface{} // Будет определен позже
	logger             *zap.Logger
	reports            []*Report
	maxReports         int
	config             *AnalyticsConfig
}

// DataCollector интерфейс для сборщиков данных
type DataCollector interface {
	CollectMetrics(ctx context.Context) (*MetricsData, error)
	GetName() string
}

// LLMClient интерфейс для LLM клиентов
type LLMClient interface {
	GenerateAnalysis(data *MetricsData) (*AnalysisResult, error)
}

// MetricsData представляет собранные метрики
type MetricsData struct {
	Timestamp time.Time
	Source    string
	Metrics   map[string]interface{}
	Anomalies []Anomaly
	Trends    []Trend
	Health    *HealthStatus
}

// Anomaly представляет обнаруженную аномалию
type Anomaly struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	Timestamp   time.Time `json:"timestamp"`
}

// Trend представляет обнаруженный тренд
type Trend struct {
	Metric     string  `json:"metric"`
	Direction  string  `json:"direction"`
	Change     float64 `json:"change"`
	Confidence float64 `json:"confidence"`
	Period     string  `json:"period"`
}

// HealthStatus представляет статус здоровья системы
type HealthStatus struct {
	Overall    string            `json:"overall"`
	Components map[string]string `json:"components"`
	Score      float64           `json:"score"`
}

// AnalysisResult представляет результат анализа от LLM
type AnalysisResult struct {
	Summary         string   `json:"summary"`
	Insights        []string `json:"insights"`
	Recommendations []string `json:"recommendations"`
	CriticalIssues  []string `json:"critical_issues"`
	Performance     string   `json:"performance"`
}

// Report представляет сгенерированный отчет
type Report struct {
	ID               string          `json:"id"`
	Timestamp        time.Time       `json:"timestamp"`
	Analysis         *AnalysisResult `json:"analysis"`
	Metrics          *MetricsData    `json:"metrics"`
	Content          string          `json:"content"`
	HTMLContent      string          `json:"html_content"`
	NotificationSent bool            `json:"notification_sent"`
}

// AnalyticsConfig конфигурация для AI-аналитики
type AnalyticsConfig struct {
	Enabled       bool               `yaml:"enabled"`
	Schedule      string             `yaml:"schedule"`
	MaxReports    int                `yaml:"max_reports"`
	RetentionDays int                `yaml:"retention_days"`
	CloudflareAI  CloudflareAIConfig `yaml:"cloudflare_ai"`
	Notifications NotificationConfig `yaml:"notifications"`
	DataSources   DataSourcesConfig  `yaml:"data_sources"`
}

// CloudflareAIConfig конфигурация для Cloudflare AI
type CloudflareAIConfig struct {
	AccountID string `yaml:"account_id"`
	APIToken  string `yaml:"api_token"`
	Model     string `yaml:"model"`
	MaxTokens int    `yaml:"max_tokens"`
}

// NotificationConfig конфигурация уведомлений
type NotificationConfig struct {
	Slack    SlackConfig `yaml:"slack"`
	Email    EmailConfig `yaml:"email"`
	Webhooks []string    `yaml:"webhooks"`
}

// SlackConfig конфигурация Slack
type SlackConfig struct {
	Enabled    bool   `yaml:"enabled"`
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
}

// EmailConfig конфигурация Email
type EmailConfig struct {
	Enabled    bool     `yaml:"enabled"`
	SMTPHost   string   `yaml:"smtp_host"`
	SMTPPort   int      `yaml:"smtp_port"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	Recipients []string `yaml:"recipients"`
}

// DataSourcesConfig конфигурация источников данных
type DataSourcesConfig struct {
	Prometheus PrometheusConfig `yaml:"prometheus"`
	Zabbix     ZabbixConfig     `yaml:"zabbix"`
	AIService  AIServiceConfig  `yaml:"ai_service"`
}

// PrometheusConfig конфигурация Prometheus
type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
}

// ZabbixConfig конфигурация Zabbix
type ZabbixConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
}

// AIServiceConfig конфигурация AI Service
type AIServiceConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
}

// NewAIAnalyticsEngine создает новый движок AI-аналитики
func NewAIAnalyticsEngine(config *AnalyticsConfig, logger *zap.Logger) *AIAnalyticsEngine {
	return &AIAnalyticsEngine{
		dataCollectors: make(map[string]DataCollector),
		scheduler:      nil, // Будет инициализирован позже
		logger:         logger,
		reports:        make([]*Report, 0),
		maxReports:     config.MaxReports,
		config:         config,
	}
}

// AddCollector добавляет сборщик данных
func (e *AIAnalyticsEngine) AddCollector(collector DataCollector) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.dataCollectors[collector.GetName()] = collector
	e.logger.Info("Added data collector", zap.String("name", collector.GetName()))
}

// SetLLMClient устанавливает LLM клиент
func (e *AIAnalyticsEngine) SetLLMClient(client LLMClient) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.llmClient = client
}

// SetNotificationSystem устанавливает систему уведомлений
func (e *AIAnalyticsEngine) SetNotificationSystem(system interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.notificationSystem = system
}

// SetReportGenerator устанавливает генератор отчетов
func (e *AIAnalyticsEngine) SetReportGenerator(generator interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.reportGenerator = generator
}

// StartScheduler запускает планировщик
func (e *AIAnalyticsEngine) StartScheduler() {
	if e.config.Schedule == "" {
		e.config.Schedule = "*/15 * * * *" // Каждые 15 минут по умолчанию
	}

	// Планировщик будет реализован позже
	e.logger.Info("Scheduler started", zap.String("schedule", e.config.Schedule))
}

// GenerateReport генерирует отчет
func (e *AIAnalyticsEngine) GenerateReport(ctx context.Context) error {
	e.logger.Info("Generating AI analytics report...")

	// Сбор данных из всех источников
	metricsData := &MetricsData{
		Timestamp: time.Now(),
		Metrics:   make(map[string]interface{}),
		Anomalies: make([]Anomaly, 0),
		Trends:    make([]Trend, 0),
	}

	for name, collector := range e.dataCollectors {
		data, err := collector.CollectMetrics(ctx)
		if err != nil {
			e.logger.Error("Failed to collect metrics",
				zap.String("collector", name),
				zap.Error(err))
			continue
		}

		metricsData.Metrics[name] = data.Metrics
		metricsData.Anomalies = append(metricsData.Anomalies, data.Anomalies...)
		metricsData.Trends = append(metricsData.Trends, data.Trends...)
	}

	// Анализ данных с помощью LLM
	var analysis *AnalysisResult
	var err error

	if e.llmClient != nil {
		analysis, err = e.llmClient.GenerateAnalysis(metricsData)
		if err != nil {
			e.logger.Error("Failed to generate analysis", zap.Error(err))
			// Создаем базовый анализ в случае ошибки
			analysis = e.createFallbackAnalysis(metricsData)
		}
	} else {
		e.logger.Warn("LLM client not available, using fallback analysis")
		analysis = e.createFallbackAnalysis(metricsData)
	}

	// Создание отчета
	report := &Report{
		ID:        fmt.Sprintf("report_%d", time.Now().Unix()),
		Timestamp: time.Now(),
		Analysis:  analysis,
		Metrics:   metricsData,
	}

	// Генерация контента отчета
	if e.reportGenerator != nil {
		if generator, ok := e.reportGenerator.(interface{ GenerateMarkdown(*Report) (string, error) }); ok {
			content, err := generator.GenerateMarkdown(report)
			if err != nil {
				e.logger.Error("Failed to generate report content", zap.Error(err))
			} else {
				report.Content = content
			}
		}
	}

	// Отправка уведомлений
	if e.notificationSystem != nil {
		if notifier, ok := e.notificationSystem.(interface{ SendReport(*Report) error }); ok {
			if err := notifier.SendReport(report); err != nil {
				e.logger.Error("Failed to send notification", zap.Error(err))
			} else {
				report.NotificationSent = true
			}
		}
	}

	// Сохранение отчета
	e.mu.Lock()
	e.reports = append(e.reports, report)
	if len(e.reports) > e.maxReports {
		e.reports = e.reports[1:]
	}
	e.mu.Unlock()

	e.logger.Info("Report generated successfully",
		zap.String("report_id", report.ID),
		zap.Bool("notification_sent", report.NotificationSent))

	return nil
}

// createFallbackAnalysis создает базовый анализ в случае недоступности LLM
func (e *AIAnalyticsEngine) createFallbackAnalysis(data *MetricsData) *AnalysisResult {
	insights := make([]string, 0)
	recommendations := make([]string, 0)
	criticalIssues := make([]string, 0)

	// Анализ аномалий
	if len(data.Anomalies) > 0 {
		criticalIssues = append(criticalIssues, fmt.Sprintf("Обнаружено %d аномалий", len(data.Anomalies)))
		for _, anomaly := range data.Anomalies {
			if anomaly.Severity == "critical" {
				criticalIssues = append(criticalIssues, anomaly.Description)
			}
		}
	}

	// Анализ здоровья системы
	if data.Health != nil {
		if data.Health.Score < 50 {
			criticalIssues = append(criticalIssues, "Критическое состояние системы")
		} else if data.Health.Score < 80 {
			insights = append(insights, "Система работает с предупреждениями")
		} else {
			insights = append(insights, "Система работает стабильно")
		}
	}

	// Базовые рекомендации
	if len(criticalIssues) > 0 {
		recommendations = append(recommendations, "Требуется немедленное вмешательство")
	}

	return &AnalysisResult{
		Summary:         "Автоматический анализ метрик CloudBridge Relay",
		Insights:        insights,
		Recommendations: recommendations,
		CriticalIssues:  criticalIssues,
		Performance:     "Анализ выполнен в базовом режиме",
	}
}

// SetupRoutes настраивает маршруты API
func (e *AIAnalyticsEngine) SetupRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/ai/analytics")
	{
		api.POST("/generate-report", e.handleGenerateReport)
		api.GET("/reports", e.handleGetReports)
		api.GET("/reports/:id", e.handleGetReport)
		api.GET("/insights", e.handleGetInsights)
		api.POST("/configure-notifications", e.handleConfigureNotifications)
		api.GET("/status", e.handleGetStatus)
	}
}

// handleGenerateReport обрабатывает запрос на генерацию отчета
func (e *AIAnalyticsEngine) handleGenerateReport(c *gin.Context) {
	if err := e.GenerateReport(c.Request.Context()); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Report generated successfully"})
}

// handleGetReports обрабатывает запрос на получение отчетов
func (e *AIAnalyticsEngine) handleGetReports(c *gin.Context) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	c.JSON(200, gin.H{"reports": e.reports})
}

// handleGetReport обрабатывает запрос на получение конкретного отчета
func (e *AIAnalyticsEngine) handleGetReport(c *gin.Context) {
	id := c.Param("id")
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, report := range e.reports {
		if report.ID == id {
			c.JSON(200, report)
			return
		}
	}
	c.JSON(404, gin.H{"error": "Report not found"})
}

// handleGetInsights обрабатывает запрос на получение инсайтов
func (e *AIAnalyticsEngine) handleGetInsights(c *gin.Context) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.reports) == 0 {
		c.JSON(200, gin.H{"insights": []string{}})
		return
	}

	latestReport := e.reports[len(e.reports)-1]
	c.JSON(200, gin.H{
		"insights":        latestReport.Analysis.Insights,
		"recommendations": latestReport.Analysis.Recommendations,
		"timestamp":       latestReport.Timestamp,
	})
}

// handleConfigureNotifications обрабатывает запрос на настройку уведомлений
func (e *AIAnalyticsEngine) handleConfigureNotifications(c *gin.Context) {
	var config NotificationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if e.notificationSystem != nil {
		if notifier, ok := e.notificationSystem.(interface{ Configure(NotificationConfig) }); ok {
			notifier.Configure(config)
			c.JSON(200, gin.H{"message": "Notifications configured successfully"})
		} else {
			c.JSON(500, gin.H{"error": "Notification system does not support configuration"})
		}
	} else {
		c.JSON(500, gin.H{"error": "Notification system not initialized"})
	}
}

// handleGetStatus обрабатывает запрос на получение статуса
func (e *AIAnalyticsEngine) handleGetStatus(c *gin.Context) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	status := gin.H{
		"status":                  "running",
		"collectors":              len(e.dataCollectors),
		"reports_count":           len(e.reports),
		"llm_available":           e.llmClient != nil,
		"notifications_available": e.notificationSystem != nil,
		"schedule":                e.config.Schedule,
	}

	c.JSON(200, status)
}

// Stop останавливает движок аналитики
func (e *AIAnalyticsEngine) Stop() {
	// Планировщик будет реализован позже
	e.logger.Info("AI Analytics Engine stopped")
}
