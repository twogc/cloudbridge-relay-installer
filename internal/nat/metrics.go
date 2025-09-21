package nat

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

// NATBehaviorType представляет тип NAT поведения согласно RFC 3489
type NATBehaviorType string

const (
	NATBehaviorUnknown           NATBehaviorType = "unknown"
	NATBehaviorFullCone          NATBehaviorType = "full_cone"
	NATBehaviorRestrictedCone    NATBehaviorType = "restricted_cone"
	NATBehaviorPortRestrictedCone NATBehaviorType = "port_restricted_cone"
	NATBehaviorSymmetric         NATBehaviorType = "symmetric"
	NATBehaviorNoNAT             NATBehaviorType = "no_nat"
)

// NATMetrics собирает метрики NAT поведения
type NATMetrics struct {
	logger *logger.Logger

	// Метрики по клиентам
	clientMetrics      map[string]*ClientNATMetrics
	clientMetricsMutex sync.RWMutex

	// Метрики по STUN серверам
	stunServerMetrics      map[string]*STUNServerMetrics
	stunServerMetricsMutex sync.RWMutex

	// Метрики по TURN серверам
	turnServerMetrics      map[string]*TURNServerMetrics
	turnServerMetricsMutex sync.RWMutex

	// Общие метрики
	totalClients      uint64
	totalSTUNRequests uint64
	totalTURNRequests uint64
	totalAllocations  uint64
	totalErrors       uint64

	ctx    context.Context
	cancel context.CancelFunc
}

// ClientNATMetrics метрики NAT поведения для клиента
type ClientNATMetrics struct {
	ClientID           string
	PublicIP           string
	NATBehavior        NATBehaviorType
	LastSeen           time.Time
	STUNRequests       uint64
	TURNRequests       uint64
	Allocations        uint64
	Errors             uint64
	AverageLatency     time.Duration
	LastLatency        time.Duration
	ConnectivityChecks uint64
	SuccessfulChecks   uint64
}

// STUNServerMetrics метрики STUN сервера
type STUNServerMetrics struct {
	ServerAddress      string
	TotalRequests      uint64
	SuccessfulRequests uint64
	FailedRequests     uint64
	AverageLatency     time.Duration
	LastRequest        time.Time
	ActiveClients      uint64
}

// TURNServerMetrics метрики TURN сервера
type TURNServerMetrics struct {
	ServerAddress      string
	TotalRequests      uint64
	SuccessfulRequests uint64
	FailedRequests     uint64
	ActiveAllocations  uint64
	TotalAllocations   uint64
	BandwidthIn        uint64
	BandwidthOut       uint64
	AverageLatency     time.Duration
	LastRequest        time.Time
}

// NewNATMetrics создает новый экземпляр NATMetrics
func NewNATMetrics(logger *logger.Logger) *NATMetrics {
	ctx, cancel := context.WithCancel(context.Background())
	nm := &NATMetrics{
		logger:            logger,
		clientMetrics:     make(map[string]*ClientNATMetrics),
		stunServerMetrics: make(map[string]*STUNServerMetrics),
		turnServerMetrics: make(map[string]*TURNServerMetrics),
		ctx:               ctx,
		cancel:            cancel,
	}
	go nm.cleanupExpiredMetrics()
	return nm
}

// RecordClientNATBehavior записывает NAT поведение клиента
func (nm *NATMetrics) RecordClientNATBehavior(clientID, publicIP string, behavior NATBehaviorType) {
	nm.clientMetricsMutex.Lock()
	defer nm.clientMetricsMutex.Unlock()

	if _, exists := nm.clientMetrics[clientID]; !exists {
		nm.clientMetrics[clientID] = &ClientNATMetrics{
			ClientID: clientID,
			PublicIP: publicIP,
		}
		nm.totalClients++
	}

	nm.clientMetrics[clientID].NATBehavior = behavior
	nm.clientMetrics[clientID].LastSeen = time.Now()

	nm.logger.Debug("Recorded client NAT behavior",
		zap.String("client_id", clientID),
		zap.String("public_ip", publicIP),
		zap.String("nat_behavior", string(behavior)),
	)
}

