package pipeline

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"go.uber.org/zap"
)

// RealPrometheusCollector собирает реальные данные из Prometheus
type RealPrometheusCollector struct {
	logger     *zap.Logger
	httpClient *http.Client
	baseURL    string
	metrics    []string
}

// NewRealPrometheusCollector создает новый Prometheus коллектор
func NewRealPrometheusCollector(logger *zap.Logger, baseURL string) *RealPrometheusCollector {
	return &RealPrometheusCollector{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
		metrics: []string{
			"relay_connections_total",
			"relay_connections_active",
			"relay_bandwidth_bytes",
			"relay_latency_seconds",
			"relay_errors_total",
			"relay_requests_total",
		},
	}
}

// Collect собирает данные из Prometheus
func (pc *RealPrometheusCollector) Collect() ([]DataPoint, error) {
	var dataPoints []DataPoint

	for _, metric := range pc.metrics {
		value, err := pc.queryPrometheus(metric)
		if err != nil {
			pc.logger.Warn("Failed to query metric",
				zap.String("metric", metric),
				zap.Error(err))
			continue
		}

		dataPoints = append(dataPoints, DataPoint{
			Timestamp: time.Now(),
			Source:    "prometheus",
			Type:      "metric",
			Data: map[string]interface{}{
				"metric_name": metric,
				"value":       value,
			},
		})
	}

	return dataPoints, nil
}

// queryPrometheus запрашивает метрику из Prometheus
func (pc *RealPrometheusCollector) queryPrometheus(metric string) (float64, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", pc.baseURL, metric)

	resp, err := pc.httpClient.Get(url)
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

	if len(result.Data.Result) == 0 {
		return 0, fmt.Errorf("no data for metric: %s", metric)
	}

	// Парсим значение (Prometheus возвращает [timestamp, value])
	if len(result.Data.Result[0].Value) < 2 {
		return 0, fmt.Errorf("invalid prometheus response format")
	}

	valueStr, ok := result.Data.Result[0].Value[1].(string)
	if !ok {
		return 0, fmt.Errorf("invalid value type")
	}

	var value float64
	if _, err := fmt.Sscanf(valueStr, "%f", &value); err != nil {
		return 0, err
	}

	return value, nil
}

// Start запускает коллектор
func (pc *RealPrometheusCollector) Start() error {
	pc.logger.Info("Real Prometheus collector started", zap.String("base_url", pc.baseURL))
	return nil
}

// Stop останавливает коллектор
func (pc *RealPrometheusCollector) Stop() error {
	pc.logger.Info("Real Prometheus collector stopped")
	return nil
}

// RealSystemCollector собирает реальные системные данные
type RealSystemCollector struct {
	logger *zap.Logger
	cache  map[string]interface{}
	mutex  sync.RWMutex
}

// NewRealSystemCollector создает новый системный коллектор
func NewRealSystemCollector(logger *zap.Logger) *RealSystemCollector {
	return &RealSystemCollector{
		logger: logger,
		cache:  make(map[string]interface{}),
	}
}

// Collect собирает системные данные
func (sc *RealSystemCollector) Collect() ([]DataPoint, error) {
	var dataPoints []DataPoint

	// CPU данные
	cpuData, err := sc.collectCPUData()
	if err != nil {
		sc.logger.Warn("Failed to collect CPU data", zap.Error(err))
	} else {
		dataPoints = append(dataPoints, cpuData...)
	}

	// Memory данные
	memData, err := sc.collectMemoryData()
	if err != nil {
		sc.logger.Warn("Failed to collect memory data", zap.Error(err))
	} else {
		dataPoints = append(dataPoints, memData...)
	}

	// Disk данные
	diskData, err := sc.collectDiskData()
	if err != nil {
		sc.logger.Warn("Failed to collect disk data", zap.Error(err))
	} else {
		dataPoints = append(dataPoints, diskData...)
	}

	// Go runtime данные
	runtimeData := sc.collectRuntimeData()
	dataPoints = append(dataPoints, runtimeData...)

	return dataPoints, nil
}

