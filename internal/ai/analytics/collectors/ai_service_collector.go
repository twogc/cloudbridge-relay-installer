package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/analytics"
	"go.uber.org/zap"
)

// AIServiceCollector сборщик метрик из AI Service
type AIServiceCollector struct {
	client *http.Client
	url    string
	logger *zap.Logger
}

// NewAIServiceCollector создает новый сборщик AI Service
func NewAIServiceCollector(url string) *AIServiceCollector {
	return &AIServiceCollector{
		client: &http.Client{Timeout: 10 * time.Second},
		url:    url,
		logger: zap.NewNop(),
	}
}

// GetName возвращает имя сборщика
func (a *AIServiceCollector) GetName() string {
	return "ai_service"
}

// CollectMetrics собирает метрики из AI Service
func (a *AIServiceCollector) CollectMetrics(ctx context.Context) (*analytics.MetricsData, error) {
	// Собираем метрики из различных endpoints AI Service
	metrics := make(map[string]interface{})
	anomalies := make([]analytics.Anomaly, 0)

	// Получаем статус AI Service
	status, err := a.getAIStatus()
	if err != nil {
		a.logger.Warn("Failed to get AI status", zap.Error(err))
	} else {
		metrics["status"] = status
	}

	// Получаем метрики AI Service
	aiMetrics, err := a.getAIMetrics()
	if err != nil {
		a.logger.Warn("Failed to get AI metrics", zap.Error(err))
	} else {
		metrics["ai_metrics"] = aiMetrics
	}

	// Получаем историю аномалий
	anomalyHistory, err := a.getAnomalyHistory()
	if err != nil {
		a.logger.Warn("Failed to get anomaly history", zap.Error(err))
	} else {
		metrics["anomaly_history"] = anomalyHistory
	}

	// Получаем историю предсказаний
	predictionHistory, err := a.getPredictionHistory()
	if err != nil {
		a.logger.Warn("Failed to get prediction history", zap.Error(err))
	} else {
		metrics["prediction_history"] = predictionHistory
	}

	// Анализируем метрики на предмет аномалий
	anomalies = append(anomalies, a.analyzeAIMetrics(aiMetrics)...)

	return &analytics.MetricsData{
		Timestamp: time.Now(),
		Source:    "ai_service",
		Metrics:   metrics,
		Anomalies: anomalies,
		Health:    a.assessAIHealth(status, aiMetrics),
	}, nil
}

// getAIStatus получает статус AI Service
func (a *AIServiceCollector) getAIStatus() (map[string]interface{}, error) {
	resp, err := a.client.Get(a.url + "/api/v1/ai/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return status, nil
}

// getAIMetrics получает метрики AI Service
func (a *AIServiceCollector) getAIMetrics() (map[string]interface{}, error) {
	resp, err := a.client.Get(a.url + "/api/v1/ai/metrics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var metrics map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

// getAnomalyHistory получает историю аномалий
func (a *AIServiceCollector) getAnomalyHistory() (map[string]interface{}, error) {
	resp, err := a.client.Get(a.url + "/api/v1/ai/history/anomalies")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var history map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, err
	}

	return history, nil
}

// getPredictionHistory получает историю предсказаний
func (a *AIServiceCollector) getPredictionHistory() (map[string]interface{}, error) {
	resp, err := a.client.Get(a.url + "/api/v1/ai/history/predictions")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var history map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, err
	}

	return history, nil
}

// analyzeAIMetrics анализирует метрики AI Service на предмет аномалий
func (a *AIServiceCollector) analyzeAIMetrics(metrics map[string]interface{}) []analytics.Anomaly {
	anomalies := make([]analytics.Anomaly, 0)

	// Проверяем компоненты AI Service
	if components, exists := metrics["components"]; exists {
		if componentsMap, ok := components.(map[string]interface{}); ok {
			for component, status := range componentsMap {
				if status == false {
					anomalies = append(anomalies, analytics.Anomaly{
						Type:        "component_down",
						Severity:    "critical",
						Description: fmt.Sprintf("AI component %s is down", component),
						Value:       0,
						Threshold:   1,
						Timestamp:   time.Now(),
					})
				}
			}
		}
	}

	// Проверяем метрики производительности
	if aiMetrics, exists := metrics["metrics"]; exists {
		if metricsMap, ok := aiMetrics.(map[string]interface{}); ok {
			// Проверяем количество обработанных запросов
			if requestsProcessed, ok := metricsMap["requests_processed"].(float64); ok {
				if requestsProcessed == 0 {
					anomalies = append(anomalies, analytics.Anomaly{
						Type:        "no_requests_processed",
						Severity:    "warning",
						Description: "No requests processed by AI service",
						Value:       requestsProcessed,
						Threshold:   1,
						Timestamp:   time.Now(),
					})
				}
			}

			// Проверяем количество обнаруженных аномалий
			if anomaliesDetected, ok := metricsMap["anomalies_detected"].(float64); ok {
				if anomaliesDetected > 100 {
					anomalies = append(anomalies, analytics.Anomaly{
						Type:        "high_anomaly_count",
						Severity:    "warning",
						Description: fmt.Sprintf("High number of anomalies detected: %.0f", anomaliesDetected),
						Value:       anomaliesDetected,
						Threshold:   100,
						Timestamp:   time.Now(),
					})
				}
			}

			// Проверяем количество сделанных предсказаний
			if predictionsMade, ok := metricsMap["predictions_made"].(float64); ok {
				if predictionsMade == 0 {
					anomalies = append(anomalies, analytics.Anomaly{
						Type:        "no_predictions_made",
						Severity:    "warning",
						Description: "No predictions made by AI service",
						Value:       predictionsMade,
						Threshold:   1,
						Timestamp:   time.Now(),
					})
				}
			}
		}
	}

	return anomalies
}

// assessAIHealth оценивает здоровье AI Service
func (a *AIServiceCollector) assessAIHealth(status, metrics map[string]interface{}) *analytics.HealthStatus {
	components := make(map[string]string)
	score := 100.0

	// Проверяем общий статус
	if serviceStatus, exists := status["status"]; exists {
		if serviceStatus != "running" {
			components["ai_service"] = "critical"
			score -= 50
		} else {
			components["ai_service"] = "healthy"
		}
	}

	// Проверяем компоненты
	if componentsData, exists := status["components"]; exists {
		if componentsMap, ok := componentsData.(map[string]interface{}); ok {
			for component, isActive := range componentsMap {
				if active, ok := isActive.(bool); ok {
					if !active {
						components[component] = "critical"
						score -= 20
					} else {
						components[component] = "healthy"
					}
				}
			}
		}
	}

	// Проверяем метрики производительности
	if metricsData, exists := metrics["metrics"]; exists {
		if metricsMap, ok := metricsData.(map[string]interface{}); ok {
			// Проверяем активность
			if requestsProcessed, ok := metricsMap["requests_processed"].(float64); ok {
				if requestsProcessed == 0 {
					components["activity"] = "warning"
					score -= 15
				} else {
					components["activity"] = "healthy"
				}
			}

			// Проверяем количество аномалий
			if anomaliesDetected, ok := metricsMap["anomalies_detected"].(float64); ok {
				if anomaliesDetected > 50 {
					components["anomaly_detection"] = "warning"
					score -= 10
				} else {
					components["anomaly_detection"] = "healthy"
				}
			}
		}
	}

	overall := "healthy"
	if score < 50 {
		overall = "critical"
	} else if score < 80 {
		overall = "warning"
	}

	return &analytics.HealthStatus{
		Overall:    overall,
		Components: components,
		Score:      score,
	}
}








