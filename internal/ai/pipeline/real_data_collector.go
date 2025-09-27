package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// RealDataCollector собирает реальные данные из различных источников
type RealDataCollector struct {
	logger         *zap.Logger
	httpClient     *http.Client
	prometheusURL  string
	cloudbridgeURL string
	cadenceURL     string
	lastDataPoints []DataPoint
	lastUpdate     time.Time
	updateInterval time.Duration
}

// NewRealDataCollector создает новый коллектор реальных данных
func NewRealDataCollector(logger *zap.Logger, prometheusURL, cloudbridgeURL, cadenceURL string) *RealDataCollector {
	return &RealDataCollector{
		logger:         logger,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		prometheusURL:  prometheusURL,
		cloudbridgeURL: cloudbridgeURL,
		cadenceURL:     cadenceURL,
		updateInterval: 30 * time.Second,
	}
}

// Collect собирает реальные данные из всех источников
func (rdc *RealDataCollector) Collect() ([]DataPoint, error) {
	var allDataPoints []DataPoint

	// Собираем данные из Prometheus
	prometheusData, err := rdc.collectFromPrometheus()
	if err != nil {
		rdc.logger.Warn("Failed to collect from Prometheus", zap.Error(err))
	} else {
		allDataPoints = append(allDataPoints, prometheusData...)
	}

	// Собираем данные из CloudBridge Relay
	relayData, err := rdc.collectFromCloudBridgeRelay()
	if err != nil {
		rdc.logger.Warn("Failed to collect from CloudBridge Relay", zap.Error(err))
	} else {
		allDataPoints = append(allDataPoints, relayData...)
	}

	// Собираем данные из Cadence
	cadenceData, err := rdc.collectFromCadence()
	if err != nil {
		rdc.logger.Warn("Failed to collect from Cadence", zap.Error(err))
	} else {
		allDataPoints = append(allDataPoints, cadenceData...)
	}

	// Собираем системные метрики
	systemData, err := rdc.collectSystemMetrics()
	if err != nil {
		rdc.logger.Warn("Failed to collect system metrics", zap.Error(err))
	} else {
		allDataPoints = append(allDataPoints, systemData...)
	}

	rdc.lastDataPoints = allDataPoints
	rdc.lastUpdate = time.Now()

	rdc.logger.Info("Collected real data points",
		zap.Int("total_points", len(allDataPoints)),
		zap.Int("prometheus_points", len(prometheusData)),
		zap.Int("relay_points", len(relayData)),
		zap.Int("cadence_points", len(cadenceData)),
		zap.Int("system_points", len(systemData)))

	return allDataPoints, nil
}

// collectFromPrometheus собирает данные из Prometheus
func (rdc *RealDataCollector) collectFromPrometheus() ([]DataPoint, error) {
	var dataPoints []DataPoint

	// Список метрик для сбора
	metrics := []string{
		"relay_active_connections",
		"go_goroutines",
		"go_memstats_alloc_bytes",
		"go_memstats_gc_duration_seconds",
		"process_cpu_seconds_total",
		"process_resident_memory_bytes",
	}

	for _, metric := range metrics {
		value, err := rdc.queryPrometheusMetric(metric)
		if err != nil {
			rdc.logger.Debug("Failed to query metric", zap.String("metric", metric), zap.Error(err))
			continue
		}

		dataPoints = append(dataPoints, DataPoint{
			Timestamp: time.Now(),
			Source:    "prometheus",
			Type:      "metric",
			Data: map[string]interface{}{
				"metric_name": metric,
				"value":       value,
				"source":      "prometheus",
			},
		})
	}

	return dataPoints, nil
}

// collectFromCloudBridgeRelay собирает данные из CloudBridge Relay
func (rdc *RealDataCollector) collectFromCloudBridgeRelay() ([]DataPoint, error) {
	var dataPoints []DataPoint

	// Проверяем здоровье сервиса
	healthData, err := rdc.getRelayHealth()
	if err != nil {
		return dataPoints, err
	}

	dataPoints = append(dataPoints, DataPoint{
		Timestamp: time.Now(),
		Source:    "cloudbridge-relay",
		Type:      "health",
		Data:      healthData,
	})

	// Получаем статистику подключений
	statsData, err := rdc.getRelayStats()
	if err != nil {
		rdc.logger.Debug("Failed to get relay stats", zap.Error(err))
	} else {
		dataPoints = append(dataPoints, DataPoint{
			Timestamp: time.Now(),
			Source:    "cloudbridge-relay",
			Type:      "stats",
			Data:      statsData,
		})
	}

	return dataPoints, nil
}

// collectFromCadence собирает данные из Cadence
func (rdc *RealDataCollector) collectFromCadence() ([]DataPoint, error) {
	var dataPoints []DataPoint

	// Проверяем здоровье Cadence
	healthData, err := rdc.getCadenceHealth()
	if err != nil {
		return dataPoints, err
	}

	dataPoints = append(dataPoints, DataPoint{
		Timestamp: time.Now(),
		Source:    "cadence",
		Type:      "health",
		Data:      healthData,
	})

	// Получаем статистику workflow
	workflowData, err := rdc.getCadenceWorkflowStats()
	if err != nil {
		rdc.logger.Debug("Failed to get Cadence workflow stats", zap.Error(err))
	} else {
		dataPoints = append(dataPoints, DataPoint{
			Timestamp: time.Now(),
			Source:    "cadence",
			Type:      "workflow_stats",
			Data:      workflowData,
		})
	}

	return dataPoints, nil
}