// collectCPUData собирает данные о CPU
func (sc *RealSystemCollector) collectCPUData() ([]DataPoint, error) {
	percentages, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, err
	}

	var dataPoints []DataPoint

	// Общее использование CPU
	totalCPU := 0.0
	for _, pct := range percentages {
		totalCPU += pct
	}
	totalCPU /= float64(len(percentages))

	dataPoints = append(dataPoints, DataPoint{
		Timestamp: time.Now(),
		Source:    "system",
		Type:      "cpu",
		Data: map[string]interface{}{
			"cpu_usage_percent": totalCPU,
			"cpu_cores":         len(percentages),
			"cpu_cores_usage":   percentages,
		},
	})

	return dataPoints, nil
}

// collectMemoryData собирает данные о памяти
func (sc *RealSystemCollector) collectMemoryData() ([]DataPoint, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "system",
			Type:      "memory",
			Data: map[string]interface{}{
				"memory_total_bytes":     memInfo.Total,
				"memory_available_bytes": memInfo.Available,
				"memory_used_bytes":      memInfo.Used,
				"memory_usage_percent":   memInfo.UsedPercent,
				"memory_cached_bytes":    memInfo.Cached,
				"memory_buffers_bytes":   memInfo.Buffers,
			},
		},
	}, nil
}

// collectDiskData собирает данные о дисках
func (sc *RealSystemCollector) collectDiskData() ([]DataPoint, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var dataPoints []DataPoint

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		dataPoints = append(dataPoints, DataPoint{
			Timestamp: time.Now(),
			Source:    "system",
			Type:      "disk",
			Data: map[string]interface{}{
				"disk_device":               partition.Device,
				"disk_mountpoint":           partition.Mountpoint,
				"disk_fstype":               partition.Fstype,
				"disk_total_bytes":          usage.Total,
				"disk_used_bytes":           usage.Used,
				"disk_free_bytes":           usage.Free,
				"disk_usage_percent":        usage.UsedPercent,
				"disk_inodes_total":         usage.InodesTotal,
				"disk_inodes_used":          usage.InodesUsed,
				"disk_inodes_free":          usage.InodesFree,
				"disk_inodes_usage_percent": usage.InodesUsedPercent,
			},
		})
	}

	return dataPoints, nil
}

// collectRuntimeData собирает данные Go runtime
func (sc *RealSystemCollector) collectRuntimeData() []DataPoint {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return []DataPoint{
		{
			Timestamp: time.Now(),
			Source:    "system",
			Type:      "runtime",
			Data: map[string]interface{}{
				"go_routines":           runtime.NumGoroutine(),
				"go_memory_alloc_bytes": m.Alloc,
				"go_memory_total_bytes": m.TotalAlloc,
				"go_memory_sys_bytes":   m.Sys,
				"go_memory_heap_bytes":  m.HeapAlloc,
				"go_memory_stack_bytes": m.StackInuse,
				"go_gc_cycles":          m.NumGC,
				"go_gc_pause_ns":        m.PauseNs[(m.NumGC+255)%256],
			},
		},
	}
}

// Start запускает коллектор
func (sc *RealSystemCollector) Start() error {
	sc.logger.Info("Real system collector started")
	return nil
}

// Stop останавливает коллектор
func (sc *RealSystemCollector) Stop() error {
	sc.logger.Info("Real system collector stopped")
	return nil
}

// RealNetworkCollector собирает реальные сетевые данные
type RealNetworkCollector struct {
	logger     *zap.Logger
	interfaces []string
	prevStats  map[string]net.IOCountersStat
	mutex      sync.RWMutex
}

// NewRealNetworkCollector создает новый сетевой коллектор
func NewRealNetworkCollector(logger *zap.Logger) *RealNetworkCollector {
	return &RealNetworkCollector{
		logger:     logger,
		interfaces: []string{"all"}, // Собираем данные по всем интерфейсам
		prevStats:  make(map[string]net.IOCountersStat),
	}
}

