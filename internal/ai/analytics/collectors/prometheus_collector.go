package collectors

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/twogc/cloudbridge-relay-installer/internal/ai/analytics"
	"go.uber.org/zap"
)

// PrometheusCollector сборщик метрик из Prometheus
type PrometheusCollector struct {
	client *http.Client
	url    string
	logger *zap.Logger
}

// NewPrometheusCollector создает новый сборщик Prometheus
func NewPrometheusCollector(url string) *PrometheusCollector {
	return &PrometheusCollector{
		client: &http.Client{Timeout: 10 * time.Second},
		url:    url,
		logger: zap.NewNop(),
	}
}

// GetName возвращает имя сборщика
func (p *PrometheusCollector) GetName() string {
	return "prometheus"
}

// CollectMetrics собирает метрики из Prometheus
func (p *PrometheusCollector) CollectMetrics(ctx context.Context) (*analytics.MetricsData, error) {
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
	anomalies := make([]analytics.Anomaly, 0)

	// AI метрики
	aiMetrics := map[string]float64{
		"anomalies_detected":  p.parseMetric(resp.Body, "ai_anomalies_detected_total"),
		"predictions_made":    p.parseMetric(resp.Body, "ai_predictions_made_total"),
		"cpu_usage":           p.parseMetric(resp.Body, "ai_cpu_usage_percent"),
		"memory_usage":        p.parseMetric(resp.Body, "ai_memory_usage_bytes"),
		"active_alerts":       p.parseMetric(resp.Body, "ai_active_alerts"),
		"prediction_accuracy": p.parseMetric(resp.Body, "ai_prediction_accuracy"),
	}

	// Relay метрики
	relayMetrics := map[string]float64{
		"active_connections": p.parseMetric(resp.Body, "relay_active_connections"),
		"connections_total":  p.parseMetric(resp.Body, "relay_connections_total"),
		"uptime_seconds":     p.parseMetric(resp.Body, "relay_uptime_seconds"),
		"bytes_transferred":  p.parseMetric(resp.Body, "relay_bytes_transferred_total"),
		"error_rate":         p.parseMetric(resp.Body, "relay_error_rate"),
		"average_latency":    p.parseMetric(resp.Body, "relay_average_latency_seconds"),
	}

	metrics["ai"] = aiMetrics
	metrics["relay"] = relayMetrics

	// Детекция аномалий
	anomalies = append(anomalies, p.detectAnomalies(aiMetrics)...)
	anomalies = append(anomalies, p.detectAnomalies(relayMetrics)...)

	return &analytics.MetricsData{
		Timestamp: time.Now(),
		Source:    "prometheus",
		Metrics:   metrics,
		Anomalies: anomalies,
		Health:    p.assessHealth(aiMetrics, relayMetrics),
	}, nil
}

// parseMetric парсит конкретную метрику из ответа Prometheus
func (p *PrometheusCollector) parseMetric(body io.Reader, metricName string) float64 {
	// Создаем новый reader для каждого поиска
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return 0.0
	}

	scanner := bufio.NewScanner(strings.NewReader(string(bodyBytes)))
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

// detectAnomalies обнаруживает аномалии в метриках
func (p *PrometheusCollector) detectAnomalies(metrics map[string]float64) []analytics.Anomaly {
	anomalies := make([]analytics.Anomaly, 0)

	// Проверка CPU usage
	if cpuUsage, exists := metrics["cpu_usage"]; exists {
		if cpuUsage > 80.0 {
			anomalies = append(anomalies, analytics.Anomaly{
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
			anomalies = append(anomalies, analytics.Anomaly{
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
			anomalies = append(anomalies, analytics.Anomaly{
				Type:        "active_alerts",
				Severity:    "critical",
				Description: fmt.Sprintf("Active alerts: %.0f", activeAlerts),
				Value:       activeAlerts,
				Threshold:   0,
				Timestamp:   time.Now(),
			})
		}
	}

	// Проверка prediction accuracy
	if accuracy, exists := metrics["prediction_accuracy"]; exists {
		if accuracy < 0.7 { // 70%
			anomalies = append(anomalies, analytics.Anomaly{
				Type:        "low_prediction_accuracy",
				Severity:    "warning",
				Description: fmt.Sprintf("Low prediction accuracy: %.1f%%", accuracy*100),
				Value:       accuracy,
				Threshold:   0.7,
				Timestamp:   time.Now(),
			})
		}
	}

	// Проверка active connections
	if activeConnections, exists := metrics["active_connections"]; exists {
		if activeConnections == 0 {
			anomalies = append(anomalies, analytics.Anomaly{
				Type:        "no_active_connections",
				Severity:    "critical",
				Description: "No active connections detected",
				Value:       activeConnections,
				Threshold:   1,
				Timestamp:   time.Now(),
			})
		}
	}

	// Проверка error rate
	if errorRate, exists := metrics["error_rate"]; exists {
		if errorRate > 0.05 { // 5%
			anomalies = append(anomalies, analytics.Anomaly{
				Type:        "high_error_rate",
				Severity:    "warning",
				Description: fmt.Sprintf("High error rate: %.2f%%", errorRate*100),
				Value:       errorRate,
				Threshold:   0.05,
				Timestamp:   time.Now(),
			})
		}
	}

	// Проверка latency
	if latency, exists := metrics["average_latency"]; exists {
		if latency > 1.0 { // 1 second
			anomalies = append(anomalies, analytics.Anomaly{
				Type:        "high_latency",
				Severity:    "warning",
				Description: fmt.Sprintf("High latency: %.2f seconds", latency),
				Value:       latency,
				Threshold:   1.0,
				Timestamp:   time.Now(),
			})
		}
	}

	return anomalies
}

// assessHealth оценивает здоровье системы
func (p *PrometheusCollector) assessHealth(aiMetrics, relayMetrics map[string]float64) *analytics.HealthStatus {
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

	// Проверка точности предсказаний
	if accuracy, exists := aiMetrics["prediction_accuracy"]; exists {
		if accuracy < 0.5 {
			components["ai_accuracy"] = "critical"
			score -= 20
		} else if accuracy < 0.7 {
			components["ai_accuracy"] = "warning"
			score -= 10
		} else {
			components["ai_accuracy"] = "healthy"
		}
	}

	// Проверка активных алертов
	if activeAlerts, exists := aiMetrics["active_alerts"]; exists {
		if activeAlerts > 5 {
			components["alerts"] = "critical"
			score -= 25
		} else if activeAlerts > 0 {
			components["alerts"] = "warning"
			score -= 10
		} else {
			components["alerts"] = "healthy"
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







