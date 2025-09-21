package nat

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// NATMetrics представляет метрики для NAT traversal
type NATMetrics struct {
	logger *zap.Logger

	// Prometheus метрики
	iceCandidatesTotal    *prometheus.CounterVec
	iceConnectionsTotal   *prometheus.CounterVec
	iceConnectionDuration *prometheus.HistogramVec
	natTypeDetected       *prometheus.CounterVec
	turnAllocationsTotal  *prometheus.CounterVec
	stunRequestsTotal     *prometheus.CounterVec
	connectionSuccessRate *prometheus.GaugeVec

	// Внутренние метрики
	mu              sync.RWMutex
	natTypeStats    map[string]int64
	connectionStats map[string]*ConnectionStats
	candidateStats  map[string]*CandidateStats
	turnServerStats map[string]*TURNServerStats
}

// ConnectionStats статистика соединений
type ConnectionStats struct {
	TotalAttempts   int64
	SuccessfulConns int64
	FailedConns     int64
	AverageLatency  time.Duration
	LastAttempt     time.Time
	LastSuccess     time.Time
}

// CandidateStats статистика кандидатов
type CandidateStats struct {
	HostCandidates       int64
	ServerReflexiveCands int64
	RelayCandidates      int64
	PeerReflexiveCands   int64
	TotalCandidates      int64
}

// TURNServerStats статистика TURN серверов
type TURNServerStats struct {
	TotalAllocations  int64
	ActiveAllocations int64
	FailedAllocations int64
	AverageLatency    time.Duration
	LastAllocation    time.Time
}

// NewNATMetrics создает новый экземпляр NAT метрик
func NewNATMetrics(logger *zap.Logger) *NATMetrics {
	return &NATMetrics{
		logger: logger,
		iceCandidatesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ice_candidates_total",
				Help: "Total number of ICE candidates gathered",
			},
			[]string{"type", "transport"},
		),
		iceConnectionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ice_connections_total",
				Help: "Total number of ICE connection attempts",
			},
			[]string{"status", "nat_type"},
		),
		iceConnectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ice_connection_duration_seconds",
				Help:    "Duration of ICE connection establishment",
				Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
			},
			[]string{"nat_type", "success"},
		),
		natTypeDetected: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "nat_type_detected_total",
				Help: "Total number of NAT type detections",
			},
			[]string{"nat_type"},
		),
		turnAllocationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "turn_allocations_total",
				Help: "Total number of TURN allocations",
			},
			[]string{"server", "status"},
		),
		stunRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "stun_requests_total",
				Help: "Total number of STUN requests",
			},
			[]string{"server", "status"},
		),
		connectionSuccessRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ice_connection_success_rate",
				Help: "ICE connection success rate by NAT type",
			},
			[]string{"nat_type"},
		),
		natTypeStats:    make(map[string]int64),
		connectionStats: make(map[string]*ConnectionStats),
		candidateStats:  make(map[string]*CandidateStats),
		turnServerStats: make(map[string]*TURNServerStats),
	}
}

// RecordICECandidate записывает метрику ICE кандидата
func (nm *NATMetrics) RecordICECandidate(candidateType, transport string) {
	nm.iceCandidatesTotal.WithLabelValues(candidateType, transport).Inc()

	nm.mu.Lock()
	defer nm.mu.Unlock()

	stats, exists := nm.candidateStats[transport]
	if !exists {
		stats = &CandidateStats{}
		nm.candidateStats[transport] = stats
	}

	stats.TotalCandidates++
	switch candidateType {
	case "host":
		stats.HostCandidates++
	case "srflx":
		stats.ServerReflexiveCands++
	case "relay":
		stats.RelayCandidates++
	case "prflx":
		stats.PeerReflexiveCands++
	}
}

// RecordICEConnection записывает метрику ICE соединения
func (nm *NATMetrics) RecordICEConnection(status, natType string, duration time.Duration) {
	nm.iceConnectionsTotal.WithLabelValues(status, natType).Inc()
	nm.iceConnectionDuration.WithLabelValues(natType, status).Observe(duration.Seconds())

	nm.mu.Lock()
	defer nm.mu.Unlock()

	stats, exists := nm.connectionStats[natType]
	if !exists {
		stats = &ConnectionStats{}
		nm.connectionStats[natType] = stats
	}

	stats.TotalAttempts++
	stats.LastAttempt = time.Now()

	if status == "success" {
		stats.SuccessfulConns++
		stats.LastSuccess = time.Now()
	} else {
		stats.FailedConns++
	}

	// Обновляем среднюю задержку
	if stats.AverageLatency == 0 {
		stats.AverageLatency = duration
	} else {
		stats.AverageLatency = (stats.AverageLatency + duration) / 2
	}

	// Обновляем rate успешности
	successRate := float64(stats.SuccessfulConns) / float64(stats.TotalAttempts)
	nm.connectionSuccessRate.WithLabelValues(natType).Set(successRate)
}