// RecordSTUNRequest записывает STUN запрос
func (nm *NATMetrics) RecordSTUNRequest(clientID, serverAddress string, success bool, latency time.Duration) {
	nm.clientMetricsMutex.Lock()
	if clientMetrics, exists := nm.clientMetrics[clientID]; exists {
		clientMetrics.STUNRequests++
		clientMetrics.LastLatency = latency
		clientMetrics.AverageLatency = nm.calculateAverageLatency(clientMetrics.AverageLatency, latency, clientMetrics.STUNRequests)
		if !success {
			clientMetrics.Errors++
		}
	}
	nm.clientMetricsMutex.Unlock()

	nm.stunServerMetricsMutex.Lock()
	defer nm.stunServerMetricsMutex.Unlock()

	if _, exists := nm.stunServerMetrics[serverAddress]; !exists {
		nm.stunServerMetrics[serverAddress] = &STUNServerMetrics{
			ServerAddress: serverAddress,
		}
	}

	serverMetrics := nm.stunServerMetrics[serverAddress]
	serverMetrics.TotalRequests++
	serverMetrics.LastRequest = time.Now()
	serverMetrics.AverageLatency = nm.calculateAverageLatency(serverMetrics.AverageLatency, latency, serverMetrics.TotalRequests)

	if success {
		serverMetrics.SuccessfulRequests++
	} else {
		serverMetrics.FailedRequests++
	}

	nm.totalSTUNRequests++
}

// RecordTURNRequest записывает TURN запрос
func (nm *NATMetrics) RecordTURNRequest(clientID, serverAddress string, success bool, latency time.Duration) {
	nm.clientMetricsMutex.Lock()
	if clientMetrics, exists := nm.clientMetrics[clientID]; exists {
		clientMetrics.TURNRequests++
		clientMetrics.LastLatency = latency
		clientMetrics.AverageLatency = nm.calculateAverageLatency(clientMetrics.AverageLatency, latency, clientMetrics.TURNRequests)
		if !success {
			clientMetrics.Errors++
		}
	}
	nm.clientMetricsMutex.Unlock()

	nm.turnServerMetricsMutex.Lock()
	defer nm.turnServerMetricsMutex.Unlock()

	if _, exists := nm.turnServerMetrics[serverAddress]; !exists {
		nm.turnServerMetrics[serverAddress] = &TURNServerMetrics{
			ServerAddress: serverAddress,
		}
	}

	serverMetrics := nm.turnServerMetrics[serverAddress]
	serverMetrics.TotalRequests++
	serverMetrics.LastRequest = time.Now()
	serverMetrics.AverageLatency = nm.calculateAverageLatency(serverMetrics.AverageLatency, latency, serverMetrics.TotalRequests)

	if success {
		serverMetrics.SuccessfulRequests++
	} else {
		serverMetrics.FailedRequests++
	}

	nm.totalTURNRequests++
}

// RecordTURNAllocation записывает TURN allocation
func (nm *NATMetrics) RecordTURNAllocation(clientID, serverAddress string, bandwidthIn, bandwidthOut uint64) {
	nm.clientMetricsMutex.Lock()
	if clientMetrics, exists := nm.clientMetrics[clientID]; exists {
		clientMetrics.Allocations++
	}
	nm.clientMetricsMutex.Unlock()

	nm.turnServerMetricsMutex.Lock()
	defer nm.turnServerMetricsMutex.Unlock()

	if serverMetrics, exists := nm.turnServerMetrics[serverAddress]; exists {
		serverMetrics.ActiveAllocations++
		serverMetrics.TotalAllocations++
		serverMetrics.BandwidthIn += bandwidthIn
		serverMetrics.BandwidthOut += bandwidthOut
	}

	nm.totalAllocations++
}

// RecordConnectivityCheck записывает ICE connectivity check
func (nm *NATMetrics) RecordConnectivityCheck(clientID string, success bool) {
	nm.clientMetricsMutex.Lock()
	defer nm.clientMetricsMutex.Unlock()

	if clientMetrics, exists := nm.clientMetrics[clientID]; exists {
		clientMetrics.ConnectivityChecks++
		if success {
			clientMetrics.SuccessfulChecks++
		}
	}
}

// GetAllMetrics возвращает все метрики
func (nm *NATMetrics) GetAllMetrics() map[string]interface{} {
	nm.clientMetricsMutex.RLock()
	nm.stunServerMetricsMutex.RLock()
	nm.turnServerMetricsMutex.RLock()
	defer nm.clientMetricsMutex.RUnlock()
	defer nm.stunServerMetricsMutex.RUnlock()
	defer nm.turnServerMetricsMutex.RUnlock()

	return map[string]interface{}{
		"total_clients":       nm.totalClients,
		"total_stun_requests": nm.totalSTUNRequests,
		"total_turn_requests": nm.totalTURNRequests,
		"total_allocations":   nm.totalAllocations,
		"total_errors":        nm.totalErrors,
		"active_clients":      len(nm.clientMetrics),
		"active_stun_servers": len(nm.stunServerMetrics),
		"active_turn_servers": len(nm.turnServerMetrics),
	}
}