// Collect собирает сетевые данные
func (nc *RealNetworkCollector) Collect() ([]DataPoint, error) {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()

	stats, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	var dataPoints []DataPoint
	now := time.Now()

	for _, stat := range stats {
		// Пропускаем loopback интерфейсы
		if stat.Name == "lo" || stat.Name == "lo0" {
			continue
		}

		// Вычисляем дельты для rates
		var bytesInRate, bytesOutRate, packetsInRate, packetsOutRate float64
		if prevStat, exists := nc.prevStats[stat.Name]; exists {
			timeDiff := float64(now.Sub(time.Now().Add(-time.Second)).Nanoseconds()) / 1e9
			if timeDiff > 0 {
				bytesInRate = float64(stat.BytesRecv-prevStat.BytesRecv) / timeDiff
				bytesOutRate = float64(stat.BytesSent-prevStat.BytesSent) / timeDiff
				packetsInRate = float64(stat.PacketsRecv-prevStat.PacketsRecv) / timeDiff
				packetsOutRate = float64(stat.PacketsSent-prevStat.PacketsSent) / timeDiff
			}
		}

		// Сохраняем текущие статистики
		nc.prevStats[stat.Name] = stat

		dataPoints = append(dataPoints, DataPoint{
			Timestamp: now,
			Source:    "network",
			Type:      "interface",
			Data: map[string]interface{}{
				"interface_name":    stat.Name,
				"bytes_in_total":    stat.BytesRecv,
				"bytes_out_total":   stat.BytesSent,
				"bytes_in_rate":     bytesInRate,
				"bytes_out_rate":    bytesOutRate,
				"packets_in_total":  stat.PacketsRecv,
				"packets_out_total": stat.PacketsSent,
				"packets_in_rate":   packetsInRate,
				"packets_out_rate":  packetsOutRate,
				"errors_in_total":   stat.Errin,
				"errors_out_total":  stat.Errout,
				"drops_in_total":    stat.Dropin,
				"drops_out_total":   stat.Dropout,
			},
		})
	}

	// Собираем общие сетевые соединения
	connections, err := net.Connections("all")
	if err == nil {
		connectionCounts := make(map[string]int)
		for _, conn := range connections {
			connectionCounts[conn.Status]++
		}

		dataPoints = append(dataPoints, DataPoint{
			Timestamp: now,
			Source:    "network",
			Type:      "connections",
			Data: map[string]interface{}{
				"total_connections":    len(connections),
				"connection_by_status": connectionCounts,
			},
		})
	}

	return dataPoints, nil
}

// Start запускает коллектор
func (nc *RealNetworkCollector) Start() error {
	nc.logger.Info("Real network collector started")
	return nil
}

// Stop останавливает коллектор
func (nc *RealNetworkCollector) Stop() error {
	nc.logger.Info("Real network collector stopped")
	return nil
}

// RelayMetricsCollector собирает метрики из Relay сервера
type RelayMetricsCollector struct {
	logger     *zap.Logger
	relayStats *RelayStats
	mutex      sync.RWMutex
	httpClient *http.Client
	relayURL   string
}

// RelayStats содержит статистики Relay сервера
type RelayStats struct {
	ConnectionsTotal     int64          `json:"connections_total"`
	ConnectionsActive    int64          `json:"connections_active"`
	BytesTransferred     int64          `json:"bytes_transferred"`
	PacketsTransferred   int64          `json:"packets_transferred"`
	AverageLatency       float64        `json:"average_latency"`
	ErrorRate            float64        `json:"error_rate"`
	ProtocolDistribution map[string]int `json:"protocol_distribution"`
	LastUpdated          time.Time      `json:"last_updated"`
}

// NewRelayMetricsCollector создает новый коллектор метрик Relay
func NewRelayMetricsCollector(logger *zap.Logger) *RelayMetricsCollector {
	return &RelayMetricsCollector{
		logger: logger,
		relayStats: &RelayStats{
			LastUpdated: time.Now(),
			ProtocolDistribution: map[string]int{
				"QUIC":     0,
				"P2P_QUIC": 0,
				"TCP":      0,
			},
		},
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		relayURL: "http://cloudbridge-relay.cloudbridge.svc.cluster.local:8080",
	}
}

