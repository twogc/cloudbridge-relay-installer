# 🤖 Демо-реализация AI-аналитика для CloudBridge Relay

## 📋 Обзор

Этот документ содержит **конкретную реализацию** AI-аналитика, который можно сразу интегрировать в существующую систему CloudBridge Relay.

---

## 🚀 **БЫСТРЫЙ СТАРТ**

### **1. Создание AI Analytics Service**

```go
// cmd/ai-analytics/main.go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    
    "cloudbridge-relay/internal/ai/analytics"
    "cloudbridge-relay/internal/ai/analytics/collectors"
    "cloudbridge-relay/internal/ai/analytics/llm"
    "cloudbridge-relay/internal/ai/analytics/notifications"
    "cloudbridge-relay/internal/ai/analytics/reports"
)

func main() {
    // Инициализация логгера
    logger, _ := zap.NewProduction()
    defer logger.Sync()
    
    // Создание AI Analytics Engine
    engine := analytics.NewAIAnalyticsEngine(logger)
    
    // Добавление сборщиков данных
    engine.AddCollector(collectors.NewPrometheusCollector("http://192.168.49.2:30083/metrics"))
    engine.AddCollector(collectors.NewZabbixCollector("http://localhost:10051"))
    engine.AddCollector(collectors.NewAIServiceCollector("http://192.168.49.2:30083"))
    
    // Настройка LLM клиента
    llmClient := llm.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"))
    engine.SetLLMClient(llmClient)
    
    // Настройка системы уведомлений
    notificationSystem := notifications.NewNotificationSystem()
    notificationSystem.AddSlackWebhook(os.Getenv("SLACK_WEBHOOK_URL"))
    engine.SetNotificationSystem(notificationSystem)
    
    // Запуск планировщика (каждые 15 минут)
    engine.StartScheduler("*/15 * * * *")
    
    // Запуск HTTP сервера
    router := gin.Default()
    engine.SetupRoutes(router)
    
    // Graceful shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        logger.Info("Shutting down AI Analytics Service...")
        engine.Stop()
        os.Exit(0)
    }()
    
    logger.Info("AI Analytics Service started on port 8084")
    router.Run(":8084")
}
```

### **2. Основной движок аналитики**

