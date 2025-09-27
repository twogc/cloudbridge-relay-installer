package routing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// NetworkMetricsCollector собирает реальные метрики сети и загрузки узлов
type NetworkMetricsCollector struct {
	logger         *zap.Logger
	httpClient     *http.Client
	prometheusURL  string
	cloudbridgeURL string
	metrics        map[string]*NodeMetrics
	lastUpdate     time.Time
	updateInterval time.Duration
	mu             sync.RWMutex
}

// NodeMetrics содержит метрики узла
type NodeMetrics struct {
	NodeID      string    `json:"node_id"`
	Location    string    `json:"location"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	NetworkLoad float64   `json:"network_load"`
	ActiveConns int       `json:"active_connections"`
	Bandwidth   float64   `json:"bandwidth"`
	Latency     float64   `json:"latency"`
	ErrorRate   float64   `json:"error_rate"`
	LastUpdate  time.Time `json:"last_update"`
	Health      string    `json:"health"`
}

// EdgeMetrics содержит метрики соединения между узлами
type EdgeMetrics struct {
	From       string    `json:"from"`
	To         string    `json:"to"`
	Latency    float64   `json:"latency"`
	Bandwidth  float64   `json:"bandwidth"`
	PacketLoss float64   `json:"packet_loss"`
	Jitter     float64   `json:"jitter"`
	LastUpdate time.Time `json:"last_update"`
}

// NewNetworkMetricsCollector создает новый коллектор метрик сети
func NewNetworkMetricsCollector(logger *zap.Logger, prometheusURL, cloudbridgeURL string) *NetworkMetricsCollector {
	return &NetworkMetricsCollector{
		logger:         logger,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		prometheusURL:  prometheusURL,
		cloudbridgeURL: cloudbridgeURL,
		metrics:        make(map[string]*NodeMetrics),
		updateInterval: 30 * time.Second,
	}
}

// CollectMetrics собирает метрики всех узлов
func (nmc *NetworkMetricsCollector) CollectMetrics() error {
	nmc.mu.Lock()
	defer nmc.mu.Unlock()

	// Собираем метрики из Prometheus
	if err := nmc.collectPrometheusMetrics(); err != nil {
		nmc.logger.Warn("Failed to collect Prometheus metrics", zap.Error(err))
	}

	// Собираем метрики из CloudBridge Relay
	if err := nmc.collectRelayMetrics(); err != nil {
		nmc.logger.Warn("Failed to collect Relay metrics", zap.Error(err))
	}

	// Собираем системные метрики
	if err := nmc.collectSystemMetrics(); err != nil {
		nmc.logger.Warn("Failed to collect system metrics", zap.Error(err))
	}

	nmc.lastUpdate = time.Now()

	nmc.logger.Info("Network metrics collected",
		zap.Int("nodes", len(nmc.metrics)),
		zap.Time("last_update", nmc.lastUpdate))

	return nil
}

// collectPrometheusMetrics собирает метрики из Prometheus
func (nmc *NetworkMetricsCollector) collectPrometheusMetrics() error {
	// Список метрик для сбора
	metrics := map[string]string{
		"relay_active_connections":  "active_connections",
		"go_memstats_alloc_bytes":   "memory_usage",
		"process_cpu_seconds_total": "cpu_usage",
		"go_goroutines":             "goroutines",
	}

	for metric, field := range metrics {
		value, err := nmc.queryPrometheusMetric(metric)
		if err != nil {
			nmc.logger.Debug("Failed to query metric", zap.String("metric", metric), zap.Error(err))
			continue
		}

		// Обновляем метрики для всех узлов
		for _, nodeMetrics := range nmc.metrics {
			switch field {
			case "active_connections":
				nodeMetrics.ActiveConns = int(value)
			case "memory_usage":
				nodeMetrics.MemoryUsage = value / (1024 * 1024) // Convert to MB
			case "cpu_usage":
				nodeMetrics.CPUUsage = value
			case "goroutines":
				// Используем количество горутин как индикатор нагрузки
				nodeMetrics.NetworkLoad = value / 1000.0 // Normalize
			}
			nodeMetrics.LastUpdate = time.Now()
		}
	}

	return nil
}

// collectRelayMetrics собирает метрики из CloudBridge Relay
func (nmc *NetworkMetricsCollector) collectRelayMetrics() error {
	// Получаем статистику из Relay
	stats, err := nmc.getRelayStats()
	if err != nil {
		return err
	}

	// Обновляем метрики на основе данных Relay
	for _, nodeMetrics := range nmc.metrics {
		if bandwidth, ok := stats["bandwidth"].(float64); ok {
			nodeMetrics.Bandwidth = bandwidth
		}
		if latency, ok := stats["latency"].(float64); ok {
			nodeMetrics.Latency = latency
		}
		if errorRate, ok := stats["error_rate"].(float64); ok {
			nodeMetrics.ErrorRate = errorRate
		}
		nodeMetrics.LastUpdate = time.Now()
	}

	return nil
}

// collectSystemMetrics собирает системные метрики
func (nmc *NetworkMetricsCollector) collectSystemMetrics() error {
	// Получаем системные метрики
	systemData, err := nmc.getSystemMetrics()
	if err != nil {
		return err
	}

	// Обновляем метрики на основе системных данных
	for _, nodeMetrics := range nmc.metrics {
		if cpuPercent, ok := systemData["cpu_percent"].(float64); ok {
			nodeMetrics.CPUUsage = cpuPercent
		}
		if memPercent, ok := systemData["memory_used_percent"].(float64); ok {
			nodeMetrics.MemoryUsage = memPercent
		}
		nodeMetrics.LastUpdate = time.Now()
	}

	return nil
}

// queryPrometheusMetric запрашивает метрику из Prometheus
func (nmc *NetworkMetricsCollector) queryPrometheusMetric(metric string) (float64, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", nmc.prometheusURL, metric)

	resp, err := nmc.httpClient.Get(url)
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

// getRelayStats получает статистику CloudBridge Relay
func (nmc *NetworkMetricsCollector) getRelayStats() (map[string]interface{}, error) {
	// Используем /health endpoint вместо несуществующего /api/v1/stats
	url := fmt.Sprintf("%s/health", nmc.cloudbridgeURL)

	resp, err := nmc.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// getSystemMetrics получает системные метрики
func (nmc *NetworkMetricsCollector) getSystemMetrics() (map[string]interface{}, error) {
	// Здесь можно использовать gopsutil или другие библиотеки для получения системных метрик
	// Для простоты возвращаем базовые метрики
	return map[string]interface{}{
		"cpu_percent":         25.0, // В реальной реализации получать из gopsutil
		"memory_used_percent": 60.0, // В реальной реализации получать из gopsutil
		"network_bytes_sent":  1024000,
		"network_bytes_recv":  2048000,
	}, nil
}

// GetNodeMetrics возвращает метрики узла
func (nmc *NetworkMetricsCollector) GetNodeMetrics(nodeID string) (*NodeMetrics, bool) {
	nmc.mu.RLock()
	defer nmc.mu.RUnlock()

	metrics, exists := nmc.metrics[nodeID]
	return metrics, exists
}

// GetAllNodeMetrics возвращает метрики всех узлов
func (nmc *NetworkMetricsCollector) GetAllNodeMetrics() map[string]*NodeMetrics {
	nmc.mu.RLock()
	defer nmc.mu.RUnlock()

	// Создаем копию для безопасности
	result := make(map[string]*NodeMetrics)
	for nodeID, metrics := range nmc.metrics {
		result[nodeID] = metrics
	}

	return result
}

// InitializeNodes инициализирует узлы с базовыми метриками
func (nmc *NetworkMetricsCollector) InitializeNodes(nodes []*Node) {
	nmc.mu.Lock()
	defer nmc.mu.Unlock()

	for _, node := range nodes {
		nmc.metrics[node.ID] = &NodeMetrics{
			NodeID:      node.ID,
			Location:    node.Location,
			CPUUsage:    0.0,
			MemoryUsage: 0.0,
			NetworkLoad: 0.0,
			ActiveConns: 0,
			Bandwidth:   0.0,
			Latency:     0.0,
			ErrorRate:   0.0,
			LastUpdate:  time.Now(),
			Health:      "unknown",
		}
	}

	nmc.logger.Info("Initialized nodes", zap.Int("count", len(nodes)))
}

// StartPeriodicCollection запускает периодический сбор метрик
func (nmc *NetworkMetricsCollector) StartPeriodicCollection(ctx context.Context) {
	ticker := time.NewTicker(nmc.updateInterval)
	defer ticker.Stop()

	nmc.logger.Info("Starting periodic network metrics collection", zap.Duration("interval", nmc.updateInterval))

	for {
		select {
		case <-ctx.Done():
			nmc.logger.Info("Stopping periodic network metrics collection")
			return
		case <-ticker.C:
			if err := nmc.CollectMetrics(); err != nil {
				nmc.logger.Error("Failed to collect network metrics", zap.Error(err))
			}
		}
	}
}

// GetLastUpdate возвращает время последнего обновления
func (nmc *NetworkMetricsCollector) GetLastUpdate() time.Time {
	nmc.mu.RLock()
	defer nmc.mu.RUnlock()
	return nmc.lastUpdate
}

// GetUpdateInterval возвращает интервал обновления
func (nmc *NetworkMetricsCollector) GetUpdateInterval() time.Duration {
	return nmc.updateInterval
}