// UpdateStats обновляет статистики Relay сервера
func (rmc *RelayMetricsCollector) UpdateStats(stats *RelayStats) {
	rmc.mutex.Lock()
	defer rmc.mutex.Unlock()

	rmc.relayStats = stats
	rmc.relayStats.LastUpdated = time.Now()
}

// Collect собирает метрики Relay сервера
func (rmc *RelayMetricsCollector) Collect() ([]DataPoint, error) {
	// Попытаемся получить реальные данные от Relay сервера
	realStats, err := rmc.fetchRelayStats()
	if err != nil {
		rmc.logger.Warn("Failed to fetch real relay stats, using cached data", zap.Error(err))
		// Используем кэшированные данные в случае ошибки
		rmc.mutex.RLock()
		defer rmc.mutex.RUnlock()

		return []DataPoint{
			{
				Timestamp: rmc.relayStats.LastUpdated,
				Source:    "relay",
				Type:      "metrics",
				Data: map[string]interface{}{
					"connections_total":     rmc.relayStats.ConnectionsTotal,
					"connections_active":    rmc.relayStats.ConnectionsActive,
					"bytes_transferred":     rmc.relayStats.BytesTransferred,
					"packets_transferred":   rmc.relayStats.PacketsTransferred,
					"average_latency":       rmc.relayStats.AverageLatency,
					"error_rate":            rmc.relayStats.ErrorRate,
					"protocol_distribution": rmc.relayStats.ProtocolDistribution,
				},
			},
		}, nil
	}

	// Обновляем кэшированные данные
	rmc.UpdateStats(realStats)

	return []DataPoint{
		{
			Timestamp: realStats.LastUpdated,
			Source:    "relay",
			Type:      "metrics",
			Data: map[string]interface{}{
				"connections_total":     realStats.ConnectionsTotal,
				"connections_active":    realStats.ConnectionsActive,
				"bytes_transferred":     realStats.BytesTransferred,
				"packets_transferred":   realStats.PacketsTransferred,
				"average_latency":       realStats.AverageLatency,
				"error_rate":            realStats.ErrorRate,
				"protocol_distribution": realStats.ProtocolDistribution,
			},
		},
	}, nil
}

// Start запускает коллектор
func (rmc *RelayMetricsCollector) Start() error {
	rmc.logger.Info("Relay metrics collector started")
	return nil
}

// Stop останавливает коллектор
func (rmc *RelayMetricsCollector) Stop() error {
	rmc.logger.Info("Relay metrics collector stopped")
	return nil
}