```go
// internal/ai/analytics/analytics_engine.go
package analytics

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/robfig/cron/v3"
    "go.uber.org/zap"
)

type AIAnalyticsEngine struct {
    mu                sync.RWMutex
    dataCollectors    map[string]DataCollector
    llmClient         LLMClient
    reportGenerator   *ReportGenerator
    notificationSystem *NotificationSystem
    scheduler         *cron.Cron
    logger            *zap.Logger
    reports           []*Report
    maxReports        int
}

type DataCollector interface {
    CollectMetrics(ctx context.Context) (*MetricsData, error)
    GetName() string
}

type LLMClient interface {
    GenerateAnalysis(data *MetricsData) (*AnalysisResult, error)
}

type MetricsData struct {
    Timestamp   time.Time
    Source      string
    Metrics     map[string]interface{}
    Anomalies   []Anomaly
    Trends      []Trend
    Health      *HealthStatus
}

type Anomaly struct {
    Type        string    `json:"type"`
    Severity    string    `json:"severity"`
    Description string    `json:"description"`
    Value       float64   `json:"value"`
    Threshold   float64   `json:"threshold"`
    Timestamp   time.Time `json:"timestamp"`
}

type Trend struct {
    Metric      string  `json:"metric"`
    Direction   string  `json:"direction"`
    Change      float64 `json:"change"`
    Confidence  float64 `json:"confidence"`
    Period      string  `json:"period"`
}

type HealthStatus struct {
    Overall    string            `json:"overall"`
    Components map[string]string `json:"components"`
    Score      float64           `json:"score"`
}

type AnalysisResult struct {
    Summary         string   `json:"summary"`
    Insights        []string `json:"insights"`
    Recommendations []string `json:"recommendations"`
    CriticalIssues  []string `json:"critical_issues"`
    Performance     string   `json:"performance"`
}

type Report struct {
    ID              string          `json:"id"`
    Timestamp       time.Time       `json:"timestamp"`
    Analysis        *AnalysisResult `json:"analysis"`
    Metrics         *MetricsData    `json:"metrics"`
    Content         string          `json:"content"`
    HTMLContent     string          `json:"html_content"`
    NotificationSent bool           `json:"notification_sent"`
}

func NewAIAnalyticsEngine(logger *zap.Logger) *AIAnalyticsEngine {
    return &AIAnalyticsEngine{
        dataCollectors: make(map[string]DataCollector),
        scheduler:      cron.New(),
        logger:         logger,
        reports:        make([]*Report, 0),
        maxReports:     100,
    }
}

func (e *AIAnalyticsEngine) AddCollector(collector DataCollector) {
    e.mu.Lock()
    defer e.mu.Unlock()
    e.dataCollectors[collector.GetName()] = collector
    e.logger.Info("Added data collector", zap.String("name", collector.GetName()))
}

func (e *AIAnalyticsEngine) SetLLMClient(client LLMClient) {
    e.mu.Lock()
    defer e.mu.Unlock()
    e.llmClient = client
}

func (e *AIAnalyticsEngine) SetNotificationSystem(system *NotificationSystem) {
    e.mu.Lock()
    defer e.mu.Unlock()
    e.notificationSystem = system
}

func (e *AIAnalyticsEngine) StartScheduler(schedule string) {
    e.scheduler.AddFunc(schedule, func() {
        e.logger.Info("Starting scheduled analysis...")
        if err := e.GenerateReport(context.Background()); err != nil {
            e.logger.Error("Failed to generate report", zap.Error(err))
        }
    })
    e.scheduler.Start()
    e.logger.Info("Scheduler started", zap.String("schedule", schedule))
}

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
    analysis, err := e.llmClient.GenerateAnalysis(metricsData)
    if err != nil {
        return fmt.Errorf("failed to generate analysis: %w", err)
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
        content, err := e.reportGenerator.GenerateMarkdown(report)
        if err != nil {
            e.logger.Error("Failed to generate report content", zap.Error(err))
        } else {
            report.Content = content
        }
    }
    
    // Отправка уведомлений
    if e.notificationSystem != nil {
        if err := e.notificationSystem.SendReport(report); err != nil {
            e.logger.Error("Failed to send notification", zap.Error(err))
        } else {
            report.NotificationSent = true
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

func (e *AIAnalyticsEngine) SetupRoutes(router *gin.Engine) {
    api := router.Group("/api/v1/ai/analytics")
    {
        api.POST("/generate-report", e.handleGenerateReport)
        api.GET("/reports", e.handleGetReports)
        api.GET("/reports/:id", e.handleGetReport)
        api.GET("/insights", e.handleGetInsights)
        api.POST("/configure-notifications", e.handleConfigureNotifications)
    }
}

func (e *AIAnalyticsEngine) handleGenerateReport(c *gin.Context) {
    if err := e.GenerateReport(c.Request.Context()); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"message": "Report generated successfully"})
}

func (e *AIAnalyticsEngine) handleGetReports(c *gin.Context) {
    e.mu.RLock()
    defer e.mu.RUnlock()
    c.JSON(200, gin.H{"reports": e.reports})
}

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

func (e *AIAnalyticsEngine) handleGetInsights(c *gin.Context) {
    e.mu.RLock()
    defer e.mu.RUnlock()
    
    if len(e.reports) == 0 {
        c.JSON(200, gin.H{"insights": []string{}})
        return
    }
    
    latestReport := e.reports[len(e.reports)-1]
    c.JSON(200, gin.H{
        "insights": latestReport.Analysis.Insights,
        "recommendations": latestReport.Analysis.Recommendations,
        "timestamp": latestReport.Timestamp,
    })
}

func (e *AIAnalyticsEngine) handleConfigureNotifications(c *gin.Context) {
    var config NotificationConfig
    if err := c.ShouldBindJSON(&config); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    if e.notificationSystem != nil {
        e.notificationSystem.Configure(config)
        c.JSON(200, gin.H{"message": "Notifications configured successfully"})
    } else {
        c.JSON(500, gin.H{"error": "Notification system not initialized"})
    }
}

func (e *AIAnalyticsEngine) Stop() {
    e.scheduler.Stop()
    e.logger.Info("AI Analytics Engine stopped")
}
```

### **3. Сборщик данных Prometheus**