// RecordNATType записывает метрику типа NAT
func (nm *NATMetrics) RecordNATType(natType string) {
	nm.natTypeDetected.WithLabelValues(natType).Inc()

	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.natTypeStats[natType]++
}

// RecordTURNAllocation записывает метрику TURN allocation
func (nm *NATMetrics) RecordTURNAllocation(server, status string, duration time.Duration) {
	nm.turnAllocationsTotal.WithLabelValues(server, status).Inc()

	nm.mu.Lock()
	defer nm.mu.Unlock()

	stats, exists := nm.turnServerStats[server]
	if !exists {
		stats = &TURNServerStats{}
		nm.turnServerStats[server] = stats
	}

	stats.TotalAllocations++
	stats.LastAllocation = time.Now()

	if status == "success" {
		stats.ActiveAllocations++
		if stats.AverageLatency == 0 {
			stats.AverageLatency = duration
		} else {
			stats.AverageLatency = (stats.AverageLatency + duration) / 2
		}
	} else {
		stats.FailedAllocations++
	}
}

// RecordSTUNRequest записывает метрику STUN запроса
func (nm *NATMetrics) RecordSTUNRequest(server, status string) {
	nm.stunRequestsTotal.WithLabelValues(server, status).Inc()
}

// GetNATTypeStats возвращает статистику типов NAT
func (nm *NATMetrics) GetNATTypeStats() map[string]int64 {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	result := make(map[string]int64)
	for natType, count := range nm.natTypeStats {
		result[natType] = count
	}
	return result
}

// GetConnectionStats возвращает статистику соединений
func (nm *NATMetrics) GetConnectionStats() map[string]*ConnectionStats {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	result := make(map[string]*ConnectionStats)
	for natType, stats := range nm.connectionStats {
		// Создаем копию для безопасности
		statsCopy := *stats
		result[natType] = &statsCopy
	}
	return result
}

// GetCandidateStats возвращает статистику кандидатов
func (nm *NATMetrics) GetCandidateStats() map[string]*CandidateStats {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	result := make(map[string]*CandidateStats)
	for transport, stats := range nm.candidateStats {
		// Создаем копию для безопасности
		statsCopy := *stats
		result[transport] = &statsCopy
	}
	return result
}

// GetTURNServerStats возвращает статистику TURN серверов
func (nm *NATMetrics) GetTURNServerStats() map[string]*TURNServerStats {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	result := make(map[string]*TURNServerStats)
	for server, stats := range nm.turnServerStats {
		// Создаем копию для безопасности
		statsCopy := *stats
		result[server] = &statsCopy
	}
	return result
}

// GetOverallStats возвращает общую статистику
func (nm *NATMetrics) GetOverallStats() map[string]interface{} {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	totalConnections := int64(0)
	totalSuccessful := int64(0)
	totalCandidates := int64(0)

	for _, stats := range nm.connectionStats {
		totalConnections += stats.TotalAttempts
		totalSuccessful += stats.SuccessfulConns
	}

	for _, stats := range nm.candidateStats {
		totalCandidates += stats.TotalCandidates
	}

	successRate := float64(0)
	if totalConnections > 0 {
		successRate = float64(totalSuccessful) / float64(totalConnections)
	}

	return map[string]interface{}{
		"total_connections":      totalConnections,
		"successful_connections": totalSuccessful,
		"success_rate":           successRate,
		"total_candidates":       totalCandidates,
		"nat_types_detected":     len(nm.natTypeStats),
		"turn_servers_active":    len(nm.turnServerStats),
	}
}

// Reset сбрасывает все метрики
func (nm *NATMetrics) Reset() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nm.natTypeStats = make(map[string]int64)
	nm.connectionStats = make(map[string]*ConnectionStats)
	nm.candidateStats = make(map[string]*CandidateStats)
	nm.turnServerStats = make(map[string]*TURNServerStats)

	nm.logger.Info("NAT metrics reset")
}