// fetchRelayStats получает реальные статистики от Relay сервера
func (rmc *RelayMetricsCollector) fetchRelayStats() (*RelayStats, error) {
	// Получаем health endpoint от Relay сервера
	healthURL := fmt.Sprintf("%s/health", rmc.relayURL)
	resp, err := rmc.httpClient.Get(healthURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relay health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("relay health check failed with status: %d", resp.StatusCode)
	}

	var healthResponse struct {
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
		Uptime    string `json:"uptime"`
		Version   string `json:"version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}

	// Получаем метрики от Relay сервера
	metricsURL := fmt.Sprintf("%s/metrics", rmc.relayURL)
	metricsResp, err := rmc.httpClient.Get(metricsURL)
	if err != nil {
		// Если метрики недоступны, попробуем получить данные через API
		rmc.logger.Warn("Failed to fetch relay metrics, trying API endpoint", zap.Error(err))

		// Попытка получить данные через API endpoint
		apiURL := fmt.Sprintf("%s/api/v1/stats", rmc.relayURL)
		apiResp, apiErr := rmc.httpClient.Get(apiURL)
		if apiErr != nil {
			rmc.logger.Warn("Failed to fetch relay API stats, using zero values", zap.Error(apiErr))
			return &RelayStats{
				ConnectionsTotal:   0,
				ConnectionsActive:  0,
				BytesTransferred:   0,
				PacketsTransferred: 0,
				AverageLatency:     0.0,
				ErrorRate:          0.0,
				LastUpdated:        time.Now(),
			}, nil
		}
		defer apiResp.Body.Close()

		// Парсим API ответ
		var apiStats struct {
			ConnectionsTotal   int64   `json:"connections_total"`
			ConnectionsActive  int64   `json:"connections_active"`
			BytesTransferred   int64   `json:"bytes_transferred"`
			PacketsTransferred int64   `json:"packets_transferred"`
			AverageLatency     float64 `json:"average_latency"`
			ErrorRate          float64 `json:"error_rate"`
		}

		if err := json.NewDecoder(apiResp.Body).Decode(&apiStats); err != nil {
			rmc.logger.Warn("Failed to decode API response, using zero values", zap.Error(err))
			return &RelayStats{
				ConnectionsTotal:   0,
				ConnectionsActive:  0,
				BytesTransferred:   0,
				PacketsTransferred: 0,
				AverageLatency:     0.0,
				ErrorRate:          0.0,
				LastUpdated:        time.Now(),
			}, nil
		}

		return &RelayStats{
			ConnectionsTotal:   apiStats.ConnectionsTotal,
			ConnectionsActive:  apiStats.ConnectionsActive,
			BytesTransferred:   apiStats.BytesTransferred,
			PacketsTransferred: apiStats.PacketsTransferred,
			AverageLatency:     apiStats.AverageLatency,
			ErrorRate:          apiStats.ErrorRate,
			LastUpdated:        time.Now(),
		}, nil
	}
	defer metricsResp.Body.Close()

	// Парсим Prometheus метрики
	stats := &RelayStats{
		LastUpdated: time.Now(),
		ProtocolDistribution: map[string]int{
			"QUIC":      0,
			"WireGuard": 0,
			"TCP":       0,
		},
	}

	// Читаем метрики построчно
	scanner := bufio.NewScanner(metricsResp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Парсим метрики
		if strings.HasPrefix(line, "relay_connections_total") {
			if value := rmc.parseMetricValue(line); value > 0 {
				stats.ConnectionsTotal = int64(value)
			}
		} else if strings.HasPrefix(line, "relay_active_connections") {
			if value := rmc.parseMetricValue(line); value > 0 {
				stats.ConnectionsActive = int64(value)
			}
		} else if strings.HasPrefix(line, "relay_bandwidth_bytes_total") {
			if value := rmc.parseMetricValue(line); value > 0 {
				stats.BytesTransferred = int64(value)
			}
		} else if strings.HasPrefix(line, "relay_latency_seconds") {
			if value := rmc.parseMetricValue(line); value > 0 {
				stats.AverageLatency = value
			}
		} else if strings.HasPrefix(line, "relay_errors_total") {
			if value := rmc.parseMetricValue(line); value > 0 {
				// Вычисляем error rate
				if stats.ConnectionsTotal > 0 {
					stats.ErrorRate = value / float64(stats.ConnectionsTotal)
				}
			}
		} else if strings.Contains(line, "protocol=\"QUIC\"") {
			if value := rmc.parseMetricValue(line); value > 0 {
				stats.ProtocolDistribution["QUIC"] = int(value)
			}
		} else if strings.Contains(line, "protocol=\"P2P_QUIC\"") {
			if value := rmc.parseMetricValue(line); value > 0 {
				stats.ProtocolDistribution["P2P_QUIC"] = int(value)
			}
		} else if strings.Contains(line, "protocol=\"TCP\"") {
			if value := rmc.parseMetricValue(line); value > 0 {
				stats.ProtocolDistribution["TCP"] = int(value)
			}
		}
	}

	return stats, nil
}

// parseMetricValue парсит значение из строки метрики Prometheus
func (rmc *RelayMetricsCollector) parseMetricValue(line string) float64 {
	// Формат: metric_name{labels} value
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return 0
	}

	value, err := strconv.ParseFloat(parts[len(parts)-1], 64)
	if err != nil {
		return 0
	}

	return value
}