```go
// internal/ai/analytics/collectors/prometheus_collector.go
package collectors

import (
    "context"
    "fmt"
    "net/http"
    "strconv"
    "strings"
    "time"
    
    "go.uber.org/zap"
)

type PrometheusCollector struct {
    client *http.Client
    url    string
    logger *zap.Logger
}

func NewPrometheusCollector(url string) *PrometheusCollector {
    return &PrometheusCollector{
        client: &http.Client{Timeout: 10 * time.Second},
        url:    url,
        logger: zap.NewNop(),
    }
}

func (p *PrometheusCollector) GetName() string {
    return "prometheus"
}

func (p *PrometheusCollector) CollectMetrics(ctx context.Context) (*MetricsData, error) {
    resp, err := p.client.Get(p.url)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch metrics: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    // Парсинг метрик
    metrics := make(map[string]interface{})
    anomalies := make([]Anomaly, 0)
    
    // AI метрики
    aiMetrics := map[string]float64{
        "anomalies_detected": p.parseMetric(resp.Body, "ai_anomalies_detected_total"),
        "predictions_made":   p.parseMetric(resp.Body, "ai_predictions_made_total"),
        "cpu_usage":         p.parseMetric(resp.Body, "ai_cpu_usage_percent"),
        "memory_usage":      p.parseMetric(resp.Body, "ai_memory_usage_bytes"),
        "active_alerts":     p.parseMetric(resp.Body, "ai_active_alerts"),
        "prediction_accuracy": p.parseMetric(resp.Body, "ai_prediction_accuracy"),
    }
    
    // Relay метрики
    relayMetrics := map[string]float64{
        "active_connections": p.parseMetric(resp.Body, "relay_active_connections"),
        "connections_total":  p.parseMetric(resp.Body, "relay_connections_total"),
        "uptime_seconds":     p.parseMetric(resp.Body, "relay_uptime_seconds"),
    }
    
    metrics["ai"] = aiMetrics
    metrics["relay"] = relayMetrics
    
    // Детекция аномалий
    anomalies = append(anomalies, p.detectAnomalies(aiMetrics)...)
    anomalies = append(anomalies, p.detectAnomalies(relayMetrics)...)
    
    return &MetricsData{
        Timestamp: time.Now(),
        Source:    "prometheus",
        Metrics:   metrics,
        Anomalies: anomalies,
        Health:    p.assessHealth(aiMetrics, relayMetrics),
    }, nil
}

func (p *PrometheusCollector) parseMetric(body io.Reader, metricName string) float64 {
    scanner := bufio.NewScanner(body)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, metricName+" ") {
            parts := strings.Fields(line)
            if len(parts) >= 2 {
                if value, err := strconv.ParseFloat(parts[1], 64); err == nil {
                    return value
                }
            }
        }
    }
    return 0.0
}

func (p *PrometheusCollector) detectAnomalies(metrics map[string]float64) []Anomaly {
    anomalies := make([]Anomaly, 0)
    
    // Проверка CPU usage
    if cpuUsage, exists := metrics["cpu_usage"]; exists {
        if cpuUsage > 80.0 {
            anomalies = append(anomalies, Anomaly{
                Type:        "high_cpu_usage",
                Severity:    "warning",
                Description: fmt.Sprintf("High CPU usage: %.1f%%", cpuUsage),
                Value:       cpuUsage,
                Threshold:   80.0,
                Timestamp:   time.Now(),
            })
        }
    }
    
    // Проверка memory usage
    if memoryUsage, exists := metrics["memory_usage"]; exists {
        if memoryUsage > 1024*1024*1024 { // 1GB
            anomalies = append(anomalies, Anomaly{
                Type:        "high_memory_usage",
                Severity:    "warning",
                Description: fmt.Sprintf("High memory usage: %.1f MB", memoryUsage/1024/1024),
                Value:       memoryUsage,
                Threshold:   1024 * 1024 * 1024,
                Timestamp:   time.Now(),
            })
        }
    }
    
    // Проверка active alerts
    if activeAlerts, exists := metrics["active_alerts"]; exists {
        if activeAlerts > 0 {
            anomalies = append(anomalies, Anomaly{
                Type:        "active_alerts",
                Severity:    "critical",
                Description: fmt.Sprintf("Active alerts: %.0f", activeAlerts),
                Value:       activeAlerts,
                Threshold:   0,
                Timestamp:   time.Now(),
            })
        }
    }
    
    return anomalies
}

func (p *PrometheusCollector) assessHealth(aiMetrics, relayMetrics map[string]float64) *HealthStatus {
    components := make(map[string]string)
    score := 100.0
    
    // AI Service health
    if cpuUsage, exists := aiMetrics["cpu_usage"]; exists {
        if cpuUsage > 90 {
            components["ai_service"] = "critical"
            score -= 30
        } else if cpuUsage > 70 {
            components["ai_service"] = "warning"
            score -= 15
        } else {
            components["ai_service"] = "healthy"
        }
    }
    
    // Relay Service health
    if activeConnections, exists := relayMetrics["active_connections"]; exists {
        if activeConnections == 0 {
            components["relay_service"] = "critical"
            score -= 40
        } else if activeConnections < 10 {
            components["relay_service"] = "warning"
            score -= 20
        } else {
            components["relay_service"] = "healthy"
        }
    }
    
    overall := "healthy"
    if score < 50 {
        overall = "critical"
    } else if score < 80 {
        overall = "warning"
    }
    
    return &HealthStatus{
        Overall:    overall,
        Components: components,
        Score:      score,
    }
}
```