// collectSystemMetrics собирает системные метрики
func (rdc *RealDataCollector) collectSystemMetrics() ([]DataPoint, error) {
	var dataPoints []DataPoint

	// Получаем системные метрики через gopsutil
	systemData := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"source":    "system",
	}

	// CPU usage
	if cpuPercent, err := getCPUPercent(); err == nil {
		systemData["cpu_percent"] = cpuPercent
	}

	// Memory usage
	if memInfo, err := getMemoryInfo(); err == nil {
		systemData["memory_used_percent"] = memInfo["used_percent"]
		systemData["memory_available"] = memInfo["available"]
	}

	// Network stats
	if netStats, err := getNetworkStats(); err == nil {
		systemData["network_bytes_sent"] = netStats["bytes_sent"]
		systemData["network_bytes_recv"] = netStats["bytes_recv"]
	}

	dataPoints = append(dataPoints, DataPoint{
		Timestamp: time.Now(),
		Source:    "system",
		Type:      "system_metrics",
		Data:      systemData,
	})

	return dataPoints, nil
}

// queryPrometheusMetric запрашивает метрику из Prometheus
func (rdc *RealDataCollector) queryPrometheusMetric(metric string) (float64, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", rdc.prometheusURL, metric)

	resp, err := rdc.httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("prometheus query failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if result.Status != "success" || len(result.Data.Result) == 0 {
		return 0, fmt.Errorf("no data returned for metric: %s", metric)
	}

	// Извлекаем значение
	valueStr, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("invalid value format for metric: %s", metric)
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse value for metric %s: %w", metric, err)
	}

	return value, nil
}

// getRelayHealth получает информацию о здоровье CloudBridge Relay
func (rdc *RealDataCollector) getRelayHealth() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/health", rdc.cloudbridgeURL)

	resp, err := rdc.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	healthData := map[string]interface{}{
		"status_code": resp.StatusCode,
		"healthy":     resp.StatusCode == http.StatusOK,
		"timestamp":   time.Now().Unix(),
	}

	if resp.StatusCode == http.StatusOK {
		var healthResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err == nil {
			healthData["details"] = healthResponse
		}
	}

	return healthData, nil
}

// getRelayStats получает статистику CloudBridge Relay
func (rdc *RealDataCollector) getRelayStats() (map[string]interface{}, error) {
	// Используем /health endpoint вместо несуществующего /api/v1/stats
	url := fmt.Sprintf("%s/health", rdc.cloudbridgeURL)

	resp, err := rdc.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	stats["timestamp"] = time.Now().Unix()
	return stats, nil
}

// getCadenceHealth получает информацию о здоровье Cadence
func (rdc *RealDataCollector) getCadenceHealth() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/health", rdc.cadenceURL)

	resp, err := rdc.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	healthData := map[string]interface{}{
		"status_code": resp.StatusCode,
		"healthy":     resp.StatusCode == http.StatusOK,
		"timestamp":   time.Now().Unix(),
	}

	if resp.StatusCode == http.StatusOK {
		var healthResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err == nil {
			healthData["details"] = healthResponse
		}
	}

	return healthData, nil
}

// getCadenceWorkflowStats получает статистику workflow Cadence
func (rdc *RealDataCollector) getCadenceWorkflowStats() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/domains", rdc.cadenceURL)

	resp, err := rdc.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var domains []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&domains); err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"domains_count": len(domains),
		"domains":       domains,
		"timestamp":     time.Now().Unix(),
	}

	return stats, nil
}

// GetLastDataPoints возвращает последние собранные данные
func (rdc *RealDataCollector) GetLastDataPoints() []DataPoint {
	return rdc.lastDataPoints
}

// GetLastUpdate возвращает время последнего обновления
func (rdc *RealDataCollector) GetLastUpdate() time.Time {
	return rdc.lastUpdate
}

// GetUpdateInterval возвращает интервал обновления
func (rdc *RealDataCollector) GetUpdateInterval() time.Duration {
	return rdc.updateInterval
}

// StartPeriodicCollection запускает периодический сбор данных
func (rdc *RealDataCollector) StartPeriodicCollection(ctx context.Context) {
	ticker := time.NewTicker(rdc.updateInterval)
	defer ticker.Stop()

	rdc.logger.Info("Starting periodic data collection", zap.Duration("interval", rdc.updateInterval))

	for {
		select {
		case <-ctx.Done():
			rdc.logger.Info("Stopping periodic data collection")
			return
		case <-ticker.C:
			if _, err := rdc.Collect(); err != nil {
				rdc.logger.Error("Failed to collect data", zap.Error(err))
			}
		}
	}
}