// calculateAverageLatency вычисляет среднюю задержку
func (nm *NATMetrics) calculateAverageLatency(current, new time.Duration, count uint64) time.Duration {
	if count == 1 {
		return new
	}
	return time.Duration((int64(current)*int64(count-1) + int64(new)) / int64(count))
}

// cleanupExpiredMetrics очищает просроченные метрики
func (nm *NATMetrics) cleanupExpiredMetrics() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-nm.ctx.Done():
			return
		case <-ticker.C:
			nm.cleanupExpiredClientMetrics()
			nm.cleanupExpiredServerMetrics()
		}
	}
}

// cleanupExpiredClientMetrics очищает просроченные метрики клиентов
func (nm *NATMetrics) cleanupExpiredClientMetrics() {
	nm.clientMetricsMutex.Lock()
	defer nm.clientMetricsMutex.Unlock()

	expiredTime := time.Now().Add(-1 * time.Hour) // Удаляем метрики старше 1 часа
	for clientID, metrics := range nm.clientMetrics {
		if metrics.LastSeen.Before(expiredTime) {
			delete(nm.clientMetrics, clientID)
			nm.logger.Debug("Cleaned up expired client metrics", zap.String("client_id", clientID))
		}
	}
}

// cleanupExpiredServerMetrics очищает просроченные метрики серверов
func (nm *NATMetrics) cleanupExpiredServerMetrics() {
	nm.stunServerMetricsMutex.Lock()
	defer nm.stunServerMetricsMutex.Unlock()

	nm.turnServerMetricsMutex.Lock()
	defer nm.turnServerMetricsMutex.Unlock()

	expiredTime := time.Now().Add(-1 * time.Hour) // Удаляем метрики старше 1 часа

	// Очищаем STUN серверы
	for serverAddress, metrics := range nm.stunServerMetrics {
		if metrics.LastRequest.Before(expiredTime) {
			delete(nm.stunServerMetrics, serverAddress)
			nm.logger.Debug("Cleaned up expired STUN server metrics", zap.String("server", serverAddress))
		}
	}

	// Очищаем TURN серверы
	for serverAddress, metrics := range nm.turnServerMetrics {
		if metrics.LastRequest.Before(expiredTime) {
			delete(nm.turnServerMetrics, serverAddress)
			nm.logger.Debug("Cleaned up expired TURN server metrics", zap.String("server", serverAddress))
		}
	}
}

// Stop останавливает NATMetrics
func (nm *NATMetrics) Stop() {
	nm.cancel()
	nm.logger.Info("NAT metrics stopped")
}

// DetectNATBehavior определяет тип NAT поведения клиента
func (nm *NATMetrics) DetectNATBehavior(clientID string, stunServers []string) (NATBehaviorType, error) {
	// Здесь должна быть логика определения NAT поведения
	// Пока возвращаем неизвестный тип
	return NATBehaviorUnknown, fmt.Errorf("NAT behavior detection not implemented yet")
}

// GetNATBehaviorStats возвращает статистику по типам NAT поведения
func (nm *NATMetrics) GetNATBehaviorStats() map[NATBehaviorType]uint64 {
	nm.clientMetricsMutex.RLock()
	defer nm.clientMetricsMutex.RUnlock()

	stats := make(map[NATBehaviorType]uint64)
	for _, metrics := range nm.clientMetrics {
		stats[metrics.NATBehavior]++
	}

	return stats
}

// GetClientMetrics возвращает метрики клиента
func (nm *NATMetrics) GetClientMetrics(clientID string) (*ClientNATMetrics, bool) {
	nm.clientMetricsMutex.RLock()
	defer nm.clientMetricsMutex.RUnlock()

	metrics, exists := nm.clientMetrics[clientID]
	return metrics, exists
}

// GetSTUNServerMetrics возвращает метрики STUN сервера
func (nm *NATMetrics) GetSTUNServerMetrics(serverAddress string) (*STUNServerMetrics, bool) {
	nm.stunServerMetricsMutex.RLock()
	defer nm.stunServerMetricsMutex.RUnlock()

	metrics, exists := nm.stunServerMetrics[serverAddress]
	return metrics, exists
}

// GetTURNServerMetrics возвращает метрики TURN сервера
func (nm *NATMetrics) GetTURNServerMetrics(serverAddress string) (*TURNServerMetrics, bool) {
	nm.turnServerMetricsMutex.RLock()
	defer nm.turnServerMetricsMutex.RUnlock()

	metrics, exists := nm.turnServerMetrics[serverAddress]
	return metrics, exists
}