### **4. LLM клиент для OpenAI**

```go
// internal/ai/analytics/llm/openai_client.go
package llm

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
    client *openai.Client
    model  string
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
    return &OpenAIClient{
        client: openai.NewClient(apiKey),
        model:  "gpt-4",
    }
}

func (o *OpenAIClient) GenerateAnalysis(data *MetricsData) (*AnalysisResult, error) {
    prompt := o.buildPrompt(data)
    
    resp, err := o.client.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model: o.model,
            Messages: []openai.ChatCompletionMessage{
                {
                    Role:    openai.ChatMessageRoleSystem,
                    Content: "You are an expert system administrator analyzing CloudBridge Relay metrics. Provide clear, actionable insights in Russian. Focus on critical issues and practical recommendations.",
                },
                {
                    Role:    openai.ChatMessageRoleUser,
                    Content: prompt,
                },
            },
            MaxTokens:   1500,
            Temperature: 0.7,
        },
    )
    
    if err != nil {
        return nil, fmt.Errorf("failed to generate analysis: %w", err)
    }
    
    content := resp.Choices[0].Message.Content
    return o.parseAnalysis(content), nil
}

func (o *OpenAIClient) buildPrompt(data *MetricsData) string {
    aiMetrics := data.Metrics["ai"].(map[string]float64)
    relayMetrics := data.Metrics["relay"].(map[string]float64)
    
    prompt := fmt.Sprintf(`
Проанализируй следующие метрики CloudBridge Relay и предоставь инсайты:

🤖 AI Service Metrics:
- Обнаружено аномалий: %.0f
- Сделано предсказаний: %.0f
- Использование CPU: %.1f%%
- Использование памяти: %.1f MB
- Активные алерты: %.0f
- Точность предсказаний: %.1f%%

🌐 Relay Service Metrics:
- Активные соединения: %.0f
- Всего соединений: %.0f
- Время работы: %.0f секунд

🚨 Обнаруженные аномалии:
%s

📊 Статус здоровья: %s (оценка: %.1f/100)

Пожалуйста, предоставь:
1. Краткое резюме (2-3 предложения)
2. Ключевые инсайты и паттерны
3. Критические проблемы, требующие внимания
4. Практические рекомендации
5. Оценку производительности

