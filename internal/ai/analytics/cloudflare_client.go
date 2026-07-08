package analytics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CloudflareAIClient клиент для работы с Cloudflare AI
type CloudflareAIClient struct {
	accountID string
	apiToken  string
	model     string
	maxTokens int
	client    *http.Client
	baseURL   string
}

// CloudflareAIRequest запрос к Cloudflare AI
type CloudflareAIRequest struct {
	Messages    []CloudflareAIMessage `json:"messages"`
	Model       string                `json:"model"`
	MaxTokens   int                   `json:"max_tokens,omitempty"`
	Temperature float64               `json:"temperature,omitempty"`
}

// CloudflareAIMessage сообщение для Cloudflare AI
type CloudflareAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CloudflareAIResponse ответ от Cloudflare AI
type CloudflareAIResponse struct {
	Result struct {
		Response string `json:"response"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

// NewCloudflareAIClient создает новый клиент Cloudflare AI
func NewCloudflareAIClient(accountID, apiToken, model string, maxTokens int) *CloudflareAIClient {
	return &CloudflareAIClient{
		accountID: accountID,
		apiToken:  apiToken,
		model:     model,
		maxTokens: maxTokens,
		client: &http.Client{
			Timeout: 90 * time.Second,
		},
		baseURL: fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run", accountID),
	}
}

// GenerateAnalysis генерирует анализ с помощью Cloudflare AI
func (c *CloudflareAIClient) GenerateAnalysis(data *MetricsData) (*AnalysisResult, error) {
	prompt := c.buildPrompt(data)

	request := CloudflareAIRequest{
		Messages: []CloudflareAIMessage{
			{
				Role:    "system",
				Content: "Ты эксперт по системному администрированию, анализирующий метрики CloudBridge Relay. Предоставляй четкие, практичные инсайты на русском языке. Фокусируйся на критических проблемах и практических рекомендациях.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Model:       c.model,
		MaxTokens:   c.maxTokens,
		Temperature: 0.7,
	}

	response, err := c.makeRequest(request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate analysis: %w", err)
	}

	return c.parseAnalysis(response.Result.Response), nil
}

// GenerateReport генерирует профессиональный отчет с помощью Cloudflare AI
func (c *CloudflareAIClient) GenerateReport(metrics AIMetricsData, prompt string) (*AIReport, error) {
	// Форматируем промпт с реальными метриками
	formattedPrompt := fmt.Sprintf(prompt,
		metrics.AnomaliesDetected,
		metrics.PredictionsMade,
		metrics.RequestsProcessed,
		metrics.RoutesOptimized,
		metrics.PredictionAccuracy,
		metrics.CPUUsage,
		metrics.MemoryUsage,
		metrics.DiskUsage,
		metrics.ActiveAlerts,
		metrics.ProcessingTime,
		metrics.ServiceUptime,
	)

	request := CloudflareAIRequest{
		Messages: []CloudflareAIMessage{
			{
				Role:    "system",
				Content: "Ты - эксперт по анализу систем мониторинга CloudBridge AI. Создавай профессиональные отчеты на русском языке с детальным анализом, конкретными рекомендациями и прогнозами.",
			},
			{
				Role:    "user",
				Content: formattedPrompt,
			},
		},
		Model:       c.model,
		MaxTokens:   c.maxTokens,
		Temperature: 0.7,
	}

	response, err := c.makeRequest(request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	// Парсим ответ и создаем AIReport
	return c.parseAIReport(response.Result.Response, metrics), nil
}

// buildPrompt строит промпт для анализа
func (c *CloudflareAIClient) buildPrompt(data *MetricsData) string {
	var prompt strings.Builder

	// System-style guardrails
	prompt.WriteString("Ты эксперт по SRE/NetOps и MLOps. Анализируй метрики CloudBridge Relay, AI-сервиса и Cadence. ")
	prompt.WriteString("Пиши по-русски, чётко и кратко, без выдумок. При нехватке данных — укажи это явно. ")
	prompt.WriteString("Сначала критические риски и их причины, затем рекомендации с приоритетами (P0..P2). Отвечай строго в указанном формате секций.\n\n")

	prompt.WriteString("Контекст системы:\n- Relay: транспорт/маршрутизация трафика\n- AI: предсказания/анализ, RT метрики и история аномалий\n- Cadence: фоновые воркфлоу/ретраи/ошибки задач\n\n")
	prompt.WriteString("Данные метрик (агрегировано):\n")

	// AI метрики
	if aiMetrics, exists := data.Metrics["ai"]; exists {
		if aiMap, ok := aiMetrics.(map[string]interface{}); ok {
			prompt.WriteString(" AI Service Metrics:\n")
			if anomalies, ok := aiMap["anomalies_detected"].(float64); ok {
				prompt.WriteString(fmt.Sprintf("- Обнаружено аномалий: %.0f\n", anomalies))
			}
			if predictions, ok := aiMap["predictions_made"].(float64); ok {
				prompt.WriteString(fmt.Sprintf("- Сделано предсказаний: %.0f\n", predictions))
			}
			if cpuUsage, ok := aiMap["cpu_usage"].(float64); ok {
				prompt.WriteString(fmt.Sprintf("- Использование CPU: %.1f%%\n", cpuUsage))
			}
			if memoryUsage, ok := aiMap["memory_usage"].(float64); ok {
				prompt.WriteString(fmt.Sprintf("- Использование памяти: %.1f MB\n", memoryUsage/1024/1024))
			}
			if activeAlerts, ok := aiMap["active_alerts"].(float64); ok {
				prompt.WriteString(fmt.Sprintf("- Активные алерты: %.0f\n", activeAlerts))
			}
			if accuracy, ok := aiMap["prediction_accuracy"].(float64); ok {
				prompt.WriteString(fmt.Sprintf("- Точность предсказаний: %.1f%%\n", accuracy*100))
			}
		}
	}

	// Relay метрики
	if relayMetrics, exists := data.Metrics["relay"]; exists {
		if relayMap, ok := relayMetrics.(map[string]interface{}); ok {
			prompt.WriteString("\nRelay Service Metrics:\n")
			// Cadence метрики
			if cadenceMetrics, exists := data.Metrics["cadence"]; exists {
				if cadenceMap, ok := cadenceMetrics.(map[string]interface{}); ok {
					prompt.WriteString("\n🕒 Cadence Workflow Metrics:\n")
					if q, ok := cadenceMap["queue_depth"].(float64); ok {
						prompt.WriteString(fmt.Sprintf("- Глубина очередей: %.0f\n", q))
					}
					if r, ok := cadenceMap["retries"].(float64); ok {
						prompt.WriteString(fmt.Sprintf("- Ретраи задач: %.0f\n", r))
					}
					if f, ok := cadenceMap["failures"].(float64); ok {
						prompt.WriteString(fmt.Sprintf("- Ошибки задач: %.0f\n", f))
					}
				}
			}

			if activeConnections, ok := relayMap["active_connections"].(float64); ok {
				prompt.WriteString(fmt.Sprintf("- Активные соединения: %.0f\n", activeConnections))
			}
			if totalConnections, ok := relayMap["connections_total"].(float64); ok {
				prompt.WriteString(fmt.Sprintf("- Всего соединений: %.0f\n", totalConnections))
			}
			if uptime, ok := relayMap["uptime_seconds"].(float64); ok {
				prompt.WriteString(fmt.Sprintf("- Время работы: %.0f секунд\n", uptime))
			}
		}
	}

	// Аномалии
	if len(data.Anomalies) > 0 {
		prompt.WriteString("\nОбнаруженные аномалии:\n")
		for _, anomaly := range data.Anomalies {
			prompt.WriteString(fmt.Sprintf("- %s (%s): %s\n",
				anomaly.Type, anomaly.Severity, anomaly.Description))
		}
	} else {
		prompt.WriteString("\nОбнаруженные аномалии: Аномалий не обнаружено\n")
	}

	// Статус здоровья
	if data.Health != nil {
		prompt.WriteString(fmt.Sprintf("\nСтатус здоровья: %s (оценка: %.1f/100)\n",
			data.Health.Overall, data.Health.Score))
	}

	prompt.WriteString(`
ВАЖНО: Отвечай КРАТКО на русском языке. Максимум 1-2 предложения на пункт.

Формат ответа:
**Резюме:** [1-2 предложения о текущем состоянии системы]
**Инсайты:**
- краткий инсайт 1 (до 50 символов)
- краткий инсайт 2 (до 50 символов)
**Критические проблемы:**
- проблема и причина (до 80 символов)
- проблема и причина (до 80 символов)
**Рекомендации:**
- [P0] действие (до 60 символов)
- [P1] действие (до 60 символов)
**Производительность:** [краткая оценка до 100 символов]

Пиши максимально сжато, без лишних слов!
`)

	return prompt.String()
}

// makeRequest выполняет запрос к Cloudflare AI API
func (c *CloudflareAIClient) makeRequest(request CloudflareAIRequest) (*CloudflareAIResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Используем название модели без кодирования (Cloudflare AI принимает @ в URL)
	req, err := http.NewRequest("POST", c.baseURL+"/"+c.model, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response CloudflareAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		if len(response.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare ai error: %s", response.Errors[0].Message)
		}
		return nil, fmt.Errorf("cloudflare ai request failed")
	}

	return &response, nil
}

// parseAnalysis парсит ответ от Cloudflare AI
func (c *CloudflareAIClient) parseAnalysis(content string) *AnalysisResult {
	// Базовая реализация парсинга анализа
	return &AnalysisResult{
		Summary:         content,
		Insights:        []string{"Анализ выполнен Cloudflare AI"},
		Recommendations: []string{"Продолжить мониторинг системы"},
		CriticalIssues:  []string{},
		Performance:     "Анализ завершен",
	}
}

// parseAIReport парсит ответ от Cloudflare AI в структуру AIReport
func (c *CloudflareAIClient) parseAIReport(content string, metrics AIMetricsData) *AIReport {
	// Определяем статус системы на основе метрик
	healthScore := c.calculateHealthScore(metrics)
	systemStatus := c.determineSystemStatus(healthScore)

	// Парсим анализ из ответа AI
	analysis := c.parseAnalysisFromContent(content)

	return &AIReport{
		Timestamp:    time.Now(),
		SystemStatus: systemStatus,
		HealthScore:  healthScore,
		Analysis: struct {
			OverallAssessment string   `json:"overall_assessment"`
			Performance      string   `json:"performance"`
			Issues           []string `json:"issues"`
			Recommendations  []string `json:"recommendations"`
			Forecast         string   `json:"forecast"`
		}{
			OverallAssessment: analysis.Summary,
			Performance:      analysis.Performance,
			Issues:          analysis.CriticalIssues,
			Recommendations: analysis.Recommendations,
			Forecast:        "Система будет стабильно работать в ближайшие 24 часа",
		},
		Metrics: struct {
			AnomaliesDetected    int     `json:"anomalies_detected"`
			PredictionsMade      int     `json:"predictions_made"`
			RequestsProcessed    int     `json:"requests_processed"`
			RoutesOptimized      int     `json:"routes_optimized"`
			PredictionAccuracy   float64 `json:"prediction_accuracy"`
			CPUUsage            float64 `json:"cpu_usage"`
			MemoryUsage         float64 `json:"memory_usage"`
			DiskUsage           float64 `json:"disk_usage"`
			ActiveAlerts        int     `json:"active_alerts"`
			ProcessingTime      float64 `json:"processing_time"`
			ServiceUptime       float64 `json:"service_uptime"`
		}{
			AnomaliesDetected:  metrics.AnomaliesDetected,
			PredictionsMade:    metrics.PredictionsMade,
			RequestsProcessed:  metrics.RequestsProcessed,
			RoutesOptimized:    metrics.RoutesOptimized,
			PredictionAccuracy: metrics.PredictionAccuracy,
			CPUUsage:          metrics.CPUUsage,
			MemoryUsage:       metrics.MemoryUsage,
			DiskUsage:         metrics.DiskUsage,
			ActiveAlerts:      metrics.ActiveAlerts,
			ProcessingTime:    metrics.ProcessingTime,
			ServiceUptime:     metrics.ServiceUptime,
		},
	}
}

// calculateHealthScore вычисляет оценку здоровья системы (0-100)
func (c *CloudflareAIClient) calculateHealthScore(metrics AIMetricsData) int {
	score := 100.0

	// Штрафы за проблемы
	if metrics.CPUUsage > 80 {
		score -= 20
	} else if metrics.CPUUsage > 60 {
		score -= 10
	}

	if metrics.MemoryUsage > 80 {
		score -= 15
	} else if metrics.MemoryUsage > 60 {
		score -= 8
	}

	if metrics.DiskUsage > 90 {
		score -= 25
	} else if metrics.DiskUsage > 80 {
		score -= 15
	}

	if metrics.ActiveAlerts > 0 {
		score -= float64(metrics.ActiveAlerts * 10)
	}

	if metrics.PredictionAccuracy < 0.7 {
		score -= 20
	} else if metrics.PredictionAccuracy < 0.8 {
		score -= 10
	}

	// Бонусы за хорошие показатели
	if metrics.PredictionAccuracy > 0.9 {
		score += 5
	}

	if metrics.ServiceUptime > 99.9 {
		score += 5
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return int(score)
}

// determineSystemStatus определяет статус системы на основе оценки здоровья
func (c *CloudflareAIClient) determineSystemStatus(healthScore int) string {
	switch {
	case healthScore >= 90:
		return "excellent"
	case healthScore >= 75:
		return "good"
	case healthScore >= 50:
		return "attention"
	default:
		return "critical"
	}
}

// parseAnalysisFromContent парсит анализ из текста ответа AI
func (c *CloudflareAIClient) parseAnalysisFromContent(content string) *AnalysisResult {
	// Простой парсинг - в реальной реализации можно использовать более сложную логику
	lines := strings.Split(content, "\n")
	
	var issues []string
	var recommendations []string
	
	// Ищем проблемы и рекомендации в тексте
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "проблема") || strings.Contains(strings.ToLower(line), "ошибка") {
			issues = append(issues, line)
		}
		if strings.Contains(strings.ToLower(line), "рекоменд") || strings.Contains(strings.ToLower(line), "совет") {
			recommendations = append(recommendations, line)
		}
	}

	return &AnalysisResult{
		Summary:         content,
		Insights:        []string{"Анализ выполнен Cloudflare AI"},
		Recommendations: recommendations,
		CriticalIssues:  issues,
		Performance:     "Анализ завершен",
	}
}
