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
			Timeout: 30 * time.Second,
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

// buildPrompt строит промпт для анализа
func (c *CloudflareAIClient) buildPrompt(data *MetricsData) string {
	var prompt strings.Builder

	prompt.WriteString("Проанализируй следующие метрики CloudBridge Relay и предоставь инсайты:\n\n")

	// AI метрики
	if aiMetrics, exists := data.Metrics["ai"]; exists {
		if aiMap, ok := aiMetrics.(map[string]interface{}); ok {
			prompt.WriteString("🤖 AI Service Metrics:\n")
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
			prompt.WriteString("\n🌐 Relay Service Metrics:\n")
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
		prompt.WriteString("\n🚨 Обнаруженные аномалии:\n")
		for _, anomaly := range data.Anomalies {
			prompt.WriteString(fmt.Sprintf("- %s (%s): %s\n",
				anomaly.Type, anomaly.Severity, anomaly.Description))
		}
	} else {
		prompt.WriteString("\n🚨 Обнаруженные аномалии: Аномалий не обнаружено\n")
	}

	// Статус здоровья
	if data.Health != nil {
		prompt.WriteString(fmt.Sprintf("\n📊 Статус здоровья: %s (оценка: %.1f/100)\n",
			data.Health.Overall, data.Health.Score))
	}

	prompt.WriteString(`
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
`)

	return prompt.String()
}

// makeRequest выполняет запрос к Cloudflare AI API
func (c *CloudflareAIClient) makeRequest(request CloudflareAIRequest) (*CloudflareAIResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

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

	// Если не удалось распарсить структурированный ответ, используем весь контент как резюме
	if result.Summary == "" {
		result.Summary = content
	}

	return result
}