Формат ответа:
**Резюме:** [краткое описание]
**Инсайты:** [список ключевых наблюдений]
**Критические проблемы:** [список проблем]
**Рекомендации:** [список действий]
**Производительность:** [оценка]
`,
        aiMetrics["anomalies_detected"],
        aiMetrics["predictions_made"],
        aiMetrics["cpu_usage"],
        aiMetrics["memory_usage"]/1024/1024,
        aiMetrics["active_alerts"],
        aiMetrics["prediction_accuracy"]*100,
        relayMetrics["active_connections"],
        relayMetrics["connections_total"],
        relayMetrics["uptime_seconds"],
        o.formatAnomalies(data.Anomalies),
        data.Health.Overall,
        data.Health.Score,
    )
    
    return prompt
}

func (o *OpenAIClient) formatAnomalies(anomalies []Anomaly) string {
    if len(anomalies) == 0 {
        return "Аномалий не обнаружено"
    }
    
    var result strings.Builder
    for _, anomaly := range anomalies {
        result.WriteString(fmt.Sprintf("- %s (%s): %s\n", 
            anomaly.Type, anomaly.Severity, anomaly.Description))
    }
    return result.String()
}

func (o *OpenAIClient) parseAnalysis(content string) *AnalysisResult {
    result := &AnalysisResult{
        Insights:        make([]string, 0),
        Recommendations: make([]string, 0),
        CriticalIssues:  make([]string, 0),
    }
    
    lines := strings.Split(content, "\n")
    currentSection := ""
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "**Резюме:**") {
            result.Summary = strings.TrimPrefix(line, "**Резюме:**")
            currentSection = "summary"
        } else if strings.HasPrefix(line, "**Инсайты:**") {
            currentSection = "insights"
        } else if strings.HasPrefix(line, "**Критические проблемы:**") {
            currentSection = "critical"
        } else if strings.HasPrefix(line, "**Рекомендации:**") {
            currentSection = "recommendations"
        } else if strings.HasPrefix(line, "**Производительность:**") {
            result.Performance = strings.TrimPrefix(line, "**Производительность:**")
            currentSection = "performance"
        } else if line != "" && strings.HasPrefix(line, "-") {
            switch currentSection {
            case "insights":
                result.Insights = append(result.Insights, strings.TrimPrefix(line, "- "))
            case "critical":
                result.CriticalIssues = append(result.CriticalIssues, strings.TrimPrefix(line, "- "))
            case "recommendations":
                result.Recommendations = append(result.Recommendations, strings.TrimPrefix(line, "- "))
            }
        }
    }
    
    return result
}
```

### **5. Система уведомлений**

```go
// internal/ai/analytics/notifications/notification_system.go
package notifications

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type NotificationSystem struct {
    slackWebhook string
    emailConfig  *EmailConfig
    webhooks     []string
}

type EmailConfig struct {
    SMTPHost     string
    SMTPPort     int
    Username     string
    Password     string
    Recipients   []string
}

type NotificationConfig struct {
    SlackWebhook string   `json:"slack_webhook"`
    EmailConfig  *EmailConfig `json:"email_config"`
    Webhooks     []string `json:"webhooks"`
}

func NewNotificationSystem() *NotificationSystem {
    return &NotificationSystem{
        webhooks: make([]string, 0),
    }
}

func (n *NotificationSystem) AddSlackWebhook(webhook string) {
    n.slackWebhook = webhook
}

func (n *NotificationSystem) Configure(config NotificationConfig) {
    n.slackWebhook = config.SlackWebhook
    n.emailConfig = config.EmailConfig
    n.webhooks = config.Webhooks
}

func (n *NotificationSystem) SendReport(report *Report) error {
    // Slack уведомление
    if n.slackWebhook != "" {
        if err := n.sendSlackNotification(report); err != nil {
            return fmt.Errorf("failed to send Slack notification: %w", err)
        }
    }
    
    // Webhook уведомления
    for _, webhook := range n.webhooks {
        if err := n.sendWebhookNotification(report, webhook); err != nil {
            // Логируем ошибку, но не прерываем процесс
            fmt.Printf("Failed to send webhook to %s: %v\n", webhook, err)
        }
    }
    
    return nil
}

func (n *NotificationSystem) sendSlackNotification(report *Report) error {
    payload := map[string]interface{}{
        "text": "🤖 CloudBridge AI Analytics Report",
        "blocks": []map[string]interface{}{
            {
                "type": "header",
                "text": map[string]interface{}{
                    "type": "plain_text",
                    "text": "🤖 CloudBridge AI Analytics Report",
                },
            },
            {
                "type": "section",
                "text": map[string]interface{}{
                    "type": "mrkdwn",
                    "text": fmt.Sprintf("*%s*\n\n%s", 
                        report.Timestamp.Format("2006-01-02 15:04:05"),
                        report.Analysis.Summary),
                },
            },
            {
                "type": "divider",
            },
        },
    }
    
    // Добавление критических проблем
    if len(report.Analysis.CriticalIssues) > 0 {
        issuesText := "*🚨 Критические проблемы:*\n"
        for _, issue := range report.Analysis.CriticalIssues {
            issuesText += fmt.Sprintf("• %s\n", issue)
        }
        
        payload["blocks"] = append(payload["blocks"].([]map[string]interface{}), map[string]interface{}{
            "type": "section",
            "text": map[string]interface{}{
                "type": "mrkdwn",
                "text": issuesText,
            },
        })
    }
    
    // Добавление рекомендаций
    if len(report.Analysis.Recommendations) > 0 {
        recommendationsText := "*💡 Рекомендации:*\n"
        for _, rec := range report.Analysis.Recommendations {
            recommendationsText += fmt.Sprintf("• %s\n", rec)
        }
        
        payload["blocks"] = append(payload["blocks"].([]map[string]interface{}), map[string]interface{}{
            "type": "section",
            "text": map[string]interface{}{
                "type": "mrkdwn",
                "text": recommendationsText,
            },
        })
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    
    resp, err := http.Post(n.slackWebhook, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    return nil
}

func (n *NotificationSystem) sendWebhookNotification(report *Report, webhookURL string) error {
    payload := map[string]interface{}{
        "timestamp": report.Timestamp,
        "report_id": report.ID,
        "analysis":  report.Analysis,
        "health":    report.Metrics.Health,
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    
    resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    return nil
}
```

---

## 🚀 **ЗАПУСК ДЕМО**

### **1. Установка зависимостей**

```bash
# Добавить в go.mod
go get github.com/sashabaranov/go-openai
go get github.com/robfig/cron/v3
go get github.com/gin-gonic/gin
```

### **2. Настройка переменных окружения**

```bash
export OPENAI_API_KEY="your-openai-api-key"
export SLACK_WEBHOOK_URL="your-slack-webhook-url"
```

### **3. Запуск сервиса**

```bash
go run cmd/ai-analytics/main.go
```

### **4. Тестирование API**

```bash
# Генерация отчета
curl -X POST http://localhost:8084/api/v1/ai/analytics/generate-report

# Получение отчетов
curl http://localhost:8084/api/v1/ai/analytics/reports

# Получение инсайтов
curl http://localhost:8084/api/v1/ai/analytics/insights
```

---

## 📊 **ПРИМЕР ВЫВОДА**

### **Slack уведомление:**
```
🤖 CloudBridge AI Analytics Report

2025-09-17 19:30:00

**Резюме:** Система работает стабильно с небольшими предупреждениями по использованию ресурсов. AI сервис показывает хорошую производительность с точностью предсказаний 85%.

🚨 Критические проблемы:
• Активные алерты: 2

💡 Рекомендации:
• Рекомендуется проверить активные алерты
• Рассмотреть оптимизацию использования памяти
• Мониторить рост активных соединений
```

### **API ответ:**
```json
{
  "insights": [
    "AI сервис показывает стабильную работу с точностью 85%",
    "Обнаружено 9800 аномалий за период",
    "Активные соединения: 0 (требует внимания)"
  ],
  "recommendations": [
    "Проверить активные алерты (2 активных)",
    "Оптимизировать использование памяти (268MB)",
    "Мониторить рост активных соединений"
  ],
  "timestamp": "2025-09-17T19:30:00Z"
}
```

---

## 🎯 **ЗАКЛЮЧЕНИЕ**

Эта демо-реализация показывает, как можно быстро интегрировать AI-аналитик в существующую систему CloudBridge Relay. 

**Ключевые преимущества:**
- ✅ **Быстрая интеграция** - готовый код для копирования
- ✅ **Автоматические сводки** - каждые 15 минут
- ✅ **Интеллектуальный анализ** - с помощью GPT-4
- ✅ **Множественные каналы** - Slack, Email, Webhooks
- ✅ **REST API** - для интеграции с внешними системами

**Следующие шаги:**
1. Настроить OpenAI API ключ
2. Добавить Slack webhook
3. Запустить сервис
4. Настроить планировщик
5. Интегрировать с существующими системами мониторинга

**Это первый шаг к созданию полностью автономной системы управления!** 🚀
